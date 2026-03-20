package client

import (
	"fmt"
	"net/http"
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

func NewClient(client *github.Client) (*GitHubClient, error) {
	return &GitHubClient{
		client:  client,
		graphql: nil,
	}, nil
}

// tokenGetter is implemented by transports that can expose their authentication token.
type tokenGetter interface {
	Token() string
}

// bearerToken returns the access token configured in this client's transport.
// It unwraps getOnlyRoundTripper if present and uses a type assertion to read the token.
// Returns an empty string if no token is available.
func (g *GitHubClient) bearerToken() string {
	tr := g.client.Client().Transport
	// Unwrap getOnlyRoundTripper since it delegates to the inner transport
	type unwrapper interface {
		Unwrap() http.RoundTripper
	}
	if u, ok := tr.(unwrapper); ok {
		tr = u.Unwrap()
	}
	if tg, ok := tr.(tokenGetter); ok {
		return tg.Token()
	}
	return ""
}

// GetClient returns the underlying GitHub client
func (g *GitHubClient) GetClient() *github.Client {
	return g.client
}

// GitURLWithToken embeds the client's bearer token into rawURL as basic-auth
// credentials for git over HTTPS. If no token is available, rawURL is returned unchanged.
func (g *GitHubClient) GitURLWithToken(rawURL string) string {
	token := g.bearerToken()
	if token == "" || rawURL == "" {
		return rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	u.User = url.UserPassword("x-access-token", token)
	return u.String()
}

// Host returns the GitHub hostname for this client.
// For github.com, returns "github.com". For GitHub Enterprise Server, returns the GHES hostname.
func (g *GitHubClient) Host() string {
	host := g.client.BaseURL.Host
	if host == "api.github.com" {
		return "github.com"
	}
	return host
}

// v4EndpointURL returns the GraphQL v4 endpoint URL for the client.
// It uses GITHUB_GRAPHQL_URL env var when targeting github.com, and
// derives the GHES endpoint from the REST API base URL otherwise.
func (g *GitHubClient) v4EndpointURL() string {
	host := g.client.BaseURL.Host
	if host != "api.github.com" {
		return fmt.Sprintf("https://%s/api/graphql", host)
	}
	if ep := os.Getenv("GITHUB_GRAPHQL_URL"); ep != "" {
		return ep
	}
	return DefaultV4Endpoint
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
