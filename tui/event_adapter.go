package tui

import (
	"time"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/runtime"
	"github.com/yaoapp/yao/tui/runtime/event"
)

// ConvertBubbleTeaMsg converts a Bubble Tea message to a Runtime event.
// This adapter maintains module boundary rules by sitting in the tui package
// (which can import both Bubble Tea and runtime).
func ConvertBubbleTeaMsg(msg tea.Msg) event.Event {
	switch m := msg.(type) {
	case tea.MouseMsg:
		return convertMouseMsg(m)
	case tea.KeyMsg:
		return convertKeyMsg(m)
	case tea.WindowSizeMsg:
		return convertWindowSizeMsg(m)
	default:
		// For custom messages, wrap as-is
		return event.Event{
			Type:   event.EventTypeCustom,
			Custom: msg,
		}
	}
}

// convertMouseMsg converts a Bubble Tea MouseMsg to a Runtime MouseEvent.
func convertMouseMsg(msg tea.MouseMsg) event.Event {
	// Determine mouse event type
	var evType event.MouseEventType
	var click event.MouseClickType

	// Bubble Tea MouseMsg contains: X, Y, Action, and Button
	switch msg.Action {
	case tea.MouseActionPress:
		evType = event.MousePress
		// Map button type
		switch msg.Button {
		case tea.MouseButtonLeft:
			click = event.MouseLeft
		case tea.MouseButtonMiddle:
			click = event.MouseMiddle
		case tea.MouseButtonRight:
			click = event.MouseRight
		default:
			click = event.MouseLeft
		}
	case tea.MouseActionRelease:
		evType = event.MouseRelease
		switch msg.Button {
		case tea.MouseButtonLeft:
			click = event.MouseLeft
		case tea.MouseButtonMiddle:
			click = event.MouseMiddle
		case tea.MouseButtonRight:
			click = event.MouseRight
		default:
			click = event.MouseLeft
		}
	case tea.MouseActionMotion:
		evType = event.MouseMove
		click = event.MouseLeft
	default:
		// Check for scroll wheel
		if msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown {
			evType = event.MouseScroll
		} else {
			evType = event.MousePress
			click = event.MouseLeft
		}
	}

	return event.Event{
		Type: event.EventTypeMouse,
		Mouse: &event.MouseEvent{
			X:     msg.X,
			Y:     msg.Y,
			Type:  evType,
			Click: click,
		},
	}
}

// convertKeyMsg converts a Bubble Tea KeyMsg to a Runtime KeyEvent.
func convertKeyMsg(msg tea.KeyMsg) event.Event {
	// Extract rune from key
	var key rune
	if len(msg.Runes) > 0 {
		key = msg.Runes[0]
	}

	// Determine key modifier
	var mod event.KeyModifier
	if msg.Alt {
		mod = event.ModAlt
	} else if msg.Type == tea.KeyShiftLeft || msg.Type == tea.KeyShiftRight {
		mod = event.ModShift
	} else if msg.String() == "ctrl+c" || msg.String() == "ctrl+v" ||
		msg.String() == "ctrl+up" || msg.String() == "ctrl+down" ||
		msg.String() == "ctrl+left" || msg.String() == "ctrl+right" {
		// Check string representation for ctrl combinations
		mod = event.ModCtrl
	}

	return event.Event{
		Type: event.EventTypeKey,
		Key: &event.KeyEvent{
			Key:  key,
			Type: event.KeyPress,
			Mod:  mod,
		},
	}
}

// convertWindowSizeMsg converts a Bubble Tea WindowSizeMsg to a Runtime ResizeEvent.
func convertWindowSizeMsg(msg tea.WindowSizeMsg) event.Event {
	return event.Event{
		Type:   event.EventTypeResize,
		Resize: &event.ResizeEvent{
			Width:  msg.Width,
			Height: msg.Height,
		},
	}
}

// DispatchEventToRuntime dispatches a Bubble Tea message through the Runtime event system.
// It converts the message to a Runtime event, performs hit testing, and dispatches
// to the appropriate component.
func (m *Model) DispatchEventToRuntime(msg tea.Msg) event.EventResult {
	// Convert to Runtime event
	ev := ConvertBubbleTeaMsg(msg)

	// ========== 新增：处理文本选择 ==========
	// 在事件分发到组件之前，先让选择管理器处理鼠标事件
	if ev.Type == event.EventTypeMouse && ev.Mouse != nil && m.RuntimeEngine != nil {
		handled := m.handleSelectionMouseEvent(ev.Mouse)
		if handled {
			m.forceRender = true // 选择变化需要重新渲染
			// 返回已处理的结果，但仍然需要处理焦点等
			return event.EventResult{
				Handled: true,
				Updated: true,
			}
		}
	}
	// ============================================

	// Get the layout boxes from Runtime engine
	// The boxes are the output of the layout phase and contain position information
	var boxes []runtime.LayoutBox
	if m.RuntimeEngine != nil {
		boxes = m.RuntimeEngine.GetBoxes()
	}

	// Dispatch the event
	result := event.DispatchEvent(ev, boxes)

	// Handle focus changes if needed
	if result.FocusChange != event.FocusChangeNone {
		m.handleFocusChange(result.FocusChange)
	}

	// Handle focus target (e.g., mouse click on a focusable component)
	if result.FocusTarget != "" {
		m.handleFocusTarget(result.FocusTarget)
	}

	return result
}

