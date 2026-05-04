package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// AuditEntry is an alias for github.AuditEntry.
type AuditEntry = github.AuditEntry

// GetAuditLogOptions holds optional parameters for querying the audit log.
type GetAuditLogOptions struct {
	// Phrase is a search phrase to filter audit log events. (Optional.)
	Phrase string
	// Include filters by event type. Can be "web", "git", or "all". Default: "web". (Optional.)
	Include string
	// Order is the sort order of events. Can be "asc" or "desc". Default: "desc". (Optional.)
	Order string
	// MaxEntries limits the total number of entries returned. Use -1 for unlimited.
	MaxEntries int
}

// toGitHubGetAuditLogOptions converts GetAuditLogOptions to github.GetAuditLogOptions.
func toGitHubGetAuditLogOptions(opts *GetAuditLogOptions) (*github.GetAuditLogOptions, int) {
	if opts == nil {
		return nil, -1
	}
	ghOpts := &github.GetAuditLogOptions{}
	if opts.Phrase != "" {
		ghOpts.Phrase = github.Ptr(opts.Phrase)
	}
	if opts.Include != "" {
		ghOpts.Include = github.Ptr(opts.Include)
	}
	if opts.Order != "" {
		ghOpts.Order = github.Ptr(opts.Order)
	}
	maxEntries := opts.MaxEntries
	if maxEntries == 0 {
		maxEntries = -1
	}
	return ghOpts, maxEntries
}

// GetAuditLog retrieves audit log entries for the organization associated with repo.
func GetAuditLog(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *GetAuditLogOptions) ([]*AuditEntry, error) {
	ghOpts, maxEntries := toGitHubGetAuditLogOptions(opts)
	entries, err := g.GetAuditLog(ctx, repo.Owner, ghOpts, maxEntries)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log for '%s': %w", repo.Owner, err)
	}
	return entries, nil
}

// AuditEntryStringField returns the string value for key from an audit log
// entry's AdditionalFields or Data maps. Returns empty string if not found.
func AuditEntryStringField(e *github.AuditEntry, key string) string {
	for _, m := range []map[string]any{e.AdditionalFields, e.Data} {
		if v, ok := m[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}
