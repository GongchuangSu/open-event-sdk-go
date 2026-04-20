// Package protocol 定义 WebSocket 协议常量（内部使用）
package protocol

import "time"

const (
	// DefaultBaseUrl 默认的 WebSocket 基础地址
	DefaultBaseUrl = "wss://openapi.wps.cn"

	// DefaultEventPath 默认的事件 WebSocket 路径
	DefaultEventPath = "/v7/event/ws"

	// DefaultEndpoint 默认的 WebSocket 连接端点
	DefaultEndpoint = DefaultBaseUrl + DefaultEventPath

	// MessageTypeGoAway 关闭通知消息类型
	// 注：事件消息不包含 type 字段，通过 topic 和 operation 识别
	MessageTypeGoAway = "goaway"
)

// GoAway 原因常量
const (
	// GoAwayReasonServerShutdown 服务器关闭
	GoAwayReasonServerShutdown = "server_shutdown"

	// GoAwayReasonConnectionReplaced 连接被新连接替换
	GoAwayReasonConnectionReplaced = "connection_replaced"

	// GoAwayReasonHeartbeatTimeout 心跳超时
	GoAwayReasonHeartbeatTimeout = "heartbeat_timeout"
)

// 默认配置常量
const (
	// DefaultAckMode 默认开启 ACK 模式
	DefaultAckMode = true

	// DefaultAutoReconnect 默认开启自动重连
	DefaultAutoReconnect = true

	// DefaultReconnectBaseInterval 默认重连基础间隔（指数退避起始值）
	DefaultReconnectBaseInterval = 1 * time.Second

	// DefaultReconnectMaxInterval 默认重连最大间隔（指数退避上限）
	DefaultReconnectMaxInterval = 60 * time.Second

	// DefaultReconnectMultiplier 默认重连间隔倍数
	DefaultReconnectMultiplier = 2.0

	// DefaultReconnectMaxRetry 默认最大重试次数，-1 表示无限重试
	DefaultReconnectMaxRetry = -1

	// DefaultReconnectJitter 默认重连抖动系数（0-1之间，表示在计算出的间隔基础上随机增减的比例）
	DefaultReconnectJitter = 0.2

	// DefaultWriteWait 默认写超时
	DefaultWriteWait = 10 * time.Second

	// DefaultPongWait 默认等待 Pong 的超时时间
	DefaultPongWait = 90 * time.Second
)
