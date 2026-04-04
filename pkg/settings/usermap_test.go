package settings_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/srz-zumix/go-gh-extension/pkg/settings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- helpers ---

func newFile(users ...settings.UserMapping) *settings.UserMappingFile {
	return &settings.UserMappingFile{Users: users}
}

func writeYAML(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "usermap-*.yaml")
	require.NoError(t, err)
	_, err = f.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

// --- LoadFile ---

func TestLoadFile(t *testing.T) {
	path := writeYAML(t, `users:
  - src: alice
    dst: alice-new
    email: alice@example.com
  - src: bob
    dst: bob-new
`)
	f, err := settings.LoadFile(path)
	require.NoError(t, err)
	require.Len(t, f.Users, 2)
	assert.Equal(t, "alice", f.Users[0].Src)
	assert.Equal(t, "alice-new", f.Users[0].Dst)
	assert.Equal(t, "alice@example.com", f.Users[0].Email)
	assert.Equal(t, "bob", f.Users[1].Src)
	assert.Equal(t, "bob-new", f.Users[1].Dst)
}

func TestLoadFile_NotFound(t *testing.T) {
	_, err := settings.LoadFile(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	assert.Error(t, err)
}

func TestLoadFile_InvalidYAML(t *testing.T) {
	path := writeYAML(t, ":\ninvalid: [yaml")
	_, err := settings.LoadFile(path)
	assert.Error(t, err)
}

// --- Load ---

func TestLoad(t *testing.T) {
	path := writeYAML(t, `users:
  - src: alice
    dst: alice-new
  - src: bob
    dst: bob-new
`)
	m, err := settings.Load(path)
	require.NoError(t, err)
	assert.Equal(t, "alice-new", m["alice"])
	assert.Equal(t, "bob-new", m["bob"])
}

// --- Marshal / Write ---

func TestMarshal(t *testing.T) {
	mappings := []settings.UserMapping{
		{Src: "alice", Dst: "alice-new", Email: "alice@example.com"},
	}
	data, err := settings.Marshal(mappings)
	require.NoError(t, err)
	assert.Contains(t, string(data), "alice")
	assert.Contains(t, string(data), "alice-new")
	assert.Contains(t, string(data), "alice@example.com")
}

func TestWrite_ToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.yaml")
	mappings := []settings.UserMapping{
		{Src: "alice", Dst: "alice-new"},
	}
	data, err := settings.Write(path, mappings)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify written content is readable
	f, err := settings.LoadFile(path)
	require.NoError(t, err)
	require.Len(t, f.Users, 1)
	assert.Equal(t, "alice", f.Users[0].Src)
}

func TestWrite_NoFile(t *testing.T) {
	mappings := []settings.UserMapping{
		{Src: "alice", Dst: "alice-new"},
	}
	// filePath="" does not write to a file and should still return YAML bytes
	data, err := settings.Write("", mappings)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

// --- LoadByEmail ---

func TestLoadByEmail(t *testing.T) {
	path := writeYAML(t, `users:
  - src: alice
    dst: alice-new
    email: alice@example.com
  - src: bob
    dst: bob-new
`)
	m, err := settings.LoadByEmail(path)
	require.NoError(t, err)
	require.Contains(t, m, "alice@example.com")
	assert.Equal(t, "alice", m["alice@example.com"].Src)
	// bob has no email, so it must not appear
	assert.Len(t, m, 1)
}

// --- NewCompiledMappings ---

func TestNewCompiledMappings_InvalidRegex(t *testing.T) {
	_, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "[invalid", Dst: "x"},
	))
	assert.Error(t, err)
}

// --- ResolveSrc: exact match ---

func TestResolveSrc_ExactMatch(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "alice", Dst: "alice-new"},
	))
	require.NoError(t, err)

	dst, ok := cm.ResolveSrc("alice")
	assert.True(t, ok)
	assert.Equal(t, "alice-new", dst)
}

func TestResolveSrc_NoMatch(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "alice", Dst: "alice-new"},
	))
	require.NoError(t, err)

	_, ok := cm.ResolveSrc("charlie")
	assert.False(t, ok)
}

// --- ResolveSrc: regex with group references ---

func TestResolveSrc_RegexCapture(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "legacy-(.*)", Dst: "$1"},
	))
	require.NoError(t, err)

	dst, ok := cm.ResolveSrc("legacy-alice")
	assert.True(t, ok)
	assert.Equal(t, "alice", dst)
}

func TestResolveSrc_RegexNamedCapture(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "(?P<name>[a-z]+)-old", Dst: "${name}-new"},
	))
	require.NoError(t, err)

	dst, ok := cm.ResolveSrc("bob-old")
	assert.True(t, ok)
	assert.Equal(t, "bob-new", dst)
}

