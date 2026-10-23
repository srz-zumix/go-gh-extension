package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/google/go-github/v93/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubRoundTripper serves responses from a handler without opening a listening socket.
type stubRoundTripper struct {
	handler func(*http.Request) (int, string)
}

func (s stubRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	status, body := s.handler(req)
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Request:    req,
	}, nil
}

func newStubClient(t *testing.T, handler func(*http.Request) (int, string)) *GitHubClient {
	t.Helper()
	return newTestClient(t, "https://api.github.com/", stubRoundTripper{handler: handler})
}

func TestBillingUsageOptionsToUsageReportOptions(t *testing.T) {
	t.Run("nil receiver", func(t *testing.T) {
		var opts *BillingUsageOptions
		got := opts.toUsageReportOptions()
		require.NotNil(t, got)
		assert.Nil(t, got.Year)
		assert.Nil(t, got.Month)
		assert.Nil(t, got.Day)
		assert.Nil(t, got.Hour)
	})

	t.Run("zero values are omitted", func(t *testing.T) {
		got := (&BillingUsageOptions{Year: 2025}).toUsageReportOptions()
		assert.Equal(t, 2025, got.GetYear())
		assert.Nil(t, got.Month)
		assert.Nil(t, got.Day)
		assert.Nil(t, got.Hour)
	})

	t.Run("all values", func(t *testing.T) {
		got := (&BillingUsageOptions{Year: 2025, Month: 3, Day: 4, Hour: github.Ptr(0)}).toUsageReportOptions()
		assert.Equal(t, 2025, got.GetYear())
		assert.Equal(t, 3, got.GetMonth())
		assert.Equal(t, 4, got.GetDay())
		require.NotNil(t, got.Hour)
		assert.Equal(t, 0, got.GetHour())
	})
}

func TestGetUsageReport(t *testing.T) {
	opts := &BillingUsageOptions{Year: 2025, Month: 3, Day: 4, Hour: github.Ptr(5), CostCenterID: "cc-1"}
	for _, tc := range []struct {
		name  string
		path  string
		query map[string]string
		call  func(g *GitHubClient) (any, error)
	}{
		{
			name:  "org",
			path:  "/organizations/example/settings/billing/usage",
			query: map[string]string{"year": "2025", "month": "3", "day": "4", "hour": "5", "cost_center_id": ""},
			call: func(g *GitHubClient) (any, error) {
				return g.GetOrgUsageReport(context.Background(), "example", opts)
			},
		},
		{
			name:  "user",
			path:  "/users/example/settings/billing/usage",
			query: map[string]string{"year": "2025", "month": "3", "day": "4", "hour": "5", "cost_center_id": ""},
			call: func(g *GitHubClient) (any, error) {
				return g.GetUserUsageReport(context.Background(), "example", opts)
			},
		},
		{
			name:  "enterprise",
			path:  "/enterprises/example/settings/billing/usage",
			query: map[string]string{"year": "2025", "month": "3", "day": "4", "hour": "", "cost_center_id": "cc-1"},
			call: func(g *GitHubClient) (any, error) {
				return g.GetEnterpriseUsageReport(context.Background(), "example", opts)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, fail := range []bool{false, true} {
				t.Run(fmt.Sprintf("fail=%v", fail), func(t *testing.T) {
					client := newStubClient(t, func(r *http.Request) (int, string) {
						assert.Equal(t, tc.path, r.URL.Path)
						for k, want := range tc.query {
							assert.Equal(t, want, r.URL.Query().Get(k), "query %q", k)
						}
						if fail {
							return http.StatusForbidden, `{"message":"forbidden"}`
						}
						return http.StatusOK, `{"usageItems":[{"product":"actions","netAmount":1.5}]}`
					})
					got, err := tc.call(client)
					if fail {
						require.Error(t, err)
						assert.IsType(t, &github.ErrorResponse{}, err)
						return
					}
					require.NoError(t, err)
					assert.NotNil(t, got)
				})
			}
		})
	}
}

