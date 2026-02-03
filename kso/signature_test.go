package kso

import (
	"net/http"
	"strings"
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
				AppId:       "test_app_id",
				AppSecret:   "test_app_secret",
				Method:      http.MethodGet,
				RequestURI:  "/v7/event/ws",
				ContentType: "",
				Body:        nil,
			},
			wantErr: false,
		},
		{
			name: "sign with body",
			params: &SignParams{
				AppId:       "test_app_id",
				AppSecret:   "test_app_secret",
				Method:      http.MethodPost,
				RequestURI:  "/v7/api/test",
				ContentType: "application/json",
				Body:        []byte(`{"key":"value"}`),
			},
			wantErr: false,
		},
		{
			name: "sign with custom date",
			params: &SignParams{
				AppId:      "test_app_id",
				AppSecret:  "test_app_secret",
				Method:     http.MethodGet,
				RequestURI: "/v7/event/ws",
				Date:       "Wed, 23 Jan 2013 06:43:08 GMT",
			},
			wantErr: false,
		},
		{
			name: "missing app_id",
			params: &SignParams{
				AppSecret:  "test_app_secret",
				Method:     http.MethodGet,
				RequestURI: "/v7/event/ws",
			},
			wantErr: true,
		},
		{
			name: "missing app_secret",
			params: &SignParams{
				AppId:      "test_app_id",
				Method:     http.MethodGet,
				RequestURI: "/v7/event/ws",
			},
			wantErr: true,
		},
		{
			name: "missing method",
			params: &SignParams{
				AppId:      "test_app_id",
				AppSecret:  "test_app_secret",
				RequestURI: "/v7/event/ws",
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
				// 验证返回结果
				if result == nil {
					t.Error("Sign() returned nil result")
					return
				}

				// 验证 Date 不为空
				if result.Date == "" {
					t.Error("Sign() Date is empty")
				}

				// 验证 Authorization 格式
				if !strings.HasPrefix(result.Authorization, "KSO-1 ") {
					t.Errorf("Sign() Authorization format invalid: %s", result.Authorization)
				}

				// 验证 Authorization 包含 app_id
				if !strings.Contains(result.Authorization, tt.params.AppId) {
					t.Errorf("Sign() Authorization does not contain app_id: %s", result.Authorization)
				}

				// 验证 Authorization 包含签名（冒号分隔）
				parts := strings.Split(result.Authorization, ":")
				if len(parts) != 2 {
					t.Errorf("Sign() Authorization format invalid, expected 2 parts: %s", result.Authorization)
				}

				// 验证签名是 64 字符的十六进制（SHA256）
				signature := parts[1]
				if len(signature) != 64 {
					t.Errorf("Sign() signature length = %d, want 64", len(signature))
				}
			}
		})
	}
}

func TestSignConsistency(t *testing.T) {
	// 使用固定的日期，验证签名一致性
	params := &SignParams{
		AppId:       "test_app_id",
		AppSecret:   "test_app_secret",
		Method:      http.MethodGet,
		RequestURI:  "/v7/event/ws",
		ContentType: "",
		Body:        nil,
		Date:        "Wed, 23 Jan 2013 06:43:08 GMT",
	}

	// 多次签名应该得到相同的结果
	result1, err := Sign(params)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	result2, err := Sign(params)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if result1.Authorization != result2.Authorization {
		t.Errorf("Sign() inconsistent results: %s vs %s", result1.Authorization, result2.Authorization)
	}

	if result1.Date != result2.Date {
		t.Errorf("Sign() inconsistent dates: %s vs %s", result1.Date, result2.Date)
	}
}

func TestNewHeaders(t *testing.T) {
	headers, err := NewHeaders("test_app_id", "test_app_secret", http.MethodGet, "/v7/event/ws", "", nil)
	if err != nil {
		t.Fatalf("NewHeaders() error = %v", err)
	}

	// 验证 headers 不为空
	if headers == nil {
		t.Fatal("NewHeaders() returned nil")
	}

	// 验证包含必要的头
	if headers.Get(HeaderKsoDate) == "" {
		t.Error("NewHeaders() missing X-Kso-Date")
	}

	if headers.Get(HeaderKsoAuthorization) == "" {
		t.Error("NewHeaders() missing X-Kso-Authorization")
	}
}

func TestSignForWebSocket(t *testing.T) {
	headers, err := SignForWebSocket("test_app_id", "test_app_secret", "/v7/event/ws")
	if err != nil {
		t.Fatalf("SignForWebSocket() error = %v", err)
	}

	// 验证 headers 不为空
	if headers == nil {
		t.Fatal("SignForWebSocket() returned nil")
	}

	// 验证包含必要的头
	if headers.Get(HeaderKsoDate) == "" {
		t.Error("SignForWebSocket() missing X-Kso-Date")
	}

	if headers.Get(HeaderKsoAuthorization) == "" {
		t.Error("SignForWebSocket() missing X-Kso-Authorization")
	}

	// Content-Type 应该为空
	if headers.Get(HeaderContentType) != "" {
		t.Errorf("SignForWebSocket() unexpected Content-Type: %s", headers.Get(HeaderContentType))
	}
}
