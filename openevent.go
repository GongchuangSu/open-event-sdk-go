// Package openevent 提供事件订阅 SDK
//
// 该 SDK 用于接收和处理来自开放平台的事件推送。
//
// 快速开始:
//
//	client := openevent.NewClient("app_id", "app_secret",
//	    openevent.WithEventHandlerFunc(func(ctx context.Context, e *openevent.Event) error {
//	        log.Printf("收到事件: %s", e.EventCode())
//	        return nil
//	    }),
//	)
//	client.Start(context.Background())
//
// 使用 Dispatcher 分发事件:
//
//	dispatcher := openevent.NewDispatcher().
//	    // 已支持的事件使用 OnV7XXX 方法（类型安全）
//	    OnV7AppChatMessageCreate(func(ctx context.Context, e *openevent.V7AppChatMessageCreateEvent) error {
//	        log.Printf("收到消息: %s", e.Data.Message.Id)
//	        return nil
//	    })
//
//	// 其他事件使用 RegisterFunc 方法
//	dispatcher.RegisterFunc("kso.other.event", func(ctx context.Context, e *openevent.Event) error {
//	    log.Printf("其他事件: %s", e.Data)
//	    return nil
//	})
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

// ----- 类型化事件 -----
// 注意: 如需使用泛型类型 TypedEvent[T] 和 TypedEventHandler[T]，请直接从 event 包导入:
//   import "github.com/GongchuangSu/open-event-sdk-go/event"

// V7AppChatMessageCreateEvent 应用收到消息事件
type V7AppChatMessageCreateEvent = event.V7AppChatMessageCreateEvent

// V7AppChatCreateEvent 应用会话创建事件
type V7AppChatCreateEvent = event.V7AppChatCreateEvent

// V7AppGroupChatDeleteEvent 群聊解散事件
type V7AppGroupChatDeleteEvent = event.V7AppGroupChatDeleteEvent

// V7AppGroupChatMemberUserCreateEvent 用户进群事件
type V7AppGroupChatMemberUserCreateEvent = event.V7AppGroupChatMemberUserCreateEvent

// V7AppGroupChatMemberUserDeleteEvent 用户退群事件
type V7AppGroupChatMemberUserDeleteEvent = event.V7AppGroupChatMemberUserDeleteEvent

// V7AppGroupChatMemberRobotCreateEvent 机器人进群事件
type V7AppGroupChatMemberRobotCreateEvent = event.V7AppGroupChatMemberRobotCreateEvent

// V7AppGroupChatMemberRobotDeleteEvent 机器人退群事件
type V7AppGroupChatMemberRobotDeleteEvent = event.V7AppGroupChatMemberRobotDeleteEvent

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

// WithAckMode 设置是否启用 ACK 模式（默认: true）
// 启用后，事件处理结果会发送给服务端，处理失败时服务端会触发重试
var WithAckMode = ws.WithAckMode

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
