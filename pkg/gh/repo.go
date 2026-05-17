package gh

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

type RepositorySubmodule = client.RepositorySubmodule

// GetRepositoryFromGitHubRepository converts github.Repository to repository.Repository.
func GetRepositoryFromGitHubRepository(repo any) (repository.Repository, error) {
	switch v := repo.(type) {
	case github.Repository:
		host := ""
		htmlURL := v.GetHTMLURL()
		if htmlURL != "" {
			parsedURL, err := url.Parse(htmlURL)
			if err != nil {
				return repository.Repository{}, fmt.Errorf("failed to parse repository HTML URL %q: %w", htmlURL, err)
			}
			host = parsedURL.Host
		}
		if host == "" {
			host, _ = auth.DefaultHost()
		}

		owner := v.GetOwner()
		if owner == nil {
			return repository.Repository{}, fmt.Errorf("repository owner is nil")
		}
		ownerLogin := owner.GetLogin()
		repoName := v.GetName()
		if ownerLogin == "" || repoName == "" {
			return repository.Repository{}, fmt.Errorf("repository owner login or name is empty")
		}
		return repository.Repository{Host: host, Owner: ownerLogin, Name: repoName}, nil
	case *github.Repository:
		if v == nil {
			return repository.Repository{}, fmt.Errorf("repository is nil")
		}
		return GetRepositoryFromGitHubRepository(*v)
	default:
		return repository.Repository{}, fmt.Errorf("unsupported repository type %T", repo)
	}
}

func GetRepositoryID(repoID any) (int, error) {
	switch v := repoID.(type) {
	case int:
		return v, nil
	case *int:
		if v == nil {
			return 0, fmt.Errorf("repository ID pointer is nil")
		}
		return *v, nil
	case int64:
		return int(v), nil
	case *int64:
		if v == nil {
			return 0, fmt.Errorf("repository ID pointer is nil")
		}
		return int(*v), nil
	case github.Repository:
		if v.ID == nil {
			return 0, fmt.Errorf("repository ID is nil")
		}
		return int(v.GetID()), nil
	case *github.Repository:
		if v == nil {
			return 0, fmt.Errorf("repository is nil")
		}
		if v.ID == nil {
			return 0, fmt.Errorf("repository ID is nil")
		}
		return int(v.GetID()), nil
	default:
		return 0, fmt.Errorf("unsupported repository type %T", repoID)
	}
}

// GetRepository retrieves a repository by owner and name (wrapper).
func GetRepository(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.Repository, error) {
	return g.GetRepository(ctx, repo.Owner, repo.Name)
}

func GetRepositoryByID(ctx context.Context, g *GitHubClient, id int64) (*github.Repository, error) {
	return g.GetRepositoryByID(ctx, id)
}

// GetBranch retrieves a branch by name (wrapper).
func GetBranch(ctx context.Context, g *GitHubClient, repo repository.Repository, branch string) (*github.Branch, error) {
	return g.GetBranch(ctx, repo.Owner, repo.Name, branch)
}

// CreateBranch creates a new branch from the given SHA (wrapper).
func CreateBranch(ctx context.Context, g *GitHubClient, repo repository.Repository, branchName string, sha string) (*github.Reference, error) {
	return g.CreateRef(ctx, repo.Owner, repo.Name, "refs/heads/"+branchName, sha)
}

// DeleteBranch deletes a branch (wrapper).
func DeleteBranch(ctx context.Context, g *GitHubClient, repo repository.Repository, branchName string) error {
	return g.DeleteRef(ctx, repo.Owner, repo.Name, "heads/"+branchName)
}

// DeleteBranchIfExists deletes a branch, silently skipping if it does not exist.
// It ignores 404 Not Found and specific 422 Unprocessable Entity responses that
// indicate the reference does not exist. Invalid or empty branch names return an error.
func DeleteBranchIfExists(ctx context.Context, g *GitHubClient, repo repository.Repository, branchName string) error {
	if branchName == "" {
		return fmt.Errorf("branch name must not be empty")
	}

	err := DeleteBranch(ctx, g, repo, branchName)
	if err == nil {
		return nil
	}

	if IsHTTPNotFound(err) {
		// Branch does not exist
		return nil
	}

	if IsHTTPUnprocessableEntity(err) {
		// Some "reference does not exist" errors are returned as 422; only
		// treat those specific cases as "branch does not exist".
		var ghErr *github.ErrorResponse
		if errors.As(err, &ghErr) {
			if strings.Contains(ghErr.Message, "Reference does not exist") {
				return nil
			}
		}
	}

	return err
}

// ListBranches retrieves all branches for a repository (wrapper).
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

