package ws

import (
	"encoding/json"
	"testing"
)

func TestEventMessage_JSON(t *testing.T) {
	// 事件消息不包含 type 字段
	msg := EventMessage{
		Topic:         "user.created",
		Operation:     "create",
		Time:          1234567890,
		Nonce:         "abc123def456ghij",
		Signature:     "test_signature",
		EncryptedData: "encrypted_data_here",
	}

	// 序列化
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// 反序列化
	var decoded EventMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// 验证字段
	if decoded.Topic != msg.Topic {
		t.Errorf("Topic = %v, want %v", decoded.Topic, msg.Topic)
	}
	if decoded.Operation != msg.Operation {
		t.Errorf("Operation = %v, want %v", decoded.Operation, msg.Operation)
	}
	if decoded.Time != msg.Time {
		t.Errorf("Time = %v, want %v", decoded.Time, msg.Time)
	}
	if decoded.Nonce != msg.Nonce {
		t.Errorf("Nonce = %v, want %v", decoded.Nonce, msg.Nonce)
	}
	if decoded.Signature != msg.Signature {
		t.Errorf("Signature = %v, want %v", decoded.Signature, msg.Signature)
	}
	if decoded.EncryptedData != msg.EncryptedData {
		t.Errorf("EncryptedData = %v, want %v", decoded.EncryptedData, msg.EncryptedData)
	}
}

func TestGoAwayMessage_JSON(t *testing.T) {
	msg := GoAwayMessage{
		Type:        MessageTypeGoAway,
		Reason:      GoAwayReasonServerShutdown,
		Message:     "server is shutting down",
		ReconnectMs: 5000,
	}

	// 序列化
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// 反序列化
	var decoded GoAwayMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// 验证字段
	if decoded.Type != msg.Type {
		t.Errorf("Type = %v, want %v", decoded.Type, msg.Type)
	}
	if decoded.Reason != msg.Reason {
		t.Errorf("Reason = %v, want %v", decoded.Reason, msg.Reason)
	}
	if decoded.Message != msg.Message {
		t.Errorf("Message = %v, want %v", decoded.Message, msg.Message)
	}
	if decoded.ReconnectMs != msg.ReconnectMs {
		t.Errorf("ReconnectMs = %v, want %v", decoded.ReconnectMs, msg.ReconnectMs)
	}
}

func TestGoAwayMessage_OmitEmpty(t *testing.T) {
	// 当 ReconnectMs 为 0 时应该被省略
	msg := GoAwayMessage{
		Type:    MessageTypeGoAway,
		Reason:  GoAwayReasonHeartbeatTimeout,
		Message: "heartbeat timeout",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// 验证 reconnect_ms 不在 JSON 中
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if _, ok := result["reconnect_ms"]; ok {
		t.Error("expected reconnect_ms to be omitted when 0")
	}
}

func TestBaseMessage_ParseType(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string
		wantType string
	}{
		{
			name:     "event message without type",
			jsonStr:  `{"topic":"user.created","operation":"create"}`,
			wantType: "", // 事件消息没有 type 字段
		},
		{
			name:     "goaway message",
			jsonStr:  `{"type":"goaway","reason":"server_shutdown"}`,
			wantType: MessageTypeGoAway,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var base baseMessage
			if err := json.Unmarshal([]byte(tt.jsonStr), &base); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if base.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", base.Type, tt.wantType)
			}
		})
	}
}
