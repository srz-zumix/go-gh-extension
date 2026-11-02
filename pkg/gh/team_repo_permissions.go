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
	plan, err := PlanCopyTeamRepoPermissions(ctx, srcClient, src, srcTeamSlug, dstClient, dst, dstTeamSlug, force)
	if err != nil {
		if plan == nil {
			return nil, err
		}
		warnings := plan.Warnings()
		if plan.Sync {
			warnings = append([]string(nil), plan.skippedWarnings...)
		}
		if applyErr := applyTeamRepoPermissionGrants(ctx, dstClient, dst, dstTeamSlug, plan.Plan.Grant); applyErr != nil {
			return warnings, applyErr
		}
		return warnings, err
	}
	return ApplyTeamRepoAccessPlan(ctx, dstClient, dst, dstTeamSlug, plan)
}

// SyncTeamRepoPermissions makes the destination team's direct repository permissions match the source team's.
func SyncTeamRepoPermissions(ctx context.Context, srcClient *GitHubClient, src repository.Repository, srcTeamSlug string, dstClient *GitHubClient, dst repository.Repository, dstTeamSlug string) ([]string, error) {
	plan, err := PlanSyncTeamRepoPermissions(ctx, srcClient, src, srcTeamSlug, dstClient, dst, dstTeamSlug)
	if err != nil {
		if plan == nil {
			return nil, err
		}
		warnings := append([]string(nil), plan.skippedWarnings...)
		if applyErr := applyTeamRepoPermissionGrants(ctx, dstClient, dst, dstTeamSlug, plan.Plan.Grant); applyErr != nil {
			return warnings, applyErr
		}
		return warnings, err
	}
	return ApplyTeamRepoAccessPlan(ctx, dstClient, dst, dstTeamSlug, plan)
}

// TeamRepoAccessPlan contains the planned direct grants and revocations for a team.
type TeamRepoAccessPlan struct {
	Plan                 TeamRepoPermissionPlan
	Source               []*github.Repository
	DestinationDirect    []*github.Repository
	DestinationEffective []*github.Repository
	Skipped              map[string]struct{}
	Sync                 bool
	skippedWarnings      []string
	destination          repository.Repository
	destinationTeamSlug  string
}

// Changes returns the changes required to apply the team repository access plan.
func (p *TeamRepoAccessPlan) Changes() []SyncChange {
	if p == nil {
		return nil
	}
	directByName := repositoryMap(p.DestinationDirect)
	effectiveByName := repositoryMap(p.DestinationEffective)
	destination := syncChangeTeamDestination(p.destination, p.destinationTeamSlug)
	var changes []SyncChange
	for _, grant := range p.Plan.Grant {
		if grant.Repository == nil || grant.Repository.GetName() == "" {
			continue
		}
		key := repositoryKey(grant.Repository)
		if _, skipped := p.Skipped[key]; skipped {
			continue
		}
		current := ""
		action := SyncChangeAdd
		if direct := directByName[key]; direct != nil {
			current = GetRepositoryPermissions(direct)
			action = SyncChangeUpdate
		} else if effective := effectiveByName[key]; effective != nil {
			current = GetRepositoryPermissions(effective)
		}
		changes = append(changes, SyncChange{
			Destination: destination,
			Action:      action,
			Target:      grant.Repository.GetName(),
			Permission:  grant.Permission,
			Current:     current,
		})
	}
	for _, repo := range p.Plan.Revoke {
		if repo == nil || repo.GetName() == "" {
			continue
		}
		changes = append(changes, SyncChange{
			Destination: destination,
			Action:      SyncChangeRemove,
			Target:      repo.GetName(),
			Current:     GetRepositoryPermissions(repo),
		})
	}
	return changes
}

// Warnings returns warnings found while planning, including destination repositories that do not exist.
func (p *TeamRepoAccessPlan) Warnings() []string {
	if p == nil {
		return nil
	}
	warnings := append([]string(nil), p.Plan.Warnings...)
	return append(warnings, p.skippedWarnings...)
}

// PlanCopyTeamRepoPermissions calculates changes to add source team repositories to the destination team.
func PlanCopyTeamRepoPermissions(ctx context.Context, srcClient *GitHubClient, src repository.Repository, srcTeamSlug string, dstClient *GitHubClient, dst repository.Repository, dstTeamSlug string, force bool) (*TeamRepoAccessPlan, error) {
	return planTeamRepoPermissions(ctx, srcClient, src, srcTeamSlug, dstClient, dst, dstTeamSlug, force, false)
}

// PlanSyncTeamRepoPermissions calculates changes to match destination team repositories to the source team.
func PlanSyncTeamRepoPermissions(ctx context.Context, srcClient *GitHubClient, src repository.Repository, srcTeamSlug string, dstClient *GitHubClient, dst repository.Repository, dstTeamSlug string) (*TeamRepoAccessPlan, error) {
	return planTeamRepoPermissions(ctx, srcClient, src, srcTeamSlug, dstClient, dst, dstTeamSlug, true, true)
}

