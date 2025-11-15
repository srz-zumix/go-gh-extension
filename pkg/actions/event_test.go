package actions

import (
	"testing"

	"github.com/google/go-github/v79/github"
)

func TestEventPayloadHelpers(t *testing.T) {
	payload := &EventPayload{
		PullRequest: &github.PullRequest{
			Number: github.Ptr(123),
			Title:  github.Ptr("Test PR"),
		},
		Issue: &github.Issue{
			Number: github.Ptr(456),
			Title:  github.Ptr("Test Issue"),
		},
		Repository: &github.Repository{
			Name:     github.Ptr("test-repo"),
			FullName: github.Ptr("owner/test-repo"),
		},
		Sender: &github.User{
			Login: github.Ptr("testuser"),
		},
	}

	// Test GetPullRequestFromPayload
	pr := payload.GetPullRequestFromPayload()
	if pr == nil || pr.GetNumber() != 123 {
		t.Errorf("GetPullRequestFromPayload() = %v, want PR with number 123", pr.GetNumber())
	}

	// Test GetIssueFromPayload
	issue := payload.GetIssueFromPayload()
	if issue == nil || issue.GetNumber() != 456 {
		t.Errorf("GetIssueFromPayload() = %v, want Issue with number 456", issue.GetNumber())
	}

	// Test GetRepositoryFromPayload
	repo := payload.GetRepositoryFromPayload()
	if repo == nil || repo.GetName() != "test-repo" {
		t.Errorf("GetRepositoryFromPayload() = %v, want Repository with name test-repo", repo.GetName())
	}

	// Test GetSenderFromPayload
	sender := payload.GetSenderFromPayload()
	if sender == nil || sender.GetLogin() != "testuser" {
		t.Errorf("GetSenderFromPayload() = %v, want User with login testuser", sender.GetLogin())
	}
}

func TestGetEventPayloadError(t *testing.T) {
	// Test with missing GITHUB_EVENT_PATH
	t.Setenv("GITHUB_EVENT_PATH", "")

	_, err := GetEventPayload()
	if err == nil {
		t.Error("GetEventPayload() expected error when GITHUB_EVENT_PATH is not set")
	}

	// Test with invalid file path
	t.Setenv("GITHUB_EVENT_PATH", "/nonexistent/path/event.json")

	_, err = GetEventPayload()
	if err == nil {
		t.Error("GetEventPayload() expected error for nonexistent file")
	}
}

func TestLoadEventJson(t *testing.T) {
	if !IsRunsOn() {
		t.Skip("Skipping test; not running in GitHub Actions environment")
	}
	payload, err := GetEventPayload()
	if err != nil {
		t.Errorf("GetEventPayload() returned error: %v", err)
	}
	if payload == nil {
		t.Error("GetEventPayload() returned nil payload")
	}
}
