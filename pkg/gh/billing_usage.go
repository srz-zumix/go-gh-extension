package gh

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-github/v90/github"
)

// BillingUsageRow is a billing usage line normalized across organization, user and enterprise reports.
type BillingUsageRow struct {
	Date           string  `json:"date"`
	Product        string  `json:"product"`
	SKU            string  `json:"sku"`
	Quantity       float64 `json:"quantity"`
	UnitType       string  `json:"unitType"`
	PricePerUnit   float64 `json:"pricePerUnit"`
	GrossAmount    float64 `json:"grossAmount"`
	DiscountAmount float64 `json:"discountAmount"`
	NetAmount      float64 `json:"netAmount"`
	Organization   string  `json:"organization,omitempty"`
	Repository     string  `json:"repository,omitempty"`
}

// BillingUsageGroup is the sum of the billing usage lines that share a group key.
type BillingUsageGroup struct {
	Key   string `json:"key"`
	Lines int    `json:"lines"`
	// Quantity is nil when the lines are measured in different units, as their quantities cannot be summed.
	Quantity       *float64 `json:"quantity,omitempty"`
	UnitType       string   `json:"unitType"`
	GrossAmount    float64  `json:"grossAmount"`
	DiscountAmount float64  `json:"discountAmount"`
	NetAmount      float64  `json:"netAmount"`
}

// BillingMixedUnit is the UnitType of a group whose lines are measured in different units.
const BillingMixedUnit = "mixed"

// BillingGroupBy is the key billing usage lines are aggregated by.
type BillingGroupBy string

const (
	BillingGroupByDate       BillingGroupBy = "date"
	BillingGroupByOrg        BillingGroupBy = "org"
	BillingGroupByProduct    BillingGroupBy = "product"
	BillingGroupByRepository BillingGroupBy = "repo"
	BillingGroupBySKU        BillingGroupBy = "sku"
)

// BillingGroupByValues lists the accepted group-by values.
func BillingGroupByValues() []string {
	return []string{
		string(BillingGroupByDate), string(BillingGroupByOrg), string(BillingGroupByProduct),
		string(BillingGroupByRepository), string(BillingGroupBySKU),
	}
}

// ParseBillingGroupBy validates a group-by value; an empty value means no aggregation.
func ParseBillingGroupBy(value string) (BillingGroupBy, error) {
	if value == "" {
		return "", nil
	}
	for _, v := range BillingGroupByValues() {
		if value == v {
			return BillingGroupBy(v), nil
		}
	}
	return "", fmt.Errorf("invalid group-by %q: must be one of %s", value, strings.Join(BillingGroupByValues(), ", "))
}

// BillingUsageRowsFromReport converts an organization or user usage report.
func BillingUsageRowsFromReport(report *github.UsageReport) []BillingUsageRow {
	if report == nil {
		return nil
	}
	rows := make([]BillingUsageRow, 0, len(report.UsageItems))
	for _, item := range report.UsageItems {
		if item == nil {
			continue
		}
		rows = append(rows, BillingUsageRow{
			Date:           item.Date,
			Product:        item.Product,
			SKU:            item.SKU,
			Quantity:       item.Quantity,
			UnitType:       item.UnitType,
			PricePerUnit:   item.PricePerUnit,
			GrossAmount:    item.GrossAmount,
			DiscountAmount: item.DiscountAmount,
			NetAmount:      item.NetAmount,
			Organization:   item.GetOrganizationName(),
			Repository:     item.GetRepositoryName(),
		})
	}
	return rows
}

// BillingUsageRowsFromEnterpriseReport converts an enterprise usage report.
func BillingUsageRowsFromEnterpriseReport(report *github.EnterpriseUsageReport) []BillingUsageRow {
	if report == nil {
		return nil
	}
	rows := make([]BillingUsageRow, 0, len(report.UsageItems))
	for _, item := range report.UsageItems {
		if item == nil {
			continue
		}
		rows = append(rows, BillingUsageRow{
			Date:           item.Date,
			Product:        item.Product,
			SKU:            item.SKU,
			Quantity:       item.Quantity,
			UnitType:       item.UnitType,
			PricePerUnit:   item.PricePerUnit,
			GrossAmount:    item.GrossAmount,
			DiscountAmount: item.DiscountAmount,
			NetAmount:      item.NetAmount,
			Organization:   item.OrganizationName,
			Repository:     item.GetRepositoryName(),
		})
	}
	return rows
}

// FilterBillingUsageRows keeps the lines of product (all products when empty) and of repository,
// given as OWNER/REPO (all repositories when empty). The API reports the repository
// either as OWNER/REPO or as the bare name, so both forms match; a bare name matches only
// when the line's organization is OWNER or unknown.
func FilterBillingUsageRows(rows []BillingUsageRow, product, repository string) []BillingUsageRow {
	if product == "" && repository == "" {
		return rows
	}
	owner, name, _ := strings.Cut(repository, "/")
	filtered := make([]BillingUsageRow, 0, len(rows))
	for _, row := range rows {
		if product != "" && !strings.EqualFold(row.Product, product) {
			continue
		}
		if repository != "" && !strings.EqualFold(row.Repository, repository) &&
			(!strings.EqualFold(row.Repository, name) || (row.Organization != "" && !strings.EqualFold(row.Organization, owner))) {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func (by BillingGroupBy) key(row BillingUsageRow) string {
	switch by {
	case BillingGroupByDate:
		return row.Date
	case BillingGroupByOrg:
		return row.Organization
	case BillingGroupByProduct:
		return row.Product
	case BillingGroupByRepository:
		// Qualify a bare repository name so that same-named repositories of different organizations stay apart.
		if row.Organization != "" && row.Repository != "" && !strings.Contains(row.Repository, "/") {
			return row.Organization + "/" + row.Repository
		}
		return row.Repository
	default:
		return row.SKU
	}
}

// GroupBillingUsageRows sums the lines by key, ordered by key. Lines without a key are collected under an empty key.
func GroupBillingUsageRows(rows []BillingUsageRow, by BillingGroupBy) []BillingUsageGroup {
	index := map[string]*BillingUsageGroup{}
	var order []string
	for _, row := range rows {
		key := by.key(row)
		group, ok := index[key]
		if !ok {
			group = &BillingUsageGroup{Key: key, UnitType: row.UnitType, Quantity: new(float64)}
			index[key] = group
			order = append(order, key)
		} else if group.UnitType != row.UnitType {
			group.UnitType = BillingMixedUnit
			group.Quantity = nil
		}
		group.Lines++
		if group.Quantity != nil {
			*group.Quantity += row.Quantity
		}
		group.GrossAmount += row.GrossAmount
		group.DiscountAmount += row.DiscountAmount
		group.NetAmount += row.NetAmount
	}
	sort.Strings(order)
	groups := make([]BillingUsageGroup, 0, len(order))
	for _, key := range order {
		groups = append(groups, *index[key])
	}
	return groups
}
