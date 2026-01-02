package parser

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/actions"
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

func RepositoryInputOptional(input string) RepositoryOption {
	return func(r *repository.Repository) error {
		_ = RepositoryInput(input)(r)
		return nil
	}
}

func RepositoryFromURL(input string) RepositoryOption {
	return func(r *repository.Repository) error {
		if input == "" {
			return nil
		}

		issueURL, err := ParseIssueURL(input)
		if err != nil {
			return fmt.Errorf(`failed to parse repository from URL %q: %w`, input, err)
		}
		if issueURL != nil && issueURL.Repo != nil {
			if r.Host != "" && r.Host != issueURL.Repo.Host {
				return errors.New("conflicting host")
			}
			if r.Name != "" && r.Name != issueURL.Repo.Name {
				return errors.New("conflicting name")
			}
			if r.Owner != "" && r.Owner != issueURL.Repo.Owner {
				return errors.New("conflicting owner")
			}
			r.Host = issueURL.Repo.Host
			r.Name = issueURL.Repo.Name
			r.Owner = issueURL.Repo.Owner
		}
		return nil
	}
}

func RepositoryOwner(input string) RepositoryOption {
	return func(r *repository.Repository) error {
		if input == "" {
			return nil
		}
		if strings.Contains(input, "/") {
			return fmt.Errorf(`expected owner name without '/', got %q`, input)
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

func RepositoryOwnerWithHost(input string) RepositoryOption {
	return func(r *repository.Repository) error {
		if input == "" {
			return nil
		}
		parsed, err := repository.Parse(input + "/dummy")
		if err != nil {
			return fmt.Errorf(`expected the "[HOST/]OWNER" format, got %q`, input)
		}
		if r.Owner != "" && r.Owner != parsed.Owner {
			return errors.New("conflicting owner")
		}
		if r.Host != "" && r.Host != parsed.Host {
			return errors.New("conflicting host")
		}
		r.Host = parsed.Host
		r.Owner = parsed.Owner
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
func Repository(opts ...RepositoryOption) (r repository.Repository, err error) {
	for _, opt := range opts {
		if err := opt(&r); err != nil {
			return repository.Repository{}, err
		}
	}
	if r.Host == "" && r.Name == "" && r.Owner == "" {
		r, err = repository.Current()
		if err != nil {
			opt := RepositoryInput(actions.GetRepositoryFullNameWithHost())
			if err := opt(&r); err != nil {
				return repository.Repository{}, err
			}
			if r.Host == "" && r.Name == "" && r.Owner == "" {
				return repository.Repository{}, errors.New("failed to parse repository")
			}
		}
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