func planTeamRepoPermissions(ctx context.Context, srcClient *GitHubClient, src repository.Repository, srcTeamSlug string, dstClient *GitHubClient, dst repository.Repository, dstTeamSlug string, force, sync bool) (*TeamRepoAccessPlan, error) {
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

	plan := &TeamRepoAccessPlan{
		Plan:                 PlanTeamRepoPermissionChanges(source, destinationDirect, destinationEffective, force, sync),
		Source:               source,
		DestinationDirect:    destinationDirect,
		DestinationEffective: destinationEffective,
		Skipped:              make(map[string]struct{}),
		Sync:                 sync,
		destination:          dst,
		destinationTeamSlug:  dstTeamSlug,
	}
	grants := make([]TeamRepoPermissionChange, 0, len(plan.Plan.Grant))
	for _, change := range plan.Plan.Grant {
		repoName := change.Repository.GetName()
		if _, err := dstClient.GetRepository(ctx, dst.Owner, repoName); err != nil {
			var response *github.ErrorResponse
			if errors.As(err, &response) && response.Response != nil && response.Response.StatusCode == http.StatusNotFound {
				plan.Skipped[repositoryKey(change.Repository)] = struct{}{}
				plan.skippedWarnings = append(plan.skippedWarnings, fmt.Sprintf("repository %s does not exist in destination organization %s; skipped", repoName, dst.Owner))
				continue
			}
			plan.Plan.Grant = grants
			return plan, fmt.Errorf("failed to check destination repository %s/%s: %w", dst.Owner, repoName, err)
		}
		grants = append(grants, change)
	}
	plan.Plan.Grant = grants
	return plan, nil
}

// ApplyTeamRepoAccessPlan applies planned direct grants and revocations.
func ApplyTeamRepoAccessPlan(ctx context.Context, dstClient *GitHubClient, dst repository.Repository, dstTeamSlug string, plan *TeamRepoAccessPlan) ([]string, error) {
	if plan == nil {
		return nil, fmt.Errorf("team repository access plan must not be nil")
	}
	warnings := plan.Warnings()
	if plan.Sync {
		warnings = append([]string(nil), plan.skippedWarnings...)
	}
	if err := applyTeamRepoPermissionGrants(ctx, dstClient, dst, dstTeamSlug, plan.Plan.Grant); err != nil {
		return warnings, err
	}
	for _, repo := range plan.Plan.Revoke {
		if err := dstClient.RemoveTeamRepo(ctx, dst.Owner, dstTeamSlug, dst.Owner, repo.GetName()); err != nil {
			return warnings, fmt.Errorf("failed to revoke team %s access to repository %s/%s: %w", dstTeamSlug, dst.Owner, repo.GetName(), err)
		}
	}
	if plan.Sync {
		destinationAfterSync, err := ListTeamRepos(ctx, dstClient, dst, dstTeamSlug, nil, true)
		if err != nil {
			return warnings, fmt.Errorf("failed to verify destination team permissions after sync; changes may already have been applied: %w", err)
		}
		warnings = append(warnings, teamRepoPermissionWarnings(plan.Source, destinationAfterSync, plan.Skipped)...)
	}
	return warnings, nil
}

func applyTeamRepoPermissionGrants(ctx context.Context, dstClient *GitHubClient, dst repository.Repository, dstTeamSlug string, grants []TeamRepoPermissionChange) error {
	for _, grant := range grants {
		repoName := grant.Repository.GetName()
		if err := dstClient.AddTeamRepo(ctx, dst.Owner, dstTeamSlug, dst.Owner, repoName, grant.Permission); err != nil {
			return fmt.Errorf("failed to grant team %s access to repository %s/%s: %w", dstTeamSlug, dst.Owner, repoName, err)
		}
	}
	return nil
}

func teamRepoPermissionWarnings(source, destinationEffective []*github.Repository, skipped map[string]struct{}) []string {
	sourceByName := repositoryMap(source)
	destinationByName := repositoryMap(destinationEffective)
	warned := make(map[string]struct{})
	var warnings []string
	appendWarning := func(category, key, warning string) {
		warningKey := category + "\x00" + key
		if _, exists := warned[warningKey]; exists {
			return
		}
		warned[warningKey] = struct{}{}
		warnings = append(warnings, warning)
	}

	for _, repo := range destinationEffective {
		if repo == nil || repo.GetName() == "" {
			continue
		}
		key := repositoryKey(repo)
		sourceRepo := sourceByName[key]
		if sourceRepo == nil {
			appendWarning("extra", key, fmt.Sprintf("repository %s remains accessible after synchronization", repo.GetName()))
			continue
		}
		actual := GetRepositoryPermissions(repo)
		expected := GetRepositoryPermissions(sourceRepo)
		if actual != expected {
			appendWarning("permission", key, fmt.Sprintf("repository %s has effective permission %s after synchronization; expected %s", repo.GetName(), actual, expected))
		}
	}

	for _, repo := range source {
		if repo == nil || repo.GetName() == "" {
			continue
		}
		key := repositoryKey(repo)
		if destinationByName[key] != nil {
			continue
		}
		if _, wasSkipped := skipped[key]; wasSkipped {
			continue
		}
		appendWarning("missing", key, fmt.Sprintf("repository %s is missing from destination effective access after synchronization", repo.GetName()))
	}

	return warnings
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
