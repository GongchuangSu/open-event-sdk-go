package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/GongchuangSu/open-event-sdk-go/event/model"
)

// TypedEvent 类型化事件，包含原始事件和解析后的数据
type TypedEvent[T any] struct {
	// Event 原始事件
	*Event

	// Data 解析后的事件数据
	Data T
}

// TypedEventHandler 类型化事件处理器函数
type TypedEventHandler[T any] func(ctx context.Context, event *TypedEvent[T]) error

// wrapTypedHandler 将类型化处理器包装为通用处理器
func wrapTypedHandler[T any](fn TypedEventHandler[T]) HandlerFunc {
	return func(ctx context.Context, event *Event) error {
		var data T
		if err := json.Unmarshal([]byte(event.Data), &data); err != nil {
			return fmt.Errorf("failed to unmarshal event data: %w", err)
		}

		typedEvent := &TypedEvent[T]{
			Event: event,
			Data:  data,
		}

		return fn(ctx, typedEvent)
	}
}

// ================== 应用消息事件 ==================

// V7AppChatMessageCreateEvent 应用收到消息事件
type V7AppChatMessageCreateEvent = TypedEvent[model.V7NotificationAppChatMessageCreateData]

// V7AppChatMessageCreateHandler 应用收到消息事件处理器
type V7AppChatMessageCreateHandler = TypedEventHandler[model.V7NotificationAppChatMessageCreateData]

// V7AppChatCreateEvent 应用会话创建事件
type V7AppChatCreateEvent = TypedEvent[model.V7NotificationAppChatCreateData]

// V7AppChatCreateHandler 应用会话创建事件处理器
type V7AppChatCreateHandler = TypedEventHandler[model.V7NotificationAppChatCreateData]

// ================== 应用群聊事件 ==================

// V7AppGroupChatDeleteEvent 群聊解散事件
type V7AppGroupChatDeleteEvent = TypedEvent[model.V7NotificationAppGroupChatData]

// V7AppGroupChatDeleteHandler 群聊解散事件处理器
type V7AppGroupChatDeleteHandler = TypedEventHandler[model.V7NotificationAppGroupChatData]

// V7AppGroupChatMemberUserCreateEvent 用户进群事件
type V7AppGroupChatMemberUserCreateEvent = TypedEvent[model.V7NotificationAppGroupChatMemberUserData]

// V7AppGroupChatMemberUserCreateHandler 用户进群事件处理器
type V7AppGroupChatMemberUserCreateHandler = TypedEventHandler[model.V7NotificationAppGroupChatMemberUserData]

// V7AppGroupChatMemberUserDeleteEvent 用户退群事件
type V7AppGroupChatMemberUserDeleteEvent = TypedEvent[model.V7NotificationAppGroupChatMemberUserData]

// V7AppGroupChatMemberUserDeleteHandler 用户退群事件处理器
type V7AppGroupChatMemberUserDeleteHandler = TypedEventHandler[model.V7NotificationAppGroupChatMemberUserData]

// V7AppGroupChatMemberRobotCreateEvent 机器人进群事件
type V7AppGroupChatMemberRobotCreateEvent = TypedEvent[model.V7NotificationAppGroupChatMemberRobotData]

// V7AppGroupChatMemberRobotCreateHandler 机器人进群事件处理器
type V7AppGroupChatMemberRobotCreateHandler = TypedEventHandler[model.V7NotificationAppGroupChatMemberRobotData]

// V7AppGroupChatMemberRobotDeleteEvent 机器人退群事件
type V7AppGroupChatMemberRobotDeleteEvent = TypedEvent[model.V7NotificationAppGroupChatMemberRobotData]

// V7AppGroupChatMemberRobotDeleteHandler 机器人退群事件处理器
type V7AppGroupChatMemberRobotDeleteHandler = TypedEventHandler[model.V7NotificationAppGroupChatMemberRobotData]
