package internal

import "strings"

// EventFilter 客户端侧事件过滤器，基于 event_code 精确匹配
type EventFilter struct {
	codes map[string]struct{}
	all   bool
}

// NewEventFilter 创建事件过滤器
// eventsFlag 为空时匹配所有事件；否则按逗号分隔解析为精确匹配集合
func NewEventFilter(eventsFlag string) *EventFilter {
	if eventsFlag == "" {
		return &EventFilter{all: true}
	}
	codes := make(map[string]struct{})
	for _, code := range strings.Split(eventsFlag, ",") {
		code = strings.TrimSpace(code)
		if code != "" {
			codes[code] = struct{}{}
		}
	}
	if len(codes) == 0 {
		return &EventFilter{all: true}
	}
	return &EventFilter{codes: codes}
}

// Match 判断事件码是否匹配过滤器
func (f *EventFilter) Match(eventCode string) bool {
	if f.all {
		return true
	}
	_, ok := f.codes[eventCode]
	return ok
}
