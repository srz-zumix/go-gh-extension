package gh

import (
	"context"
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// SecretScanningPushProtectionSettings is the list of valid push protection settings.
var SecretScanningPushProtectionSettings = []string{
	"enabled",
	"disabled",
	"not_set",
}

// SecretScanningAlertStates is the list of valid state values for filtering secret scanning alerts.
var SecretScanningAlertStates = []string{
	"open",
	"resolved",
}

// SecretScanningAlertResolutions is the list of valid resolution values for secret scanning alerts.
var SecretScanningAlertResolutions = []string{
	"false_positive",
	"wont_fix",
	"revoked",
	"used_in_tests",
	"pattern_edited",
	"pattern_deleted",
}

// SecretScanningAlertUpdateStates is the list of valid state values for updating a secret scanning alert.
var SecretScanningAlertUpdateStates = []string{
	"open",
	"resolved",
}

// SecretScanningAlertUpdateResolutions is the list of valid resolution values when updating.
var SecretScanningAlertUpdateResolutions = []string{
	"false_positive",
	"wont_fix",
	"revoked",
	"used_in_tests",
}

// SecretScanningAlertValidities is the list of valid validity values.
var SecretScanningAlertValidities = []string{
	"active",
	"inactive",
	"unknown",
}

// SecretScanningAlertSortOptions is the list of valid sort values for secret scanning alerts.
var SecretScanningAlertSortOptions = []string{
	"created",
	"updated",
}

// ListSecretScanningAlertsOptions holds filter/sort options for listing secret scanning alerts.
type ListSecretScanningAlertsOptions struct {
	State      string
	SecretType string
	Resolution string
	Validity   string
	Sort       string
	Direction  string
}

// UpdateSecretScanningAlertOptions holds options for updating a secret scanning alert.
type UpdateSecretScanningAlertOptions struct {
	State             string
	Resolution        string
	ResolutionComment string
}

// SecretScanningProviderPatternSetting holds a push protection setting for a provider pattern.
type SecretScanningProviderPatternSetting struct {
	TokenType             string
	PushProtectionSetting string
}

// SecretScanningCustomPatternSetting holds a push protection setting for a custom pattern.
type SecretScanningCustomPatternSetting struct {
	TokenType             string
	CustomPatternVersion  string
	PushProtectionSetting string
}

// SecretScanningPatternConfigsUpdateOptions holds the options for updating secret scanning pattern configurations.
type SecretScanningPatternConfigsUpdateOptions struct {
	PatternConfigVersion    string
	ProviderPatternSettings []*SecretScanningProviderPatternSetting
	CustomPatternSettings   []*SecretScanningCustomPatternSetting
}

// toGitHubSecretScanningPatternConfigsUpdateOptions converts the wrapper type to a github.SecretScanningPatternConfigsUpdateOptions.
func toGitHubSecretScanningPatternConfigsUpdateOptions(opts *SecretScanningPatternConfigsUpdateOptions) *github.SecretScanningPatternConfigsUpdateOptions {
	if opts == nil {
		return nil
	}
	ghOpts := &github.SecretScanningPatternConfigsUpdateOptions{}
	ghOpts.PatternConfigVersion = stringPtrIfSet(opts.PatternConfigVersion)
	for _, p := range opts.ProviderPatternSettings {
		if p == nil {
			continue
		}
		ghOpts.ProviderPatternSettings = append(ghOpts.ProviderPatternSettings, &github.SecretScanningProviderPatternSetting{
			TokenType:             p.TokenType,
			PushProtectionSetting: p.PushProtectionSetting,
		})
	}
	for _, p := range opts.CustomPatternSettings {
		if p == nil {
			continue
		}
		s := &github.SecretScanningCustomPatternSetting{
			TokenType:             p.TokenType,
			PushProtectionSetting: p.PushProtectionSetting,
		}
		s.CustomPatternVersion = stringPtrIfSet(p.CustomPatternVersion)
		ghOpts.CustomPatternSettings = append(ghOpts.CustomPatternSettings, s)
	}
	return ghOpts
}

// ParseProviderPattern parses a "TOKEN_TYPE=SETTING" string into a SecretScanningProviderPatternSetting.
func ParseProviderPattern(s string) (*SecretScanningProviderPatternSetting, error) {
	if strings.Count(s, "=") != 1 {
		return nil, fmt.Errorf("invalid provider pattern %q: expected exactly one '=' separator", s)
	}
	before, after, ok := strings.Cut(s, "=")
	if !ok {
		return nil, fmt.Errorf("invalid provider pattern %q: expected TOKEN_TYPE=SETTING", s)
	}
	if before == "" || after == "" {
		return nil, fmt.Errorf("invalid provider pattern %q: TOKEN_TYPE and SETTING must not be empty", s)
	}
	return &SecretScanningProviderPatternSetting{
		TokenType:             before,
		PushProtectionSetting: after,
	}, nil
}

// ParseCustomPattern parses a "TOKEN_TYPE=SETTING" or "TOKEN_TYPE:VERSION=SETTING" string
// into a SecretScanningCustomPatternSetting.
func ParseCustomPattern(s string) (*SecretScanningCustomPatternSetting, error) {
	idx := strings.LastIndex(s, "=")
	if idx < 0 {
		return nil, fmt.Errorf("invalid custom pattern %q: expected TOKEN_TYPE=SETTING or TOKEN_TYPE:VERSION=SETTING", s)
	}
	key := s[:idx]
	setting := s[idx+1:]
	if key == "" || setting == "" {
		return nil, fmt.Errorf("invalid custom pattern %q: TOKEN_TYPE and SETTING must not be empty", s)
	}
	result := &SecretScanningCustomPatternSetting{
		PushProtectionSetting: setting,
	}
	if colon := strings.Index(key, ":"); colon >= 0 {
		tokenType := key[:colon]
		version := key[colon+1:]
		if tokenType == "" || version == "" {
			return nil, fmt.Errorf("invalid custom pattern %q: TOKEN_TYPE and VERSION must not be empty", s)
		}
		result.TokenType = tokenType
		result.CustomPatternVersion = version
	} else {
		result.TokenType = key
	}
	return result, nil
}

// ListSecretScanningPatternConfigs lists secret scanning pattern configurations for an organization.
func ListSecretScanningPatternConfigs(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.SecretScanningPatternConfigs, error) {
	return g.ListOrgSecretScanningPatternConfigs(ctx, repo.Owner)
}

// UpdateSecretScanningPatternConfigs updates secret scanning pattern configurations for an organization.
func UpdateSecretScanningPatternConfigs(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *SecretScanningPatternConfigsUpdateOptions) (*github.SecretScanningPatternConfigsUpdate, error) {
	return g.UpdateOrgSecretScanningPatternConfigs(ctx, repo.Owner, toGitHubSecretScanningPatternConfigsUpdateOptions(opts))
}

// toGitHubSecretScanningAlertListOptions converts ListSecretScanningAlertsOptions to github.SecretScanningAlertListOptions.
func toGitHubSecretScanningAlertListOptions(opts *ListSecretScanningAlertsOptions) *github.SecretScanningAlertListOptions {
	if opts == nil {
		return nil
	}
	o := &github.SecretScanningAlertListOptions{}
	o.State = opts.State
	o.SecretType = opts.SecretType
	o.Resolution = opts.Resolution
	o.Validity = opts.Validity
	o.Sort = opts.Sort
	o.Direction = opts.Direction
	return o
}

// toGitHubSecretScanningAlertUpdateOptions converts UpdateSecretScanningAlertOptions to github.SecretScanningAlertUpdateOptions.
func toGitHubSecretScanningAlertUpdateOptions(opts *UpdateSecretScanningAlertOptions) *github.SecretScanningAlertUpdateOptions {
	if opts == nil {
		return nil
	}
	o := &github.SecretScanningAlertUpdateOptions{
		State: opts.State,
	}
	if opts.Resolution != "" {
		o.Resolution = &opts.Resolution
	}
	if opts.ResolutionComment != "" {
		o.ResolutionComment = &opts.ResolutionComment
	}
	return o
}

// ListSecretScanningOrgAlerts lists secret scanning alerts for an organization.
func ListSecretScanningOrgAlerts(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *ListSecretScanningAlertsOptions) ([]*github.SecretScanningAlert, error) {
	alerts, err := g.ListOrgSecretScanningAlerts(ctx, repo.Owner, toGitHubSecretScanningAlertListOptions(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to list secret scanning alerts for org %s: %w", repo.Owner, err)
	}
	return alerts, nil
}

// ListSecretScanningRepoAlerts lists secret scanning alerts for a repository.
func ListSecretScanningRepoAlerts(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *ListSecretScanningAlertsOptions) ([]*github.SecretScanningAlert, error) {
	alerts, err := g.ListRepoSecretScanningAlerts(ctx, repo.Owner, repo.Name, toGitHubSecretScanningAlertListOptions(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to list secret scanning alerts for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return alerts, nil
}

// ListSecretScanningAlerts lists secret scanning alerts for an organization or repository.
// When repo.Name is empty, alerts are listed for the entire organization; otherwise for the specific repository.
func ListSecretScanningAlerts(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *ListSecretScanningAlertsOptions) ([]*github.SecretScanningAlert, error) {
	if repo.Name == "" {
		return ListSecretScanningOrgAlerts(ctx, g, repo, opts)
	}
	return ListSecretScanningRepoAlerts(ctx, g, repo, opts)
}

// GetSecretScanningAlert gets a single secret scanning alert for a repository.
func GetSecretScanningAlert(ctx context.Context, g *GitHubClient, repo repository.Repository, number int64) (*github.SecretScanningAlert, error) {
	alert, err := g.GetSecretScanningAlert(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret scanning alert #%d for %s/%s: %w", number, repo.Owner, repo.Name, err)
	}
	return alert, nil
}

// UpdateSecretScanningAlert updates a secret scanning alert for a repository.
func UpdateSecretScanningAlert(ctx context.Context, g *GitHubClient, repo repository.Repository, number int64, opts *UpdateSecretScanningAlertOptions) (*github.SecretScanningAlert, error) {
	alert, err := g.UpdateSecretScanningAlert(ctx, repo.Owner, repo.Name, number, toGitHubSecretScanningAlertUpdateOptions(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to update secret scanning alert #%d for %s/%s: %w", number, repo.Owner, repo.Name, err)
	}
	return alert, nil
}

// ListSecretScanningAlertLocations lists all locations for a secret scanning alert.
func ListSecretScanningAlertLocations(ctx context.Context, g *GitHubClient, repo repository.Repository, number int64) ([]*github.SecretScanningAlertLocation, error) {
	locations, err := g.ListSecretScanningAlertLocations(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to list locations for secret scanning alert #%d in %s/%s: %w", number, repo.Owner, repo.Name, err)
	}
	return locations, nil
}

// GetSecretScanningScanHistory gets the secret scanning scan history for a repository.
func GetSecretScanningScanHistory(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.SecretScanningScanHistory, error) {
	history, err := g.GetSecretScanningScanHistory(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret scanning scan history for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return history, nil
}
