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

// --- NewCompiledMappings: nil guard ---

func TestNewCompiledMappings_NilFile(t *testing.T) {
	_, err := settings.NewCompiledMappings(nil)
	assert.Error(t, err)
}

// --- NewCompiledMappings: empty dst is skipped ---

func TestNewCompiledMappings_EmptyDstSkipped(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "alice", Dst: ""},
		settings.UserMapping{Src: "bob", Dst: "bob-new"},
	))
	require.NoError(t, err)

	_, ok := cm.ResolveSrc("alice")
	assert.False(t, ok, "entry with empty dst must be skipped")

	dst, ok := cm.ResolveSrc("bob")
	assert.True(t, ok)
	assert.Equal(t, "bob-new", dst)
}

func TestNewCompiledMappings_EmptyDstRegexSkipped(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "legacy-(.*)", Dst: ""},
		settings.UserMapping{Src: "legacy-(.*)", Dst: "$1"},
	))
	require.NoError(t, err)

	// The first regex mapping is skipped because it has an empty dst.
	// Therefore, the second mapping becomes the first applicable regex entry
	// and should match.
	dst, ok := cm.ResolveSrc("legacy-alice")
	assert.True(t, ok)
	assert.Equal(t, "alice", dst)
}

// --- NewCompiledMappings: duplicate literal src (first wins) ---

func TestNewCompiledMappings_DuplicateLiteralSrcFirstWins(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "alice", Dst: "first"},
		settings.UserMapping{Src: "alice", Dst: "second"},
	))
	require.NoError(t, err)

	dst, ok := cm.ResolveSrc("alice")
	assert.True(t, ok)
	assert.Equal(t, "first", dst)
}

// --- NewCompiledMappings: duplicate email (first wins) ---

func TestNewCompiledMappings_DuplicateEmailFirstWins(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "alice", Dst: "first", Email: "alice@example.com"},
		settings.UserMapping{Src: "alice2", Dst: "second", Email: "alice@example.com"},
	))
	require.NoError(t, err)

	m, ok := cm.ResolveEmail("alice@example.com")
	assert.True(t, ok)
	assert.Equal(t, "first", m.Dst)
}

// --- NewCompiledMappings: blank email after trimming is skipped ---

func TestNewCompiledMappings_BlankEmailSkipped(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "alice", Dst: "alice-new", Email: "   "},
	))
	require.NoError(t, err)

	_, ok := cm.ResolveEmail("   ")
	assert.False(t, ok, "blank-after-trim email must not be indexed")
}

// --- NewCompiledMappings: regex src combined with email is an error ---

func TestNewCompiledMappings_RegexWithEmailError(t *testing.T) {
	_, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "legacy-(.*)", Dst: "$1", Email: "alice@example.com"},
	))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "legacy-(.*)")
}

func TestNewCompiledMappings_MultipleErrorsCollected(t *testing.T) {
	// All errors across multiple entries must be returned together, not just the first.
	_, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "[invalid", Dst: "x"},
		settings.UserMapping{Src: "legacy-(.*)", Dst: "$1", Email: "alice@example.com"},
	))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "[invalid")
	assert.Contains(t, err.Error(), "legacy-(.*)")
}

// --- ResolveEmail: case-insensitive and trims whitespace ---

func TestResolveEmail_CaseInsensitive(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "alice", Dst: "alice-new", Email: "Alice@Example.COM"},
	))
	require.NoError(t, err)

	m, ok := cm.ResolveEmail("alice@example.com")
	assert.True(t, ok)
	assert.Equal(t, "alice-new", m.Dst)

	m, ok = cm.ResolveEmail("ALICE@EXAMPLE.COM")
	assert.True(t, ok)
	assert.Equal(t, "alice-new", m.Dst)
}

func TestResolveEmail_TrimSpace(t *testing.T) {
	cm, err := settings.NewCompiledMappings(newFile(
		settings.UserMapping{Src: "alice", Dst: "alice-new", Email: "alice@example.com"},
	))
	require.NoError(t, err)

	m, ok := cm.ResolveEmail("  alice@example.com  ")
	assert.True(t, ok)
	assert.Equal(t, "alice-new", m.Dst)
}

