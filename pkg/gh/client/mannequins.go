package client

import (
	"context"

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
			ID    githubv4.ID
			Login githubv4.String
		}
		ClientMutationID githubv4.String
	} `graphql:"createAttributionInvitation(input: $input)"`
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

// ReattributeMannequinToUserInput is the input type for reattributeMannequinToUser mutation.
// This is an undocumented mutation; see https://github.com/github/gh-gei for reference.
type ReattributeMannequinToUserInput struct {
	OwnerID  githubv4.ID `json:"ownerId"`
	SourceID githubv4.ID `json:"sourceId"`
	TargetID githubv4.ID `json:"targetId"`
}

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
	} `graphql:"reattributeMannequinToUser(input: $input)"`
}

// ReattributeMannequinToUser directly reclaims a mannequin to a user without sending an invitation.
// This mutation is undocumented and requires the feature to be enabled for the organization by GitHub Support.
func (g *GitHubClient) ReattributeMannequinToUser(ctx context.Context, ownerID, sourceID, targetID string) error {
	graphqlClient, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}

	var mutation reattributeMannequinToUserMutation

	input := ReattributeMannequinToUserInput{
		OwnerID:  githubv4.ID(ownerID),
		SourceID: githubv4.ID(sourceID),
		TargetID: githubv4.ID(targetID),
	}

	return graphqlClient.Mutate(ctx, &mutation, input, nil)
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
