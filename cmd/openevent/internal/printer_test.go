package internal

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GongchuangSu/open-event-sdk-go/event"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	color.NoColor = false
}

func newTestEvent(eventCode, data string) *event.Event {
	parts := strings.SplitN(eventCode, ".", 2)
	topic := ""
	operation := ""
	if len(parts) == 2 {
		topic = parts[0]
		operation = parts[1]
	} else {
		topic = eventCode
	}
	return event.NewEvent(topic, operation, time.Date(2026, 3, 26, 14, 30, 5, 0, time.UTC).Unix(), data)
}

func TestPrinter_ColoredEvent(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	p, err := NewPrinter(PrinterConfig{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	require.NoError(t, err)
	p.Start()

	e := newTestEvent("kso.app_chat.message.create", `{"chat_id":"c1","message":"hello"}`)
	p.SendEvent(e)

	time.Sleep(50 * time.Millisecond)
	p.Close()

	output := stdout.String()
	assert.Contains(t, output, "app_chat.message.create")
	assert.Contains(t, output, "📨")
	assert.Contains(t, output, "chat_id")
	assert.Contains(t, output, "hello")
}

func TestPrinter_JsonEvent(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	p, err := NewPrinter(PrinterConfig{
		JsonMode: true,
		Stdout:   &stdout,
		Stderr:   &stderr,
	})
	require.NoError(t, err)
	p.Start()

	e := newTestEvent("kso.app_chat.message.create", `{"chat_id":"c1"}`)
	p.SendEvent(e)

	time.Sleep(50 * time.Millisecond)
	p.Close()

	output := strings.TrimSpace(stdout.String())
	var envelope jsonEnvelope
	err = json.Unmarshal([]byte(output), &envelope)
	require.NoError(t, err, "output should be valid JSON: %s", output)

	assert.Equal(t, "kso.app_chat.message.create", envelope.EventCode)
	assert.NotEmpty(t, envelope.Time)
	assert.JSONEq(t, `{"chat_id":"c1"}`, string(envelope.Data))
}

func TestPrinter_JsonEvent_InvalidData(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	p, err := NewPrinter(PrinterConfig{
		JsonMode: true,
		Stdout:   &stdout,
		Stderr:   &stderr,
	})
	require.NoError(t, err)
	p.Start()

	e := newTestEvent("kso.test", "not valid json")
	p.SendEvent(e)

	time.Sleep(50 * time.Millisecond)
	p.Close()

	output := strings.TrimSpace(stdout.String())
	var envelope jsonEnvelope
	err = json.Unmarshal([]byte(output), &envelope)
	require.NoError(t, err)
	assert.Equal(t, `"not valid json"`, string(envelope.Data))
}

func TestPrinter_StatusMessages(t *testing.T) {
	tests := []struct {
		name     string
		jsonMode bool
		level    StatusLevel
		text     string
		checkOut func(t *testing.T, stdout, stderr string)
	}{
		{
			name:  "colored info goes to stdout",
			level: StatusInfo,
			text:  "✔ 已连接",
			checkOut: func(t *testing.T, stdout, stderr string) {
				assert.Contains(t, stdout, "已连接")
				assert.Empty(t, stderr)
			},
		},
		{
			name:  "colored error goes to stdout",
			level: StatusError,
			text:  "✖ 失败",
			checkOut: func(t *testing.T, stdout, stderr string) {
				assert.Contains(t, stdout, "失败")
				assert.Empty(t, stderr)
			},
		},
		{
			name:     "json mode status goes to stderr",
			jsonMode: true,
			level:    StatusInfo,
			text:     "connecting",
			checkOut: func(t *testing.T, stdout, stderr string) {
				assert.Empty(t, stdout)
				assert.Contains(t, stderr, "connecting")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			p, err := NewPrinter(PrinterConfig{
				JsonMode: tt.jsonMode,
				Stdout:   &stdout,
				Stderr:   &stderr,
			})
			require.NoError(t, err)
			p.Start()

			p.SendStatus(tt.level, tt.text)

			time.Sleep(50 * time.Millisecond)
			p.Close()

			tt.checkOut(t, stdout.String(), stderr.String())
		})
	}
}

func TestPrinter_FileOutput(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "output.txt")

	var stdout bytes.Buffer
	p, err := NewPrinter(PrinterConfig{
		FilePath: filePath,
		Stdout:   &stdout,
		Stderr:   &bytes.Buffer{},
	})
	require.NoError(t, err)
	p.Start()

	e := newTestEvent("kso.test.event", `{"key":"value"}`)
	p.SendEvent(e)
	p.SendStatus(StatusInfo, "✔ status line")

	time.Sleep(50 * time.Millisecond)
	p.Close()

	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	content := string(data)

	assert.Contains(t, content, "kso.test.event")
	assert.Contains(t, content, "key")
	assert.Contains(t, content, "status line")
}

func TestPrinter_FileOutput_Json(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "output.ndjson")

	var stdout bytes.Buffer
	p, err := NewPrinter(PrinterConfig{
		JsonMode: true,
		FilePath: filePath,
		Stdout:   &stdout,
		Stderr:   &bytes.Buffer{},
	})
	require.NoError(t, err)
	p.Start()

	e := newTestEvent("kso.test", `{"k":"v"}`)
	p.SendEvent(e)

	time.Sleep(50 * time.Millisecond)
	p.Close()

	data, err := os.ReadFile(filePath)
	require.NoError(t, err)

	var envelope jsonEnvelope
	err = json.Unmarshal(bytes.TrimSpace(data), &envelope)
	require.NoError(t, err)
	assert.Equal(t, "kso.test", envelope.EventCode)
}

func TestPrinter_ConcurrentSend(t *testing.T) {
	var stdout bytes.Buffer
	p, err := NewPrinter(PrinterConfig{
		JsonMode: true,
		Stdout:   &stdout,
		Stderr:   &bytes.Buffer{},
	})
	require.NoError(t, err)
	p.Start()

	const numEvents = 100
	var wg sync.WaitGroup
	wg.Add(numEvents)

	for i := 0; i < numEvents; i++ {
		go func(idx int) {
			defer wg.Done()
			e := newTestEvent("kso.concurrent", `{"idx":`+strings.Repeat("1", 1)+`}`)
			p.SendEvent(e)
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	p.Close()

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	assert.Equal(t, numEvents, len(lines), "all events should be printed")

	for _, line := range lines {
		var envelope jsonEnvelope
		err := json.Unmarshal([]byte(line), &envelope)
		assert.NoError(t, err, "each line should be valid JSON: %s", line)
	}
}

func TestPrinter_SendAfterClose(t *testing.T) {
	var stdout bytes.Buffer
	p, err := NewPrinter(PrinterConfig{
		Stdout: &stdout,
		Stderr: &bytes.Buffer{},
	})
	require.NoError(t, err)
	p.Start()
	p.Close()

	assert.NotPanics(t, func() {
		p.SendEvent(newTestEvent("kso.test", `{}`))
		p.SendStatus(StatusInfo, "after close")
	})
}

func TestPrinter_DoubleClose(t *testing.T) {
	p, err := NewPrinter(PrinterConfig{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	})
	require.NoError(t, err)
	p.Start()

	assert.NotPanics(t, func() {
		p.Close()
		p.Close()
	})
}
