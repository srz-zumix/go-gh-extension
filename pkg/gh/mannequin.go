package gh

import (
	"context"
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/shurcooL/githubv4"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// ListMannequins retrieves all mannequins for the organization associated with the repository.
func ListMannequins(ctx context.Context, g *GitHubClient, repo repository.Repository, orderBy *githubv4.MannequinOrder) ([]client.Mannequin, error) {
	mannequins, err := g.ListMannequins(ctx, repo.Owner, orderBy)
	if err != nil {
		return nil, fmt.Errorf("failed to list mannequins for organization '%s': %w", repo.Owner, err)
	}
	return mannequins, nil
}

// CreateAttributionInvitation creates an attribution invitation to claim a mannequin in the organization.
func CreateAttributionInvitation(ctx context.Context, g *GitHubClient, repo repository.Repository, ownerID, sourceID, targetID string) error {
	err := g.CreateAttributionInvitation(ctx, ownerID, sourceID, targetID)
	if err != nil {
		return fmt.Errorf("failed to create attribution invitation in organization '%s': %w", repo.Owner, err)
	}
	return nil
}

// ReattributeMannequinToUser directly reclaims a mannequin to a user without sending an invitation.
// This mutation is undocumented and requires the feature to be enabled for the organization by GitHub Support.
func ReattributeMannequinToUser(ctx context.Context, g *GitHubClient, repo repository.Repository, ownerID, sourceID, targetID string) error {
	err := g.ReattributeMannequinToUser(ctx, ownerID, sourceID, targetID)
	if err != nil {
		if strings.Contains(err.Error(), "Field 'reattributeMannequinToUser' doesn't exist on type 'Mutation'") {
			return fmt.Errorf("reattributeMannequinToUser is not enabled for organization '%s': contact GitHub Support to enable this feature: %w", repo.Owner, err)
		}
		return fmt.Errorf("failed to reattribute mannequin to user in organization '%s': %w", repo.Owner, err)
	}
	return nil
}
