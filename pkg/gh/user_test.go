package gh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-github/v90/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

func TestIsObfuscatedManagedUserLogin(t *testing.T) {
	tests := []struct {
		login string
		want  bool
	}{
		{"0123456789abcdef0123456789abcd_example", true},
		{"0123456789ABCDEF0123456789ABCD_EXAMPLE", true},
		{"octocat_example", false},
		{"octocat", false},
		{"0123456789abcdef0123456789abcd", false},
		{"abcdef_example", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsObfuscatedManagedUserLogin(tt.login); got != tt.want {
			t.Errorf("IsObfuscatedManagedUserLogin(%q) = %v, want %v", tt.login, got, tt.want)
		}
	}
}

func TestSuspendedUserFilters(t *testing.T) {
	now := time.Now()
	obfuscated := &GitHubUser{Login: github.Ptr("0123456789abcdef0123456789abcd_example")}
	suspendedAt := &GitHubUser{Login: github.Ptr("alice"), SuspendedAt: &github.Timestamp{Time: now}}
	active := &GitHubUser{Login: github.Ptr("bob_example")}
	users := []*GitHubUser{obfuscated, suspendedAt, active}

	if got := CollectSuspendedUsers(users); len(got) != 2 || got[0] != obfuscated || got[1] != suspendedAt {
		t.Errorf("CollectSuspendedUsers() = %v", got)
	}
	if got := ExcludeSuspendedUsers(users); len(got) != 1 || got[0] != active {
		t.Errorf("ExcludeSuspendedUsers() = %v", got)
	}
	if IsSuspendedUser(nil) {
		t.Error("IsSuspendedUser(nil) = true, want false")
	}
}

func TestUpdateUsersAfterSuspendedFilter(t *testing.T) {
	login := "0123456789abcdef0123456789abcd_example"
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests++
		if request.URL.Path != "/users/"+login {
			t.Errorf("unexpected user request: %s", request.URL.Path)
			http.NotFound(writer, request)
			return
		}
		if err := json.NewEncoder(writer).Encode(&GitHubUser{
			Login: github.Ptr(login),
			Name:  github.Ptr("Alice Example"),
			Email: github.Ptr("alice@example.com"),
		}); err != nil {
			t.Errorf("encode user: %v", err)
		}
	}))
	defer server.Close()
	githubClient := redirectingClient(t, server)
	users := []*GitHubUser{
		{Login: github.Ptr("bob_example")},
		{Login: github.Ptr(login), RoleName: github.Ptr("member")},
	}

	users, err := UpdateUsersForSuspension(context.Background(), githubClient, users)
	if err != nil {
		t.Fatal(err)
	}
	if requests != 0 {
		t.Fatalf("suspension detection made %d requests, want 0", requests)
	}
	users, err = UpdateUsers(context.Background(), githubClient, CollectSuspendedUsers(users))
	if err != nil {
		t.Fatal(err)
	}
	if requests != 1 || len(users) != 1 {
		t.Fatalf("got %d requests and %d users, want 1 each", requests, len(users))
	}
	if users[0].GetName() != "Alice Example" || users[0].GetEmail() != "alice@example.com" {
		t.Errorf("user profile not populated: %v", users[0])
	}
	if users[0].GetLogin() != login || users[0].GetRoleName() != "member" || !IsSuspendedUser(users[0]) {
		t.Errorf("user identity, role or suspension changed: %v", users[0])
	}
}

// enterpriseRedirectingClient returns a GitHubClient whose base URL points at a
// GitHub Enterprise Server host while every request is redirected to srv.
func enterpriseRedirectingClient(t *testing.T, srv *httptest.Server) *GitHubClient {
	t.Helper()
	target, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse server URL: %v", err)
	}
	base := srv.Client()
	base.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req = req.Clone(req.Context())
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		return http.DefaultTransport.RoundTrip(req)
	})
	baseURL := "https://ghes.example.com/api/v3/"
	gc, err := github.NewClient(github.WithHTTPClient(base), github.WithURLs(&baseURL, nil))
	if err != nil {
		t.Fatalf("new go-github client: %v", err)
	}
	g, err := client.NewClient(gc)
	if err != nil {
		t.Fatalf("new github client: %v", err)
	}
	return g
}

func TestUpdateUsersForSuspensionOnEnterprise(t *testing.T) {
	login := "alice"
	suspendedAt := time.Now().UTC().Truncate(time.Second)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests++
		if request.URL.Path != "/api/v3/users/"+login {
			t.Errorf("unexpected user request: %s", request.URL.Path)
			http.NotFound(writer, request)
			return
		}
		if err := json.NewEncoder(writer).Encode(&GitHubUser{
			Login:       github.Ptr(login),
			SuspendedAt: &github.Timestamp{Time: suspendedAt},
		}); err != nil {
			t.Errorf("encode user: %v", err)
		}
	}))
	defer server.Close()

	users, err := UpdateUsersForSuspension(context.Background(), enterpriseRedirectingClient(t, server), []*GitHubUser{
		{Login: github.Ptr(login), RoleName: github.Ptr("member")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if requests != 1 {
		t.Fatalf("suspension detection made %d requests, want 1", requests)
	}
	if !IsSuspendedUser(users[0]) {
		t.Errorf("suspended_at not populated: %v", users[0])
	}
	if users[0].GetRoleName() != "member" {
		t.Errorf("role changed: %v", users[0])
	}
}
