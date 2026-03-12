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

// v4EndpointURL returns the GraphQL v4 endpoint URL for the client.
// It uses GITHUB_GRAPHQL_URL env var when targetting github.com, and
// derives the GHES endpoint from the REST API base URL otherwise.
func (g *GitHubClient) v4EndpointURL() string {
	host := g.client.BaseURL.Host
	if host != "api.github.com" {
		return fmt.Sprintf("https://%s/api/graphql", host)
	}
	if ep := os.Getenv("GITHUB_GRAPHQL_URL"); ep != "" {
		return ep
	}
	return defaultV4Endpoint
}

func (g *GitHubClient) GetOrCreateGraphQLClient() (*githubv4.Client, error) {
	if g.graphql != nil {
		return g.graphql, nil
	}
	client := githubv4.NewEnterpriseClient(g.v4EndpointURL(), g.client.Client())
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
