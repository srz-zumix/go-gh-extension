package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/shurcooL/githubv4"
)

// RepositorySubmodule represents a git submodule in a repository.
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

// resolveGitURL resolves a potentially relative git URL against a base URL.
// If gitURL starts with "./" or "../", it is resolved relative to baseURL.
func resolveGitURL(baseURL, gitURL string) (string, error) {
	if !strings.HasPrefix(gitURL, "./") && !strings.HasPrefix(gitURL, "../") {
		return gitURL, nil
	}
	base, err := url.Parse(baseURL + "/")
	if err != nil {
		return "", err
	}
	rel, err := url.Parse(gitURL)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(rel).String(), nil
}

func convertSubmodule(submodule repositorySubmoduleObject, baseURL string) (*RepositorySubmodule, error) {
	gitURL, err := resolveGitURL(baseURL, string(submodule.GitUrl))
	if err != nil {
		return nil, err
	}
	repo, err := repository.Parse(gitURL)
	if err != nil {
		return nil, err
	}
	return &RepositorySubmodule{
		Name:       string(submodule.Name),
		GitUrl:     string(submodule.GitUrl),
		Repository: repo,
		Branch:     string(submodule.Branch),
		Path:       string(submodule.Path),
	}, nil
}

func convertSubmodules(nodes []repositorySubmoduleObject, baseURL string) []RepositorySubmodule {
	submodules := make([]RepositorySubmodule, len(nodes))
	for i, node := range nodes {
		submodule, err := convertSubmodule(node, baseURL)
		if err != nil {
			continue
		}
		submodules[i] = *submodule
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
	variables := map[string]any{
		"owner":  githubv4.String(owner),
		"name":   githubv4.String(repo),
		"cursor": (*githubv4.String)(nil), // Null after argument to get first page.
	}
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	baseURL := fmt.Sprintf("https://%s/%s/%s", g.Host(), owner, repo)
	allSubmodules := []RepositorySubmodule{}
	for {
		err := graphql.Query(ctx, &query, variables)
		if err != nil {
			return nil, err
		}
		allSubmodules = append(allSubmodules, convertSubmodules(query.Repository.Submodules.Nodes, baseURL)...)
		if !query.Repository.Submodules.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Repository.Submodules.PageInfo.EndCursor)
	}
	return allSubmodules, nil
}
