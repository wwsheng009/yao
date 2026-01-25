package platform

// ==============================================================================
// Platform Abstraction (V3)
// ==============================================================================
// RuntimePlatform 运行时平台（为 Runtime 提供的简化接口）
// Platform 将所有底层能力组合在一起，但避免接口方法冲突
type RuntimePlatform interface {
	// Init 初始化平台
	Init() error

	// Close 关闭平台
	Close() error

	// Size 获取屏幕尺寸
	Size() (width, height int)

	// ReadInput 读取输入
	ReadInput() *RawInput

	// WriteString 写入字符串
	WriteString(s string) (int, error)

	// Clear 清屏
	Clear() error
}

// DefaultPlatform 默认平台实现
type DefaultPlatform struct {
	*DefaultScreen
	*DefaultInputReader
	*DefaultSignalHandler
}

// NewDefaultPlatform 创建默认平台
func NewDefaultPlatform() *DefaultPlatform {
	return &DefaultPlatform{
		DefaultScreen:        NewDefaultScreen(),
		DefaultInputReader:   NewDefaultInputReader(),
		DefaultSignalHandler: NewDefaultSignalHandler(),
	}
}

// Init 初始化平台
func (p *DefaultPlatform) Init() error {
	if err := p.DefaultScreen.Init(); err != nil {
		return err
	}
	return nil
}

// Close 关闭平台
func (p *DefaultPlatform) Close() error {
	_ = p.DefaultScreen.Close()
	_ = p.DefaultInputReader.Stop()
	_ = p.DefaultSignalHandler.Stop()
	return nil
}

// Size 获取屏幕尺寸
func (p *DefaultPlatform) Size() (width, height int) {
	return p.DefaultScreen.Size()
}

// ReadInput 读取输入（同步方式）
func (p *DefaultPlatform) ReadInput() *RawInput {
	input, _ := p.DefaultInputReader.ReadEvent()
	return &input
}

// WriteString 写入字符串
func (p *DefaultPlatform) WriteString(s string) (int, error) {
	return p.DefaultScreen.Write([]byte(s))
}

// Clear 清屏
func (p *DefaultPlatform) Clear() error {
	return p.DefaultScreen.Clear()
}
