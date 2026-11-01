package gh

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v93/github"
)

// TeamRepoPermissionChange describes a repository permission to grant.
type TeamRepoPermissionChange struct {
	Repository *github.Repository
	Permission string
}

// TeamRepoPermissionPlan contains direct grants, revocations, and warnings for a team.
type TeamRepoPermissionPlan struct {
	Grant    []TeamRepoPermissionChange
	Revoke   []*github.Repository
	Warnings []string
}

// PlanTeamRepoPermissionChanges calculates direct permission grants and revocations.
func PlanTeamRepoPermissionChanges(source, destinationDirect, destinationEffective []*github.Repository, force, sync bool) TeamRepoPermissionPlan {
	plan := TeamRepoPermissionPlan{}
	sourceByName := repositoryMap(source)
	directByName := repositoryMap(destinationDirect)
	effectiveByName := repositoryMap(destinationEffective)

	for _, repo := range source {
		if repo == nil || repo.GetName() == "" {
			continue
		}
		name := repo.GetName()
		key := repositoryKey(repo)
		permission := GetRepositoryPermissions(repo)
		directRepo := directByName[key]
		if directRepo != nil {
			if GetRepositoryPermissions(directRepo) == permission {
				continue
			}
			if !force && !sync {
				plan.Warnings = append(plan.Warnings, fmt.Sprintf("repository %s has a different direct permission; skipped", name))
				continue
			}
			plan.Grant = append(plan.Grant, TeamRepoPermissionChange{Repository: repo, Permission: permission})
			continue
		}
		if effectiveRepo := effectiveByName[key]; effectiveRepo != nil && GetRepositoryPermissions(effectiveRepo) == permission && !sync {
			continue
		}
		if effectiveRepo := effectiveByName[key]; effectiveRepo != nil && !force && !sync {
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("repository %s has a different effective permission; skipped", name))
			continue
		}
		plan.Grant = append(plan.Grant, TeamRepoPermissionChange{Repository: repo, Permission: permission})
	}

	if sync {
		for _, repo := range destinationDirect {
			if repo != nil && repo.GetName() != "" {
				if _, exists := sourceByName[repositoryKey(repo)]; !exists {
					plan.Revoke = append(plan.Revoke, repo)
				}
			}
		}
		for _, repo := range destinationEffective {
			if repo == nil || repo.GetName() == "" {
				continue
			}
			name := repo.GetName()
			key := repositoryKey(repo)
			if directRepo := directByName[key]; directRepo != nil && GetRepositoryPermissions(directRepo) == GetRepositoryPermissions(repo) {
				continue
			}
			sourceRepo := sourceByName[key]
			if sourceRepo == nil {
				plan.Warnings = append(plan.Warnings, fmt.Sprintf("repository %s remains accessible through an inherited permission", name))
			} else if GetRepositoryPermissions(repo) != GetRepositoryPermissions(sourceRepo) {
				plan.Warnings = append(plan.Warnings, fmt.Sprintf("repository %s has an inherited permission that may differ from the synchronized direct permission", name))
			}
		}
	}

	return plan
}

// CopyTeamRepoPermissions adds the source team's direct repository permissions to the destination team.
func CopyTeamRepoPermissions(ctx context.Context, srcClient *GitHubClient, src repository.Repository, srcTeamSlug string, dstClient *GitHubClient, dst repository.Repository, dstTeamSlug string, force bool) ([]string, error) {
	return syncTeamRepoPermissions(ctx, srcClient, src, srcTeamSlug, dstClient, dst, dstTeamSlug, force, false)
}

// SyncTeamRepoPermissions makes the destination team's direct repository permissions match the source team's.
func SyncTeamRepoPermissions(ctx context.Context, srcClient *GitHubClient, src repository.Repository, srcTeamSlug string, dstClient *GitHubClient, dst repository.Repository, dstTeamSlug string) ([]string, error) {
	return syncTeamRepoPermissions(ctx, srcClient, src, srcTeamSlug, dstClient, dst, dstTeamSlug, true, true)
}

func syncTeamRepoPermissions(ctx context.Context, srcClient *GitHubClient, src repository.Repository, srcTeamSlug string, dstClient *GitHubClient, dst repository.Repository, dstTeamSlug string, force, sync bool) ([]string, error) {
	if srcTeamSlug == "" || dstTeamSlug == "" {
		return nil, fmt.Errorf("source and destination team slugs must not be empty")
	}
	source, err := ListTeamRepos(ctx, srcClient, src, srcTeamSlug, nil, false)
	if err != nil {
		return nil, fmt.Errorf("failed to list source team repositories: %w", err)
	}
	destinationDirect, err := ListTeamRepos(ctx, dstClient, dst, dstTeamSlug, nil, false)
	if err != nil {
		return nil, fmt.Errorf("failed to list destination team's direct repositories: %w", err)
	}
	destinationEffective, err := ListTeamRepos(ctx, dstClient, dst, dstTeamSlug, nil, true)
	if err != nil {
		return nil, fmt.Errorf("failed to list destination team's repositories: %w", err)
	}

	plan := PlanTeamRepoPermissionChanges(source, destinationDirect, destinationEffective, force, sync)
	for _, change := range plan.Grant {
		repoName := change.Repository.GetName()
		if _, err := dstClient.GetRepository(ctx, dst.Owner, repoName); err != nil {
			var response *github.ErrorResponse
			if errors.As(err, &response) && response.Response != nil && response.Response.StatusCode == http.StatusNotFound {
				plan.Warnings = append(plan.Warnings, fmt.Sprintf("repository %s does not exist in destination organization %s; skipped", repoName, dst.Owner))
				continue
			}
			return plan.Warnings, fmt.Errorf("failed to check destination repository %s/%s: %w", dst.Owner, repoName, err)
		}
		if err := dstClient.AddTeamRepo(ctx, dst.Owner, dstTeamSlug, dst.Owner, repoName, change.Permission); err != nil {
			return plan.Warnings, fmt.Errorf("failed to grant team %s access to repository %s/%s: %w", dstTeamSlug, dst.Owner, repoName, err)
		}
	}
	for _, repo := range plan.Revoke {
		if err := dstClient.RemoveTeamRepo(ctx, dst.Owner, dstTeamSlug, dst.Owner, repo.GetName()); err != nil {
			return plan.Warnings, fmt.Errorf("failed to revoke team %s access to repository %s/%s: %w", dstTeamSlug, dst.Owner, repo.GetName(), err)
		}
	}
	return plan.Warnings, nil
}

func repositoryMap(repos []*github.Repository) map[string]*github.Repository {
	result := make(map[string]*github.Repository, len(repos))
	for _, repo := range repos {
		if repo != nil && repo.GetName() != "" {
			result[repositoryKey(repo)] = repo
		}
	}
	return result
}

func repositoryKey(repo *github.Repository) string {
	return strings.ToLower(repo.GetName())
}
