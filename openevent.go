// Package openevent 提供事件订阅 SDK
//
// 该 SDK 用于接收和处理来自开放平台的事件推送。
//
// 快速开始:
//
//	import openevent "github.com/GongchuangSu/open-event-sdk-go"
//
//	func main() {
//	    client := openevent.NewClient("your_app_id", "your_app_secret",
//	        openevent.WithEventHandlerFunc(func(ctx context.Context, e *openevent.Event) error {
//	            // EventCode = Topic.Operation，如 "kso.app_chat.message.create"
//	            log.Printf("收到事件: event_code=%s, time=%d", e.EventCode(), e.Time)
//	            log.Printf("事件数据: %s", e.Data)
//	            return nil
//	        }),
//	    )
//
//	    if err := client.Start(context.Background()); err != nil {
//	        log.Fatal(err)
//	    }
//	}
//
// 使用事件分发器:
//
//	dispatcher := openevent.NewDispatcher()
//	dispatcher.RegisterFunc("kso.app_chat.message.create", handleChatMessage)
//	dispatcher.RegisterFunc("kso.user.status.update", handleUserStatusUpdate)
//
//	client := openevent.NewClient("app_id", "app_secret",
//	    openevent.WithDispatcher(dispatcher),
//	)
package openevent

import (
	"github.com/GongchuangSu/open-event-sdk-go/core"
	"github.com/GongchuangSu/open-event-sdk-go/event"
	"github.com/GongchuangSu/open-event-sdk-go/ws"
)

// ----- 核心类型 -----

// Client WebSocket 长连接客户端
type Client = ws.Client

// Option 客户端配置选项
type Option = ws.Option

// Event 事件实体
type Event = event.Event

// Handler 事件处理器接口
type Handler = event.Handler

// HandlerFunc 函数类型的事件处理器
type HandlerFunc = event.HandlerFunc

// Dispatcher 事件分发器
type Dispatcher = event.Dispatcher

// Logger 日志接口
type Logger = core.Logger

// LogLevel 日志级别
type LogLevel = core.LogLevel

// ----- 错误类型 -----

// ClientError 客户端错误
type ClientError = ws.ClientError

// ServerError 服务端错误
type ServerError = ws.ServerError

// ----- 构造函数 -----

// NewClient 创建 WebSocket 客户端
//
// 参数:
//   - appId: 应用 ID
//   - appSecret: 应用密钥
//   - opts: 可选配置项
var NewClient = ws.NewClient

// NewDispatcher 创建事件分发器
var NewDispatcher = event.NewDispatcher

// NewDefaultLogger 创建默认日志实例
var NewDefaultLogger = core.NewDefaultLogger

// NewNopLogger 创建空日志实例（不输出任何日志）
var NewNopLogger = core.NewNopLogger

// ----- 配置选项 -----

// WithEndpoint 设置 WebSocket 连接端点
// 默认使用 SDK 内置端点，一般无需设置
var WithEndpoint = ws.WithEndpoint

// WithLogger 设置自定义日志实例
var WithLogger = ws.WithLogger

// WithLogLevel 设置日志级别
var WithLogLevel = ws.WithLogLevel

// WithEventHandler 设置单一事件处理器
var WithEventHandler = ws.WithEventHandler

// WithEventHandlerFunc 设置函数类型的事件处理器
var WithEventHandlerFunc = ws.WithEventHandlerFunc

// WithDispatcher 设置事件分发器
var WithDispatcher = ws.WithDispatcher

// WithAutoReconnect 设置是否开启自动重连（默认: true）
var WithAutoReconnect = ws.WithAutoReconnect

// WithReconnectBaseInterval 设置重连基础间隔（默认: 1秒）
var WithReconnectBaseInterval = ws.WithReconnectBaseInterval

// WithReconnectMaxInterval 设置重连最大间隔（默认: 60秒）
var WithReconnectMaxInterval = ws.WithReconnectMaxInterval

// WithReconnectMultiplier 设置重连间隔倍数（默认: 2.0）
var WithReconnectMultiplier = ws.WithReconnectMultiplier

// WithReconnectMaxRetry 设置最大重试次数（默认: -1 无限重试）
var WithReconnectMaxRetry = ws.WithReconnectMaxRetry

// WithReconnectJitter 设置重连抖动系数（默认: 0.2）
var WithReconnectJitter = ws.WithReconnectJitter

// WithWriteWait 设置写操作超时时间（默认: 10秒）
var WithWriteWait = ws.WithWriteWait

// WithPongWait 设置等待 Pong 响应超时时间（默认: 90秒）
var WithPongWait = ws.WithPongWait

// ----- 日志级别常量 -----

const (
	// LogLevelDebug 调试级别
	LogLevelDebug = core.LogLevelDebug

	// LogLevelInfo 信息级别
	LogLevelInfo = core.LogLevelInfo

	// LogLevelWarn 警告级别
	LogLevelWarn = core.LogLevelWarn

	// LogLevelError 错误级别
	LogLevelError = core.LogLevelError
)

// ----- 错误变量 -----

var (
	// ErrHandlerNotSet 事件处理器未设置
	ErrHandlerNotSet = ws.ErrHandlerNotSet

	// ErrClientClosed 客户端已关闭
	ErrClientClosed = ws.ErrClientClosed

	// ErrReconnectExceeded 超过最大重连次数
	ErrReconnectExceeded = ws.ErrReconnectExceeded
)
