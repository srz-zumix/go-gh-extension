package actions

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unicode/utf16"
)

// StepLog represents a single step's log file information
type StepLog struct {
	StepNumber int
	StepName   string
	FilePath   string
	Content    []byte
}

// JobLog represents a job's log structure containing all step logs
type JobLog struct {
	JobName  string
	StepLogs map[int]*StepLog // key: step number
}

// ListSteps returns a list of all steps in the job, sorted by step number
func (j *JobLog) ListSteps() []*StepLog {
	// Collect step numbers
	stepNumbers := make([]int, 0, len(j.StepLogs))
	for stepNumber := range j.StepLogs {
		stepNumbers = append(stepNumbers, stepNumber)
	}

	// Sort step numbers
	for i := 0; i < len(stepNumbers); i++ {
		for k := i + 1; k < len(stepNumbers); k++ {
			if stepNumbers[i] > stepNumbers[k] {
				stepNumbers[i], stepNumbers[k] = stepNumbers[k], stepNumbers[i]
			}
		}
	}

	// Build sorted StepLog array
	steps := make([]*StepLog, 0, len(stepNumbers))
	for _, stepNumber := range stepNumbers {
		steps = append(steps, j.StepLogs[stepNumber])
	}
	return steps
}

// GetStepLog retrieves the log for a specific step by number
func (j *JobLog) GetStepLog(stepNumber int) (*StepLog, error) {
	stepLog, exists := j.StepLogs[stepNumber]
	if !exists {
		return nil, fmt.Errorf("step number %d not found in job %q", stepNumber, j.JobName)
	}
	return stepLog, nil
}

// GetStepLogByName retrieves the log for a specific step by name
func (j *JobLog) GetStepLogByName(stepName string) (*StepLog, error) {
	for _, stepLog := range j.StepLogs {
		if stepLog.StepName == stepName {
			return stepLog, nil
		}
	}
	return nil, fmt.Errorf("step log not found for step name: %s", stepName)
}

// WorkflowRunLogArchive represents the entire workflow run log archive structure
type WorkflowRunLogArchive struct {
	JobLogs map[string]*JobLog // key: job name
}

// NewWorkflowRunLogArchive creates a new WorkflowRunLogArchive instance by fetching and parsing logs
func NewWorkflowRunLogArchive(ctx context.Context, zipReader *zip.Reader) (*WorkflowRunLogArchive, error) {
	// Create the archive instance
	archive := &WorkflowRunLogArchive{
		JobLogs: make(map[string]*JobLog),
	}

	// Parse the zip file
	if err := archive.parseZipArchive(zipReader); err != nil {
		return nil, fmt.Errorf("failed to parse log archive: %w", err)
	}

	return archive, nil
}

// parseZipArchive parses the downloaded zip file and builds the log structure
func (w *WorkflowRunLogArchive) parseZipArchive(zipReader *zip.Reader) error {
	// Iterate through files in the zip archive
	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		// Parse the file path to extract job and step information
		// Expected format: "jobName/stepNumber_stepName.txt"
		parts := strings.Split(filepath.ToSlash(file.Name), "/")
		if len(parts) != 2 {
			// Skip files that don't match the expected structure
			continue
		}

		jobName := parts[0]
		stepFileName := parts[1]

		// Parse step number and name from the filename
		stepNumber, stepName, err := parseStepFileName(stepFileName)
		if err != nil {
			// Skip files that don't match the expected format
			continue
		}

		// Read the file content
		content, err := readZipFile(file)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", file.Name, err)
		}

		// Ensure the job exists in the map
		if _, exists := w.JobLogs[jobName]; !exists {
			w.JobLogs[jobName] = &JobLog{
				JobName:  jobName,
				StepLogs: make(map[int]*StepLog),
			}
		}

		// Add the step log
		w.JobLogs[jobName].StepLogs[stepNumber] = &StepLog{
			StepNumber: stepNumber,
			StepName:   stepName,
			FilePath:   file.Name,
			Content:    content,
		}
	}

	return nil
}

// parseStepFileName parses a step filename to extract step number and name
// Expected format: "1_stepName.txt" or "10_stepName.txt"
func parseStepFileName(filename string) (int, string, error) {
	// Remove .txt extension
	name := strings.TrimSuffix(filename, ".txt")

	// Find the first underscore
	underscoreIndex := strings.Index(name, "_")
	if underscoreIndex == -1 {
		return 0, "", fmt.Errorf("invalid step filename format: %s", filename)
	}

	// Parse step number
	var stepNumber int
	_, err := fmt.Sscanf(name[:underscoreIndex], "%d", &stepNumber)
	if err != nil {
		return 0, "", fmt.Errorf("failed to parse step number from %s: %w", filename, err)
	}

	// Extract step name
	stepName := name[underscoreIndex+1:]

	return stepNumber, stepName, nil
}

