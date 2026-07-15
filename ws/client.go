// Package ws 提供 WebSocket 长连接客户端
package ws

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/GongchuangSu/open-event-sdk-go/core"
	"github.com/GongchuangSu/open-event-sdk-go/event"
	"github.com/GongchuangSu/open-event-sdk-go/internal/kso"
	"github.com/GongchuangSu/open-event-sdk-go/internal/protocol"
)

// Client WebSocket 长连接客户端
type Client struct {
	// 认证信息
	appId     string
	appSecret string

	// 连接配置
	endpoint string // 完整连接 URL
	signPath string // 签名用的路径（由 WithBaseUrl 设置，为空时从 endpoint 解析）
	conn     *websocket.Conn
	connURL  *url.URL

	// 事件处理
	eventHandler event.Handler
	dispatcher   *event.Dispatcher

	// 日志
	logger   core.Logger
	logLevel core.LogLevel

	// 重连配置（指数退避策略）
	autoReconnect         bool
	reconnectBaseInterval time.Duration // 基础间隔（起始值）
	reconnectMaxInterval  time.Duration // 最大间隔（上限）
	reconnectMultiplier   float64       // 间隔倍数
	reconnectMaxRetry     int           // 最大重试次数
	reconnectJitter       float64       // 抖动系数（0-1）

	// 超时配置
	writeWait time.Duration
	pongWait  time.Duration

	// 状态管理
	mu             sync.Mutex
	closed         bool
	stopChan       chan struct{}
	receivedGoAway bool // 是否已收到 GoAway 消息

	// 写操作保护（所有 WebSocket 写操作必须通过 writeMessage 方法）
	writeMu sync.Mutex

	// 消息发送通道（用于串行化 WebSocket 写操作）
	sendChan chan []byte

	// ACK 模式配置
	ackMode bool // 是否启用 ACK 模式（处理结果反馈，支持服务端重试）

	// TLS 配置
	tlsVerify bool // 是否校验服务端 TLS 证书（默认 false，跳过校验）
}

// NewClient 创建 WebSocket 客户端
//
// 参数:
//   - appId: 应用 ID
//   - appSecret: 应用密钥
//   - opts: 可选配置项
//
// 示例:
//
//	client := ws.NewClient("your_app_id", "your_app_secret",
//	    ws.WithEventHandler(handler),
//	)
func NewClient(appId, appSecret string, opts ...Option) *Client {
	c := &Client{
		appId:     appId,
		appSecret: appSecret,
		endpoint:  protocol.DefaultEndpoint,

		// 默认配置（指数退避重连策略）
		autoReconnect:         protocol.DefaultAutoReconnect,
		reconnectBaseInterval: protocol.DefaultReconnectBaseInterval,
		reconnectMaxInterval:  protocol.DefaultReconnectMaxInterval,
		reconnectMultiplier:   protocol.DefaultReconnectMultiplier,
		reconnectMaxRetry:     protocol.DefaultReconnectMaxRetry,
		reconnectJitter:       protocol.DefaultReconnectJitter,
		writeWait:             protocol.DefaultWriteWait,
		pongWait:              protocol.DefaultPongWait,
		ackMode:               protocol.DefaultAckMode,

		logLevel: core.LogLevelInfo,
		sendChan: make(chan []byte, 256),
		stopChan: make(chan struct{}),
	}

	// 应用配置选项
	for _, opt := range opts {
		opt(c)
	}

	// 如果没有设置 logger，使用默认 logger
	if c.logger == nil {
		c.logger = core.NewDefaultLogger(c.logLevel)
	}

	return c
}

