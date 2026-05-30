package actions

import (
	"os"
	"strings"
)

// GetWorkflowRef returns the reference path of the workflow that is running.
// It reads from the GITHUB_WORKFLOW_REF environment variable, e.g.
// "owner/repo/.github/workflows/my-workflow.yml@refs/heads/main".
func GetWorkflowRef() string {
	return os.Getenv("GITHUB_WORKFLOW_REF")
}

// GetJobName returns the job_id of the current job.
// It reads from the GITHUB_JOB environment variable.
func GetJobName() string {
	return os.Getenv("GITHUB_JOB")
}

// GetSHA returns the commit SHA that triggered the workflow run.
// It reads from the GITHUB_SHA environment variable.
func GetSHA() string {
	return os.Getenv("GITHUB_SHA")
}

// GetRunID returns the unique number for the current workflow run.
// It reads from the GITHUB_RUN_ID environment variable.
func GetRunID() string {
	return os.Getenv("GITHUB_RUN_ID")
}

// WorkflowFilePathFromRef extracts the repository-relative workflow file path
// from a GITHUB_WORKFLOW_REF value. For
// "owner/repo/.github/workflows/my-workflow.yml@refs/heads/main" it returns
// ".github/workflows/my-workflow.yml". It returns an empty string when the
// reference cannot be parsed.
func WorkflowFilePathFromRef(ref string) string {
	if ref == "" {
		return ""
	}
	// Strip the trailing "@<git-ref>" portion.
	if at := strings.Index(ref, "@"); at >= 0 {
		ref = ref[:at]
	}
	// The remaining value is "owner/repo/<path>"; drop the first two segments.
	parts := strings.SplitN(ref, "/", 3)
	if len(parts) < 3 {
		return ""
	}
	return parts[2]
}

// GetWorkflowFilePath returns the repository-relative path of the running
// workflow file, derived from GITHUB_WORKFLOW_REF.
func GetWorkflowFilePath() string {
	return WorkflowFilePathFromRef(GetWorkflowRef())
}
