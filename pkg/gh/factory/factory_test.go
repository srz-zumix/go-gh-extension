package factory

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetOnlyRoundTripper(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		wantStatusCode int
		wantErr        bool
	}{
		{
			name:           "GET method allowed",
			method:         http.MethodGet,
			wantStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "HEAD method allowed",
			method:         http.MethodHead,
			wantStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:    "POST method blocked",
			method:  http.MethodPost,
			wantErr: true,
		},
		{
			name:    "PUT method blocked",
			method:  http.MethodPut,
			wantErr: true,
		},
		{
			name:    "PATCH method blocked",
			method:  http.MethodPatch,
			wantErr: true,
		},
		{
			name:    "DELETE method blocked",
			method:  http.MethodDelete,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server that returns 200 OK for all requests
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			// Create getOnlyRoundTripper with default transport
			rt := &getOnlyRoundTripper{
				transport: http.DefaultTransport,
			}

			// Create HTTP request
			req, err := http.NewRequest(tt.method, server.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Execute RoundTrip
			resp, err := rt.RoundTrip(req)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("RoundTrip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For allowed methods, check status code
			if !tt.wantErr {
				if resp == nil {
					t.Errorf("RoundTrip() returned nil response for allowed method")
					return
				}
				if resp.StatusCode != tt.wantStatusCode {
					t.Errorf("RoundTrip() status code = %v, want %v", resp.StatusCode, tt.wantStatusCode)
				}
			}

			// For blocked methods, check error message and response is nil
			if tt.wantErr {
				if resp != nil {
					t.Errorf("RoundTrip() returned non-nil response for blocked method")
				}
				if err != nil {
					expectedErrMsg := "only GET and HEAD methods are allowed, got " + tt.method
					if err.Error() != expectedErrMsg {
						t.Errorf("RoundTrip() error message = %v, want %v", err.Error(), expectedErrMsg)
					}
				}
			}

			// Close response body if not nil
			if resp != nil && resp.Body != nil {
				err = resp.Body.Close()
				if err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}
		})
	}
}

func TestReadOnlyClient(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK")) // nolint
	}))
	defer server.Close()

	tests := []struct {
		name           string
		readOnly       bool
		method         string
		wantStatusCode int
		wantErr        bool
	}{
		{
			name:           "ReadOnly: false, GET allowed",
			readOnly:       false,
			method:         http.MethodGet,
			wantStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "ReadOnly: false, POST allowed",
			readOnly:       false,
			method:         http.MethodPost,
			wantStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "ReadOnly: true, GET allowed",
			readOnly:       true,
			method:         http.MethodGet,
			wantStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "ReadOnly: true, POST blocked",
			readOnly:       true,
			method:         http.MethodPost,
			wantStatusCode: http.StatusMethodNotAllowed,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config
			c := &Config{
				Token:               "",
				DialTimeout:         5 * time.Second,
				TLSHandshakeTimeout: 5 * time.Second,
				Timeout:             30 * time.Second,
				ReadOnly:            tt.readOnly,
			}

			// Create HTTP client
			client := httpClient(c)

			// Create HTTP request
			req, err := http.NewRequest(tt.method, server.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Execute request
			resp, err := client.Do(req)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("client.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check status code only if response is not nil
			if !tt.wantErr && resp == nil {
				t.Errorf("client.Do() returned nil response")
				return
			}
			if resp != nil && resp.StatusCode != tt.wantStatusCode {
				t.Errorf("client.Do() status code = %v, want %v", resp.StatusCode, tt.wantStatusCode)
			}

			// Close response body if not nil
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
		})
	}
}
