package parser

import (
	"testing"
)

func TestIsEnableEnvFlag(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"ENABLED_TRUE", "true", true},
		{"ENABLED_1", "1", true},
		{"ENABLED_YES", "yes", true},
		{"ENABLED_ON", "on", true},
		{"DISABLED_FALSE", "false", false},
		{"DISABLED_0", "0", false},
		{"DISABLED_NO", "no", false},
		{"DISABLED_OFF", "off", false},
		{"EMPTY", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.name, tt.envValue)
			result := IsEnableEnvFlag(tt.name)
			if result != tt.expected {
				t.Errorf("IsEnableEnvFlag(%s) = %v; want %v", tt.envValue, result, tt.expected)
			}
		})
	}
}

func TestIsDisableEnvFlag(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"DISABLED_FALSE", "false", true},
		{"DISABLED_0", "0", true},
		{"DISABLED_NO", "no", true},
		{"DISABLED_OFF", "off", true},
		{"ENABLED_TRUE", "true", false},
		{"ENABLED_1", "1", false},
		{"ENABLED_YES", "yes", false},
		{"ENABLED_ON", "on", false},
		{"EMPTY", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.name, tt.envValue)
			result := IsDisableEnvFlag(tt.name)
			if result != tt.expected {
				t.Errorf("IsDisableEnvFlag(%s) = %v; want %v", tt.envValue, result, tt.expected)
			}
		})
	}
}
