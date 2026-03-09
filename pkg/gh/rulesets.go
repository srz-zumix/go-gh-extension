package gh

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

type RepositoryRulesetConfig struct {
	ID           *int64                              `json:"id,omitempty"`
	Name         string                              `json:"name"`
	Target       *string                             `json:"target,omitempty"`
	SourceType   *string                             `json:"source_type,omitempty"`
	Source       string                              `json:"source"`
	Enforcement  string                              `json:"enforcement"`
	Conditions   *github.RepositoryRulesetConditions `json:"conditions,omitempty"`
	Rules        *github.RepositoryRulesetRules      `json:"rules,omitempty"`
	BypassActors []*github.BypassActor               `json:"bypass_actors,omitempty"`
}

type RepositoryRulesetMigrateConfig struct {
	Ruleset      *github.RepositoryRuleset
	Teams        map[int64]*github.Team
	CheckRuns    map[int64]*CheckRun
	Repositories map[int64]*github.Repository
}

func ListRepositoryRulesets(ctx context.Context, g *GitHubClient, repo repository.Repository, includesParents bool) ([]*github.RepositoryRuleset, error) {
	return g.ListRepositoryRulesets(ctx, repo.Owner, repo.Name, includesParents)
}

func GetRepositoryRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, rulesetID int64, includesParents bool) (*github.RepositoryRuleset, error) {
	return g.GetRepositoryRuleset(ctx, repo.Owner, repo.Name, rulesetID, includesParents)
}

func CreateRepositoryRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleset *github.RepositoryRuleset) (*github.RepositoryRuleset, error) {
	return g.CreateRepositoryRuleset(ctx, repo.Owner, repo.Name, ruleset)
}

func UpdateRepositoryRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, rulesetID int64, ruleset *github.RepositoryRuleset) (*github.RepositoryRuleset, error) {
	return g.UpdateRepositoryRuleset(ctx, repo.Owner, repo.Name, rulesetID, ruleset)
}

func CreateOrUpdateRepositoryRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleset *github.RepositoryRuleset) (*github.RepositoryRuleset, error) {
	existingRuleset, err := FindRepositoryRuleset(ctx, g, repo, *ruleset.ID, ruleset.Name, false)
	if err != nil {
		return nil, err
	}
	if existingRuleset != nil {
		return UpdateRepositoryRuleset(ctx, g, repo, *existingRuleset.ID, ruleset)
	}
	return CreateRepositoryRuleset(ctx, g, repo, ruleset)
}

func DeleteRepositoryRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, rulesetID int64) error {
	return g.DeleteRepositoryRuleset(ctx, repo.Owner, repo.Name, rulesetID)
}

func FindRepositoryRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, rulesetID int64, name string, includesParents bool) (*github.RepositoryRuleset, error) {
	ruleset, err := g.GetRepositoryRuleset(ctx, repo.Owner, repo.Name, rulesetID, includesParents)
	if err == nil {
		return ruleset, nil
	}
	return FindRepositoryRulesetByName(ctx, g, repo, name, includesParents)
}

func FindRepositoryRulesetByName(ctx context.Context, g *GitHubClient, repo repository.Repository, name string, includesParents bool) (*github.RepositoryRuleset, error) {
	rulesets, err := ListRepositoryRulesets(ctx, g, repo, includesParents)
	if err != nil {
		return nil, err
	}
	for _, ruleset := range rulesets {
		if ruleset.Name == name {
			return ruleset, nil
		}
	}
	return nil, nil
}

// ListOrgRulesets retrieves all rulesets for a specific organization
func ListOrgRulesets(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.RepositoryRuleset, error) {
	return g.ListOrgRulesets(ctx, repo.Owner)
}

// GetOrgRuleset retrieves a single ruleset for a specific organization by ruleset ID
func GetOrgRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, rulesetID int64) (*github.RepositoryRuleset, error) {
	return g.GetOrgRuleset(ctx, repo.Owner, rulesetID)
}

// CreateOrgRuleset creates a new ruleset for a specific organization
func CreateOrgRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleset *github.RepositoryRuleset) (*github.RepositoryRuleset, error) {
	return g.CreateOrgRuleset(ctx, repo.Owner, ruleset)
}

// UpdateOrgRuleset updates an existing ruleset for a specific organization
func UpdateOrgRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, rulesetID int64, ruleset *github.RepositoryRuleset) (*github.RepositoryRuleset, error) {
	return g.UpdateOrgRuleset(ctx, repo.Owner, rulesetID, ruleset)
}

// CreateOrUpdateOrgRuleset creates or updates an organization ruleset
func CreateOrUpdateOrgRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleset *github.RepositoryRuleset) (*github.RepositoryRuleset, error) {
	existingRuleset, err := FindOrgRuleset(ctx, g, repo, *ruleset.ID, ruleset.Name)
	if err != nil {
		return nil, err
	}
	if existingRuleset != nil {
		return UpdateOrgRuleset(ctx, g, repo, *existingRuleset.ID, ruleset)
	}
	return CreateOrgRuleset(ctx, g, repo, ruleset)
}

