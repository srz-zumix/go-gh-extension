package gh

import (
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
)

// SyncChangeAction describes the operation needed to synchronize an item.
type SyncChangeAction string

const (
	SyncChangeAdd    SyncChangeAction = "add"
	SyncChangeUpdate SyncChangeAction = "update"
	SyncChangeRemove SyncChangeAction = "remove"
)

// SyncChange describes a planned change to a destination.
type SyncChange struct {
	Destination string           `json:"destination"`
	Action      SyncChangeAction `json:"action"`
	Target      string           `json:"target"`
	Permission  string           `json:"permission,omitempty"`
	Current     string           `json:"current,omitempty"`
	TeamOwner   string           `json:"-"`
}

func syncChangeDestination(repo repository.Repository) string {
	destination := repo.Owner + "/" + repo.Name
	if repo.Host != "" {
		destination = repo.Host + "/" + destination
	}
	return destination
}

func syncChangeTeamDestination(repo repository.Repository, teamSlug string) string {
	destination := repo.Owner + "/" + teamSlug
	if repo.Host != "" {
		destination = repo.Host + "/" + destination
	}
	return destination
}

func syncChangeTeamTarget(owner, slug string) string {
	return strings.Trim(owner+"/"+slug, "/")
}
