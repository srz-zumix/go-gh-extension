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
	"github.com/google/go-github/v73/github"
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
	Ruleset   *github.RepositoryRuleset
	Teams     map[int64]*github.Team
	CheckRuns map[int64]*github.CheckRun
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
	ruleset, err := GetRepositoryRuleset(ctx, g, repo, rulesetID, false)
	if err != nil {
		return nil, err
	}

	teams := GetRulesetActorsTeams(ctx, g, repo, ruleset)

	checkRuns := make(map[int64]*github.CheckRun)
	if ruleset.Rules.RequiredStatusChecks != nil {
		ref, err := FindRulesetRequireStatusCheckRunRef(ctx, g, repo, ruleset)
		if err != nil {
			ref = "HEAD"
		}
		for _, check := range ruleset.Rules.RequiredStatusChecks.RequiredStatusChecks {
			if check.IntegrationID != nil {
				if _, ok := checkRuns[*check.IntegrationID]; ok {
					continue
				}
				checkRun, err := findIntegrationID(ctx, g, repo, ref, check.Context, check.IntegrationID, nil)
				if err != nil {
					return nil, err
				}
				if checkRun != nil {
					checkRuns[*check.IntegrationID] = checkRun
				}
			}
		}
	}

	return &RepositoryRulesetMigrateConfig{
		Ruleset:   ruleset,
		Teams:     teams,
		CheckRuns: checkRuns,
	}, nil
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

	foundIntegrations := make(map[int64]*int64)

	if ruleset.Rules.RequiredStatusChecks != nil {
		ref, err := FindRulesetRequireStatusCheckRunRef(ctx, g, repo, ruleset)
		if err != nil {
			ref = "HEAD"
		}
		for _, check := range ruleset.Rules.RequiredStatusChecks.RequiredStatusChecks {
			if check.IntegrationID != nil {
				id := *check.IntegrationID
				if dstID, ok := foundIntegrations[id]; ok {
					check.IntegrationID = dstID
					continue
				}
				found, err := findIntegrationID(ctx, g, repo, ref, check.Context, check.IntegrationID, nil)
				if err == nil && found != nil {
					continue
				}
				check.IntegrationID = nil
				checkRun := migrateConfig.CheckRuns[id]
				if checkRun != nil {
					found, err = findIntegrationID(ctx, g, repo, ref, check.Context, nil, checkRun.App)
					if err == nil {
						if found != nil {
							check.IntegrationID = found.App.ID
							foundIntegrations[id] = found.App.ID
							logger.Info("Mapped required status check integration to target repository", "integration", check.Context, "id", found.App.GetID())
							continue
						}

						if checkRun.App != nil && checkRun.App.GetSlug() == "github-actions" {
							actionAppId := gitHubActionsAppID
							if actionAppId == nil && org.IsGitHubCom() {
								actionAppId = &GitHubComGitHubActionsAppID
							}
							if actionAppId != nil {
								check.IntegrationID = actionAppId
								foundIntegrations[id] = actionAppId
								logger.Info("Mapped required status check integration to GitHub Actions in target repository", "integration", check.Context)
								continue
							}
						}
					}
				}
				logger.Warn("Required status check integration not found in target repository, replace to any-source", "integration", check.Context)
			}
		}
	}

	if ruleset.Rules.RequiredDeployments != nil {
		newEnvrionments := []string{}
		for _, env := range ruleset.Rules.RequiredDeployments.RequiredDeploymentEnvironments {
			deployments, err := ListEnvrionmentDeployments(ctx, g, repo, env)
			if err != nil || len(deployments) == 0 {
				logger.Warn("Required deployment environment not found in target repository, skipping...", "environment", env)
			} else {
				newEnvrionments = append(newEnvrionments, env)
			}
		}
		if len(newEnvrionments) == 0 {
			ruleset.Rules.RequiredDeployments = nil
			logger.Warn("No valid required deployment environments found in target repository, removing required deployments rule...")
		} else {
			ruleset.Rules.RequiredDeployments.RequiredDeploymentEnvironments = newEnvrionments
		}
	}

	result, err := CreateOrUpdateRepositoryRuleset(ctx, g, repo, ruleset)
	if err != nil {
		return nil, err
	}

	return result, nil
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

func findIntegrationID(ctx context.Context, g *GitHubClient, repo repository.Repository, ref string, name string, appID *int64, app *github.App) (*github.CheckRun, error) {
	checkRuns, err := ListCheckRunsForRef(ctx, g, repo, ref, &ListChecksRunFilterOptions{
		AppID: appID,
	})
	if err != nil {
		return nil, err
	}
	if checkRuns.Total == nil || *checkRuns.Total == 0 {
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
