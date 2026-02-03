// Package main 演示使用单一 Handler 接收和处理事件
package main

import (
	"context"
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

	// 创建事件处理器
	handler := event.HandlerFunc(func(ctx context.Context, e *event.Event) error {
		log.Printf("收到事件: event_code=%s, time=%d", e.EventCode(), e.Time)
		log.Printf("事件数据: %s", e.Data)

		// 在这里处理事件...
		// EventCode = Topic.Operation，如 "kso.app_chat.message.create"
		// Data 字段是解密后的 JSON 字符串，可以根据业务需要进行解析

		return nil
	})

	// 创建 WebSocket 客户端
	// 使用默认端点 wss://openapi.wps.cn/v7/event/ws
	client := ws.NewClient(appId, appSecret,
		ws.WithEventHandler(handler),
		ws.WithLogLevel(core.LogLevelDebug),
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
