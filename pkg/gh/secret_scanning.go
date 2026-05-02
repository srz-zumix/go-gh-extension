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
	"not-set",
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
		result.TokenType = key[:colon]
		result.CustomPatternVersion = key[colon+1:]
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
