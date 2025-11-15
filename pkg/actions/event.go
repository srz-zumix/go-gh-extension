package actions

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/go-github/v79/github"
)

// GetEventName returns the name of the GitHub Actions event that triggered the workflow
// It reads from the GITHUB_EVENT_NAME environment variable
func GetEventName() string {
	return os.Getenv("GITHUB_EVENT_NAME")
}

// GetEventJsonPath returns the path to the file containing the event payload
// It reads from the GITHUB_EVENT_PATH environment variable
func GetEventJsonPath() string {
	return os.Getenv("GITHUB_EVENT_PATH")
}

// EventPayload represents the complete GitHub Actions event payload
type EventPayload struct {
	Action            string                     `json:"action,omitempty"`
	Number            int                        `json:"number,omitempty"`
	PullRequest       *github.PullRequest        `json:"pull_request,omitempty"`
	Issue             *github.Issue              `json:"issue,omitempty"`
	Repository        *github.Repository         `json:"repository,omitempty"`
	Organization      *github.Organization       `json:"organization,omitempty"`
	Sender            *github.User               `json:"sender,omitempty"`
	Ref               string                     `json:"ref,omitempty"`
	Before            string                     `json:"before,omitempty"`
	After             string                     `json:"after,omitempty"`
	Commits           []*github.HeadCommit       `json:"commits,omitempty"`
	HeadCommit        *github.HeadCommit         `json:"head_commit,omitempty"`
	Compare           string                     `json:"compare,omitempty"`
	Created           bool                       `json:"created,omitempty"`
	Deleted           bool                       `json:"deleted,omitempty"`
	Forced            bool                       `json:"forced,omitempty"`
	BaseRef           string                     `json:"base_ref,omitempty"`
	Pusher            *github.CommitAuthor       `json:"pusher,omitempty"`
	Inputs            map[string]interface{}     `json:"inputs,omitempty"`
	Workflow          *github.Workflow           `json:"workflow,omitempty"`
	WorkflowRun       *github.WorkflowRun        `json:"workflow_run,omitempty"`
	Release           *github.RepositoryRelease  `json:"release,omitempty"`
	Deployment        *github.Deployment         `json:"deployment,omitempty"`
	Comment           *github.IssueComment       `json:"comment,omitempty"`
	Review            *github.PullRequestReview  `json:"review,omitempty"`
	ReviewComment     *github.PullRequestComment `json:"review_comment,omitempty"`
	Label             *github.Label              `json:"label,omitempty"`
	Milestone         *github.Milestone          `json:"milestone,omitempty"`
	Team              *github.Team               `json:"team,omitempty"`
	Member            *github.User               `json:"member,omitempty"`
	Installation      *github.Installation       `json:"installation,omitempty"`
	CheckRun          *github.CheckRun           `json:"check_run,omitempty"`
	CheckSuite        *github.CheckSuite         `json:"check_suite,omitempty"`
	Package           *github.Package            `json:"package,omitempty"`
	Discussion        interface{}                `json:"discussion,omitempty"`
	DiscussionComment interface{}                `json:"discussion_comment,omitempty"`
}

// GetEventPayload reads and parses the complete GitHub Actions event payload
func GetEventPayload() (*EventPayload, error) {
	path := GetEventJsonPath()
	if path == "" {
		return nil, fmt.Errorf("GITHUB_EVENT_PATH is not set")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open event file: %w", err)
	}
	defer f.Close() // nolint

	var payload EventPayload
	if err := json.NewDecoder(f).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode event payload: %w", err)
	}

	return &payload, nil
}

