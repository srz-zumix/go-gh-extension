package client

import (
	"context"

	"github.com/google/go-github/v84/github"
)

// GetAuditLog retrieves all audit-log entries for an organization matching the
// given options, following cursor-based pagination until all results are fetched.
// maxEntries limits the total number of entries returned; use -1 for unlimited.
func (g *GitHubClient) GetAuditLog(ctx context.Context, org string, opts *github.GetAuditLogOptions, maxEntries int) ([]*github.AuditEntry, error) {
	if opts == nil {
		opts = &github.GetAuditLogOptions{}
	}
	opts.PerPage = defaultPerPage

	var all []*github.AuditEntry
	for {
		entries, resp, err := g.client.Organizations.GetAuditLog(ctx, org, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, entries...)
		if maxEntries > 0 && len(all) >= maxEntries {
			return all[:maxEntries], nil
		}
		if resp.Cursor == "" {
			break
		}
		opts.Cursor = resp.Cursor
	}
	return all, nil
}
