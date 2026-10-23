package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v93/github"
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

	// The API transport plays the role of the authenticated client: it attaches a
	// dummy token and redirects to the download server. The token must reach the
	// API hop and must not reach the download hop.
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host == "api.github.com" {
			r.Header.Set("Authorization", "token ghp_dummy_token")
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

func TestFetchDependencyGraphSBOMReport_SameOriginRedirectThenFound(t *testing.T) {
	const sbomBody = `{"SPDXID":"SPDXRef-DOCUMENT","spdxVersion":"SPDX-2.3"}`
	downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, sbomBody)
	}))
	defer downloadServer.Close()

	var hops []string
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		hops = append(hops, r.URL.Path)
		if r.URL.Path == "/repos/owner/repo/dependency-graph/sbom/fetch-report/uuid-1" {
			return &http.Response{
				StatusCode: http.StatusMovedPermanently,
				Header:     http.Header{"Location": []string{"https://api.github.com/repos/owner/repo/dependency-graph/sbom/fetch-report/uuid-2"}},
				Body:       io.NopCloser(strings.NewReader(``)),
				Request:    r,
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusFound,
			Header:     http.Header{"Location": []string{downloadServer.URL + "/download/report.json"}},
			Body:       io.NopCloser(strings.NewReader(``)),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	report, err := g.FetchDependencyGraphSBOMReport(t.Context(), "owner", "repo", "uuid-1")
	require.NoError(t, err)
	require.NotNil(t, report.SBOM)
	assert.Equal(t, "SPDXRef-DOCUMENT", report.SBOM.GetSPDXID())
	assert.Equal(t, []string{
		"/repos/owner/repo/dependency-graph/sbom/fetch-report/uuid-1",
		"/repos/owner/repo/dependency-graph/sbom/fetch-report/uuid-2",
	}, hops)
}

func TestFetchDependencyGraphSBOMReport_CrossOriginRedirectRejected(t *testing.T) {
	var downloadHits int
	downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		downloadHits++
		_, _ = io.WriteString(w, `{}`)
	}))
	defer downloadServer.Close()

	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusMovedPermanently,
			Header:     http.Header{"Location": []string{downloadServer.URL + "/download/report.json"}},
			Body:       io.NopCloser(strings.NewReader(``)),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	report, err := g.FetchDependencyGraphSBOMReport(t.Context(), "owner", "repo", "uuid-1")
	require.Error(t, err)
	assert.Nil(t, report)
	assert.Equal(t, 0, downloadHits)
}

func TestFetchDependencyGraphSBOMReport_DownloadHost202IsNotPending(t *testing.T) {
	downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer downloadServer.Close()

	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusFound,
			Header:     http.Header{"Location": []string{downloadServer.URL + "/download/report.json"}},
			Body:       io.NopCloser(strings.NewReader(``)),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	report, err := g.FetchDependencyGraphSBOMReport(t.Context(), "owner", "repo", "uuid-1")
	require.Error(t, err)
	assert.Nil(t, report)
	_, ok := errors.AsType[*github.AcceptedError](err)
	assert.True(t, ok, "expected *github.AcceptedError, got %T", err)
}

const testSBOMFetchPath = "/repos/owner/repo/dependency-graph/sbom/fetch-report/uuid-1"

