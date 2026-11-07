package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-github/v93/github"
)

// CopilotCodeReviewSourceKind is the kind of a source Copilot code review
// reported as contributing to a review comment.
type CopilotCodeReviewSourceKind string

const (
	CopilotCodeReviewSourceKindSkill              CopilotCodeReviewSourceKind = "SKILL"
	CopilotCodeReviewSourceKindMCP                CopilotCodeReviewSourceKind = "MCP"
	CopilotCodeReviewSourceKindCustomInstructions CopilotCodeReviewSourceKind = "CUSTOM_INSTRUCTIONS"
)

// CopilotCodeReviewSource is a skill, MCP server, or custom instructions file
// Copilot code review reported as a source for a review comment.
type CopilotCodeReviewSource struct {
	Kind CopilotCodeReviewSourceKind `json:"kind"`
	Name string                      `json:"name"`
	Path string                      `json:"path,omitempty"`
}

// String returns the source as KIND:NAME. Custom instructions use their path
// instead of the name when a path is reported.
func (s CopilotCodeReviewSource) String() string {
	if s.Kind == CopilotCodeReviewSourceKindCustomInstructions && s.Path != "" {
		return fmt.Sprintf("%s:%s", s.Kind, s.Path)
	}
	return fmt.Sprintf("%s:%s", s.Kind, s.Name)
}

// CopilotAttributedPullRequestReview is a pull request review with the Copilot
// code review attribution fields that go-github does not decode yet.
type CopilotAttributedPullRequestReview struct {
	github.PullRequestReview
	IsCopilotCodeReview     *bool   `json:"is_copilot_code_review,omitempty"`
	CopilotCodeReviewEffort *string `json:"copilot_code_review_effort,omitempty"`
}

// CopilotAttributedPullRequestComment is a pull request review comment with the
// Copilot code review attribution fields that go-github does not decode yet.
type CopilotAttributedPullRequestComment struct {
	github.PullRequestComment
	IsCopilotCodeReview      *bool                     `json:"is_copilot_code_review,omitempty"`
	CopilotCodeReviewSources []CopilotCodeReviewSource `json:"copilot_code_review_sources,omitempty"`
}

// CopilotCodeReviewRequest is the response of a Copilot code review request.
type CopilotCodeReviewRequest struct {
	NodeID string `json:"node_id"`
	State  string `json:"state"`
}

// CopilotCodeReviewRequestBody is the body of a Copilot code review request.
type CopilotCodeReviewRequestBody struct {
	ReviewEffort string `json:"review_effort,omitempty"`
}

// RepositoryCopilotCodeReviewEffort is the review effort configured for a repository.
type RepositoryCopilotCodeReviewEffort struct {
	Configured *string `json:"configured"`
	Effective  string  `json:"effective"`
}

// RepositoryCopilotCodeReviewSettings is the Copilot code review settings of a repository.
type RepositoryCopilotCodeReviewSettings struct {
	ReviewEffort              RepositoryCopilotCodeReviewEffort `json:"review_effort"`
	AvailableReviewEfforts    []string                          `json:"available_review_efforts"`
	SeverityFilter            string                            `json:"severity_filter"`
	CustomInstructionsEnabled bool                              `json:"custom_instructions_enabled"`
	MCPEnabled                bool                              `json:"mcp_enabled"`
}

// RepositoryCopilotCodeReviewSettingsUpdate is the body to update the Copilot
// code review settings of a repository. Nil fields are left unchanged.
type RepositoryCopilotCodeReviewSettingsUpdate struct {
	ReviewEffort              *string `json:"review_effort,omitempty"`
	SeverityFilter            *string `json:"severity_filter,omitempty"`
	CustomInstructionsEnabled *bool   `json:"custom_instructions_enabled,omitempty"`
	MCPEnabled                *bool   `json:"mcp_enabled,omitempty"`
}

// OrganizationCopilotCodeReviewEffort is the default review effort of an organization.
type OrganizationCopilotCodeReviewEffort struct {
	Configured    *string `json:"configured"`
	Effective     string  `json:"effective"`
	GitHubDefault string  `json:"github_default"`
	Editable      bool    `json:"editable"`
	ManagedBy     string  `json:"managed_by"`
}

// OrganizationCopilotCodeReviewSettings is the Copilot code review settings of an organization.
type OrganizationCopilotCodeReviewSettings struct {
	DefaultReviewEffort    OrganizationCopilotCodeReviewEffort `json:"default_review_effort"`
	AvailableReviewEfforts []string                            `json:"available_review_efforts"`
}

// OrganizationCopilotCodeReviewSettingsUpdate is the body to update the Copilot
// code review settings of an organization. Nil fields are left unchanged.
type OrganizationCopilotCodeReviewSettingsUpdate struct {
	DefaultReviewEffort *string `json:"default_review_effort,omitempty"`
}

