package platform

import (
	"fmt"
	"os"
)

// Terminal 终端接口
type Terminal interface {
	// 初始化
	Init() error
	Close() error

	// 屏幕操作
	EnterAlternateScreen() error
	ExitAlternateScreen() error
	EnableRawMode() error
	DisableRawMode() error

	// 光标操作
	ShowCursor() error
	HideCursor() error
	MoveCursor(x, y int) error

	// 输出
	Write(data []byte) (int, error)
	WriteString(s string) (int, error)

	// 窗口
	GetSize() (width, height int)
}

// DefaultTerminal 默认终端实现
type DefaultTerminal struct {
	out *os.File
}

// NewDefaultTerminal 创建默认终端
func NewDefaultTerminal() Terminal {
	return &DefaultTerminal{
		out: os.Stdout,
	}
}

// Init 初始化终端
func (t *DefaultTerminal) Init() error {
	return nil
}

// Close 关闭终端
func (t *DefaultTerminal) Close() error {
	return nil
}

// EnterAlternateScreen 进入备用屏幕
func (t *DefaultTerminal) EnterAlternateScreen() error {
	_, err := t.WriteString("\x1b[?1049h")
	return err
}

// ExitAlternateScreen 退出备用屏幕
func (t *DefaultTerminal) ExitAlternateScreen() error {
	_, err := t.WriteString("\x1b[?1049l")
	return err
}

// EnableRawMode 启用原始模式
func (t *DefaultTerminal) EnableRawMode() error {
	// Unix raw mode 将在后续实现
	fmt.Print("\x1b[?7l") // 禁用自动换行
	return nil
}

// DisableRawMode 禁用原始模式
func (t *DefaultTerminal) DisableRawMode() error {
	fmt.Print("\x1b[?7h") // 启用自动换行
	return nil
}

// ShowCursor 显示光标
func (t *DefaultTerminal) ShowCursor() error {
	_, err := t.WriteString("\x1b[?25h")
	return err
}

// HideCursor 隐藏光标
func (t *DefaultTerminal) HideCursor() error {
	_, err := t.WriteString("\x1b[?25l")
	return err
}

// MoveCursor 移动光标
func (t *DefaultTerminal) MoveCursor(x, y int) error {
	_, err := t.WriteString(fmt.Sprintf("\x1b[%d;%dH", y+1, x+1))
	return err
}

// Write 写入数据
func (t *DefaultTerminal) Write(data []byte) (int, error) {
	return t.out.Write(data)
}

// WriteString 写入字符串
func (t *DefaultTerminal) WriteString(s string) (int, error) {
	return t.out.WriteString(s)
}

// GetSize 获取终端尺寸
func (t *DefaultTerminal) GetSize() (width, height int) {
	// 默认尺寸
	// TODO: 实现通过 syscall 获取实际尺寸
	return 80, 24
}
