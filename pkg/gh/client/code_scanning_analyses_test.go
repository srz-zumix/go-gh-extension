package client

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteAnalysis(t *testing.T) {
	tests := []struct {
		name          string
		confirmDelete bool
		expectedQuery string
	}{
		{
			name:          "without confirm_delete",
			confirmDelete: false,
			expectedQuery: "",
		},
		{
			name:          "with confirm_delete",
			confirmDelete: true,
			expectedQuery: "confirm_delete=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath, gotQuery string
			tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				gotMethod = r.Method
				gotPath = r.URL.EscapedPath()
				gotQuery = r.URL.RawQuery
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"next_analysis_url":"https://example.test/next"}`)),
					Request:    r,
				}, nil
			})
			g := newTestClient(t, "https://api.github.com/", tr)

			result, err := g.DeleteAnalysis(t.Context(), "owner", "repo", 42, tt.confirmDelete)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, http.MethodDelete, gotMethod)
			assert.Equal(t, "/repos/owner/repo/code-scanning/analyses/42", gotPath)
			assert.Equal(t, tt.expectedQuery, gotQuery)
			assert.Equal(t, "https://example.test/next", result.GetNextAnalysisURL())
		})
	}
}
