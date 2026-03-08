package gh

import (
	"testing"
	"time"

	"github.com/google/go-github/v79/github"
	"github.com/stretchr/testify/assert"
)

// makeVersion creates a PackageVersion with the given ID and creation time.
func makeVersion(id int64, createdAt *time.Time) *github.PackageVersion {
	v := &github.PackageVersion{ID: github.Ptr(id)}
	if createdAt != nil {
		v.CreatedAt = &github.Timestamp{Time: *createdAt}
	}
	return v
}

var (
	t1 = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 = time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	t3 = time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	t4 = time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
)

func TestContainerRegistry(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		expected string
	}{
		{name: "github.com", host: "github.com", expected: "ghcr.io"},
		{name: "empty host treated as github.com", host: "", expected: "ghcr.io"},
		{name: "enterprise host", host: "github.example.com", expected: "containers.github.example.com"},
		{name: "enterprise host short", host: "ghe.internal", expected: "containers.ghe.internal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ContainerRegistry(tt.host))
		})
	}
}

func TestContainerImageBase(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		owner    string
		pkg      string
		expected string
	}{
		{
			name:     "github.com lowercase",
			host:     "github.com",
			owner:    "MyOwner",
			pkg:      "MyPkg",
			expected: "ghcr.io/myowner/mypkg",
		},
		{
			name:     "empty host treated as github.com",
			host:     "",
			owner:    "Owner",
			pkg:      "pkg",
			expected: "ghcr.io/owner/pkg",
		},
		{
			name:     "enterprise host",
			host:     "ghe.internal",
			owner:    "Owner",
			pkg:      "Pkg",
			expected: "containers.ghe.internal/owner/pkg",
		},
		{
			name:     "scoped package lowercase",
			host:     "github.com",
			owner:    "Owner",
			pkg:      "Scope/Pkg",
			expected: "ghcr.io/owner/scope/pkg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ContainerImageBase(tt.host, tt.owner, tt.pkg))
		})
	}
}

func TestFilterVersions_NoFilter(t *testing.T) {
	versions := []*github.PackageVersion{
		makeVersion(1, &t1),
		makeVersion(2, &t2),
	}
	result := FilterVersions(versions, VersionFilter{})
	assert.Equal(t, versions, result)
}

func TestFilterVersions_NoFilterDoesNotMutateInput(t *testing.T) {
	versions := []*github.PackageVersion{
		makeVersion(1, &t1),
		makeVersion(2, &t2),
	}
	original := make([]*github.PackageVersion, len(versions))
	copy(original, versions)
	FilterVersions(versions, VersionFilter{})
	assert.Equal(t, original, versions)
}

func TestFilterVersions_VersionIDs(t *testing.T) {
	versions := []*github.PackageVersion{
		makeVersion(1, &t1),
		makeVersion(2, &t2),
		makeVersion(3, &t3),
	}
	result := FilterVersions(versions, VersionFilter{VersionIDs: []int64{1, 3}})
	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), *result[0].ID)
	assert.Equal(t, int64(3), *result[1].ID)
}

func TestFilterVersions_VersionIDs_NilID(t *testing.T) {
	v := &github.PackageVersion{} // no ID
	versions := []*github.PackageVersion{v, makeVersion(2, &t2)}
	result := FilterVersions(versions, VersionFilter{VersionIDs: []int64{2}})
	assert.Len(t, result, 1)
	assert.Equal(t, int64(2), *result[0].ID)
}

func TestFilterVersions_Since(t *testing.T) {
	versions := []*github.PackageVersion{
		makeVersion(1, &t1),
		makeVersion(2, &t2),
		makeVersion(3, &t3),
	}
	result := FilterVersions(versions, VersionFilter{Since: &t2})
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), *result[0].ID)
	assert.Equal(t, int64(3), *result[1].ID)
}

func TestFilterVersions_Until(t *testing.T) {
	versions := []*github.PackageVersion{
		makeVersion(1, &t1),
		makeVersion(2, &t2),
		makeVersion(3, &t3),
	}
	result := FilterVersions(versions, VersionFilter{Until: &t2})
	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), *result[0].ID)
	assert.Equal(t, int64(2), *result[1].ID)
}

