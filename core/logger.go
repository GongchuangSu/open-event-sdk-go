// Package core 提供 SDK 的核心公共组件
package core

import (
	"context"
	"fmt"
	"log"
	"os"
)

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger 日志接口
type Logger interface {
	Debug(ctx context.Context, args ...interface{})
	Info(ctx context.Context, args ...interface{})
	Warn(ctx context.Context, args ...interface{})
	Error(ctx context.Context, args ...interface{})
}

// DefaultLogger 默认日志实现
type DefaultLogger struct {
	level  LogLevel
	logger *log.Logger
}

// NewDefaultLogger 创建默认日志实例
func NewDefaultLogger(level LogLevel) *DefaultLogger {
	return &DefaultLogger{
		level:  level,
		logger: log.New(os.Stdout, "[open-event-sdk] ", log.LstdFlags|log.Lmicroseconds),
	}
}

// Debug 输出 Debug 级别日志
func (l *DefaultLogger) Debug(ctx context.Context, args ...interface{}) {
	if l.level <= LogLevelDebug {
		l.logger.Printf("[DEBUG] %s", fmt.Sprint(args...))
	}
}

// Info 输出 Info 级别日志
func (l *DefaultLogger) Info(ctx context.Context, args ...interface{}) {
	if l.level <= LogLevelInfo {
		l.logger.Printf("[INFO] %s", fmt.Sprint(args...))
	}
}

// Warn 输出 Warn 级别日志
func (l *DefaultLogger) Warn(ctx context.Context, args ...interface{}) {
	if l.level <= LogLevelWarn {
		l.logger.Printf("[WARN] %s", fmt.Sprint(args...))
	}
}

// Error 输出 Error 级别日志
func (l *DefaultLogger) Error(ctx context.Context, args ...interface{}) {
	if l.level <= LogLevelError {
		l.logger.Printf("[ERROR] %s", fmt.Sprint(args...))
	}
}

// NopLogger 空日志实现，不输出任何日志
type NopLogger struct{}

// NewNopLogger 创建空日志实例
func NewNopLogger() *NopLogger {
	return &NopLogger{}
}

// Debug 不输出任何内容
func (l *NopLogger) Debug(ctx context.Context, args ...interface{}) {}

// Info 不输出任何内容
func (l *NopLogger) Info(ctx context.Context, args ...interface{}) {}

// Warn 不输出任何内容
func (l *NopLogger) Warn(ctx context.Context, args ...interface{}) {}

// Error 不输出任何内容
func (l *NopLogger) Error(ctx context.Context, args ...interface{}) {}
