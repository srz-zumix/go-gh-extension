package gh

import (
	"net/http"
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCopilotLatestMetricsMergesLinks(t *testing.T) {
	g, _ := newFixtureClient(t, map[string]fixtureResponse{
		"GET " + fixtureAPI + "orgs/octo/copilot/metrics/reports/organization-28-day/latest": ok(`{"download_links":["https://dl.example.test/a","https://dl.example.test/b"],"report_start_day":"2030-01-01","report_end_day":"2030-01-28"}`),
		"GET https://dl.example.test/a": ok(`{"report_start_day":"2030-01-01","report_end_day":"2030-01-28","day_totals":[{"day":"2030-01-01","daily_active_users":3}]}`),
		"GET https://dl.example.test/b": ok(`{"report_start_day":"2030-01-01","report_end_day":"2030-01-28","day_totals":[{"day":"2030-01-02","daily_active_users":4}]}`),
	})
	m, err := GetCopilotLatestMetrics(t.Context(), g, CopilotMetricsSource{Owner: repository.Repository{Owner: "octo"}})
	require.NoError(t, err)
	assert.Equal(t, CopilotMetricsWindow{StartDay: "2030-01-01", EndDay: "2030-01-28"}, m.Window)
	require.Len(t, m.DayTotals, 2)
	assert.Equal(t, "2030-01-02", m.DayTotals[1].Day)
}

func TestGetCopilotDailyMetricsEnterpriseAndMissing(t *testing.T) {
	g, _ := newFixtureClient(t, map[string]fixtureResponse{
		"GET " + fixtureAPI + "enterprises/acme/copilot/metrics/reports/enterprise-1-day?day=2030-01-01": ok(`{"download_links":["https://dl.example.test/e1"],"report_day":"2030-01-01"}`),
		"GET https://dl.example.test/e1": ok(`{"day":"2030-01-01","daily_active_users":9}`),
		"GET " + fixtureAPI + "enterprises/acme/copilot/metrics/reports/enterprise-1-day?day=2030-01-03": ok(`{"download_links":[],"report_day":"2030-01-03"}`),
	})
	src := CopilotMetricsSource{Owner: repository.Repository{Owner: "acme"}, Enterprise: true}
	assert.Equal(t, "acme (enterprise)", src.String())

	records, found, err := GetCopilotDailyMetrics(t.Context(), g, src, "2030-01-01")
	require.NoError(t, err)
	assert.True(t, found)
	require.Len(t, records, 1)
	assert.Equal(t, 9, *records[0].DailyActiveUsers)

	_, found, err = GetCopilotDailyMetrics(t.Context(), g, src, "2030-01-02")
	require.NoError(t, err, "404 is reported as a missing day")
	assert.False(t, found)

	_, found, err = GetCopilotDailyMetrics(t.Context(), g, src, "2030-01-03")
	require.NoError(t, err)
	assert.False(t, found, "empty download links are reported as a missing day")
}

func TestGetCopilotDailyUserMetricsNDJSON(t *testing.T) {
	g, _ := newFixtureClient(t, map[string]fixtureResponse{
		"GET " + fixtureAPI + "orgs/octo/copilot/metrics/reports/users-1-day?day=2030-01-01": ok(`{"download_links":["https://dl.example.test/u1","https://dl.example.test/u2"],"report_day":"2030-01-01"}`),
		"GET https://dl.example.test/u1": ok("{\"user_id\":1,\"user_login\":\"user-a\",\"day\":\"2030-01-01\"}\n{\"user_id\":2,\"user_login\":\"user-b\",\"day\":\"2030-01-01\"}\n"),
		"GET https://dl.example.test/u2": ok("{\"user_id\":3,\"user_login\":\"user-c\",\"day\":\"2030-01-01\"}\n"),
	})
	records, found, err := GetCopilotDailyUserMetrics(t.Context(), g, CopilotMetricsSource{Owner: repository.Repository{Owner: "octo"}}, "2030-01-01")
	require.NoError(t, err)
	assert.True(t, found)
	require.Len(t, records, 3)
	assert.Equal(t, "user-c", records[2].UserLogin)
}

func TestGetCopilotDailyMetricsServerError(t *testing.T) {
	g, _ := newFixtureClient(t, map[string]fixtureResponse{
		"GET " + fixtureAPI + "orgs/octo/copilot/metrics/reports/repos-1-day?day=2030-01-01": {Status: http.StatusForbidden, Body: `{"message":"Forbidden"}`},
	})
	_, _, err := GetCopilotDailyRepositoryMetrics(t.Context(), g, CopilotMetricsSource{Owner: repository.Repository{Owner: "octo"}}, "2030-01-01")
	require.Error(t, err)
}
