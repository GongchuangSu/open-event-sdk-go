package event

import (
	"context"
	"testing"

	"github.com/GongchuangSu/open-event-sdk-go/event/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypedEvent_ParseData(t *testing.T) {
	tests := []struct {
		name      string
		eventData string
		wantErr   bool
		validate  func(t *testing.T, data model.V7NotificationAppChatMessageCreateData)
	}{
		{
			name: "valid app chat message create data",
			eventData: `{
				"company_id": "comp_123",
				"chat": {
					"id": "chat_456",
					"type": "single"
				},
				"sender": {
					"id": "user_789",
					"type": "user",
					"name": "张三"
				},
				"send_time": 1706150400,
				"message": {
					"id": "msg_001",
					"type": "text",
					"content": {
						"text": {
							"content": "Hello World"
						}
					}
				}
			}`,
			wantErr: false,
			validate: func(t *testing.T, data model.V7NotificationAppChatMessageCreateData) {
				assert.Equal(t, "comp_123", data.CompanyId)
				assert.Equal(t, "chat_456", data.Chat.Id)
				assert.Equal(t, "single", data.Chat.Type)
				assert.Equal(t, "user_789", data.Sender.Id)
				assert.Equal(t, "user", data.Sender.Type)
				assert.Equal(t, "张三", data.Sender.Name)
				assert.Equal(t, int64(1706150400), data.SendTime)
				assert.Equal(t, "msg_001", data.Message.Id)
				assert.Equal(t, "text", data.Message.Type)
			},
		},
		{
			name: "data with mentions",
			eventData: `{
				"company_id": "comp_123",
				"chat": {"id": "chat_456", "type": "group"},
				"sender": {"id": "user_789", "type": "user"},
				"send_time": 1706150400,
				"message": {
					"id": "msg_001",
					"type": "text",
					"content": {},
					"mentions": [
						{"id": "0", "type": "user", "identity": {"id": "user_001", "type": "user"}}
					]
				}
			}`,
			wantErr: false,
			validate: func(t *testing.T, data model.V7NotificationAppChatMessageCreateData) {
				assert.Len(t, data.Message.Mentions, 1)
				assert.Equal(t, "0", data.Message.Mentions[0].Id)
				assert.Equal(t, "user", data.Message.Mentions[0].Type)
				require.NotNil(t, data.Message.Mentions[0].Identity)
				assert.Equal(t, "user_001", data.Message.Mentions[0].Identity.Id)
			},
		},
		{
			name:      "invalid json",
			eventData: `{invalid json}`,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &Event{
				Topic:     "kso.app_chat.message",
				Operation: "create",
				Time:      1706150400,
				Data:      tt.eventData,
			}

			var handlerCalled bool
			var capturedData model.V7NotificationAppChatMessageCreateData

			handler := wrapTypedHandler(func(ctx context.Context, e *TypedEvent[model.V7NotificationAppChatMessageCreateData]) error {
				handlerCalled = true
				capturedData = e.Data
				return nil
			})

			err := handler.Handle(context.Background(), event)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, handlerCalled)
			if tt.validate != nil {
				tt.validate(t, capturedData)
			}
		})
	}
}

func TestDispatcher_OnV7AppChatMessageCreate(t *testing.T) {
	dispatcher := NewDispatcher()

	var receivedEvent *V7AppChatMessageCreateEvent

	dispatcher.OnV7AppChatMessageCreate(func(ctx context.Context, e *V7AppChatMessageCreateEvent) error {
		receivedEvent = e
		return nil
	})

	// 验证处理器已注册
	assert.True(t, dispatcher.HasHandler(model.EventCodeAppChatMessageCreate))

	// 模拟事件
	event := &Event{
		Topic:     "kso.app_chat.message",
		Operation: "create",
		Time:      1706150400,
		Data: `{
			"company_id": "comp_123",
			"chat": {"id": "chat_456", "type": "single"},
			"sender": {"id": "user_789", "type": "user"},
			"send_time": 1706150400,
			"message": {"id": "msg_001", "type": "text", "content": {}}
		}`,
	}

	err := dispatcher.Dispatch(context.Background(), event)
	require.NoError(t, err)

	require.NotNil(t, receivedEvent)
	assert.Equal(t, "comp_123", receivedEvent.Data.CompanyId)
	assert.Equal(t, "chat_456", receivedEvent.Data.Chat.Id)
}

