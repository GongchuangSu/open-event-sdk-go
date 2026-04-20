package ws

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GongchuangSu/open-event-sdk-go/core"
	"github.com/GongchuangSu/open-event-sdk-go/event"
	"github.com/GongchuangSu/open-event-sdk-go/internal/protocol"
)

func TestNewClient_DefaultOptions(t *testing.T) {
	c := NewClient("app_id", "app_secret")

	assert.Equal(t, "app_id", c.appId)
	assert.Equal(t, "app_secret", c.appSecret)
	assert.Equal(t, protocol.DefaultEndpoint, c.endpoint)
	assert.Equal(t, protocol.DefaultAutoReconnect, c.autoReconnect)
	assert.Equal(t, protocol.DefaultReconnectBaseInterval, c.reconnectBaseInterval)
	assert.Equal(t, protocol.DefaultReconnectMaxInterval, c.reconnectMaxInterval)
	assert.Equal(t, protocol.DefaultReconnectMultiplier, c.reconnectMultiplier)
	assert.Equal(t, protocol.DefaultReconnectMaxRetry, c.reconnectMaxRetry)
	assert.Equal(t, protocol.DefaultReconnectJitter, c.reconnectJitter)
	assert.Equal(t, protocol.DefaultWriteWait, c.writeWait)
	assert.Equal(t, protocol.DefaultPongWait, c.pongWait)
	assert.Equal(t, protocol.DefaultAckMode, c.ackMode)
	assert.NotNil(t, c.logger)
	assert.NotNil(t, c.sendChan)
	assert.NotNil(t, c.stopChan)
}

func TestNewClient_CustomOptions(t *testing.T) {
	handler := event.HandlerFunc(func(ctx context.Context, e *event.Event) error {
		return nil
	})
	logger := core.NewNopLogger()

	c := NewClient("id", "secret",
		WithEndpoint("wss://custom.endpoint/ws"),
		WithLogger(logger),
		WithAutoReconnect(false),
		WithReconnectBaseInterval(5*time.Second),
		WithReconnectMaxInterval(120*time.Second),
		WithReconnectMultiplier(3.0),
		WithReconnectMaxRetry(10),
		WithReconnectJitter(0.5),
		WithWriteWait(20*time.Second),
		WithPongWait(180*time.Second),
		WithAckMode(false),
		WithEventHandler(handler),
	)

	assert.Equal(t, "wss://custom.endpoint/ws", c.endpoint)
	assert.Equal(t, logger, c.logger)
	assert.False(t, c.autoReconnect)
	assert.Equal(t, 5*time.Second, c.reconnectBaseInterval)
	assert.Equal(t, 120*time.Second, c.reconnectMaxInterval)
	assert.Equal(t, 3.0, c.reconnectMultiplier)
	assert.Equal(t, 10, c.reconnectMaxRetry)
	assert.Equal(t, 0.5, c.reconnectJitter)
	assert.Equal(t, 20*time.Second, c.writeWait)
	assert.Equal(t, 180*time.Second, c.pongWait)
	assert.False(t, c.ackMode)
	assert.NotNil(t, c.eventHandler)
}

func TestWithEndpoint_IgnoresEmpty(t *testing.T) {
	c := NewClient("id", "secret", WithEndpoint(""))
	assert.Equal(t, protocol.DefaultEndpoint, c.endpoint)
}