// Start 启动 WebSocket 长连接
// 该方法会阻塞直到连接关闭或发生不可恢复的错误
//
// 使用 context 可以控制连接的生命周期：
//
//	ctx, cancel := context.WithCancel(context.Background())
//	go func() {
//	    time.Sleep(10 * time.Minute)
//	    cancel() // 10 分钟后关闭连接
//	}()
//	client.Start(ctx)
func (c *Client) Start(ctx context.Context) error {
	// 检查是否已关闭
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrClientClosed
	}
	c.mu.Unlock()

	// 检查是否设置了事件处理器
	if c.eventHandler == nil && c.dispatcher == nil {
		return ErrHandlerNotSet
	}

	// 建立连接
	if err := c.connect(ctx); err != nil {
		c.logger.Error(ctx, fmt.Sprintf("connect failed: %v", err))

		// 如果是客户端错误，不重试
		if _, ok := err.(*ClientError); ok {
			return err
		}

		// 断开连接并尝试重连
		c.disconnect(ctx)

		if c.autoReconnect {
			if err := c.reconnect(ctx); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// 启动消息接收循环
	c.receiveLoop(ctx)

	return nil
}

// Stop 停止 WebSocket 连接
func (c *Client) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	close(c.stopChan)

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

// IsConnected 检查是否已连接
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil && !c.closed
}

// connect 建立 WebSocket 连接
func (c *Client) connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return ErrAlreadyConnected
	}

	// 解析端点 URL
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %w", err)
	}

	// 确定签名用的 URI
	// 当通过 WithBaseUrl 设置时，使用固定的事件路径（避免反向代理路径前缀导致签名不一致）
	uri := c.signPath
	if uri == "" {
		uri = u.RequestURI()
		if uri == "" {
			uri = u.Path
			if u.RawQuery != "" {
				uri = uri + "?" + u.RawQuery
			}
		}
	}

	headers, err := kso.SignForWebSocket(c.appId, c.appSecret, uri)
	if err != nil {
		return fmt.Errorf("sign failed: %w", err)
	}

	// 如果启用 ACK 模式，添加协商 Header
	if c.ackMode {
		headers.Set("X-Ack-Mode", "required")
	}

	c.logger.Debug(ctx, fmt.Sprintf("connecting to %s, ack_mode: %v", c.endpoint, c.ackMode))

	// 建立 WebSocket 连接
	dialer := c.newDialer()
	conn, resp, err := dialer.DialContext(ctx, c.endpoint, headers)
	if err != nil {
		if resp != nil {
			return c.parseConnectError(resp)
		}
		return fmt.Errorf("dial failed: %w", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusSwitchingProtocols {
		return c.parseConnectError(resp)
	}

	c.conn = conn
	c.connURL = u
	c.receivedGoAway = false // 重置 GoAway 标志

	// 设置 Ping 处理器（服务端主动发送 Ping，客户端需要回复 Pong 并刷新超时）
	// 通过统一的 writeMessage 方法回复 Pong，确保与 ACK 等写操作串行化
	conn.SetPingHandler(func(appData string) error {
		c.logger.Debug(ctx, "received ping from server")
		// 刷新读超时
		if err := conn.SetReadDeadline(time.Now().Add(c.pongWait)); err != nil {
			c.logger.Error(ctx, fmt.Sprintf("set read deadline failed in ping handler: %v", err))
			return err
		}
		// 回复 Pong（通过 writeMessage 统一写入路径）
		if err := c.writeMessage(websocket.PongMessage, []byte(appData)); err != nil {
			c.logger.Error(ctx, fmt.Sprintf("write pong failed: %v", err))
			return err
		}
		return nil
	})

	c.logger.Info(ctx, fmt.Sprintf("connected to %s", c.endpoint))

	return nil
}

// newDialer 创建 WebSocket 拨号器（拷贝默认配置，避免污染全局 Dialer）
func (c *Client) newDialer() *websocket.Dialer {
	dialer := *websocket.DefaultDialer
	if !c.tlsVerify {
		if dialer.TLSClientConfig == nil {
			dialer.TLSClientConfig = &tls.Config{}
		}
		dialer.TLSClientConfig.InsecureSkipVerify = true
	}
	return &dialer
}

// disconnect 断开连接
func (c *Client) disconnect(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return
	}

	_ = c.conn.Close()
	c.conn = nil
	c.connURL = nil

	c.logger.Info(ctx, "disconnected")
}

