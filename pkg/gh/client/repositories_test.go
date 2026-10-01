package client

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-github/v90/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTagsPage(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		want    int
		wantErr bool
	}{
		{name: "non-empty", status: http.StatusOK, body: `[{"name":"v1.0.0"}]`, want: 1},
		{name: "empty", status: http.StatusOK, body: `[]`, want: 0},
		{name: "api error", status: http.StatusInternalServerError, body: `{"message":"boom"}`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath, gotPerPage string
			requests := 0
			transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				requests++
				gotPath = r.URL.Path
				gotPerPage = r.URL.Query().Get("per_page")
				return &http.Response{
					StatusCode: tt.status,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
						"Link":         []string{`<https://api.example.com/repos/owner/repo/tags?per_page=1&page=2>; rel="next"`},
					},
					Body:    io.NopCloser(strings.NewReader(tt.body)),
					Request: r,
				}, nil
			})

			g := newTestClient(t, "https://api.example.com/", transport)
			got, err := g.ListTagsPage(context.Background(), "owner", "repo", &github.ListOptions{PerPage: 1})

			assert.Equal(t, "/repos/owner/repo/tags", gotPath)
			assert.Equal(t, "1", gotPerPage)
			assert.Equal(t, 1, requests)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, tt.want)
		})
	}
}
