package gh

import (
	"testing"
	"time"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
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

func TestDockerRegistry(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		expected string
	}{
		{name: "github.com", host: "github.com", expected: "docker.pkg.github.com"},
		{name: "empty host treated as github.com", host: "", expected: "docker.pkg.github.com"},
		{name: "enterprise host", host: "github.example.com", expected: "docker.pkg.github.example.com"},
		{name: "enterprise host short", host: "ghe.internal", expected: "docker.pkg.ghe.internal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, DockerRegistry(tt.host))
		})
	}
}

func TestDockerImageBase(t *testing.T) {
	tests := []struct {
		name     string
		repo     repository.Repository
		pkg      string
		expected string
	}{
		{
			name:     "github.com",
			repo:     repository.Repository{Host: "github.com", Owner: "MyOrg", Name: "repo"},
			pkg:      "MyImage",
			expected: "docker.pkg.github.com/myorg/repo/myimage",
		},
		{
			name:     "empty host treated as github.com",
			repo:     repository.Repository{Host: "", Owner: "MyOrg", Name: "repo"},
			pkg:      "MyImage",
			expected: "docker.pkg.github.com/myorg/repo/myimage",
		},
		{
			name:     "enterprise host",
			repo:     repository.Repository{Host: "ghe.internal", Owner: "MyOrg", Name: "repo"},
			pkg:      "MyImage",
			expected: "docker.pkg.ghe.internal/myorg/repo/myimage",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, DockerImageBase(tt.repo, tt.pkg))
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

func TestFilterVersions_VersionIDs_PreservesInputOrder(t *testing.T) {
	// Result order must follow the input slice, not the order of filter.VersionIDs.
	versions := []*github.PackageVersion{
		makeVersion(3, &t3),
		makeVersion(1, &t1),
		makeVersion(2, &t2),
	}
	result := FilterVersions(versions, VersionFilter{VersionIDs: []int64{1, 3}})
	assert.Len(t, result, 2)
	assert.Equal(t, int64(3), *result[0].ID) // 3 appears first in the input
	assert.Equal(t, int64(1), *result[1].ID)
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

func TestFilterVersions_Names(t *testing.T) {
	versions := []*github.PackageVersion{
		{ID: github.Ptr(int64(1)), Name: github.Ptr("v1.0"), CreatedAt: &github.Timestamp{Time: t1}},
		{ID: github.Ptr(int64(2)), Name: github.Ptr("v2.0"), CreatedAt: &github.Timestamp{Time: t2}},
		{ID: github.Ptr(int64(3)), Name: github.Ptr("v3.0"), CreatedAt: &github.Timestamp{Time: t3}},
	}
	result := FilterVersions(versions, VersionFilter{Names: []string{"v3.0", "v1.0"}})
	assert.Len(t, result, 2)
	// Results should preserve the input slice order, not the filter.Names order.
	assert.Equal(t, "v1.0", result[0].GetName())
	assert.Equal(t, "v3.0", result[1].GetName())
}

func TestFilterVersions_Names_NoMatch(t *testing.T) {
	versions := []*github.PackageVersion{
		{ID: github.Ptr(int64(1)), Name: github.Ptr("v1.0"), CreatedAt: &github.Timestamp{Time: t1}},
	}
	result := FilterVersions(versions, VersionFilter{Names: []string{"v9.9"}})
	assert.Empty(t, result)
}

func TestFilterVersions_NamesAndSince(t *testing.T) {
	// Combining Names and Since: only versions matching both criteria are returned.
	versions := []*github.PackageVersion{
		{ID: github.Ptr(int64(1)), Name: github.Ptr("v1.0"), CreatedAt: &github.Timestamp{Time: t1}},
		{ID: github.Ptr(int64(2)), Name: github.Ptr("v2.0"), CreatedAt: &github.Timestamp{Time: t2}},
		{ID: github.Ptr(int64(3)), Name: github.Ptr("v3.0"), CreatedAt: &github.Timestamp{Time: t3}},
	}
	result := FilterVersions(versions, VersionFilter{Names: []string{"v1.0", "v3.0"}, Since: &t2})
	// v1.0 is before Since (t2), so only v3.0 passes both filters.
	assert.Len(t, result, 1)
	assert.Equal(t, "v3.0", result[0].GetName())
}

func TestFilterVersions_NamesAndLatest(t *testing.T) {
	// Combining Names and Latest: latest N from the name-filtered set.
	versions := []*github.PackageVersion{
		{ID: github.Ptr(int64(1)), Name: github.Ptr("v1.0"), CreatedAt: &github.Timestamp{Time: t1}},
		{ID: github.Ptr(int64(2)), Name: github.Ptr("v2.0"), CreatedAt: &github.Timestamp{Time: t2}},
		{ID: github.Ptr(int64(3)), Name: github.Ptr("v3.0"), CreatedAt: &github.Timestamp{Time: t3}},
		{ID: github.Ptr(int64(4)), Name: github.Ptr("v4.0"), CreatedAt: &github.Timestamp{Time: t4}},
	}
	result := FilterVersions(versions, VersionFilter{
		Names:  []string{"v1.0", "v3.0", "v4.0"},
		Latest: 2,
	})
	// Latest 2 from {v1.0, v3.0, v4.0}, newest first.
	assert.Len(t, result, 2)
	assert.Equal(t, "v4.0", result[0].GetName())
	assert.Equal(t, "v3.0", result[1].GetName())
}

func TestFilterVersions_EmptyInput(t *testing.T) {
	result := FilterVersions(nil, VersionFilter{Latest: 5, Since: &t1, Until: &t4})
	assert.Empty(t, result)
}

func TestContainerImageBase(t *testing.T) {
	tests := []struct {
		name     string
		repo     repository.Repository
		pkg      string
		expected string
	}{
		{
			name: "github.com with lowercase",
			repo: repository.Repository{
				Host:  "github.com",
				Owner: "myowner",
			},
			pkg:      "mypackage",
			expected: "ghcr.io/myowner/mypackage",
		},
		{
			name: "github.com with empty host",
			repo: repository.Repository{
				Owner: "myowner",
			},
			pkg:      "mypackage",
			expected: "ghcr.io/myowner/mypackage",
		},
		{
			name: "github.com with mixed case owner and package",
			repo: repository.Repository{
				Owner: "MyOwner",
			},
			pkg:      "MyPackage",
			expected: "ghcr.io/myowner/mypackage",
		},
		{
			name: "GHES with different host",
			repo: repository.Repository{
				Host:  "ghe.example.com",
				Owner: "ghesowner",
			},
			pkg:      "ghespackage",
			expected: "containers.ghe.example.com/ghesowner/ghespackage",
		},
		{
			name: "GHES with mixed case",
			repo: repository.Repository{
				Host:  "GHE.Example.Com",
				Owner: "GHESOwner",
			},
			pkg:      "GHESPackage",
			expected: "containers.GHE.Example.Com/ghesowner/ghespackage",
		},
		{
			name: "owner with hyphen",
			repo: repository.Repository{
				Owner: "my-owner",
			},
			pkg:      "my-package",
			expected: "ghcr.io/my-owner/my-package",
		},
		{
			name: "owner with underscore",
			repo: repository.Repository{
				Owner: "my_owner",
			},
			pkg:      "my_package",
			expected: "ghcr.io/my_owner/my_package",
		},
		{
			name: "with repository name set (owner only case)",
			repo: repository.Repository{
				Owner: "myowner",
				Name:  "myrepo",
			},
			pkg:      "mypackage",
			expected: "ghcr.io/myowner/mypackage",
		},
		{
			name: "with repository name set and GHES",
			repo: repository.Repository{
				Host:  "ghe.example.com",
				Owner: "ghesowner",
				Name:  "ghesrepo",
			},
			pkg:      "ghespackage",
			expected: "containers.ghe.example.com/ghesowner/ghespackage",
		},
		{
			name: "with repository name set, different from package",
			repo: repository.Repository{
				Owner: "myowner",
				Name:  "my-repo",
			},
			pkg:      "my-other-package",
			expected: "ghcr.io/myowner/my-other-package",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainerImageBase(tt.repo, tt.pkg)
			assert.Equal(t, tt.expected, got)
		})
	}
}
