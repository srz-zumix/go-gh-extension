package gitutil

import (
	"net/url"
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/stretchr/testify/assert"
)

func TestRemoteURLMatchesRepoIgnoresCase(t *testing.T) {
	u, err := url.Parse("https://GitHub.com/Example-Org/Example-Repo.git")
	assert.NoError(t, err)
	assert.True(t, remoteURLMatchesRepo(repository.Repository{Host: "github.com", Owner: "example-org", Name: "example-repo"}, u))
	assert.False(t, remoteURLMatchesRepo(repository.Repository{Host: "github.com", Owner: "example-org", Name: "other-repo"}, u))
}
