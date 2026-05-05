package gitutil

import (
	"context"
	"errors"
	"fmt"
	"net/url"
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
	// git fsck writes unreachable/dangling object lines to stderr, not stdout.
	// Use CombinedOutput so that both streams are captured together.
	out, err := cmd.CombinedOutput()
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

// remoteURLMatchesRepo checks whether a single remote URL matches the target repository.
func remoteURLMatchesRepo(repo repository.Repository, url *url.URL) bool {
	if url == nil {
		return false
	}

	parsedRepo, err := repository.Parse(url.String())
	if err != nil {
		return false
	}

	if parsedRepo.Owner != repo.Owner || parsedRepo.Name != repo.Name {
		return false
	}

	return repo.Host == "" || parsedRepo.Host == repo.Host
}

// repoMatchesRemote checks if the repository matches any of the git remotes
func repoMatchesRemote(repo repository.Repository, remotes git.RemoteSet) bool {
	for _, remote := range remotes {
		if remoteURLMatchesRepo(repo, remote.FetchURL) || remoteURLMatchesRepo(repo, remote.PushURL) {
			return true
		}
	}
	return false
}
