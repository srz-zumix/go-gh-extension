package parser

import (
	"fmt"
	"strconv"
)

func GetNumberFromString(s string) (int, error) {
	number, err := strconv.Atoi(s)
	if err == nil {
		return number, nil
	}
	_, err = fmt.Sscanf(s, "#%d", &number)
	if err == nil {
		return number, nil
	}
	return 0, fmt.Errorf("invalid number format: %s", s)
}

func GetIssueNumberFromString(s string) (int, error) {
	if num, err := GetNumberFromString(s); err == nil && num > 0 {
		return num, nil
	}
	if issue, err := ParseIssueURL(s); err == nil && issue != nil && issue.Number != nil {
		return *issue.Number, nil
	}
	return 0, fmt.Errorf("unable to parse issue number from: %s", s)
}

func GetPullRequestNumberFromString(s string) (int, error) {
	if num, err := GetNumberFromString(s); err == nil && num > 0 {
		return num, nil
	}
	if pr, err := ParsePullRequestURL(s); err == nil && pr != nil && pr.Number != nil {
		return *pr.Number, nil
	}
	return 0, fmt.Errorf("unable to parse pull request number from: %s", s)
}

func GetDiscussionNumberFromString(s string) (int, error) {
	if num, err := GetNumberFromString(s); err == nil && num > 0 {
		return num, nil
	}
	if discussion, err := ParseDiscussionURL(s); err == nil && discussion != nil && discussion.Number != nil {
		return *discussion.Number, nil
	}
	return 0, fmt.Errorf("unable to parse discussion number from: %s", s)
}

func GetMilestoneNumberFromString(s string) (int, error) {
	if num, err := GetNumberFromString(s); err == nil && num > 0 {
		return num, nil
	}
	if milestone, err := ParseMilestoneURL(s); err == nil && milestone != nil && milestone.Number != nil {
		return *milestone.Number, nil
	}
	return 0, fmt.Errorf("unable to parse milestone number from: %s", s)
}
