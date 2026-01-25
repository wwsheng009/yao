package event

import "time"

// KeyEvent 键盘事件
type KeyEvent struct {
	BaseEvent
	Key      rune
	Special  SpecialKey
	Modifiers KeyModifier
	Repeat   bool
}

// SpecialKey 特殊键定义
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

	// Vim 风格导航键
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

// Has 检查是否有修饰键
func (m KeyModifier) Has(mod KeyModifier) bool {
	return m&mod != 0
}

// NewKeyEvent 创建键盘事件
func NewKeyEvent(key rune) *KeyEvent {
	return &KeyEvent{
		BaseEvent: BaseEvent{
			eventType: EventKeyPress,
			timestamp: time.Now(),
		},
		Key: key,
	}
}

// NewSpecialKeyEvent 创建特殊键事件
func NewSpecialKeyEvent(special SpecialKey) *KeyEvent {
	return &KeyEvent{
		BaseEvent: BaseEvent{
			eventType: EventKeyPress,
			timestamp: time.Now(),
		},
		Special: special,
	}
}
