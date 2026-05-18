package gitglob
package gitglob

import (
	"testing"
)

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		// --- basename-only patterns (no '/' in pattern, not anchored) ---
		{"*.go", "file.go", true},
		{"*.go", "dir/file.go", true},
		{"*.go", "a/b/c/file.go", true},
		{"*.go", "file.txt", false},
		{"file.go", "file.go", true},
		{"file.go", "dir/file.go", true},
		{"file.go", "dir/other.go", false},
		{"*.min.js", "dist/app.min.js", true},
		{"*.min.js", "dist/app.js", false},
		{"?ile.go", "file.go", true},
		{"?ile.go", "pile.go", true},
		{"?ile.go", "xile.go", true},
		{"?ile.go", "ile.go", false},

		// --- root-anchored patterns (leading '/') ---
		{"/vendor/**", "vendor/foo.go", true},
		{"/vendor/**", "vendor/a/b/c.go", true},
		{"/vendor/**", "src/vendor/foo.go", false},
		{"/gen/*.go", "gen/file.go", true},
		{"/gen/*.go", "gen/sub/file.go", false},
		{"/gen/*.go", "src/gen/file.go", false},
		{"/*.go", "main.go", true},
		{"/*.go", "dir/main.go", false},

		// --- path patterns (contain '/' but not anchored) ---
		{"vendor/**", "vendor/foo.go", true},
		{"vendor/**", "vendor/a/b/c.go", true},
		{"vendor/**", "src/vendor/foo.go", false},
		{"src/*.go", "src/main.go", true},
		{"src/*.go", "src/sub/main.go", false},
		{"src/*.go", "other/main.go", false},

		// --- globstar: **/ at start ---
		{"**/foo.go", "foo.go", true},
		{"**/foo.go", "dir/foo.go", true},
		{"**/foo.go", "a/b/foo.go", true},
		{"**/foo.go", "a/b/bar.go", false},

		// --- globstar: /**/ in middle ---
		{"src/**/file.go", "src/file.go", true},
		{"src/**/file.go", "src/a/file.go", true},
		{"src/**/file.go", "src/a/b/file.go", true},
		{"src/**/file.go", "other/a/file.go", false},

		// --- globstar: ** at end ---
		{"src/**", "src/file.go", true},
		{"src/**", "src/a/b/file.go", true},
		{"src/**", "other/file.go", false},

		// --- globstar: multiple ** ---
		{"gen/**/*.go", "gen/file.go", true},
		{"gen/**/*.go", "gen/sub/file.go", true},
		{"gen/**/*.go", "gen/a/b/c.go", true},
		{"gen/**/*.go", "gen/a/b/c.txt", false},
		{"gen/**/*.go", "other/file.go", false},
		{"a/**/b/**/*.go", "a/b/file.go", true},
		{"a/**/b/**/*.go", "a/x/b/file.go", true},
		{"a/**/b/**/*.go", "a/x/y/b/sub/file.go", true},
		{"a/**/b/**/*.go", "a/x/y/b/sub/file.txt", false},

		// --- character classes ---
		{"[abc].go", "a.go", true},
		{"[abc].go", "b.go", true},
		{"[abc].go", "d.go", false},
		{"[!abc].go", "d.go", true},
		{"[!abc].go", "a.go", false},

		// --- trailing slash stripped ---
		{"/vendor/", "vendor/foo.go", true},

		// --- literal special regex chars in path ---
		{"src/main.go", "src/main.go", true},
		{"src/main.go", "src/mainXgo", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"__"+tt.path, func(t *testing.T) {
			got := MatchPattern(tt.pattern, tt.path)
			if got != tt.want {
				t.Errorf("MatchPattern(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}
