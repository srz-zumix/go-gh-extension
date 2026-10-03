package client

import (
	"context"
	"encoding/base64"
	"sort"

	"github.com/shurcooL/githubv4"
)

type createCommitOnBranchMutation struct {
	CreateCommitOnBranch struct {
		Commit struct {
			OID githubv4.GitObjectID `graphql:"oid"`
		}
	} `graphql:"createCommitOnBranch(input: $input)"`
}

// CreateCommitOnBranch adds a commit that writes the given files (path -> content) to a branch
// using the GraphQL createCommitOnBranch mutation. Unlike the Contents REST API, commits
// created this way are signed by GitHub and show up as verified.
// expectedHeadOid must be the current head commit SHA of the branch. It returns the new commit SHA.
func (g *GitHubClient) CreateCommitOnBranch(ctx context.Context, owner, repo, branch, expectedHeadOid, headline string, additions map[string][]byte) (string, error) {
	graphqlClient, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return "", err
	}

	paths := make([]string, 0, len(additions))
	for p := range additions {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	fileAdditions := make([]githubv4.FileAddition, 0, len(paths))
	for _, p := range paths {
		fileAdditions = append(fileAdditions, githubv4.FileAddition{
			Path:     githubv4.String(p),
			Contents: githubv4.Base64String(base64.StdEncoding.EncodeToString(additions[p])),
		})
	}

	input := githubv4.CreateCommitOnBranchInput{
		Branch: githubv4.CommittableBranch{
			RepositoryNameWithOwner: githubv4.NewString(githubv4.String(owner + "/" + repo)),
			BranchName:              githubv4.NewString(githubv4.String(branch)),
		},
		Message:         githubv4.CommitMessage{Headline: githubv4.String(headline)},
		ExpectedHeadOid: githubv4.GitObjectID(expectedHeadOid),
		FileChanges:     &githubv4.FileChanges{Additions: &fileAdditions},
	}

	var mutation createCommitOnBranchMutation
	if err := graphqlClient.Mutate(ctx, &mutation, input, nil); err != nil {
		return "", err
	}
	return string(mutation.CreateCommitOnBranch.Commit.OID), nil
}
