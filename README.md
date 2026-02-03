# open-event-sdk-go

开放平台事件订阅 SDK（Go 语言版），支持通过 WebSocket 长连接接收和处理事件。

## 特性

- **WebSocket 长连接**：与 HTTP 回调相比，延迟更低、实时性更好
- **自动重连**：网络断开时自动重连，支持指数退避策略
- **KSO-1 签名认证**：安全的认证机制
- **灵活的事件处理**：支持单一 Handler 和 Dispatcher 分发两种模式
- **开箱即用**：内置默认配置，无需额外设置即可使用

## 安装

```bash
go get github.com/GongchuangSu/open-event-sdk-go
```

## 快速开始

### 方式一：简洁导入（推荐）

使用根包导入，API 更简洁：

```go
package main

import (
    "context"
    "log"

    openevent "github.com/GongchuangSu/open-event-sdk-go"
)

func main() {
    // 创建客户端，开箱即用
    client := openevent.NewClient("your_app_id", "your_app_secret",
        openevent.WithEventHandlerFunc(func(ctx context.Context, e *openevent.Event) error {
            log.Printf("收到事件: event_code=%s", e.EventCode())
            log.Printf("事件数据: %s", e.Data)
            return nil
        }),
    )

    // 启动长连接（阻塞）
    if err := client.Start(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

### 方式二：分包导入

适用于需要更细粒度控制的场景：

```go
package main

import (
    "context"
    "log"

    "github.com/GongchuangSu/open-event-sdk-go/core"
    "github.com/GongchuangSu/open-event-sdk-go/event"
    "github.com/GongchuangSu/open-event-sdk-go/ws"
)

