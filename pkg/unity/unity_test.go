package unity

import "testing"

func TestToPackages_Nil(t *testing.T) {
	var m *UnityManifest
	if got := m.ToPackages(); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestToPackages_Empty(t *testing.T) {
	m := &UnityManifest{Dependencies: map[string]string{}}
	got := m.ToPackages()
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestToPackages_Semver(t *testing.T) {
	m := &UnityManifest{Dependencies: map[string]string{
		"com.unity.inputsystem": "1.5.0",
	}}
	got := m.ToPackages()
	if len(got) != 1 {
		t.Fatalf("expected 1 package, got %d", len(got))
	}
	pkg := got[0]
	if pkg.Name != "com.unity.inputsystem" {
		t.Errorf("Name: got %q, want %q", pkg.Name, "com.unity.inputsystem")
	}
	if pkg.Version != "1.5.0" {
		t.Errorf("Version: got %q, want %q", pkg.Version, "1.5.0")
	}
	if pkg.Path != "" {
		t.Errorf("Path should be empty, got %q", pkg.Path)
	}
	if pkg.URL != "" {
		t.Errorf("URL should be empty, got %q", pkg.URL)
	}
}

func TestToPackages_FilePath(t *testing.T) {
	m := &UnityManifest{Dependencies: map[string]string{
		"com.example.local": "file:../LocalPackage",
	}}
	got := m.ToPackages()
	if len(got) != 1 {
		t.Fatalf("expected 1 package, got %d", len(got))
	}
	pkg := got[0]
	if pkg.Path != "../LocalPackage" {
		t.Errorf("Path: got %q, want %q", pkg.Path, "../LocalPackage")
	}
	if pkg.Version != "" {
		t.Errorf("Version should be empty for file: dep, got %q", pkg.Version)
	}
	if pkg.URL != "" {
		t.Errorf("URL should be empty, got %q", pkg.URL)
	}
}

func TestToPackages_URL_WithFragment(t *testing.T) {
	m := &UnityManifest{Dependencies: map[string]string{
		"com.example.remote": "https://github.com/example/package.git#1.2.3",
	}}
	got := m.ToPackages()
	if len(got) != 1 {
		t.Fatalf("expected 1 package, got %d", len(got))
	}
	pkg := got[0]
	want := "https://github.com/example/package.git"
	if pkg.URL != want {
		t.Errorf("URL: got %q, want %q", pkg.URL, want)
	}
	if pkg.Version != "1.2.3" {
		t.Errorf("Version: got %q, want %q", pkg.Version, "1.2.3")
	}
	if pkg.Path != "" {
		t.Errorf("Path should be empty, got %q", pkg.Path)
	}
}

func TestToPackages_URL_WithoutFragment(t *testing.T) {
	m := &UnityManifest{Dependencies: map[string]string{
		"com.example.remote": "https://github.com/example/package.git",
	}}
	got := m.ToPackages()
	if len(got) != 1 {
		t.Fatalf("expected 1 package, got %d", len(got))
	}
	pkg := got[0]
	want := "https://github.com/example/package.git"
	if pkg.URL != want {
		t.Errorf("URL: got %q, want %q", pkg.URL, want)
	}
	if pkg.Version != "" {
		t.Errorf("Version should be empty when no fragment, got %q", pkg.Version)
	}
}

func TestToPackages_GitPlus(t *testing.T) {
	m := &UnityManifest{Dependencies: map[string]string{
		"com.example.gitplus": "git+https://github.com/example/package.git#v2.0.0",
	}}
	got := m.ToPackages()
	if len(got) != 1 {
		t.Fatalf("expected 1 package, got %d", len(got))
	}
	pkg := got[0]
	if pkg.Version != "v2.0.0" {
		t.Errorf("Version: got %q, want %q", pkg.Version, "v2.0.0")
	}
	wantURL := "git+https://github.com/example/package.git"
	if pkg.URL != wantURL {
		t.Errorf("URL: got %q, want %q", pkg.URL, wantURL)
	}
}

func TestToPackages_SortedByName(t *testing.T) {
	m := &UnityManifest{Dependencies: map[string]string{
		"com.z.last":   "1.0.0",
		"com.a.first":  "2.0.0",
		"com.m.middle": "3.0.0",
	}}
	got := m.ToPackages()
	if len(got) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(got))
	}
	names := []string{got[0].Name, got[1].Name, got[2].Name}
	want := []string{"com.a.first", "com.m.middle", "com.z.last"}
	for i, w := range want {
		if names[i] != w {
			t.Errorf("index %d: got %q, want %q", i, names[i], w)
		}
	}
}

func TestIsURL(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"http://example.com/pkg.git", true},
		{"https://github.com/foo/bar.git", true},
		{"git+https://github.com/foo/bar.git", true},
		{"ssh://git@github.com/foo/bar.git", true},
		{"git://github.com/foo/bar.git", true},
		{"1.2.3", false},
		{"file:../local", false},
		{"", false},
		{"com.unity.modules.physics", false},
	}
	for _, tc := range cases {
		got := isURL(tc.input)
		if got != tc.want {
			t.Errorf("isURL(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestIsCompressedFile(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"package.tgz", true},
		{"archive.zip", true},
		{"dist.tar.gz", true},
		{"dist.tar.bz2", true},
		{"dist.tar.xz", true},
		{"file.gz", true},
		{"file.bz2", true},
		{"../LocalPackage", false},
		{"./EmbeddedPackage", false},
		{"1.2.3", false},
		{"", false},
	}
	for _, tc := range cases {
		got := isCompressedFile(tc.input)
		if got != tc.want {
			t.Errorf("isCompressedFile(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}
