package client

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-github/v90/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAgentsRepoSecretsPagination(t *testing.T) {
	var gotMethods []string
	var gotPaths []string
	var gotPages []string
	var gotPerPage []string
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotMethods = append(gotMethods, r.Method)
		gotPaths = append(gotPaths, r.URL.EscapedPath())
		gotPages = append(gotPages, r.URL.Query().Get("page"))
		gotPerPage = append(gotPerPage, r.URL.Query().Get("per_page"))
		header := make(http.Header)
		body := `{"total_count":2,"secrets":[{"name":"SECOND"}]}`
		if r.URL.Query().Get("page") == "1" {
			// Advertise a next page via the Link header so resp.NextPage is populated.
			header.Set("Link", `<https://api.github.com/repos/owner/repo/agents/secrets?per_page=100&page=2>; rel="next"`)
			body = `{"total_count":2,"secrets":[{"name":"FIRST"}]}`
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     header,
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	secrets, err := g.ListAgentsRepoSecrets(t.Context(), "owner", "repo")
	require.NoError(t, err)

	// Exactly two requests: page 1 then the advertised page 2.
	require.Len(t, gotPaths, 2)
	assert.Equal(t, []string{http.MethodGet, http.MethodGet}, gotMethods)
	assert.Equal(t, "/repos/owner/repo/agents/secrets", gotPaths[0])
	assert.Equal(t, "/repos/owner/repo/agents/secrets", gotPaths[1])
	assert.Equal(t, []string{"1", "2"}, gotPages)
	assert.Equal(t, []string{"100", "100"}, gotPerPage)

	// Results are aggregated in order.
	require.Len(t, secrets, 2)
	assert.Equal(t, "FIRST", secrets[0].GetName())
	assert.Equal(t, "SECOND", secrets[1].GetName())
}

func TestListAgentsOrgSecretsPath(t *testing.T) {
	var gotPath string
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotPath = r.URL.EscapedPath()
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"total_count":0,"secrets":[]}`)),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	_, err := g.ListAgentsOrgSecrets(t.Context(), "octo-org")
	require.NoError(t, err)
	assert.Equal(t, "/orgs/octo-org/agents/secrets", gotPath)
}

func TestGetAgentsRepoPublicKey(t *testing.T) {
	var gotMethod, gotPath string
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.EscapedPath()
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"key_id":"kid-1","key":"pub"}`)),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	key, err := g.GetAgentsRepoPublicKey(t.Context(), "owner", "repo")
	require.NoError(t, err)
	assert.Equal(t, http.MethodGet, gotMethod)
	assert.Equal(t, "/repos/owner/repo/agents/secrets/public-key", gotPath)
	assert.Equal(t, "kid-1", key.GetKeyID())
}

func TestCreateOrUpdateAgentsRepoSecretBody(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.EscapedPath()
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(``)),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	// Deliberately populate Visibility/SelectedRepositoryIDs to prove they are
	// NOT forwarded to the repository endpoint (only key_id/encrypted_value are).
	eSecret := &github.EncryptedSecret{
		Name:                  "MY_SECRET",
		KeyID:                 "kid-1",
		EncryptedValue:        "enc",
		Visibility:            "all",
		SelectedRepositoryIDs: github.SelectedRepoIDs{1, 2},
	}
	require.NoError(t, g.CreateOrUpdateAgentsRepoSecret(t.Context(), "owner", "repo", eSecret))

	assert.Equal(t, http.MethodPut, gotMethod)
	assert.Equal(t, "/repos/owner/repo/agents/secrets/MY_SECRET", gotPath)
	assert.Equal(t, "kid-1", gotBody["key_id"])
	assert.Equal(t, "enc", gotBody["encrypted_value"])
	assert.NotContains(t, gotBody, "visibility")
	assert.NotContains(t, gotBody, "selected_repository_ids")
	assert.NotContains(t, gotBody, "name")
}

func TestDeleteAgentsRepoSecret(t *testing.T) {
	var gotMethod, gotPath string
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.EscapedPath()
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(``)),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	require.NoError(t, g.DeleteAgentsRepoSecret(t.Context(), "owner", "repo", "MY_SECRET"))
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/repos/owner/repo/agents/secrets/MY_SECRET", gotPath)
}
