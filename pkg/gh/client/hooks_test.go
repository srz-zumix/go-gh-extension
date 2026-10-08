package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v90/github"
)

func TestListHooks(t *testing.T) {
	for _, org := range []bool{true, false} {
		for _, fail := range []bool{false, true} {
			t.Run(fmt.Sprintf("org=%v/fail=%v", org, fail), func(t *testing.T) {
				requests := 0
				path := "/repos/example/repo/hooks"
				if org {
					path = "/orgs/example/hooks"
				}
				server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					requests++
					if request.URL.Path != path || request.URL.Query().Get("per_page") != fmt.Sprint(defaultPerPage) {
						t.Errorf("unexpected URL: %s", request.URL)
					}
					writer.Header().Set("Content-Type", "application/json")
					if request.URL.Query().Get("page") == "2" {
						if fail {
							writer.WriteHeader(http.StatusForbidden)
							fmt.Fprint(writer, `{"message":"forbidden"}`)
						} else {
							fmt.Fprint(writer, `[{"id":2}]`)
						}
						return
					}
					writer.Header().Set("Link", fmt.Sprintf(`<http://%s%s?page=2>; rel="next"`, request.Host, path))
					fmt.Fprint(writer, `[{"id":1}]`)
				}))
				defer server.Close()
				client := newTestClient(t, server.URL, http.DefaultTransport)
				var hooks []*github.Hook
				var err error
				if org {
					hooks, err = client.ListOrgHooks(context.Background(), "example")
				} else {
					hooks, err = client.ListRepoHooks(context.Background(), "example", "repo")
				}
				if requests != 2 {
					t.Fatalf("requests = %d, want 2", requests)
				}
				if fail {
					if _, ok := err.(*github.ErrorResponse); !ok || hooks != nil {
						t.Fatalf("expected unwrapped error and no partial result: hooks=%v, err=%v", hooks, err)
					}
				} else if err != nil || len(hooks) != 2 || hooks[0].GetID() != 1 || hooks[1].GetID() != 2 {
					t.Fatalf("hooks=%v, err=%v", hooks, err)
				}
			})
		}
	}
}
