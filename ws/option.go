package ws

import (
	"time"

	"github.com/GongchuangSu/open-event-sdk-go/core"
	"github.com/GongchuangSu/open-event-sdk-go/event"
)

// Option 客户端配置选项
type Option func(*Client)

// WithEndpoint 设置 WebSocket 连接端点
// 默认值: wss://open.wps.cn/v7/event/ws
func WithEndpoint(endpoint string) Option {
	return func(c *Client) {
		if endpoint != "" {
			c.endpoint = endpoint
		}
	}
}

// WithLogger 设置自定义日志实例
func WithLogger(logger core.Logger) Option {
	return func(c *Client) {
		if logger != nil {
			c.logger = logger
		}
	}
}

// WithLogLevel 设置日志级别
// 仅在使用默认日志时生效
func WithLogLevel(level core.LogLevel) Option {
	return func(c *Client) {
		c.logLevel = level
	}
}

// WithEventHandler 设置单一事件处理器
// 所有事件都会交给这个处理器处理
func WithEventHandler(handler event.Handler) Option {
	return func(c *Client) {
		c.eventHandler = handler
	}
}

// WithEventHandlerFunc 设置函数类型的事件处理器（便捷方法）
func WithEventHandlerFunc(fn event.HandlerFunc) Option {
	return func(c *Client) {
		c.eventHandler = fn
	}
}

// WithDispatcher 设置事件分发器
// 事件会根据 event_code 分发到不同的处理器
func WithDispatcher(dispatcher *event.Dispatcher) Option {
	return func(c *Client) {
		c.dispatcher = dispatcher
	}
}

// WithAutoReconnect 设置是否开启自动重连
// 默认值: true
func WithAutoReconnect(enable bool) Option {
	return func(c *Client) {
		c.autoReconnect = enable
	}
}

// WithReconnectBaseInterval 设置重连基础间隔（指数退避起始值）
// 默认值: 1 秒
func WithReconnectBaseInterval(interval time.Duration) Option {
	return func(c *Client) {
		if interval > 0 {
			c.reconnectBaseInterval = interval
		}
	}
}

// WithReconnectMaxInterval 设置重连最大间隔（指数退避上限）
// 默认值: 60 秒
func WithReconnectMaxInterval(interval time.Duration) Option {
	return func(c *Client) {
		if interval > 0 {
			c.reconnectMaxInterval = interval
		}
	}
}

// WithReconnectMultiplier 设置重连间隔倍数
// 每次重连失败后，间隔时间乘以该倍数
// 默认值: 2.0
func WithReconnectMultiplier(multiplier float64) Option {
	return func(c *Client) {
		if multiplier > 1 {
			c.reconnectMultiplier = multiplier
		}
	}
}

// WithReconnectMaxRetry 设置最大重试次数
// -1 表示无限重试（默认）
// 0 表示不重试
// >0 表示最大重试次数
func WithReconnectMaxRetry(maxRetry int) Option {
	return func(c *Client) {
		c.reconnectMaxRetry = maxRetry
	}
}

// WithReconnectJitter 设置重连抖动系数
// 用于避免大量客户端同时重连（惊群效应）
// 取值范围: 0-1，表示在计算出的间隔基础上随机增减的比例
// 例如: 0.2 表示实际间隔在 [interval*0.8, interval*1.2] 范围内随机
// 默认值: 0.2
func WithReconnectJitter(jitter float64) Option {
	return func(c *Client) {
		if jitter >= 0 && jitter <= 1 {
			c.reconnectJitter = jitter
		}
	}
}

// WithWriteWait 设置写操作超时时间
// 默认值: 10 秒
func WithWriteWait(wait time.Duration) Option {
	return func(c *Client) {
		if wait > 0 {
			c.writeWait = wait
		}
	}
}

// WithPongWait 设置等待 Pong 响应的超时时间
// 默认值: 90 秒
func WithPongWait(wait time.Duration) Option {
	return func(c *Client) {
		if wait > 0 {
			c.pongWait = wait
		}
	}
}