func TestGetDependencyGraphSBOM_PollsUntilReady(t *testing.T) {
	const sbomBody = `{"SPDXID":"SPDXRef-DOCUMENT","spdxVersion":"SPDX-2.3"}`
	downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, sbomBody)
	}))
	defer downloadServer.Close()

	var generateCalls, fetchCalls int
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/repos/owner/repo/dependency-graph/sbom/generate-report":
			generateCalls++
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"sbom_url":"https://api.github.com` + testSBOMFetchPath + `"}`)),
				Request:    r,
			}, nil
		case testSBOMFetchPath:
			fetchCalls++
			if fetchCalls < 3 {
				return &http.Response{StatusCode: http.StatusAccepted, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(``)), Request: r}, nil
			}
			return &http.Response{
				StatusCode: http.StatusFound,
				Header:     http.Header{"Location": []string{downloadServer.URL + "/download/report.json"}},
				Body:       io.NopCloser(strings.NewReader(``)),
				Request:    r,
			}, nil
		default:
			t.Errorf("unexpected request path: %s", r.URL.Path)
			return &http.Response{StatusCode: http.StatusNotFound, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(``)), Request: r}, nil
		}
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	sbom, err := g.getDependencyGraphSBOM(t.Context(), "owner", "repo", time.Millisecond, 10*time.Second)
	require.NoError(t, err)
	require.NotNil(t, sbom.SBOM)
	assert.Equal(t, "SPDXRef-DOCUMENT", sbom.SBOM.GetSPDXID())
	assert.Equal(t, 1, generateCalls)
	assert.Equal(t, 3, fetchCalls)
}

func TestGetDependencyGraphSBOM_TimesOutWhilePending(t *testing.T) {
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/repos/owner/repo/dependency-graph/sbom/generate-report" {
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"sbom_url":"https://api.github.com` + testSBOMFetchPath + `"}`)), Request: r}, nil
		}
		return &http.Response{StatusCode: http.StatusAccepted, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(``)), Request: r}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	sbom, err := g.getDependencyGraphSBOM(t.Context(), "owner", "repo", time.Millisecond, 30*time.Millisecond)
	require.Error(t, err)
	assert.Nil(t, sbom)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestGetDependencyGraphSBOM_GenerateErrorStopsBeforeFetch(t *testing.T) {
	var fetchCalls int
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/repos/owner/repo/dependency-graph/sbom/generate-report" {
			fetchCalls++
		}
		return &http.Response{StatusCode: http.StatusForbidden, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"message":"forbidden"}`)), Request: r}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	sbom, err := g.getDependencyGraphSBOM(t.Context(), "owner", "repo", time.Millisecond, 10*time.Second)
	require.Error(t, err)
	assert.Nil(t, sbom)
	assert.Equal(t, 0, fetchCalls)
}

func TestGetDependencyGraphSBOM_InvalidGenerationResponse(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "missing sbom_url", body: `{}`},
		{name: "empty body", body: ``},
		{name: "trailing slash without uuid", body: `{"sbom_url":"https://api.github.com/repos/owner/repo/dependency-graph/sbom/fetch-report/"}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var fetchCalls int
			tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path != "/repos/owner/repo/dependency-graph/sbom/generate-report" {
					fetchCalls++
				}
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(tc.body)), Request: r}, nil
			})
			g := newTestClient(t, "https://api.github.com/", tr)

			sbom, err := g.getDependencyGraphSBOM(t.Context(), "owner", "repo", time.Millisecond, 10*time.Second)
			require.Error(t, err)
			assert.Nil(t, sbom)
			assert.Equal(t, 0, fetchCalls)
		})
	}
}

func TestGetDependencyGraphSBOM_NullReportIsError(t *testing.T) {
	downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `null`)
	}))
	defer downloadServer.Close()

	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/repos/owner/repo/dependency-graph/sbom/generate-report" {
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"sbom_url":"https://api.github.com` + testSBOMFetchPath + `"}`)), Request: r}, nil
		}
		return &http.Response{StatusCode: http.StatusFound, Header: http.Header{"Location": []string{downloadServer.URL + "/download/report.json"}}, Body: io.NopCloser(strings.NewReader(``)), Request: r}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	sbom, err := g.getDependencyGraphSBOM(t.Context(), "owner", "repo", time.Millisecond, 10*time.Second)
	require.Error(t, err)
	assert.Nil(t, sbom)
}

func TestSBOMUUIDFromURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		want    string
		wantErr bool
	}{
		{name: "standard", rawURL: "https://api.github.com/repos/owner/repo/dependency-graph/sbom/fetch-report/uuid-1", want: "uuid-1"},
		{name: "query ignored", rawURL: "https://ghe.example.com/api/v3/repos/owner/repo/dependency-graph/sbom/fetch-report/uuid-1?x=1", want: "uuid-1"},
		{name: "escaped id is decoded", rawURL: "https://api.github.com/repos/owner/repo/dependency-graph/sbom/fetch-report/uuid%2D1", want: "uuid-1"},
		{name: "trailing slash", rawURL: "https://api.github.com/repos/owner/repo/dependency-graph/sbom/fetch-report/", wantErr: true},
		{name: "missing segment", rawURL: "https://api.github.com/repos/owner/repo/dependency-graph/sbom/uuid-1", wantErr: true},
		{name: "dot segment", rawURL: "https://api.github.com/repos/owner/repo/dependency-graph/sbom/fetch-report/..", wantErr: true},
		{name: "encoded slash", rawURL: "https://api.github.com/repos/owner/repo/dependency-graph/sbom/fetch-report/a%2Fb", wantErr: true},
		{name: "empty", rawURL: "", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := sbomUUIDFromURL(tc.rawURL)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
