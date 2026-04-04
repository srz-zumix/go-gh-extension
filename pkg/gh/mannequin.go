package gh

import (
	"context"
	"errors"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/shurcooL/githubv4"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
	"github.com/srz-zumix/go-gh-extension/pkg/parser"
)

// ListMannequins retrieves all mannequins for the organization associated with the repository.
func ListMannequins(ctx context.Context, g *GitHubClient, repo repository.Repository, orderBy *githubv4.MannequinOrder) ([]client.Mannequin, error) {
	mannequins, err := g.ListMannequins(ctx, repo.Owner, orderBy)
	if err != nil {
		return nil, fmt.Errorf("failed to list mannequins for organization '%s': %w", repo.Owner, err)
	}
	return mannequins, nil
}

// FindMannequinByLogin retrieves a single mannequin by login from the organization associated with the repository.
// Returns nil if not found.
func FindMannequinByLogin(ctx context.Context, g *GitHubClient, repo repository.Repository, login string) (*client.Mannequin, error) {
	m, err := g.FindMannequinByLogin(ctx, repo.Owner, login)
	if err != nil {
		return nil, fmt.Errorf("failed to find mannequin '%s' in organization '%s': %w", login, repo.Owner, err)
	}
	return m, nil
}

// FindMannequinInListByEmail searches an already-fetched mannequin slice for the given email.
// Returns nil if email is empty or no match is found.
// Use this when the mannequin list has already been fetched to avoid redundant API calls.
func FindMannequinInListByEmail(mannequins []client.Mannequin, email string) *client.Mannequin {
	normalized := parser.NormalizeEmail(email)
	if normalized == "" {
		return nil
	}
	for i := range mannequins {
		m := &mannequins[i]
		if m.Email != nil && parser.NormalizeEmail(string(*m.Email)) == normalized {
			return m
		}
	}
	return nil
}

// FindMannequinByEmail finds a mannequin by email in the organization.
// Returns nil if not found. Returns nil immediately if email is empty.
// If the mannequin list has already been fetched, prefer FindMannequinInListByEmail
// to avoid a redundant API request.
func FindMannequinByEmail(ctx context.Context, g *GitHubClient, repo repository.Repository, email string) (*client.Mannequin, error) {
	if parser.NormalizeEmail(email) == "" {
		return nil, nil
	}
	mannequins, err := ListMannequins(ctx, g, repo, nil)
	if err != nil {
		return nil, err
	}
	return FindMannequinInListByEmail(mannequins, email), nil
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
		if errors.Is(err, client.ErrMutationUnavailable) {
			return fmt.Errorf("reattributeMannequinToUser is not enabled for organization '%s': contact GitHub Support to enable this feature: %w", repo.Owner, err)
		}
		return fmt.Errorf("failed to reattribute mannequin to user in organization '%s': %w", repo.Owner, err)
	}
	return nil
}
