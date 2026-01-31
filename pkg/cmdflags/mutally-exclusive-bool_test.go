package cmdflags

import (
	"testing"
)

func TestMutuallyExclusiveBoolFlags_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		flags    MutuallyExclusiveBoolFlags
		expected bool
	}{
		{
			name:     "enabled true",
			flags:    MutuallyExclusiveBoolFlags{Enabled: true, Disabled: false},
			expected: true,
		},
		{
			name:     "enabled false",
			flags:    MutuallyExclusiveBoolFlags{Enabled: false, Disabled: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.flags.IsEnabled(); got != tt.expected {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMutuallyExclusiveBoolFlags_IsTrue(t *testing.T) {
	tests := []struct {
		name     string
		flags    MutuallyExclusiveBoolFlags
		expected bool
	}{
		{
			name:     "enabled true",
			flags:    MutuallyExclusiveBoolFlags{Enabled: true, Disabled: false},
			expected: true,
		},
		{
			name:     "enabled false",
			flags:    MutuallyExclusiveBoolFlags{Enabled: false, Disabled: true},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.flags.IsTrue(); got != tt.expected {
				t.Errorf("IsTrue() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMutuallyExclusiveBoolFlags_IsDisabled(t *testing.T) {
	tests := []struct {
		name     string
		flags    MutuallyExclusiveBoolFlags
		expected bool
	}{
		{
			name:     "disabled true",
			flags:    MutuallyExclusiveBoolFlags{Enabled: false, Disabled: true},
			expected: true,
		},
		{
			name:     "disabled false",
			flags:    MutuallyExclusiveBoolFlags{Enabled: true, Disabled: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.flags.IsDisabled(); got != tt.expected {
				t.Errorf("IsDisabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMutuallyExclusiveBoolFlags_IsFalse(t *testing.T) {
	tests := []struct {
		name     string
		flags    MutuallyExclusiveBoolFlags
		expected bool
	}{
		{
			name:     "disabled true",
			flags:    MutuallyExclusiveBoolFlags{Enabled: false, Disabled: true},
			expected: true,
		},
		{
			name:     "disabled false",
			flags:    MutuallyExclusiveBoolFlags{Enabled: false, Disabled: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.flags.IsFalse(); got != tt.expected {
				t.Errorf("IsFalse() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMutuallyExclusiveBoolFlags_IsSet(t *testing.T) {
	tests := []struct {
		name     string
		flags    MutuallyExclusiveBoolFlags
		expected bool
	}{
		{
			name:     "enabled set",
			flags:    MutuallyExclusiveBoolFlags{Enabled: true, Disabled: false},
			expected: true,
		},
		{
			name:     "disabled set",
			flags:    MutuallyExclusiveBoolFlags{Enabled: false, Disabled: true},
			expected: true,
		},
		{
			name:     "neither set",
			flags:    MutuallyExclusiveBoolFlags{Enabled: false, Disabled: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.flags.IsSet(); got != tt.expected {
				t.Errorf("IsSet() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMutuallyExclusiveBoolFlags_GetValue(t *testing.T) {
	tests := []struct {
		name     string
		flags    MutuallyExclusiveBoolFlags
		expected *bool
	}{
		{
			name:     "enabled set",
			flags:    MutuallyExclusiveBoolFlags{Enabled: true, Disabled: false},
			expected: boolPtr(true),
		},
		{
			name:     "disabled set",
			flags:    MutuallyExclusiveBoolFlags{Enabled: false, Disabled: true},
			expected: boolPtr(false),
		},
		{
			name:     "neither set",
			flags:    MutuallyExclusiveBoolFlags{Enabled: false, Disabled: false},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.flags.GetValue()
			if tt.expected == nil {
				if got != nil {
					t.Errorf("GetValue() = %v, want nil", *got)
				}
			} else {
				if got == nil {
					t.Errorf("GetValue() = nil, want %v", *tt.expected)
				} else if *got != *tt.expected {
					t.Errorf("GetValue() = %v, want %v", *got, *tt.expected)
				}
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}