// DeleteOrgRuleset deletes a single ruleset for a specific organization by ruleset ID
func DeleteOrgRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, rulesetID int64) error {
	return g.DeleteOrgRuleset(ctx, repo.Owner, rulesetID)
}

// FindOrgRuleset finds an organization ruleset by ID or name
func FindOrgRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, rulesetID int64, name string) (*github.RepositoryRuleset, error) {
	ruleset, err := g.GetOrgRuleset(ctx, repo.Owner, rulesetID)
	if err == nil {
		return ruleset, nil
	}
	return FindOrgRulesetByName(ctx, g, repo, name)
}

// FindOrgRulesetByName finds an organization ruleset by name
func FindOrgRulesetByName(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) (*github.RepositoryRuleset, error) {
	rulesets, err := ListOrgRulesets(ctx, g, repo)
	if err != nil {
		return nil, err
	}
	for _, ruleset := range rulesets {
		if ruleset.Name == name {
			return ruleset, nil
		}
	}
	return nil, nil
}

func ListRulesets(ctx context.Context, g *GitHubClient, repo repository.Repository, includesParents bool) ([]*github.RepositoryRuleset, error) {
	if repo.Name == "" {
		return ListOrgRulesets(ctx, g, repo)
	}
	return ListRepositoryRulesets(ctx, g, repo, includesParents)
}

func GetRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, rulesetID int64, includesParents bool) (*github.RepositoryRuleset, error) {
	if repo.Name == "" {
		return GetOrgRuleset(ctx, g, repo, rulesetID)
	}
	return GetRepositoryRuleset(ctx, g, repo, rulesetID, includesParents)
}

func CreateRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleset *github.RepositoryRuleset) (*github.RepositoryRuleset, error) {
	if repo.Name == "" {
		return CreateOrgRuleset(ctx, g, repo, ruleset)
	}
	return CreateRepositoryRuleset(ctx, g, repo, ruleset)
}

func UpdateRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, rulesetID int64, ruleset *github.RepositoryRuleset) (*github.RepositoryRuleset, error) {
	if repo.Name == "" {
		return UpdateOrgRuleset(ctx, g, repo, rulesetID, ruleset)
	}
	return UpdateRepositoryRuleset(ctx, g, repo, rulesetID, ruleset)
}

func CreateOrUpdateRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleset *github.RepositoryRuleset) (*github.RepositoryRuleset, error) {
	if repo.Name == "" {
		return CreateOrUpdateOrgRuleset(ctx, g, repo, ruleset)
	}
	return CreateOrUpdateRepositoryRuleset(ctx, g, repo, ruleset)
}

func DeleteRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, rulesetID int64) error {
	if repo.Name == "" {
		return DeleteOrgRuleset(ctx, g, repo, rulesetID)
	}
	return DeleteRepositoryRuleset(ctx, g, repo, rulesetID)
}

func FindRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, rulesetID int64, name string, includesParents bool) (*github.RepositoryRuleset, error) {
	if repo.Name == "" {
		return FindOrgRuleset(ctx, g, repo, rulesetID, name)
	}
	return FindRepositoryRuleset(ctx, g, repo, rulesetID, name, includesParents)
}

func FindRulesetByName(ctx context.Context, g *GitHubClient, repo repository.Repository, name string, includesParents bool) (*github.RepositoryRuleset, error) {
	if repo.Name == "" {
		return FindOrgRulesetByName(ctx, g, repo, name)
	}
	return FindRepositoryRulesetByName(ctx, g, repo, name, includesParents)
}

func ExportRuleset(ruleset *github.RepositoryRuleset) *RepositoryRulesetConfig {
	config := &RepositoryRulesetConfig{
		ID:           ruleset.ID,
		Name:         ruleset.Name,
		Target:       (*string)(ruleset.Target),
		SourceType:   (*string)(ruleset.SourceType),
		Source:       ruleset.Source,
		Enforcement:  (string)(ruleset.Enforcement),
		BypassActors: ruleset.BypassActors,
		Conditions:   ruleset.Conditions,
		Rules:        ruleset.Rules,
	}

	return config
}

func ImportRuleset(config *RepositoryRulesetConfig, ruleset *github.RepositoryRuleset) *github.RepositoryRuleset {
	if ruleset == nil {
		ruleset = &github.RepositoryRuleset{}
	}
	ruleset.Name = config.Name
	ruleset.Target = (*github.RulesetTarget)(config.Target)
	ruleset.SourceType = (*github.RulesetSourceType)(config.SourceType)
	ruleset.Source = config.Source
	ruleset.Enforcement = github.RulesetEnforcement(config.Enforcement)
	ruleset.BypassActors = config.BypassActors
	ruleset.Conditions = config.Conditions
	ruleset.Rules = config.Rules

	return ruleset
}

