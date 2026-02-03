package event

import (
	"github.com/GongchuangSu/open-event-sdk-go/event/model"
)

// ================== 应用消息事件 OnXXX 方法 ==================

// OnV7AppChatMessageCreate 注册应用收到消息事件处理器
// 事件编码: kso.app_chat.message.create
// 当用户在单聊/群聊中给应用发送消息时触发
func (d *Dispatcher) OnV7AppChatMessageCreate(handler V7AppChatMessageCreateHandler) *Dispatcher {
	return d.Register(model.EventCodeAppChatMessageCreate, wrapTypedHandler(handler))
}

// OnV7AppChatCreate 注册应用会话创建事件处理器
// 事件编码: kso.app_chat.create
// 当首次创建用户和机器人的会话时触发
func (d *Dispatcher) OnV7AppChatCreate(handler V7AppChatCreateHandler) *Dispatcher {
	return d.Register(model.EventCodeAppChatCreate, wrapTypedHandler(handler))
}

// ================== 应用群聊事件 OnXXX 方法 ==================

// OnV7AppGroupChatDelete 注册群聊解散事件处理器
// 事件编码: kso.xz.app.group_chat.delete
// 当群聊解散时触发
func (d *Dispatcher) OnV7AppGroupChatDelete(handler V7AppGroupChatDeleteHandler) *Dispatcher {
	return d.Register(model.EventCodeAppGroupChatDelete, wrapTypedHandler(handler))
}

// OnV7AppGroupChatMemberUserCreate 注册用户进群事件处理器
// 事件编码: kso.xz.app.group_chat.member.user.create
// 当用户加入群聊时触发
func (d *Dispatcher) OnV7AppGroupChatMemberUserCreate(handler V7AppGroupChatMemberUserCreateHandler) *Dispatcher {
	return d.Register(model.EventCodeAppGroupChatMemberUserCreate, wrapTypedHandler(handler))
}

// OnV7AppGroupChatMemberUserDelete 注册用户退群事件处理器
// 事件编码: kso.xz.app.group_chat.member.user.delete
// 当用户退出群聊时触发
func (d *Dispatcher) OnV7AppGroupChatMemberUserDelete(handler V7AppGroupChatMemberUserDeleteHandler) *Dispatcher {
	return d.Register(model.EventCodeAppGroupChatMemberUserDelete, wrapTypedHandler(handler))
}

// OnV7AppGroupChatMemberRobotCreate 注册机器人进群事件处理器
// 事件编码: kso.xz.app.group_chat.member.robot.create
// 当机器人加入群聊时触发
func (d *Dispatcher) OnV7AppGroupChatMemberRobotCreate(handler V7AppGroupChatMemberRobotCreateHandler) *Dispatcher {
	return d.Register(model.EventCodeAppGroupChatMemberRobotCreate, wrapTypedHandler(handler))
}

// OnV7AppGroupChatMemberRobotDelete 注册机器人退群事件处理器
// 事件编码: kso.xz.app.group_chat.member.robot.delete
// 当机器人退出群聊时触发
func (d *Dispatcher) OnV7AppGroupChatMemberRobotDelete(handler V7AppGroupChatMemberRobotDeleteHandler) *Dispatcher {
	return d.Register(model.EventCodeAppGroupChatMemberRobotDelete, wrapTypedHandler(handler))
}
