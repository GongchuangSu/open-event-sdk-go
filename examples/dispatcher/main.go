// Package main 演示使用 Dispatcher 处理事件
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	openevent "github.com/GongchuangSu/open-event-sdk-go"
)

func main() {
	appId := os.Getenv("APP_ID")
	appSecret := os.Getenv("APP_SECRET")

	if appId == "" || appSecret == "" {
		log.Fatal("请设置 APP_ID 和 APP_SECRET 环境变量")
	}

	// 创建分发器
	dispatcher := openevent.NewDispatcher()

	// 使用 OnV7XXX 方法处理已支持的事件（类型安全）
	dispatcher.
		OnV7AppChatMessageCreate(func(ctx context.Context, e *openevent.V7AppChatMessageCreateEvent) error {
			log.Printf("[消息] chat=%s, sender=%s, msg=%s", e.Data.Chat.Id, e.Data.Sender.Id, e.Data.Message.Id)
			return nil
		}).
		OnV7AppChatCreate(func(ctx context.Context, e *openevent.V7AppChatCreateEvent) error {
			log.Printf("[会话创建] chat=%s", e.Data.ChatId)
			return nil
		}).
		OnV7AppGroupChatMemberUserCreate(func(ctx context.Context, e *openevent.V7AppGroupChatMemberUserCreateEvent) error {
			log.Printf("[用户进群] chat=%s", e.Data.ChatId)
			return nil
		})

	// 使用 RegisterFunc 处理其他事件（需自行解析 Data）
	dispatcher.RegisterFunc("kso.user.status.update", func(ctx context.Context, e *openevent.Event) error {
		var data map[string]any
		json.Unmarshal([]byte(e.Data), &data)
		log.Printf("[用户状态] %+v", data)
		return nil
	})

	// 兜底处理器
	dispatcher.RegisterFallbackFunc(func(ctx context.Context, e *openevent.Event) error {
		log.Printf("[未处理] event_code=%s", e.EventCode())
		return nil
	})

	client := openevent.NewClient(appId, appSecret,
		openevent.WithDispatcher(dispatcher),
		openevent.WithLogLevel(openevent.LogLevelInfo),
	)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		cancel()
		client.Stop()
	}()

	// 启动长连接
	log.Println("正在连接到 WebSocket 服务...")
	if err := client.Start(ctx); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
