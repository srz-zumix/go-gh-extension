package client

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadCopilotMetricsOmitsGitHubAuthorization(t *testing.T) {
	var gotAuth []string
	raw := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotAuth = append(gotAuth, r.Header.Get("Authorization"))
		body := `{"report_start_day":"2030-01-01","report_end_day":"2030-01-28","day_totals":[]}`
		if strings.Contains(r.URL.Path, "users") {
			body = "{\"user_id\":1,\"user_login\":\"octocat\",\"day\":\"2030-01-01\"}\n"
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}, Request: r}, nil
	})
	g := newTestClient(t, "https://api.example.test/", &authInjectingTransport{raw: raw, token: "secret"})

	ctx := context.Background()
	periodic, err := g.DownloadCopilotPeriodicMetrics(ctx, "https://reports.example.test/org.json?sig=fake")
	require.NoError(t, err)
	assert.Equal(t, "2030-01-28", periodic.ReportEndDay)
	users, err := g.DownloadCopilotUserDailyMetrics(ctx, "https://reports.example.test/users.json?sig=fake")
	require.NoError(t, err)
	require.Len(t, users, 1)

	assert.Equal(t, []string{"", ""}, gotAuth, "pre-signed download links must not receive the GitHub token")
}

type fakeTimeoutError struct{}

func (fakeTimeoutError) Error() string   { return "fake timeout" }
func (fakeTimeoutError) Timeout() bool   { return true }
func (fakeTimeoutError) Temporary() bool { return true }

func TestDownloadCopilotMetricsRetriesTransientFailures(t *testing.T) {
	orig := copilotMetricsRetryBackoff
	copilotMetricsRetryBackoff = func(int) time.Duration { return time.Millisecond }
	t.Cleanup(func() { copilotMetricsRetryBackoff = orig })

	tests := []struct {
		name      string
		failures  []func(*http.Request) (*http.Response, error)
		wantErr   bool
		wantCalls int
	}{
		{
			name: "timeout then success",
			failures: []func(*http.Request) (*http.Response, error){
				func(*http.Request) (*http.Response, error) { return nil, fakeTimeoutError{} },
			},
			wantCalls: 2,
		},
		{
			name: "server error then success",
			failures: []func(*http.Request) (*http.Response, error){
				func(r *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}, Request: r}, nil
				},
			},
			wantCalls: 2,
		},
		{
			name: "gives up after max attempts",
			failures: []func(*http.Request) (*http.Response, error){
				func(*http.Request) (*http.Response, error) { return nil, fakeTimeoutError{} },
				func(*http.Request) (*http.Response, error) { return nil, fakeTimeoutError{} },
				func(*http.Request) (*http.Response, error) { return nil, fakeTimeoutError{} },
			},
			wantErr:   true,
			wantCalls: copilotMetricsRequestAttempts,
		},
		{
			name: "does not retry client errors",
			failures: []func(*http.Request) (*http.Response, error){
				func(r *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusForbidden, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}, Request: r}, nil
				},
			},
			wantErr:   true,
			wantCalls: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			raw := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if calls <= len(tt.failures) {
					return tt.failures[calls-1](r)
				}
				body := `{"report_start_day":"2030-01-01","report_end_day":"2030-01-28","day_totals":[]}`
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}, Request: r}, nil
			})
			g := newTestClient(t, "https://api.example.test/", &authInjectingTransport{raw: raw, token: "secret"})

			m, err := g.DownloadCopilotPeriodicMetrics(context.Background(), "https://reports.example.test/org.json?sig=fake")
			assert.Equal(t, tt.wantCalls, calls)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "2030-01-28", m.ReportEndDay)
		})
	}
}
