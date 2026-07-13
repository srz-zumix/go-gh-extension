package client

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteCodeQLDatabase(t *testing.T) {
	tests := []struct {
		name         string
		language     string
		expectedPath string
	}{
		{
			name:         "plain language",
			language:     "go",
			expectedPath: "/repos/owner/repo/code-scanning/codeql/databases/go",
		},
		{
			name:         "language with slash is escaped into a single path segment",
			language:     "lang/evil",
			expectedPath: "/repos/owner/repo/code-scanning/codeql/databases/lang%2Fevil",
		},
		{
			name:         "language with space is escaped",
			language:     "foo bar",
			expectedPath: "/repos/owner/repo/code-scanning/codeql/databases/foo%20bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath string
			tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				gotMethod = r.Method
				gotPath = r.URL.EscapedPath()
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("")),
					Request:    r,
				}, nil
			})
			g := newTestClient(t, "https://api.github.com/", tr)

			err := g.DeleteCodeQLDatabase(t.Context(), "owner", "repo", tt.language)
			require.NoError(t, err)
			assert.Equal(t, http.MethodDelete, gotMethod)
			assert.Equal(t, tt.expectedPath, gotPath)
		})
	}
}
