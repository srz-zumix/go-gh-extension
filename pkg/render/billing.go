package render

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

func formatBillingNumber(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func formatBillingAmount(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func (r *Renderer) writeBillingTotal(gross, discount, net float64) {
	r.WriteLine(fmt.Sprintf("Total (USD): gross %s, discount %s, net %s",
		formatBillingAmount(gross), formatBillingAmount(discount), formatBillingAmount(net)))
}

// RenderBillingUsage prints the usage lines, or their sums by groupBy when it is set.
func (r *Renderer) RenderBillingUsage(rows []gh.BillingUsageRow, groupBy gh.BillingGroupBy) error {
	if groupBy != "" {
		groups := gh.GroupBillingUsageRows(rows, groupBy)
		if r.HasExporter() {
			return r.RenderExportedData(groups)
		}
		return r.renderBillingUsageGroups(groups, groupBy)
	}
	if r.HasExporter() {
		return r.RenderExportedData(rows)
	}
	if len(rows) == 0 {
		r.WriteLine("No billing usage.")
		return nil
	}
	table := r.NewTableWriter([]string{"DATE", "PRODUCT", "SKU", "QUANTITY", "UNIT", "PRICE", "GROSS", "DISCOUNT", "NET", "ORGANIZATION", "REPOSITORY"})
	var gross, discount, net float64
	for _, row := range rows {
		table.Append([]string{
			row.Date, row.Product, row.SKU,
			formatBillingNumber(row.Quantity), row.UnitType, formatBillingNumber(row.PricePerUnit),
			formatBillingAmount(row.GrossAmount), formatBillingAmount(row.DiscountAmount), formatBillingAmount(row.NetAmount),
			row.Organization, row.Repository,
		})
		gross += row.GrossAmount
		discount += row.DiscountAmount
		net += row.NetAmount
	}
	if err := table.Render(); err != nil {
		return err
	}
	r.writeBillingTotal(gross, discount, net)
	return nil
}

func (r *Renderer) renderBillingUsageGroups(groups []gh.BillingUsageGroup, groupBy gh.BillingGroupBy) error {
	if len(groups) == 0 {
		r.WriteLine("No billing usage.")
		return nil
	}
	table := r.NewTableWriter([]string{strings.ToUpper(string(groupBy)), "LINES", "QUANTITY", "UNIT", "GROSS", "DISCOUNT", "NET"})
	var gross, discount, net float64
	for _, group := range groups {
		quantity := ""
		if group.Quantity != nil {
			quantity = formatBillingNumber(*group.Quantity)
		}
		table.Append([]string{
			group.Key, strconv.Itoa(group.Lines), quantity, group.UnitType,
			formatBillingAmount(group.GrossAmount), formatBillingAmount(group.DiscountAmount), formatBillingAmount(group.NetAmount),
		})
		gross += group.GrossAmount
		discount += group.DiscountAmount
		net += group.NetAmount
	}
	if err := table.Render(); err != nil {
		return err
	}
	r.writeBillingTotal(gross, discount, net)
	return nil
}

// RenderBillingSummary prints the lines of a usage summary.
func (r *Renderer) RenderBillingSummary(summary *gh.BillingUsageSummary) error {
	if r.HasExporter() {
		return r.RenderExportedData(summary)
	}
	if summary == nil || !slices.ContainsFunc(summary.UsageItems, func(item *gh.BillingUsageSummaryItem) bool { return item != nil }) {
		r.WriteLine("No billing usage.")
		return nil
	}
	table := r.NewTableWriter([]string{"PRODUCT", "SKU", "UNIT", "PRICE", "GROSS QTY", "GROSS", "DISCOUNT QTY", "DISCOUNT", "NET QTY", "NET"})
	var gross, discount, net float64
	for _, item := range summary.UsageItems {
		if item == nil {
			continue
		}
		table.Append([]string{
			item.Product, item.SKU, item.UnitType, formatBillingNumber(item.PricePerUnit),
			formatBillingNumber(item.GrossQuantity), formatBillingAmount(item.GrossAmount),
			formatBillingNumber(item.DiscountQuantity), formatBillingAmount(item.DiscountAmount),
			formatBillingNumber(item.NetQuantity), formatBillingAmount(item.NetAmount),
		})
		gross += item.GrossAmount
		discount += item.DiscountAmount
		net += item.NetAmount
	}
	if err := table.Render(); err != nil {
		return err
	}
	r.writeBillingTotal(gross, discount, net)
	return nil
}
