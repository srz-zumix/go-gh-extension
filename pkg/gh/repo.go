package gh

import (
	"context"
	"fmt"
	"slices"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

type RepositorySubmodule = client.RepositorySubmodule

func GetRepository(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.Repository, error) {
	return g.GetRepository(ctx, repo.Owner, repo.Name)
}

func GetRepositoryByID(ctx context.Context, g *GitHubClient, id int64) (*github.Repository, error) {
	return g.GetRepositoryByID(ctx, id)
}

func ListBranches(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Branch, error) {
	return g.ListBranches(ctx, repo.Owner, repo.Name, nil)
}

func ListTags(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.RepositoryTag, error) {
	return g.ListTags(ctx, repo.Owner, repo.Name)
}

func ListProtectedBranches(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Branch, error) {
	protected := true
	return g.ListBranches(ctx, repo.Owner, repo.Name, &protected)
}

// CheckTeamPermissions is a wrapper function to check team permissions for a repository.
func CheckTeamPermissions(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string) (*github.Repository, bool, error) {
	if teamSlug == "" {
		return nil, false, nil
	}
	teamRepo, err := g.CheckTeamPermissions(ctx, repo.Owner, teamSlug, repo.Owner, repo.Name)
	if err != nil {
		return nil, false, err
	}
	hasPermission := teamRepo != nil
	if teamRepo == nil {
		teamRepo, err = GetRepository(ctx, g, repo)
		if err != nil {
			return nil, false, fmt.Errorf("failed to get repository: %w", err)
		}
		teamRepo.RoleName = github.Ptr("none")
		teamRepo.Permissions = CreatePermissionMap([]string{})
	}
	return teamRepo, hasPermission, nil
}

func CheckTeamPermissionsWithSubmodules(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string) ([]*github.Repository, bool, error) {
	var teamRepos []*github.Repository
	hasPermissions := true
	teamRepo, hasPermission, err := CheckTeamPermissions(ctx, g, repo, teamSlug)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check team permissions: %w", err)
	}
	if !hasPermission {
		hasPermissions = false
	}
	teamRepos = append(teamRepos, teamRepo)

	submodules, err := GetRepositorySubmodules(ctx, g, repo, true)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get submodules: %w", err)
	}
	submodules = FlattenRepositorySubmodules(submodules)
	for _, submodule := range submodules {
		submoduleTeamRepo, hasPermission, err := CheckTeamPermissions(ctx, g, submodule.Repository, teamSlug)
		if err != nil {
			return nil, false, fmt.Errorf("failed to check team permissions for submodule '%s': %w", submodule, err)
		}
		if !hasPermission {
			hasPermissions = false
		}
		teamRepos = append(teamRepos, submoduleTeamRepo)
	}

	return teamRepos, hasPermissions, nil
}

func ListTeamRepos(ctx context.Context, g *GitHubClient, repo repository.Repository, teamName string, roles []string, inherit bool) ([]*github.Repository, error) {
	if teamName == "" {
		return nil, nil
	}
	repos, err := g.ListTeamRepos(ctx, repo.Owner, teamName)
	if err != nil {
		return nil, err
	}

	if !inherit {
		var noInheritRepos []*github.Repository
		team, err := g.GetTeamBySlug(ctx, repo.Owner, teamName)
		if err != nil {
			return nil, err
		}
		if team != nil && team.Parent != nil {
			parentRepos, err := g.ListTeamRepos(ctx, repo.Owner, *team.Parent.Slug)
			if err != nil {
				return nil, err
			}
			for _, repo := range repos {
				d := FindRepository(repo, parentRepos)
				if CompareRepository(repo, d) != nil {
					noInheritRepos = append(noInheritRepos, repo)
				} else {
					teams, err := g.ListRepositoryTeams(ctx, *repo.Owner.Login, *repo.Name)
					if err != nil {
						return nil, err
					}
					if slices.ContainsFunc(teams, func(t *github.Team) bool {
						return *t.Slug == teamName
					}) {
						noInheritRepos = append(noInheritRepos, repo)
					}
				}
			}
			repos = noInheritRepos
		}
	}

	return FilterByRepositoryPermissions(repos, roles), nil
}