// --- SplitEMUSuffix ---

func TestSplitEMUSuffix_WithSuffix(t *testing.T) {
	base, suffix := settings.SplitEMUSuffix("alice_corp")
	assert.Equal(t, "alice", base)
	assert.Equal(t, "corp", suffix)
}

func TestSplitEMUSuffix_MultipleSuffixes(t *testing.T) {
	// Last underscore is used as the split point.
	base, suffix := settings.SplitEMUSuffix("alice_foo_corp")
	assert.Equal(t, "alice_foo", base)
	assert.Equal(t, "corp", suffix)
}

func TestSplitEMUSuffix_NoUnderscore(t *testing.T) {
	base, suffix := settings.SplitEMUSuffix("alice")
	assert.Equal(t, "alice", base)
	assert.Equal(t, "", suffix)
}

func TestSplitEMUSuffix_TrailingUnderscore(t *testing.T) {
	// Underscore at the end is treated as no-suffix.
	base, suffix := settings.SplitEMUSuffix("alice_")
	assert.Equal(t, "alice_", base)
	assert.Equal(t, "", suffix)
}

func TestSplitEMUSuffix_Empty(t *testing.T) {
	base, suffix := settings.SplitEMUSuffix("")
	assert.Equal(t, "", base)
	assert.Equal(t, "", suffix)
}

// --- CompactEMUMappings ---

func TestCompactEMUMappings_BothSuffix(t *testing.T) {
	// alice_corp → alice_new  ⇒  (.+)_corp → $1_new
	mappings := []settings.UserMapping{
		{Src: "alice_corp", Dst: "alice_new"},
		{Src: "bob_corp", Dst: "bob_new"},
	}
	result := settings.CompactEMUMappings(mappings)
	require.Len(t, result, 1)
	assert.Equal(t, `(.+)_corp`, result[0].Src)
	assert.Equal(t, `$1_new`, result[0].Dst)
}

func TestCompactEMUMappings_SrcSuffixOnly(t *testing.T) {
	// alice_corp → alice  ⇒  (.+)_corp → $1
	mappings := []settings.UserMapping{
		{Src: "alice_corp", Dst: "alice"},
		{Src: "bob_corp", Dst: "bob"},
	}
	result := settings.CompactEMUMappings(mappings)
	require.Len(t, result, 1)
	assert.Equal(t, `(.+)_corp`, result[0].Src)
	assert.Equal(t, `$1`, result[0].Dst)
}

func TestCompactEMUMappings_DstSuffixOnly(t *testing.T) {
	// alice → alice_new  ⇒  (.+) → $1_new
	mappings := []settings.UserMapping{
		{Src: "alice", Dst: "alice_new"},
		{Src: "bob", Dst: "bob_new"},
	}
	result := settings.CompactEMUMappings(mappings)
	require.Len(t, result, 1)
	assert.Equal(t, `(.+)`, result[0].Src)
	assert.Equal(t, `$1_new`, result[0].Dst)
}

func TestCompactEMUMappings_NoSuffix_SameLogin(t *testing.T) {
	// alice → alice  ⇒  kept as exact entry
	mappings := []settings.UserMapping{
		{Src: "alice", Dst: "alice"},
	}
	result := settings.CompactEMUMappings(mappings)
	require.Len(t, result, 1)
	assert.Equal(t, "alice", result[0].Src)
	assert.Equal(t, "alice", result[0].Dst)
}

func TestCompactEMUMappings_BaseMismatch(t *testing.T) {
	// Different base: alice_corp → carol_new  ⇒  kept as exact
	mappings := []settings.UserMapping{
		{Src: "alice_corp", Dst: "carol_new"},
	}
	result := settings.CompactEMUMappings(mappings)
	require.Len(t, result, 1)
	assert.Equal(t, "alice_corp", result[0].Src)
	assert.Equal(t, "carol_new", result[0].Dst)
}