func TestUsageSummaryFromReport(t *testing.T) {
	t.Run("nil report yields empty summary", func(t *testing.T) {
		got := usageSummaryFromReport(nil)
		require.NotNil(t, got)
		assert.Equal(t, BillingUsageSummary{}, *got)
	})

	t.Run("all fields are copied", func(t *testing.T) {
		month, day := 5, 6
		report := &github.UsageSummaryReport{
			TimePeriod:   github.UsageSummaryTimePeriod{Year: 2025, Month: &month, Day: &day},
			Organization: github.Ptr("example-org"),
			User:         github.Ptr("example-user"),
			Repository:   github.Ptr("example/repo"),
			Product:      github.Ptr("actions"),
			SKU:          github.Ptr("actions_linux"),
			UsageItems: []*github.UsageSummaryItem{{
				Product:          "actions",
				SKU:              "actions_linux",
				UnitType:         "minutes",
				PricePerUnit:     0.008,
				GrossQuantity:    200,
				GrossAmount:      1.6,
				DiscountQuantity: 100,
				DiscountAmount:   0.8,
				NetQuantity:      100,
				NetAmount:        0.8,
			}},
		}
		got := usageSummaryFromReport(report)
		require.NotNil(t, got)
		assert.Equal(t, 2025, got.TimePeriod.Year)
		require.NotNil(t, got.TimePeriod.Month)
		assert.Equal(t, 5, *got.TimePeriod.Month)
		require.NotNil(t, got.TimePeriod.Day)
		assert.Equal(t, 6, *got.TimePeriod.Day)
		require.NotNil(t, got.Organization)
		assert.Equal(t, "example-org", *got.Organization)
		require.NotNil(t, got.User)
		assert.Equal(t, "example-user", *got.User)
		require.NotNil(t, got.Repository)
		assert.Equal(t, "example/repo", *got.Repository)
		require.NotNil(t, got.Product)
		assert.Equal(t, "actions", *got.Product)
		require.NotNil(t, got.SKU)
		assert.Equal(t, "actions_linux", *got.SKU)
		assert.Nil(t, got.Enterprise)
		require.Len(t, got.UsageItems, 1)
		item := got.UsageItems[0]
		assert.Equal(t, "actions", item.Product)
		assert.Equal(t, "actions_linux", item.SKU)
		assert.Equal(t, "minutes", item.UnitType)
		assert.InDelta(t, 0.008, item.PricePerUnit, 1e-9)
		assert.InDelta(t, 200, item.GrossQuantity, 1e-9)
		assert.InDelta(t, 1.6, item.GrossAmount, 1e-9)
		assert.InDelta(t, 100, item.DiscountQuantity, 1e-9)
		assert.InDelta(t, 0.8, item.DiscountAmount, 1e-9)
		assert.InDelta(t, 100, item.NetQuantity, 1e-9)
		assert.InDelta(t, 0.8, item.NetAmount, 1e-9)
	})

	t.Run("nil and empty usage items keep their shape", func(t *testing.T) {
		assert.Nil(t, usageSummaryFromReport(&github.UsageSummaryReport{}).UsageItems)
		got := usageSummaryFromReport(&github.UsageSummaryReport{UsageItems: []*github.UsageSummaryItem{}})
		require.NotNil(t, got.UsageItems)
		assert.Empty(t, got.UsageItems)
	})

	t.Run("nil items are preserved", func(t *testing.T) {
		got := usageSummaryFromReport(&github.UsageSummaryReport{
			UsageItems: []*github.UsageSummaryItem{nil, {SKU: "actions_linux"}},
		})
		require.Len(t, got.UsageItems, 2)
		assert.Nil(t, got.UsageItems[0])
		require.NotNil(t, got.UsageItems[1])
		assert.Equal(t, "actions_linux", got.UsageItems[1].SKU)
	})
}

func TestBillingUsageSummaryOptionsToUsageSummaryOptions(t *testing.T) {
	t.Run("nil receiver", func(t *testing.T) {
		var opts *BillingUsageSummaryOptions
		assert.Nil(t, opts.toUsageSummaryOptions())
	})

	t.Run("enterprise-only fields are dropped and the input is not mutated", func(t *testing.T) {
		in := &BillingUsageSummaryOptions{
			Year:         2025,
			Month:        3,
			Day:          4,
			CostCenterID: "cc-1",
			Organization: "example-org",
			Repository:   "example/repo",
			Product:      "actions",
			SKU:          "actions_linux",
		}
		got := in.toUsageSummaryOptions()
		require.NotNil(t, got)
		assert.Equal(t, github.UsageSummaryOptions{
			Year:       2025,
			Month:      3,
			Day:        4,
			Repository: "example/repo",
			Product:    "actions",
			SKU:        "actions_linux",
		}, *got)
		assert.Equal(t, "cc-1", in.CostCenterID)
		assert.Equal(t, "example-org", in.Organization)
	})
}

