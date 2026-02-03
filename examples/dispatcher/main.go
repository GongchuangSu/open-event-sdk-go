// Package main 演示使用 Dispatcher 按事件编码分发处理
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/GongchuangSu/open-event-sdk-go/core"
	"github.com/GongchuangSu/open-event-sdk-go/event"
	"github.com/GongchuangSu/open-event-sdk-go/ws"
)

func main() {
	// 从环境变量获取配置
	appId := os.Getenv("APP_ID")
	appSecret := os.Getenv("APP_SECRET")

	if appId == "" || appSecret == "" {
		log.Fatal("请设置 APP_ID 和 APP_SECRET 环境变量")
	}

	// 创建事件分发器
	dispatcher := event.NewDispatcher()

	// 注册聊天消息创建事件处理器
	// 事件编码 = topic.operation = "kso.app_chat.message.create"
	dispatcher.RegisterFunc("kso.app_chat.message.create", func(ctx context.Context, e *event.Event) error {
		log.Printf("[聊天消息] event_code=%s", e.EventCode())

		// 解析事件数据
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(e.Data), &data); err != nil {
			log.Printf("解析事件数据失败: %v", err)
			return err
		}

		log.Printf("[聊天消息] 事件数据: %+v", data)
		return nil
	})

	// 注册用户状态变更事件处理器
	dispatcher.RegisterFunc("kso.user.status.update", func(ctx context.Context, e *event.Event) error {
		log.Printf("[用户状态变更] event_code=%s", e.EventCode())
		log.Printf("[用户状态变更] 事件数据: %s", e.Data)
		return nil
	})

	// 注册兜底处理器，处理未注册的事件编码
	dispatcher.RegisterFallbackFunc(func(ctx context.Context, e *event.Event) error {
		log.Printf("[未知事件] event_code=%s", e.EventCode())
		log.Printf("[未知事件] 事件数据: %s", e.Data)
		return nil
	})

	// 打印已注册的事件编码
	log.Printf("已注册的事件编码: %v", dispatcher.EventCodes())

	// 创建 WebSocket 客户端
	// 使用默认端点 wss://openapi.wps.cn/v7/event/ws
	client := ws.NewClient(appId, appSecret,
		ws.WithDispatcher(dispatcher),
		ws.WithLogLevel(core.LogLevelInfo),
	)

	// 创建可取消的 context
	ctx, cancel := context.WithCancel(context.Background())

	// 监听退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("收到退出信号，正在关闭...")
		cancel()
		client.Stop()
	}()

	// 启动长连接
	log.Println("正在连接到 WebSocket 服务...")
	if err := client.Start(ctx); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
