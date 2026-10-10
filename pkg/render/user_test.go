package render

import (
	"testing"

	"github.com/google/go-github/v90/github"
	"github.com/stretchr/testify/assert"
)

func TestRenderUserDetails(t *testing.T) {
	for _, test := range []struct {
		label string
		name  *string
		email *string
	}{
		{"profile available", github.Ptr("Alice Example"), github.Ptr("alice@example.com")},
		{"profile unavailable", nil, nil},
	} {
		t.Run(test.label, func(t *testing.T) {
			renderer := NewStringRenderer(nil)
			user := &github.User{
				Login:    github.Ptr("0123456789abcdef0123456789abcd_example"),
				Name:     test.name,
				Email:    test.email,
				RoleName: github.Ptr("member"),
			}
			assert.NoError(t, renderer.Renderer.RenderUserDetails([]*github.User{user}))
			output := renderer.Stdout.String()
			for _, header := range []string{"LOGIN", "NAME", "ROLE", "EMAIL", "SUSPENDED"} {
				assert.Contains(t, output, header)
			}
			assert.Contains(t, output, user.GetLogin())
			assert.Contains(t, output, "member")
			getter := NewUserFieldGetters()
			assert.Equal(t, user.GetName(), getter.GetField(user, "NAME"))
			assert.Equal(t, user.GetEmail(), getter.GetField(user, "EMAIL"))
			assert.Equal(t, "YES", getter.GetField(user, "SUSPENDED"))
			if test.name != nil {
				assert.Contains(t, output, *test.name)
				assert.Contains(t, output, *test.email)
			}
		})
	}
}
