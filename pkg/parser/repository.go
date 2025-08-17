package parser

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/cli/go-gh/v2/pkg/repository"
)

type RepositoryOption func(*repository.Repository) error

func RepositoryInput(input string) RepositoryOption {
	return func(r *repository.Repository) error {
		if input == "" {
			return nil
		}
		repo, err := repository.Parse(input)
		if err != nil {
			repo, err = parseFilePath(input)
			if err != nil {
				return fmt.Errorf(`expected the "[HOST/]OWNER/REPO" format or PATH, got %q`, input)
			}
		}
		if r.Host != "" && r.Host != repo.Host {
			return errors.New("conflicting host")
		}
		if r.Name != "" && r.Name != repo.Name {
			return errors.New("conflicting name")
		}
		if r.Owner != "" && r.Owner != repo.Owner {
			return errors.New("conflicting owner")
		}
		r.Host = repo.Host
		r.Name = repo.Name
		r.Owner = repo.Owner
		return nil
	}
}

func RepositoryOwner(input string) RepositoryOption {
	return func(r *repository.Repository) error {
		if input == "" {
			return nil
		}
		if r.Owner != "" && r.Owner != input {
			return errors.New("conflicting owner")
		}
		if r.Host == "" {
			r.Host, _ = auth.DefaultHost()
		}
		r.Owner = input
		return nil
	}
}

func RepositoryOwners(inputs []string) RepositoryOption {
	return func(r *repository.Repository) error {
		for _, input := range inputs {
			if input == "" {
				continue
			}
			if r.Owner != "" && r.Owner != input {
				return errors.New("conflicting owner")
			}
			r.Owner = input
			break
		}
		return nil
	}
}

// ParseRepository parses a string into a go-gh Repository object. If the string is empty, it returns the current repository.
func Repository(opts ...RepositoryOption) (repository.Repository, error) {
	var r repository.Repository
	for _, opt := range opts {
		if err := opt(&r); err != nil {
			return repository.Repository{}, err
		}
	}
	if r.Host == "" && r.Name == "" && r.Owner == "" {
		return repository.Current()
	}
	return r, nil
}

func GetRepositoryFullName(repo repository.Repository) string {
	return fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
}

func GetRepositoryFullNameWithHost(repo repository.Repository) string {
	if repo.Host != "" {
		return fmt.Sprintf("%s/%s/%s", repo.Host, repo.Owner, repo.Name)
	}
	return fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
}

func parseFilePath(input string) (repo repository.Repository, err error) {
	if input == "" {
		return repo, nil
	}
	// Check if the file path exists
	input += "/.git"
	if _, err := os.Stat(input); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return repo, errors.New("file path does not exist")
		}
		return repo, err
	}
	absPath, err := filepath.Abs(input)
	if err != nil {
		return repo, fmt.Errorf("failed to get absolute path: %w", err)
	}
	gitDir := os.Getenv("GIT_DIR")
	defer func() {
		if gitDir == "" {
			err = os.Unsetenv("GIT_DIR")
		} else {
			err = os.Setenv("GIT_DIR", gitDir)
		}
	}()
	err = os.Setenv("GIT_DIR", absPath)
	if err != nil {
		return repo, err
	}
	return repository.Current()
}
