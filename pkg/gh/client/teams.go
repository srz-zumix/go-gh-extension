package client

import (
	"context"

	"github.com/google/go-github/v73/github"
	"github.com/shurcooL/githubv4"
)

// ListTeams retrieves all teams in the specified organization with pagination support.
func (g *GitHubClient) ListTeams(ctx context.Context, org string) ([]*github.Team, error) {
	var allTeams []*github.Team
	opt := &github.ListOptions{PerPage: 50}

	for {
		teams, resp, err := g.client.Teams.ListTeams(ctx, org, opt)
		if err != nil {
			return nil, err
		}
		allTeams = append(allTeams, teams...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allTeams, nil
}

// ListChildTeams retrieves all child teams of a specified team.
func (g *GitHubClient) ListChildTeams(ctx context.Context, org string, parentSlug string) ([]*github.Team, error) {
	var allChildTeams []*github.Team
	opt := &github.ListOptions{PerPage: 50}

	for {
		teams, resp, err := g.client.Teams.ListChildTeamsByParentSlug(ctx, org, parentSlug, opt)
		if err != nil {
			return nil, err
		}
		allChildTeams = append(allChildTeams, teams...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allChildTeams, nil
}

// GetTeamBySlug retrieves a team by its slug name.
func (g *GitHubClient) GetTeamBySlug(ctx context.Context, org string, teamSlug string) (*github.Team, error) {
	team, _, err := g.client.Teams.GetTeamBySlug(ctx, org, teamSlug)
	if err != nil {
		return nil, err
	}
	return team, nil
}

// FindTeamBySlug retrieves a team by its slug name.
// If the team is not found, it returns nil without an error.
func (g *GitHubClient) FindTeamBySlug(ctx context.Context, org string, teamSlug string) (*github.Team, error) {
	t, resp, err := g.client.Teams.GetTeamBySlug(ctx, org, teamSlug)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil, nil
		}
		return nil, err
	}
	return t, nil
}

// ListTeamRepos retrieves all repositories associated with a specific team in the organization.
func (g *GitHubClient) ListTeamRepos(ctx context.Context, org string, teamSlug string) ([]*github.Repository, error) {
	var allRepos []*github.Repository
	opt := &github.ListOptions{PerPage: 50}

	for {
		repos, resp, err := g.client.Teams.ListTeamReposBySlug(ctx, org, teamSlug, opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

// CheckTeamPermissions checks the permissions of a team for a specific repository.
func (g *GitHubClient) CheckTeamPermissions(ctx context.Context, org string, teamSlug string, owner string, repo string) (*github.Repository, error) {
	teamRepo, resp, err := g.client.Teams.IsTeamRepoBySlug(ctx, org, teamSlug, owner, repo)
	if err != nil {
		if resp.StatusCode == 404 {
			return nil, nil
		}
		return nil, err
	}
	return teamRepo, nil
}

// RemoveTeamRepo removes a repository from a team in the organization.
func (g *GitHubClient) RemoveTeamRepo(ctx context.Context, org string, teamSlug string, owner string, repoName string) error {
	_, err := g.client.Teams.RemoveTeamRepoBySlug(ctx, org, teamSlug, owner, repoName)
	if err != nil {
		return err
	}
	return nil
}

// AddTeamRepo adds a repository to a team in the organization.
func (g *GitHubClient) AddTeamRepo(ctx context.Context, org string, teamSlug string, owner string, repoName string, permission string) error {
	opt := &github.TeamAddTeamRepoOptions{
		Permission: permission,
	}
	_, err := g.client.Teams.AddTeamRepoBySlug(ctx, org, teamSlug, owner, repoName, opt)
	if err != nil {
		return err
	}
	return nil
}

// AddOrUpdateTeamMembership adds or updates the membership of a user in a team.
func (g *GitHubClient) AddOrUpdateTeamMembership(ctx context.Context, org string, teamSlug string, username string, role string) (*github.Membership, error) {
	opt := &github.TeamAddTeamMembershipOptions{
		Role: role,
	}
	membership, _, err := g.client.Teams.AddTeamMembershipBySlug(ctx, org, teamSlug, username, opt)
	if err != nil {
		return nil, err
	}
	return membership, nil
}

// RemoveTeamMember removes a user from a team in the organization.
func (g *GitHubClient) RemoveTeamMember(ctx context.Context, org string, teamSlug string, username string) error {
	_, err := g.client.Teams.RemoveTeamMembershipBySlug(ctx, org, teamSlug, username)
	if err != nil {
		return err
	}
	return nil
}

// ListTeamMembers retrieves all members of a specific team in the organization.
func (g *GitHubClient) ListTeamMembers(ctx context.Context, org string, teamSlug string, role string) ([]*github.User, error) {
	var allMembers []*github.User
	opt := &github.TeamListTeamMembersOptions{
		Role:        role,
		ListOptions: github.ListOptions{PerPage: 50},
	}

	for {
		members, resp, err := g.client.Teams.ListTeamMembersBySlug(ctx, org, teamSlug, opt)
		if err != nil {
			return nil, err
		}
		allMembers = append(allMembers, members...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allMembers, nil
}

// GetTeamMembership retrieves the membership details of a user in a specific team.
func (g *GitHubClient) GetTeamMembership(ctx context.Context, org string, teamSlug string, username string) (*github.Membership, error) {
	membership, _, err := g.client.Teams.GetTeamMembershipBySlug(ctx, org, teamSlug, username)
	if err != nil {
		return nil, err
	}
	return membership, nil
}

// FindTeamMembership retrieves the membership details of a user in a specific team.
// If the user is not a member, it returns nil without an error.
func (g *GitHubClient) FindTeamMembership(ctx context.Context, org string, teamSlug string, username string) (*github.Membership, error) {
	membership, resp, err := g.client.Teams.GetTeamMembershipBySlug(ctx, org, teamSlug, username)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil, nil // User is not a member
		}
		return nil, err
	}
	return membership, nil
}

// TeamCodeReviewSettings represents the code review settings for a team
type TeamCodeReviewSettings struct {
	TeamSlug                     string
	NotifyTeam                   bool
	Enabled                      bool
	Algorithm                    string
	TeamMemberCount              int
	ExcludedTeamMemberIDs        []string
	IncludeChildTeamMembers      bool
	CountMembersAlreadyRequested bool
	RemoveTeamRequest            bool
}

// GetTeamCodeReviewSettings retrieves the code review assignment settings for a team using GraphQL
func (g *GitHubClient) GetTeamCodeReviewSettings(ctx context.Context, org string, teamSlug string) (*TeamCodeReviewSettings, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}

	var query struct {
		Organization struct {
			Team struct {
				Slug                               githubv4.String
				ReviewRequestDelegationEnabled     githubv4.Boolean
				ReviewRequestDelegationAlgorithm   githubv4.String
				ReviewRequestDelegationMemberCount githubv4.Int
				ReviewRequestDelegationNotifyTeam  githubv4.Boolean
			} `graphql:"team(slug: $teamSlug)"`
		} `graphql:"organization(login: $org)"`
	}

	variables := map[string]interface{}{
		"org":      githubv4.String(org),
		"teamSlug": githubv4.String(teamSlug),
	}

	if err := graphql.Query(ctx, &query, variables); err != nil {
		return nil, err
	}

	settings := &TeamCodeReviewSettings{
		TeamSlug:        teamSlug,
		Enabled:         bool(query.Organization.Team.ReviewRequestDelegationEnabled),
		Algorithm:       string(query.Organization.Team.ReviewRequestDelegationAlgorithm),
		TeamMemberCount: int(query.Organization.Team.ReviewRequestDelegationMemberCount),
		NotifyTeam:      bool(query.Organization.Team.ReviewRequestDelegationNotifyTeam),
	}

	return settings, nil
}

// UpdateTeamReviewAssignmentInput represents the input for updating team review assignment settings
type UpdateTeamReviewAssignmentInput struct {
	ID                           githubv4.ID       `json:"id"`
	NotifyTeam                   *githubv4.Boolean `json:"notifyTeam,omitempty"`
	Enabled                      githubv4.Boolean  `json:"enabled"`
	Algorithm                    *githubv4.String  `json:"algorithm,omitempty"`
	TeamMemberCount              *githubv4.Int     `json:"teamMemberCount,omitempty"`
	ExcludedTeamMemberIDs        []githubv4.ID     `json:"excludedTeamMemberIds,omitempty"`
	IncludeChildTeamMembers      *githubv4.Boolean `json:"includeChildTeamMembers,omitempty"`
	CountMembersAlreadyRequested *githubv4.Boolean `json:"countMembersAlreadyRequested,omitempty"`
	RemoveTeamRequest            *githubv4.Boolean `json:"removeTeamRequest,omitempty"`
	ClientMutationID             *githubv4.String  `json:"clientMutationId,omitempty"`
}

// SetTeamCodeReviewSettings updates the code review assignment settings for a team using GraphQL mutation
func (g *GitHubClient) SetTeamCodeReviewSettings(ctx context.Context, teamID any, settings *TeamCodeReviewSettings) error {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}

	var mutation struct {
		UpdateTeamReviewAssignment struct {
			Team struct {
				ID                                 githubv4.ID
				ReviewRequestDelegationEnabled     githubv4.Boolean
				ReviewRequestDelegationAlgorithm   githubv4.String
				ReviewRequestDelegationMemberCount githubv4.Int
				ReviewRequestDelegationNotifyTeam  githubv4.Boolean
			}
			ClientMutationID githubv4.String
		} `graphql:"updateTeamReviewAssignment(input: $input)"`
	}

	input := UpdateTeamReviewAssignmentInput{
		ID: githubv4.ID(teamID),
	}

	input.Enabled = githubv4.Boolean(settings.Enabled)
	input.Algorithm = github.Ptr(githubv4.String(settings.Algorithm))
	input.TeamMemberCount = github.Ptr(githubv4.Int(settings.TeamMemberCount))
	input.NotifyTeam = github.Ptr(githubv4.Boolean(settings.NotifyTeam))
	input.ExcludedTeamMemberIDs = make([]githubv4.ID, len(settings.ExcludedTeamMemberIDs))
	for i, id := range settings.ExcludedTeamMemberIDs {
		input.ExcludedTeamMemberIDs[i] = githubv4.ID(id)
	}
	input.IncludeChildTeamMembers = github.Ptr(githubv4.Boolean(settings.IncludeChildTeamMembers))
	input.CountMembersAlreadyRequested = github.Ptr(githubv4.Boolean(settings.CountMembersAlreadyRequested))
	input.RemoveTeamRequest = github.Ptr(githubv4.Boolean(settings.RemoveTeamRequest))

	return graphql.Mutate(ctx, &mutation, input, nil)
}
