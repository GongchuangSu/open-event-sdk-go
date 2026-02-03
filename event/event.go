// Package event 提供事件处理相关的类型和接口
package event

// Event 事件实体（解密后的事件数据）
type Event struct {
	// Topic 消息主题
	Topic string `json:"topic"`

	// Operation 消息变更动作
	Operation string `json:"operation"`

	// Time 时间（秒为单位的时间戳）
	Time int64 `json:"time"`

	// Data 解密后的事件数据（JSON 字符串）
	Data string `json:"data"`
}

// EventCode 返回事件编码，全局唯一，用于事件订阅和分发
// 事件编码格式: topic.operation，如 "kso.app_chat.message.create"
func (e *Event) EventCode() string {
	return BuildEventCode(e.Topic, e.Operation)
}

// BuildEventCode 根据 topic 和 operation 生成事件编码
// 事件编码格式: topic.operation，如 "kso.app_chat.message.create"
func BuildEventCode(topic, operation string) string {
	return topic + "." + operation
}

// NewEvent 创建事件实体
func NewEvent(topic, operation string, time int64, data string) *Event {
	return &Event{
		Topic:     topic,
		Operation: operation,
		Time:      time,
		Data:      data,
	}
}