// reconnect 重连（使用指数退避策略）
func (c *Client) reconnect(ctx context.Context) error {
	retryCount := 0
	currentInterval := c.reconnectBaseInterval

	for {
		// 检查是否超过最大重试次数
		if c.reconnectMaxRetry >= 0 && retryCount >= c.reconnectMaxRetry {
			return ErrReconnectExceeded
		}

		retryCount++

		// 计算带抖动的等待时间
		waitTime := c.calculateBackoffWithJitter(currentInterval)
		c.logger.Info(ctx, fmt.Sprintf("reconnecting in %v, attempt %d", waitTime, retryCount))

		// 等待重连间隔
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.stopChan:
			return ErrClientClosed
		case <-time.After(waitTime):
		}

		// 尝试连接
		err := c.connect(ctx)
		if err == nil {
			c.logger.Info(ctx, fmt.Sprintf("reconnected successfully after %d attempts", retryCount))
			return nil
		}

		// 如果是客户端错误，不继续重试
		if _, ok := err.(*ClientError); ok {
			return err
		}

		c.logger.Error(ctx, fmt.Sprintf("reconnect failed: %v", err))
		c.disconnect(ctx)

		// 计算下次重连间隔（指数增长，但不超过最大值）
		currentInterval = time.Duration(float64(currentInterval) * c.reconnectMultiplier)
		if currentInterval > c.reconnectMaxInterval {
			currentInterval = c.reconnectMaxInterval
		}
	}
}

// calculateBackoffWithJitter 计算带抖动的退避时间
// 使用 Full Jitter 策略：在 [0, interval] 范围内随机
func (c *Client) calculateBackoffWithJitter(interval time.Duration) time.Duration {
	if c.reconnectJitter <= 0 {
		return interval
	}

	// 计算抖动范围：interval * (1 - jitter) 到 interval * (1 + jitter)
	minInterval := float64(interval) * (1 - c.reconnectJitter)
	maxInterval := float64(interval) * (1 + c.reconnectJitter)

	// 在范围内随机
	jitteredInterval := minInterval + rand.Float64()*(maxInterval-minInterval)
	return time.Duration(jitteredInterval)
}

// receiveLoop 消息接收循环
func (c *Client) receiveLoop(ctx context.Context) {
	// 创建用于通知 writeLoop 退出的 channel
	writeDone := make(chan struct{})

	// 启动写入协程，串行化所有 WebSocket 写操作
	go c.writeLoop(ctx, writeDone)

	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(ctx, fmt.Sprintf("receive loop panic: %v\n%s", r, debug.Stack()))
		}

		// 通知 writeLoop 退出并等待其完成
		close(writeDone)

		c.disconnect(ctx)

		// 如果开启了自动重连且未关闭，尝试重连
		c.mu.Lock()
		shouldReconnect := c.autoReconnect && !c.closed
		c.mu.Unlock()

		if shouldReconnect {
			if err := c.reconnect(ctx); err != nil {
				c.logger.Error(ctx, fmt.Sprintf("reconnect failed: %v", err))
				return
			}
			// 重连成功，重新进入接收循环
			c.receiveLoop(ctx)
		}
	}()

	for {
		// 检查是否关闭
		select {
		case <-ctx.Done():
			return
		case <-c.stopChan:
			return
		default:
		}

		c.mu.Lock()
		conn := c.conn
		c.mu.Unlock()

		if conn == nil {
			c.logger.Error(ctx, "connection is nil, exiting receive loop")
			return
		}

		// 设置读取超时
		if err := conn.SetReadDeadline(time.Now().Add(c.pongWait)); err != nil {
			c.logger.Error(ctx, fmt.Sprintf("set read deadline failed: %v", err))
			return
		}

		// 读取消息
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			// 检查是否因为收到 GoAway 后服务端关闭了连接
			c.mu.Lock()
			goAwayReceived := c.receivedGoAway
			c.mu.Unlock()

			if goAwayReceived {
				c.logger.Debug(ctx, fmt.Sprintf("connection closed after goaway: %v", err))
			} else {
				c.logger.Error(ctx, fmt.Sprintf("read message failed: %v", err))
			}
			return
		}

		// 只处理文本消息
		if messageType != websocket.TextMessage {
			c.logger.Warn(ctx, fmt.Sprintf("received non-text message, type: %d", messageType))
			continue
		}

		// 异步处理消息
		go c.handleMessage(ctx, message)
	}
}

