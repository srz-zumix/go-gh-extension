package parser

import (
	"testing"
)

func TestValidateTokenSecretName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		// Known token prefixes should be rejected.
		{"classic_pat", "ghp_ABCDEFGHIJKLMNOPabcdefg", true},
		{"fine_grained_pat", "github_pat_ABCDEFGHIJKLMNOP_abcdefg", true},
		{"oauth_app_token", "gho_ABCDEFGHIJKLMNOPabcdefg", true},
		{"actions_token", "ghs_ABCDEFGHIJKLMNOPabcdefg", true},
		{"refresh_token", "ghr_ABCDEFGHIJKLMNOPabcdefg", true},

		// Typical secret names should be accepted.
		{"uppercase_name", "MY_SECRET", false},
		{"uppercase_with_number", "MY_TOKEN_123", false},
		{"simple_name", "SECRET", false},
		{"name_with_ghp_elsewhere", "NOT_ghp_PREFIX", false},
		{"empty_string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTokenSecretName(tt.input)
			if tt.wantError && err == nil {
				t.Errorf("ValidateTokenSecretName(%q) = nil; want error", tt.input)
			}
			if !tt.wantError && err != nil {
				t.Errorf("ValidateTokenSecretName(%q) = %v; want nil", tt.input, err)
			}
		})
	}
}
