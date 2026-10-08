package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListEnterpriseRunners(t *testing.T) {
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
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			var base string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				calls++
				if req.Method != http.MethodGet || req.URL.Path != "/enterprises/octo/actions/runners" || req.URL.Query().Get("per_page") != "100" {
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
					w.Header().Set("Link", fmt.Sprintf(`<%s/enterprises/octo/actions/runners?per_page=100&page=2>; rel="next"`, base))
					_, _ = fmt.Fprint(w, `{"total_count":2,"runners":[{"id":1,"name":"first","status":"online","busy":true}]}`)
				} else {
					if req.URL.Query().Get("page") != "2" {
						t.Errorf("unexpected page: %s", req.URL.Query().Get("page"))
					}
					_, _ = fmt.Fprint(w, `{"total_count":2,"runners":[{"id":2,"name":"second","status":"offline","busy":false}]}`)
				}
			}))
			defer server.Close()
			base = server.URL
			client := newTestClient(t, base, http.DefaultTransport)
			runners, err := client.ListEnterpriseRunners(context.Background(), "octo")
			if test.errorPage != 0 {
				if err == nil || runners != nil || calls != test.errorPage {
					t.Fatalf("runners = %+v, err = %v, calls = %d", runners, err, calls)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if test.empty {
				if runners == nil || len(runners) != 0 || calls != 1 {
					t.Fatalf("runners = %+v, calls = %d", runners, calls)
				}
				return
			}
			if len(runners) != 2 || calls != 2 {
				t.Fatalf("runners = %+v, calls = %d", runners, calls)
			}
			if runners[0].GetID() != 1 || runners[1].GetID() != 2 || !runners[0].GetBusy() || runners[1].GetStatus() != "offline" {
				t.Fatalf("unexpected runners: %+v", runners)
			}
		})
	}
}
