package actions

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
)

func GetErrorPrefix() string {
	return "::error::"
}

func Output(name, value string) (err error) {
	outputPath := os.Getenv("GITHUB_OUTPUT")
	if outputPath == "" {
		return nil
	}
	outputFile, err := os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open GITHUB_OUTPUT file: %w", err)
	}
	defer func() {
		err = outputFile.Close()
	}()
	// If value contains newline, use multiline format
	if strings.Contains(value, "\n") {
		delimiter := randomDelimiter()
		_, err = fmt.Fprintf(outputFile, "%s<<%s\n%s\n%s\n", name, delimiter, value, delimiter)
		return err
	} else {
		_, err = fmt.Fprintf(outputFile, "%s=%s\n", name, value)
	}
	return err
}

func randomDelimiter() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