// writeLoop 写入循环，消费 sendChan 中的消息并通过 writeMessage 写入
// 与 PingHandler 共享 writeMu 锁，确保所有写操作串行化
func (c *Client) writeLoop(ctx context.Context, done <-chan struct{}) {
	for {
		select {
		case data := <-c.sendChan:
			if err := c.writeMessage(websocket.TextMessage, data); err != nil {
				c.logger.Error(ctx, fmt.Sprintf("[WebSocket] write message failed: %v", err))
			}
		case <-done:
			c.drainSendChan(ctx)
			return
		case <-c.stopChan:
			c.drainSendChan(ctx)
			return
		case <-ctx.Done():
			return
		}
	}
}

// drainSendChan 排空 sendChan 中的剩余消息
func (c *Client) drainSendChan(ctx context.Context) {
	for {
		select {
		case data := <-c.sendChan:
			if err := c.writeMessage(websocket.TextMessage, data); err != nil {
				c.logger.Debug(ctx, fmt.Sprintf("[WebSocket] drain write failed: %v", err))
				return
			}
		default:
			return
		}
	}
}

// handleMessage 处理消息
// 事件消息不包含 type 字段，goaway 消息包含 type="goaway"
func (c *Client) handleMessage(ctx context.Context, message []byte) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(ctx, fmt.Sprintf("handle message panic: %v\n%s", r, debug.Stack()))
		}
	}()

	// 解析消息基础字段
	var base struct {
		Type      string `json:"type"`
		Topic     string `json:"topic"`
		Operation string `json:"operation"`
	}
	if err := json.Unmarshal(message, &base); err != nil {
		c.logger.Error(ctx, fmt.Sprintf("unmarshal message failed: %v", err))
		return
	}

	// goaway 消息包含 type="goaway"
	if base.Type == protocol.MessageTypeGoAway {
		c.handleGoAwayMessage(ctx, message)
		return
	}

	// 事件消息需要验证 topic 和 operation 不能为空
	if base.Topic == "" || base.Operation == "" {
		c.logger.Error(ctx, fmt.Sprintf("invalid event message: topic or operation is empty, message=%s", string(message)))
		return
	}

	c.handleEventMessage(ctx, message)
}

// handleEventMessage 处理事件消息
func (c *Client) handleEventMessage(ctx context.Context, message []byte) {
	var msg protocol.EventMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		c.logger.Error(ctx, fmt.Sprintf("unmarshal event message failed: %v", err))
		return
	}

	// 生成事件编码
	eventCode := event.BuildEventCode(msg.Topic, msg.Operation)
	c.logger.Debug(ctx, fmt.Sprintf("received event: event_code=%s", eventCode))

	// 验证签名
	if err := kso.VerifySignature(&kso.VerifySignatureParams{
		AccessKey:     c.appId,
		SecretKey:     c.appSecret,
		Topic:         msg.Topic,
		Nonce:         msg.Nonce,
		Time:          msg.Time,
		EncryptedData: msg.EncryptedData,
		Signature:     msg.Signature,
	}); err != nil {
		c.logger.Error(ctx, fmt.Sprintf("verify signature failed: %v", err))
		return
	}

	// 解密数据
	decryptedData, err := kso.Decrypt(&kso.DecryptParams{
		SecretKey:     c.appSecret,
		EncryptedData: msg.EncryptedData,
		Nonce:         msg.Nonce,
	})
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("decrypt event data failed: %v", err))
		return
	}

	// 转换为 Event
	evt := event.NewEvent(msg.Topic, msg.Operation, msg.Time, decryptedData)

	// 调用处理器
	var handleErr error
	if c.dispatcher != nil {
		handleErr = c.dispatcher.Handle(ctx, evt)
	} else if c.eventHandler != nil {
		handleErr = c.eventHandler.Handle(ctx, evt)
	}

	// 如果启用 ACK 模式，发送 ACK
	if c.ackMode {
		c.sendAck(ctx, msg.Nonce, handleErr)
	}

	if handleErr != nil {
		c.logger.Error(ctx, fmt.Sprintf("handle event failed: %v", handleErr))
		return
	}

	c.logger.Debug(ctx, fmt.Sprintf("event handled: event_code=%s", eventCode))
}

