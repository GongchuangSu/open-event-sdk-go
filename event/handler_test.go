package event

import (
	"context"
	"errors"
	"testing"
)

func TestHandlerFunc(t *testing.T) {
	called := false
	var receivedEvent *Event

	// 创建函数类型的处理器
	handler := HandlerFunc(func(ctx context.Context, event *Event) error {
		called = true
		receivedEvent = event
		return nil
	})

	// 验证实现了 Handler 接口
	var _ Handler = handler

	// 调用处理器
	testEvent := NewEvent("test", "event", 1234567890, `{"key":"value"}`)

	err := handler.Handle(context.Background(), testEvent)

	if err != nil {
		t.Errorf("Handle() error = %v", err)
	}

	if !called {
		t.Error("expected handler to be called")
	}

	if receivedEvent != testEvent {
		t.Error("expected to receive the same event")
	}
}

func TestHandlerFunc_Error(t *testing.T) {
	expectedErr := errors.New("test error")

	handler := HandlerFunc(func(ctx context.Context, event *Event) error {
		return expectedErr
	})

	err := handler.Handle(context.Background(), &Event{})

	if err != expectedErr {
		t.Errorf("expected error = %v, got %v", expectedErr, err)
	}
}

func TestHandlerFunc_Context(t *testing.T) {
	type ctxKey string

	handler := HandlerFunc(func(ctx context.Context, event *Event) error {
		// 验证 context 正确传递
		val := ctx.Value(ctxKey("test_key"))
		if val != "test_value" {
			t.Errorf("expected context value = test_value, got %v", val)
		}
		return nil
	})

	ctx := context.WithValue(context.Background(), ctxKey("test_key"), "test_value")
	err := handler.Handle(ctx, &Event{})

	if err != nil {
		t.Errorf("Handle() error = %v", err)
	}
}
