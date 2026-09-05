package gh

import (
	"context"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v90/github"
)

// ListRunners lists all self-hosted runners for a repository (wrapper)
func ListRunners(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Runner, error) {
	if repo.Name == "" {
		return ListOrgRunners(ctx, g, repo)
	}
	return g.ListRunners(ctx, repo.Owner, repo.Name)
}

// FindRunner finds a self-hosted runner by name for a repository (wrapper)
func FindRunner(ctx context.Context, g *GitHubClient, repo repository.Repository, runnerName string) (*github.Runner, error) {
	if repo.Name == "" {
		return FindOrgRunner(ctx, g, repo, runnerName)
	}
	return g.FindRunner(ctx, repo.Owner, repo.Name, runnerName)
}

// FindRunnersByLabel finds every self-hosted runner that has the given label for a repository or organization (wrapper)
func FindRunnersByLabel(ctx context.Context, g *GitHubClient, repo repository.Repository, label string) ([]*github.Runner, error) {
	runners, err := ListRunners(ctx, g, repo)
	if err != nil {
		return nil, err
	}

	matched := make([]*github.Runner, 0, len(runners))
	for _, runner := range runners {
		if HasRunnerLabel(runner, label) {
			matched = append(matched, runner)
		}
	}
	return matched, nil
}

// HasRunnerLabel reports whether runner has a label matching name (case-insensitive)
func HasRunnerLabel(runner *github.Runner, name string) bool {
	for _, label := range runner.Labels {
		if strings.EqualFold(label.GetName(), name) {
			return true
		}
	}
	return false
}

// FilterRunnersByStatus returns the runners whose status matches status.
// Runners are returned unchanged when status is empty.
func FilterRunnersByStatus(runners []*github.Runner, status string) []*github.Runner {
	if status == "" {
		return runners
	}

	matched := make([]*github.Runner, 0, len(runners))
	for _, runner := range runners {
		if strings.EqualFold(runner.GetStatus(), status) {
			matched = append(matched, runner)
		}
	}
	return matched
}

// GetRunner gets a single self-hosted runner for a repository or organization (wrapper)
func GetRunner(ctx context.Context, g *GitHubClient, repo repository.Repository, runnerID int64) (*github.Runner, error) {
	if repo.Name == "" {
		return GetOrgRunner(ctx, g, repo, runnerID)
	}
	return g.GetRunner(ctx, repo.Owner, repo.Name, runnerID)
}

// ListOrgRunners lists all self-hosted runners for an organization (wrapper)
func ListOrgRunners(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Runner, error) {
	return g.ListOrgRunners(ctx, repo.Owner)
}

// FindOrgRunner finds a self-hosted runner by name for an organization (wrapper)
func FindOrgRunner(ctx context.Context, g *GitHubClient, repo repository.Repository, runnerName string) (*github.Runner, error) {
	return g.FindOrgRunner(ctx, repo.Owner, runnerName)
}

// GetOrgRunner gets a single self-hosted runner for an organization (wrapper)
func GetOrgRunner(ctx context.Context, g *GitHubClient, repo repository.Repository, runnerID int64) (*github.Runner, error) {
	return g.GetOrgRunner(ctx, repo.Owner, runnerID)
}

// CreateRegistrationToken creates a registration token for a repository or organization (wrapper)
func CreateRegistrationToken(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.RegistrationToken, error) {
	if repo.Name == "" {
		return g.CreateOrgRegistrationToken(ctx, repo.Owner)
	}
	return g.CreateRegistrationToken(ctx, repo.Owner, repo.Name)
}

// RemoveRunner removes a self-hosted runner from a repository or organization (wrapper)
func RemoveRunner(ctx context.Context, g *GitHubClient, repo repository.Repository, runnerID int64) error {
	if repo.Name == "" {
		return g.RemoveOrgRunner(ctx, repo.Owner, runnerID)
	}
	return g.RemoveRunner(ctx, repo.Owner, repo.Name, runnerID)
}

// ListOrgRunnerGroups lists all organization runner groups (wrapper)
func ListOrgRunnerGroups(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.RunnerGroup, error) {
	return g.ListOrgRunnerGroups(ctx, repo.Owner)
}

// FindOrgRunnerGroup finds an organization runner group by name (wrapper)
func FindOrgRunnerGroup(ctx context.Context, g *GitHubClient, repo repository.Repository, groupName string) (*github.RunnerGroup, error) {
	groups, err := ListOrgRunnerGroups(ctx, g, repo)
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		if group.GetName() == groupName {
			return group, nil
		}
	}
	return nil, nil // Group not found
}

// CreateOrgRunnerGroup creates a new organization runner group (wrapper)
func CreateOrgRunnerGroup(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) (*github.RunnerGroup, error) {
	return g.CreateOrgRunnerGroup(ctx, repo.Owner, name)
}

// ListOrgRunnerGroupRunners lists all self-hosted runners belonging to an organization runner group (wrapper)
func ListOrgRunnerGroupRunners(ctx context.Context, g *GitHubClient, repo repository.Repository, groupID int64) ([]*github.Runner, error) {
	return g.ListOrgRunnerGroupRunners(ctx, repo.Owner, groupID)
}

// ListOrgRunnerGroupRepositories lists all repositories that have access to an organization runner group (wrapper)
func ListOrgRunnerGroupRepositories(ctx context.Context, g *GitHubClient, repo repository.Repository, groupID int64) ([]*github.Repository, error) {
	return g.ListOrgRunnerGroupRepositories(ctx, repo.Owner, groupID)
}