// sendAck 发送 ACK 消息（ACK 模式下使用）
// 通过 sendChan 投递到 writeLoop 串行写入，避免并发写入 WebSocket 连接
func (c *Client) sendAck(ctx context.Context, nonce string, err error) {
	if nonce == "" {
		c.logger.Warn(ctx, "[WebSocket] ack mode enabled but event nonce is empty, skip sending ack")
		return
	}

	ack := &protocol.AckMessage{
		Type:  "ack",
		Nonce: nonce,
		Code:  200,
	}

	if err != nil {
		ack.Code = 500
		errMsg := err.Error()
		if len(errMsg) > 256 {
			errMsg = errMsg[:256] + "..."
		}
		ack.Msg = errMsg
	}

	data, marshalErr := json.Marshal(ack)
	if marshalErr != nil {
		c.logger.Error(ctx, fmt.Sprintf("[WebSocket] marshal ack failed: %v", marshalErr))
		return
	}

	select {
	case c.sendChan <- data:
		c.logger.Debug(ctx, fmt.Sprintf("[WebSocket] ack queued, nonce: %s, code: %d", nonce, ack.Code))
	default:
		c.logger.Error(ctx, fmt.Sprintf("[WebSocket] sendChan full, ack dropped, nonce: %s", nonce))
	}
}

// handleGoAwayMessage 处理 GoAway 消息
func (c *Client) handleGoAwayMessage(ctx context.Context, message []byte) {
	var msg protocol.GoAwayMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		c.logger.Error(ctx, fmt.Sprintf("unmarshal goaway message failed: %v", err))
		return
	}

	c.logger.Info(ctx, fmt.Sprintf("received goaway: reason=%s, message=%s, reconnect_ms=%d",
		msg.Reason, msg.Message, msg.ReconnectMs))

	// 标记已收到 GoAway
	c.mu.Lock()
	c.receivedGoAway = true

	// 如果是连接被替换，不重连
	if msg.Reason == protocol.GoAwayReasonConnectionReplaced {
		c.autoReconnect = false
		c.mu.Unlock()
		c.logger.Warn(ctx, "connection replaced by another client, will not reconnect")
		return
	}
	c.mu.Unlock()

	// 如果服务端建议了重连时间，使用该时间作为基础间隔
	if msg.ReconnectMs > 0 {
		c.mu.Lock()
		c.reconnectBaseInterval = time.Duration(msg.ReconnectMs) * time.Millisecond
		c.mu.Unlock()
	}
}

// writeMessage 统一的 WebSocket 写操作方法
// 所有写操作（ACK、Pong 等）都必须通过此方法，确保写入串行化
func (c *Client) writeMessage(messageType int, data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	if err := conn.SetWriteDeadline(time.Now().Add(c.writeWait)); err != nil {
		return err
	}
	return conn.WriteMessage(messageType, data)
}

// parseConnectError 解析连接错误
func (c *Client) parseConnectError(resp *http.Response) error {
	// 根据状态码判断错误类型
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return NewClientError(resp.StatusCode, "authentication failed")
	case http.StatusForbidden:
		return NewClientError(resp.StatusCode, "forbidden")
	case http.StatusTooManyRequests:
		return NewServerError(resp.StatusCode, "too many connections")
	default:
		return NewServerError(resp.StatusCode, fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
	}
}
