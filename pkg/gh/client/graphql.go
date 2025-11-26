package client

import (
	"strings"

	"github.com/shurcooL/githubv4"
)

// GraphQLOrderByOption represents ordering options for GraphQL queries.
type GraphQLOrderByOption struct {
	Field     *string
	Direction *string
}

func (opt *GraphQLOrderByOption) Asc() {
	asc := "ASC"
	opt.Direction = &asc
}

func (opt *GraphQLOrderByOption) Desc() {
	desc := "DESC"
	opt.Direction = &desc
}

func (opt *GraphQLOrderByOption) CreatedAt() {
	createdAt := "CREATED_AT"
	opt.Field = &createdAt
}

func (opt *GraphQLOrderByOption) UpdatedAt() {
	updatedAt := "UPDATED_AT"
	opt.Field = &updatedAt
}

func (opt *GraphQLOrderByOption) Comments() {
	comments := "COMMENTS"
	opt.Field = &comments
}

// ToPullRequestOrder converts GraphQLOrderByOption to githubv4.PullRequestOrder.
func (opt *GraphQLOrderByOption) ToPullRequestOrder() *githubv4.PullRequestOrder {
	order := &githubv4.PullRequestOrder{
		Field:     githubv4.PullRequestOrderFieldCreatedAt,
		Direction: githubv4.OrderDirectionDesc,
	}
	if opt == nil {
		return order
	}
	if opt.Field != nil {
		switch strings.ToLower(*opt.Field) {
		case "created_at":
			order.Field = githubv4.PullRequestOrderFieldCreatedAt
		case "updated_at":
			order.Field = githubv4.PullRequestOrderFieldUpdatedAt
		}
	}
	if opt.Direction != nil {
		switch strings.ToLower(*opt.Direction) {
		case "asc":
			order.Direction = githubv4.OrderDirectionAsc
		case "desc":
			order.Direction = githubv4.OrderDirectionDesc
		}
	}
	return order
}

// ToIssueOrder converts GraphQLOrderByOption to githubv4.IssueOrder.
func (opt *GraphQLOrderByOption) ToIssueOrder() *githubv4.IssueOrder {
	order := &githubv4.IssueOrder{
		Field:     githubv4.IssueOrderFieldCreatedAt,
		Direction: githubv4.OrderDirectionDesc,
	}
	if opt == nil {
		return order
	}
	if opt.Field != nil {
		switch strings.ToLower(*opt.Field) {
		case "created_at":
			order.Field = githubv4.IssueOrderFieldCreatedAt
		case "updated_at":
			order.Field = githubv4.IssueOrderFieldUpdatedAt
		case "comments":
			order.Field = githubv4.IssueOrderFieldComments
		}
	}
	if opt.Direction != nil {
		switch strings.ToLower(*opt.Direction) {
		case "asc":
			order.Direction = githubv4.OrderDirectionAsc
		case "desc":
			order.Direction = githubv4.OrderDirectionDesc
		}
	}
	return order
}

// parsePullRequestStates converts a slice of string states to a slice of githubv4.PullRequestState.
func ParsePullRequestStates(states []string) []githubv4.PullRequestState {
	result := []githubv4.PullRequestState{}
	for _, s := range states {
		result = append(result, ParsePullRequestState(s))
	}
	return result
}

// ParsePullRequestState converts a string state to githubv4.PullRequestState.
func ParsePullRequestState(state string) githubv4.PullRequestState {
	switch strings.ToLower(state) {
	case "open":
		return githubv4.PullRequestStateOpen
	case "closed":
		return githubv4.PullRequestStateClosed
	case "merged":
		return githubv4.PullRequestStateMerged
	}
	return githubv4.PullRequestStateOpen
}
