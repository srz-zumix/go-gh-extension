package gh

import (
	"testing"

	"github.com/google/go-github/v79/github"
)

func TestSelectSBOMPackages(t *testing.T) {
	deps := []*github.RepoDependencies{
		{Name: github.Ptr("npm:react")},
		{Name: github.Ptr("actions:actions/checkout")},
		{Name: github.Ptr("gomod:github.com/pkg/errors")},
		{Name: github.Ptr("actions:actions/cache")},
		{Name: github.Ptr("npm:lodash")},
		{Name: nil},
	}

	selected := SelectSBOMPackages(deps, []string{"actions", "npm"})

	if len(selected) != 4 {
		t.Fatalf("expected 4 packages, got %d", len(selected))
	}

	gotNames := make([]string, 0, len(selected))
	for _, dep := range selected {
		if dep.Name == nil {
			t.Fatal("selected package name should not be nil")
		}
		gotNames = append(gotNames, dep.GetName())
	}

	wantNames := []string{
		"actions:actions/cache",
		"actions:actions/checkout",
		"npm:lodash",
		"npm:react",
	}
	for i := range wantNames {
		if gotNames[i] != wantNames[i] {
			t.Fatalf("unexpected order/content at index %d: got %q, want %q", i, gotNames[i], wantNames[i])
		}
	}
}

func TestFilterSBOMPackages(t *testing.T) {
	sbom := &github.SBOM{
		SBOM: &github.SBOMInfo{
			SPDXID:            github.Ptr("SPDXRef-DOCUMENT"),
			SPDXVersion:       github.Ptr("SPDX-2.3"),
			DataLicense:       github.Ptr("CC0-1.0"),
			DocumentNamespace: github.Ptr("https://example.com/spdx"),
			Name:              github.Ptr("test-sbom"),
			Packages: []*github.RepoDependencies{
				{Name: github.Ptr("npm:react")},
				{Name: github.Ptr("actions:actions/checkout")},
				{Name: nil},
				{Name: github.Ptr("npm:lodash")},
			},
		},
	}

	filtered := FilterSBOMPackages(sbom, []string{"npm"})

	if filtered == sbom {
		t.Fatal("expected a new SBOM instance")
	}
	if filtered.SBOM == sbom.SBOM {
		t.Fatal("expected a new SBOMInfo instance")
	}

	if filtered.SBOM.GetSPDXID() != sbom.SBOM.GetSPDXID() {
		t.Fatalf("SPDXID should be preserved: got %q, want %q", filtered.SBOM.GetSPDXID(), sbom.SBOM.GetSPDXID())
	}
	if filtered.SBOM.GetDocumentNamespace() != sbom.SBOM.GetDocumentNamespace() {
		t.Fatalf("DocumentNamespace should be preserved: got %q, want %q", filtered.SBOM.GetDocumentNamespace(), sbom.SBOM.GetDocumentNamespace())
	}

	if len(filtered.SBOM.Packages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(filtered.SBOM.Packages))
	}
	if filtered.SBOM.Packages[0].GetName() != "npm:lodash" {
		t.Fatalf("unexpected first package: got %q", filtered.SBOM.Packages[0].GetName())
	}
	if filtered.SBOM.Packages[1].GetName() != "npm:react" {
		t.Fatalf("unexpected second package: got %q", filtered.SBOM.Packages[1].GetName())
	}
}

func TestExcludeSBOMPackages(t *testing.T) {
	sbom := &github.SBOM{
		SBOM: &github.SBOMInfo{
			SPDXID:      github.Ptr("SPDXRef-DOCUMENT"),
			SPDXVersion: github.Ptr("SPDX-2.3"),
			Packages: []*github.RepoDependencies{
				{Name: github.Ptr("npm:react")},
				{Name: github.Ptr("actions:actions/checkout")},
				{Name: nil},
				{Name: github.Ptr("gomod:github.com/pkg/errors")},
				{Name: github.Ptr("plainpackage")},
			},
		},
	}

	t.Run("exclude target ecosystems and keep nil names", func(t *testing.T) {
		filtered := ExcludeSBOMPackages(sbom, []string{"actions", "gomod"})

		if len(filtered.SBOM.Packages) != 3 {
			t.Fatalf("expected 3 packages, got %d", len(filtered.SBOM.Packages))
		}
		if filtered.SBOM.Packages[0].GetName() != "npm:react" {
			t.Fatalf("unexpected first package: got %q", filtered.SBOM.Packages[0].GetName())
		}
		if filtered.SBOM.Packages[1].Name != nil {
			t.Fatalf("expected second package name to be nil, got %q", filtered.SBOM.Packages[1].GetName())
		}
		if filtered.SBOM.Packages[2].GetName() != "plainpackage" {
			t.Fatalf("unexpected third package: got %q", filtered.SBOM.Packages[2].GetName())
		}
	})

	t.Run("no ecosystems returns original pointer", func(t *testing.T) {
		filtered := ExcludeSBOMPackages(sbom, nil)
		if filtered != sbom {
			t.Fatal("expected original SBOM pointer when ecosystems is empty")
		}
	})
}
