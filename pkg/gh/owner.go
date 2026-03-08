package gh

import (
	"context"
	"fmt"
)

// OwnerType represents the type of a GitHub owner.
type OwnerType string

const (
	OwnerTypeOrg  OwnerType = "Organization"
	OwnerTypeUser OwnerType = "User"
)

// DetectOwnerType determines whether the given owner is an organization or a user.
func DetectOwnerType(ctx context.Context, g *GitHubClient, owner string) (OwnerType, error) {
	user, err := g.GetUser(ctx, owner)
	if err != nil {
		return "", fmt.Errorf("failed to detect owner type for '%s': %w", owner, err)
	}
	if user.Type != nil && *user.Type == string(OwnerTypeOrg) {
		return OwnerTypeOrg, nil
	}
	return OwnerTypeUser, nil
}