func TestResolveSrc_RegexNoMatch(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "legacy-(.*)", Dst: "$1"},
	))
	require.NoError(t, err)

	_, ok := cm.ResolveSrc("modern-alice")
	assert.False(t, ok)
}

// Exact match must take priority over regex when both could match.
func TestResolveSrc_ExactBeforeRegex(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: ".*", Dst: "wildcard"},
		settings.UserMapping{Src: "alice", Dst: "alice-specific"},
	))
	require.NoError(t, err)

	dst, ok := cm.ResolveSrc("alice")
	assert.True(t, ok)
	assert.Equal(t, "alice-specific", dst)
}

// When multiple regex patterns match, the first entry in the YAML (file order) wins.
func TestResolveSrc_RegexFirstMatchWins(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "prefix-(.*)", Dst: "first-$1"},
		settings.UserMapping{Src: "prefix-(.*)", Dst: "second-$1"},
	))
	require.NoError(t, err)

	dst, ok := cm.ResolveSrc("prefix-alice")
	assert.True(t, ok)
	assert.Equal(t, "first-alice", dst)
}

// When multiple regex patterns could match, the earlier entry in the YAML wins.
func TestResolveSrc_RegexOrderPreserved(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "a.*", Dst: "matched-by-a"},
		settings.UserMapping{Src: ".*lice", Dst: "matched-by-lice"},
	))
	require.NoError(t, err)

	// "alice" matches both patterns; the first entry must win.
	dst, ok := cm.ResolveSrc("alice")
	assert.True(t, ok)
	assert.Equal(t, "matched-by-a", dst)

	// "bob" matches neither entry.
	_, ok = cm.ResolveSrc("bob")
	assert.False(t, ok)

	// "malice" only matches the second pattern.
	dst, ok = cm.ResolveSrc("malice")
	assert.True(t, ok)
	assert.Equal(t, "matched-by-lice", dst)
}

// File order is preserved when loaded via NewCompiledMappingsFromFile.
func TestResolveSrc_FileOrderPreserved(t *testing.T) {
	path := writeYAML(t, `users:
  - src: "a.*"
    dst: first
  - src: ".*lice"
    dst: second
`)
	cm, err := settings.NewCompiledMappingsFromFile(path)
	require.NoError(t, err)

	dst, ok := cm.ResolveSrc("alice")
	assert.True(t, ok)
	assert.Equal(t, "first", dst)
}

// Regex must not match a partial string because matching is anchored with ^...$.
func TestResolveSrc_FullStringAnchor(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "(ali)", Dst: "x"},
	))
	require.NoError(t, err)

	dst, ok := cm.ResolveSrc("ali")
	assert.True(t, ok)
	assert.Equal(t, "x", dst)

	_, ok = cm.ResolveSrc("alice")
	assert.False(t, ok)
}

// Empty src matches only an empty login.
func TestResolveSrc_EmptySrc(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "", Dst: "nobody"},
	))
	require.NoError(t, err)

	dst, ok := cm.ResolveSrc("")
	assert.True(t, ok)
	assert.Equal(t, "nobody", dst)

	_, ok = cm.ResolveSrc("alice")
	assert.False(t, ok)
}

// --- ResolveEmail ---

func TestResolveEmail_Found(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "alice", Dst: "alice-new", Email: "alice@example.com"},
	))
	require.NoError(t, err)

	m, ok := cm.ResolveEmail("alice@example.com")
	assert.True(t, ok)
	assert.Equal(t, "alice-new", m.Dst)
}

func TestResolveEmail_NotFound(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "alice", Dst: "alice-new", Email: "alice@example.com"},
	))
	require.NoError(t, err)

	_, ok := cm.ResolveEmail("other@example.com")
	assert.False(t, ok)
}

func TestResolveEmail_EmptyEmailNotIndexed(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "alice", Dst: "alice-new"},
	))
	require.NoError(t, err)

	_, ok := cm.ResolveEmail("")
	assert.False(t, ok)
}

// --- NewCompiledMappingsFromFile ---

func TestNewCompiledMappingsFromFile(t *testing.T) {
	path := writeYAML(t, `users:
  - src: alice
    dst: alice-new
    email: alice@example.com
`)
	cm, err := settings.NewCompiledMappingsFromFile(path)
	require.NoError(t, err)

	dst, ok := cm.ResolveSrc("alice")
	assert.True(t, ok)
	assert.Equal(t, "alice-new", dst)

	m, ok := cm.ResolveEmail("alice@example.com")
	assert.True(t, ok)
	assert.Equal(t, "alice-new", m.Dst)
}

func TestNewCompiledMappingsFromFile_NotFound(t *testing.T) {
	_, err := settings.NewCompiledMappingsFromFile(filepath.Join(t.TempDir(), "missing.yaml"))
	assert.Error(t, err)
}
