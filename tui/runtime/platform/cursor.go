package platform

import "fmt"

// Cursor 光标控制抽象 (V3: 从 Terminal 拆分)
type Cursor interface {
	// 显示控制
	Show() error
	Hide() error

	// 移动
	Move(x, y int) error

	// 位置查询
	Position() (x, y int, err error)

	// 样式
	SetStyle(style CursorStyle) error
}

// CursorStyle 光标样式
type CursorStyle int

const (
	CursorBlock   CursorStyle = iota
	CursorUnderline
	CursorBar
)

// DefaultCursor 默认光标实现
type DefaultCursor struct {
	screen Screen
}

// NewDefaultCursor 创建默认光标
func NewDefaultCursor(screen Screen) *DefaultCursor {
	return &DefaultCursor{screen: screen}
}

// Show 显示光标
func (c *DefaultCursor) Show() error {
	_, err := c.screen.Write([]byte("\x1b[?25h"))
	return err
}

// Hide 隐藏光标
func (c *DefaultCursor) Hide() error {
	_, err := c.screen.Write([]byte("\x1b[?25l"))
	return err
}

// Move 移动光标
// x, y 是 0-based 坐标
func (c *DefaultCursor) Move(x, y int) error {
	_, err := c.screen.Write([]byte(fmt.Sprintf("\x1b[%d;%dH", y+1, x+1)))
	return err
}

// Position 获取光标位置
// 注意：ANSI 转义码查询光标位置需要终端响应，实现较复杂
// 这里返回默认值
func (c *DefaultCursor) Position() (x, y int, err error) {
	// TODO: 实现 DSR 查询
	return 0, 0, nil
}

// SetStyle 设置光标样式
func (c *DefaultCursor) SetStyle(style CursorStyle) error {
	// DECSCUSR 序列设置光标样式
	// 0 = blinking block (default)
	// 1 = blinking block (default)
	// 2 = steady block
	// 3 = blinking underline
	// 4 = steady underline
	// 5 = blinking bar
	// 6 = steady bar
	var seq string
	switch style {
	case CursorBlock:
		seq = "\x1b[2 q" // steady block
	case CursorUnderline:
		seq = "\x1b[4 q" // steady underline
	case CursorBar:
		seq = "\x1b[6 q" // steady bar
	}
	_, err := c.screen.Write([]byte(seq))
	return err
}
