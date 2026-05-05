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

// NewClientWithDir creates a new git client operating in the specified directory.
func NewClientWithDir(dir string) *git.Client {
	return &git.Client{RepoDir: dir}
}
