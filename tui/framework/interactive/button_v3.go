package interactive

import (
	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/style"
	"github.com/yaoapp/yao/tui/runtime/action"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// ==============================================================================
// Button Component V3
// ==============================================================================
// V3 按钮组件，使用 CellBuffer 绘制，处理语义化 Action

// ButtonV3 V3 按钮组件
type ButtonV3 struct {
	*component.BaseComponentV3
	*component.StateHolder

	label       string
	normalStyle style.Style
	focusStyle  style.Style
	onClick     func()
}

// NewButtonV3 创建 V3 按钮组件
func NewButtonV3(label string) *ButtonV3 {
	return &ButtonV3{
		BaseComponentV3: component.NewBaseComponentV3("button"),
		StateHolder:     component.NewStateHolder(),
		label:           label,
		normalStyle:     style.Style{},
		focusStyle:      style.Style{}.Reverse(true),
		onClick:         nil,
	}
}

// NewButtonV3WithAction 创建带点击事件的 V3 按钮
func NewButtonV3WithAction(label string, onClick func()) *ButtonV3 {
	b := NewButtonV3(label)
	b.onClick = onClick
	return b
}

// ============================================================================
// 链式设置方法
// ============================================================================

// SetLabel 设置标签文本
func (b *ButtonV3) SetLabel(label string) *ButtonV3 {
	b.label = label
	return b
}

// GetLabel 获取标签文本
func (b *ButtonV3) GetLabel() string {
	return b.label
}

// SetNormalStyle 设置普通状态样式
func (b *ButtonV3) SetNormalStyle(s style.Style) *ButtonV3 {
	b.normalStyle = s
	return b
}

// SetFocusStyle 设置焦点状态样式
func (b *ButtonV3) SetFocusStyle(s style.Style) *ButtonV3 {
	b.focusStyle = s
	return b
}

// SetOnClick 设置点击事件处理
func (b *ButtonV3) SetOnClick(onClick func()) *ButtonV3 {
	b.onClick = onClick
	return b
}

// WithLabel 链式设置标签
func (b *ButtonV3) WithLabel(label string) *ButtonV3 {
	return b.SetLabel(label)
}

// WithNormalStyle 链式设置普通样式
func (b *ButtonV3) WithNormalStyle(s style.Style) *ButtonV3 {
	return b.SetNormalStyle(s)
}

// WithFocusStyle 链式设置焦点样式
func (b *ButtonV3) WithFocusStyle(s style.Style) *ButtonV3 {
	return b.SetFocusStyle(s)
}

// WithOnClick 链式设置点击事件
func (b *ButtonV3) WithOnClick(onClick func()) *ButtonV3 {
	return b.SetOnClick(onClick)
}

// ============================================================================
// Measurable 接口实现
// ============================================================================

// Measure 测量理想尺寸
// 按钮尺寸 = "[label]" + 左右各 1 空格
func (b *ButtonV3) Measure(maxWidth, maxHeight int) (width, height int) {
	labelWidth := buttonRuneCount(b.label)
	width = labelWidth + 2 // 左右括号
	height = 1

	// 加上内边距
	paddingLeft := 1
	paddingRight := 1
	width += paddingLeft + paddingRight

	// 不超过最大宽度
	if maxWidth > 0 && width > maxWidth {
		width = maxWidth
	}
	if maxHeight > 0 && height > maxHeight {
		height = maxHeight
	}

	return width, height
}

// ============================================================================
// Paintable 接口实现
// ============================================================================

// Paint 绘制组件到 CellBuffer
func (b *ButtonV3) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	if !b.IsVisible() {
		return
	}

	width := ctx.AvailableWidth
	height := ctx.AvailableHeight

	if width <= 0 || height <= 0 {
		return
	}

	// 选择样式
	drawStyle := b.normalStyle
	if b.IsFocused() {
		drawStyle = b.focusStyle
	}

	// 计算按钮文本
	labelWidth := buttonRuneCount(b.label)
	buttonText := "[" + b.label + "]"
	buttonWidth := labelWidth + 2

	// 计算水平居中位置
	paddingLeft := (width - buttonWidth) / 2
	if paddingLeft < 1 {
		paddingLeft = 1
	}

	// 计算垂直居中位置
	y := ctx.Y
	if height > 1 {
		y += (height - 1) / 2
	}

	// 绘制按钮
	x := ctx.X
	for i := 0; i < width; i++ {
		if i < paddingLeft || i >= paddingLeft+buttonWidth {
			// 绘制空格
			buf.SetCell(x+i, y, ' ', style.Style{})
		} else {
			// 绘制按钮字符
			charIndex := i - paddingLeft
			if charIndex < len(buttonText) {
				char := rune(buttonText[charIndex])
				buf.SetCell(x+i, y, char, drawStyle)
			} else {
				buf.SetCell(x+i, y, ' ', style.Style{})
			}
		}
	}
}

// ============================================================================
// ActionTarget 接口实现
// ============================================================================

// HandleAction 处理语义化 Action
func (b *ButtonV3) HandleAction(a action.Action) bool {
	switch a.Type {
	case action.ActionSubmit:
		fallthrough
	case action.ActionSelectItem:
		if b.onClick != nil {
			b.onClick()
		}
		return true
	}
	return false
}

// ============================================================================
// Focusable 接口实现
// ============================================================================

// FocusID 返回焦点标识符
func (b *ButtonV3) FocusID() string {
	return b.ID()
}

// OnFocus 获得焦点时调用
func (b *ButtonV3) OnFocus() {
	// 可以在这里添加获得焦点时的逻辑
	// 例如触发动画或状态更新
}

// OnBlur 失去焦点时调用
func (b *ButtonV3) OnBlur() {
	// 可以在这里添加失去焦点时的逻辑
}

// ============================================================================
// 内部方法
// ============================================================================

// buttonRuneCount 计算 rune 数量
func buttonRuneCount(s string) int {
	count := 0
	for range s {
		count++
	}
	return count
}

// ============================================================================
// 辅助函数 - 创建常用按钮样式
// ============================================================================

// PrimaryStyle 创建主要按钮样式
func PrimaryStyle() style.Style {
	return style.Style{}.
		Foreground(style.White).
		Background(style.Blue)
}

// SecondaryStyle 创建次要按钮样式
func SecondaryStyle() style.Style {
	return style.Style{}.
		Foreground(style.Black).
		Background(style.BrightBlack)
}

// DangerStyle 创建危险按钮样式
func DangerStyle() style.Style {
	return style.Style{}.
		Foreground(style.White).
		Background(style.Red)
}

// SuccessStyle 创建成功按钮样式
func SuccessStyle() style.Style {
	return style.Style{}.
		Foreground(style.White).
		Background(style.Green)
}
