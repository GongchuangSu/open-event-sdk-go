package event

import (
	"context"
	"errors"
	"testing"
)

func TestDispatcher_Register(t *testing.T) {
	dispatcher := NewDispatcher()

	// 注册处理器
	handler := HandlerFunc(func(ctx context.Context, event *Event) error {
		return nil
	})

	dispatcher.Register("user.created", handler)

	// 验证已注册
	if !dispatcher.HasHandler("user.created") {
		t.Error("expected handler to be registered")
	}

	// 验证未注册的事件类型
	if dispatcher.HasHandler("user.deleted") {
		t.Error("expected handler not to be registered")
	}
}

func TestDispatcher_RegisterFunc(t *testing.T) {
	dispatcher := NewDispatcher()

	// 使用 RegisterFunc 便捷方法
	dispatcher.RegisterFunc("file.created", func(ctx context.Context, event *Event) error {
		return nil
	})

	if !dispatcher.HasHandler("file.created") {
		t.Error("expected handler to be registered")
	}
}

func TestDispatcher_EventCodes(t *testing.T) {
	dispatcher := NewDispatcher()

	dispatcher.RegisterFunc("kso.user.create", func(ctx context.Context, event *Event) error {
		return nil
	})
	dispatcher.RegisterFunc("kso.file.update", func(ctx context.Context, event *Event) error {
		return nil
	})

	codes := dispatcher.EventCodes()

	if len(codes) != 2 {
		t.Errorf("expected 2 event codes, got %d", len(codes))
	}

	// 检查是否包含预期的事件编码
	codesMap := make(map[string]bool)
	for _, code := range codes {
		codesMap[code] = true
	}

	if !codesMap["kso.user.create"] {
		t.Error("expected kso.user.create in event codes")
	}
	if !codesMap["kso.file.update"] {
		t.Error("expected kso.file.update in event codes")
	}
}

func TestDispatcher_Dispatch(t *testing.T) {
	dispatcher := NewDispatcher()

	// 记录处理的事件
	handledEvents := make(map[string]bool)

	dispatcher.RegisterFunc("user.created", func(ctx context.Context, event *Event) error {
		handledEvents["user.created"] = true
		return nil
	})

	dispatcher.RegisterFunc("file.updated", func(ctx context.Context, event *Event) error {
		handledEvents["file.updated"] = true
		return nil
	})

	// 测试分发 user.created (event_code = topic.operation)
	err := dispatcher.Dispatch(context.Background(), NewEvent("user", "created", 0, ""))

	if err != nil {
		t.Errorf("Dispatch() error = %v", err)
	}

	if !handledEvents["user.created"] {
		t.Error("expected user.created to be handled")
	}

	// 测试分发 file.updated
	err = dispatcher.Dispatch(context.Background(), NewEvent("file", "updated", 0, ""))

	if err != nil {
		t.Errorf("Dispatch() error = %v", err)
	}

	if !handledEvents["file.updated"] {
		t.Error("expected file.updated to be handled")
	}
}

func TestDispatcher_Dispatch_NoHandler(t *testing.T) {
	dispatcher := NewDispatcher()

	// 分发未注册的事件编码
	err := dispatcher.Dispatch(context.Background(), NewEvent("unknown", "event", 0, ""))

	if err == nil {
		t.Error("expected error for unregistered event_code")
	}
}

func TestDispatcher_Fallback(t *testing.T) {
	dispatcher := NewDispatcher()

	fallbackCalled := false

	// 注册兜底处理器
	dispatcher.RegisterFallbackFunc(func(ctx context.Context, event *Event) error {
		fallbackCalled = true
		return nil
	})

	// 分发未注册的事件编码
	err := dispatcher.Dispatch(context.Background(), NewEvent("unknown", "event", 0, ""))

	if err != nil {
		t.Errorf("Dispatch() error = %v", err)
	}

	if !fallbackCalled {
		t.Error("expected fallback handler to be called")
	}
}

func TestDispatcher_HandlerError(t *testing.T) {
	dispatcher := NewDispatcher()

	expectedErr := errors.New("handler error")

	dispatcher.RegisterFunc("user.created", func(ctx context.Context, event *Event) error {
		return expectedErr
	})

	err := dispatcher.Dispatch(context.Background(), NewEvent("user", "created", 0, ""))

	if err != expectedErr {
		t.Errorf("expected error = %v, got %v", expectedErr, err)
	}
}

func TestDispatcher_Handle(t *testing.T) {
	dispatcher := NewDispatcher()

	handled := false

	dispatcher.RegisterFunc("test.event", func(ctx context.Context, event *Event) error {
		handled = true
		return nil
	})

	// 使用 Handle 方法（实现 Handler 接口）
	err := dispatcher.Handle(context.Background(), NewEvent("test", "event", 0, ""))

	if err != nil {
		t.Errorf("Handle() error = %v", err)
	}

	if !handled {
		t.Error("expected event to be handled")
	}
}

func TestDispatcher_ChainedRegister(t *testing.T) {
	// 测试链式调用
	dispatcher := NewDispatcher().
		RegisterFunc("event.a", func(ctx context.Context, event *Event) error {
			return nil
		}).
		RegisterFunc("event.b", func(ctx context.Context, event *Event) error {
			return nil
		}).
		RegisterFallbackFunc(func(ctx context.Context, event *Event) error {
			return nil
		})

	if !dispatcher.HasHandler("event.a") {
		t.Error("expected event.a to be registered")
	}

	if !dispatcher.HasHandler("event.b") {
		t.Error("expected event.b to be registered")
	}
}
