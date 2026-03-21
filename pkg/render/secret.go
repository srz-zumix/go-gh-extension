package render

import (
	"strings"

	"github.com/google/go-github/v79/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

// SecretFieldGetter defines a function to get a field value from a github.Secret
type SecretFieldGetter func(secret *github.Secret) string

// SecretFieldGetters holds field getters for Secret table rendering.
type SecretFieldGetters struct {
	Func map[string]SecretFieldGetter
}

// NewSecretFieldGetters creates field getters for Secret table rendering
func NewSecretFieldGetters() *SecretFieldGetters {
	return &SecretFieldGetters{
		Func: map[string]SecretFieldGetter{
			"NAME": func(secret *github.Secret) string {
				return secret.Name
			},
			"CREATED_AT": func(secret *github.Secret) string {
				return ToString(secret.CreatedAt)
			},
			"UPDATED_AT": func(secret *github.Secret) string {
				return ToString(secret.UpdatedAt)
			},
			"VISIBILITY": func(secret *github.Secret) string {
				return secret.Visibility
			},
		},
	}
}

// GetField returns the value of the specified field for the given secret.
func (g *SecretFieldGetters) GetField(secret *github.Secret, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(secret)
	}
	return ""
}

// SecretsFieldGetter defines a function to get a field value from a []*github.Secret slice
type SecretsFieldGetter func(secrets []*github.Secret) string

// SecretsFieldGetters holds field getters for []*github.Secret slice rendering.
type SecretsFieldGetters struct {
	Func map[string]SecretsFieldGetter
}

// NewSecretsFieldGetters creates field getters for []*github.Secret slice rendering
func NewSecretsFieldGetters() *SecretsFieldGetters {
	return &SecretsFieldGetters{
		Func: map[string]SecretsFieldGetter{
			"NAMES": func(secrets []*github.Secret) string {
				names := gh.GetObjectNames(secrets)
				return strings.Join(names, ", ")
			},
			"COUNT": func(secrets []*github.Secret) string {
				return ToString(len(secrets))
			},
			"SECRETS": func(secrets []*github.Secret) string {
				names := gh.GetObjectNames(secrets)
				return strings.Join(names, ", ")
			},
		},
	}
}

// GetField returns the value of the specified field for the given secrets slice.
func (g *SecretsFieldGetters) GetField(secrets []*github.Secret, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(secrets)
	}
	return ""
}

// RenderSecrets renders a list of secrets as a table using the given headers.
func (r *Renderer) RenderSecrets(secrets []*github.Secret, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(secrets)
	}

	if len(secrets) == 0 {
		r.writeLine("No secrets found.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"NAME", "CREATED_AT", "UPDATED_AT"}
	}

	getter := NewSecretFieldGetters()
	table := r.newTableWriter(headers)

	for _, secret := range secrets {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(secret, header)
		}
		table.Append(row)
	}
	return table.Render()
}
