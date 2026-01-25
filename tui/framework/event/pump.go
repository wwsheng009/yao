package event

import (
	"time"

	"github.com/yaoapp/yao/tui/framework/platform"
)

// Pump 事件泵 - 从平台读取原始输入并转换为事件
type Pump struct {
	input  platform.InputReader
	events chan Event
	quit   chan struct{}
	running bool
}

// NewPump 创建事件泵
func NewPump(reader platform.InputReader) *Pump {
	return &Pump{
		input:  reader,
		events: make(chan Event, 100),
		quit:   make(chan struct{}),
		running: false,
	}
}

// Start 启动事件泵
func (p *Pump) Start() error {
	if p.running {
		return nil
	}

	// 创建原始输入通道
	rawInputs := make(chan platform.RawInput, 50)

	// 启动平台输入读取器
	if err := p.input.Start(rawInputs); err != nil {
		return err
	}

	p.running = true

	// 启动转换循环
	go p.convertLoop(rawInputs)

	return nil
}

// convertLoop 转换原始输入为事件
func (p *Pump) convertLoop(rawInputs <-chan platform.RawInput) {
	for {
		select {
		case <-p.quit:
			return

		case raw, ok := <-rawInputs:
			if !ok {
				return
			}
			ev := p.convertToEvent(raw)
			if ev != nil {
				select {
				case p.events <- ev:
				case <-p.quit:
					return
				}
			}
		}
	}
}

// convertToEvent 将原始输入转换为事件
func (p *Pump) convertToEvent(raw platform.RawInput) Event {
	switch raw.Type {
	case platform.InputKeyPress:
		return p.convertKeyEvent(raw)

	case platform.InputResize:
		return p.convertResizeEvent(raw)

	case platform.InputMouse:
		return p.convertMouseEvent(raw)

	default:
		return nil
	}
}

// convertKeyEvent 转换键盘事件
func (p *Pump) convertKeyEvent(raw platform.RawInput) Event {
	ev := &KeyEvent{
		BaseEvent: BaseEvent{
			eventType: EventKeyPress,
			timestamp: raw.Timestamp,
		},
		Special:  SpecialKey(raw.Special),
		Modifiers: KeyModifier(raw.Modifiers),
	}

	// 设置字符键
	if raw.Key > 0 && raw.Special == platform.KeyUnknown {
		ev.Key = raw.Key
	}

	return ev
}

// convertResizeEvent 转换窗口调整事件
func (p *Pump) convertResizeEvent(raw platform.RawInput) Event {
	ev := &ResizeEvent{
		BaseEvent: BaseEvent{
			eventType: EventResize,
			timestamp: raw.Timestamp,
		},
	}

	// 从 Data 解析尺寸 (格式: "width,height")
	if len(raw.Data) >= 2 {
		// 简化处理，实际应该解析 ANSI 窗口调整序列
	}

	return ev
}

// convertMouseEvent 转换鼠标事件
func (p *Pump) convertMouseEvent(raw platform.RawInput) Event {
	ev := &MousePressEvent{
		BaseEvent: BaseEvent{
			eventType: EventMousePress,
			timestamp: raw.Timestamp,
		},
		X:      raw.MouseX,
		Y:      raw.MouseY,
		Button: MouseButton(raw.MouseButton),
	}

	switch raw.MouseAction {
	case platform.MousePress:
		ev.eventType = EventMousePress
	case platform.MouseRelease:
		ev.eventType = EventMouseRelease
	case platform.MouseMotion:
		ev.eventType = EventMouseMove
	case platform.MouseWheelUp:
		ev.eventType = EventMouseWheel
	}

	return ev
}

// Stop 停止事件泵
func (p *Pump) Stop() {
	if !p.running {
		return
	}

	p.running = false

	// 发送停止信号
	close(p.quit)

	// 停止输入读取器
	if p.input != nil {
		p.input.Stop()
	}

	// 关闭事件通道
	close(p.events)
}

// Events 返回事件通道
func (p *Pump) Events() <-chan Event {
	return p.events
}

// IsRunning 检查是否正在运行
func (p *Pump) IsRunning() bool {
	return p.running
}

// PumpWithTimeout 带超时的事件获取
func (p *Pump) PumpWithTimeout(timeout time.Duration) (Event, bool) {
	select {
	case ev := <-p.events:
		return ev, true
	case <-time.After(timeout):
		return nil, false
	}
}
