package gh

import (
	"context"
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/stretchr/testify/assert"
)

func TestListAvailableRunnersRejectsOrgMode(t *testing.T) {
	// An empty repo.Name means organization mode, which is not supported here.
	// The guard must return before any API call, so a nil client is safe.
	repo := repository.Repository{Owner: "octo-org", Name: ""}

	runners, err := ListAvailableRunners(context.Background(), nil, repo)

	assert.Error(t, err)
	assert.Nil(t, runners)
}
