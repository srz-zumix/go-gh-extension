package settings

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/srz-zumix/go-gh-extension/pkg/parser"
)

// UserMapping represents a single user mapping from a source login to a destination login,
// optionally identified by email address.
type UserMapping struct {
	Src   string `json:"src" yaml:"src"`
	Dst   string `json:"dst" yaml:"dst"`
	Email string `json:"email" yaml:"email"`
}

// UserMappingFile represents the YAML structure for user mappings.
type UserMappingFile struct {
	Users []UserMapping `json:"users" yaml:"users"`
}

// LoadFile reads a YAML mapping file and returns the parsed UserMappingFile.
func LoadFile(filePath string) (*UserMappingFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mapping file %q: %w", filePath, err)
	}
	var f UserMappingFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("failed to parse mapping file %q: %w", filePath, err)
	}
	return &f, nil
}

// Marshal converts a list of user mappings to YAML bytes.
func Marshal(mappings []UserMapping) ([]byte, error) {
	mappingFile := UserMappingFile{Users: mappings}
	return yaml.Marshal(mappingFile)
}

// Write serializes mappings to YAML and writes them to filePath when it is not empty.
// It always returns the marshaled YAML bytes.
func Write(filePath string, mappings []UserMapping) ([]byte, error) {
	data, err := Marshal(mappings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal mapping file: %w", err)
	}
	if filePath != "" {
		if err := os.WriteFile(filePath, data, 0o644); err != nil {
			return nil, fmt.Errorf("failed to write mapping file %q: %w", filePath, err)
		}
	}
	return data, nil
}

// compiledMapping holds a UserMapping with its pre-compiled src regex.
type compiledMapping struct {
	mapping  UserMapping
	srcRegex *regexp.Regexp
}

// CompiledMappings holds pre-compiled UserMapping entries for efficient matching.
// Plain src values (no regex metacharacters) are stored in bySrc for O(1) exact lookup.
// src values containing regex metacharacters are compiled and stored in entries for
// linear regex scanning. The dst field of regex entries may contain $N or ${name}
// capture-group references.
type CompiledMappings struct {
	entries []compiledMapping // regex-only entries
	bySrc   map[string]string // exact src -> dst for O(1) lookup
	byEmail map[string]UserMapping
}

// NewCompiledMappings builds a CompiledMappings from a UserMappingFile.
// Plain src values are stored in an exact-match map for O(1) lookup.
// src values containing regex metacharacters are compiled and kept for linear regex scanning.
func NewCompiledMappings(file *UserMappingFile) (*CompiledMappings, error) {
	if file == nil {
		return nil, fmt.Errorf("usermap: NewCompiledMappings called with nil file")
	}
	cm := &CompiledMappings{
		entries: make([]compiledMapping, 0, len(file.Users)),
		bySrc:   make(map[string]string, len(file.Users)),
		byEmail: make(map[string]UserMapping, len(file.Users)),
	}
	for _, m := range file.Users {
		if m.Dst == "" {
			slog.Warn("dst value is empty, skipping", "src", m.Src)
			continue
		} else if regexp.QuoteMeta(m.Src) == m.Src {
			// Plain literal: store in the exact-match map.
			if _, exists := cm.bySrc[m.Src]; exists {
				slog.Warn("duplicate src value in mapping file, skipping", "src", m.Src)
			} else {
				cm.bySrc[m.Src] = m.Dst
			}
		} else {
			// Contains regex metacharacters: compile and store for regex scanning.
			re, err := regexp.Compile("^(?:" + m.Src + ")$")
			if err != nil {
				return nil, fmt.Errorf("invalid regex in src field %q: %w", m.Src, err)
			}
			cm.entries = append(cm.entries, compiledMapping{mapping: m, srcRegex: re})
		}
		if m.Email != "" {
			normalizedEmail := parser.NormalizeEmail(m.Email)
			if normalizedEmail == "" {
				slog.Warn("email value is blank after trimming, skipping", "email", m.Email)
			} else if _, exists := cm.byEmail[normalizedEmail]; exists {
				slog.Warn("duplicate email value in mapping file, skipping", "email", m.Email)
			} else {
				cm.byEmail[normalizedEmail] = m
			}
		}
	}
	return cm, nil
}

