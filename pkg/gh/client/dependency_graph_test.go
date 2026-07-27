package client

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateDependencyGraphSBOMReport(t *testing.T) {
	var gotMethod, gotPath string
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.EscapedPath()
		return &http.Response{
			StatusCode: http.StatusCreated,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"sbom_url":"https://api.github.com/repos/owner/repo/dependency-graph/sbom/fetch-report/uuid-1"}`)),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	result, err := g.GenerateDependencyGraphSBOMReport(t.Context(), "owner", "repo")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, http.MethodGet, gotMethod)
	assert.Equal(t, "/repos/owner/repo/dependency-graph/sbom/generate-report", gotPath)
	assert.Equal(t, "https://api.github.com/repos/owner/repo/dependency-graph/sbom/fetch-report/uuid-1", result.SBOMURL)
}

func TestFetchDependencyGraphSBOMReport_Pending(t *testing.T) {
	var gotPath string
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotPath = r.URL.EscapedPath()
		return &http.Response{
			StatusCode: http.StatusAccepted,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(``)),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	report, err := g.FetchDependencyGraphSBOMReport(t.Context(), "owner", "repo", "uuid-1")
	require.NoError(t, err)
	require.NotNil(t, report)
	assert.True(t, report.Pending)
	assert.Equal(t, "/repos/owner/repo/dependency-graph/sbom/fetch-report/uuid-1", gotPath)
}

func TestFetchDependencyGraphSBOMReport_RedirectDoesNotForwardAuth(t *testing.T) {
	const downloadHost = "https://objects.example.test/download/report.json"
	const sbomBody = `{"SPDXID":"SPDXRef-DOCUMENT","spdxVersion":"SPDX-2.3"}`

	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host == "api.github.com" {
			return &http.Response{
				StatusCode: http.StatusFound,
				Header:     http.Header{"Location": []string{downloadHost}},
				Body:       io.NopCloser(strings.NewReader(``)),
				Request:    r,
			}, nil
		}
		t.Fatalf("unexpected request to authenticated transport: %s", r.URL.String())
		return nil, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	// The download hop must not use the authenticated transport; swap the
	// default transport temporarily so we can assert no Authorization header
	// is sent to the third-party download host.
	var gotAuth string
	var gotHost string
	origTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotAuth = r.Header.Get("Authorization")
		gotHost = r.URL.Host
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(sbomBody)),
			Request:    r,
		}, nil
	})
	defer func() { http.DefaultTransport = origTransport }()

	report, err := g.FetchDependencyGraphSBOMReport(t.Context(), "owner", "repo", "uuid-1")
	require.NoError(t, err)
	require.NotNil(t, report)
	assert.False(t, report.Pending)
	require.NotNil(t, report.SBOM)
	assert.Equal(t, "SPDXRef-DOCUMENT", report.SBOM.GetSPDXID())
	assert.Equal(t, "objects.example.test", gotHost)
	assert.Empty(t, gotAuth)
}
