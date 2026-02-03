// Package protocol 定义 WebSocket 协议消息结构（内部使用）
package protocol

// EventMessage 事件消息（服务端 -> 客户端）
// 事件消息不包含 type 字段，通过 topic 和 operation 字段判断
type EventMessage struct {
	// Topic 消息主题（根据不同事件而定）
	Topic string `json:"topic"`

	// Operation 消息变更动作（根据不同事件而定）
	Operation string `json:"operation"`

	// Time 时间（秒为单位的时间戳）
	Time int64 `json:"time"`

	// Nonce iv 向量（解密时使用）
	Nonce string `json:"nonce"`

	// Signature 消息签名
	Signature string `json:"signature"`

	// EncryptedData 消息变更的加密字段
	EncryptedData string `json:"encrypted_data"`
}

// GoAwayMessage 关闭通知消息（服务端 -> 客户端）
type GoAwayMessage struct {
	// Type 消息类型，固定为 "goaway"
	Type string `json:"type"`

	// Reason 关闭原因
	// 可选值: server_shutdown, connection_replaced, heartbeat_timeout
	Reason string `json:"reason"`

	// Message 关闭原因描述
	Message string `json:"message"`

	// ReconnectMs 建议重连延迟（毫秒），仅在 server_shutdown 时提供
	ReconnectMs int `json:"reconnect_ms,omitempty"`
}