// ListOrganizationRepositories lists all repositories in an organization (wrapper).
func ListOrganizationRepositories(ctx context.Context, g *GitHubClient, org string) ([]*github.Repository, error) {
	return g.ListOrganizationRepositories(ctx, org, "all")
}

// ListUserRepositories lists all repositories owned by a user (wrapper).
func ListUserRepositories(ctx context.Context, g *GitHubClient, user string) ([]*github.Repository, error) {
	return g.ListUserRepositories(ctx, user, "all")
}

// ListOwnerRepositories lists all repositories for the given owner.
// It first attempts to list by organization; if the owner is not an organization (HTTP 404),
// it falls back to listing by user.
func ListOwnerRepositories(ctx context.Context, g *GitHubClient, owner string) ([]*github.Repository, error) {
	repos, err := g.ListOrganizationRepositories(ctx, owner, "all")
	if err != nil {
		var errResp *github.ErrorResponse
		if errors.As(err, &errResp) && errResp.Response != nil && errResp.Response.StatusCode == 404 {
			// Owner is not an organization; fall back to user repositories
			return g.ListUserRepositories(ctx, owner, "all")
		}
		return nil, err
	}
	return repos, nil
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
		teamRepo.Permissions = CreateRepositoryPermissions([]string{})
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
				r.Permissions = CreateRepositoryPermissions([]string{"pull"})
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

func repoKey(g *GitHubClient, repo repository.Repository) string {
	host := repo.Host
	if host == "" && g != nil {
		host = g.Host()
	}

	// Normalize repository coordinates so circular detection is case-insensitive
	// and treats implicit and explicit hosts as the same repository.
	return strings.ToLower(host) + "/" + strings.ToLower(repo.Owner) + "/" + strings.ToLower(repo.Name)
}

func GetRepositorySubmodules(ctx context.Context, g *GitHubClient, repo repository.Repository, recursive bool) ([]RepositorySubmodule, error) {
	visited := map[string]struct{}{
		repoKey(g, repo): {},
	}
	return getRepositorySubmodulesInternal(ctx, g, repo, recursive, visited)
}

func getRepositorySubmodulesInternal(ctx context.Context, g *GitHubClient, repo repository.Repository, recursive bool, visited map[string]struct{}) ([]RepositorySubmodule, error) {
	logger.Debug("getting submodules", "repo", repo.Owner+"/"+repo.Name, "recursive", recursive)
	allSubmodules, err := g.GetRepositorySubmodules(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get submodules for repository %s/%s: %w", repo.Owner, repo.Name, err)
	}
	logger.Debug("found submodules", "repo", repo.Owner+"/"+repo.Name, "count", len(allSubmodules))
	if recursive {
		for i, submodule := range allSubmodules {
			if repo.Host == submodule.Repository.Host {
				key := repoKey(g, submodule.Repository)
				if _, seen := visited[key]; seen {
					logger.Warn("circular submodule reference detected, skipping", "submodule", submodule.Name, "repo", submodule.Repository.Owner+"/"+submodule.Repository.Name)
					continue
				}
				logger.Debug("recursing into submodule", "submodule", submodule.Name, "repo", submodule.Repository.Owner+"/"+submodule.Repository.Name)
				visited[key] = struct{}{}
				allSubmodules[i].Submodules, err = getRepositorySubmodulesInternal(ctx, g, submodule.Repository, recursive, visited)
				delete(visited, key)
				if err != nil {
					return nil, fmt.Errorf("failed to get nested submodules for submodule %s: %w", submodule.Name, err)
				}
			} else {
				logger.Debug("skipping submodule on different host", "submodule", submodule.Name, "host", submodule.Repository.Host)
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

// GetFileContent retrieves the decoded content of a file from the repository
func GetFileContent(ctx context.Context, g *GitHubClient, repo repository.Repository, path string, ref *string) ([]byte, error) {
	fileContent, err := GetRepositoryFileContent(ctx, g, repo, path, ref)
	if err != nil {
		return nil, err
	}
	if fileContent == nil || fileContent.Content == nil {
		return nil, fmt.Errorf("file content is empty for %s", path)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode file content for %s: %w", path, err)
	}
	return []byte(content), nil
}

// RepositoryContentFileOptions holds options for creating, updating, or deleting a repository file.
// Message and Content are required.
// SHA is required for update and delete operations.
// Branch, Author, and Committer are optional.
type RepositoryContentFileOptions struct {
	Message   string        // required
	Content   []byte        // required
	Branch    *string       // optional: defaults to the repository's default branch
	SHA       *string       // optional: required for update/delete
	Author    *CommitAuthor // optional
	Committer *CommitAuthor // optional
}

// toGitHubRepositoryContentFileOptions converts RepositoryContentFileOptions to github.RepositoryContentFileOptions.
func toGitHubRepositoryContentFileOptions(opts *RepositoryContentFileOptions) *github.RepositoryContentFileOptions {
	if opts == nil {
		return nil
	}
	return &github.RepositoryContentFileOptions{
		Message:   &opts.Message,
		Content:   opts.Content,
		Branch:    opts.Branch,
		SHA:       opts.SHA,
		Author:    toGitHubCommitAuthor(opts.Author),
		Committer: toGitHubCommitAuthor(opts.Committer),
	}
}

// CreateRepositoryFile creates a new file in a repository (wrapper).
func CreateRepositoryFile(ctx context.Context, g *GitHubClient, repo repository.Repository, path string, opts *RepositoryContentFileOptions) (*github.RepositoryContentResponse, error) {
	return g.CreateFile(ctx, repo.Owner, repo.Name, path, toGitHubRepositoryContentFileOptions(opts))
}

// UpdateRepositoryFile updates an existing file in a repository (wrapper).
func UpdateRepositoryFile(ctx context.Context, g *GitHubClient, repo repository.Repository, path string, opts *RepositoryContentFileOptions) (*github.RepositoryContentResponse, error) {
	return g.UpdateFile(ctx, repo.Owner, repo.Name, path, toGitHubRepositoryContentFileOptions(opts))
}

// DeleteRepositoryFile deletes a file in a repository (wrapper).
func DeleteRepositoryFile(ctx context.Context, g *GitHubClient, repo repository.Repository, path string, opts *RepositoryContentFileOptions) error {
	return g.DeleteFile(ctx, repo.Owner, repo.Name, path, toGitHubRepositoryContentFileOptions(opts))
}

// RepoWithSecrets holds a repository and its associated secrets.
type RepoWithSecrets struct {
	Repository *github.Repository
	Secrets    []*github.Secret
	EnvSecrets map[string][]*github.Secret // environment name -> secrets
}

// SecretCount returns the number of repository-level secrets.
func (r *RepoWithSecrets) SecretCount() int {
	return len(r.Secrets)
}

// EnvSecretCount returns the total number of environment secrets across all environments.
func (r *RepoWithSecrets) EnvSecretCount() int {
	total := 0
	for _, secrets := range r.EnvSecrets {
		total += len(secrets)
	}
	return total
}

// TotalSecretCount returns the total number of secrets (repository + environment).
func (r *RepoWithSecrets) TotalSecretCount() int {
	return r.SecretCount() + r.EnvSecretCount()
}

// HasAnySecrets returns true if the repository has any secrets (repository or environment level).
func (r *RepoWithSecrets) HasAnySecrets() bool {
	return r.TotalSecretCount() > 0
}

// EditRepository updates a repository with the given settings (wrapper).
func EditRepository(ctx context.Context, g *GitHubClient, repo repository.Repository, repoUpdate *github.Repository) (*github.Repository, error) {
	return g.EditRepository(ctx, repo.Owner, repo.Name, repoUpdate)
}

// RenameBranch renames a branch in a repository (wrapper).
func RenameBranch(ctx context.Context, g *GitHubClient, repo repository.Repository, oldName, newName string) (*github.Branch, error) {
	return g.RenameBranch(ctx, repo.Owner, repo.Name, oldName, newName)
}

// UnarchiveRepository unarchives a repository (wrapper).
func UnarchiveRepository(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.Repository, error) {
	return EditRepository(ctx, g, repo, &github.Repository{Archived: github.Ptr(false)})
}

// ArchiveRepository archives a repository (wrapper).
func ArchiveRepository(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.Repository, error) {
	return EditRepository(ctx, g, repo, &github.Repository{Archived: github.Ptr(true)})
}

// EnableDiscussions enables Discussions on a repository.
func EnableDiscussions(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.Repository, error) {
	return EditRepository(ctx, g, repo, &github.Repository{HasDiscussions: github.Ptr(true)})
}

// GetRepositoryNodeID retrieves the GraphQL node ID of a repository.
func GetRepositoryNodeID(ctx context.Context, g *GitHubClient, repo repository.Repository) (string, error) {
	id, err := g.GetRepositoryNodeID(ctx, repo.Owner, repo.Name)
	if err != nil {
		return "", err
	}
	// githubv4.ID is an interface{}; the underlying value is expected to be a string.
	if s, ok := id.(string); ok {
		return s, nil
	}
	return "", fmt.Errorf("unexpected repository node ID type %T", id)
}
