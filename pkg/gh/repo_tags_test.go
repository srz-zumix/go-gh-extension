package gh

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v90/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
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
			gc, err := github.NewClient(github.WithHTTPClient(&http.Client{Transport: transport}))
			if err != nil {
				t.Fatalf("new go-github client: %v", err)
			}
			g, err := client.NewClient(gc)
			if err != nil {
				t.Fatalf("new github client: %v", err)
			}

			got, err := HasTags(context.Background(), g, repository.Repository{Host: "github.com", Owner: "owner", Name: "repo"})

			if gotPath != "/repos/owner/repo/tags" {
				t.Errorf("path = %q, want /repos/owner/repo/tags", gotPath)
			}
			if gotPerPage != "1" {
				t.Errorf("per_page = %q, want 1", gotPerPage)
			}
			if tt.wantErr {
				if err == nil {
					t.Fatalf("HasTags() error = nil, want error")
				}
				if got {
					t.Errorf("HasTags() = true on error, want false")
				}
				return
			}
			if err != nil {
				t.Fatalf("HasTags() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("HasTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