func TestWithBaseUrl(t *testing.T) {
	tests := []struct {
		name         string
		baseUrl      string
		wantEndpoint string
		wantSignPath string
	}{
		{
			name:         "simple base url",
			baseUrl:      "wss://openapi-intl.example.com",
			wantEndpoint: "wss://openapi-intl.example.com/v7/event/ws",
			wantSignPath: "/v7/event/ws",
		},
		{
			name:         "base url with path prefix",
			baseUrl:      "wss://ap.wps.com/openapi",
			wantEndpoint: "wss://ap.wps.com/openapi/v7/event/ws",
			wantSignPath: "/v7/event/ws",
		},
		{
			name:         "base url with trailing slash",
			baseUrl:      "wss://ap.wps.com/openapi/",
			wantEndpoint: "wss://ap.wps.com/openapi/v7/event/ws",
			wantSignPath: "/v7/event/ws",
		},
		{
			name:         "empty base url keeps default",
			baseUrl:      "",
			wantEndpoint: protocol.DefaultEndpoint,
			wantSignPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient("id", "secret", WithBaseUrl(tt.baseUrl))
			assert.Equal(t, tt.wantEndpoint, c.endpoint)
			assert.Equal(t, tt.wantSignPath, c.signPath)
		})
	}
}

func TestWithBaseUrl_SignPathUsedForSigning(t *testing.T) {
	c := NewClient("id", "secret",
		WithBaseUrl("wss://ap.wps.com/openapi"),
	)

	assert.Equal(t, "wss://ap.wps.com/openapi/v7/event/ws", c.endpoint)
	assert.Equal(t, "/v7/event/ws", c.signPath)
}

func TestWithEndpoint_SignPathNotSet(t *testing.T) {
	c := NewClient("id", "secret",
		WithEndpoint("wss://custom.example.com/v7/event/ws"),
	)

	assert.Equal(t, "wss://custom.example.com/v7/event/ws", c.endpoint)
	assert.Equal(t, "", c.signPath)
}

func TestWithLogger_IgnoresNil(t *testing.T) {
	c := NewClient("id", "secret", WithLogger(nil))
	assert.NotNil(t, c.logger)
}

func TestWithReconnectBaseInterval_IgnoresInvalid(t *testing.T) {
	c := NewClient("id", "secret", WithReconnectBaseInterval(0))
	assert.Equal(t, protocol.DefaultReconnectBaseInterval, c.reconnectBaseInterval)

	c2 := NewClient("id", "secret", WithReconnectBaseInterval(-1*time.Second))
	assert.Equal(t, protocol.DefaultReconnectBaseInterval, c2.reconnectBaseInterval)
}

func TestWithReconnectMultiplier_IgnoresInvalid(t *testing.T) {
	c := NewClient("id", "secret", WithReconnectMultiplier(0.5))
	assert.Equal(t, protocol.DefaultReconnectMultiplier, c.reconnectMultiplier)

	c2 := NewClient("id", "secret", WithReconnectMultiplier(1.0))
	assert.Equal(t, protocol.DefaultReconnectMultiplier, c2.reconnectMultiplier)
}

func TestWithReconnectJitter_IgnoresInvalid(t *testing.T) {
	c := NewClient("id", "secret", WithReconnectJitter(-0.1))
	assert.Equal(t, protocol.DefaultReconnectJitter, c.reconnectJitter)

	c2 := NewClient("id", "secret", WithReconnectJitter(1.5))
	assert.Equal(t, protocol.DefaultReconnectJitter, c2.reconnectJitter)
}

func TestWithDispatcher(t *testing.T) {
	d := event.NewDispatcher()
	c := NewClient("id", "secret", WithDispatcher(d))
	assert.Equal(t, d, c.dispatcher)
}

func TestWithEventHandlerFunc(t *testing.T) {
	called := false
	fn := event.HandlerFunc(func(ctx context.Context, e *event.Event) error {
		called = true
		return nil
	})
	c := NewClient("id", "secret", WithEventHandlerFunc(fn))
	assert.NotNil(t, c.eventHandler)

	_ = c.eventHandler.Handle(context.Background(), &event.Event{})
	assert.True(t, called)
}

func TestWithLogLevel(t *testing.T) {
	c := NewClient("id", "secret", WithLogLevel(core.LogLevelDebug))
	assert.Equal(t, core.LogLevelDebug, c.logLevel)
}

func TestClient_Start_NoHandler(t *testing.T) {
	c := NewClient("id", "secret")
	err := c.Start(context.Background())
	assert.ErrorIs(t, err, ErrHandlerNotSet)
}

