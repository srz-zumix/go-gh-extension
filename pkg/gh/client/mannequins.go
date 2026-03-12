package client

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
)

// Mannequin represents a placeholder account for unclaimed users in GitHub.
type Mannequin struct {
	ID         githubv4.ID
	Login      githubv4.String
	Email      *githubv4.String
	AvatarURL  githubv4.String `graphql:"avatarUrl"`
	URL        githubv4.String `graphql:"url"`
	DatabaseID *githubv4.Int   `graphql:"databaseId"`
	CreatedAt  githubv4.DateTime
	Claimant   MannequinClaimant
}

// MannequinClaimant represents the user who has claimed the identity of a mannequin.
type MannequinClaimant struct {
	ID    githubv4.ID
	Login githubv4.String
	URL   githubv4.String `graphql:"url"`
}

// NodeID returns the node ID of the mannequin as a string.
func (m *Mannequin) NodeID() string {
	return string(m.ID)
}

type mannequinConnectionResult struct {
	TotalCount githubv4.Int
	Nodes      []Mannequin
	PageInfo   struct {
		EndCursor   githubv4.String
		HasNextPage bool
	}
}

type mannequinsQuery struct {
	Organization struct {
		Mannequins mannequinConnectionResult `graphql:"mannequins(first: $first, after: $cursor, orderBy: $orderBy)"`
	} `graphql:"organization(login: $org)"`
}

type mannequinsQueryDefault struct {
	Organization struct {
		Mannequins mannequinConnectionResult `graphql:"mannequins(first: $first, after: $cursor)"`
	} `graphql:"organization(login: $org)"`
}

type mannequinByLoginQuery struct {
	Organization struct {
		Mannequins struct {
			Nodes []Mannequin
		} `graphql:"mannequins(login: $login, first: 1)"`
	} `graphql:"organization(login: $org)"`
}

type createAttributionInvitationMutation struct {
	CreateAttributionInvitation struct {
		Owner struct {
			ID    githubv4.ID
			Login githubv4.String
		}
		Source struct {
			Mannequin struct {
				ID    githubv4.ID
				Login githubv4.String
			} `graphql:"... on Mannequin"`
		}
		Target struct {
			User struct {
				ID    githubv4.ID
				Login githubv4.String
			} `graphql:"... on User"`
		}
		ClientMutationID githubv4.String
	} `graphql:"createAttributionInvitation(input: $input)"`
}

// FindMannequinByLogin retrieves a single mannequin by login for the given organization.
// Returns nil if not found.
func (g *GitHubClient) FindMannequinByLogin(ctx context.Context, org, login string) (*Mannequin, error) {
	graphqlClient, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}

	var query mannequinByLoginQuery
	variables := map[string]any{
		"org":   githubv4.String(org),
		"login": githubv4.String(login),
	}
	if err := graphqlClient.Query(ctx, &query, variables); err != nil {
		return nil, err
	}
	if len(query.Organization.Mannequins.Nodes) == 0 {
		return nil, nil
	}
	m := query.Organization.Mannequins.Nodes[0]
	return &m, nil
}

// ListMannequins retrieves all mannequins for an organization using GraphQL with cursor-based pagination.
func (g *GitHubClient) ListMannequins(ctx context.Context, org string, orderBy *githubv4.MannequinOrder) ([]Mannequin, error) {
	graphqlClient, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}

	var allMannequins []Mannequin

	if orderBy == nil {
		var query mannequinsQueryDefault
		variables := map[string]any{
			"org":    githubv4.String(org),
			"first":  githubv4.Int(100),
			"cursor": (*githubv4.String)(nil),
		}
		for {
			if err := graphqlClient.Query(ctx, &query, variables); err != nil {
				return nil, err
			}
			allMannequins = append(allMannequins, query.Organization.Mannequins.Nodes...)
			if !query.Organization.Mannequins.PageInfo.HasNextPage {
				break
			}
			variables["cursor"] = githubv4.NewString(query.Organization.Mannequins.PageInfo.EndCursor)
		}
	} else {
		var query mannequinsQuery
		variables := map[string]any{
			"org":     githubv4.String(org),
			"first":   githubv4.Int(100),
			"cursor":  (*githubv4.String)(nil),
			"orderBy": orderBy,
		}
		for {
			if err := graphqlClient.Query(ctx, &query, variables); err != nil {
				return nil, err
			}
			allMannequins = append(allMannequins, query.Organization.Mannequins.Nodes...)
			if !query.Organization.Mannequins.PageInfo.HasNextPage {
				break
			}
			variables["cursor"] = githubv4.NewString(query.Organization.Mannequins.PageInfo.EndCursor)
		}
	}

	return allMannequins, nil
}

// reattributeMannequinToUserMutation defines the mutation used to directly reattribute
// a mannequin to a user without sending an invitation.
// Individual ID variables are used because the server does not accept a typed
// input variable for this undocumented mutation.
type reattributeMannequinToUserMutation struct {
	ReattributeMannequinToUser struct {
		Source struct {
			Mannequin struct {
				ID    githubv4.ID
				Login githubv4.String
			} `graphql:"... on Mannequin"`
		}
		Target struct {
			User struct {
				ID    githubv4.ID
				Login githubv4.String
			} `graphql:"... on User"`
		}
	} `graphql:"reattributeMannequinToUser(input: {ownerId: $orgId, sourceId: $sourceId, targetId: $targetId})"`
}

// ReattributeMannequinToUser directly reclaims a mannequin to a user without sending an invitation.
// This mutation is undocumented and requires the feature to be enabled for the organization by GitHub Support.
func (g *GitHubClient) ReattributeMannequinToUser(ctx context.Context, ownerID, sourceID, targetID string) error {
	graphqlClient := g.newGraphQLClientWithFeatures("import_api,mannequin_claiming_emu,org_import_api")
	var mutation reattributeMannequinToUserMutation
	variables := map[string]interface{}{
		"orgId":    githubv4.ID(ownerID),
		"sourceId": githubv4.ID(sourceID),
		"targetId": githubv4.ID(targetID),
	}
	if err := graphqlClient.Mutate(ctx, &mutation, variables); err != nil {
		if isMutationFieldNotFoundError(err, "reattributeMannequinToUser") {
			return fmt.Errorf("%w: %v", ErrMutationUnavailable, err)
		}
		return err
	}
	return nil
}

// CreateAttributionInvitation creates an attribution invitation to claim a mannequin.
func (g *GitHubClient) CreateAttributionInvitation(ctx context.Context, ownerID, sourceID, targetID string) error {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}

	var mutation createAttributionInvitationMutation

	input := githubv4.CreateAttributionInvitationInput{
		OwnerID:  githubv4.ID(ownerID),
		SourceID: githubv4.ID(sourceID),
		TargetID: githubv4.ID(targetID),
	}

	return graphql.Mutate(ctx, &mutation, input, nil)
}