func ListUserAccessableRepositories(ctx context.Context, g *GitHubClient, repo repository.Repository, username string, roles []string, opt *RespositorySearchOptions) ([]*github.Repository, error) {
	loginUser, err := g.GetUser(ctx, "")
	if err != nil {
		return nil, err
	}
	if username == "" || username == "@me" {
		username = *loginUser.Login
	}

	repos, err := g.ListOrganizationRepositories(ctx, repo.Owner, opt.GetFilterString())
	if err != nil {
		return nil, err
	}

	repos = opt.Filter(repos)

	var filteredRepos []*github.Repository
	for _, r := range repos {
		permissions, err := g.GetRepositoryPermission(ctx, *r.Owner.Login, *r.Name, username)
		if err != nil {
			return nil, err
		}
		if permissions == nil {
			if username == *loginUser.Login {
				r.RoleName = github.Ptr("pull")
				r.Permissions = CreatePermissionMap([]string{"pull"})
			}
		} else {
			r.RoleName = permissions.RoleName
			r.Permissions = permissions.User.Permissions
		}

		if len(roles) == 0 || HasPermission(r, roles) {
			filteredRepos = append(filteredRepos, r)
		}
	}
	return filteredRepos, nil
}

// ListRepositoryCollaborators retrieves all collaborators for a specific repository.
func ListRepositoryCollaborators(ctx context.Context, g *GitHubClient, repo repository.Repository, affiliations []string, roles []string) ([]*github.User, error) {
	collaborators, err := g.ListRepositoryCollaborators(ctx, repo.Owner, repo.Name, GetCollaboratorAffiliationsFilter(affiliations))
	if err != nil {
		return nil, fmt.Errorf("failed to list collaborators for repository %s/%s: %w", repo.Owner, repo.Name, err)
	}

	return FilterByUserPermissions(collaborators, roles), nil
}

// AddRepositoryCollaborator is a wrapper function to add a collaborator to a repository.
func AddRepositoryCollaborator(ctx context.Context, g *GitHubClient, repo repository.Repository, username string, permission string) (*github.CollaboratorInvitation, error) {
	return g.AddRepositoryCollaborator(ctx, repo.Owner, repo.Name, username, permission)
}

// RemoveRepositoryCollaborator removes a collaborator from a specific repository.
func RemoveRepositoryCollaborator(ctx context.Context, g *GitHubClient, repo repository.Repository, username string) error {
	return g.RemoveRepositoryCollaborator(ctx, repo.Owner, repo.Name, username)
}

type RepositoryPermissionLevel struct {
	PermissionLevel *github.RepositoryPermissionLevel
	Repository      repository.Repository
}

func (p *RepositoryPermissionLevel) GetPermission() string {
	if p.PermissionLevel == nil {
		return "none"
	}
	return *p.PermissionLevel.Permission
}

// GetRepositoryPermission retrieves the permission level of a user for a specific repository.
func GetRepositoryPermission(ctx context.Context, g *GitHubClient, repo repository.Repository, username string) (*RepositoryPermissionLevel, error) {
	permissionLevel, err := g.GetRepositoryPermission(ctx, repo.Owner, repo.Name, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository permission for user '%s' in repository '%s/%s': %w", username, repo.Owner, repo.Name, err)
	}
	return &RepositoryPermissionLevel{
		PermissionLevel: permissionLevel,
		Repository:      repo,
	}, nil
}

