package client

import (
	"fmt"
	"net/url"
	"os"
	"reflect"

	"github.com/google/go-github/v79/github"
	"github.com/google/go-querystring/query"
	"github.com/shurcooL/githubv4"
)

type GitHubClient struct {
	client  *github.Client
	graphql *githubv4.Client
}

var (
	defaultPerPage = 100
)

const defaultV4Endpoint = "https://api.github.com/graphql"

func NewClient(client *github.Client) (*GitHubClient, error) {
	return &GitHubClient{
		client:  client,
		graphql: nil,
	}, nil
}

// GetClient returns the underlying GitHub client
func (g *GitHubClient) GetClient() *github.Client {
	return g.client
}

func (g *GitHubClient) GetOrCreateGraphQLClient() (*githubv4.Client, error) {
	if g.graphql != nil {
		return g.graphql, nil
	}
	httpClient := g.client.Client()
	host := g.client.BaseURL.Host
	v4ep := defaultV4Endpoint
	if host != "api.github.com" {
		// If the base URL is not the default GitHub API, we need to create a new HTTP client
		// with the correct base URL for GraphQL.
		v4ep = fmt.Sprintf("https://%s/api/graphql", host)
	} else {
		if os.Getenv("GITHUB_GRAPHQL_URL") != "" {
			v4ep = os.Getenv("GITHUB_GRAPHQL_URL")
		}
	}
	client := githubv4.NewEnterpriseClient(v4ep, httpClient)
	g.graphql = client
	return client, nil
}

// addOptions adds the parameters in opts as URL query parameters to s.
// opts must be a struct whose fields may contain "url" tags.
func addOptions(s string, opts any) (string, error) {
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opts)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}
