package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/GongchuangSu/open-event-sdk-go/event"
	"github.com/fatih/color"
)

// StatusLevel 状态消息级别
type StatusLevel int

const (
	StatusInfo StatusLevel = iota
	StatusWarn
	StatusError
)

const eventChannelBuffer = 256

type printItem struct {
	event *event.Event
	text  string
	level StatusLevel
}

// PrinterConfig 打印器配置
type PrinterConfig struct {
	JsonMode bool
	NoColor  bool
	FilePath string
	Stdout   io.Writer
	Stderr   io.Writer
}

// Printer 事件打印器，通过 channel 序列化所有输出
type Printer struct {
	ch       chan printItem
	done     chan struct{}
	closed   atomic.Bool
	stdout   io.Writer
	stderr   io.Writer
	file     *os.File
	jsonMode bool
}

// NewPrinter 创建打印器
func NewPrinter(cfg PrinterConfig) (*Printer, error) {
	if cfg.NoColor {
		color.NoColor = true
	}

	stdout := cfg.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := cfg.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	var file *os.File
	if cfg.FilePath != "" {
		f, err := os.Create(cfg.FilePath)
		if err != nil {
			return nil, fmt.Errorf("无法打开输出文件: %w", err)
		}
		file = f
	}

	return &Printer{
		ch:       make(chan printItem, eventChannelBuffer),
		done:     make(chan struct{}),
		stdout:   stdout,
		stderr:   stderr,
		file:     file,
		jsonMode: cfg.JsonMode,
	}, nil
}

// Start 启动消费 goroutine
func (p *Printer) Start() {
	go p.consumeLoop()
}

// SendEvent 发送事件到打印队列
func (p *Printer) SendEvent(e *event.Event) {
	if p.closed.Load() {
		return
	}
	defer func() { recover() }()
	p.ch <- printItem{event: e}
}

// SendStatus 发送状态消息到打印队列
func (p *Printer) SendStatus(level StatusLevel, text string) {
	if p.closed.Load() {
		return
	}
	defer func() { recover() }()
	p.ch <- printItem{text: text, level: level}
}

// Close 关闭打印器，排空队列并关闭文件
func (p *Printer) Close() {
	if p.closed.Swap(true) {
		return
	}
	close(p.ch)
	<-p.done
	if p.file != nil {
		p.file.Close()
	}
}

func (p *Printer) consumeLoop() {
	defer close(p.done)
	for item := range p.ch {
		if item.event != nil {
			p.renderEvent(item.event)
		} else if item.text != "" {
			p.renderStatus(item.text, item.level)
		}
	}
}

func (p *Printer) renderEvent(e *event.Event) {
	if p.jsonMode {
		p.renderJsonEvent(e)
	} else {
		p.renderColoredEvent(e)
	}
}

type jsonEnvelope struct {
	Time      string          `json:"time"`
	EventCode string          `json:"event_code"`
	Data      json.RawMessage `json:"data"`
}

func (p *Printer) renderJsonEvent(e *event.Event) {
	var data json.RawMessage
	if json.Valid([]byte(e.Data)) {
		data = json.RawMessage(e.Data)
	} else {
		data, _ = json.Marshal(e.Data)
	}

	envelope := jsonEnvelope{
		Time:      time.Unix(e.Time, 0).Format(time.RFC3339),
		EventCode: e.EventCode(),
		Data:      data,
	}

	line, err := json.Marshal(envelope)
	if err != nil {
		fmt.Fprintf(p.stderr, "JSON marshal error: %v\n", err)
		return
	}

	fmt.Fprintln(p.stdout, string(line))
	if p.file != nil {
		fmt.Fprintln(p.file, string(line))
	}
}

func (p *Printer) renderColoredEvent(e *event.Event) {
	ts := time.Unix(e.Time, 0).Format("15:04:05")
	eventCode := e.EventCode()

	var prettyData string
	if json.Valid([]byte(e.Data)) {
		var raw json.RawMessage
		if err := json.Unmarshal([]byte(e.Data), &raw); err == nil {
			if pretty, err := json.MarshalIndent(raw, "", "  "); err == nil {
				prettyData = string(pretty)
			}
		}
	}
	if prettyData == "" {
		prettyData = e.Data
	}

	gray := color.New(color.FgHiBlack)
	cyan := color.New(color.FgCyan)

	fmt.Fprintf(p.stdout, "\n%s 📨 %s\n%s\n",
		gray.Sprintf("[%s]", ts),
		cyan.Sprint(eventCode),
		prettyData,
	)

	if p.file != nil {
		fmt.Fprintf(p.file, "\n[%s] 📨 %s\n%s\n", ts, eventCode, prettyData)
	}
}

func (p *Printer) renderStatus(text string, level StatusLevel) {
	if p.jsonMode {
		fmt.Fprintln(p.stderr, text)
	} else {
		switch level {
		case StatusInfo:
			fmt.Fprintln(p.stdout, color.GreenString(text))
		case StatusWarn:
			fmt.Fprintln(p.stdout, color.YellowString(text))
		case StatusError:
			fmt.Fprintln(p.stdout, color.RedString(text))
		}
	}

	if p.file != nil {
		fmt.Fprintln(p.file, text)
	}
}
