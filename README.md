# open-event-sdk-go

开放平台事件订阅 SDK（Go 语言版），支持通过 WebSocket 长连接接收和处理事件。

## 特性

- **WebSocket 长连接**：无需公网 IP，延迟更低、实时性更好
- **自动重连**：网络断开时自动重连，支持指数退避策略
- **KSO-1 签名认证**：安全的认证机制
- **灵活的事件处理**：支持单一 Handler 和 Dispatcher 分发两种模式
- **开箱即用**：内置默认配置，无需额外设置即可使用
- **CLI 工具**：零代码接收事件，终端实时查看，适合快速验证和调试

## CLI 工具 (openevent)

`openevent` 是基于本 SDK 的命令行工具，无需编写代码即可接收事件推送，适合快速验证事件配置。

### 安装

**方式一：一键安装脚本**（推荐，自动检测平台）

```bash
curl -fsSL https://raw.githubusercontent.com/GongchuangSu/open-event-sdk-go/main/scripts/install.sh | bash
```

指定版本或安装目录：

```bash
VERSION=1.1.0 INSTALL_DIR=~/.local/bin bash -c "$(curl -fsSL https://raw.githubusercontent.com/GongchuangSu/open-event-sdk-go/main/scripts/install.sh)"
```

**方式二：go install**

```bash
go install github.com/GongchuangSu/open-event-sdk-go/cmd/openevent@latest
```

**方式三：从源码构建**

```bash
git clone https://github.com/GongchuangSu/open-event-sdk-go.git
cd open-event-sdk-go
make build            # 产物在 output/openevent
make install          # 安装到 $GOPATH/bin
```