func TestClient_Start_AfterClosed(t *testing.T) {
	c := NewClient("id", "secret",
		WithEventHandlerFunc(func(ctx context.Context, e *event.Event) error {
			return nil
		}),
	)
	c.Stop()

	err := c.Start(context.Background())
	assert.ErrorIs(t, err, ErrClientClosed)
}

func TestClient_Stop_Idempotent(t *testing.T) {
	c := NewClient("id", "secret")

	err1 := c.Stop()
	assert.NoError(t, err1)

	err2 := c.Stop()
	assert.NoError(t, err2)
}

func TestClient_IsConnected(t *testing.T) {
	c := NewClient("id", "secret")
	assert.False(t, c.IsConnected())

	c.Stop()
	assert.False(t, c.IsConnected())
}

func TestCalculateBackoffWithJitter(t *testing.T) {
	c := NewClient("id", "secret")

	t.Run("no jitter", func(t *testing.T) {
		c.reconnectJitter = 0
		result := c.calculateBackoffWithJitter(10 * time.Second)
		assert.Equal(t, 10*time.Second, result)
	})

	t.Run("with jitter stays in range", func(t *testing.T) {
		c.reconnectJitter = 0.2
		interval := 10 * time.Second

		for i := 0; i < 100; i++ {
			result := c.calculateBackoffWithJitter(interval)
			minExpected := time.Duration(float64(interval) * 0.8)
			maxExpected := time.Duration(float64(interval) * 1.2)
			assert.GreaterOrEqual(t, result, minExpected,
				"result %v should be >= %v", result, minExpected)
			assert.LessOrEqual(t, result, maxExpected,
				"result %v should be <= %v", result, maxExpected)
		}
	})

	t.Run("full jitter", func(t *testing.T) {
		c.reconnectJitter = 1.0
		interval := 10 * time.Second

		for i := 0; i < 100; i++ {
			result := c.calculateBackoffWithJitter(interval)
			assert.GreaterOrEqual(t, result, time.Duration(0))
			assert.LessOrEqual(t, result, 2*interval)
		}
	})
}

func TestParseConnectError(t *testing.T) {
	c := NewClient("id", "secret")

	tests := []struct {
		name       string
		statusCode int
		wantType   string
		wantMsg    string
	}{
		{
			name:       "401 unauthorized",
			statusCode: http.StatusUnauthorized,
			wantType:   "client",
			wantMsg:    "authentication failed",
		},
		{
			name:       "403 forbidden",
			statusCode: http.StatusForbidden,
			wantType:   "client",
			wantMsg:    "forbidden",
		},
		{
			name:       "429 too many requests",
			statusCode: http.StatusTooManyRequests,
			wantType:   "server",
			wantMsg:    "too many connections",
		},
		{
			name:       "500 internal server error",
			statusCode: http.StatusInternalServerError,
			wantType:   "server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: tt.statusCode}
			err := c.parseConnectError(resp)
			require.Error(t, err)

			switch tt.wantType {
			case "client":
				var clientErr *ClientError
				assert.ErrorAs(t, err, &clientErr)
				assert.Equal(t, tt.statusCode, clientErr.Code)
				if tt.wantMsg != "" {
					assert.Contains(t, clientErr.Message, tt.wantMsg)
				}
			case "server":
				var serverErr *ServerError
				assert.ErrorAs(t, err, &serverErr)
				assert.Equal(t, tt.statusCode, serverErr.Code)
				if tt.wantMsg != "" {
					assert.Contains(t, serverErr.Message, tt.wantMsg)
				}
			}
		})
	}
}

func TestClientError(t *testing.T) {
	err := NewClientError(401, "auth failed")
	assert.Equal(t, "auth failed", err.Error())
	assert.Equal(t, 401, err.Code)
}

func TestServerError(t *testing.T) {
	err := NewServerError(500, "internal error")
	assert.Equal(t, "internal error", err.Error())
	assert.Equal(t, 500, err.Code)
}
