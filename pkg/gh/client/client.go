package client

import (
	"encoding/base64"
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

// GitAuthEnvs returns GIT_CONFIG_* environment variables that inject an HTTP
// Authorization header scoped to the host of rawURL. Passing these to a git
// command's Env avoids embedding the token in the URL and prevents it from
// being written to .git/config or shell history, though environment variables
// may still be observable depending on the operating system and configuration.
// The GIT_CONFIG_COUNT/KEY/VALUE mechanism requires git ≥ 2.31.
func (g *GitHubClient) GitAuthEnvs(rawURL string) []string {
	token := g.bearerToken()
	if token == "" || rawURL == "" {
		return nil
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}
	// Scope the header to the specific host to avoid sending credentials to
	// unintended servers (e.g., when intermediate redirects occur).
	configKey := fmt.Sprintf("http.%s://%s/.extraHeader", u.Scheme, u.Host)
	creds := base64.StdEncoding.EncodeToString([]byte("x-access-token:" + token))
	return []string{
		"GIT_CONFIG_COUNT=1",
		"GIT_CONFIG_KEY_0=" + configKey,
		"GIT_CONFIG_VALUE_0=Authorization: Basic " + creds,
	}
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