// ListAvailableRunners lists the self-hosted runners a repository can schedule jobs on:
// the runners registered to the repository itself plus the organization runners belonging
// to every runner group that is visible to the repository (wrapper)
func ListAvailableRunners(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Runner, error) {
	runners, err := ListRunners(ctx, g, repo)
	if err != nil {
		return nil, err
	}

	seen := make(map[int64]bool, len(runners))
	for _, runner := range runners {
		seen[runner.GetID()] = true
	}

	target, err := GetRepository(ctx, g, repo)
	if err != nil {
		return nil, err
	}

	groups, err := ListOrgRunnerGroups(ctx, g, repo)
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		visible, err := IsOrgRunnerGroupVisibleTo(ctx, g, repo, group, target)
		if err != nil {
			return nil, err
		}
		if !visible {
			continue
		}

		groupRunners, err := ListOrgRunnerGroupRunners(ctx, g, repo, group.GetID())
		if err != nil {
			return nil, err
		}
		for _, runner := range groupRunners {
			if seen[runner.GetID()] {
				continue
			}
			seen[runner.GetID()] = true
			runners = append(runners, runner)
		}
	}

	return runners, nil
}

// IsOrgRunnerGroupVisibleTo reports whether target can use the runners of an organization runner group (wrapper)
func IsOrgRunnerGroupVisibleTo(ctx context.Context, g *GitHubClient, repo repository.Repository, group *github.RunnerGroup, target *github.Repository) (bool, error) {
	switch group.GetVisibility() {
	case "all":
		return true, nil
	case "private":
		return target.GetPrivate(), nil
	case "selected":
		repos, err := ListOrgRunnerGroupRepositories(ctx, g, repo, group.GetID())
		if err != nil {
			return false, err
		}
		for _, r := range repos {
			if r.GetID() == target.GetID() {
				return true, nil
			}
		}
	}
	return false, nil
}

// DeleteOrgRunnerGroup deletes an organization runner group by ID (wrapper)
func DeleteOrgRunnerGroup(ctx context.Context, g *GitHubClient, repo repository.Repository, groupID int64) error {
	return g.DeleteOrgRunnerGroup(ctx, repo.Owner, groupID)
}

// CreateOrgRunnerGroupWithRequest creates a new organization runner group using the given request (wrapper)
func CreateOrgRunnerGroupWithRequest(ctx context.Context, g *GitHubClient, repo repository.Repository, request github.CreateRunnerGroupRequest) (*github.RunnerGroup, error) {
	return g.CreateOrgRunnerGroupWithRequest(ctx, repo.Owner, request)
}

// AddOrgRunnerGroupRunner adds a self-hosted runner to an organization runner group (wrapper)
func AddOrgRunnerGroupRunner(ctx context.Context, g *GitHubClient, repo repository.Repository, groupID, runnerID int64) error {
	return g.AddOrgRunnerGroupRunner(ctx, repo.Owner, groupID, runnerID)
}

// RemoveOrgRunnerGroupRunner removes a self-hosted runner from an organization runner group,
// returning it to the default group (wrapper)
func RemoveOrgRunnerGroupRunner(ctx context.Context, g *GitHubClient, repo repository.Repository, groupID, runnerID int64) error {
	return g.RemoveOrgRunnerGroupRunner(ctx, repo.Owner, groupID, runnerID)
}

// ListRunnerLabels lists all labels for a self-hosted runner for a repository or organization (wrapper)
func ListRunnerLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, runnerID int64) ([]*github.RunnerLabels, error) {
	if repo.Name == "" {
		return g.ListOrgRunnerLabels(ctx, repo.Owner, runnerID)
	}
	return g.ListRunnerLabels(ctx, repo.Owner, repo.Name, runnerID)
}

// AddRunnerLabels adds custom labels to a self-hosted runner for a repository or organization (wrapper)
func AddRunnerLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, runnerID int64, labels []string) ([]*github.RunnerLabels, error) {
	if repo.Name == "" {
		return g.AddOrgRunnerLabels(ctx, repo.Owner, runnerID, labels)
	}
	return g.AddRunnerLabels(ctx, repo.Owner, repo.Name, runnerID, labels)
}

// SetRunnerLabels replaces all custom labels for a self-hosted runner for a repository or organization (wrapper)
func SetRunnerLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, runnerID int64, labels []string) ([]*github.RunnerLabels, error) {
	if repo.Name == "" {
		return g.SetOrgRunnerLabels(ctx, repo.Owner, runnerID, labels)
	}
	return g.SetRunnerLabels(ctx, repo.Owner, repo.Name, runnerID, labels)
}

// RemoveRunnerLabel removes a single custom label from a self-hosted runner for a repository or organization (wrapper)
func RemoveRunnerLabel(ctx context.Context, g *GitHubClient, repo repository.Repository, runnerID int64, name string) ([]*github.RunnerLabels, error) {
	if repo.Name == "" {
		return g.RemoveOrgRunnerLabel(ctx, repo.Owner, runnerID, name)
	}
	return g.RemoveRunnerLabel(ctx, repo.Owner, repo.Name, runnerID, name)
}

// RemoveAllRunnerLabels removes all custom labels from a self-hosted runner for a repository or organization (wrapper)
func RemoveAllRunnerLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, runnerID int64) ([]*github.RunnerLabels, error) {
	if repo.Name == "" {
		return g.RemoveAllOrgRunnerLabels(ctx, repo.Owner, runnerID)
	}
	return g.RemoveAllRunnerLabels(ctx, repo.Owner, repo.Name, runnerID)
}
