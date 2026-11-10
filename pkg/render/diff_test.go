package render

import (
	"errors"
	"fmt"
	"testing"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/fatih/color"
	"github.com/google/go-github/v93/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

func TestRenderDiffEmpty(t *testing.T) {
	tests := []struct {
		name        string
		diff        any
		left, right any
	}{
		{
			name:  "nil repository differences",
			diff:  gh.RepositoryPermissionsDiffs(nil),
			left:  "team-a",
			right: "team-b",
		},
		{
			name:  "empty repository differences",
			diff:  gh.RepositoryPermissionsDiffs{},
			left:  "team-a",
			right: "team-b",
		},
		{
			name:  "nil team differences",
			diff:  gh.TeamPermissionsDiffs(nil),
			left:  repository.Repository{Owner: "example-org", Name: "repo-a"},
			right: repository.Repository{Owner: "example-org", Name: "repo-b"},
		},
		{
			name:  "empty team differences",
			diff:  gh.TeamPermissionsDiffs{},
			left:  repository.Repository{Owner: "example-org", Name: "repo-a"},
			right: repository.Repository{Owner: "example-org", Name: "repo-b"},
		},
	}
	for _, tt := range tests {
		for _, colored := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/color=%t", tt.name, colored), func(t *testing.T) {
				renderer := NewStringRenderer(nil)
				renderer.Renderer.Color = colored
				builder := func(left, right, target string) string {
					t.Fatal("command builder called for empty differences")
					return ""
				}
				if err := renderer.Renderer.RenderDiff(tt.diff, tt.left, tt.right, builder); err != nil {
					t.Fatalf("RenderDiff() error = %v", err)
				}
				if got := renderer.Stdout.String(); got != "" {
					t.Errorf("stdout = %q, want no output", got)
				}
				if got := renderer.Stderr.String(); got != "" {
					t.Errorf("stderr = %q, want no output", got)
				}
			})
		}

		t.Run(tt.name+"/JSON", func(t *testing.T) {
			renderer := NewStringRenderer(cmdutil.NewJSONExporter())
			if err := renderer.Renderer.RenderDiff(tt.diff, tt.left, tt.right, nil); err != nil {
				t.Fatalf("RenderDiff() error = %v", err)
			}
			if got := renderer.Stdout.String(); got != "[]\n" {
				t.Errorf("stdout = %q, want an empty JSON array", got)
			}
		})
	}
}

func TestRenderDiffNonEmpty(t *testing.T) {
	tests := []struct {
		name        string
		diff        any
		left, right any
		header      string
	}{
		{
			name: "repository differences",
			diff: gh.RepositoryPermissionsDiffs{{
				Left:  &github.Repository{Name: github.Ptr("shared"), FullName: github.Ptr("example-org/shared"), RoleName: github.Ptr("pull")},
				Right: &github.Repository{Name: github.Ptr("shared"), FullName: github.Ptr("example-org/shared"), RoleName: github.Ptr("push")},
			}},
			left:  "team-a",
			right: "team-b",
			header: "diff --example compare team-a team-b example-org/shared\n" +
				"--- example-org/shared team-a\n+++ example-org/shared team-b\n",
		},
		{
			name: "team differences",
			diff: gh.TeamPermissionsDiffs{{
				Left:  &github.Team{Slug: github.Ptr("sample-team"), Permission: github.Ptr("pull")},
				Right: &github.Team{Slug: github.Ptr("sample-team"), Permission: github.Ptr("push")},
			}},
			left:  repository.Repository{Owner: "example-org", Name: "repo-a"},
			right: repository.Repository{Owner: "example-org", Name: "repo-b"},
			header: "diff --example compare example-org/repo-a example-org/repo-b sample-team\n" +
				"--- sample-team example-org/repo-a\n+++ sample-team example-org/repo-b\n",
		},
	}
	builder := func(left, right, target string) string {
		return fmt.Sprintf("example compare %s %s %s", left, right, target)
	}
	for _, tt := range tests {
		for _, colored := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/color=%t", tt.name, colored), func(t *testing.T) {
				renderer := NewStringRenderer(nil)
				renderer.Renderer.Color = colored
				if err := renderer.Renderer.RenderDiff(tt.diff, tt.left, tt.right, builder); err != nil {
					t.Fatalf("RenderDiff() error = %v", err)
				}
				want := tt.header + "- pull\n+ push\n\n"
				if colored {
					want = tt.header + color.RedString("- pull") + "\n" + color.GreenString("+ push") + "\n\n\n"
				}
				if got := renderer.Stdout.String(); got != want {
					t.Errorf("stdout = %q, want %q", got, want)
				}
			})
		}
	}
}

type failingDiffExporter struct {
	err error
}

func (e failingDiffExporter) Fields() []string {
	return nil
}

func (e failingDiffExporter) Write(_ *iostreams.IOStreams, _ any) error {
	return e.err
}

func TestRenderDiffEmptyExporterError(t *testing.T) {
	want := errors.New("export failed")
	renderer := NewStringRenderer(failingDiffExporter{err: want})
	err := renderer.Renderer.RenderDiff(gh.RepositoryPermissionsDiffs(nil), "team-a", "team-b", nil)
	if !errors.Is(err, want) {
		t.Errorf("RenderDiff() error = %v, want %v", err, want)
	}
	if got := renderer.Stdout.String(); got != "" {
		t.Errorf("stdout = %q, want no output", got)
	}
}
