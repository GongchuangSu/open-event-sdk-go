package kso

import (
	"net/http"
	"testing"
)

func TestSign(t *testing.T) {
	tests := []struct {
		name    string
		params  *SignParams
		wantErr bool
	}{
		{
			name: "basic sign without body",
			params: &SignParams{
				AppId:      "test_app_id",
				AppSecret:  "test_app_secret",
				Method:     http.MethodGet,
				RequestURI: "/v1/test",
				Date:       "Mon, 02 Jan 2006 15:04:05 GMT",
			},
			wantErr: false,
		},
		{
			name: "sign with body",
			params: &SignParams{
				AppId:       "test_app_id",
				AppSecret:   "test_app_secret",
				Method:      http.MethodPost,
				RequestURI:  "/v1/test",
				ContentType: "application/json",
				Body:        []byte(`{"key":"value"}`),
				Date:        "Mon, 02 Jan 2006 15:04:05 GMT",
			},
			wantErr: false,
		},
		{
			name: "sign with custom date",
			params: &SignParams{
				AppId:      "test_app_id",
				AppSecret:  "test_app_secret",
				Method:     http.MethodGet,
				RequestURI: "/v1/test",
				Date:       "Tue, 03 Jan 2006 10:00:00 GMT",
			},
			wantErr: false,
		},
		{
			name: "missing app_id",
			params: &SignParams{
				AppSecret:  "test_app_secret",
				Method:     http.MethodGet,
				RequestURI: "/v1/test",
			},
			wantErr: true,
		},
		{
			name: "missing app_secret",
			params: &SignParams{
				AppId:      "test_app_id",
				Method:     http.MethodGet,
				RequestURI: "/v1/test",
			},
			wantErr: true,
		},
		{
			name: "missing method",
			params: &SignParams{
				AppId:      "test_app_id",
				AppSecret:  "test_app_secret",
				RequestURI: "/v1/test",
			},
			wantErr: true,
		},
		{
			name: "missing request_uri",
			params: &SignParams{
				AppId:     "test_app_id",
				AppSecret: "test_app_secret",
				Method:    http.MethodGet,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Sign(tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("Sign() result is nil")
					return
				}
				if result.Date == "" {
					t.Error("Sign() result.Date is empty")
				}
				if result.Authorization == "" {
					t.Error("Sign() result.Authorization is empty")
				}
			}
		})
	}
}

func TestSignConsistency(t *testing.T) {
	params := &SignParams{
		AppId:      "test_app_id",
		AppSecret:  "test_app_secret",
		Method:     http.MethodGet,
		RequestURI: "/v1/test",
		Date:       "Mon, 02 Jan 2006 15:04:05 GMT",
	}

	result1, err := Sign(params)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	result2, err := Sign(params)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if result1.Authorization != result2.Authorization {
		t.Errorf("Sign() not consistent: %s != %s", result1.Authorization, result2.Authorization)
	}
}

func TestNewHeaders(t *testing.T) {
	headers, err := NewHeaders("app_id", "app_secret", http.MethodGet, "/v1/test", "", nil)
	if err != nil {
		t.Fatalf("NewHeaders() error = %v", err)
	}

	if headers.Get(HeaderKsoDate) == "" {
		t.Error("NewHeaders() missing X-Kso-Date header")
	}

	if headers.Get(HeaderKsoAuthorization) == "" {
		t.Error("NewHeaders() missing X-Kso-Authorization header")
	}
}

func TestSignForWebSocket(t *testing.T) {
	headers, err := SignForWebSocket("app_id", "app_secret", "/v7/event/ws")
	if err != nil {
		t.Fatalf("SignForWebSocket() error = %v", err)
	}

	if headers.Get(HeaderKsoDate) == "" {
		t.Error("SignForWebSocket() missing X-Kso-Date header")
	}

	if headers.Get(HeaderKsoAuthorization) == "" {
		t.Error("SignForWebSocket() missing X-Kso-Authorization header")
	}
}
