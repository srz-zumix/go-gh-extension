package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-github/v93/github"
	"github.com/stretchr/testify/require"
)

type stepLogResponseBody struct {
	io.Reader
	closed bool
}

func (b *stepLogResponseBody) Close() error {
	b.closed = true
	return nil
}

func TestGetWorkflowJobStepLogs(t *testing.T) {
	for _, tt := range []struct {
		name     string
		status   int
		location string
		wantErr  bool
	}{
		{"redirect", 302, "https://storage.example.invalid/step?signature=fixture", false},
		{"missing location", 302, "", true},
		{"invalid location", 302, "https://[invalid", true},
		{"unexpected success", 200, "", true},
		{"unavailable", 404, "", true},
		{"expired", 410, "", true},
		{"unauthorized", 401, "", true},
		{"forbidden", 403, "", true},
		{"rate limited", 429, "", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			body := &stepLogResponseBody{Reader: strings.NewReader(`{"message":"fixture error"}`)}
			calls := 0
			g := newTestClient(t, "https://github.example.invalid/api/v3/", roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				require.Equal(t, "GET", r.Method)
				require.Equal(t, "/api/v3/repos/example-owner/example-repo/actions/jobs/1001/steps/27/logs", r.URL.Path)
				return &http.Response{
					StatusCode: tt.status, Header: http.Header{"Location": []string{tt.location}},
					Body: body, Request: r,
				}, nil
			}))
			url, err := g.GetWorkflowJobStepLogs(t.Context(), "example-owner", "example-repo", 1001, 27)
			if tt.wantErr {
				require.Error(t, err)
				require.Empty(t, url)
				if tt.status >= 400 {
					var response *github.ErrorResponse
					require.ErrorAs(t, err, &response)
					require.Equal(t, tt.status, response.Response.StatusCode)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.location, url)
			}
			require.Equal(t, 1, calls)
			require.True(t, body.closed)
			require.Nil(t, g.client.Client().CheckRedirect)
		})
	}
}

func TestGetWorkflowJobStepLogsCancellation(t *testing.T) {
	g := newTestClient(t, "https://api.example.invalid/", roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, r.Context().Err()
	}))
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := g.GetWorkflowJobStepLogs(ctx, "example-owner", "example-repo", 1001, 27)
	require.True(t, errors.Is(err, context.Canceled))
}