func LoadRepositoryRulesetConfigFromReader(r io.Reader) (*RepositoryRulesetConfig, error) {
	var config RepositoryRulesetConfig
	jsonData, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonData, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func LoadRepositoryRulesetConfig(path string) (*RepositoryRulesetConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close() // nolint
	return LoadRepositoryRulesetConfigFromReader(f)
}

func ExportMigrateRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, rulesetID int64) (*RepositoryRulesetMigrateConfig, error) {
	ruleset, err := GetRuleset(ctx, g, repo, rulesetID, false)
	if err != nil {
		return nil, err
	}

	teams := GetRulesetActorsTeams(ctx, g, repo, ruleset)

	checkRuns, err := exportRulesetCheckRuns(ctx, g, repo, ruleset)
	if err != nil {
		return nil, err
	}

	repositories := make(map[int64]*github.Repository)
	if ruleset.Conditions != nil && ruleset.Conditions.RepositoryID != nil {
		for _, id := range ruleset.Conditions.RepositoryID.RepositoryIDs {
			r, err := GetRepositoryByID(ctx, g, id)
			if err != nil {
				logger.Warn("Failed to get repository by ID", "id", id, "error", err)
			} else {
				repositories[id] = r
			}
		}
	}
	if ruleset.Rules != nil && ruleset.Rules.Workflows != nil {
		for _, workflow := range ruleset.Rules.Workflows.Workflows {
			r, err := GetRepositoryByID(ctx, g, workflow.GetRepositoryID())
			if err != nil {
				logger.Warn("Failed to get workflow repository by ID", "id", workflow.GetRepositoryID(), "error", err)
			} else {
				repositories[workflow.GetRepositoryID()] = r
			}
		}
	}

	return &RepositoryRulesetMigrateConfig{
		Ruleset:      ruleset,
		Teams:        teams,
		CheckRuns:    checkRuns,
		Repositories: repositories,
	}, nil
}

// exportRulesetCheckRuns collects check run app metadata for all required status checks in a ruleset.
// This information is used to remap integration IDs during migration to a different organization.
func exportRulesetCheckRuns(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleset *github.RepositoryRuleset) (map[int64]*CheckRun, error) {
	checkRuns := make(map[int64]*CheckRun)
	if ruleset.Rules == nil || ruleset.Rules.RequiredStatusChecks == nil {
		return checkRuns, nil
	}

	// Get the target repository for resolving check runs.
	// For org rulesets, this is a repository where checks are actually run.
	// For repo rulesets, this returns the repo itself.
	checkRunRepo, err := GetRulesetTargetRepository(ctx, g, repo, ruleset)
	if err != nil {
		return nil, err
	}
	if checkRunRepo == nil {
		logger.Warn("Ruleset target repository not found, falling back to source repository for check run resolution")
		checkRunRepo = &repo
	}
	if checkRunRepo.Name == "" {
		logger.Warn("No target repository available for resolving required status check integrations, skipping check run collection...")
		return checkRuns, nil
	}
	ref, err := FindRulesetRequireStatusCheckRunRef(ctx, g, *checkRunRepo, ruleset)
	if err != nil {
		ref = "HEAD"
	}
	for _, check := range ruleset.Rules.RequiredStatusChecks.RequiredStatusChecks {
		if check.IntegrationID == nil {
			continue
		}
		if _, ok := checkRuns[*check.IntegrationID]; ok {
			continue
		}
		checkRun, err := findIntegrationID(ctx, g, *checkRunRepo, ref, check.Context, check.IntegrationID, nil)
		if err != nil {
			return nil, err
		}
		if checkRun != nil {
			checkRuns[*check.IntegrationID] = checkRun
		}
	}
	return checkRuns, nil
}

var GitHubComGitHubActionsAppID int64 = 15368

