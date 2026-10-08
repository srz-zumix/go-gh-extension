package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v90/github"
)

// ListOrgHostedRunners lists the configured GitHub-hosted pools of repo's owner.
func ListOrgHostedRunners(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.HostedRunner, error) {
	return g.ListOrgHostedRunners(ctx, repo.Owner)
}

func ListEnterpriseHostedRunners(ctx context.Context, g *GitHubClient, enterprise repository.Repository) ([]*github.HostedRunner, error) {
	return g.ListEnterpriseHostedRunners(ctx, enterprise.Owner)
}

// ListOrgHostedRunnersWithInherited lists the organization pools plus those of runner groups inherited from the enterprise,
// which the organization list omits and which only need organization admin permission to read.
func ListOrgHostedRunnersWithInherited(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.HostedRunner, error) {
	runners, err := ListOrgHostedRunners(ctx, g, repo)
	if err != nil {
		return nil, err
	}
	groups, err := ListOrgRunnerGroups(ctx, g, repo)
	if err != nil {
		return nil, err
	}
	seen := make(map[[2]int64]bool, len(runners))
	for _, runner := range runners {
		seen[[2]int64{runner.GetRunnerGroupID(), runner.GetID()}] = true
	}
	for _, group := range groups {
		if !group.GetInherited() {
			continue
		}
		groupRunners, err := g.ListOrgRunnerGroupHostedRunners(ctx, repo.Owner, group.GetID())
		if err != nil {
			return nil, err
		}
		for _, runner := range groupRunners {
			key := [2]int64{group.GetID(), runner.GetID()}
			if seen[key] {
				continue
			}
			seen[key] = true
			runners = append(runners, runner)
		}
	}
	return runners, nil
}
