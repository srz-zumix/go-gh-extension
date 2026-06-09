package client

import (
	"context"
)

// CodeQualitySetup represents the code quality setup configuration for a repository.
type CodeQualitySetup struct {
	State       string   `json:"state"`
	Languages   []string `json:"languages"`
	RunnerType  string   `json:"runner_type"`
	RunnerLabel *string  `json:"runner_label"`
	UpdatedAt   string   `json:"updated_at"`
	Schedule    string   `json:"schedule"`
}

// CodeQualitySetupUpdate represents the request body for updating a code quality setup configuration.
type CodeQualitySetupUpdate struct {
	State       string   `json:"state,omitempty"`
	RunnerType  string   `json:"runner_type,omitempty"`
	RunnerLabel *string  `json:"runner_label,omitempty"`
	Languages   []string `json:"languages,omitempty"`
}

// GetCodeQualitySetup gets the code quality setup configuration for a repository.
func (g *GitHubClient) GetCodeQualitySetup(ctx context.Context, owner, repo string) (*CodeQualitySetup, error) {
	u := "repos/" + owner + "/" + repo + "/code-quality/setup"
	req, err := g.client.NewRequest(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}

	result := new(CodeQualitySetup)
	_, err = g.client.Do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateCodeQualitySetup updates the code quality setup configuration for a repository.
func (g *GitHubClient) UpdateCodeQualitySetup(ctx context.Context, owner, repo string, update *CodeQualitySetupUpdate) error {
	u := "repos/" + owner + "/" + repo + "/code-quality/setup"
	req, err := g.client.NewRequest(ctx, "PATCH", u, update)
	if err != nil {
		return err
	}

	_, err = g.client.Do(req, nil)
	return err
}
