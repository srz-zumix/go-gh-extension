package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// Codespaces organization access visibilities.
const (
	CodespacesOrgAccessDisabled                          = "disabled"
	CodespacesOrgAccessSelectedMembers                   = "selected_members"
	CodespacesOrgAccessAllMembers                        = "all_members"
	CodespacesOrgAccessAllMembersAndOutsideCollaborators = "all_members_and_outside_collaborators"
)

// GetCodespacesOrgAccess gets the Codespaces access setting of the organization that owns repo (wrapper).
func GetCodespacesOrgAccess(ctx context.Context, g *GitHubClient, repo repository.Repository) (*client.CodespacesOrgAccess, error) {
	return g.GetCodespacesOrgAccess(ctx, repo.Owner)
}
