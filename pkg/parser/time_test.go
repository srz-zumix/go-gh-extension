package parser

import (
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:  "RFC3339",
			input: "2026-09-01T10:00:00Z",
			want:  time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			name:  "plain date",
			input: "2026-09-01",
			want:  time.Date(2026, 9, 1, 0, 0, 0, 0, time.Local),
		},
		{
			name:    "invalid format",
			input:   "not-a-time",
			wantErr: true,
		},
		{
			name:    "invalid relative unit",
			input:   "7x",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTime(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseTime(%q) error = nil, want error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseTime(%q) error = %v", tt.input, err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("ParseTime(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseTimeRelative(t *testing.T) {
	before := time.Now()
	got, err := ParseTime("24h")
	after := time.Now()
	if err != nil {
		t.Fatalf("ParseTime(\"24h\") error = %v", err)
	}
	if got.Before(before.Add(-24 * time.Hour)) {
		t.Fatalf("ParseTime(\"24h\") = %v, want at or after %v", got, before.Add(-24*time.Hour))
	}
	if got.After(after.Add(-24 * time.Hour)) {
		t.Fatalf("ParseTime(\"24h\") = %v, want at or before %v", got, after.Add(-24*time.Hour))
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{name: "minutes", input: "30m", want: 30 * time.Minute},
		{name: "hours", input: "24h", want: 24 * time.Hour},
		{name: "days", input: "7d", want: 7 * 24 * time.Hour},
		{name: "weeks", input: "2w", want: 2 * 7 * 24 * time.Hour},
		{name: "invalid unit", input: "7x", wantErr: true},
		{name: "invalid format", input: "bogus", wantErr: true},
		{name: "overflow", input: "9223372036854775807m", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseDuration(%q) error = nil, want error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseDuration(%q) error = %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ParseDuration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParsePeriod(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{name: "days", input: "7d", want: 7 * 24 * time.Hour},
		{name: "weeks", input: "3w", want: 3 * 7 * 24 * time.Hour},
		{name: "months", input: "6m", want: 6 * 30 * 24 * time.Hour},
		{name: "years", input: "1y", want: 365 * 24 * time.Hour},
		{name: "hours are unsupported", input: "24h", wantErr: true},
		{name: "invalid format", input: "bogus", wantErr: true},
		{name: "overflow", input: "9223372036854775807y", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePeriod(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParsePeriod(%q) error = nil, want error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePeriod(%q) error = %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ParsePeriod(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsFiscalPeriod(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{input: "FY26", want: true},
		{input: "FY2026", want: true},
		{input: "FY26H1", want: true},
		{input: "FY26Q3", want: true},
		{input: "fy26q3", want: true},
		{input: "7d", want: false},
		{input: "FY26Q5", want: false},
		{input: "bogus", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsFiscalPeriod(tt.input); got != tt.want {
				t.Fatalf("IsFiscalPeriod(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseFiscalPeriod(t *testing.T) {
	date := func(y int, m time.Month, d int) time.Time {
		return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	}
	tests := []struct {
		name      string
		input     string
		wantSince time.Time
		wantUntil time.Time
		wantErr   bool
	}{
		{
			name:      "full fiscal year two digit",
			input:     "FY26",
			wantSince: date(2026, time.April, 1),
			wantUntil: date(2027, time.April, 1).Add(-time.Nanosecond),
		},
		{
			name:      "full fiscal year four digit",
			input:     "FY2026",
			wantSince: date(2026, time.April, 1),
			wantUntil: date(2027, time.April, 1).Add(-time.Nanosecond),
		},
		{
			name:      "first half",
			input:     "FY26H1",
			wantSince: date(2026, time.April, 1),
			wantUntil: date(2026, time.October, 1).Add(-time.Nanosecond),
		},
		{
			name:      "second half",
			input:     "FY26H2",
			wantSince: date(2026, time.October, 1),
			wantUntil: date(2027, time.April, 1).Add(-time.Nanosecond),
		},
		{
			name:      "first quarter",
			input:     "FY26Q1",
			wantSince: date(2026, time.April, 1),
			wantUntil: date(2026, time.July, 1).Add(-time.Nanosecond),
		},
		{
			name:      "fourth quarter crosses year",
			input:     "FY26Q4",
			wantSince: date(2027, time.January, 1),
			wantUntil: date(2027, time.April, 1).Add(-time.Nanosecond),
		},
		{
			name:      "case insensitive",
			input:     "fy26q3",
			wantSince: date(2026, time.October, 1),
			wantUntil: date(2027, time.January, 1).Add(-time.Nanosecond),
		},
		{name: "invalid quarter", input: "FY26Q5", wantErr: true},
		{name: "not fiscal", input: "7d", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			since, until, err := ParseFiscalPeriod(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseFiscalPeriod(%q) error = nil, want error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFiscalPeriod(%q) error = %v", tt.input, err)
			}
			if !since.Equal(tt.wantSince) {
				t.Fatalf("ParseFiscalPeriod(%q) since = %v, want %v", tt.input, since, tt.wantSince)
			}
			if !until.Equal(tt.wantUntil) {
				t.Fatalf("ParseFiscalPeriod(%q) until = %v, want %v", tt.input, until, tt.wantUntil)
			}
		})
	}
}