// OrganizationCopilotCodeReviewRunner is the runner configuration Copilot code
// review uses for an organization.
type OrganizationCopilotCodeReviewRunner struct {
	RunnerType                   string   `json:"runner_type"`
	RunnerLabels                 []string `json:"runner_labels"`
	RunnerGroupName              *string  `json:"runner_group_name"`
	AllowRepositoryCustomization bool     `json:"allow_repository_customization"`
	ConfiguredBy                 string   `json:"configured_by,omitempty"`
}

// OrganizationCopilotCodeReviewRunnerUpdate is the body to set the runner
// configuration Copilot code review uses for an organization.
type OrganizationCopilotCodeReviewRunnerUpdate struct {
	RunnerType                   string   `json:"runner_type"`
	RunnerLabels                 []string `json:"runner_labels,omitempty"`
	RunnerGroupName              *string  `json:"runner_group_name,omitempty"`
	AllowRepositoryCustomization *bool    `json:"allow_repository_customization,omitempty"`
}

// OrganizationCopilotCodeReviewPolicy is the Copilot code review policy of an organization.
type OrganizationCopilotCodeReviewPolicy struct {
	ConfiguredState *string `json:"configured_state"`
	EffectiveState  string  `json:"effective_state"`
	Editable        bool    `json:"editable"`
	ManagedBy       string  `json:"managed_by"`
}

// EnterpriseCopilotCodeReviewPolicy is the Copilot code review policy of an enterprise.
type EnterpriseCopilotCodeReviewPolicy struct {
	PolicyState *string `json:"policy_state"`
}

func repoCopilotCodeReviewSettingsURL(owner, repo string) string {
	return fmt.Sprintf("repos/%s/%s/copilot/code-review/settings", url.PathEscape(owner), url.PathEscape(repo))
}

func orgCopilotCodeReviewURL(org, suffix string) string {
	return fmt.Sprintf("orgs/%s/copilot/%s", url.PathEscape(org), suffix)
}

// doJSON sends a request with an optional JSON body and decodes the response into out when non-nil.
func (g *GitHubClient) doJSON(ctx context.Context, method, u string, body, out any) error {
	req, err := g.client.NewRequest(ctx, method, u, body)
	if err != nil {
		return err
	}
	_, err = g.client.Do(req, out)
	return err
}

