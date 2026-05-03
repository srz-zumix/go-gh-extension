package gitutil

import (
	"sync"

	"github.com/cli/cli/v2/git"
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
