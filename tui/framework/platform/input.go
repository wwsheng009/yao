package platform

import "time"

// InputReader 输入读取抽象 (V3: 从 Terminal 拆分)
// Platform 只产生 RawInput，不产生语义化的 Action
// Action 转换由 Runtime 的 KeyMap 负责
type InputReader interface {
	// 读取单个输入
	ReadEvent() (RawInput, error)

	// 启动读取循环
	Start(events chan<- RawInput) error

	// 停止读取
	Stop() error
}

// RawInput 原始输入 (平台无关的表示)
type RawInput struct {
	Type RawInputType

	// 键盘
	Key      rune
	Special  SpecialKey
	Modifiers KeyModifier

	// 鼠标
	MouseX      int
	MouseY      int
	MouseButton MouseButton
	MouseAction MouseAction

	// 其他
	Data     []byte
	Timestamp time.Time
}

// RawInputType 输入类型
type RawInputType int

const (
	InputKeyPress RawInputType = iota
	InputKeyRelease
	InputMouse
	InputResize
	InputPaste
	InputSignal
)

// SpecialKey 特殊键
type SpecialKey int

const (
	KeyUnknown SpecialKey = iota

	// 控制键
	KeyEscape
	KeyEnter
	KeyTab
	KeyBackspace
	KeyDelete
	KeyInsert

	// 光标键
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown

	// 功能键
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12

	// 组合键
	KeySpace

	// Vim 风格
	KeyK // vim up
	KeyJ // vim down
	KeyH // vim left
	KeyL // vim right
)

// KeyModifier 修饰键
type KeyModifier uint8

const (
	ModShift KeyModifier = 1 << iota
	ModAlt
	ModCtrl
	ModMeta
)

// MouseButton 鼠标按钮
type MouseButton int

const (
	MouseNone MouseButton = iota
	MouseLeft
	MouseMiddle
	MouseRight
)

// MouseAction 鼠标动作
type MouseAction int

const (
	MousePress MouseAction = iota
	MouseRelease
	MouseMotion
	MouseWheelUp
	MouseWheelDown
)

// DefaultInputReader 默认输入读取器实现
// 实际实现将根据平台 (Unix/Windows) 不同
type DefaultInputReader struct {
	events chan<- RawInput
	quit   chan struct{}
}

// NewDefaultInputReader 创建默认输入读取器
func NewDefaultInputReader() *DefaultInputReader {
	return &DefaultInputReader{
		quit: make(chan struct{}),
	}
}

// Start 启动读取循环
func (r *DefaultInputReader) Start(events chan<- RawInput) error {
	r.events = events
	// TODO: 实现平台特定的输入读取
	return nil
}

// Stop 停止读取
func (r *DefaultInputReader) Stop() error {
	close(r.quit)
	return nil
}

// ReadEvent 读取单个事件 (同步方式)
func (r *DefaultInputReader) ReadEvent() (RawInput, error) {
	// TODO: 实现平台特定的输入读取
	return RawInput{}, nil
}
