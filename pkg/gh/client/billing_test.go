package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/google/go-github/v90/github"
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
	wantQuery := map[string]string{
		"year":           "2025",
		"month":          "3",
		"day":            "4",
		"cost_center_id": "cc-1",
		"organization":   "example-org",
		"repository":     "example/repo",
		"product":        "actions",
		"sku":            "actions_linux",
	}
	for _, tc := range []struct {
		name string
		path string
		call func(g *GitHubClient, opts *BillingUsageSummaryOptions) (*BillingUsageSummary, error)
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
			name: "enterprise",
			path: "/enterprises/example/settings/billing/usage/summary",
			call: func(g *GitHubClient, o *BillingUsageSummaryOptions) (*BillingUsageSummary, error) {
				return g.GetEnterpriseUsageSummary(context.Background(), "example", o)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("success", func(t *testing.T) {
				client := newStubClient(t, func(r *http.Request) (int, string) {
					assert.Equal(t, tc.path, r.URL.Path)
					for k, want := range wantQuery {
						assert.Equal(t, want, r.URL.Query().Get(k), "query %q", k)
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
