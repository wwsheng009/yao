package platform

import "os"

// Screen 屏幕输出抽象 (V3: 从 Terminal 拆分)
// Platform 层只提供"能力抽象"，不包含"语义"
// 它不应该知道 Focus、Event、Component、Layout
type Screen interface {
	// 初始化
	Init() error
	Close() error

	// 尺寸
	Size() (width, height int)

	// 输出
	Write(data []byte) (int, error)
	Flush() error

	// 清屏
	Clear() error

	// 备用屏幕
	EnterAlternateScreen() error
	ExitAlternateScreen() error
}

// DefaultScreen 默认实现 (Unix/Windows 通用部分)
type DefaultScreen struct {
	file *os.File
}

// NewDefaultScreen 创建默认屏幕
func NewDefaultScreen() *DefaultScreen {
	return &DefaultScreen{
		file: os.Stdout,
	}
}

// Init 初始化屏幕
func (s *DefaultScreen) Init() error {
	return nil
}

// Close 关闭屏幕
func (s *DefaultScreen) Close() error {
	return nil
}

// Size 获取屏幕尺寸
func (s *DefaultScreen) Size() (width, height int) {
	// TODO: 实现通过 syscall 获取实际尺寸
	return 80, 24
}

// Write 写入数据
func (s *DefaultScreen) Write(data []byte) (int, error) {
	return s.file.Write(data)
}

// Flush 刷新缓冲区
func (s *DefaultScreen) Flush() error {
	// 对于 os.Stdout，Flush 是 no-op
	return nil
}

// Clear 清屏
func (s *DefaultScreen) Clear() error {
	_, err := s.Write([]byte("\x1b[2J"))
	return err
}

// EnterAlternateScreen 进入备用屏幕
func (s *DefaultScreen) EnterAlternateScreen() error {
	_, err := s.Write([]byte("\x1b[?1049h"))
	return err
}

// ExitAlternateScreen 退出备用屏幕
func (s *DefaultScreen) ExitAlternateScreen() error {
	_, err := s.Write([]byte("\x1b[?1049l"))
	return err
}
