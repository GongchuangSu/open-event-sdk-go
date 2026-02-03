package model

// ================== 消息相关 ==================

// V7MessageType 消息类型常量
const (
	V7MessageTypeText      = "text"       // 文本
	V7MessageTypeRichText  = "rich_text"  // 富文本
	V7MessageTypeImage     = "image"      // 图片
	V7MessageTypeFile      = "file"       // 文件
	V7MessageTypeSticker   = "sticker"    // 表情
	V7MessageTypeAudio     = "audio"      // 音频
	V7MessageTypeVideo     = "video"      // 视频
	V7MessageTypeLocation  = "location"   // 位置
	V7MessageTypeLink      = "link"       // 链接
	V7MessageTypeMiniApp   = "mini_app"   // 小程序
	V7MessageTypeTemplate  = "template"   // 模板
	V7MessageTypeShareChat = "share_chat" // 分享群聊
)

// V7MessageContent 消息内容
// 消息内容结构较复杂，使用 any 类型，订阅方可根据 message.type 自行解析
type V7MessageContent = any

// V7ChatMessageMention 消息@信息
type V7ChatMessageMention struct {
	// Id 指定聊天消息中at操作的实体索引ID
	// 与消息正文中相应 <at id={index}> 标记中的 {index} 值匹配
	Id string `json:"id"`

	// Type at操作对象类型
	Type string `json:"type"`

	// Identity 被at的用户信息，当at所有人时该值为空
	Identity *V7Identity `json:"identity,omitempty"`
}

// ================== 会话相关 ==================

// V7ChatType 会话类型常量
const (
	V7ChatTypeSingle = "single" // 单聊
	V7ChatTypeGroup  = "group"  // 群聊
)