// handleFocusChange handles focus change requests from event dispatch.
func (m *Model) handleFocusChange(change event.FocusChangeType) {
	if m.RuntimeEngine == nil {
		return
	}

	focusMgr := m.RuntimeEngine.GetFocusManager()
	if focusMgr == nil {
		return
	}

	switch change {
	case event.FocusChangeNext:
		focused := focusMgr.FocusNext()
		if focused != nil {
			log.Trace("Focus moved to next: %s", focused.ID)
			m.forceRender = true // 焦点变化需要重新渲染
		}
	case event.FocusChangePrev:
		focused := focusMgr.FocusPrev()
		if focused != nil {
			log.Trace("Focus moved to previous: %s", focused.ID)
			m.forceRender = true // 焦点变化需要重新渲染
		}
	case event.FocusChangeSpecific:
		// Focus would be set to a specific component
		// This requires additional context about which component to focus
		log.Trace("Focus change to specific component requested")
		m.forceRender = true // 焦点变化需要重新渲染
	}
}

// handleFocusTarget handles focusing a specific component by ID.
func (m *Model) handleFocusTarget(targetID string) {
	if m.RuntimeEngine == nil {
		return
	}

	focusMgr := m.RuntimeEngine.GetFocusManager()
	if focusMgr == nil {
		return
	}

	// Try to focus the specific component
	if focusMgr.Focus(targetID) {
		log.Trace("Focus set to component: %s", targetID)
		m.forceRender = true // 焦点变化需要重新渲染
	}
}

// handleSelectionMouseEvent handles mouse events for text selection.
// Returns true if the event was handled by the selection system.
// Supports single-click drag, double-click (word), triple-click (line).
func (m *Model) handleSelectionMouseEvent(ev *event.MouseEvent) bool {
	if m.RuntimeEngine == nil {
		log.Trace("handleSelectionMouseEvent: RuntimeEngine is nil")
		return false
	}

	selection := m.RuntimeEngine.GetSelection()
	if selection == nil {
		log.Trace("handleSelectionMouseEvent: selection is nil")
		return false
	}
	if !selection.IsEnabled() {
		log.Trace("handleSelectionMouseEvent: selection not enabled")
		return false
	}

	log.Trace("handleSelectionMouseEvent: ev.Type=%d, Click=%d, X=%d, Y=%d",
		ev.Type, ev.Click, ev.X, ev.Y)

	switch ev.Type {
	case event.MousePress:
		if ev.Click == event.MouseLeft {
			// 检测双击/三击
			currentTime := time.Now().UnixNano()
			clickTimeThreshold := int64(500 * 1000000) // 500ms

			// 检查是否是双击/三击
			isDoubleClick := (currentTime-m.lastClickTime) < clickTimeThreshold &&
				ev.X == m.lastClickX && ev.Y == m.lastClickY

			if isDoubleClick {
				m.clickCount++
			} else {
				m.clickCount = 1
			}
			m.lastClickTime = currentTime
			m.lastClickX = ev.X
			m.lastClickY = ev.Y

			// 开始选择
			m.mouseButtonDown = 1
			m.mouseDragStartX = ev.X
			m.mouseDragStartY = ev.Y
			m.lastMouseX = ev.X
			m.lastMouseY = ev.Y

			// 根据点击次数选择不同的模式
			switch m.clickCount {
			case 1:
				// 单击 - 字符选择
				selection.Start(ev.X, ev.Y)
				log.Trace("Selection started (char) at (%d, %d)", ev.X, ev.Y)
			case 2:
				// 双击 - 单词选择
				selection.SelectWord(ev.X, ev.Y)
				log.Trace("Selection started (word) at (%d, %d)", ev.X, ev.Y)
			case 3:
				// 三击 - 行选择
				selection.SelectLine(ev.Y)
				log.Trace("Selection started (line) at y=%d", ev.Y)
				m.clickCount = 0 // 重置
			}
			return true
		}

	case event.MouseMove:
		// 鼠标移动 - 如果左键按下，更新选择
		if m.mouseButtonDown == 1 {
			// 只有位置改变时才更新
			if ev.X != m.lastMouseX || ev.Y != m.lastMouseY {
				// 如果已经移动到不同位置，拖动模式会覆盖点击模式
				selection.Update(ev.X, ev.Y)
				m.lastMouseX = ev.X
				m.lastMouseY = ev.Y
				log.Trace("Selection updated to (%d, %d)", ev.X, ev.Y)
				return true
			}
		}

	case event.MouseRelease:
		if ev.Click == event.MouseLeft {
			// 左键释放 - 结束拖动
			m.mouseButtonDown = 0
			log.Trace("Selection drag ended at (%d, %d)", ev.X, ev.Y)
			return true
		}
	}

	return false
}

