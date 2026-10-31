package gh

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/google/go-github/v93/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
	"github.com/stretchr/testify/require"
)

type recordedRequest struct {
	Method  string
	URL     string
	Version string
	Body    string
}

type fixtureResponse struct {
	Status int
	Body   string
	Link   string
}

// fixtureTransport serves canned responses keyed by "METHOD URL" and records every request.
type fixtureTransport struct {
	mu        sync.Mutex
	responses map[string]fixtureResponse
	requests  []recordedRequest
}

func (f *fixtureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
	}
	key := req.Method + " " + req.URL.String()
	f.mu.Lock()
	f.requests = append(f.requests, recordedRequest{Method: req.Method, URL: req.URL.String(), Version: req.Header.Get("X-GitHub-Api-Version"), Body: string(body)})
	res, ok := f.responses[key]
	f.mu.Unlock()
	if !ok {
		res = fixtureResponse{Status: http.StatusNotFound, Body: `{"message":"Not Found"}`}
	}
	header := http.Header{"Content-Type": []string{"application/json"}}
	if res.Link != "" {
		header.Set("Link", res.Link)
	}
	return &http.Response{StatusCode: res.Status, Header: header, Body: io.NopCloser(bytes.NewBufferString(res.Body)), Request: req}, nil
}

// RawTransport returns the fixture itself because it does not inject authentication.
func (f *fixtureTransport) RawTransport() http.RoundTripper {
	return f
}

func newFixtureClient(t *testing.T, responses map[string]fixtureResponse) (*GitHubClient, *fixtureTransport) {
	t.Helper()
	tr := &fixtureTransport{responses: responses}
	base := "https://api.example.test/"
	gc, err := github.NewClient(github.WithHTTPClient(&http.Client{Transport: tr}), github.WithURLs(&base, nil))
	require.NoError(t, err)
	g, err := client.NewClient(gc)
	require.NoError(t, err)
	return g, tr
}

func ok(body string) fixtureResponse {
	return fixtureResponse{Status: http.StatusOK, Body: body}
}

const fixtureAPI = "https://api.example.test/"
