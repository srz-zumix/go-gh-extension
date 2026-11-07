package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

type (
	CopilotCodeReviewSource                     = client.CopilotCodeReviewSource
	CopilotCodeReviewSourceKind                 = client.CopilotCodeReviewSourceKind
	CopilotAttributedPullRequestReview          = client.CopilotAttributedPullRequestReview
	CopilotAttributedPullRequestComment         = client.CopilotAttributedPullRequestComment
	CopilotCodeReviewRequest                    = client.CopilotCodeReviewRequest
	RepositoryCopilotCodeReviewEffort           = client.RepositoryCopilotCodeReviewEffort
	RepositoryCopilotCodeReviewSettings         = client.RepositoryCopilotCodeReviewSettings
	RepositoryCopilotCodeReviewSettingsUpdate   = client.RepositoryCopilotCodeReviewSettingsUpdate
	OrganizationCopilotCodeReviewEffort         = client.OrganizationCopilotCodeReviewEffort
	OrganizationCopilotCodeReviewSettings       = client.OrganizationCopilotCodeReviewSettings
	OrganizationCopilotCodeReviewSettingsUpdate = client.OrganizationCopilotCodeReviewSettingsUpdate
	OrganizationCopilotCodeReviewRunner         = client.OrganizationCopilotCodeReviewRunner
	OrganizationCopilotCodeReviewRunnerUpdate   = client.OrganizationCopilotCodeReviewRunnerUpdate
	OrganizationCopilotCodeReviewPolicy         = client.OrganizationCopilotCodeReviewPolicy
	EnterpriseCopilotCodeReviewPolicy           = client.EnterpriseCopilotCodeReviewPolicy
)

const (
	CopilotCodeReviewSourceKindSkill              = client.CopilotCodeReviewSourceKindSkill
	CopilotCodeReviewSourceKindMCP                = client.CopilotCodeReviewSourceKindMCP
	CopilotCodeReviewSourceKindCustomInstructions = client.CopilotCodeReviewSourceKindCustomInstructions
)

// CopilotCodeReviewEffortInherit clears a configured review effort so the value is inherited.
const CopilotCodeReviewEffortInherit = "inherit"

// CopilotCodeReviewEfforts lists the known review effort levels. The API reports the
// levels actually available through available_review_efforts.
var CopilotCodeReviewEfforts = []string{"lite", "balanced", "max"}

// CopilotCodeReviewSeverityFilters lists the valid severity_filter values.
var CopilotCodeReviewSeverityFilters = []string{"no_filter", "low", "low_and_medium"}

// CopilotCodeReviewRunnerTypes lists the valid runner_type values.
var CopilotCodeReviewRunnerTypes = []string{"standard", "labeled"}

// CopilotCodeReviewOrganizationPolicyStates lists the valid organization policy states.
var CopilotCodeReviewOrganizationPolicyStates = []string{"enabled", "disabled"}

// CopilotCodeReviewEnterprisePolicyStates lists the valid enterprise policy states.
var CopilotCodeReviewEnterprisePolicyStates = []string{
	"enabled_for_all_organizations",
	"disabled_for_all_organizations",
	"configured_by_organization_admins",
}

