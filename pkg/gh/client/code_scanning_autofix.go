package client

import (
	"context"
	"fmt"
)

// CodeScanningAutofix represents the status of a code scanning autofix.
type CodeScanningAutofix struct {
	Status      string  `json:"status"`
	Description *string `json:"description,omitempty"`
	StartedAt   *string `json:"started_at,omitempty"`
}

// CodeScanningAutofixCommitOptions specifies parameters for committing an autofix.
type CodeScanningAutofixCommitOptions struct {
	TargetRef *string `json:"target_ref,omitempty"`
	Message   *string `json:"message,omitempty"`
}

// CodeScanningAutofixCommit represents the result of committing an autofix.
type CodeScanningAutofixCommit struct {
	TargetRef string `json:"target_ref"`
	SHA       string `json:"sha"`
}

// GetCodeScanningAutofix gets the status and description of an autofix for a code scanning alert.
func (g *GitHubClient) GetCodeScanningAutofix(ctx context.Context, owner, repo string, alertNumber int64) (*CodeScanningAutofix, error) {
	url := fmt.Sprintf("repos/%v/%v/code-scanning/alerts/%v/autofix", owner, repo, alertNumber)
	req, err := g.client.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	autofix := new(CodeScanningAutofix)
	_, err = g.client.Do(ctx, req, autofix)
	if err != nil {
		return nil, err
	}
	return autofix, nil
}

// CreateCodeScanningAutofix creates an autofix for a code scanning alert.
// Returns the autofix status (200 if already exists, 202 if being generated).
func (g *GitHubClient) CreateCodeScanningAutofix(ctx context.Context, owner, repo string, alertNumber int64) (*CodeScanningAutofix, error) {
	url := fmt.Sprintf("repos/%v/%v/code-scanning/alerts/%v/autofix", owner, repo, alertNumber)
	req, err := g.client.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	autofix := new(CodeScanningAutofix)
	_, err = g.client.Do(ctx, req, autofix)
	if err != nil {
		return nil, err
	}
	return autofix, nil
}

// CommitCodeScanningAutofix commits an autofix for a code scanning alert.
func (g *GitHubClient) CommitCodeScanningAutofix(ctx context.Context, owner, repo string, alertNumber int64, opts *CodeScanningAutofixCommitOptions) (*CodeScanningAutofixCommit, error) {
	url := fmt.Sprintf("repos/%v/%v/code-scanning/alerts/%v/autofix/commits", owner, repo, alertNumber)
	req, err := g.client.NewRequest("POST", url, opts)
	if err != nil {
		return nil, err
	}
	commit := new(CodeScanningAutofixCommit)
	_, err = g.client.Do(ctx, req, commit)
	if err != nil {
		return nil, err
	}
	return commit, nil
}