func CheckRepositoryPermissionWithSubmodules(ctx context.Context, g *GitHubClient, repo repository.Repository, username string) ([]*RepositoryPermissionLevel, bool, error) {
	var repoRermissions []*RepositoryPermissionLevel
	hasPermissions := true
	repoPermission, err := GetRepositoryPermission(ctx, g, repo, username)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get repository permission: %w", err)
	}
	repoRermissions = append(repoRermissions, repoPermission)
	if repoPermission.GetPermission() == "none" {
		hasPermissions = false
	}

	submodules, err := GetRepositorySubmodules(ctx, g, repo, true)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get submodules: %w", err)
	}
	submodules = FlattenRepositorySubmodules(submodules)
	for _, submodule := range submodules {
		submoduleRepoPermission, err := GetRepositoryPermission(ctx, g, submodule.Repository, username)
		if err != nil {
			return nil, false, fmt.Errorf("failed to get repository permission for submodule '%s': %w", submodule, err)
		}
		if submoduleRepoPermission.GetPermission() == "none" {
			hasPermissions = false
		}
		repoRermissions = append(repoRermissions, submoduleRepoPermission)
	}

	return repoRermissions, hasPermissions, nil
}

func GetRepositorySubmodules(ctx context.Context, g *GitHubClient, repo repository.Repository, recursive bool) ([]RepositorySubmodule, error) {
	allSubmodules, err := g.GetRepositorySubmodules(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get submodules for repository %s/%s: %w", repo.Owner, repo.Name, err)
	}
	if recursive {
		for i, submodule := range allSubmodules {
			if repo.Host == submodule.Repository.Host {
				allSubmodules[i].Submodules, err = GetRepositorySubmodules(ctx, g, submodule.Repository, recursive)
				if err != nil {
					return nil, fmt.Errorf("failed to get nested submodules for submodule %s: %w", submodule.Name, err)
				}
			}
		}
	}
	return allSubmodules, nil
}

func FlattenRepositorySubmodules(submodules []RepositorySubmodule) []RepositorySubmodule {
	var flattened []RepositorySubmodule
	for _, submodule := range submodules {
		flattened = append(flattened, submodule)
		flattened = append(flattened, FlattenRepositorySubmodules(submodule.Submodules)...)
	}
	return flattened
}

func GetRepositoryContent(ctx context.Context, g *GitHubClient, repo repository.Repository, path string, ref *string) ([]*github.RepositoryContent, error) {
	fileContent, dirContents, err := g.GetRepositoryContent(ctx, repo.Owner, repo.Name, path, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get content for repository %s/%s at path '%s': %w", repo.Owner, repo.Name, path, err)
	}
	var content []*github.RepositoryContent
	if fileContent != nil {
		content = append(content, fileContent)
	}
	if dirContents != nil {
		content = append(content, dirContents...)
	}
	return content, nil
}

func GetRepositoryFileContent(ctx context.Context, g *GitHubClient, repo repository.Repository, path string, ref *string) (*github.RepositoryContent, error) {
	fileContent, _, err := g.GetRepositoryContent(ctx, repo.Owner, repo.Name, path, ref)
	if err != nil {
		if ref != nil {
			return nil, fmt.Errorf("failed to get content for repository %s/%s at path '%s' and ref '%s': %w", repo.Owner, repo.Name, path, *ref, err)
		}
		return nil, fmt.Errorf("failed to get content for repository %s/%s at path '%s': %w", repo.Owner, repo.Name, path, err)
	}
	return fileContent, nil
}

// CreateRepositoryFile creates a new file in a repository (wrapper).
func CreateRepositoryFile(ctx context.Context, g *GitHubClient, repo repository.Repository, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, error) {
	return g.CreateFile(ctx, repo.Owner, repo.Name, path, opts)
}

// UpdateRepositoryFile updates an existing file in a repository (wrapper).
func UpdateRepositoryFile(ctx context.Context, g *GitHubClient, repo repository.Repository, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContentResponse, error) {
	return g.UpdateFile(ctx, repo.Owner, repo.Name, path, opts)
}

// DeleteRepositoryFile deletes a file in a repository (wrapper).
func DeleteRepositoryFile(ctx context.Context, g *GitHubClient, repo repository.Repository, path string, opts *github.RepositoryContentFileOptions) error {
	return g.DeleteFile(ctx, repo.Owner, repo.Name, path, opts)
}
