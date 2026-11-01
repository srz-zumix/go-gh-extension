package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// IssueArtifact is a named piece of content attached to an issue.
type IssueArtifact = client.IssueArtifact

// IssueArtifactRequest is the request body to create or update an issue artifact.
type IssueArtifactRequest = client.IssueArtifactRequest

// Issue artifact types accepted by the REST API.
const (
	IssueArtifactTypeGeneric      = client.IssueArtifactTypeGeneric
	IssueArtifactTypeLink         = client.IssueArtifactTypeLink
	IssueArtifactTypePlan         = client.IssueArtifactTypePlan
	IssueArtifactTypeEmbeddedHTML = client.IssueArtifactTypeEmbeddedHTML
)

// ListIssueArtifacts lists the artifacts attached to an issue. artifactType filters by type when not empty.
func ListIssueArtifacts(ctx context.Context, g *GitHubClient, repo repository.Repository, issue any, artifactType string) ([]*IssueArtifact, error) {
	number, err := GetIssueNumber(issue)
	if err != nil {
		return nil, err
	}
	return g.ListIssueArtifacts(ctx, repo.Owner, repo.Name, number, artifactType)
}

// CreateIssueArtifact creates an artifact attached to an issue.
func CreateIssueArtifact(ctx context.Context, g *GitHubClient, repo repository.Repository, issue any, artifact *IssueArtifactRequest) (*IssueArtifact, error) {
	number, err := GetIssueNumber(issue)
	if err != nil {
		return nil, err
	}
	return g.CreateIssueArtifact(ctx, repo.Owner, repo.Name, number, artifact)
}

// UpdateIssueArtifact updates the name and/or body of an artifact attached to an issue.
func UpdateIssueArtifact(ctx context.Context, g *GitHubClient, repo repository.Repository, issue any, artifactNumber int, artifact *IssueArtifactRequest) (*IssueArtifact, error) {
	number, err := GetIssueNumber(issue)
	if err != nil {
		return nil, err
	}
	return g.UpdateIssueArtifact(ctx, repo.Owner, repo.Name, number, artifactNumber, artifact)
}

// FindIssueArtifactByName returns the first artifact on the issue that has the given name and type,
// or nil when none matches.
func FindIssueArtifactByName(artifacts []*IssueArtifact, name, artifactType string) *IssueArtifact {
	for _, a := range artifacts {
		if a == nil || a.Name != name {
			continue
		}
		if artifactType != "" && a.Type != artifactType {
			continue
		}
		return a
	}
	return nil
}

// UpsertIssueArtifact updates the artifact with the same name and type on the issue, or creates it when absent.
// The returned bool reports whether a new artifact was created.
func UpsertIssueArtifact(ctx context.Context, g *GitHubClient, repo repository.Repository, issue any, artifact *IssueArtifactRequest) (*IssueArtifact, bool, error) {
	artifactType := artifact.Type
	if artifactType == "" {
		artifactType = IssueArtifactTypeGeneric
	}
	existing, err := ListIssueArtifacts(ctx, g, repo, issue, artifactType)
	if err != nil {
		return nil, false, err
	}
	if found := FindIssueArtifactByName(existing, artifact.Name, artifactType); found != nil {
		updated, err := UpdateIssueArtifact(ctx, g, repo, issue, found.Number, artifact)
		return updated, false, err
	}
	created, err := CreateIssueArtifact(ctx, g, repo, issue, artifact)
	return created, err == nil, err
}
