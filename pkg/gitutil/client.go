package gitutil

import (
	"context"
	"fmt"
	"sync"

	"github.com/cli/cli/v2/git"
	"github.com/cli/go-gh/v2/pkg/repository"
)

var client *git.Client
var clientOnce sync.Once

// NewClient creates a new git client
func NewClient() *git.Client {
	clientOnce.Do(func() {
		c := git.Client{}
		client = &c
	})
	return client
}

// GetCurrentBranchIfRepoMatches returns the current branch if the repository matches the git remotes
func GetCurrentBranchIfRepoMatches(ctx context.Context, repo repository.Repository) (string, error) {
	gitClient := NewClient()

	currentBranch, err := gitClient.CurrentBranch(ctx)
	if err != nil {
		return "", err
	}

	if currentBranch == "" {
		return "", nil
	}

	// Verify the repository matches
	remotes, err := gitClient.Remotes(ctx)
	if err != nil {
		return "", err
	}

	if !repoMatchesRemote(repo, remotes) {
		return "", fmt.Errorf("repository %s/%s does not match any git remotes", repo.Owner, repo.Name)
	}

	return currentBranch, nil
}

// repoMatchesRemote checks if the repository matches any of the git remotes
func repoMatchesRemote(repo repository.Repository, remotes git.RemoteSet) bool {
	for _, remote := range remotes {
		url := remote.FetchURL
		if url == nil {
			if remote.PushURL == nil {
				continue
			}
			url = remote.PushURL
		}
		parsedRepo, err := repository.Parse(url.String())
		if err != nil {
			continue
		}
		if parsedRepo.Owner == repo.Owner && parsedRepo.Name == repo.Name {
			if repo.Host == "" || parsedRepo.Host == repo.Host {
				return true
			}
		}
	}
	return false
}
