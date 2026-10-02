package client

// GitHub Codespaces organization API functions
// See: https://docs.github.com/rest/codespaces/organizations

import (
	"context"
)

// CodespacesOrgAccess is the Codespaces access setting of an organization.
type CodespacesOrgAccess struct {
	// Visibility is one of "disabled", "selected_members", "all_members", or
	// "all_members_and_outside_collaborators".
	Visibility string `json:"visibility"`
}

// GetCodespacesOrgAccess gets which users can use Codespaces on the private repositories of an organization.
func (g *GitHubClient) GetCodespacesOrgAccess(ctx context.Context, org string) (*CodespacesOrgAccess, error) {
	u := "orgs/" + org + "/codespaces/access"
	req, err := g.client.NewRequest(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}

	result := new(CodespacesOrgAccess)
	_, err = g.client.Do(req, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
