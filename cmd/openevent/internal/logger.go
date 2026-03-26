package internal

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/GongchuangSu/open-event-sdk-go/core"
)

// stderrLogger 将 SDK 日志输出到 stderr，避免污染 stdout 的事件数据
type stderrLogger struct {
	level  core.LogLevel
	logger *log.Logger
}

// NewStderrLogger 创建写入 stderr 的日志实例
func NewStderrLogger(level core.LogLevel) core.Logger {
	return &stderrLogger{
		level:  level,
		logger: log.New(os.Stderr, "[open-event-sdk] ", log.LstdFlags|log.Lmicroseconds),
	}
}

func (l *stderrLogger) Debug(ctx context.Context, args ...interface{}) {
	if l.level <= core.LogLevelDebug {
		l.logger.Printf("[DEBUG] %s", fmt.Sprint(args...))
	}
}

func (l *stderrLogger) Info(ctx context.Context, args ...interface{}) {
	if l.level <= core.LogLevelInfo {
		l.logger.Printf("[INFO] %s", fmt.Sprint(args...))
	}
}

func (l *stderrLogger) Warn(ctx context.Context, args ...interface{}) {
	if l.level <= core.LogLevelWarn {
		l.logger.Printf("[WARN] %s", fmt.Sprint(args...))
	}
}

func (l *stderrLogger) Error(ctx context.Context, args ...interface{}) {
	if l.level <= core.LogLevelError {
		l.logger.Printf("[ERROR] %s", fmt.Sprint(args...))
	}
}
