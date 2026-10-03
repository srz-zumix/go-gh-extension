package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCommitOnBranch(t *testing.T) {
	type fileAddition struct {
		Path     string `json:"path"`
		Contents string `json:"contents"`
	}
	type requestInput struct {
		Branch struct {
			RepositoryNameWithOwner string `json:"repositoryNameWithOwner"`
			BranchName              string `json:"branchName"`
		} `json:"branch"`
		Message struct {
			Headline string `json:"headline"`
		} `json:"message"`
		ExpectedHeadOid string `json:"expectedHeadOid"`
		FileChanges     struct {
			Additions []fileAddition `json:"additions"`
		} `json:"fileChanges"`
	}

	var input requestInput
	response := `{"data":{"createCommitOnBranch":{"commit":{"oid":"deadbeef"}}}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Query     string `json:"query"`
			Variables struct {
				Input requestInput `json:"input"`
			} `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		assert.Contains(t, request.Query, "createCommitOnBranch(input: $input)")
		input = request.Variables.Input
		_, err := w.Write([]byte(response))
		assert.NoError(t, err)
	}))
	defer server.Close()

	g := newTestClient(t, server.URL, server.Client().Transport)
	g.graphql = githubv4.NewEnterpriseClient(server.URL, server.Client())

	additions := map[string][]byte{
		"b.txt":     []byte("bbb"),
		"a/dir.txt": []byte("aaa"),
	}
	oid, err := g.CreateCommitOnBranch(context.Background(), "owner", "repo", "topic", "cafebabe", "headline", additions)
	require.NoError(t, err)
	assert.Equal(t, "deadbeef", oid)
	assert.Equal(t, "owner/repo", input.Branch.RepositoryNameWithOwner)
	assert.Equal(t, "topic", input.Branch.BranchName)
	assert.Equal(t, "headline", input.Message.Headline)
	assert.Equal(t, "cafebabe", input.ExpectedHeadOid)
	assert.Equal(t, []fileAddition{
		{Path: "a/dir.txt", Contents: "YWFh"},
		{Path: "b.txt", Contents: "YmJi"},
	}, input.FileChanges.Additions)

	// API error propagation
	response = `{"errors":[{"message":"permission denied"}]}`
	_, err = g.CreateCommitOnBranch(context.Background(), "owner", "repo", "topic", "cafebabe", "headline", additions)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}
