package client

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasTags(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		want    bool
		wantErr bool
	}{
		{name: "non-empty", status: http.StatusOK, body: `[{"name":"v1.0.0"}]`, want: true},
		{name: "empty", status: http.StatusOK, body: `[]`, want: false},
		{name: "api error", status: http.StatusInternalServerError, body: `{"message":"boom"}`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath, gotPerPage string
			transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				gotPath = r.URL.Path
				gotPerPage = r.URL.Query().Get("per_page")
				return &http.Response{
					StatusCode: tt.status,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(tt.body)),
					Request:    r,
				}, nil
			})

			g := newTestClient(t, "https://api.example.com/", transport)
			got, err := g.HasTags(context.Background(), "owner", "repo")

			assert.Equal(t, "/repos/owner/repo/tags", gotPath)
			assert.Equal(t, "1", gotPerPage)
			if tt.wantErr {
				require.Error(t, err)
				assert.False(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