func TestCompactEMUMappings_EmptyDst(t *testing.T) {
	// Empty dst is kept as exact entry.
	mappings := []settings.UserMapping{
		{Src: "alice_corp", Dst: ""},
	}
	result := settings.CompactEMUMappings(mappings)
	require.Len(t, result, 1)
	assert.Equal(t, "alice_corp", result[0].Src)
}

func TestCompactEMUMappings_DeduplicateSuffixPair(t *testing.T) {
	// Same suffix pair seen twice: only one regex entry generated.
	mappings := []settings.UserMapping{
		{Src: "alice_corp", Dst: "alice_new"},
		{Src: "bob_corp", Dst: "bob_new"},
		{Src: "charlie_corp", Dst: "charlie_new"},
	}
	result := settings.CompactEMUMappings(mappings)
	require.Len(t, result, 1)
	assert.Equal(t, `(.+)_corp`, result[0].Src)
	assert.Equal(t, `$1_new`, result[0].Dst)
}

func TestCompactEMUMappings_MultipleSuffixPairs(t *testing.T) {
	// Two distinct suffix pairs produce two regex entries.
	mappings := []settings.UserMapping{
		{Src: "alice_corp", Dst: "alice_new"},
		{Src: "bob_old", Dst: "bob_prod"},
	}
	result := settings.CompactEMUMappings(mappings)
	require.Len(t, result, 2)
	// Regex entries come first; order matches insertion order.
	srcs := []string{result[0].Src, result[1].Src}
	assert.Contains(t, srcs, `(.+)_corp`)
	assert.Contains(t, srcs, `(.+)_old`)
}

func TestCompactEMUMappings_ExactEntriesBeforeRegex(t *testing.T) {
	// Exact entries come first in the result slice, followed by regex entries.
	// In NewCompiledMappings, exact entries are stored in the bySrc map (O(1) lookup)
	// so their position relative to regex entries does not affect matching priority.
	mappings := []settings.UserMapping{
		{Src: "alice", Dst: "carol"},      // base mismatch → exact
		{Src: "bob_corp", Dst: "bob_new"}, // regex
	}
	result := settings.CompactEMUMappings(mappings)
	require.Len(t, result, 2)
	assert.Equal(t, "alice", result[0].Src, "exact entry must come first")
	assert.Equal(t, `(.+)_corp`, result[1].Src, "regex entry must come second")
}

func TestCompactEMUMappings_Empty(t *testing.T) {
	result := settings.CompactEMUMappings(nil)
	assert.Empty(t, result)
}

func TestCompactEMUMappings_MultipleCatchAllDstSuffixes(t *testing.T) {
	// When srcSuffix=="" candidates have more than one distinct dst suffix,
	// a single (.+) regex would shadow later (.+) rules, changing semantics.
	// All such candidates must be kept as exact entries instead.
	mappings := []settings.UserMapping{
		{Src: "alice", Dst: "alice_new"},
		{Src: "bob", Dst: "bob_prod"},
	}
	result := settings.CompactEMUMappings(mappings)
	require.Len(t, result, 2)
	for _, r := range result {
		assert.NotEqual(t, `(.+)`, r.Src, "no catch-all regex should be emitted")
	}
	srcs := []string{result[0].Src, result[1].Src}
	assert.Contains(t, srcs, "alice")
	assert.Contains(t, srcs, "bob")
}

func TestCompactEMUMappings_MultipleCatchAllDstSuffixes_OrderPreserved(t *testing.T) {
	// When falling back to exact entries, the original input order must be preserved,
	// including interleaving with regular exact entries.
	mappings := []settings.UserMapping{
		{Src: "alice", Dst: "alice_new"},  // catch-all candidate → exact fallback
		{Src: "zara", Dst: "carol"},       // base mismatch → exact
		{Src: "bob", Dst: "bob_prod"},     // catch-all candidate → exact fallback
	}
	result := settings.CompactEMUMappings(mappings)
	require.Len(t, result, 3)
	assert.Equal(t, "alice", result[0].Src)
	assert.Equal(t, "zara", result[1].Src)
	assert.Equal(t, "bob", result[2].Src)
}
