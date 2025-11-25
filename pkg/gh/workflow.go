package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
)

type LogUrlFetcher interface {
	FetchLogURL(ctx context.Context, g *GitHubClient, repo repository.Repository, maxRedirects int) (string, error)
}

type JobLogUrlFetcher struct {
	CheckRunID int64
}

func (j *JobLogUrlFetcher) FetchLogURL(ctx context.Context, g *GitHubClient, repo repository.Repository, maxRedirects int) (string, error) {
	return GetWorkflowJobLogsURL(ctx, g, repo, j.CheckRunID, maxRedirects)
}

type RunLogUrlFetcher struct {
	RunID   int64
	Attempt *int
}

func (r *RunLogUrlFetcher) FetchLogURL(ctx context.Context, g *GitHubClient, repo repository.Repository, maxRedirects int) (string, error) {
	if r.Attempt != nil {
		return GetWorkflowRunAttemptLogsURL(ctx, g, repo, r.RunID, *r.Attempt, maxRedirects)
	}
	return GetWorkflowRunLogsURL(ctx, g, repo, r.RunID, maxRedirects)
}

func GetLogUrlFetcher(context any) LogUrlFetcher {
	switch v := context.(type) {
	case int64:
		return &JobLogUrlFetcher{CheckRunID: v}
	case *int64:
		if v == nil {
			return nil
		}
		return &JobLogUrlFetcher{CheckRunID: *v}
	case CheckRun:
		if v.ID == nil {
			return nil
		}
		return &JobLogUrlFetcher{CheckRunID: *v.ID}
	case *CheckRun:
		if v == nil {
			return nil
		}
		if v.ID == nil {
			return nil
		}
		return &JobLogUrlFetcher{CheckRunID: *v.ID}
	case github.WorkflowJob:
		if v.RunID == nil {
			return nil
		}
		if v.RunAttempt == nil {
			return &RunLogUrlFetcher{RunID: *v.RunID}
		}
		attempt := int(*v.RunAttempt)
		return &RunLogUrlFetcher{RunID: *v.RunID, Attempt: &attempt}
	case *github.WorkflowJob:
		if v == nil {
			return nil
		}
		if v.RunID == nil {
			return nil
		}
		if v.RunAttempt == nil {
			return &RunLogUrlFetcher{RunID: *v.RunID}
		}
		attempt := int(*v.RunAttempt)
		return &RunLogUrlFetcher{RunID: *v.RunID, Attempt: &attempt}
	case github.WorkflowRun:
		// Check for nil ID to avoid panic
		if v.ID == nil {
			return nil
		}
		return &RunLogUrlFetcher{RunID: *v.ID, Attempt: v.RunAttempt}
	case *github.WorkflowRun:
		if v == nil {
			return nil
		}
		// Check for nil ID to avoid panic
		if v.ID == nil {
			return nil
		}
		return &RunLogUrlFetcher{RunID: *v.ID, Attempt: v.RunAttempt}
	default:
		return nil
	}
}

func GetWorkflowRunLogUrlFetcher(context any) LogUrlFetcher {
	switch v := context.(type) {
	case int64:
		return &RunLogUrlFetcher{RunID: v}
	case *int64:
		if v == nil {
			return nil
		}
		return &RunLogUrlFetcher{RunID: *v}
	case CheckRun:
		id, err := ExtractRunIDFromCheckRunAsInt64(&v)
		if err != nil {
			return nil
		}
		return &RunLogUrlFetcher{RunID: id}
	case *CheckRun:
		id, err := ExtractRunIDFromCheckRunAsInt64(v)
		if err != nil {
			return nil
		}
		return &RunLogUrlFetcher{RunID: id}
	default:
		return GetLogUrlFetcher(context)
	}
}