// readZipFile reads the content of a file from the zip archive
func readZipFile(file *zip.File) ([]byte, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return io.ReadAll(rc)
}

// sanitizeJobName converts a job name to the format used in log file names.
// GitHub Actions sanitizes job names by replacing certain characters and truncates to 90 UTF-16 code units.
// This is based on GitHub CLI's implementation.
func sanitizeJobName(jobName string) string {
	// Replace characters that GitHub sanitizes in log file names
	replacer := strings.NewReplacer(
		"/", "",
		":", "",
		"*", "",
		"?", "",
		"<", "",
		">", "",
		"|", "",
		"\"", "",
		"\\", "",
	)
	sanitized := replacer.Replace(jobName)

	// Truncate to 90 UTF-16 code units
	// GitHub Actions limits job names to 90 UTF-16 code units
	return truncateToUTF16Length(sanitized, 90)
}

// truncateToUTF16Length truncates a string to a maximum UTF-16 code unit length.
// It encodes the string to UTF-16, truncates to maxLen code units, and decodes back to UTF-8.
func truncateToUTF16Length(s string, maxLen int) string {
	// Encode to UTF-16
	utf16Encoded := utf16.Encode([]rune(s))

	// If already within limit, return original string
	if len(utf16Encoded) <= maxLen {
		return s
	}

	// Truncate UTF-16 slice
	truncated := utf16Encoded[:maxLen]

	// Decode back to UTF-8
	return string(utf16.Decode(truncated))
}

// findJobByName attempts to find a job by name, trying both the original name
// and the sanitized version.
func (w *WorkflowRunLogArchive) findJobByName(jobName string) (*JobLog, string, bool) {
	// Try the original name first
	if jobLog, exists := w.JobLogs[jobName]; exists {
		return jobLog, jobName, true
	}

	// Try the sanitized version
	sanitized := sanitizeJobName(jobName)
	if jobLog, exists := w.JobLogs[sanitized]; exists {
		return jobLog, sanitized, true
	}

	// Try case-insensitive match with sanitized names
	sanitizedLower := strings.ToLower(sanitized)
	for name, jobLog := range w.JobLogs {
		if strings.ToLower(name) == sanitizedLower {
			return jobLog, name, true
		}
	}

	return nil, "", false
}

// GetJobLog retrieves the log for a specific job by name.
// It attempts to match the job name by trying the original name, sanitized name,
// and case-insensitive matches.
func (w *WorkflowRunLogArchive) GetJobLog(jobName string) (*JobLog, error) {
	jobLog, _, exists := w.findJobByName(jobName)
	if !exists {
		return nil, fmt.Errorf("job %q not found in the logs", jobName)
	}
	return jobLog, nil
}

// GetStepLog retrieves the log for a specific step within a job
func (w *WorkflowRunLogArchive) GetStepLog(jobName string, stepNumber int) (*StepLog, error) {
	jobLog, err := w.GetJobLog(jobName)
	if err != nil {
		return nil, err
	}

	return jobLog.GetStepLog(stepNumber)
}

// GetStepLogByName retrieves the log for a specific step by name within a job
func (w *WorkflowRunLogArchive) GetStepLogByName(jobName string, stepName string) (*StepLog, error) {
	jobLog, err := w.GetJobLog(jobName)
	if err != nil {
		return nil, err
	}

	return jobLog.GetStepLogByName(stepName)
}

// ListJobs returns a list of all job names in the archive
func (w *WorkflowRunLogArchive) ListJobs() []string {
	jobs := make([]string, 0, len(w.JobLogs))
	for jobName := range w.JobLogs {
		jobs = append(jobs, jobName)
	}
	return jobs
}

// ListSteps returns a list of all steps for a given job, sorted by step number
func (w *WorkflowRunLogArchive) ListSteps(jobName string) ([]*StepLog, error) {
	jobLog, err := w.GetJobLog(jobName)
	if err != nil {
		return nil, err
	}

	return jobLog.ListSteps(), nil
}

// Walk iterates over all steps in a job and calls the provided function for each step.
func (w *WorkflowRunLogArchive) Walk(jobName string, fn func(stepLog *StepLog) error) error {
	jobLog, actualName, exists := w.findJobByName(jobName)
	if !exists {
		return fmt.Errorf("job %q not found in the logs", jobName)
	}

	_ = actualName // actualName is available if needed for logging
	for _, stepLog := range jobLog.StepLogs {
		if err := fn(stepLog); err != nil {
			return err
		}
	}
	return nil
}