// ListPullRequestReviewsWithCopilotAttribution lists the reviews of a pull request,
// including the Copilot code review attribution fields when the server reports them.
func ListPullRequestReviewsWithCopilotAttribution(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any) ([]*CopilotAttributedPullRequestReview, error) {
	number, err := GetPullRequestNumber(pull_request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pull request number from '%v': %w", pull_request, err)
	}
	reviews, err := g.ListPullRequestReviewsWithCopilotAttribution(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to list reviews for pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return reviews, nil
}

// ListPullRequestReviewCommentsWithCopilotAttribution lists the review comments of a pull
// request, including the Copilot code review attribution fields when the server reports them.
func ListPullRequestReviewCommentsWithCopilotAttribution(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any) ([]*CopilotAttributedPullRequestComment, error) {
	number, err := GetPullRequestNumber(pull_request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pull request number from '%v': %w", pull_request, err)
	}
	comments, err := g.ListPullRequestReviewCommentsWithCopilotAttribution(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to list review comments for pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return comments, nil
}

// RequestCopilotCodeReview requests a Copilot code review for a pull request.
// An empty effort uses the default review effort.
func RequestCopilotCodeReview(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any, effort string) (*CopilotCodeReviewRequest, error) {
	number, err := GetPullRequestNumber(pull_request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pull request number from '%v': %w", pull_request, err)
	}
	result, err := g.RequestCopilotCodeReview(ctx, repo.Owner, repo.Name, number, &client.CopilotCodeReviewRequestBody{ReviewEffort: effort})
	if err != nil {
		return nil, fmt.Errorf("failed to request a Copilot code review for pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return result, nil
}

// GetRepositoryCopilotCodeReviewSettings gets the Copilot code review settings of a repository.
func GetRepositoryCopilotCodeReviewSettings(ctx context.Context, g *GitHubClient, repo repository.Repository) (*RepositoryCopilotCodeReviewSettings, error) {
	settings, err := g.GetRepositoryCopilotCodeReviewSettings(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get Copilot code review settings for repository '%s/%s': %w", repo.Owner, repo.Name, err)
	}
	return settings, nil
}

// UpdateRepositoryCopilotCodeReviewSettings updates the Copilot code review settings of a repository.
func UpdateRepositoryCopilotCodeReviewSettings(ctx context.Context, g *GitHubClient, repo repository.Repository, update *RepositoryCopilotCodeReviewSettingsUpdate) (*RepositoryCopilotCodeReviewSettings, error) {
	if update == nil {
		return nil, fmt.Errorf("repository Copilot code review settings update must not be nil")
	}
	settings, err := g.UpdateRepositoryCopilotCodeReviewSettings(ctx, repo.Owner, repo.Name, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update Copilot code review settings for repository '%s/%s': %w", repo.Owner, repo.Name, err)
	}
	return settings, nil
}

// GetOrganizationCopilotCodeReviewSettings gets the Copilot code review settings of the organization repo.Owner.
func GetOrganizationCopilotCodeReviewSettings(ctx context.Context, g *GitHubClient, repo repository.Repository) (*OrganizationCopilotCodeReviewSettings, error) {
	settings, err := g.GetOrganizationCopilotCodeReviewSettings(ctx, repo.Owner)
	if err != nil {
		return nil, fmt.Errorf("failed to get Copilot code review settings for organization '%s': %w", repo.Owner, err)
	}
	return settings, nil
}

// UpdateOrganizationCopilotCodeReviewSettings updates the Copilot code review settings of the organization repo.Owner.
func UpdateOrganizationCopilotCodeReviewSettings(ctx context.Context, g *GitHubClient, repo repository.Repository, update *OrganizationCopilotCodeReviewSettingsUpdate) (*OrganizationCopilotCodeReviewSettings, error) {
	if update == nil {
		return nil, fmt.Errorf("organization Copilot code review settings update must not be nil")
	}
	settings, err := g.UpdateOrganizationCopilotCodeReviewSettings(ctx, repo.Owner, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update Copilot code review settings for organization '%s': %w", repo.Owner, err)
	}
	return settings, nil
}

// GetOrganizationCopilotCodeReviewRunner gets the Copilot code review runner configuration of the organization repo.Owner.
func GetOrganizationCopilotCodeReviewRunner(ctx context.Context, g *GitHubClient, repo repository.Repository) (*OrganizationCopilotCodeReviewRunner, error) {
	runner, err := g.GetOrganizationCopilotCodeReviewRunner(ctx, repo.Owner)
	if err != nil {
		return nil, fmt.Errorf("failed to get Copilot code review runner for organization '%s': %w", repo.Owner, err)
	}
	return runner, nil
}

// SetOrganizationCopilotCodeReviewRunner sets the Copilot code review runner configuration of the organization repo.Owner.
func SetOrganizationCopilotCodeReviewRunner(ctx context.Context, g *GitHubClient, repo repository.Repository, update *OrganizationCopilotCodeReviewRunnerUpdate) (*OrganizationCopilotCodeReviewRunner, error) {
	if update == nil {
		return nil, fmt.Errorf("organization Copilot code review runner update must not be nil")
	}
	runner, err := g.SetOrganizationCopilotCodeReviewRunner(ctx, repo.Owner, update)
	if err != nil {
		return nil, fmt.Errorf("failed to set Copilot code review runner for organization '%s': %w", repo.Owner, err)
	}
	return runner, nil
}

// GetOrganizationCopilotCodeReviewPolicy gets the Copilot code review policy of the organization repo.Owner.
func GetOrganizationCopilotCodeReviewPolicy(ctx context.Context, g *GitHubClient, repo repository.Repository) (*OrganizationCopilotCodeReviewPolicy, error) {
	policy, err := g.GetOrganizationCopilotCodeReviewPolicy(ctx, repo.Owner)
	if err != nil {
		return nil, fmt.Errorf("failed to get Copilot code review policy for organization '%s': %w", repo.Owner, err)
	}
	return policy, nil
}

// SetOrganizationCopilotCodeReviewPolicy sets the Copilot code review policy of the organization repo.Owner.
func SetOrganizationCopilotCodeReviewPolicy(ctx context.Context, g *GitHubClient, repo repository.Repository, state string) error {
	if err := g.SetOrganizationCopilotCodeReviewPolicy(ctx, repo.Owner, state); err != nil {
		return fmt.Errorf("failed to set Copilot code review policy for organization '%s': %w", repo.Owner, err)
	}
	return nil
}

// GetEnterpriseCopilotCodeReviewPolicy gets the Copilot code review policy of the enterprise enterprise.Owner.
func GetEnterpriseCopilotCodeReviewPolicy(ctx context.Context, g *GitHubClient, enterprise repository.Repository) (*EnterpriseCopilotCodeReviewPolicy, error) {
	policy, err := g.GetEnterpriseCopilotCodeReviewPolicy(ctx, enterprise.Owner)
	if err != nil {
		return nil, fmt.Errorf("failed to get Copilot code review policy for enterprise '%s': %w", enterprise.Owner, err)
	}
	return policy, nil
}

// SetEnterpriseCopilotCodeReviewPolicy sets the Copilot code review policy of the enterprise enterprise.Owner.
func SetEnterpriseCopilotCodeReviewPolicy(ctx context.Context, g *GitHubClient, enterprise repository.Repository, policyState string) error {
	if err := g.SetEnterpriseCopilotCodeReviewPolicy(ctx, enterprise.Owner, policyState); err != nil {
		return fmt.Errorf("failed to set Copilot code review policy for enterprise '%s': %w", enterprise.Owner, err)
	}
	return nil
}