**方式四：从 [GitHub Releases](https://github.com/GongchuangSu/open-event-sdk-go/releases) 手动下载**

### 使用

**凭证提供方式**（按优先级从高到低）：

| 优先级 | 方式 | 示例 |
|--------|------|------|
| 1 | 命令行参数 | `--app-id xxx --app-secret yyy` |
| 2 | 环境变量 | `APP_ID=xxx APP_SECRET=yyy` |
| 3 | 交互式输入 | 自动检测 TTY，密码隐藏显示 |

```bash
# 使用命令行参数
openevent listen --app-id YOUR_APP_ID --app-secret YOUR_APP_SECRET

# 使用环境变量
export APP_ID=your_app_id
export APP_SECRET=your_app_secret
openevent listen

# 交互式输入（仅 TTY 终端，密码不回显）
openevent listen

# 过滤特定事件
openevent listen --events "kso.app_chat.message.create,kso.app_chat.create"

# JSON 输出（适合管道和日志采集）
openevent listen --json | jq .

# 同时输出到文件
openevent listen --output events.log

# 查看 SDK 内部调试日志
openevent listen --verbose
```

### 命令参考

```
openevent listen [flags]
openevent version
openevent help
```

**listen 选项：**

| 选项 | 缩写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--app-id` | | string | | 应用 ID |
| `--app-secret` | | string | | 应用密钥 |
| `--events` | `-e` | string | | 事件类型过滤（逗号分隔，精确匹配） |
| `--json` | | bool | false | 以 NDJSON 格式输出（每行一个 JSON 对象） |
| `--output` | `-o` | string | | 同时输出到文件（create/truncate 模式） |
| `--verbose` | `-v` | bool | false | 显示详细日志（DEBUG 级别，输出到 stderr） |
| `--no-color` | | bool | false | 禁用彩色输出 |
| `--no-ack` | | bool | false | 禁用 ACK 模式（默认开启，失败时服务端重试） |
| `--endpoint` | | string | | 自定义 WebSocket 端点 |

**退出码：**

| 退出码 | 含义 | 场景 |
|--------|------|------|
| 0 | 正常退出 | Ctrl+C / SIGTERM 优雅关闭 |
| 1 | 凭证错误 | 缺少 APP_ID/APP_SECRET、认证失败 |
| 2 | 连接失败 | 重连次数耗尽、服务端不可达 |
| 3 | 运行时错误 | 文件打开失败、其他未预期错误 |

### 本地开发

项目提供 `Makefile` 管理构建和测试流程：

```bash
make help             # 查看所有可用目标
make build            # 构建 CLI 到 output/openevent（自动注入版本信息）
make test             # 运行全部单元测试
make test-race        # 运行测试（含竞态检测）
make test-cover       # 运行测试并生成覆盖率报告
make lint             # 代码静态分析（go vet）
make run ARGS="listen --help"  # go run 直接运行
make clean            # 清理构建产物
make install          # 安装到 $GOPATH/bin
make                  # 默认: lint + test + build
```

构建产物会通过 `-ldflags` 自动注入版本信息，`openevent version` 将输出：

```
openevent v1.0.2 (125f725, 2026-03-26T03:01:35Z)
```

---

## SDK 使用

### 安装

```bash
go get github.com/GongchuangSu/open-event-sdk-go@v1.0.1
```

## 快速开始

```go
package main

import (
    "context"
    "log"

    openevent "github.com/GongchuangSu/open-event-sdk-go"
)

func main() {
    client := openevent.NewClient("your_app_id", "your_app_secret",
        openevent.WithEventHandlerFunc(func(ctx context.Context, e *openevent.Event) error {
            log.Printf("收到事件: event_code=%s", e.EventCode())
            log.Printf("事件数据: %s", e.Data)
            return nil
        }),
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

### 类型化事件处理

目前部分事件已支持 `OnV7XXX` 方法，可使用链式调用注册类型化处理器，事件数据会自动解析为对应的结构体。其他事件请使用 `RegisterFunc` 方法处理。

**已支持 OnV7XXX 方法的事件：**

| 方法 | 事件编码 | 说明 |
|------|---------|------|
| `OnV7AppChatMessageCreate` | `kso.app_chat.message.create` | 用户给应用发送消息 |
| `OnV7AppChatCreate` | `kso.app_chat.create` | 首次创建用户和机器人的会话 |
| `OnV7AppGroupChatDelete` | `kso.xz.app.group_chat.delete` | 群聊解散 |
| `OnV7AppGroupChatMemberUserCreate` | `kso.xz.app.group_chat.member.user.create` | 用户进群 |
| `OnV7AppGroupChatMemberUserDelete` | `kso.xz.app.group_chat.member.user.delete` | 用户退群 |
| `OnV7AppGroupChatMemberRobotCreate` | `kso.xz.app.group_chat.member.robot.create` | 机器人进群 |
| `OnV7AppGroupChatMemberRobotDelete` | `kso.xz.app.group_chat.member.robot.delete` | 机器人退群 |

**组合使用示例：**

```go
package main

import (
    "context"
    "encoding/json"
    "log"

    openevent "github.com/GongchuangSu/open-event-sdk-go"
)

func main() {
    dispatcher := openevent.NewDispatcher()

    // ========== 方式一: OnV7XXX 方法（类型安全，推荐） ==========
    dispatcher.
        OnV7AppChatMessageCreate(func(ctx context.Context, e *openevent.V7AppChatMessageCreateEvent) error {
            log.Printf("收到消息: chat_id=%s, sender=%s", e.Data.Chat.Id, e.Data.Sender.Id)
            return nil
        }).
        OnV7AppChatCreate(func(ctx context.Context, e *openevent.V7AppChatCreateEvent) error {
            log.Printf("会话创建: chat_id=%s", e.Data.ChatId)
            return nil
        }).
        OnV7AppGroupChatMemberUserCreate(func(ctx context.Context, e *openevent.V7AppGroupChatMemberUserCreateEvent) error {
            log.Printf("用户进群: chat_id=%s", e.Data.ChatId)
            return nil
        })

    // ========== 方式二: RegisterFunc（处理其他事件，需自行解析 Data） ==========
    dispatcher.RegisterFunc("kso.user.status.update", func(ctx context.Context, e *openevent.Event) error {
        log.Printf("用户状态变更: %s", e.EventCode())
        var data map[string]any
        json.Unmarshal([]byte(e.Data), &data)
        log.Printf("数据: %+v", data)
        return nil
    })

    // ========== 兜底处理器 ==========
    dispatcher.RegisterFallbackFunc(func(ctx context.Context, e *openevent.Event) error {
        log.Printf("未处理的事件: %s", e.EventCode())
        return nil
    })

    client := openevent.NewClient("your_app_id", "your_app_secret",
        openevent.WithDispatcher(dispatcher),
    )

    if err := client.Start(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

**使用事件数据模型：**

如需直接使用事件数据模型，请导入 `model` 包：

```go
import "github.com/GongchuangSu/open-event-sdk-go/event/model"

var sender model.V7Identity
var data model.V7NotificationAppChatMessageCreateData
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
├── Makefile                # 构建/测试/安装
├── .goreleaser.yaml        # 多平台发布配置
├── .github/workflows/      # CI/CD
│   ├── ci.yaml             # 测试 + lint（Go 1.21-1.23）
│   └── release.yaml        # tag → GoReleaser 发布
├── openevent.go            # 根包入口（推荐使用）
├── cmd/openevent/          # CLI 工具
│   ├── main.go             # 入口（版本注入点）
│   ├── cmd/                # Cobra 命令
│   │   ├── root.go         # 根命令 + 退出码
│   │   ├── listen.go       # listen 子命令
│   │   └── version.go      # version 子命令
│   └── internal/           # CLI 内部实现
│       ├── credential.go   # 凭证解析（flag → env → TTY）
│       ├── printer.go      # 事件输出（channel 序列化）
│       ├── filter.go       # 事件过滤
│       └── logger.go       # Stderr 日志
├── ws/                     # WebSocket 客户端
│   ├── client.go           # 客户端主逻辑
│   ├── option.go           # 配置选项
│   └── error.go            # 错误定义
├── event/                  # 事件处理
│   ├── dispatcher.go       # 事件分发器
│   ├── dispatcher_typed.go # OnXXX 类型化处理方法
│   ├── typed_event.go      # 类型化事件定义
│   ├── handler.go          # Handler 接口
│   ├── event.go            # 事件实体
│   └── model/              # 事件数据模型
├── core/                   # 核心公共组件
│   └── logger.go           # 日志接口
├── internal/               # 内部实现（不对外暴露）
│   ├── kso/                # KSO-1 签名和加解密
│   └── protocol/           # WebSocket 协议定义
├── scripts/                # 安装脚本
│   └── install.sh
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
