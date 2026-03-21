package render

import (
	"maps"
	"slices"
	"strings"

	"github.com/google/go-github/v84/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

// RepositoryFieldGetter defines a function to get a field value from a github.Repository
type RepositoryFieldGetter func(repo *github.Repository) string

// RepositoryFieldGetters holds field getters for Repository table rendering.
type RepositoryFieldGetters struct {
	Func map[string]RepositoryFieldGetter
}

// NewRepositoryFieldGetters creates field getters for Repository table rendering
func NewRepositoryFieldGetters() *RepositoryFieldGetters {
	return &RepositoryFieldGetters{
		Func: map[string]RepositoryFieldGetter{
			"REPOSITORY": func(repo *github.Repository) string {
				return ToString(repo.FullName)
			},
			"NAME": func(repo *github.Repository) string {
				return ToString(repo.FullName)
			},
			"OWNER": func(repo *github.Repository) string {
				if repo.Owner == nil {
					return ""
				}
				return ToString(repo.Owner.Login)
			},
			"REPO": func(repo *github.Repository) string {
				return ToString(repo.Name)
			},
			"PERMISSION": func(repo *github.Repository) string {
				return gh.GetRepositoryPermissions(repo)
			},
			"VISIBILITY": func(repo *github.Repository) string {
				return ToString(repo.Visibility)
			},
			"DESCRIPTION": func(repo *github.Repository) string {
				return ToString(repo.Description)
			},
			"ARCHIVED": func(repo *github.Repository) string {
				return ToString(repo.Archived)
			},
			"FORK": func(repo *github.Repository) string {
				return ToString(repo.Fork)
			},
			"LANGUAGE": func(repo *github.Repository) string {
				return ToString(repo.Language)
			},
			"URL": func(repo *github.Repository) string {
				return ToString(repo.HTMLURL)
			},
		},
	}
}

// GetField returns the value of the specified field for the given repository.
func (g *RepositoryFieldGetters) GetField(repo *github.Repository, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(repo)
	}
	return ""
}

// RepoWithSecretFieldGetters holds combined field getters for RepoWithSecrets table rendering.
// Repository fields take precedence over Secrets fields when the same key exists in both.
type RepoWithSecretFieldGetters struct {
	Repo    *RepositoryFieldGetters
	Secrets *SecretsFieldGetters
}

// NewRepoWithSecretFieldGetters creates a combined field getter for RepoWithSecrets rendering
func NewRepoWithSecretFieldGetters() *RepoWithSecretFieldGetters {
	return &RepoWithSecretFieldGetters{
		Repo:    NewRepositoryFieldGetters(),
		Secrets: NewSecretsFieldGetters(),
	}
}

// GetField resolves the field from Repository first, then Secrets slice.
func (g *RepoWithSecretFieldGetters) GetField(repo *github.Repository, secrets []*github.Secret, field string) string {
	field = strings.ToUpper(field)
	if _, ok := g.Repo.Func[field]; ok {
		return g.Repo.GetField(repo, field)
	}
	return g.Secrets.GetField(secrets, field)
}

// RenderRepository renders a list of repositories as a table using the given headers.
func (r *Renderer) RenderRepository(repos []*github.Repository, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(repos)
	}

	if len(repos) == 0 {
		r.writeLine("No repositories.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"NAME", "PERMISSION", "VISIBILITY"}
	}

	getter := NewRepositoryFieldGetters()
	table := r.newTableWriter(headers)

	for _, repo := range repos {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(repo, header)
		}
		table.Append(row)
	}
	return table.Render()
}

// RenderRepositoriesWithSecret renders a table of repositories with their secrets, one row per repository.
func (r *Renderer) RenderRepositoriesWithSecret(repos []gh.RepoWithSecrets, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(repos)
	}

	if len(repos) == 0 {
		r.writeLine("No repositories with secrets found.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"REPOSITORY", "NAMES"}
	}

	getter := NewRepoWithSecretFieldGetters()
	table := r.newTableWriter(headers)

	for i := range repos {
		row := make([]string, len(headers))
		for j, header := range headers {
			row[j] = getter.GetField(repos[i].Repository, repos[i].Secrets, header)
		}
		table.Append(row)
	}
	return table.Render()
}

// RenderRepositoriesWithSecretCount renders a table of repositories and their secret counts.
func (r *Renderer) RenderRepositoriesWithSecretCount(repos []gh.RepoWithSecrets) error {
	return r.RenderRepositoriesWithSecret(repos, []string{"REPOSITORY", "COUNT"})
}

// RenderRepositoriesWithScopedSecretCount renders a table of repositories with per-scope secret counts.
// Each repository may generate multiple rows: one for repository-level secrets and one for each environment.
func (r *Renderer) RenderRepositoriesWithScopedSecretCount(repos []gh.RepoWithSecrets) error {
	if r.exporter != nil {
		return r.RenderExportedData(repos)
	}

	if len(repos) == 0 {
		r.writeLine("No repositories with secrets found.")
		return nil
	}

	headers := []string{"REPOSITORY", "SCOPE", "COUNT"}
	table := r.newTableWriter(headers)

	for i := range repos {
		repoName := ToString(repos[i].Repository.FullName)

		if repos[i].SecretCount() > 0 {
			table.Append([]string{repoName, "repository", ToString(repos[i].SecretCount())})
		}

		// Sort environment names for consistent output
		envNames := slices.Collect(maps.Keys(repos[i].EnvSecrets))

		for _, envName := range envNames {
			secrets := repos[i].EnvSecrets[envName]
			if len(secrets) > 0 {
				table.Append([]string{repoName, "env:" + envName, ToString(len(secrets))})
			}
		}
	}
	return table.Render()
}
