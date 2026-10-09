package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListHostedRunnersPagesAndErrors(t *testing.T) {
	for _, scope := range []string{"orgs", "enterprises"} {
		for _, test := range []struct {
			name      string
			errorPage int
			empty     bool
		}{
			{name: "pagination"},
			{name: "empty", empty: true},
			{name: "first page error", errorPage: 1},
			{name: "second page error", errorPage: 2},
		} {
			t.Run(scope+"/"+test.name, func(t *testing.T) {
				calls := 0
				var base string
				path := "/" + scope + "/octo/actions/hosted-runners"
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					calls++
					if req.Method != http.MethodGet || req.URL.Path != path || req.URL.Query().Get("per_page") != "100" {
						t.Errorf("unexpected request: %s %s", req.Method, req.URL)
					}
					w.Header().Set("Content-Type", "application/json")
					if calls == test.errorPage {
						w.WriteHeader(http.StatusForbidden)
						_, _ = fmt.Fprint(w, `{"message":"Forbidden"}`)
						return
					}
					if test.empty {
						_, _ = fmt.Fprint(w, `{"total_count":0,"runners":[]}`)
						return
					}
					if req.URL.Query().Get("page") == "" {
						w.Header().Set("Link", fmt.Sprintf(`<%s%s?per_page=100&page=2>; rel="next"`, base, path))
						_, _ = fmt.Fprint(w, `{"total_count":2,"runners":[{"id":1,"name":"first"}]}`)
					} else {
						if req.URL.Query().Get("page") != "2" {
							t.Errorf("unexpected page: %s", req.URL.Query().Get("page"))
						}
						_, _ = fmt.Fprint(w, `{"total_count":2,"runners":[{"id":2,"name":"second"}]}`)
					}
				}))
				defer server.Close()
				base = server.URL
				client := newTestClient(t, base, http.DefaultTransport)
				list := client.ListOrgHostedRunners
				if scope == "enterprises" {
					list = client.ListEnterpriseHostedRunners
				}
				runners, err := list(context.Background(), "octo")
				if test.errorPage != 0 {
					if err == nil || runners != nil || calls != test.errorPage {
						t.Fatalf("result = %+v, %v, calls=%d", runners, err, calls)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if test.empty {
					if len(runners) != 0 || calls != 1 {
						t.Fatalf("result = %+v, calls=%d", runners, calls)
					}
					return
				}
				if len(runners) != 2 || calls != 2 {
					t.Fatalf("result = %+v, calls=%d", runners, calls)
				}
				if runners[0].GetID() != 1 || runners[1].GetID() != 2 || runners[0].GetName() != "first" || runners[1].GetName() != "second" {
					t.Fatalf("unexpected runners: %+v", runners)
				}
			})
		}
	}
}
