package ws

import "errors"

// SDK 错误定义
var (
	// ErrNotConnected WebSocket 未连接
	ErrNotConnected = errors.New("websocket not connected")

	// ErrAlreadyConnected WebSocket 已连接
	ErrAlreadyConnected = errors.New("websocket already connected")

	// ErrConnectionClosed 连接已关闭
	ErrConnectionClosed = errors.New("connection closed")

	// ErrAuthFailed 认证失败
	ErrAuthFailed = errors.New("authentication failed")

	// ErrInvalidMessage 无效的消息格式
	ErrInvalidMessage = errors.New("invalid message format")

	// ErrHandlerNotSet 事件处理器未设置
	ErrHandlerNotSet = errors.New("event handler not set")

	// ErrReconnectExceeded 超过最大重连次数
	ErrReconnectExceeded = errors.New("max reconnect attempts exceeded")

	// ErrClientClosed 客户端已关闭
	ErrClientClosed = errors.New("client is closed")
)

// ClientError 客户端错误，通常是配置或使用问题，不应重试
type ClientError struct {
	Code    int
	Message string
}

func (e *ClientError) Error() string {
	return e.Message
}

// NewClientError 创建客户端错误
func NewClientError(code int, message string) *ClientError {
	return &ClientError{
		Code:    code,
		Message: message,
	}
}

// ServerError 服务端错误，可能是临时性的，可以重试
type ServerError struct {
	Code    int
	Message string
}

func (e *ServerError) Error() string {
	return e.Message
}

// NewServerError 创建服务端错误
func NewServerError(code int, message string) *ServerError {
	return &ServerError{
		Code:    code,
		Message: message,
	}
}