func TestGetUsageSummary(t *testing.T) {
	opts := &BillingUsageSummaryOptions{
		Year:         2025,
		Month:        3,
		Day:          4,
		CostCenterID: "cc-1",
		Organization: "example-org",
		Repository:   "example/repo",
		Product:      "actions",
		SKU:          "actions_linux",
	}
	commonQuery := map[string]string{
		"year":       "2025",
		"month":      "3",
		"day":        "4",
		"repository": "example/repo",
		"product":    "actions",
		"sku":        "actions_linux",
	}
	// cost_center_id and organization are accepted by the enterprise summary endpoint only.
	enterpriseOnlyQuery := map[string]string{
		"cost_center_id": "cc-1",
		"organization":   "example-org",
	}
	for _, tc := range []struct {
		name           string
		path           string
		wantEnterprise bool
		call           func(g *GitHubClient, opts *BillingUsageSummaryOptions) (*BillingUsageSummary, error)
	}{
		{
			name: "org",
			path: "/organizations/example/settings/billing/usage/summary",
			call: func(g *GitHubClient, o *BillingUsageSummaryOptions) (*BillingUsageSummary, error) {
				return g.GetOrgUsageSummary(context.Background(), "example", o)
			},
		},
		{
			name: "user",
			path: "/users/example/settings/billing/usage/summary",
			call: func(g *GitHubClient, o *BillingUsageSummaryOptions) (*BillingUsageSummary, error) {
				return g.GetUserUsageSummary(context.Background(), "example", o)
			},
		},
		{
			name:           "enterprise",
			path:           "/enterprises/example/settings/billing/usage/summary",
			wantEnterprise: true,
			call: func(g *GitHubClient, o *BillingUsageSummaryOptions) (*BillingUsageSummary, error) {
				return g.GetEnterpriseUsageSummary(context.Background(), "example", o)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("success", func(t *testing.T) {
				client := newStubClient(t, func(r *http.Request) (int, string) {
					assert.Equal(t, tc.path, r.URL.Path)
					for k, want := range commonQuery {
						assert.Equal(t, want, r.URL.Query().Get(k), "query %q", k)
					}
					for k, want := range enterpriseOnlyQuery {
						if tc.wantEnterprise {
							assert.Equal(t, want, r.URL.Query().Get(k), "query %q", k)
						} else {
							assert.Empty(t, r.URL.Query().Get(k), "query %q", k)
						}
					}
					return http.StatusOK, `{"timePeriod":{"year":2025,"month":3},"organization":"example","usageItems":[{"product":"actions","sku":"actions_linux","unitType":"minutes","pricePerUnit":0.008,"netQuantity":100,"netAmount":0.8}]}`
				})
				summary, err := tc.call(client, opts)
				require.NoError(t, err)
				require.NotNil(t, summary)
				assert.Equal(t, 2025, summary.TimePeriod.Year)
				require.NotNil(t, summary.TimePeriod.Month)
				assert.Equal(t, 3, *summary.TimePeriod.Month)
				assert.Nil(t, summary.TimePeriod.Day)
				require.Len(t, summary.UsageItems, 1)
				assert.Equal(t, "actions_linux", summary.UsageItems[0].SKU)
				assert.InDelta(t, 0.8, summary.UsageItems[0].NetAmount, 1e-9)
			})

			t.Run("nil options sends no query", func(t *testing.T) {
				client := newStubClient(t, func(r *http.Request) (int, string) {
					assert.Empty(t, r.URL.RawQuery)
					return http.StatusOK, `{"timePeriod":{"year":2025},"usageItems":[]}`
				})
				summary, err := tc.call(client, nil)
				require.NoError(t, err)
				require.NotNil(t, summary)
			})

			t.Run("api error", func(t *testing.T) {
				client := newStubClient(t, func(r *http.Request) (int, string) {
					return http.StatusNotFound, `{"message":"not found"}`
				})
				summary, err := tc.call(client, opts)
				require.Error(t, err)
				assert.Nil(t, summary)
				assert.IsType(t, &github.ErrorResponse{}, err)
			})
		})
	}
}
