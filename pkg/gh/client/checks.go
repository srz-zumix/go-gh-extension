package client

import (
	"context"
	"strings"

	"github.com/google/go-github/v79/github"
	"github.com/shurcooL/githubv4"
)

// CheckRun represents a GitHub check run that is part of a check suite.
type CheckRun struct {
	github.CheckRun
	WorkflowName *string `json:"workflow_name,omitempty"`
	IsRequired   *bool   `json:"is_required,omitempty"`
}

func (c *CheckRun) GetFullName() string {
	if c.WorkflowName != nil && *c.WorkflowName != "" {
		return *c.WorkflowName + " / " + c.GetName()
	}
	return c.GetName()
}

// ListCheckRunsForRef retrieves all check runs for a specific Git reference.
func (g *GitHubClient) ListCheckRunsForRef(ctx context.Context, owner string, repo string, ref string, options *github.ListCheckRunsOptions) (*github.ListCheckRunsResults, error) {
	opt := github.ListCheckRunsOptions{ListOptions: github.ListOptions{PerPage: defaultPerPage}}
	if options != nil {
		opt = *options
		opt.PerPage = defaultPerPage
	}

	var allCheckRuns []*github.CheckRun
	for {
		checkRuns, resp, err := g.client.Checks.ListCheckRunsForRef(ctx, owner, repo, ref, &opt)
		if err != nil {
			return nil, err
		}
		allCheckRuns = append(allCheckRuns, checkRuns.CheckRuns...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	total := len(allCheckRuns)
	return &github.ListCheckRunsResults{
		Total:     &total,
		CheckRuns: allCheckRuns,
	}, nil
}

// ListCheckSuitesForRef retrieves all check suites for a specific Git reference.
func (g *GitHubClient) ListCheckSuitesForRef(ctx context.Context, owner string, repo string, ref string, options *github.ListCheckSuiteOptions) (*github.ListCheckSuiteResults, error) {
	opt := github.ListCheckSuiteOptions{ListOptions: github.ListOptions{PerPage: defaultPerPage}}
	if options != nil {
		opt = *options
		opt.PerPage = defaultPerPage
	}

	var allCheckSuites []*github.CheckSuite
	for {
		checkSuites, resp, err := g.client.Checks.ListCheckSuitesForRef(ctx, owner, repo, ref, &opt)
		if err != nil {
			return nil, err
		}
		allCheckSuites = append(allCheckSuites, checkSuites.CheckSuites...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	total := len(allCheckSuites)
	return &github.ListCheckSuiteResults{
		Total:       &total,
		CheckSuites: allCheckSuites,
	}, nil
}

// ListCheckRunsForRefWithGraphQL retrieves all check runs for a specific Git reference using GraphQL.
// This function includes the isRequired field which is not available in the REST API.
// The prNumber parameter is required to determine if a check run is required for the pull request.
func (g *GitHubClient) ListCheckRunsForRefWithGraphQL(ctx context.Context, owner string, repo string, ref string, prNumber int) ([]*CheckRun, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}

	var query struct {
		Repository struct {
			Object struct {
				Commit struct {
					StatusCheckRollup struct {
						Contexts struct {
							Nodes []struct {
								CheckRun struct {
									DatabaseID  githubv4.Int
									ID          githubv4.String
									Name        githubv4.String
									Status      githubv4.String
									Conclusion  githubv4.String
									IsRequired  githubv4.Boolean `graphql:"isRequired(pullRequestNumber: $prNumber)"`
									StartedAt   githubv4.DateTime
									CompletedAt githubv4.DateTime
									DetailsURL  githubv4.URI
									URL         githubv4.URI
									ExternalID  githubv4.String
									Title       githubv4.String
									Text        githubv4.String
									Summary     githubv4.String
									Annotations struct {
										TotalCount githubv4.Int
									}
									CheckSuite struct {
										DatabaseID githubv4.Int
										App        struct {
											ID          githubv4.String
											DatabaseID  githubv4.Int
											Name        githubv4.String
											Slug        githubv4.String
											Description githubv4.String
											URL         githubv4.URI
											CreatedAt   githubv4.DateTime
											UpdatedAt   githubv4.DateTime
										}
										WorkflowRun struct {
											DatabaseID githubv4.Int
											Workflow   struct {
												Name githubv4.String
											}
										}
									}
								} `graphql:"... on CheckRun"`
							}
							PageInfo struct {
								HasNextPage githubv4.Boolean
								EndCursor   githubv4.String
							}
						} `graphql:"contexts(first: 100, after: $cursor)"`
					}
				} `graphql:"... on Commit"`
			} `graphql:"object(expression: $ref)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":    githubv4.String(owner),
		"name":     githubv4.String(repo),
		"ref":      githubv4.String(ref),
		"prNumber": githubv4.Int(prNumber),
		"cursor":   (*githubv4.String)(nil),
	}

	var allCheckRuns []*CheckRun

	// Loop through status check rollup contexts with pagination
	for {
		if err := graphql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}

		for _, contextNode := range query.Repository.Object.Commit.StatusCheckRollup.Contexts.Nodes {
			checkRunNode := contextNode.CheckRun

			// Convert GraphQL CheckRun to REST API CheckRun
			id := int64(checkRunNode.DatabaseID)
			name := string(checkRunNode.Name)
			status := strings.ToLower(string(checkRunNode.Status))
			conclusion := strings.ToLower(string(checkRunNode.Conclusion))
			detailsURL := checkRunNode.DetailsURL.String()
			htmlURL := checkRunNode.URL.String()
			externalID := string(checkRunNode.ExternalID)
			startedAt := github.Timestamp{Time: checkRunNode.StartedAt.Time}
			completedAt := github.Timestamp{Time: checkRunNode.CompletedAt.Time}
			isRequired := bool(checkRunNode.IsRequired)

			// Output fields
			title := string(checkRunNode.Title)
			summary := string(checkRunNode.Summary)
			text := string(checkRunNode.Text)
			annotationsCount := int(checkRunNode.Annotations.TotalCount)

			checkRun := &CheckRun{
				CheckRun: github.CheckRun{
					ID:          &id,
					Name:        &name,
					Status:      &status,
					Conclusion:  &conclusion,
					DetailsURL:  &detailsURL,
					HTMLURL:     &htmlURL,
					ExternalID:  &externalID,
					StartedAt:   &startedAt,
					CompletedAt: &completedAt,
					Output: &github.CheckRunOutput{
						AnnotationsCount: &annotationsCount,
					},
				},
				IsRequired: &isRequired,
			}

			if title != "" {
				checkRun.Output.Title = &title
			}
			if summary != "" {
				checkRun.Output.Summary = &summary
			}
			if text != "" {
				checkRun.Output.Text = &text
			}

			// Convert CheckSuite information
			if checkRunNode.CheckSuite.DatabaseID != 0 {
				checkSuiteID := int64(checkRunNode.CheckSuite.DatabaseID)
				checkRun.CheckSuite = &github.CheckSuite{
					ID: &checkSuiteID,
				}
				workflowName := string(checkRunNode.CheckSuite.WorkflowRun.Workflow.Name)
				checkRun.WorkflowName = &workflowName
			}

			// Convert App information
			if checkRunNode.CheckSuite.App.Name != "" {
				appID := int64(checkRunNode.CheckSuite.App.DatabaseID)
				appName := string(checkRunNode.CheckSuite.App.Name)
				appSlug := string(checkRunNode.CheckSuite.App.Slug)
				appDescription := string(checkRunNode.CheckSuite.App.Description)
				appURL := checkRunNode.CheckSuite.App.URL.String()
				appCreatedAt := github.Timestamp{Time: checkRunNode.CheckSuite.App.CreatedAt.Time}
				appUpdatedAt := github.Timestamp{Time: checkRunNode.CheckSuite.App.UpdatedAt.Time}

				checkRun.App = &github.App{
					ID:          &appID,
					Name:        &appName,
					Slug:        &appSlug,
					Description: &appDescription,
					ExternalURL: &appURL,
					CreatedAt:   &appCreatedAt,
					UpdatedAt:   &appUpdatedAt,
				}
			}

			allCheckRuns = append(allCheckRuns, checkRun)
		}

		if !query.Repository.Object.Commit.StatusCheckRollup.Contexts.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Repository.Object.Commit.StatusCheckRollup.Contexts.PageInfo.EndCursor)
	}

	return allCheckRuns, nil
}

// GetCheckRun retrieves a specific check run.
func (g *GitHubClient) GetCheckRun(ctx context.Context, owner string, repo string, checkRunID int64) (*CheckRun, error) {
	checkRun, _, err := g.client.Checks.GetCheckRun(ctx, owner, repo, checkRunID)
	if err != nil {
		return nil, err
	}
	return &CheckRun{CheckRun: *checkRun}, nil
}
