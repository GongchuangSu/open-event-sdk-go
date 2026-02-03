package core

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level LogLevel
		want  string
	}{
		{LogLevelDebug, "DEBUG"},
		{LogLevelInfo, "INFO"},
		{LogLevelWarn, "WARN"},
		{LogLevelError, "ERROR"},
		{LogLevel(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultLogger_Levels(t *testing.T) {
	tests := []struct {
		name      string
		level     LogLevel
		logMethod string
		shouldLog bool
	}{
		{"debug at debug level", LogLevelDebug, "debug", true},
		{"info at debug level", LogLevelDebug, "info", true},
		{"warn at debug level", LogLevelDebug, "warn", true},
		{"error at debug level", LogLevelDebug, "error", true},

		{"debug at info level", LogLevelInfo, "debug", false},
		{"info at info level", LogLevelInfo, "info", true},
		{"warn at info level", LogLevelInfo, "warn", true},
		{"error at info level", LogLevelInfo, "error", true},

		{"debug at warn level", LogLevelWarn, "debug", false},
		{"info at warn level", LogLevelWarn, "info", false},
		{"warn at warn level", LogLevelWarn, "warn", true},
		{"error at warn level", LogLevelWarn, "error", true},

		{"debug at error level", LogLevelError, "debug", false},
		{"info at error level", LogLevelError, "info", false},
		{"warn at error level", LogLevelError, "warn", false},
		{"error at error level", LogLevelError, "error", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建一个缓冲区来捕获日志输出
			var buf bytes.Buffer
			logger := &DefaultLogger{
				level:  tt.level,
				logger: log.New(&buf, "[test] ", 0),
			}

			ctx := context.Background()

			switch tt.logMethod {
			case "debug":
				logger.Debug(ctx, "test message")
			case "info":
				logger.Info(ctx, "test message")
			case "warn":
				logger.Warn(ctx, "test message")
			case "error":
				logger.Error(ctx, "test message")
			}

			output := buf.String()
			hasOutput := len(output) > 0

			if hasOutput != tt.shouldLog {
				t.Errorf("expected shouldLog = %v, but hasOutput = %v, output = %q",
					tt.shouldLog, hasOutput, output)
			}

			if tt.shouldLog {
				if !strings.Contains(output, "test message") {
					t.Errorf("expected output to contain 'test message', got %q", output)
				}
			}
		})
	}
}

func TestNewDefaultLogger(t *testing.T) {
	logger := NewDefaultLogger(LogLevelInfo)

	if logger == nil {
		t.Fatal("NewDefaultLogger() returned nil")
	}

	if logger.level != LogLevelInfo {
		t.Errorf("expected level = LogLevelInfo, got %v", logger.level)
	}

	if logger.logger == nil {
		t.Error("expected logger.logger to be initialized")
	}
}

func TestNopLogger(t *testing.T) {
	logger := NewNopLogger()

	// 验证不会 panic
	ctx := context.Background()
	logger.Debug(ctx, "test")
	logger.Info(ctx, "test")
	logger.Warn(ctx, "test")
	logger.Error(ctx, "test")

	// 验证实现了 Logger 接口
	var _ Logger = logger
}

func TestLogger_Interface(t *testing.T) {
	// 验证 DefaultLogger 实现了 Logger 接口
	var _ Logger = &DefaultLogger{}

	// 验证 NopLogger 实现了 Logger 接口
	var _ Logger = &NopLogger{}
}