func ImportMigrateRuleset(ctx context.Context, g *GitHubClient, repo repository.Repository, migrateConfig *RepositoryRulesetMigrateConfig, gitHubActionsAppID *int64) (*github.RepositoryRuleset, error) {
	ruleset := migrateConfig.Ruleset
	teams := GetRulesetActorsTeams(ctx, g, repo, ruleset)

	org, err := GetOrganizationProfile(ctx, g, repo)
	if err != nil {
		return nil, err
	}

	newBypassActors := []*github.BypassActor{}
	for _, actor := range ruleset.BypassActors {
		if *actor.ActorType == github.BypassActorTypeOrganizationAdmin {
			if org.IsUser() {
				logger.Warn("Bypass actor organization admin is not supported on user accounts, skipping...")
				continue
			}
		}
		if *actor.ActorType != github.BypassActorTypeTeam {
			newBypassActors = append(newBypassActors, actor)
			continue
		}
		if _, ok := teams[*actor.ActorID]; ok {
			newBypassActors = append(newBypassActors, actor)
			continue
		}
		if team, ok := migrateConfig.Teams[*actor.ActorID]; ok {
			t, err := g.GetTeamBySlug(ctx, repo.Owner, team.GetSlug())
			if err == nil {
				*actor.ActorID = t.GetID()
				logger.Info("Team ID has been converted to team with same name in migration destination", "team", t.GetSlug(), "id", t.GetID())
				newBypassActors = append(newBypassActors, actor)
			}
		}
		teamName := fmt.Sprintf("%d", *actor.ActorID)
		orgTeam := migrateConfig.Teams[*actor.ActorID]
		if orgTeam != nil {
			teamName = orgTeam.GetName()
		}
		logger.Warn("Bypass actor team not found in target repository, skipping...", "team", teamName)
	}
	ruleset.BypassActors = newBypassActors

	if org.IsUser() || org.IsGitHubEnterpriseServer() {
		if *ruleset.Target == github.RulesetTargetPush {
			logger.Warn("Push target rulesets are not supported on user accounts or GitHub Enterprise Server, skipping...")
			return nil, nil
		}
	}
	if !org.IsGitHubEnterprise() {
		if ruleset.Rules != nil {
			ruleset.Rules.MergeQueue = nil
			logger.Warn("Merge Queue settings are not supported on GitHub.com or GitHub Team plan, removing...")
			ruleset.Rules.CommitMessagePattern = nil
			ruleset.Rules.CommitAuthorEmailPattern = nil
			ruleset.Rules.CommitterEmailPattern = nil
			logger.Warn("Restrict commit metadata settings are not supported on GitHub.com or GitHub Team plan, removing...")
			ruleset.Rules.BranchNamePattern = nil
			ruleset.Rules.TagNamePattern = nil
			logger.Warn("Restrict branch and tag names settings are not supported on GitHub.com or GitHub Team plan, removing...")
		}
	}
	if org.IsGitHubEnterpriseServer() {
		if ruleset.Rules.PullRequest != nil {
			ruleset.Rules.PullRequest.AllowedMergeMethods = nil
			logger.Warn("Allowed merge methods are not supported on GitHub Enterprise Server, removing...")
			ruleset.Rules.PullRequest.AutomaticCopilotCodeReviewEnabled = nil
			logger.Warn("Automatic Copilot code review is not supported on GitHub Enterprise Server, removing...")
		}
	} else {
		if ruleset.Rules.PullRequest != nil {
			ruleset.Rules.PullRequest.AllowedMergeMethods = []github.PullRequestMergeMethod{
				github.PullRequestMergeMethodSquash,
				github.PullRequestMergeMethodRebase,
				github.PullRequestMergeMethodMerge,
			}
			logger.Info("Allowed merge methods have been set to all methods supported on GitHub.com")
		}
	}

	if err := importRulesetRequiredStatusChecks(ctx, g, repo, ruleset, migrateConfig, org, gitHubActionsAppID); err != nil {
		return nil, err
	}
	if err := importRulesetRequiredDeployments(ctx, g, repo, ruleset); err != nil {
		return nil, err
	}

	migrateRepositoryNames := map[int64]string{}
	migrateRepositories := map[int64]*github.Repository{}
	for id, r := range migrateConfig.Repositories {
		migrateRepositoryNames[id] = r.GetName()
		dstRepo, err := g.GetRepository(ctx, repo.Owner, r.GetName())
		if err != nil {
			logger.Warn("Repository ID condition target repository not found in target organization, skipping...", "name", r.GetName())
			continue
		}
		migrateRepositories[id] = dstRepo
	}
	importRulesetConditions(repo, ruleset, migrateRepositoryNames, migrateRepositories)
	importRulesetWorkflows(ruleset, migrateRepositories)

	result, err := CreateOrUpdateRuleset(ctx, g, repo, ruleset)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// importRulesetRequiredStatusChecks remaps required status checks integration IDs for the destination repository.
func importRulesetRequiredStatusChecks(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleset *github.RepositoryRuleset, migrateConfig *RepositoryRulesetMigrateConfig, org *OrganizationProfile, gitHubActionsAppID *int64) error {
	if ruleset == nil || ruleset.Rules == nil || ruleset.Rules.RequiredStatusChecks == nil {
		return nil
	}

	foundIntegrations := make(map[int64]*int64)

	// Get the target repository for resolving check runs in the destination.
	// For org rulesets, this is a repository where checks are actually run.
	// For repo rulesets, this returns the repo itself.
	checkRunRepo, err := GetRulesetTargetRepository(ctx, g, repo, ruleset)
	if err != nil {
		return err
	}
	if checkRunRepo == nil {
		checkRunRepo = &repo
	}
	if checkRunRepo.Name == "" {
		logger.Warn("No target repository available for resolving required status check integrations, falling back to any-source...")
		for _, check := range ruleset.Rules.RequiredStatusChecks.RequiredStatusChecks {
			if check.IntegrationID != nil {
				check.IntegrationID = nil
				logger.Warn("Required status check integration replaced to any-source due to missing target repository", "integration", check.Context)
			}
		}
		return nil
	}
	ref, err := FindRulesetRequireStatusCheckRunRef(ctx, g, *checkRunRepo, ruleset)
	if err != nil {
		ref = "HEAD"
	}

	for _, check := range ruleset.Rules.RequiredStatusChecks.RequiredStatusChecks {
		if check.IntegrationID == nil {
			continue
		}
		id := *check.IntegrationID
		if dstID, ok := foundIntegrations[id]; ok {
			check.IntegrationID = dstID
			continue
		}

		found, err := findIntegrationID(ctx, g, *checkRunRepo, ref, check.Context, check.IntegrationID, nil)
		if err == nil && found != nil {
			// cache successful no-op integration mapping to avoid duplicate lookups
			foundIntegrations[id] = check.IntegrationID
			continue
		}

		check.IntegrationID = nil
		checkRun := migrateConfig.CheckRuns[id]
		if checkRun != nil {
			found, err = findIntegrationID(ctx, g, *checkRunRepo, ref, check.Context, nil, checkRun.App)
			if err == nil {
				if found != nil {
					check.IntegrationID = found.App.ID
					foundIntegrations[id] = found.App.ID
					logger.Info("Mapped required status check integration to target repository", "integration", check.Context, "id", found.App.GetID())
					continue
				}

				if checkRun.App != nil && checkRun.App.GetSlug() == "github-actions" {
					actionAppID := gitHubActionsAppID
					if actionAppID == nil && org.IsGitHubCom() {
						actionAppID = &GitHubComGitHubActionsAppID
					}
					if actionAppID != nil {
						check.IntegrationID = actionAppID
						foundIntegrations[id] = actionAppID
						logger.Info("Mapped required status check integration to GitHub Actions in target repository", "integration", check.Context)
						continue
					}
				}
			}
		}
		logger.Warn("Required status check integration not found in target repository, replace to any-source", "integration", check.Context)
	}

	return nil
}

// importRulesetRequiredDeployments validates required deployment environments for repository rulesets.
// Organization rulesets do not support required deployments and this rule will be removed.
func importRulesetRequiredDeployments(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleset *github.RepositoryRuleset) error {
	if ruleset == nil || ruleset.Rules == nil || ruleset.Rules.RequiredDeployments == nil {
		return nil
	}

	if repo.Name == "" {
		ruleset.Rules.RequiredDeployments = nil
		logger.Warn("Required deployments rule is not supported for Organization rulesets, removing...")
		return nil
	}

	newEnvironments := []string{}
	for _, env := range ruleset.Rules.RequiredDeployments.RequiredDeploymentEnvironments {
		deployments, err := ListEnvrionmentDeployments(ctx, g, repo, env)
		if err != nil || len(deployments) == 0 {
			logger.Warn("Required deployment environment not found in target repository, skipping...", "environment", env)
			continue
		}
		newEnvironments = append(newEnvironments, env)
	}

	if len(newEnvironments) == 0 {
		ruleset.Rules.RequiredDeployments = nil
		logger.Warn("No valid required deployment environments found in target repository, removing required deployments rule...")
		return nil
	}

	ruleset.Rules.RequiredDeployments.RequiredDeploymentEnvironments = newEnvironments
	return nil
}

// importRulesetWorkflows remaps workflow repository IDs for the destination organization.
func importRulesetWorkflows(ruleset *github.RepositoryRuleset, migrateRepositories map[int64]*github.Repository) {
	if ruleset == nil || ruleset.Rules == nil || ruleset.Rules.Workflows == nil {
		return
	}

	newWorkflows := []*github.RuleWorkflow{}
	for _, workflow := range ruleset.Rules.Workflows.Workflows {
		r, ok := migrateRepositories[workflow.GetRepositoryID()]
		if !ok {
			logger.Warn("Workflow repository not found in target organization, skipping...", "id", workflow.GetRepositoryID())
			continue
		}
		workflow.RepositoryID = r.ID
		newWorkflows = append(newWorkflows, workflow)
		logger.Info("Workflow repository has been mapped to target repository", "name", r.GetName(), "id", r.GetID())
	}
	if len(newWorkflows) == 0 {
		ruleset.Rules.Workflows = nil
		logger.Warn("No valid workflows found in target repository, removing workflows rule...")
		return
	}

	ruleset.Rules.Workflows.Workflows = newWorkflows
}

// importRulesetConditions adjusts ruleset conditions to fit destination scope and maps repository IDs.
func importRulesetConditions(repo repository.Repository, ruleset *github.RepositoryRuleset, migrateRepositoryNames map[int64]string, migrateRepositories map[int64]*github.Repository) {
	if ruleset == nil || ruleset.Conditions == nil {
		return
	}

	// Repository rulesets do not support RepositoryName, RepositoryID, or RepositoryProperty conditions.
	// These conditions are only valid for Organization rulesets.
	if repo.Name != "" {
		if ruleset.Conditions.RepositoryName != nil {
			ruleset.Conditions.RepositoryName = nil
			logger.Warn("Repository name condition is not supported for Repository rulesets, removing...")
		}
		if ruleset.Conditions.RepositoryID != nil {
			ruleset.Conditions.RepositoryID = nil
			logger.Warn("Repository ID condition is not supported for Repository rulesets, removing...")
		}
		if ruleset.Conditions.RepositoryProperty != nil {
			ruleset.Conditions.RepositoryProperty = nil
			logger.Warn("Repository property condition is not supported for Repository rulesets, removing...")
		}
		return
	}

	if ruleset.Conditions.RepositoryID == nil {
		return
	}

	newRepoIDs := []int64{}
	newRepoNames := []string{}
	for _, id := range ruleset.Conditions.RepositoryID.RepositoryIDs {
		r, ok := migrateRepositories[id]
		if ok {
			newRepoIDs = append(newRepoIDs, r.GetID())
			newRepoNames = append(newRepoNames, r.GetName())
			logger.Info("Repository ID condition has been mapped to target repository", "name", r.GetName(), "id", r.GetID())
			continue
		}
		if name, nameOk := migrateRepositoryNames[id]; nameOk && name != "" {
			newRepoNames = append(newRepoNames, name)
			logger.Warn("Repository ID condition not found in target organization, will use repository name as fallback...", "id", id, "name", name)
		} else {
			logger.Warn("Repository ID condition not found in target organization and name unknown, skipping...", "id", id)
		}
	}
	if len(newRepoIDs) == 0 {
		// When no repository IDs could be mapped, try to fall back to repository names if available.
		ruleset.Conditions.RepositoryID = nil
		if len(newRepoNames) == 0 {
			// No repository names are available either; remove repository conditions entirely.
			ruleset.Conditions.RepositoryName = nil
			logger.Warn("No valid repository ID or repository name conditions found in target organization, removing repository conditions...")
			return
		}
		ruleset.Conditions.RepositoryName = &github.RepositoryRulesetRepositoryNamesConditionParameters{
			Include:   newRepoNames,
			Exclude:   []string{},
			Protected: nil,
		}
		logger.Warn("No valid repository ID conditions found in target organization, converting to repository name conditions...")
		return
	}

	ruleset.Conditions.RepositoryID.RepositoryIDs = newRepoIDs
}

func GetRulesetActorsTeams(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleset *github.RepositoryRuleset) map[int64]*github.Team {
	teams := make(map[int64]*github.Team)

	var allTeams []*github.Team
	for _, actor := range ruleset.BypassActors {
		if *actor.ActorType == github.BypassActorTypeTeam {
			if allTeams == nil {
				teamTree, err := TeamByOwner(ctx, g, repo, true)
				if err != nil {
					return teams
				}
				allTeams = teamTree.Flatten()
			}
			for _, t := range allTeams {
				if t.GetID() == *actor.ActorID {
					teams[*actor.ActorID] = t
					break
				}
			}
		}
	}

	return teams
}

func findIntegrationID(ctx context.Context, g *GitHubClient, repo repository.Repository, ref string, name string, appID *int64, app *github.App) (*CheckRun, error) {
	checkRuns, err := ListCheckRunsForRef(ctx, g, repo, ref, &ListChecksRunFilterOptions{
		AppID: appID,
	})
	if err != nil {
		return nil, err
	}
	if checkRuns.Total == 0 {
		return nil, nil
	}
	for i, checkRun := range checkRuns.CheckRuns {
		if checkRun.GetName() == name {
			return checkRuns.CheckRuns[i], nil
		}
		if app != nil && checkRun.App != nil {
			if checkRun.App.GetName() == app.GetName() {
				return checkRuns.CheckRuns[i], nil
			}
			if checkRun.App.GetSlug() == app.GetSlug() {
				return checkRuns.CheckRuns[i], nil
			}
		}
	}

	if appID == nil {
		return nil, nil
	}
	return checkRuns.CheckRuns[0], nil
}

func FindRulesetRequireStatusCheckRunRef(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleset *github.RepositoryRuleset) (string, error) {
	refs, err := FindRulesetTargetRefs(ctx, g, repo, ruleset)
	if err != nil {
		return "HEAD", err
	}
	if ruleset.Target == nil {
		return refs[0], nil
	}
	if *ruleset.Target == github.RulesetTargetBranch {
		for _, ref := range refs {
			pr, err := FindPullRequest(ctx, g, repo, &ListPullRequestsOptionBase{Base: ref}, ListPullRequestsOptionStateAll())
			if err == nil && pr != nil {
				return pr.GetHead().GetSHA(), nil
			}
		}
	}
	return refs[0], nil
}

func FindRulesetTargetRefs(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleset *github.RepositoryRuleset) ([]string, error) {
	if ruleset.Target == nil {
		return []string{"HEAD"}, nil
	}
	switch *ruleset.Target {
	case github.RulesetTargetBranch:
		return FindRulesetTargetBranches(ctx, g, repo, ruleset)
	case github.RulesetTargetPush:
		return FindRulesetTargetBranches(ctx, g, repo, ruleset)
	case github.RulesetTargetTag:
		return FindRulesetTargetTags(ctx, g, repo, ruleset)
	default:
		return []string{"HEAD"}, nil
	}
}

func FindRulesetTargetBranches(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleset *github.RepositoryRuleset) ([]string, error) {
	if ruleset.Conditions == nil || ruleset.Conditions.RefName == nil {
		return []string{"HEAD"}, nil
	}

	// Get all branches
	branches, err := ListProtectedBranches(ctx, g, repo)
	if err != nil {
		return nil, err
	}

	if len(branches) == 0 {
		branches, err = ListBranches(ctx, g, repo)
		if err != nil {
			return nil, err
		}
	}

	defaultBranch, err := getDefaultBranchIfNeeded(ctx, g, repo, ruleset.Conditions.RefName)
	if err != nil {
		return nil, err
	}

	// Search for a branch that matches the conditions
	var matchedBranches []string
	for _, branch := range branches {
		branchName := branch.GetName()
		if MatchRulesetRefName(branchName, defaultBranch, ruleset.Conditions.RefName) {
			matchedBranches = append(matchedBranches, branchName)
		}
	}

	if len(matchedBranches) > 0 {
		return matchedBranches, nil
	}

	if defaultBranch != "" {
		return []string{defaultBranch}, nil
	}

	// If no branch matches, return HEAD
	return []string{"HEAD"}, nil
}

func FindRulesetTargetTags(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleset *github.RepositoryRuleset) ([]string, error) {
	if ruleset.Conditions == nil || ruleset.Conditions.RefName == nil {
		return []string{"HEAD"}, nil
	}

	defaultBranch, err := getDefaultBranchIfNeeded(ctx, g, repo, ruleset.Conditions.RefName)
	if err != nil {
		return nil, err
	}

	// Get all tags
	tags, err := ListTags(ctx, g, repo)
	if err != nil {
		return nil, err
	}

	// Search for tags that match the conditions
	var matchedTags []string
	for _, tag := range tags {
		tagName := tag.GetName()
		if MatchRulesetRefName(tagName, defaultBranch, ruleset.Conditions.RefName) {
			matchedTags = append(matchedTags, tagName)
		}
	}

	if len(matchedTags) > 0 {
		return matchedTags, nil
	}

	// If no tag matches, return HEAD
	return []string{"HEAD"}, nil
}

// matchRefName checks if a branch name matches the RefName conditions
func MatchRulesetRefName(branchName string, defaultBranch string, refName *github.RepositoryRulesetRefConditionParameters) bool {
	// Check exclude patterns first
	for _, exclude := range refName.Exclude {
		if matchPattern(branchName, defaultBranch, exclude) {
			return false
		}
	}

	// Check include patterns
	for _, include := range refName.Include {
		if matchPattern(branchName, defaultBranch, include) {
			return true
		}
	}

	return false
}

// matchPattern checks if a branch name matches a pattern using fnmatch syntax
func matchPattern(branchName string, defaultBranch string, pattern string) bool {
	// Handle special patterns
	if pattern == "~DEFAULT_BRANCH" {
		return branchName == defaultBranch
	}
	if pattern == "~ALL" {
		return true
	}

	// Remove refs/heads/ prefix if present in branch name for matching
	branchNameForMatch := strings.TrimPrefix(branchName, "refs/heads/")
	patternForMatch := strings.TrimPrefix(pattern, "refs/heads/")

	// Use filepath.Match for fnmatch-style pattern matching
	// This supports:
	// - * matches any sequence of non-separator characters
	// - ? matches any single non-separator character
	// - [...] matches any character in the brackets
	// - [^...] or [!...] matches any character not in the brackets
	matched, err := filepath.Match(patternForMatch, branchNameForMatch)
	if err != nil {
		// If pattern is invalid, fall back to exact match
		return pattern == branchName
	}

	return matched
}

func getDefaultBranchIfNeeded(ctx context.Context, g *GitHubClient, repo repository.Repository, refName *github.RepositoryRulesetRefConditionParameters) (string, error) {
	// Get the default branch if ~DEFAULT_BRANCH is specified
	var defaultBranch string
	for _, ref := range refName.Include {
		if ref == "~DEFAULT_BRANCH" {
			r, err := GetRepository(ctx, g, repo)
			if err != nil {
				return "HEAD", err
			}
			defaultBranch = r.GetDefaultBranch()
			break
		}
	}

	return defaultBranch, nil
}

func GetRulesetTargetRepository(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleset *github.RepositoryRuleset) (*repository.Repository, error) {
	if repo.Name != "" {
		return &repo, nil
	}
	repos, err := GetRulesetTargetRepositories(ctx, g, repo, ruleset)
	if err != nil {
		return nil, err
	}
	if len(repos) > 0 {
		return repos[0], nil
	}
	return nil, nil
}

func GetRulesetTargetRepositories(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleset *github.RepositoryRuleset) ([]*repository.Repository, error) {
	if ruleset.Conditions == nil {
		return nil, nil
	}
	if ruleset.Conditions.RepositoryName != nil {
		repos, err := getRulesetTargetRepositoriesFromName(ctx, g, repo, ruleset.Conditions.RepositoryName)
		if err != nil {
			return nil, err
		}
		return repos, nil
	}
	if ruleset.Conditions.RepositoryProperty != nil {
		repos, err := getRulesetTargetRepositoriesFromProperty(ctx, g, repo, ruleset.Conditions.RepositoryProperty)
		if err != nil {
			return nil, err
		}
		return repos, nil
	}
	if ruleset.Conditions.RepositoryID != nil {
		repos, err := getRulesetTargetRepositoriesFromID(ctx, g, ruleset.Conditions.RepositoryID)
		if err != nil {
			return nil, err
		}
		return repos, nil
	}
	return nil, nil
}

func getRulesetTargetRepositoriesFromName(ctx context.Context, g *GitHubClient, repo repository.Repository, condition *github.RepositoryRulesetRepositoryNamesConditionParameters) ([]*repository.Repository, error) {
	// Get all repositories in the organization
	ghRepos, err := g.ListOrganizationRepositories(ctx, repo.Owner, "all")
	if err != nil {
		return nil, err
	}

	var targetRepos []*repository.Repository

	for _, ghRepo := range ghRepos {
		repoName := ghRepo.GetName()

		// Check if repository name matches the conditions
		if matchRepositoryName(repoName, condition) {
			targetRepos = append(targetRepos, &repository.Repository{
				Owner: repo.Owner,
				Name:  repoName,
			})
		}
	}

	return targetRepos, nil
}

// matchRepositoryName checks if a repository name matches the RepositoryName conditions
func matchRepositoryName(repoName string, condition *github.RepositoryRulesetRepositoryNamesConditionParameters) bool {
	// Check exclude patterns first
	for _, exclude := range condition.Exclude {
		if matchPattern(repoName, "", exclude) {
			return false
		}
	}

	// Check include patterns
	for _, include := range condition.Include {
		if matchPattern(repoName, "", include) {
			return true
		}
	}

	return false
}

func getRulesetTargetRepositoriesFromProperty(ctx context.Context, g *GitHubClient, repo repository.Repository, condition *github.RepositoryRulesetRepositoryPropertyConditionParameters) ([]*repository.Repository, error) {
	// Get all repositories in the organization
	ghRepos, err := g.ListOrganizationRepositories(ctx, repo.Owner, "all")
	if err != nil {
		return nil, err
	}

	var targetRepos []*repository.Repository

	for _, ghRepo := range ghRepos {
		// Check if repository properties match the conditions
		if matchRepositoryProperty(ghRepo, condition) {
			targetRepos = append(targetRepos, &repository.Repository{
				Owner: repo.Owner,
				Name:  ghRepo.GetName(),
			})
		}
	}

	return targetRepos, nil
}

// matchRepositoryProperty checks if a repository's custom properties match the conditions
func matchRepositoryProperty(ghRepo *github.Repository, condition *github.RepositoryRulesetRepositoryPropertyConditionParameters) bool {
	// Build a map of repository properties for quick lookup
	repoProperties := make(map[string]string)
	if ghRepo.CustomProperties != nil {
		for key, value := range ghRepo.CustomProperties {
			// CustomProperties values can be of different types, convert to string
			if strValue, ok := value.(string); ok {
				repoProperties[key] = strValue
			}
		}
	}

	// Check exclude patterns first
	for _, exclude := range condition.Exclude {
		if matchPropertyPattern(repoProperties, exclude) {
			return false
		}
	}

	// Check include patterns
	for _, include := range condition.Include {
		if matchPropertyPattern(repoProperties, include) {
			return true
		}
	}

	return false
}

// matchPropertyPattern checks if a repository's properties match a specific pattern
func matchPropertyPattern(repoProperties map[string]string, pattern *github.RepositoryRulesetRepositoryPropertyTargetParameters) bool {
	propertyValue, exists := repoProperties[pattern.Name]

	// If property doesn't exist, it doesn't match
	if !exists {
		return false
	}

	// If no specific values are required, any value matches
	if len(pattern.PropertyValues) == 0 {
		return true
	}

	// Check if the property value matches any of the required values
	for _, requiredValue := range pattern.PropertyValues {
		if propertyValue == requiredValue {
			return true
		}
	}

	return false
}

func getRulesetTargetRepositoriesFromID(ctx context.Context, g *GitHubClient, condition *github.RepositoryRulesetRepositoryIDsConditionParameters) ([]*repository.Repository, error) {
	if len(condition.RepositoryIDs) == 0 {
		return []*repository.Repository{}, nil
	}

	// Get repositories by ID
	var targetRepos []*repository.Repository
	for _, id := range condition.RepositoryIDs {
		ghRepo, err := GetRepositoryByID(ctx, g, id)
		if err != nil {
			logger.Warn("Failed to get repository by ID, skipping...", "id", id, "error", err)
			continue
		}

		targetRepos = append(targetRepos, &repository.Repository{
			Owner: ghRepo.GetOwner().GetLogin(),
			Name:  ghRepo.GetName(),
		})
	}

	return targetRepos, nil
}
