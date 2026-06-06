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
// HTTP 200 means an autofix already exists; HTTP 202 (github.AcceptedError) means
// a new autofix is being generated. Both cases are treated as success and the
// autofix status is returned. Callers can poll GetCodeScanningAutofix to track progress.
func (g *GitHubClient) CreateCodeScanningAutofix(ctx context.Context, owner, repo string, alertNumber int64) (*CodeScanningAutofix, error) {
	url := fmt.Sprintf("repos/%v/%v/code-scanning/alerts/%v/autofix", owner, repo, alertNumber)
	req, err := g.client.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	autofix := new(CodeScanningAutofix)
	_, err = g.client.Do(ctx, req, autofix)
	// HTTP 202: autofix is being generated asynchronously.
	// go-github surfaces this as AcceptedError with the raw response body.
	if err := handleAcceptedError(err, autofix); err != nil {
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
