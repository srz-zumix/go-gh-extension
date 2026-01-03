package parser

import "testing"

func TestGetNumberFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:  "valid positive number",
			input: "123",
			want:  123,
		},
		{
			name:  "valid single digit number",
			input: "1",
			want:  1,
		},
		{
			name:  "valid large number",
			input: "999999",
			want:  999999,
		},
		{
			name:  "number with hash prefix",
			input: "#456",
			want:  456,
		},
		{
			name:  "zero number",
			input: "0",
			want:  0,
		},
		{
			name:  "negative number",
			input: "-1",
			want:  -1,
		},
		{
			name:  "number with leading zeros",
			input: "000123",
			want:  123,
		},
		{
			name:  "number with plus sign",
			input: "+123",
			want:  123,
		},
		{
			name:    "non-numeric string",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "mixed alphanumeric string",
			input:   "123abc",
			wantErr: true,
		},
		{
			name:    "hash without number",
			input:   "#",
			wantErr: true,
		},
		{
			name:    "hash with non-numeric",
			input:   "#abc",
			wantErr: true,
		},
		{
			name:    "special characters",
			input:   "!@#$%",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetNumberFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNumberFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetNumberFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetIssueNumberFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:  "valid positive number",
			input: "123",
			want:  123,
		},
		{
			name:  "number with hash prefix",
			input: "#456",
			want:  456,
		},
		{
			name:  "valid issue URL",
			input: "https://github.com/owner/repo/issues/789",
			want:  789,
		},
		{
			name:  "valid pull request URL (treated as issue)",
			input: "https://github.com/owner/repo/pull/101",
			want:  101,
		},
		{
			name:    "zero number",
			input:   "0",
			wantErr: true,
		},
		{
			name:    "negative number",
			input:   "-1",
			wantErr: true,
		},
		{
			name:    "non-numeric string",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "discussion URL",
			input:   "https://github.com/owner/repo/discussions/123",
			wantErr: true,
		},
		{
			name:    "milestone URL",
			input:   "https://github.com/owner/repo/milestone/123",
			wantErr: true,
		},
		{
			name:    "malformed URL",
			input:   "https://not a valid url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetIssueNumberFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIssueNumberFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetIssueNumberFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPullRequestNumberFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:  "valid positive number",
			input: "123",
			want:  123,
		},
		{
			name:  "number with hash prefix",
			input: "#456",
			want:  456,
		},
		{
			name:  "valid pull request URL",
			input: "https://github.com/owner/repo/pull/789",
			want:  789,
		},
		{
			name:  "PR URL with query parameter",
			input: "https://github.com/owner/repo/actions/runs/123?pr=999",
			want:  999,
		},
		{
			name:    "zero number",
			input:   "0",
			wantErr: true,
		},
		{
			name:    "negative number",
			input:   "-1",
			wantErr: true,
		},
		{
			name:    "non-numeric string",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "issue URL",
			input:   "https://github.com/owner/repo/issues/123",
			wantErr: true,
		},
		{
			name:    "discussion URL",
			input:   "https://github.com/owner/repo/discussions/123",
			wantErr: true,
		},
		{
			name:    "milestone URL",
			input:   "https://github.com/owner/repo/milestone/123",
			wantErr: true,
		},
		{
			name:    "malformed URL",
			input:   "https://not a valid url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPullRequestNumberFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPullRequestNumberFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetPullRequestNumberFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDiscussionNumberFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:  "valid positive number",
			input: "123",
			want:  123,
		},
		{
			name:  "number with hash prefix",
			input: "#456",
			want:  456,
		},
		{
			name:  "valid discussion URL",
			input: "https://github.com/owner/repo/discussions/789",
			want:  789,
		},
		{
			name:    "zero number",
			input:   "0",
			wantErr: true,
		},
		{
			name:    "negative number",
			input:   "-1",
			wantErr: true,
		},
		{
			name:    "non-numeric string",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "issue URL",
			input:   "https://github.com/owner/repo/issues/123",
			wantErr: true,
		},
		{
			name:    "pull request URL",
			input:   "https://github.com/owner/repo/pull/123",
			wantErr: true,
		},
		{
			name:    "milestone URL",
			input:   "https://github.com/owner/repo/milestone/123",
			wantErr: true,
		},
		{
			name:    "malformed URL",
			input:   "https://not a valid url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDiscussionNumberFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDiscussionNumberFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetDiscussionNumberFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMilestoneNumberFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:  "valid positive number",
			input: "123",
			want:  123,
		},
		{
			name:  "number with hash prefix",
			input: "#456",
			want:  456,
		},
		{
			name:  "valid milestone URL",
			input: "https://github.com/owner/repo/milestone/789",
			want:  789,
		},
		{
			name:    "zero number",
			input:   "0",
			wantErr: true,
		},
		{
			name:    "negative number",
			input:   "-1",
			wantErr: true,
		},
		{
			name:    "non-numeric string",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "issue URL",
			input:   "https://github.com/owner/repo/issues/123",
			wantErr: true,
		},
		{
			name:    "pull request URL",
			input:   "https://github.com/owner/repo/pull/123",
			wantErr: true,
		},
		{
			name:    "discussion URL",
			input:   "https://github.com/owner/repo/discussions/123",
			wantErr: true,
		},
		{
			name:    "malformed URL",
			input:   "https://not a valid url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMilestoneNumberFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMilestoneNumberFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetMilestoneNumberFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
