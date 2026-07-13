package client

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-github/v88/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultSetupConfiguration(t *testing.T) {
	var gotMethod, gotPath string
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.EscapedPath()
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"state":"configured"}`)),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	config, err := g.GetDefaultSetupConfiguration(t.Context(), "owner", "repo")
	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, http.MethodGet, gotMethod)
	assert.Equal(t, "/repos/owner/repo/code-scanning/default-setup", gotPath)
	assert.Equal(t, "configured", config.GetState())
}

func TestUpdateDefaultSetupConfiguration(t *testing.T) {
	const body = `{"run_id":123,"run_url":"https://example.test/run/123"}`

	tests := []struct {
		name       string
		statusCode int
	}{
		{
			name:       "normal 200 response",
			statusCode: http.StatusOK,
		},
		{
			name:       "202 accepted with response body",
			statusCode: http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath string
			var gotBody map[string]any
			tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				gotMethod = r.Method
				gotPath = r.URL.EscapedPath()
				if r.Body != nil {
					require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
				}
				return &http.Response{
					StatusCode: tt.statusCode,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(body)),
					Request:    r,
				}, nil
			})
			g := newTestClient(t, "https://api.github.com/", tr)

			opts := &github.UpdateDefaultSetupConfigurationOptions{State: "configured"}
			result, err := g.UpdateDefaultSetupConfiguration(t.Context(), "owner", "repo", opts)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, http.MethodPatch, gotMethod)
			assert.Equal(t, "/repos/owner/repo/code-scanning/default-setup", gotPath)
			assert.Equal(t, "configured", gotBody["state"])
			assert.Equal(t, int64(123), result.GetRunID())
			assert.Equal(t, "https://example.test/run/123", result.GetRunURL())
		})
	}
}
