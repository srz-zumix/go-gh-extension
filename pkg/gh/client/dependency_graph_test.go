package client

import (
	"io"
	"net/http"
	"net/http/httptest"
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
	const sbomBody = `{"SPDXID":"SPDXRef-DOCUMENT","spdxVersion":"SPDX-2.3"}`

	// Real download endpoint hosted on a local test server. Using httptest.Server
	// avoids mutating process-wide globals such as http.DefaultTransport.
	var gotAuth string
	var gotDownload bool
	downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotDownload = true
		gotAuth = r.Header.Get("Authorization")
		_, _ = io.WriteString(w, sbomBody)
	}))
	defer downloadServer.Close()

	// The authenticated transport injects an Authorization header on the GitHub
	// API hop and redirects to the download server. This proves credentials are
	// present on the authenticated hop and absent on the download hop.
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host == "api.github.com" {
			return &http.Response{
				StatusCode: http.StatusFound,
				Header:     http.Header{"Location": []string{downloadServer.URL + "/download/report.json"}},
				Body:       io.NopCloser(strings.NewReader(``)),
				Request:    r,
			}, nil
		}
		t.Fatalf("unexpected request to authenticated transport: %s", r.URL.String())
		return nil, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	report, err := g.FetchDependencyGraphSBOMReport(t.Context(), "owner", "repo", "uuid-1")
	require.NoError(t, err)
	require.NotNil(t, report)
	assert.False(t, report.Pending)
	require.NotNil(t, report.SBOM)
	assert.Equal(t, "SPDXRef-DOCUMENT", report.SBOM.GetSPDXID())
	assert.True(t, gotDownload)
	assert.Empty(t, gotAuth)
}
