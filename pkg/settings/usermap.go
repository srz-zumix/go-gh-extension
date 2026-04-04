package settings

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// UserMapping represents a single user mapping between two hosts.
type UserMapping struct {
	Src   string `yaml:"src"`
	Dst   string `yaml:"dst"`
	Email string `yaml:"email"`
}

// File represents the YAML structure for user mappings.
type File struct {
	Users []UserMapping `yaml:"users"`
}

// LoadFile reads a YAML mapping file and returns the parsed File.
func LoadFile(filePath string) (*File, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mapping file: %w", err)
	}
	var f File
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("failed to parse mapping file: %w", err)
	}
	return &f, nil
}

// Load reads a YAML mapping file and returns a map of src login to dst login.
func Load(filePath string) (map[string]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mapping file: %w", err)
	}

	var mappingFile File
	if err := yaml.Unmarshal(data, &mappingFile); err != nil {
		return nil, fmt.Errorf("failed to parse mapping file: %w", err)
	}

	result := make(map[string]string)
	for _, mapping := range mappingFile.Users {
		result[mapping.Src] = mapping.Dst
	}
	return result, nil
}

// Marshal converts a list of user mappings to YAML bytes.
func Marshal(mappings []UserMapping) ([]byte, error) {
	mappingFile := File{Users: mappings}
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
			return nil, fmt.Errorf("failed to write mapping file '%s': %w", filePath, err)
		}
	}
	return data, nil
}

// LoadByEmail reads a mapping file and returns a map of email to UserMapping.
func LoadByEmail(filePath string) (map[string]UserMapping, error) {
	f, err := LoadFile(filePath)
	if err != nil {
		return nil, err
	}
	result := make(map[string]UserMapping, len(f.Users))
	for _, m := range f.Users {
		if m.Email != "" {
			result[m.Email] = m
		}
	}
	return result, nil
}

// compiledMapping holds a UserMapping with its pre-compiled src regex.
type compiledMapping struct {
	mapping  UserMapping
	srcRegex *regexp.Regexp
}

// CompiledMappings holds pre-compiled UserMapping entries for efficient regex-based matching.
// The src field of each entry is treated as a full-string anchored regular expression,
// so plain strings without regex metacharacters behave as exact matches.
// The dst field may contain $N or ${name} capture-group references.
type CompiledMappings struct {
	entries []compiledMapping
	byEmail map[string]UserMapping
}

// NewCompiledMappings compiles all src fields as full-string anchored regular expressions
// and builds an email-keyed lookup table.
func NewCompiledMappings(file *File) (*CompiledMappings, error) {
	cm := &CompiledMappings{
		entries: make([]compiledMapping, 0, len(file.Users)),
		byEmail: make(map[string]UserMapping, len(file.Users)),
	}
	for _, m := range file.Users {
		re, err := regexp.Compile("^(?:" + m.Src + ")$")
		if err != nil {
			return nil, fmt.Errorf("invalid regex in src field %q: %w", m.Src, err)
		}
		cm.entries = append(cm.entries, compiledMapping{mapping: m, srcRegex: re})
		if m.Email != "" {
			cm.byEmail[m.Email] = m
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
// Exact matches are tried first (fast path); then regex patterns with $N group substitution.
// Returns ("", false) if no matching entry is found.
func (c *CompiledMappings) ResolveSrc(login string) (string, bool) {
	// First pass: exact string match
	for _, e := range c.entries {
		if e.mapping.Src == login {
			return e.mapping.Dst, true
		}
	}
	// Second pass: regex match with group substitution in dst
	for _, e := range c.entries {
		if e.srcRegex.MatchString(login) {
			dst := e.srcRegex.ReplaceAllString(login, e.mapping.Dst)
			return dst, true
		}
	}
	return "", false
}

// ResolveEmail returns the UserMapping for the given email address (exact match only).
// Returns (UserMapping{}, false) if not found.
func (c *CompiledMappings) ResolveEmail(email string) (UserMapping, bool) {
	m, ok := c.byEmail[email]
	return m, ok
}