// Event type constants for common GitHub Actions events
const (
	EventPush                     = "push"
	EventPullRequest              = "pull_request"
	EventPullRequestTarget        = "pull_request_target"
	EventIssues                   = "issues"
	EventIssueComment             = "issue_comment"
	EventWorkflowDispatch         = "workflow_dispatch"
	EventSchedule                 = "schedule"
	EventRelease                  = "release"
	EventCreate                   = "create"
	EventDelete                   = "delete"
	EventFork                     = "fork"
	EventWatch                    = "watch"
	EventStar                     = "star"
	EventMember                   = "member"
	EventTeam                     = "team"
	EventTeamAdd                  = "team_add"
	EventOrganization             = "organization"
	EventRepository               = "repository"
	EventRepositoryDispatch       = "repository_dispatch"
	EventCheckRun                 = "check_run"
	EventCheckSuite               = "check_suite"
	EventDeployment               = "deployment"
	EventDeploymentStatus         = "deployment_status"
	EventLabel                    = "label"
	EventMilestone                = "milestone"
	EventPageBuild                = "page_build"
	EventProject                  = "project"
	EventProjectCard              = "project_card"
	EventProjectColumn            = "project_column"
	EventPublic                   = "public"
	EventPullRequestReview        = "pull_request_review"
	EventPullRequestReviewComment = "pull_request_review_comment"
	EventWorkflowRun              = "workflow_run"
	EventPackage                  = "package"
	EventDiscussion               = "discussion"
	EventDiscussionComment        = "discussion_comment"
)

// Action constants for common GitHub Actions event actions
const (
	ActionOpened          = "opened"
	ActionClosed          = "closed"
	ActionReopened        = "reopened"
	ActionEdited          = "edited"
	ActionSynchronize     = "synchronize"
	ActionCreated         = "created"
	ActionDeleted         = "deleted"
	ActionUpdated         = "updated"
	ActionAssigned        = "assigned"
	ActionUnassigned      = "unassigned"
	ActionLabeled         = "labeled"
	ActionUnlabeled       = "unlabeled"
	ActionLocked          = "locked"
	ActionUnlocked        = "unlocked"
	ActionTransferred     = "transferred"
	ActionPinned          = "pinned"
	ActionUnpinned        = "unpinned"
	ActionMilestoned      = "milestoned"
	ActionDemilestoned    = "demilestoned"
	ActionSubmitted       = "submitted"
	ActionDismissed       = "dismissed"
	ActionCompleted       = "completed"
	ActionRequested       = "requested"
	ActionRequestedAction = "requested_action"
	ActionRerequested     = "rerequested"
	ActionPublished       = "published"
	ActionUnpublished     = "unpublished"
	ActionPreReleased     = "prereleased"
	ActionReleased        = "released"
)

// Helper methods for common event types

// IsPushEvent returns true if the current event is a push event
func IsPushEvent() bool {
	return GetEventName() == EventPush
}

// IsPullRequestEvent returns true if the current event is a pull request event
func IsPullRequestEvent() bool {
	eventName := GetEventName()
	return eventName == EventPullRequest || eventName == EventPullRequestTarget
}

// IsIssueEvent returns true if the current event is an issue event
func IsIssueEvent() bool {
	return GetEventName() == EventIssues
}

// IsWorkflowDispatchEvent returns true if the current event is a workflow_dispatch event
func IsWorkflowDispatchEvent() bool {
	return GetEventName() == EventWorkflowDispatch
}

// IsScheduleEvent returns true if the current event is a schedule event
func IsScheduleEvent() bool {
	return GetEventName() == EventSchedule
}

// GetPullRequestFromPayload extracts the pull request from the event payload
func (p *EventPayload) GetPullRequestFromPayload() *github.PullRequest {
	return p.PullRequest
}

// GetIssueFromPayload extracts the issue from the event payload
func (p *EventPayload) GetIssueFromPayload() *github.Issue {
	return p.Issue
}

// GetRepositoryFromPayload extracts the repository from the event payload
func (p *EventPayload) GetRepositoryFromPayload() *github.Repository {
	return p.Repository
}

// GetSenderFromPayload extracts the sender (user who triggered the event) from the event payload
func (p *EventPayload) GetSenderFromPayload() *github.User {
	return p.Sender
}
