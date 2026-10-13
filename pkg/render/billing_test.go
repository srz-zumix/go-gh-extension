package render

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

func billingTestRows() []gh.BillingUsageRow {
	return []gh.BillingUsageRow{
		{Date: "2026-01-01", Product: "actions", SKU: "actions_linux", Quantity: 10, UnitType: "minutes", PricePerUnit: 0.008, GrossAmount: 0.08, DiscountAmount: 0.08, Organization: "example-org", Repository: "example-org/repo-a"},
		{Date: "2026-01-02", Product: "actions", SKU: "actions_linux", Quantity: 5, UnitType: "minutes", PricePerUnit: 0.008, GrossAmount: 0.04, NetAmount: 0.04, Organization: "example-org", Repository: "example-org/repo-b"},
	}
}

func assertBillingOutput(t *testing.T, output string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(output, want) {
			t.Errorf("output %q does not contain %q", output, want)
		}
	}
}

func TestRenderBillingUsage(t *testing.T) {
	r := NewStringRenderer(nil)
	if err := r.Renderer.RenderBillingUsage(billingTestRows(), ""); err != nil {
		t.Fatal(err)
	}
	assertBillingOutput(t, r.Stdout.String(), "DATE", "REPOSITORY", "actions_linux", "0.008", "example-org/repo-b", "Total (USD): gross 0.12, discount 0.08, net 0.04")
}

func TestRenderBillingUsageGrouped(t *testing.T) {
	r := NewStringRenderer(nil)
	if err := r.Renderer.RenderBillingUsage(billingTestRows(), gh.BillingGroupBySKU); err != nil {
		t.Fatal(err)
	}
	assertBillingOutput(t, r.Stdout.String(), "SKU", "LINES", "actions_linux", "15", "minutes")
}

func TestRenderBillingUsageEmpty(t *testing.T) {
	r := NewStringRenderer(nil)
	if err := r.Renderer.RenderBillingUsage(nil, ""); err != nil {
		t.Fatal(err)
	}
	assertBillingOutput(t, r.Stdout.String(), "No billing usage.")
}

func TestRenderBillingUsageJSON(t *testing.T) {
	for _, groupBy := range []gh.BillingGroupBy{"", gh.BillingGroupBySKU} {
		r := NewStringRenderer(cmdutil.NewJSONExporter())
		if err := r.Renderer.RenderBillingUsage(billingTestRows(), groupBy); err != nil {
			t.Fatal(err)
		}
		var decoded []map[string]any
		if err := json.Unmarshal(r.Stdout.Bytes(), &decoded); err != nil {
			t.Fatalf("invalid JSON %q: %v", r.Stdout.String(), err)
		}
		if len(decoded) == 0 {
			t.Errorf("groupBy %q: no JSON entries", groupBy)
		}
	}
}

func TestRenderBillingSummary(t *testing.T) {
	summary := &gh.BillingUsageSummary{UsageItems: []*gh.BillingUsageSummaryItem{
		{Product: "actions", SKU: "actions_linux", UnitType: "minutes", PricePerUnit: 0.008, GrossQuantity: 100, GrossAmount: 0.8, DiscountQuantity: 100, DiscountAmount: 0.8},
	}}
	r := NewStringRenderer(nil)
	if err := r.Renderer.RenderBillingSummary(summary); err != nil {
		t.Fatal(err)
	}
	assertBillingOutput(t, r.Stdout.String(), "GROSS QTY", "actions_linux", "100", "Total (USD): gross 0.80, discount 0.80, net 0.00")

	empty := NewStringRenderer(nil)
	if err := empty.Renderer.RenderBillingSummary(&gh.BillingUsageSummary{}); err != nil {
		t.Fatal(err)
	}
	assertBillingOutput(t, empty.Stdout.String(), "No billing usage.")
}