// NewCompiledMappingsFromFile loads a mapping YAML file and compiles it.
func NewCompiledMappingsFromFile(filePath string) (*CompiledMappings, error) {
	f, err := LoadFile(filePath)
	if err != nil {
		return nil, err
	}
	return NewCompiledMappings(f)
}

// ResolveSrc resolves a login against src patterns, returning the dst login.
// Plain src entries are looked up in O(1) via an exact-match map.
// If no exact match is found, regex entries are scanned linearly with $N group substitution.
// Returns ("", false) if no matching entry is found.
func (c *CompiledMappings) ResolveSrc(login string) (string, bool) {
	if dst, ok := c.bySrc[login]; ok {
		return dst, true
	}
	for _, e := range c.entries {
		match := e.srcRegex.FindStringSubmatchIndex(login)
		if match != nil {
			dst := string(e.srcRegex.ExpandString(nil, e.mapping.Dst, login, match))
			return dst, true
		}
	}
	return "", false
}

// ResolveEmail returns the UserMapping for the given email address.
// Matching is case-insensitive and ignores leading/trailing whitespace.
// Returns (UserMapping{}, false) if not found.
func (c *CompiledMappings) ResolveEmail(email string) (UserMapping, bool) {
	m, ok := c.byEmail[parser.NormalizeEmail(email)]
	return m, ok
}

// SplitEMUSuffix splits an EMU login into a base and suffix by cutting at the last underscore.
// For example, "alice_corp" → ("alice", "corp") and "alice_my_corp" → ("alice_my", "corp").
// Returns (login, "") if there is no underscore or the underscore is at the end.
func SplitEMUSuffix(login string) (base, suffix string) {
	idx := strings.LastIndex(login, "_")
	if idx < 0 || idx == len(login)-1 {
		return login, ""
	}
	return login[:idx], login[idx+1:]
}

// CompactEMUMappings groups matched pairs that share the same base login (login without the
// EMU _<slug> suffix) into a single regex entry per (srcSuffix, dstSuffix) combination.
//
// Supported patterns:
//   - both have suffix, same base:  src=alice_corp  dst=alice_new  → (.+)_corp → $1_new
//   - src has suffix, dst has none: src=alice_corp  dst=alice      → (.+)_corp → $1
//   - src has none, dst has suffix: src=alice       dst=alice_new  → (.+)      → $1_new
//
// Pairs with empty dst, or where bases differ, are kept as exact entries.
func CompactEMUMappings(mappings []UserMapping) []UserMapping {
	type suffixPair struct{ src, dst string }

	seen := make(map[suffixPair]struct{})
	var specificRegexEntries []UserMapping // patterns with an explicit _suffix (e.g. (.+)_corp)
	var catchAllRegexEntries []UserMapping // catch-all patterns with no src suffix (e.g. (.+))
	var exact []UserMapping

	for _, m := range mappings {
		if m.Dst == "" {
			exact = append(exact, m)
			continue
		}
		srcBase, srcSuffix := SplitEMUSuffix(m.Src)
		dstBase, dstSuffix := SplitEMUSuffix(m.Dst)
		if srcBase != dstBase {
			exact = append(exact, m)
			continue
		}
		// Both have no suffix and same base → same login, keep as exact
		if srcSuffix == "" && dstSuffix == "" {
			exact = append(exact, m)
			continue
		}
		pair := suffixPair{src: srcSuffix, dst: dstSuffix}
		if _, ok := seen[pair]; !ok {
			seen[pair] = struct{}{}
			var srcPattern, dstPattern string
			if srcSuffix == "" {
				srcPattern = `(.+)`
			} else {
				srcPattern = `(.+)_` + regexp.QuoteMeta(pair.src)
			}
			if dstSuffix == "" {
				dstPattern = `$1`
			} else {
				dstPattern = `$1_` + strings.ReplaceAll(pair.dst, "$", "$$")
			}
			entry := UserMapping{
				Src: srcPattern,
				Dst: dstPattern,
			}
			// Catch-all patterns (no src suffix) must come after specific suffix patterns
			// to prevent shadowing more specific rules like (.+)_corp.
			if srcSuffix == "" {
				catchAllRegexEntries = append(catchAllRegexEntries, entry)
			} else {
				specificRegexEntries = append(specificRegexEntries, entry)
			}
		}
	}
	// Order: specific regex entries first, then catch-all regex entries, then exact entries.
	return append(append(specificRegexEntries, catchAllRegexEntries...), exact...)
}
