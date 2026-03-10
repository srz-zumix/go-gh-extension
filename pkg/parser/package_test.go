package parser

import (
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/stretchr/testify/assert"
)

func TestParsePackageRef(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		defaultPackage string
		expectErr      bool
		expected       PackageRef
	}{
		{
			name:           "owner only",
			input:          "myowner",
			defaultPackage: "defaultpkg",
			expected:       PackageRef{Repository: repository.Repository{Owner: "myowner"}, Package: "defaultpkg"},
		},
		{
			name:           "owner/pkg",
			input:          "myowner/mypkg",
			defaultPackage: "defaultpkg",
			expected:       PackageRef{Repository: repository.Repository{Owner: "myowner"}, Package: "mypkg"},
		},
		{
			name:           "owner/scope/pkg",
			input:          "myowner/scope/mypkg",
			defaultPackage: "defaultpkg",
			expected:       PackageRef{Repository: repository.Repository{Owner: "myowner"}, Package: "scope/mypkg"},
		},
		{
			name:           "host/owner",
			input:          "ghcr.io/myowner",
			defaultPackage: "defaultpkg",
			expected:       PackageRef{Repository: repository.Repository{Host: "ghcr.io", Owner: "myowner"}, Package: "defaultpkg"},
		},
		{
			name:           "host/owner/pkg",
			input:          "ghcr.io/myowner/mypkg",
			defaultPackage: "defaultpkg",
			expected:       PackageRef{Repository: repository.Repository{Host: "ghcr.io", Owner: "myowner"}, Package: "mypkg"},
		},
		{
			name:           "host/owner/scope/pkg",
			input:          "ghcr.io/myowner/scope/mypkg",
			defaultPackage: "defaultpkg",
			expected:       PackageRef{Repository: repository.Repository{Host: "ghcr.io", Owner: "myowner"}, Package: "scope/mypkg"},
		},
		{
			name:           "empty input",
			input:          "",
			defaultPackage: "defaultpkg",
			expectErr:      true,
		},
		{
			name:           "host with empty owner",
			input:          "ghcr.io/",
			defaultPackage: "defaultpkg",
			expectErr:      true,
		},
		{
			name:           "host with empty owner before slash",
			input:          "ghcr.io//mypkg",
			defaultPackage: "defaultpkg",
			expectErr:      true,
		},
		{
			name:           "host/owner with empty package",
			input:          "ghcr.io/myowner/",
			defaultPackage: "defaultpkg",
			expectErr:      true,
		},
		{
			name:           "owner with trailing slash uses default",
			input:          "myowner/",
			defaultPackage: "defaultpkg",
			expected:       PackageRef{Repository: repository.Repository{Owner: "myowner"}, Package: "defaultpkg"},
		},
		{
			name:           "empty default package with owner only",
			input:          "myowner",
			defaultPackage: "",
			expected:       PackageRef{Repository: repository.Repository{Owner: "myowner"}, Package: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePackageRef(tt.input, tt.defaultPackage)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
