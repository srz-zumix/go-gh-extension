package client

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v73/github"
	"github.com/shurcooL/githubv4"
)

func (g *GitHubClient) GetRepository(ctx context.Context, owner string, repo string) (*github.Repository, error) {
	repository, _, err := g.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return repository, nil
}

// ListRepositoryTeams retrieves all teams associated with a specific repository.
func (g *GitHubClient) ListRepositoryTeams(ctx context.Context, owner string, repo string) ([]*github.Team, error) {
	var allTeams []*github.Team
	opt := &github.ListOptions{PerPage: 50}

	for {
		teams, resp, err := g.client.Repositories.ListTeams(ctx, owner, repo, opt)
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

// ListOrganizationRepositories retrieves all repositories for a specific organization.
func (g *GitHubClient) ListOrganizationRepositories(ctx context.Context, org string, repoType string) ([]*github.Repository, error) {
	var allRepos []*github.Repository
	opt := &github.RepositoryListByOrgOptions{Type: repoType, ListOptions: github.ListOptions{PerPage: 50}}

	for {
		repos, resp, err := g.client.Repositories.ListByOrg(ctx, org, opt)
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

func (g *GitHubClient) CheckRepositoryCollaborators(ctx context.Context, owner string, repo string, username string) (bool, error) {
	collaborator, _, err := g.client.Repositories.IsCollaborator(ctx, owner, repo, username)
	if err != nil {
		return false, err
	}
	return collaborator, nil
}

// GetRepositoryPermission retrieves the permission level of a user for a specific repository.
func (g *GitHubClient) GetRepositoryPermission(ctx context.Context, owner string, repo string, username string) (*github.RepositoryPermissionLevel, error) {
	permission, resp, err := g.client.Repositories.GetPermissionLevel(ctx, owner, repo, username)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil, nil // User not found or no permission
		}
		return nil, err
	}
	return permission, nil
}

// ListRepositoryCollaborators retrieves all collaborators for a specific repository.
func (g *GitHubClient) ListRepositoryCollaborators(ctx context.Context, owner string, repo string, affiliation string) ([]*github.User, error) {
	var allCollaborators []*github.User
	opt := &github.ListCollaboratorsOptions{
		Affiliation: affiliation,
		ListOptions: github.ListOptions{
			PerPage: 50,
		},
	}

	for {
		collaborators, resp, err := g.client.Repositories.ListCollaborators(ctx, owner, repo, opt)
		if err != nil {
			return nil, err
		}
		allCollaborators = append(allCollaborators, collaborators...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allCollaborators, nil
}

// RemoveRepositoryCollaborator removes a collaborator from a specific repository.
func (g *GitHubClient) RemoveRepositoryCollaborator(ctx context.Context, owner string, repo string, username string) error {
	_, err := g.client.Repositories.RemoveCollaborator(ctx, owner, repo, username)
	return err
}

// AddRepositoryCollaborator adds a collaborator to a specific repository with a given permission.
func (g *GitHubClient) AddRepositoryCollaborator(ctx context.Context, owner string, repo string, username string, permission string) (*github.CollaboratorInvitation, error) {
	invitation, _, err := g.client.Repositories.AddCollaborator(ctx, owner, repo, username, &github.RepositoryAddCollaboratorOptions{Permission: permission})
	if err != nil {
		return nil, err
	}
	return invitation, nil
}

type RepositorySubmodule struct {
	Name       string
	GitUrl     string
	Repository repository.Repository
	Branch     string
	Path       string
	Submodules []RepositorySubmodule
}

type repositorySubmoduleObject struct {
	Name   githubv4.String
	GitUrl githubv4.String
	Branch githubv4.String
	Path   githubv4.String
}

func convertSubmodules(nodes []repositorySubmoduleObject) []RepositorySubmodule {
	submodules := make([]RepositorySubmodule, len(nodes))
	for i, node := range nodes {

		repo, err := repository.Parse(string(node.GitUrl))
		if err != nil {
			continue // Handle error appropriately, e.g., log it or return an error
		}
		submodules[i] = RepositorySubmodule{
			Name:       string(node.Name),
			GitUrl:     string(node.GitUrl),
			Repository: repo,
			Branch:     string(node.Branch),
			Path:       string(node.Path),
		}
	}
	return submodules
}

// GetRepositorySubmodules retrieves all submodules for a specific repository using GraphQL.
func (g *GitHubClient) GetRepositorySubmodules(ctx context.Context, owner string, repo string) ([]RepositorySubmodule, error) {
	var query struct {
		Repository struct {
			Submodules struct {
				Nodes    []repositorySubmoduleObject
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"submodules(first: 100, after: $cursor)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}
	variables := map[string]interface{}{
		"owner":  githubv4.String(owner),
		"name":   githubv4.String(repo),
		"cursor": (*githubv4.String)(nil), // Null after argument to get first page.
	}
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	allSubmodules := []RepositorySubmodule{}
	for {
		err := graphql.Query(ctx, &query, variables)
		if err != nil {
			return nil, err
		}
		allSubmodules = append(allSubmodules, convertSubmodules(query.Repository.Submodules.Nodes)...)
		if !query.Repository.Submodules.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Repository.Submodules.PageInfo.EndCursor)
	}
	return allSubmodules, nil
}

// GetRepositoryContent retrieves the content of a file or directory in a repository.
func (g *GitHubClient) GetRepositoryContent(ctx context.Context, owner, repo, path string, ref *string) (*github.RepositoryContent, []*github.RepositoryContent, error) {
	opt := &github.RepositoryContentGetOptions{}
	if ref != nil && *ref != "" {
		opt.Ref = *ref
	}
	fileContent, dirContent, _, err := g.client.Repositories.GetContents(ctx, owner, repo, path, opt)
	if err != nil {
		return nil, nil, err
	}
	return fileContent, dirContent, nil
}
