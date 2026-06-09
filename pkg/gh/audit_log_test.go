package gh

import (
	"testing"

	"github.com/google/go-github/v88/github"
	"github.com/stretchr/testify/assert"
)

func TestToGitHubGetAuditLogOptions_Nil(t *testing.T) {
	ghOpts, maxEntries := toGitHubGetAuditLogOptions(nil)
	assert.Nil(t, ghOpts)
	assert.Equal(t, -1, maxEntries)
}

func TestToGitHubGetAuditLogOptions_Defaults(t *testing.T) {
	// MaxEntries == 0 should default to -1 (unlimited).
	ghOpts, maxEntries := toGitHubGetAuditLogOptions(&GetAuditLogOptions{})
	assert.NotNil(t, ghOpts)
	assert.Nil(t, ghOpts.Phrase)
	assert.Nil(t, ghOpts.Include)
	assert.Nil(t, ghOpts.Order)
	assert.Equal(t, -1, maxEntries)
}

func TestToGitHubGetAuditLogOptions_AllFields(t *testing.T) {
	opts := &GetAuditLogOptions{
		Phrase:     "action:org.invite_member",
		Include:    "all",
		Order:      "asc",
		MaxEntries: 50,
	}
	ghOpts, maxEntries := toGitHubGetAuditLogOptions(opts)
	assert.NotNil(t, ghOpts)
	assert.Equal(t, "action:org.invite_member", *ghOpts.Phrase)
	assert.Equal(t, "all", *ghOpts.Include)
	assert.Equal(t, "asc", *ghOpts.Order)
	assert.Equal(t, 50, maxEntries)
}

func TestToGitHubGetAuditLogOptions_NegativeMaxEntries(t *testing.T) {
	// Explicit -1 should pass through unchanged.
	_, maxEntries := toGitHubGetAuditLogOptions(&GetAuditLogOptions{MaxEntries: -1})
	assert.Equal(t, -1, maxEntries)
}

func TestAuditEntryStringField_AdditionalFields(t *testing.T) {
	e := &github.AuditEntry{
		AdditionalFields: map[string]any{"repo": "owner/repo"},
	}
	assert.Equal(t, "owner/repo", AuditEntryStringField(e, "repo"))
}

func TestAuditEntryStringField_Data(t *testing.T) {
	e := &github.AuditEntry{
		Data: map[string]any{"actor": "octocat"},
	}
	assert.Equal(t, "octocat", AuditEntryStringField(e, "actor"))
}

func TestAuditEntryStringField_AdditionalFieldsTakesPrecedence(t *testing.T) {
	// AdditionalFields is checked before Data; whichever has the key first wins.
	e := &github.AuditEntry{
		AdditionalFields: map[string]any{"key": "from_additional"},
		Data:             map[string]any{"key": "from_data"},
	}
	assert.Equal(t, "from_additional", AuditEntryStringField(e, "key"))
}

func TestAuditEntryStringField_FallsBackToData(t *testing.T) {
	// When AdditionalFields has the key but value is empty string, fall back to Data.
	e := &github.AuditEntry{
		AdditionalFields: map[string]any{"key": ""},
		Data:             map[string]any{"key": "from_data"},
	}
	assert.Equal(t, "from_data", AuditEntryStringField(e, "key"))
}

func TestAuditEntryStringField_NonStringValue(t *testing.T) {
	// Non-string values in AdditionalFields should be skipped; fall back to Data.
	e := &github.AuditEntry{
		AdditionalFields: map[string]any{"count": 42},
		Data:             map[string]any{"count": "42"},
	}
	assert.Equal(t, "42", AuditEntryStringField(e, "count"))
}

func TestAuditEntryStringField_Missing(t *testing.T) {
	e := &github.AuditEntry{
		AdditionalFields: map[string]any{"other": "value"},
		Data:             map[string]any{},
	}
	assert.Equal(t, "", AuditEntryStringField(e, "missing"))
}

func TestAuditEntryStringField_NilMaps(t *testing.T) {
	e := &github.AuditEntry{}
	assert.Equal(t, "", AuditEntryStringField(e, "key"))
}

func TestAuditEntryStringField_NilEntry(t *testing.T) {
	assert.Equal(t, "", AuditEntryStringField(nil, "key"))
}