func TestDispatcher_ChainedRegistration(t *testing.T) {
	dispatcher := NewDispatcher()

	// 测试链式调用
	dispatcher.
		OnV7AppChatMessageCreate(func(ctx context.Context, e *V7AppChatMessageCreateEvent) error {
			return nil
		}).
		OnV7AppChatCreate(func(ctx context.Context, e *V7AppChatCreateEvent) error {
			return nil
		}).
		OnV7AppGroupChatDelete(func(ctx context.Context, e *V7AppGroupChatDeleteEvent) error {
			return nil
		}).
		OnV7AppGroupChatMemberUserCreate(func(ctx context.Context, e *V7AppGroupChatMemberUserCreateEvent) error {
			return nil
		}).
		OnV7AppGroupChatMemberUserDelete(func(ctx context.Context, e *V7AppGroupChatMemberUserDeleteEvent) error {
			return nil
		}).
		OnV7AppGroupChatMemberRobotCreate(func(ctx context.Context, e *V7AppGroupChatMemberRobotCreateEvent) error {
			return nil
		}).
		OnV7AppGroupChatMemberRobotDelete(func(ctx context.Context, e *V7AppGroupChatMemberRobotDeleteEvent) error {
			return nil
		})

	// 验证所有处理器都已注册
	assert.True(t, dispatcher.HasHandler(model.EventCodeAppChatMessageCreate))
	assert.True(t, dispatcher.HasHandler(model.EventCodeAppChatCreate))
	assert.True(t, dispatcher.HasHandler(model.EventCodeAppGroupChatDelete))
	assert.True(t, dispatcher.HasHandler(model.EventCodeAppGroupChatMemberUserCreate))
	assert.True(t, dispatcher.HasHandler(model.EventCodeAppGroupChatMemberUserDelete))
	assert.True(t, dispatcher.HasHandler(model.EventCodeAppGroupChatMemberRobotCreate))
	assert.True(t, dispatcher.HasHandler(model.EventCodeAppGroupChatMemberRobotDelete))
}

func TestDispatcher_OnV7AppGroupChatMemberUserCreate(t *testing.T) {
	dispatcher := NewDispatcher()

	var receivedEvent *V7AppGroupChatMemberUserCreateEvent

	dispatcher.OnV7AppGroupChatMemberUserCreate(func(ctx context.Context, e *V7AppGroupChatMemberUserCreateEvent) error {
		receivedEvent = e
		return nil
	})

	event := &Event{
		Topic:     "kso.xz.app.group_chat.member.user",
		Operation: "create",
		Time:      1706150400,
		Data: `{
			"chat_id": "chat_123",
			"company_id": "comp_456",
			"operator": {"id": "user_admin", "type": "user"},
			"users": [
				{"id": "user_001", "type": "user"},
				{"id": "user_002", "type": "user"}
			]
		}`,
	}

	err := dispatcher.Dispatch(context.Background(), event)
	require.NoError(t, err)

	require.NotNil(t, receivedEvent)
	assert.Equal(t, "chat_123", receivedEvent.Data.ChatId)
	assert.Equal(t, "comp_456", receivedEvent.Data.CompanyId)
	assert.Equal(t, "user_admin", receivedEvent.Data.Operator.Id)
	assert.Len(t, receivedEvent.Data.Users, 2)
}

func TestTypedEvent_OriginalEventAccess(t *testing.T) {
	dispatcher := NewDispatcher()

	var originalTopic string
	var originalOperation string

	dispatcher.OnV7AppChatMessageCreate(func(ctx context.Context, e *V7AppChatMessageCreateEvent) error {
		// 验证可以访问原始事件数据
		originalTopic = e.Topic
		originalOperation = e.Operation
		return nil
	})

	event := &Event{
		Topic:     "kso.app_chat.message",
		Operation: "create",
		Time:      1706150400,
		Data:      `{"company_id": "comp_123", "chat": {"id": "chat_456", "type": "single"}, "sender": {"id": "user_789", "type": "user"}, "send_time": 1706150400, "message": {"id": "msg_001", "type": "text", "content": {}}}`,
	}

	err := dispatcher.Dispatch(context.Background(), event)
	require.NoError(t, err)

	assert.Equal(t, "kso.app_chat.message", originalTopic)
	assert.Equal(t, "create", originalOperation)
}
