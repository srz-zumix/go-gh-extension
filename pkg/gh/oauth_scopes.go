package gh

import (
	"context"
	"strings"
)

// GetTokenScopes returns the OAuth scopes of the client's token.
// ok is false when the token does not report scopes (e.g. fine-grained or GitHub App tokens).
func GetTokenScopes(ctx context.Context, g *GitHubClient) (scopes []string, ok bool, err error) {
	header, ok, err := g.GetOAuthScopesHeader(ctx)
	if err != nil || !ok {
		return nil, ok, err
	}
	for scope := range strings.SplitSeq(header, ",") {
		if scope = strings.TrimSpace(scope); scope != "" {
			scopes = append(scopes, scope)
		}
	}
	return scopes, true, nil
}
