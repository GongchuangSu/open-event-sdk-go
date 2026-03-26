package event

import "context"

// Handler 事件处理器接口
type Handler interface {
	// Handle 处理事件
	// ACK 模式下：处理成功时发送 code=200 的 ACK，处理失败时发送 code=500 的 ACK（服务端会触发重试）
	Handle(ctx context.Context, event *Event) error
}

// HandlerFunc 函数类型的事件处理器适配器
// 允许使用普通函数作为 Handler
type HandlerFunc func(ctx context.Context, event *Event) error

// Handle 实现 Handler 接口
func (f HandlerFunc) Handle(ctx context.Context, event *Event) error {
	return f(ctx, event)
}
