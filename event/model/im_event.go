package model

// ================== 应用消息事件数据结构 ==================

// V7NotificationChatInfo 会话信息
type V7NotificationChatInfo struct {
	// Id 会话id
	Id string `json:"id"`

	// Type 会话类型
	Type string `json:"type"`
}

// V7NotificationMessageInfo 消息信息
type V7NotificationMessageInfo struct {
	// Id 消息id
	Id string `json:"id"`

	// Type 消息类型
	Type string `json:"type"`

	// Content 消息内容
	Content V7MessageContent `json:"content"`

	// Mentions 消息被@人列表
	Mentions []V7ChatMessageMention `json:"mentions,omitempty"`

	// QuoteMsgId 被引用的消息ID
	QuoteMsgId *string `json:"quote_msg_id,omitempty"`
}

// V7NotificationAppChatMessageCreateData 用户在会话（单聊/群聊）中给应用发送消息
// 事件编码: kso.app_chat.message.create
type V7NotificationAppChatMessageCreateData struct {
	// CompanyId 企业id
	CompanyId string `json:"company_id"`

	// Chat 会话
	Chat V7NotificationChatInfo `json:"chat"`

	// Sender 消息发送者
	Sender V7Identity `json:"sender"`

	// SendTime 消息发送时间戳（秒）
	SendTime int64 `json:"send_time"`

	// Message 消息
	Message V7NotificationMessageInfo `json:"message"`
}

// V7NotificationAppChatCreateData 首次创建用户和机器人的会话
// 事件编码: kso.app_chat.create
type V7NotificationAppChatCreateData struct {
	// ChatId 会话id
	ChatId string `json:"chat_id"`

	// Creator 会话创建者
	Creator V7Identity `json:"creator"`

	// CompanyId 企业id
	CompanyId string `json:"company_id"`
}

// ================== 应用群聊事件数据结构 ==================

// V7NotificationAppGroupChatData 群状态变更-应用事件通知
// 事件编码: kso.xz.app.group_chat.delete
type V7NotificationAppGroupChatData struct {
	// ChatId 会话id
	ChatId string `json:"chat_id"`

	// CompanyId 企业id
	CompanyId string `json:"company_id"`

	// Operator 操作人
	Operator V7Identity `json:"operator"`
}

// V7NotificationAppGroupChatMemberUserData 用户进出群-应用事件通知
// 事件编码: kso.xz.app.group_chat.member.user.create / kso.xz.app.group_chat.member.user.delete
type V7NotificationAppGroupChatMemberUserData struct {
	// ChatId 会话id
	ChatId string `json:"chat_id"`

	// CompanyId 企业id
	CompanyId string `json:"company_id"`

	// Operator 操作人
	Operator V7Identity `json:"operator"`

	// Users 进出群用户列表
	Users []V7Identity `json:"users"`
}

// V7NotificationAppGroupChatMemberRobotData 机器人进出群-应用事件通知
// 事件编码: kso.xz.app.group_chat.member.robot.create / kso.xz.app.group_chat.member.robot.delete
type V7NotificationAppGroupChatMemberRobotData struct {
	// ChatId 会话id
	ChatId string `json:"chat_id"`

	// CompanyId 企业id
	CompanyId string `json:"company_id"`

	// Operator 操作人
	Operator V7Identity `json:"operator"`
}
