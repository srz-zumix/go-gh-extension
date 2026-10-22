package gh

import (
	"testing"

	"github.com/google/go-github/v93/github"
)

func sampleBillingRows() []BillingUsageRow {
	return []BillingUsageRow{
		{Date: "2026-01-01", Product: "actions", SKU: "actions_linux", Quantity: 10, UnitType: "minutes", GrossAmount: 0.08, DiscountAmount: 0.08, Organization: "example-org", Repository: "example-org/repo-a"},
		{Date: "2026-01-02", Product: "actions", SKU: "actions_linux", Quantity: 5, UnitType: "minutes", GrossAmount: 0.04, NetAmount: 0.04, Organization: "example-org", Repository: "example-org/repo-b"},
		{Date: "2026-01-02", Product: "actions", SKU: "actions_storage", Quantity: 2, UnitType: "gigabyte-hours", GrossAmount: 0.01, NetAmount: 0.01, Organization: "example-org", Repository: "example-org/repo-b"},
		{Date: "2026-01-02", Product: "Copilot", SKU: "copilot_seat", Quantity: 1, UnitType: "user-months", GrossAmount: 19, NetAmount: 19, Organization: "example-org"},
	}
}

func TestParseBillingGroupBy(t *testing.T) {
	for _, value := range append(BillingGroupByValues(), "") {
		if _, err := ParseBillingGroupBy(value); err != nil {
			t.Errorf("ParseBillingGroupBy(%q) error: %v", value, err)
		}
	}
	if _, err := ParseBillingGroupBy("team"); err == nil {
		t.Error("ParseBillingGroupBy(team) succeeded")
	}
}

func TestFilterBillingUsageRows(t *testing.T) {
	for _, test := range []struct {
		name          string
		product, repo string
		want          int
	}{
		{"none", "", "", 4},
		{"product case insensitive", "ACTIONS", "", 3},
		{"repository full name", "", "example-org/repo-b", 2},
		{"repository and product", "copilot", "example-org/repo-b", 0},
		{"unknown", "packages", "", 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := len(FilterBillingUsageRows(sampleBillingRows(), test.product, test.repo)); got != test.want {
				t.Errorf("got %d rows, want %d", got, test.want)
			}
		})
	}
}

func TestFilterBillingUsageRowsBareRepositoryName(t *testing.T) {
	rows := []BillingUsageRow{{Repository: "repo-a"}, {Repository: "repo-b"}}
	if got := FilterBillingUsageRows(rows, "", "example-org/repo-a"); len(got) != 1 || got[0].Repository != "repo-a" {
		t.Errorf("got %+v", got)
	}
}

func TestGroupBillingUsageRows(t *testing.T) {
	groups := GroupBillingUsageRows(sampleBillingRows(), BillingGroupBySKU)
	if len(groups) != 3 {
		t.Fatalf("got %d groups, want 3", len(groups))
	}
	linux := groups[0]
	if linux.Key != "actions_linux" || linux.Lines != 2 || linux.Quantity == nil || *linux.Quantity != 15 || linux.UnitType != "minutes" {
		t.Errorf("unexpected group %+v", linux)
	}
	if linux.GrossAmount != 0.12 || linux.NetAmount != 0.04 {
		t.Errorf("unexpected amounts %+v", linux)
	}
}

func TestGroupBillingUsageRowsMixedUnit(t *testing.T) {
	groups := GroupBillingUsageRows(sampleBillingRows(), BillingGroupByDate)
	if len(groups) != 2 || groups[1].Key != "2026-01-02" || groups[1].UnitType != BillingMixedUnit || groups[1].Quantity != nil {
		t.Errorf("unexpected groups %+v", groups)
	}
}

func TestBillingUsageRowsFromReport(t *testing.T) {
	report := &github.UsageReport{UsageItems: []*github.UsageItem{
		nil,
		{Date: "2026-01-01", Product: "actions", SKU: "actions_linux", Quantity: 3, UnitType: "minutes", OrganizationName: github.Ptr("example-org"), RepositoryName: github.Ptr("example-org/repo-a")},
	}}
	rows := BillingUsageRowsFromReport(report)
	if len(rows) != 1 || rows[0].Organization != "example-org" || rows[0].Repository != "example-org/repo-a" {
		t.Errorf("unexpected rows %+v", rows)
	}
	if BillingUsageRowsFromReport(nil) != nil {
		t.Error("nil report must yield no rows")
	}
}

func TestBillingUsageRowsFromEnterpriseReport(t *testing.T) {
	report := &github.EnterpriseUsageReport{UsageItems: []*github.EnterpriseUsageItem{
		{Date: "2026-01-01", Product: "actions", SKU: "actions_linux", OrganizationName: "example-org"},
	}}
	rows := BillingUsageRowsFromEnterpriseReport(report)
	if len(rows) != 1 || rows[0].Organization != "example-org" || rows[0].Repository != "" {
		t.Errorf("unexpected rows %+v", rows)
	}
}

func TestFilterBillingUsageRowsBareRepositoryNameOtherOrganization(t *testing.T) {
	rows := []BillingUsageRow{{Organization: "org-a", Repository: "shared"}, {Organization: "org-b", Repository: "shared"}, {Repository: "shared"}}
	got := FilterBillingUsageRows(rows, "", "org-a/shared")
	if len(got) != 2 || got[0].Organization != "org-a" || got[1].Organization != "" {
		t.Fatalf("unexpected rows: %+v", got)
	}
}

func TestGroupBillingUsageRowsBareRepositoryNameOtherOrganization(t *testing.T) {
	rows := []BillingUsageRow{{Organization: "org-a", Repository: "shared"}, {Organization: "org-b", Repository: "shared"}, {Organization: "org-a", Repository: "org-a/shared"}}
	groups := GroupBillingUsageRows(rows, BillingGroupByRepository)
	if len(groups) != 2 || groups[0].Key != "org-a/shared" || groups[0].Lines != 2 || groups[1].Key != "org-b/shared" {
		t.Fatalf("unexpected groups: %+v", groups)
	}
}
