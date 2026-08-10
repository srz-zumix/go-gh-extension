package parser

import "testing"

func TestUUID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "valid lowercase uuid",
			input: "00000000-0000-0000-0000-000000000000",
			want:  "00000000-0000-0000-0000-000000000000",
		},
		{
			name:  "valid uppercase uuid is normalized to lowercase",
			input: "A1B2C3D4-E5F6-7890-ABCD-EF1234567890",
			want:  "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		},
		{
			name:  "surrounding whitespace is trimmed and normalized",
			input: "  A1B2C3D4-E5F6-7890-ABCD-EF1234567890  ",
			want:  "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		},
		{
			name:    "missing hyphens is invalid",
			input:   "00000000000000000000000000000000",
			wantErr: true,
		},
		{
			name:    "wrong segment length is invalid",
			input:   "00000000-0000-0000-0000-00000000000",
			wantErr: true,
		},
		{
			name:    "non-hex characters is invalid",
			input:   "gggggggg-gggg-gggg-gggg-gggggggggggg",
			wantErr: true,
		},
		{
			name:    "empty string is invalid",
			input:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UUID(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
