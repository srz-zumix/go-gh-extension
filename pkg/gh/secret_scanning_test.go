package gh

import (
	"testing"
)

func TestParseCustomPattern(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantTokenType string
		wantVersion   string
		wantSetting   string
		wantErr       bool
	}{
		{
			name:          "TOKEN_TYPE=SETTING",
			input:         "MY_TOKEN=enabled",
			wantTokenType: "MY_TOKEN",
			wantVersion:   "",
			wantSetting:   "enabled",
		},
		{
			name:          "TOKEN_TYPE:VERSION=SETTING",
			input:         "MY_TOKEN:v1=disabled",
			wantTokenType: "MY_TOKEN",
			wantVersion:   "v1",
			wantSetting:   "disabled",
		},
		{
			name:          "multiple colons uses first colon as separator",
			input:         "A:B:C=enabled",
			wantTokenType: "A",
			wantVersion:   "B:C",
			wantSetting:   "enabled",
		},
		{
			name:          "multiple equals uses last equals as separator",
			input:         "A=B=enabled",
			wantTokenType: "A=B",
			wantVersion:   "",
			wantSetting:   "enabled",
		},
		{
			// Last '=' is used as the key/setting separator, so the version
			// absorbs any intermediate '=' characters from the key portion.
			name:          "TOKEN_TYPE:VERSION with multiple equals",
			input:         "TOK:v1=foo=bar",
			wantTokenType: "TOK",
			wantVersion:   "v1=foo",
			wantSetting:   "bar",
		},
		{
			name:    "no equals sign",
			input:   "TOKEN_TYPE",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "empty key (leading equals)",
			input:   "=enabled",
			wantErr: true,
		},
		{
			name:    "empty setting (trailing equals)",
			input:   "MY_TOKEN=",
			wantErr: true,
		},
		{
			name:    "colon with empty token type",
			input:   ":v1=enabled",
			wantErr: true,
		},
		{
			name:    "colon with empty version",
			input:   "MY_TOKEN:=enabled",
			wantErr: true,
		},
		{
			name:    "equals only",
			input:   "=",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseCustomPattern(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ParseCustomPattern(%q) = %+v, want error", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseCustomPattern(%q) unexpected error: %v", tc.input, err)
			}
			if got.TokenType != tc.wantTokenType {
				t.Errorf("TokenType = %q, want %q", got.TokenType, tc.wantTokenType)
			}
			if got.CustomPatternVersion != tc.wantVersion {
				t.Errorf("CustomPatternVersion = %q, want %q", got.CustomPatternVersion, tc.wantVersion)
			}
			if got.PushProtectionSetting != tc.wantSetting {
				t.Errorf("PushProtectionSetting = %q, want %q", got.PushProtectionSetting, tc.wantSetting)
			}
		})
	}
}
