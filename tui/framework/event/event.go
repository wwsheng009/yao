package event

import (
	"time"
)

// EventType 事件类型
type EventType int

const (
	// 键盘事件
	EventKeyPress EventType = iota + 1000
	EventKeyRelease
	EventKeyRepeat

	// 鼠标事件
	EventMousePress
	EventMouseRelease
	EventMouseMove
	EventMouseWheel
	EventMouseEnter
	EventMouseLeave

	// 窗口事件
	EventResize
	EventFocus
	EventBlur
	EventClose

	// 组件事件
	EventClick
	EventDoubleClick
	EventContextMenu
	EventChange
	EventSubmit
	EventCancel
	EventSelect
	EventExpand
	EventCollapse

	// 自定义事件
	EventCustom = 10000
)

// Event 事件接口
type Event interface {
	// Type 返回事件类型
	Type() EventType

	// Timestamp 返回事件时间戳
	Timestamp() time.Time

	// Source 返回事件源组件
	Source() Component

	// Target 返回目标组件
	Target() Component

	// PreventDefault 阻止默认行为
	PreventDefault()

	// IsDefaultPrevented 是否已阻止默认行为
	IsDefaultPrevented() bool

	// StopPropagation 停止事件传播
	StopPropagation()

	// IsPropagationStopped 是否已停止传播
	IsPropagationStopped() bool
}

// Component 组件接口 (用于事件)
type Component interface {
	HandleEvent(Event) bool
}

// BaseEvent 基础事件实现
type BaseEvent struct {
	eventType EventType
	timestamp time.Time
	source    Component
	target    Component
	prevented bool
	stopped   bool
}

// Type 返回事件类型
func (e *BaseEvent) Type() EventType {
	return e.eventType
}

// Timestamp 返回事件时间戳
func (e *BaseEvent) Timestamp() time.Time {
	return e.timestamp
}

// Source 返回事件源
func (e *BaseEvent) Source() Component {
	return e.source
}

// Target 返回事件目标
func (e *BaseEvent) Target() Component {
	return e.target
}

// PreventDefault 阻止默认行为
func (e *BaseEvent) PreventDefault() {
	e.prevented = true
}

// IsDefaultPrevented 是否已阻止默认行为
func (e *BaseEvent) IsDefaultPrevented() bool {
	return e.prevented
}

// StopPropagation 停止事件传播
func (e *BaseEvent) StopPropagation() {
	e.stopped = true
}

// IsPropagationStopped 是否已停止传播
func (e *BaseEvent) IsPropagationStopped() bool {
	return e.stopped
}

// NewBaseEvent 创建基础事件
func NewBaseEvent(eventType EventType) *BaseEvent {
	return &BaseEvent{
		eventType: eventType,
		timestamp: time.Now(),
	}
}
