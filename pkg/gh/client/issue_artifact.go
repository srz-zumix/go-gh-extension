package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v93/github"
)

// issueArtifactAPIVersion is the REST API version that serves the issue artifacts endpoints.
const issueArtifactAPIVersion = "2026-03-10"

// Issue artifact types accepted by the REST API.
const (
	IssueArtifactTypeGeneric      = "generic"
	IssueArtifactTypeLink         = "link"
	IssueArtifactTypePlan         = "plan"
	IssueArtifactTypeEmbeddedHTML = "embedded_html"
)

// IssueArtifactActor is the account that created or updated an issue artifact.
type IssueArtifactActor struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
}

// IssueArtifact is a named piece of content attached to an issue.
type IssueArtifact struct {
	ID             int64               `json:"id"`
	HTMLURL        string              `json:"html_url,omitempty"`
	Number         int                 `json:"number"`
	Type           string              `json:"type"`
	Name           string              `json:"name"`
	Body           string              `json:"body"`
	BodyHTML       *string             `json:"body_html,omitempty"`
	Description    string              `json:"description,omitempty"`
	Creator        *IssueArtifactActor `json:"creator,omitempty"`
	UpdatedByActor *IssueArtifactActor `json:"updated_by_actor,omitempty"`
	CreatedAt      *time.Time          `json:"created_at,omitempty"`
	UpdatedAt      *time.Time          `json:"updated_at,omitempty"`
}

// IssueArtifactRequest is the request body to create or update an issue artifact.
// Type is only honored on creation. Body is a pointer so that an explicitly empty body
// (pointer to "") can be distinguished from an omitted one (nil).
type IssueArtifactRequest struct {
	Type string  `json:"type,omitempty"`
	Name string  `json:"name,omitempty"`
	Body *string `json:"body,omitempty"`
}

type issueArtifactListOptions struct {
	Type string `url:"type,omitempty"`
	github.ListOptions
}

// ListIssueArtifacts lists all artifacts attached to an issue. artifactType filters by type when not empty.
func (g *GitHubClient) ListIssueArtifacts(ctx context.Context, owner, repo string, number int, artifactType string) ([]*IssueArtifact, error) {
	opt := &issueArtifactListOptions{Type: artifactType, ListOptions: github.ListOptions{PerPage: defaultPerPage}}
	var all []*IssueArtifact
	for {
		u, err := addOptions(fmt.Sprintf("repos/%v/%v/issues/%d/artifacts", owner, repo, number), opt)
		if err != nil {
			return nil, err
		}
		req, err := g.client.NewRequest(ctx, "GET", u, nil, github.WithVersion(issueArtifactAPIVersion))
		if err != nil {
			return nil, err
		}
		var page []*IssueArtifact
		resp, err := g.client.Do(req, &page)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return all, nil
}

// CreateIssueArtifact creates an artifact attached to an issue.
func (g *GitHubClient) CreateIssueArtifact(ctx context.Context, owner, repo string, number int, artifact *IssueArtifactRequest) (*IssueArtifact, error) {
	u := fmt.Sprintf("repos/%v/%v/issues/%d/artifacts", owner, repo, number)
	req, err := g.client.NewRequest(ctx, "POST", u, artifact, github.WithVersion(issueArtifactAPIVersion))
	if err != nil {
		return nil, err
	}
	created := new(IssueArtifact)
	if _, err := g.client.Do(req, created); err != nil {
		return nil, err
	}
	return created, nil
}

// UpdateIssueArtifact updates the name and/or body of an artifact attached to an issue.
func (g *GitHubClient) UpdateIssueArtifact(ctx context.Context, owner, repo string, number, artifactNumber int, artifact *IssueArtifactRequest) (*IssueArtifact, error) {
	u := fmt.Sprintf("repos/%v/%v/issues/%d/artifacts/%d", owner, repo, number, artifactNumber)
	body := &IssueArtifactRequest{Name: artifact.Name, Body: artifact.Body}
	req, err := g.client.NewRequest(ctx, "PATCH", u, body, github.WithVersion(issueArtifactAPIVersion))
	if err != nil {
		return nil, err
	}
	updated := new(IssueArtifact)
	if _, err := g.client.Do(req, updated); err != nil {
		return nil, err
	}
	return updated, nil
}
