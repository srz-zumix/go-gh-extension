package gh

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-github/v90/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTokenScopes(t *testing.T) {
	tests := []struct {
		name         string
		header       http.Header
		status       int
		transportErr error
		wantScopes   []string
		wantOK       bool
		wantErr      bool
	}{
		{
			name:       "populated",
			header:     http.Header{"X-Oauth-Scopes": []string{"repo, read:org,workflow"}},
			status:     http.StatusOK,
			wantScopes: []string{"repo", "read:org", "workflow"},
			wantOK:     true,
		},
		{
			name:       "whitespace and empty entries",
			header:     http.Header{"X-Oauth-Scopes": []string{" repo , , gist ,"}},
			status:     http.StatusOK,
			wantScopes: []string{"repo", "gist"},
			wantOK:     true,
		},
		{
			name:       "present but empty",
			header:     http.Header{"X-Oauth-Scopes": []string{""}},
			status:     http.StatusOK,
			wantScopes: nil,
			wantOK:     true,
		},
		{
			name:       "absent header",
			header:     http.Header{},
			status:     http.StatusOK,
			wantScopes: nil,
			wantOK:     false,
		},
		{
			name:    "api error",
			header:  http.Header{},
			status:  http.StatusUnauthorized,
			wantErr: true,
		},
		{
			name:         "transport error",
			transportErr: errors.New("connection refused"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				gotPath = r.URL.Path
				if tt.transportErr != nil {
					return nil, tt.transportErr
				}
				header := tt.header.Clone()
				header.Set("Content-Type", "application/json")
				return &http.Response{
					StatusCode: tt.status,
					Header:     header,
					Body:       io.NopCloser(strings.NewReader(`{}`)),
					Request:    r,
				}, nil
			})
			gc, err := github.NewClient(github.WithHTTPClient(&http.Client{Transport: transport}))
			require.NoError(t, err)
			g, err := client.NewClient(gc)
			require.NoError(t, err)

			scopes, ok, err := GetTokenScopes(t.Context(), g)

			assert.Equal(t, "/", gotPath)
			if tt.wantErr {
				require.Error(t, err)
				assert.False(t, ok)
				assert.Nil(t, scopes)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantScopes, scopes)
		})
	}
}
