package client

import (
	"context"
	"net/http"
	"strings"
)

// oauthScopesHeader is the response header that lists the scopes of an OAuth or classic personal access token.
const oauthScopesHeader = "X-OAuth-Scopes"

// GetOAuthScopesHeader returns the raw X-OAuth-Scopes header of the API root.
// Multiple header field lines are combined into a single comma-separated value.
// ok is false when the header is absent, as with fine-grained or GitHub App tokens.
func (g *GitHubClient) GetOAuthScopesHeader(ctx context.Context) (header string, ok bool, err error) {
	req, err := g.client.NewRequest(ctx, "GET", "", nil)
	if err != nil {
		return "", false, err
	}
	resp, err := g.client.Do(req, nil)
	if err != nil {
		return "", false, err
	}
	values, ok := resp.Header[http.CanonicalHeaderKey(oauthScopesHeader)]
	if !ok || len(values) == 0 {
		return "", false, nil
	}
	return strings.Join(values, ", "), true, nil
}
