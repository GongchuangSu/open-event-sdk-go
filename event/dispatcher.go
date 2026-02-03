package event

import (
	"context"
	"fmt"
	"sync"
)

// Dispatcher 事件分发器
// 支持按事件编码（event_code）注册不同的处理器
type Dispatcher struct {
	handlers map[string]Handler
	fallback Handler
	mu       sync.RWMutex
}

// NewDispatcher 创建事件分发器
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string]Handler),
	}
}

// Register 注册特定事件编码的处理器
// eventCode: 事件编码，由 topic.operation 组成，如 "kso.app_chat.message.create"
// handler: 事件处理器
func (d *Dispatcher) Register(eventCode string, handler Handler) *Dispatcher {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[eventCode] = handler
	return d
}

// RegisterFunc 注册函数类型的处理器（便捷方法）
func (d *Dispatcher) RegisterFunc(eventCode string, fn func(ctx context.Context, event *Event) error) *Dispatcher {
	return d.Register(eventCode, HandlerFunc(fn))
}

// RegisterFallback 注册兜底处理器
// 当事件没有匹配的处理器时，会调用兜底处理器
func (d *Dispatcher) RegisterFallback(handler Handler) *Dispatcher {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.fallback = handler
	return d
}

// RegisterFallbackFunc 注册函数类型的兜底处理器（便捷方法）
func (d *Dispatcher) RegisterFallbackFunc(fn func(ctx context.Context, event *Event) error) *Dispatcher {
	return d.RegisterFallback(HandlerFunc(fn))
}

// Dispatch 分发事件到对应的处理器
func (d *Dispatcher) Dispatch(ctx context.Context, event *Event) error {
	d.mu.RLock()
	handler, ok := d.handlers[event.EventCode()]
	fallback := d.fallback
	d.mu.RUnlock()

	if ok {
		return handler.Handle(ctx, event)
	}

	if fallback != nil {
		return fallback.Handle(ctx, event)
	}

	// 没有匹配的处理器，返回错误
	return fmt.Errorf("no handler registered for event_code: %s", event.EventCode())
}

// Handle 实现 Handler 接口，使 Dispatcher 可以作为 Handler 使用
func (d *Dispatcher) Handle(ctx context.Context, event *Event) error {
	return d.Dispatch(ctx, event)
}

// HasHandler 检查是否有处理器注册了指定的事件编码
func (d *Dispatcher) HasHandler(eventCode string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, ok := d.handlers[eventCode]
	return ok
}

// EventCodes 返回所有已注册的事件编码
func (d *Dispatcher) EventCodes() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	codes := make([]string, 0, len(d.handlers))
	for code := range d.handlers {
		codes = append(codes, code)
	}
	return codes
}
