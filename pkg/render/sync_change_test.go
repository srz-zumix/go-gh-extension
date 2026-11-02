package render

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

func TestSyncChangeFieldGetters(t *testing.T) {
	change := gh.SyncChange{
		Destination: "github.com/example/destination",
		Action:      gh.SyncChangeUpdate,
		Target:      "sample-user",
		Permission:  "maintainer",
		Current:     "member",
	}
	getters := NewSyncChangeFieldGetters()
	for field, want := range map[string]string{
		"destination": "github.com/example/destination",
		"ACTION":      "update",
		"target":      "sample-user",
		"permission":  "maintainer",
		"current":     "member",
	} {
		if got := getters.GetField(change, field); got != want {
			t.Errorf("GetField(%q) = %q, want %q", field, got, want)
		}
	}
	if got := getters.GetField(change, "unknown"); got != "" {
		t.Errorf("GetField(unknown) = %q, want empty", got)
	}
}

func TestRenderSyncChanges(t *testing.T) {
	t.Run("renders changes", func(t *testing.T) {
		renderer := NewStringRenderer(nil)
		err := renderer.Renderer.RenderSyncChanges([]gh.SyncChange{{
			Destination: "github.com/example/destination",
			Action:      gh.SyncChangeUpdate,
			Target:      "sample-user",
			Permission:  "maintainer",
			Current:     "member",
		}}, nil)
		if err != nil {
			t.Fatalf("RenderSyncChanges() error = %v", err)
		}
		for _, want := range []string{"DESTINATION", "ACTION", "TARGET", "PERMISSION", "CURRENT", "sample-user", "maintainer"} {
			if !strings.Contains(renderer.Stdout.String(), want) {
				t.Errorf("output %q does not contain %q", renderer.Stdout.String(), want)
			}
		}
	})

	t.Run("renders empty changes", func(t *testing.T) {
		renderer := NewStringRenderer(nil)
		if err := renderer.Renderer.RenderSyncChanges(nil, nil); err != nil {
			t.Fatalf("RenderSyncChanges() error = %v", err)
		}
		if got := renderer.Stdout.String(); got != "No changes.\n" {
			t.Errorf("output = %q, want %q", got, "No changes.\n")
		}
	})

	t.Run("exports empty changes as an array", func(t *testing.T) {
		renderer := NewStringRenderer(cmdutil.NewJSONExporter())
		if err := renderer.Renderer.RenderSyncChanges(nil, nil); err != nil {
			t.Fatalf("RenderSyncChanges() error = %v", err)
		}
		if got := strings.TrimSpace(renderer.Stdout.String()); got != "[]" {
			t.Errorf("JSON output = %q, want []", got)
		}
	})

	t.Run("exports changes as JSON", func(t *testing.T) {
		renderer := NewStringRenderer(cmdutil.NewJSONExporter())
		if err := renderer.Renderer.RenderSyncChanges([]gh.SyncChange{{
			Destination: "github.com/example/destination",
			Action:      gh.SyncChangeAdd,
			Target:      "sample-user",
			Permission:  "member",
		}}, nil); err != nil {
			t.Fatalf("RenderSyncChanges() error = %v", err)
		}
		var changes []map[string]any
		if err := json.Unmarshal(renderer.Stdout.Bytes(), &changes); err != nil {
			t.Fatalf("JSON output is invalid: %v", err)
		}
		if len(changes) != 1 || changes[0]["destination"] != "github.com/example/destination" || changes[0]["action"] != "add" {
			t.Fatalf("unexpected JSON output: %s", renderer.Stdout.String())
		}
		if _, exists := changes[0]["teamOwner"]; exists {
			t.Errorf("internal team owner leaked into JSON output: %s", renderer.Stdout.String())
		}
	})
}
