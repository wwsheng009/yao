package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// EventClass 表示事件的类型分类，用于决定消息处理路径
type EventClass int

const (
	// GeometryEvent: 几何事件，通过 Runtime 事件系统处理
	// 特点：需要命中测试、焦点导航，不需要返回 tea.Cmd
	// 例如：鼠标点击、Tab/Shift+Tab 导航
	GeometryEvent EventClass = iota

	// ComponentEvent: 组件消息，必须通过 Bubble Tea 消息路径处理
	// 特点：需要保留 tea.Cmd 传递到主消息循环
	// 例如：键盘输入、光标闪烁、组件状态更新
	ComponentEvent

	// SystemEvent: 系统消息，需要特殊处理
	// 特点：影响整个 UI 状态，需要通知所有组件
	// 例如：窗口大小变化
	SystemEvent
)

// ClassifyMessage 对消息进行分类，决定使用哪种处理路径
//
// 几何事件路径 (GeometryEvent):
//   - 使用 Runtime 事件系统进行命中测试
//   - 处理焦点导航
//   - 不需要返回 tea.Cmd
//
// 组件消息路径 (ComponentEvent):
//   - 直接广播到订阅的组件
//   - 收集并返回所有 tea.Cmd
//   - 确保 Bubble Tea 消息循环正常工作
func ClassifyMessage(msg tea.Msg) EventClass {
	switch msg.(type) {
	case tea.MouseMsg:
		// 鼠标消息需要命中测试 → 几何事件
		return GeometryEvent

	case tea.KeyMsg:
		// 键盘消息默认为组件事件（需要传递 Cmd）
		// Tab/Shift+Tab 会在处理阶段特殊处理
		return ComponentEvent

	case tea.WindowSizeMsg:
		// 窗口大小是系统消息
		return SystemEvent

	default:
		// cursor.BlinkMsg, tea.FocusMsg, tea.BlurMsg 等都是组件事件
		// 这些消息必须传递到组件，且返回的 Cmd 必须执行
		return ComponentEvent
	}
}

// IsNavigationKey 检查是否为焦点导航键
// 导航键使用 Runtime 焦点系统，不传递到组件
func IsNavigationKey(msg tea.KeyMsg) bool {
	return msg.Type == tea.KeyTab ||
		msg.Type == tea.KeyShiftTab ||
		(msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == '\t')
}

// ShouldDispatchToRuntime 检查消息是否应该通过 Runtime 事件系统分发
func ShouldDispatchToRuntime(msg tea.Msg) bool {
	return ClassifyMessage(msg) == GeometryEvent
}

// ShouldPreserveCommands 检查消息是否应该保留 tea.Cmd
func ShouldPreserveCommands(msg tea.Msg) bool {
	return ClassifyMessage(msg) == ComponentEvent
}
