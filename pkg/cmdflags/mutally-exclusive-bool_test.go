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








































































































































































}	return &bfunc boolPtr(b bool) *bool {}	}		})			}				}					t.Errorf("GetValue() = %v, want %v", *got, *tt.expected)				} else if *got != *tt.expected {					t.Errorf("GetValue() = nil, want %v", *tt.expected)				if got == nil {			} else {				}					t.Errorf("GetValue() = %v, want nil", *got)				if got != nil {			if tt.expected == nil {			got := tt.flags.GetValue()		t.Run(tt.name, func(t *testing.T) {	for _, tt := range tests {	}		},			expected: nil,			flags:    MutuallyExclusiveBoolFlags{Enabled: false, Disabled: false},			name:     "neither set",		{		},			expected: boolPtr(false),			flags:    MutuallyExclusiveBoolFlags{Enabled: false, Disabled: true},			name:     "disabled set",		{		},			expected: boolPtr(true),			flags:    MutuallyExclusiveBoolFlags{Enabled: true, Disabled: false},			name:     "enabled set",		{	}{		expected *bool		flags    MutuallyExclusiveBoolFlags		name     string	tests := []struct {func TestMutuallyExclusiveBoolFlags_GetValue(t *testing.T) {}	}		})			}				t.Errorf("IsSet() = %v, want %v", got, tt.expected)			if got := tt.flags.IsSet(); got != tt.expected {		t.Run(tt.name, func(t *testing.T) {	for _, tt := range tests {	}		},			expected: false,			flags:    MutuallyExclusiveBoolFlags{Enabled: false, Disabled: false},			name:     "neither set",		{		},			expected: true,			flags:    MutuallyExclusiveBoolFlags{Enabled: false, Disabled: true},			name:     "disabled set",		{		},			expected: true,			flags:    MutuallyExclusiveBoolFlags{Enabled: true, Disabled: false},			name:     "enabled set",		{	}{		expected bool		flags    MutuallyExclusiveBoolFlags		name     string	tests := []struct {func TestMutuallyExclusiveBoolFlags_IsSet(t *testing.T) {}	}		})			}				t.Errorf("IsFalse() = %v, want %v", got, tt.expected)			if got := tt.flags.IsFalse(); got != tt.expected {		t.Run(tt.name, func(t *testing.T) {	for _, tt := range tests {	}		},			expected: false,			flags:    MutuallyExclusiveBoolFlags{Enabled: false, Disabled: false},			name:     "disabled false",		{		},			expected: true,			flags:    MutuallyExclusiveBoolFlags{Enabled: false, Disabled: true},			name:     "disabled true",		{	}{		expected bool		flags    MutuallyExclusiveBoolFlags		name     string	tests := []struct {func TestMutuallyExclusiveBoolFlags_IsFalse(t *testing.T) {}	}		})			}				t.Errorf("IsDisabled() = %v, want %v", got, tt.expected)			if got := tt.flags.IsDisabled(); got != tt.expected {		t.Run(tt.name, func(t *testing.T) {	for _, tt := range tests {	}		},			expected: false,			flags:    MutuallyExclusiveBoolFlags{Enabled: true, Disabled: false},			name:     "disabled false",		{		},			expected: true,			flags:    MutuallyExclusiveBoolFlags{Enabled: false, Disabled: true},			name:     "disabled true",		{	}{		expected bool		flags    MutuallyExclusiveBoolFlags		name     string	tests := []struct {func TestMutuallyExclusiveBoolFlags_IsDisabled(t *testing.T) {}	}		})			}				t.Errorf("IsTrue() = %v, want %v", got, tt.expected)			if got := tt.flags.IsTrue(); got != tt.expected {		t.Run(tt.name, func(t *testing.T) {	for _, tt := range tests {	}		},			expected: false,			flags:    MutuallyExclusiveBoolFlags{Enabled: false, Disabled: true},			name:     "enabled false",		{		},			expected: true,			flags:    MutuallyExclusiveBoolFlags{Enabled: true, Disabled: false},			name:     "enabled true",		{	}{		expected bool		flags    MutuallyExclusiveBoolFlags		name     string	tests := []struct {func TestMutuallyExclusiveBoolFlags_IsTrue(t *testing.T) {}	}		})			}				t.Errorf("IsEnabled() = %v, want %v", got, tt.expected)			if got := tt.flags.IsEnabled(); got != tt.expected {		t.Run(tt.name, func(t *testing.T) {	for _, tt := range tests {	}
