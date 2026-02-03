package model

// 事件编码常量
// 事件编码格式: topic.operation，如 "kso.app_chat.message.create"
const (
	// EventCodeAppChatMessageCreate 用户在会话（单聊/群聊）中给应用发送消息
	EventCodeAppChatMessageCreate = "kso.app_chat.message.create"

	// EventCodeAppChatCreate 首次创建用户和机器人的会话
	EventCodeAppChatCreate = "kso.app_chat.create"

	// EventCodeAppGroupChatDelete 群聊解散
	EventCodeAppGroupChatDelete = "kso.xz.app.group_chat.delete"

	// EventCodeAppGroupChatMemberUserCreate 用户进群
	EventCodeAppGroupChatMemberUserCreate = "kso.xz.app.group_chat.member.user.create"

	// EventCodeAppGroupChatMemberUserDelete 用户退群
	EventCodeAppGroupChatMemberUserDelete = "kso.xz.app.group_chat.member.user.delete"

	// EventCodeAppGroupChatMemberRobotCreate 机器人进群
	EventCodeAppGroupChatMemberRobotCreate = "kso.xz.app.group_chat.member.robot.create"

	// EventCodeAppGroupChatMemberRobotDelete 机器人退群
	EventCodeAppGroupChatMemberRobotDelete = "kso.xz.app.group_chat.member.robot.delete"
)