func main() {
    handler := event.HandlerFunc(func(ctx context.Context, e *event.Event) error {
        log.Printf("收到事件: event_code=%s", e.EventCode())
        return nil
    })

    client := ws.NewClient("your_app_id", "your_app_secret",
        ws.WithEventHandler(handler),
        ws.WithLogLevel(core.LogLevelDebug),
    )

    if err := client.Start(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

### Dispatcher 分发模式

按事件编码（event_code）分别处理不同事件：

```go
package main

import (
    "context"
    "log"

    openevent "github.com/GongchuangSu/open-event-sdk-go"
)

func main() {
    // 创建分发器
    dispatcher := openevent.NewDispatcher()

    // 注册不同事件编码的处理器
    // 事件编码 = topic.operation，如 "kso.app_chat.message.create"
    dispatcher.RegisterFunc("kso.app_chat.message.create", func(ctx context.Context, e *openevent.Event) error {
        log.Printf("处理聊天消息事件: %s", e.Data)
        return nil
    })

    dispatcher.RegisterFunc("kso.user.status.update", func(ctx context.Context, e *openevent.Event) error {
        log.Printf("处理用户状态变更事件: %s", e.Data)
        return nil
    })

    // 注册兜底处理器
    dispatcher.RegisterFallbackFunc(func(ctx context.Context, e *openevent.Event) error {
        log.Printf("未知事件: event_code=%s", e.EventCode())
        return nil
    })

    // 创建客户端
    client := openevent.NewClient("your_app_id", "your_app_secret",
        openevent.WithDispatcher(dispatcher),
    )

    if err := client.Start(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

## 配置选项

### 基础配置

```go
client := openevent.NewClient(appId, appSecret,
    // 自定义 WebSocket 端点（可选，默认 wss://openapi.wps.cn/v7/event/ws）
    openevent.WithEndpoint("wss://custom-endpoint.com/event/ws"),

    // 设置日志级别
    openevent.WithLogLevel(openevent.LogLevelDebug),

    // 使用自定义日志
    openevent.WithLogger(customLogger),
)
```

### 重连配置（指数退避策略）

SDK 采用指数退避（Exponential Backoff）策略进行重连，避免在网络恢复时产生惊群效应。

**重连间隔计算公式**：`interval = min(baseInterval * multiplier^(retryCount-1), maxInterval) * (1 ± jitter)`

```go
client := openevent.NewClient(appId, appSecret,
    // 开启/关闭自动重连（默认开启）
    openevent.WithAutoReconnect(true),

    // 重连基础间隔（默认 1 秒）
    openevent.WithReconnectBaseInterval(1 * time.Second),

    // 重连最大间隔（默认 60 秒）
    openevent.WithReconnectMaxInterval(60 * time.Second),

    // 重连间隔倍数（默认 2.0）
    openevent.WithReconnectMultiplier(2.0),

    // 最大重试次数（-1 表示无限重试，默认 -1）
    openevent.WithReconnectMaxRetry(10),

    // 重连抖动系数（默认 0.2，表示 ±20% 随机抖动）
    openevent.WithReconnectJitter(0.2),
)
```

**默认重连时间序列示例**（baseInterval=1s, multiplier=2, maxInterval=60s, jitter=0.2）：

| 重试次数 | 基础间隔 | 实际间隔范围 |
|---------|---------|-------------|
| 1 | 1s | 0.8s ~ 1.2s |
| 2 | 2s | 1.6s ~ 2.4s |
| 3 | 4s | 3.2s ~ 4.8s |
| 4 | 8s | 6.4s ~ 9.6s |
| 5 | 16s | 12.8s ~ 19.2s |
| 6 | 32s | 25.6s ~ 38.4s |
| 7+ | 60s | 48s ~ 72s |

### 超时配置

```go
client := openevent.NewClient(appId, appSecret,
    // 写操作超时（默认 10 秒）
    openevent.WithWriteWait(10 * time.Second),

    // Pong 等待超时（默认 90 秒）
    openevent.WithPongWait(90 * time.Second),
)
```

## 事件结构

### 原始事件消息（加密）

SDK 接收到的原始消息包含以下字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `topic` | string | 消息主题（根据不同事件而定） |
| `operation` | string | 消息变更动作（根据不同事件而定） |
| `time` | int64 | 时间（秒为单位的时间戳） |
| `nonce` | string | iv 向量（解密时使用） |
| `signature` | string | 消息签名 |
| `encrypted_data` | string | 加密的消息数据 |

### 解密后的事件结构

SDK 会自动验证签名并解密数据，处理器接收到的是解密后的事件：

```go
type Event struct {
    Topic     string `json:"topic"`      // 消息主题
    Operation string `json:"operation"`  // 变更动作
    Time      int64  `json:"time"`       // 时间戳（秒）
    Data      string `json:"data"`       // 解密后的事件数据（JSON 字符串）
}

// 获取事件编码
func (e *Event) EventCode() string
```

**事件编码（event_code）说明**：
- 事件编码 = `topic` + `.` + `operation`，全局唯一
- 例如：`topic="kso.app_chat.message"`, `operation="create"` → `event_code="kso.app_chat.message.create"`
- 通过 `e.EventCode()` 方法动态获取事件编码
- Dispatcher 按事件编码进行事件分发

### 签名验证

签名计算方式：
1. 构建签名原文：`content = access_key:topic:nonce:time:encrypted_data`
2. 计算签名：`signature = HMAC-SHA256(content, secret_key)`
3. 签名使用 URL 安全的无填充 base64 编码

### 数据解密

解密方式：
1. `encrypted_data` 使用标准的有填充 base64 编码
2. 密钥 `cipher = MD5(secret_key)`
3. 使用 AES-CBC 模式解密，iv 为 nonce 的前 16 字节
4. 解密后移除 PKCS7 填充

## 事件处理

### 处理成功

```go
handler := openevent.HandlerFunc(func(ctx context.Context, e *openevent.Event) error {
    log.Printf("处理事件: event_code=%s", e.EventCode())
    
    // 解析事件数据
    var data map[string]interface{}
    if err := json.Unmarshal([]byte(e.Data), &data); err != nil {
        return err
    }
    
    // 处理业务逻辑...
    return nil
})
```

### 处理失败

返回 `error` 表示处理失败：

```go
handler := openevent.HandlerFunc(func(ctx context.Context, e *openevent.Event) error {
    if err := processEvent(e); err != nil {
        return err // 处理失败
    }
    return nil
})
```

## 优雅关闭

使用 `context` 控制连接生命周期：

```go
ctx, cancel := context.WithCancel(context.Background())

// 监听退出信号
go func() {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    cancel()       // 取消 context
    client.Stop()  // 停止客户端
}()

client.Start(ctx)
```

## 自定义日志

实现 `Logger` 接口：

```go
type Logger interface {
    Debug(ctx context.Context, args ...interface{})
    Info(ctx context.Context, args ...interface{})
    Warn(ctx context.Context, args ...interface{})
    Error(ctx context.Context, args ...interface{})
}
```

示例：

```go
type MyLogger struct{}

func (l *MyLogger) Debug(ctx context.Context, args ...interface{}) {
    // 自定义实现
}

// ... 其他方法

client := openevent.NewClient(appId, appSecret,
    openevent.WithLogger(&MyLogger{}),
)
```

## 目录结构

```
open-event-sdk-go/
├── openevent.go            # 根包入口（推荐使用）
├── ws/                     # WebSocket 客户端
│   ├── client.go           # 客户端主逻辑
│   ├── option.go           # 配置选项
│   └── error.go            # 错误定义
├── event/                  # 事件处理
│   ├── dispatcher.go       # 事件分发器
│   ├── handler.go          # Handler 接口
│   └── event.go            # 事件实体
├── core/                   # 核心公共组件
│   └── logger.go           # 日志接口
├── internal/               # 内部实现（不对外暴露）
│   ├── kso/                # KSO-1 签名和加解密
│   └── protocol/           # WebSocket 协议定义
└── examples/               # 使用示例
    ├── simple/             # 简单示例
    └── dispatcher/         # Dispatcher 模式示例
```

## 协议说明

### 消息类型

服务端通过 WebSocket 向客户端推送两种类型的消息：

#### 1. 事件消息

携带加密的事件数据，SDK 会自动验签和解密。

```json
{
    "topic": "kso.app_chat.message",
    "operation": "create",
    "time": 1704067200,
    "nonce": "b38c06ba3330d2a3",
    "signature": "EzpccL5eAnDbOH2qtZK3fHNBaO0UV3xvYvhbWLp1wuQ",
    "encrypted_data": "oadgi2+nWGZal2EfCSlxbLrr2Aog..."
}
```

#### 2. 关闭通知（goaway）

服务端主动关闭连接时发送，包含 `type: "goaway"` 和关闭原因：

```json
{
    "type": "goaway",
    "reason": "server_shutdown",
    "message": "服务器维护中",
    "reconnect_ms": 5000
}
```

**GoAway 原因类型**：

| 原因 | 说明 | 是否重连 |
|------|------|---------|
| `server_shutdown` | 服务器关闭（如维护升级） | 是，按 `reconnect_ms` 延迟重连 |
| `connection_replaced` | 连接被新连接替换（同一应用重复连接） | 否 |
| `heartbeat_timeout` | 心跳超时 | 是 |

### 心跳机制

- 服务端每 30 秒发送 WebSocket Ping
- 客户端自动回复 Pong
- 90 秒内未收到 Pong 则断开连接

### 认证方式

使用 KSO-1 签名认证，需要在 WebSocket 握手时携带以下 HTTP 头：

- `X-Kso-Date`: 请求时间（RFC1123 格式）
- `X-Kso-Authorization`: 签名（格式：`KSO-1 {app_id}:{signature}`）

## 许可证

MIT License