func TestFilterVersions_SinceAndUntil(t *testing.T) {
	versions := []*github.PackageVersion{
		makeVersion(1, &t1),
		makeVersion(2, &t2),
		makeVersion(3, &t3),
		makeVersion(4, &t4),
	}
	result := FilterVersions(versions, VersionFilter{Since: &t2, Until: &t3})
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), *result[0].ID)
	assert.Equal(t, int64(3), *result[1].ID)
}

func TestFilterVersions_NilCreatedAt_ExcludedBySince(t *testing.T) {
	versions := []*github.PackageVersion{
		makeVersion(1, nil), // nil CreatedAt excluded by Since filter
		makeVersion(2, &t2),
	}
	result := FilterVersions(versions, VersionFilter{Since: &t1})
	assert.Len(t, result, 1)
	assert.Equal(t, int64(2), *result[0].ID)
}

func TestFilterVersions_NilCreatedAt_ExcludedByUntil(t *testing.T) {
	versions := []*github.PackageVersion{
		makeVersion(1, nil), // nil CreatedAt excluded by Until filter
		makeVersion(2, &t2),
	}
	result := FilterVersions(versions, VersionFilter{Until: &t4})
	assert.Len(t, result, 1)
	assert.Equal(t, int64(2), *result[0].ID)
}

func TestFilterVersions_Latest_Sorts(t *testing.T) {
	versions := []*github.PackageVersion{
		makeVersion(1, &t1),
		makeVersion(3, &t3),
		makeVersion(2, &t2),
	}
	result := FilterVersions(versions, VersionFilter{Latest: 2})
	assert.Len(t, result, 2)
	// newest first
	assert.Equal(t, int64(3), *result[0].ID)
	assert.Equal(t, int64(2), *result[1].ID)
}

func TestFilterVersions_Latest_LessThanCount_Sorts(t *testing.T) {
	versions := []*github.PackageVersion{
		makeVersion(1, &t1),
		makeVersion(2, &t2),
	}
	// Latest >= len: no truncation, but still sorted
	result := FilterVersions(versions, VersionFilter{Latest: 10})
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), *result[0].ID)
	assert.Equal(t, int64(1), *result[1].ID)
}

func TestFilterVersions_Latest_DoesNotMutateInput(t *testing.T) {
	versions := []*github.PackageVersion{
		makeVersion(1, &t1),
		makeVersion(3, &t3),
		makeVersion(2, &t2),
	}
	original := make([]*github.PackageVersion, len(versions))
	copy(original, versions)
	FilterVersions(versions, VersionFilter{Latest: 2})
	assert.Equal(t, original, versions)
}

func TestFilterVersions_Latest_NilCreatedAt_SortedLast(t *testing.T) {
	versions := []*github.PackageVersion{
		makeVersion(1, nil),
		makeVersion(2, &t2),
		makeVersion(3, &t3),
	}
	result := FilterVersions(versions, VersionFilter{Latest: 3})
	assert.Len(t, result, 3)
	assert.Equal(t, int64(3), *result[0].ID)
	assert.Equal(t, int64(2), *result[1].ID)
	assert.Equal(t, int64(1), *result[2].ID) // nil CreatedAt last
}

func TestFilterVersions_VersionIDsAndLatest(t *testing.T) {
	versions := []*github.PackageVersion{
		makeVersion(1, &t1),
		makeVersion(2, &t2),
		makeVersion(3, &t3),
		makeVersion(4, &t4),
	}
	result := FilterVersions(versions, VersionFilter{
		VersionIDs: []int64{1, 3, 4},
		Latest:     2,
	})
	assert.Len(t, result, 2)
	// newest first from {1,3,4}
	assert.Equal(t, int64(4), *result[0].ID)
	assert.Equal(t, int64(3), *result[1].ID)
}

func TestFilterVersions_EmptyInput(t *testing.T) {
	result := FilterVersions(nil, VersionFilter{Latest: 5, Since: &t1, Until: &t4})
	assert.Empty(t, result)
}