// listCopilotAttributed fetches every page of a list endpoint that returns Copilot attributed items.
func listCopilotAttributed[T any](ctx context.Context, g *GitHubClient, u string) ([]*T, error) {
	var all []*T
	page := 1
	for {
		req, err := g.client.NewRequest(ctx, http.MethodGet, fmt.Sprintf("%s?per_page=%d&page=%d", u, defaultPerPage, page), nil)
		if err != nil {
			return nil, err
		}
		var items []*T
		resp, err := g.client.Do(req, &items)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return all, nil
}

// ListPullRequestReviewsWithCopilotAttribution lists the reviews of a pull request,
// including the Copilot code review attribution fields.
func (g *GitHubClient) ListPullRequestReviewsWithCopilotAttribution(ctx context.Context, owner, repo string, number int) ([]*CopilotAttributedPullRequestReview, error) {
	u := fmt.Sprintf("repos/%s/%s/pulls/%d/reviews", url.PathEscape(owner), url.PathEscape(repo), number)
	return listCopilotAttributed[CopilotAttributedPullRequestReview](ctx, g, u)
}

// ListPullRequestReviewCommentsWithCopilotAttribution lists the review comments of a
// pull request, including the Copilot code review attribution fields.
func (g *GitHubClient) ListPullRequestReviewCommentsWithCopilotAttribution(ctx context.Context, owner, repo string, number int) ([]*CopilotAttributedPullRequestComment, error) {
	u := fmt.Sprintf("repos/%s/%s/pulls/%d/comments", url.PathEscape(owner), url.PathEscape(repo), number)
	return listCopilotAttributed[CopilotAttributedPullRequestComment](ctx, g, u)
}

// RequestCopilotCodeReview requests a Copilot code review for a pull request.
func (g *GitHubClient) RequestCopilotCodeReview(ctx context.Context, owner, repo string, number int, body *CopilotCodeReviewRequestBody) (*CopilotCodeReviewRequest, error) {
	u := fmt.Sprintf("repos/%s/%s/pulls/%d/copilot-code-review", url.PathEscape(owner), url.PathEscape(repo), number)
	result := new(CopilotCodeReviewRequest)
	if err := g.doJSON(ctx, http.MethodPost, u, body, result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetRepositoryCopilotCodeReviewSettings gets the Copilot code review settings of a repository.
func (g *GitHubClient) GetRepositoryCopilotCodeReviewSettings(ctx context.Context, owner, repo string) (*RepositoryCopilotCodeReviewSettings, error) {
	result := new(RepositoryCopilotCodeReviewSettings)
	if err := g.doJSON(ctx, http.MethodGet, repoCopilotCodeReviewSettingsURL(owner, repo), nil, result); err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateRepositoryCopilotCodeReviewSettings updates the Copilot code review settings of a repository.
func (g *GitHubClient) UpdateRepositoryCopilotCodeReviewSettings(ctx context.Context, owner, repo string, update *RepositoryCopilotCodeReviewSettingsUpdate) (*RepositoryCopilotCodeReviewSettings, error) {
	result := new(RepositoryCopilotCodeReviewSettings)
	if err := g.doJSON(ctx, http.MethodPatch, repoCopilotCodeReviewSettingsURL(owner, repo), update, result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetOrganizationCopilotCodeReviewSettings gets the Copilot code review settings of an organization.
func (g *GitHubClient) GetOrganizationCopilotCodeReviewSettings(ctx context.Context, org string) (*OrganizationCopilotCodeReviewSettings, error) {
	result := new(OrganizationCopilotCodeReviewSettings)
	if err := g.doJSON(ctx, http.MethodGet, orgCopilotCodeReviewURL(org, "code-review/settings"), nil, result); err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateOrganizationCopilotCodeReviewSettings updates the Copilot code review settings of an organization.
func (g *GitHubClient) UpdateOrganizationCopilotCodeReviewSettings(ctx context.Context, org string, update *OrganizationCopilotCodeReviewSettingsUpdate) (*OrganizationCopilotCodeReviewSettings, error) {
	result := new(OrganizationCopilotCodeReviewSettings)
	if err := g.doJSON(ctx, http.MethodPatch, orgCopilotCodeReviewURL(org, "code-review/settings"), update, result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetOrganizationCopilotCodeReviewRunner gets the Copilot code review runner configuration of an organization.
func (g *GitHubClient) GetOrganizationCopilotCodeReviewRunner(ctx context.Context, org string) (*OrganizationCopilotCodeReviewRunner, error) {
	result := new(OrganizationCopilotCodeReviewRunner)
	if err := g.doJSON(ctx, http.MethodGet, orgCopilotCodeReviewURL(org, "code-review/runner"), nil, result); err != nil {
		return nil, err
	}
	return result, nil
}

// SetOrganizationCopilotCodeReviewRunner sets the Copilot code review runner configuration of an organization.
func (g *GitHubClient) SetOrganizationCopilotCodeReviewRunner(ctx context.Context, org string, update *OrganizationCopilotCodeReviewRunnerUpdate) (*OrganizationCopilotCodeReviewRunner, error) {
	result := new(OrganizationCopilotCodeReviewRunner)
	if err := g.doJSON(ctx, http.MethodPut, orgCopilotCodeReviewURL(org, "code-review/runner"), update, result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetOrganizationCopilotCodeReviewPolicy gets the Copilot code review policy of an organization.
func (g *GitHubClient) GetOrganizationCopilotCodeReviewPolicy(ctx context.Context, org string) (*OrganizationCopilotCodeReviewPolicy, error) {
	result := new(OrganizationCopilotCodeReviewPolicy)
	if err := g.doJSON(ctx, http.MethodGet, orgCopilotCodeReviewURL(org, "policies/code-review"), nil, result); err != nil {
		return nil, err
	}
	return result, nil
}

// SetOrganizationCopilotCodeReviewPolicy sets the Copilot code review policy of an organization.
func (g *GitHubClient) SetOrganizationCopilotCodeReviewPolicy(ctx context.Context, org, state string) error {
	body := map[string]string{"state": state}
	return g.doJSON(ctx, http.MethodPut, orgCopilotCodeReviewURL(org, "policies/code-review"), body, nil)
}

// GetEnterpriseCopilotCodeReviewPolicy gets the Copilot code review policy of an enterprise.
func (g *GitHubClient) GetEnterpriseCopilotCodeReviewPolicy(ctx context.Context, enterprise string) (*EnterpriseCopilotCodeReviewPolicy, error) {
	u := fmt.Sprintf("enterprises/%s/copilot/policies/code-review", url.PathEscape(enterprise))
	result := new(EnterpriseCopilotCodeReviewPolicy)
	if err := g.doJSON(ctx, http.MethodGet, u, nil, result); err != nil {
		return nil, err
	}
	return result, nil
}

// SetEnterpriseCopilotCodeReviewPolicy sets the Copilot code review policy of an enterprise.
func (g *GitHubClient) SetEnterpriseCopilotCodeReviewPolicy(ctx context.Context, enterprise, policyState string) error {
	u := fmt.Sprintf("enterprises/%s/copilot/policies/code-review", url.PathEscape(enterprise))
	body := map[string]string{"policy_state": policyState}
	return g.doJSON(ctx, http.MethodPut, u, body, nil)
}
