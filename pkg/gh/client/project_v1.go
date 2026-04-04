// Package client provides GitHub API client methods, including GitHub Projects (classic) v1.
package client

import (
	"context"
	"fmt"
	"net/http"
)

// inertiaPreview is the Accept header value required by all GitHub Projects (classic) v1 API endpoints.
// Without this preview header the endpoints return 404 even when Classic Projects are enabled.
const inertiaPreview = "application/vnd.github.inertia-preview+json"

// setInertiaPreview sets the Accept header required by the Classic Projects v1 API on req.
func setInertiaPreview(req *http.Request) {
	req.Header.Set("Accept", inertiaPreview)
}

// ProjectV1 represents a GitHub Project (classic).
type ProjectV1 struct {
	ID         int64    `json:"id"`
	NodeID     string   `json:"node_id"`
	Number     int      `json:"number"`
	Name       string   `json:"name"`
	Body       string   `json:"body"`
	State      string   `json:"state"`
	HTMLURL    string   `json:"html_url"`
	URL        string   `json:"url"`
	ColumnsURL string   `json:"columns_url"`
	Creator    *V1Actor `json:"creator,omitempty"`
}

// ProjectV1Column represents a column in a GitHub Project (classic).
type ProjectV1Column struct {
	ID         int64  `json:"id"`
	NodeID     string `json:"node_id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	ProjectURL string `json:"project_url"`
	CardsURL   string `json:"cards_url"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// ProjectV1Card represents a card in a GitHub Project (classic) column.
type ProjectV1Card struct {
	ID         int64    `json:"id"`
	NodeID     string   `json:"node_id"`
	Note       *string  `json:"note"`
	Archived   bool     `json:"archived"`
	ProjectURL string   `json:"project_url"`
	ColumnURL  string   `json:"column_url"`
	ContentURL *string  `json:"content_url"`
	URL        string   `json:"url"`
	Creator    *V1Actor `json:"creator,omitempty"`
}

// V1Actor represents a minimal GitHub user reference used by classic project responses.
type V1Actor struct {
	Login string `json:"login"`
}

// ListOrgProjectsV1 lists all classic projects for an organization.
func (g *GitHubClient) ListOrgProjectsV1(ctx context.Context, org string) ([]ProjectV1, error) {
	return g.listProjectsV1(ctx, fmt.Sprintf("orgs/%v/projects", org))
}

// ListUserProjectsV1 lists all classic projects for a user.
func (g *GitHubClient) ListUserProjectsV1(ctx context.Context, username string) ([]ProjectV1, error) {
	return g.listProjectsV1(ctx, fmt.Sprintf("users/%v/projects", username))
}

// ListRepoProjectsV1 lists all classic projects for a repository.
func (g *GitHubClient) ListRepoProjectsV1(ctx context.Context, owner, repo string) ([]ProjectV1, error) {
	return g.listProjectsV1(ctx, fmt.Sprintf("repos/%v/%v/projects", owner, repo))
}

// listProjectsV1 is the shared implementation for listing classic projects from any base path.
// A 404 response is treated as an empty list because it indicates that Classic Projects
// are not enabled or do not exist for the owner, rather than a fatal error.
func (g *GitHubClient) listProjectsV1(ctx context.Context, basePath string) ([]ProjectV1, error) {
	var all []ProjectV1
	page := 1
	for {
		u := fmt.Sprintf("%s?state=all&per_page=%d&page=%d", basePath, defaultPerPage, page)
		req, err := g.client.NewRequest("GET", u, nil)
		if err != nil {
			return nil, err
		}
		setInertiaPreview(req)
		var projects []ProjectV1
		resp, err := g.client.Do(ctx, req, &projects)
		if err != nil {
			if resp != nil && resp.StatusCode == 404 {
				return nil, nil
			}
			return nil, err
		}
		all = append(all, projects...)
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return all, nil
}

// ListProjectV1Columns lists all columns in a classic project.
func (g *GitHubClient) ListProjectV1Columns(ctx context.Context, projectID int64) ([]ProjectV1Column, error) {
	var all []ProjectV1Column
	page := 1
	for {
		u := fmt.Sprintf("projects/%d/columns?per_page=%d&page=%d", projectID, defaultPerPage, page)
		req, err := g.client.NewRequest("GET", u, nil)
		if err != nil {
			return nil, err
		}
		setInertiaPreview(req)
		var columns []ProjectV1Column
		resp, err := g.client.Do(ctx, req, &columns)
		if err != nil {
			return nil, err
		}
		all = append(all, columns...)
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return all, nil
}

// createProjectV1 is the shared implementation for creating a classic project at a given API path.
func (g *GitHubClient) createProjectV1(ctx context.Context, path, name, body string) (*ProjectV1, error) {
	input := struct {
		Name string `json:"name"`
		Body string `json:"body"`
	}{Name: name, Body: body}
	req, err := g.client.NewRequest("POST", path, &input)
	if err != nil {
		return nil, err
	}
	setInertiaPreview(req)
	var project ProjectV1
	if _, err := g.client.Do(ctx, req, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// CreateOrgProjectV1 creates a classic project in an organization.
func (g *GitHubClient) CreateOrgProjectV1(ctx context.Context, org, name, body string) (*ProjectV1, error) {
	return g.createProjectV1(ctx, fmt.Sprintf("orgs/%v/projects", org), name, body)
}

// CreateUserProjectV1 creates a classic project for the authenticated user.
func (g *GitHubClient) CreateUserProjectV1(ctx context.Context, name, body string) (*ProjectV1, error) {
	return g.createProjectV1(ctx, "user/projects", name, body)
}

// CreateRepoProjectV1 creates a classic project in a repository.
func (g *GitHubClient) CreateRepoProjectV1(ctx context.Context, owner, repo, name, body string) (*ProjectV1, error) {
	return g.createProjectV1(ctx, fmt.Sprintf("repos/%v/%v/projects", owner, repo), name, body)
}

// DeleteProjectV1 deletes a classic project by its ID.
func (g *GitHubClient) DeleteProjectV1(ctx context.Context, projectID int64) error {
	req, err := g.client.NewRequest("DELETE", fmt.Sprintf("projects/%d", projectID), nil)
	if err != nil {
		return err
	}
	setInertiaPreview(req)
	_, err = g.client.Do(ctx, req, nil)
	return err
}

// CreateProjectV1Column creates a column in a classic project.
func (g *GitHubClient) CreateProjectV1Column(ctx context.Context, projectID int64, name string) (*ProjectV1Column, error) {
	input := struct {
		Name string `json:"name"`
	}{Name: name}
	req, err := g.client.NewRequest("POST", fmt.Sprintf("projects/%d/columns", projectID), &input)
	if err != nil {
		return nil, err
	}
	setInertiaPreview(req)
	var col ProjectV1Column
	if _, err := g.client.Do(ctx, req, &col); err != nil {
		return nil, err
	}
	return &col, nil
}

// CreateProjectV1Card creates a note card in a classic project column.
func (g *GitHubClient) CreateProjectV1Card(ctx context.Context, columnID int64, note string) (*ProjectV1Card, error) {
	input := struct {
		Note string `json:"note"`
	}{Note: note}
	req, err := g.client.NewRequest("POST", fmt.Sprintf("projects/columns/%d/cards", columnID), &input)
	if err != nil {
		return nil, err
	}
	setInertiaPreview(req)
	var card ProjectV1Card
	if _, err := g.client.Do(ctx, req, &card); err != nil {
		return nil, err
	}
	return &card, nil
}

// ListProjectV1Cards lists all cards in a classic project column.
func (g *GitHubClient) ListProjectV1Cards(ctx context.Context, columnID int64) ([]ProjectV1Card, error) {
	var all []ProjectV1Card
	page := 1
	for {
		u := fmt.Sprintf("projects/columns/%d/cards?per_page=%d&page=%d", columnID, defaultPerPage, page)
		req, err := g.client.NewRequest("GET", u, nil)
		if err != nil {
			return nil, err
		}
		setInertiaPreview(req)
		var cards []ProjectV1Card
		resp, err := g.client.Do(ctx, req, &cards)
		if err != nil {
			return nil, err
		}
		all = append(all, cards...)
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return all, nil
}
