package gitutil

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cli/cli/v2/git"
	"github.com/cli/go-gh/v2/pkg/repository"
)

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

// ListUnreachableCommits returns the SHAs of commit objects that are not
// reachable from any ref in the local git repository.
// When noReflogs is true, reflog entries are excluded from reachability analysis,
// which may surface additional candidates.
func ListUnreachableCommits(ctx context.Context, noReflogs bool) ([]string, error) {
	c := NewClient()
	args := []string{"fsck", "--unreachable"}
	if noReflogs {
		args = append(args, "--no-reflogs")
	}
	cmd, err := c.Command(ctx, args...)
	if err != nil {
		return nil, err
	}
	out, err := cmd.Output()
	if err != nil {
		// git fsck exits non-zero when it finds connectivity or fsck errors.
		// We still want to parse whatever output was produced.
		var ge *git.GitError
		if !errors.As(err, &ge) {
			return nil, err
		}
	}
	var shas []string
	for _, line := range strings.Split(string(out), "\n") {
		// Output lines: "unreachable commit <sha>" or "dangling commit <sha>"
		parts := strings.Fields(line)
		if len(parts) == 3 && (parts[0] == "unreachable" || parts[0] == "dangling") && parts[1] == "commit" {
			shas = append(shas, parts[2])
		}
	}
	return shas, nil
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
