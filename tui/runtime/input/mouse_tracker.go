package input

import (
	"github.com/yaoapp/yao/tui/runtime/action"
	"github.com/yaoapp/yao/tui/runtime/platform"
	"time"
)

// ==============================================================================
// 鼠标状态跟踪器 (V3)
// ==============================================================================
// 跟踪鼠标状态以支持双击、拖动等高级操作

// MouseTracker 鼠标状态跟踪器
type MouseTracker struct {
	// 上次点击信息
	lastClickButton platform.MouseButton
	lastClickTime   time.Time
	lastClickX      int
	lastClickY      int
	lastClickCount  int

	// 当前拖动状态
	isDragging        bool
	dragStartX        int
	dragStartY        int
	dragButton        platform.MouseButton
	dragTargetID      string

	// 双击检测配置
	doubleClickTimeout time.Duration
	doubleClickDistance int
}

// NewMouseTracker 创建鼠标跟踪器
func NewMouseTracker() *MouseTracker {
	return &MouseTracker{
		doubleClickTimeout:  500 * time.Millisecond,
		doubleClickDistance: 5, // 像素
	}
}

// MouseEvent 鼠标事件结果
type MouseEvent struct {
	Action      *action.Action
	IsDoubleClick bool
	IsTripleClick bool
	IsDragStart   bool
	IsDragMove    bool
	IsDragEnd     bool
	DragStartX    int
	DragStartY    int
	DragDeltaX    int
	DragDeltaY    int
}

// ProcessInput 处理鼠标输入，返回相应的 Action
func (m *MouseTracker) ProcessInput(input platform.RawInput) *MouseEvent {
	if input.Type != platform.InputMouse {
		return nil
	}

	result := &MouseEvent{}

	switch input.MouseAction {
	case platform.MousePress:
		result = m.handlePress(input)
	case platform.MouseRelease:
		result = m.handleRelease(input)
	case platform.MouseMotion:
		result = m.handleMotion(input)
	case platform.MouseWheelUp, platform.MouseWheelDown:
		result = m.handleWheel(input)
	}

	return result
}

// handlePress 处理鼠标按下
func (m *MouseTracker) handlePress(input platform.RawInput) *MouseEvent {
	now := time.Now()
	result := &MouseEvent{}

	// 检查是否是双击/三击
	isDoubleClick := false
	isTripleClick := false

	if input.MouseButton == m.lastClickButton &&
		input.MouseButton != platform.MouseNone &&
		now.Sub(m.lastClickTime) < m.doubleClickTimeout &&
		abs(input.MouseX-m.lastClickX) <= m.doubleClickDistance &&
		abs(input.MouseY-m.lastClickY) <= m.doubleClickDistance {
		m.lastClickCount++
		if m.lastClickCount == 2 {
			isDoubleClick = true
		} else if m.lastClickCount >= 3 {
			isTripleClick = true
			m.lastClickCount = 0 // 重置
		}
	} else {
		m.lastClickCount = 1
	}

	m.lastClickButton = input.MouseButton
	m.lastClickTime = now
	m.lastClickX = input.MouseX
	m.lastClickY = input.MouseY

	// 开始拖动
	if input.MouseButton != platform.MouseNone &&
		input.MouseAction != platform.MouseWheelUp &&
		input.MouseAction != platform.MouseWheelDown {
		m.isDragging = true
		m.dragStartX = input.MouseX
		m.dragStartY = input.MouseY
		m.dragButton = input.MouseButton
		result.IsDragStart = true
		result.DragStartX = m.dragStartX
		result.DragStartY = m.dragStartY
	}

	// 创建 Action
	var actionType action.ActionType
	switch input.MouseButton {
	case platform.MouseLeft:
		if isTripleClick {
			actionType = action.ActionMouseTripleClick
		} else if isDoubleClick {
			actionType = action.ActionMouseDoubleClick
		} else {
			actionType = action.ActionMouseClick
		}
	case platform.MouseRight:
		actionType = action.ActionMouseRightClick
	case platform.MouseMiddle:
		actionType = action.ActionMouseMiddleClick
	default:
		return nil
	}

	result.Action = action.NewAction(actionType).
		WithPayload(MouseEventPayload{
			X:     input.MouseX,
			Y:     input.MouseY,
			Button: input.MouseButton,
		})

	result.IsDoubleClick = isDoubleClick
	result.IsTripleClick = isTripleClick

	return result
}

// handleRelease 处理鼠标释放
func (m *MouseTracker) handleRelease(input platform.RawInput) *MouseEvent {
	result := &MouseEvent{}

	if m.isDragging {
		m.isDragging = false
		result.IsDragEnd = true
		result.DragStartX = m.dragStartX
		result.DragStartY = m.dragStartY
		result.DragDeltaX = input.MouseX - m.dragStartX
		result.DragDeltaY = input.MouseY - m.dragStartY

		// 如果拖动距离很小，当作普通点击
		if abs(result.DragDeltaX) <= m.doubleClickDistance &&
			abs(result.DragDeltaY) <= m.doubleClickDistance {
			result.IsDragEnd = false
		}

		m.dragButton = platform.MouseNone
	}

	result.Action = action.NewAction(action.ActionMouseRelease).
		WithPayload(MouseEventPayload{
			X:     input.MouseX,
			Y:     input.MouseY,
			Button: input.MouseButton,
		})

	return result
}

// handleMotion 处理鼠标移动
func (m *MouseTracker) handleMotion(input platform.RawInput) *MouseEvent {
	result := &MouseEvent{}

	if m.isDragging {
		result.IsDragMove = true
		result.DragStartX = m.dragStartX
		result.DragStartY = m.dragStartY
		result.DragDeltaX = input.MouseX - m.dragStartX
		result.DragDeltaY = input.MouseY - m.dragStartY

		result.Action = action.NewAction(action.ActionMouseDrag).
			WithPayload(MouseEventPayload{
				X:         input.MouseX,
				Y:         input.MouseY,
				Button:    m.dragButton,
				StartX:    m.dragStartX,
				StartY:    m.dragStartY,
				DeltaX:    result.DragDeltaX,
				DeltaY:    result.DragDeltaY,
			})
	} else {
		result.Action = action.NewAction(action.ActionMouseMotion).
			WithPayload(MouseEventPayload{
			X:     input.MouseX,
			Y:     input.MouseY,
		})
	}

	return result
}

// handleWheel 处理鼠标滚轮
func (m *MouseTracker) handleWheel(input platform.RawInput) *MouseEvent {
	result := &MouseEvent{}

	var actionType action.ActionType
	if input.MouseAction == platform.MouseWheelUp {
		actionType = action.ActionMouseWheelUp
	} else {
		actionType = action.ActionMouseWheelDown
	}

	result.Action = action.NewAction(actionType).
		WithPayload(MouseEventPayload{
			X:     input.MouseX,
			Y:     input.MouseY,
			Button: input.MouseButton,
		})

	return result
}

// MouseEventPayload 鼠标事件负载
type MouseEventPayload struct {
	X      int
	Y      int
	Button platform.MouseButton
	StartX int // 拖动起始位置
	StartY int
	DeltaX int // 拖动偏移
	DeltaY int
}

// abs 返回绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
