package interactive

import (
	"strings"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/event"
)

// Button 按钮组件
type Button struct {
	*component.BaseComponent

	label   string
	onClick func()
}

// NewButton 创建按钮
func NewButton(label string) *Button {
	return &Button{
		BaseComponent: component.NewBaseComponent("button"),
		label:         label,
	}
}

// NewButtonWithAction 创建带点击事件的按钮
func NewButtonWithAction(label string, onClick func()) *Button {
	b := NewButton(label)
	b.onClick = onClick
	return b
}

// SetLabel 设置标签
func (b *Button) SetLabel(label string) {
	b.label = label
}

// GetLabel 获取标签
func (b *Button) GetLabel() string {
	return b.label
}

// SetOnClick 设置点击事件
func (b *Button) SetOnClick(onClick func()) {
	b.onClick = onClick
}

// Render 渲染按钮
func (b *Button) Render(ctx *component.RenderContext) string {
	if !b.IsVisible() {
		return ""
	}

	s := b.GetStyle()

	// 如果有焦点，使用焦点样式
	if b.HasFocus() {
		s = s.Reverse(true)
	}

	labelWidth := runeCount(b.label)
	width, height := ctx.AvailableWidth, ctx.AvailableHeight

	if width <= 0 {
		return ""
	}

	// 居中显示
	padding := (width - labelWidth - 2) / 2
	if padding < 0 {
		padding = 0
	}

	// 渲染按钮
	leftPad := strings.Repeat(" ", padding)
	rightPad := strings.Repeat(" ", width-labelWidth-2*padding-2)

	if height > 0 && height <= 3 {
		vPadding := (height - 1) / 2
		topPad := strings.Repeat("\n", vPadding)
		return topPad + leftPad + "[" + b.label + "]" + rightPad
	}

	return leftPad + "[" + s.Apply(b.label) + "]" + rightPad
}

// HasFocus 检查是否有焦点
func (b *Button) HasFocus() bool {
	return false
}

// HandleEvent 处理事件
func (b *Button) HandleEvent(ev event.Event) bool {
	switch e := ev.(type) {
	case *event.KeyEvent:
		if e.Key == '\r' || e.Key == ' ' {
			if b.onClick != nil {
				b.onClick()
			}
			return true
		}
	}
	return false
}

// runeCount 计算 rune 数量
func runeCount(s string) int {
	count := 0
	for range s {
		count++
	}
	return count
}
