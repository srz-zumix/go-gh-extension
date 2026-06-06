package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// CommitCodeScanningAutofixOptions holds options for committing a code scanning autofix.
type CommitCodeScanningAutofixOptions struct {
	TargetRef string
	Message   string
}

// toClientCommitOptions converts CommitCodeScanningAutofixOptions to client.CodeScanningAutofixCommitOptions.
func toClientCommitOptions(opts *CommitCodeScanningAutofixOptions) *client.CodeScanningAutofixCommitOptions {
	if opts == nil {
		return nil
	}
	return &client.CodeScanningAutofixCommitOptions{
		TargetRef: stringPtrIfSet(opts.TargetRef),
		Message:   stringPtrIfSet(opts.Message),
	}
}

// GetCodeScanningAutofix gets the autofix status for a code scanning alert.
func GetCodeScanningAutofix(ctx context.Context, g *GitHubClient, repo repository.Repository, alertNumber int64) (*client.CodeScanningAutofix, error) {
	autofix, err := g.GetCodeScanningAutofix(ctx, repo.Owner, repo.Name, alertNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get autofix for code scanning alert #%d in %s/%s: %w", alertNumber, repo.Owner, repo.Name, err)
	}
	return autofix, nil
}

// CreateCodeScanningAutofix creates an autofix for a code scanning alert.
func CreateCodeScanningAutofix(ctx context.Context, g *GitHubClient, repo repository.Repository, alertNumber int64) (*client.CodeScanningAutofix, error) {
	autofix, err := g.CreateCodeScanningAutofix(ctx, repo.Owner, repo.Name, alertNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to create autofix for code scanning alert #%d in %s/%s: %w", alertNumber, repo.Owner, repo.Name, err)
	}
	return autofix, nil
}

// CommitCodeScanningAutofix commits an autofix for a code scanning alert.
func CommitCodeScanningAutofix(ctx context.Context, g *GitHubClient, repo repository.Repository, alertNumber int64, opts *CommitCodeScanningAutofixOptions) (*client.CodeScanningAutofixCommit, error) {
	commit, err := g.CommitCodeScanningAutofix(ctx, repo.Owner, repo.Name, alertNumber, toClientCommitOptions(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to commit autofix for code scanning alert #%d in %s/%s: %w", alertNumber, repo.Owner, repo.Name, err)
	}
	return commit, nil
}
