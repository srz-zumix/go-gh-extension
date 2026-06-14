package client

import (
	"context"
	"fmt"
	"net/url"
)

// EnableDisableOrgSecurityFeature enables or disables a security feature for all eligible repositories in an organization.
func (g *GitHubClient) EnableDisableOrgSecurityFeature(ctx context.Context, org, securityProduct, enablement, querySuite string) error {
	u := fmt.Sprintf("orgs/%v/%v/%v", org, securityProduct, enablement)
	if querySuite != "" {
		params := url.Values{}
		params.Set("query_suite", querySuite)
		u = fmt.Sprintf("%s?%s", u, params.Encode())
	}

	req, err := g.client.NewRequest(ctx, "POST", u, nil)
	if err != nil {
		return err
	}

	_, err = g.client.Do(req, nil)
	if err != nil {
		return err
	}

	return nil
}
