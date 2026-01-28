package framework

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/debug"
	frameworkevent "github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/theme"
	"github.com/yaoapp/yao/tui/runtime/core"
	"github.com/yaoapp/yao/tui/runtime/paint"
	"github.com/yaoapp/yao/tui/runtime/platform"
	"github.com/yaoapp/yao/tui/runtime/render"
	"github.com/yaoapp/yao/tui/runtime/style"
)

// AppState 应用状态
type AppState int

const (
	StateCreated AppState = iota
	StateInitializing
	StateRunning
	StatePaused
	StateStopping
	StateStopped
	StateError
)

// App 主应用程序
type App struct {
	// 组件树
	root component.Node

	// 事件
	router       *frameworkevent.Router
	keyMap       *frameworkevent.KeyMap
	pump         *frameworkevent.Pump
	eventFilter  func(frameworkevent.Event) bool // 事件过滤器回调，返回 false 表示拦截

	// 生命周期
	state AppState
	quit  chan struct{}
	dirty bool

	// 终端尺寸
	terminalWidth  int
	terminalHeight int

	// 首次渲染标记
	firstRender bool

	// 上一帧缓冲区（用于局部刷新）
	prevBuffer [][]paint.Cell

	// 光标位置跟踪（用于强制刷新光标区域）
	lastCursorX int
	lastCursorY int

	// 配置
	tickInterval time.Duration

	// 调试模式
	debugMode     bool
	debugRecorder *debug.Recorder
	debugLogFile  string

	// Panic 恢复管理器
	recovery *core.Recovery

	// 渲染节流器
	throttler *render.Throttler

	// 上下文管理器
	contextMgr *core.ContextManager

	// 主题管理器
	themeMgr     *theme.Manager
	themeName    string // 当前主题名称
	themeEnabled bool   // 是否启用主题系统

	// 用户数据存储（用于存储任意用户定义数据）
	userData map[string]interface{}
}

// NewApp 创建新应用
func NewApp() *App {
	return &App{
		router:       frameworkevent.NewRouter(),
		keyMap:       frameworkevent.NewKeyMap(),
		eventFilter:  func(ev frameworkevent.Event) bool { return true }, // 默认放行所有事件
		quit:         make(chan struct{}),
		tickInterval: 16 * time.Millisecond, // ~60fps
		firstRender:  true,
		debugMode:    os.Getenv("TUI_DEBUG") == "true",
		debugLogFile: os.Getenv("TUI_DEBUG_LOG"),
		throttler:    render.NewThrottler(60), // 默认 60 FPS
		contextMgr:   core.NewContextManager(context.Background()),
		userData:     make(map[string]interface{}),
	}
}

// SetDebugMode 设置调试模式
func (a *App) SetDebugMode(enabled bool) {
	a.debugMode = enabled
	if enabled && a.debugRecorder == nil {
		logFile := a.debugLogFile
		if logFile == "" {
			logFile = fmt.Sprintf("tui_debug_%s.log", time.Now().Format("20060102_150405"))
		}
		recorder, err := debug.NewRecorder(logFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create debug recorder: %v\n", err)
			return
		}
		a.debugRecorder = recorder
		fmt.Fprintf(os.Stderr, "Debug mode enabled, logging to: %s\n", logFile)
	}
}

// IsDebugMode 检查是否在调试模式
func (a *App) IsDebugMode() bool {
	return a.debugMode
}

// ============================================================================
// 事件过滤器配置
// ============================================================================

// SetEventFilter 设置事件过滤器回调
// 返回 false 表示拦截该事件，不再继续处理
func (a *App) SetEventFilter(filter func(frameworkevent.Event) bool) {
	a.eventFilter = filter
}

// ClearEventFilter 清除事件过滤器
func (a *App) ClearEventFilter() {
	a.eventFilter = func(ev frameworkevent.Event) bool { return true }
}

// ============================================================================
// Panic 恢复配置
// ============================================================================

// EnableRecovery 启用 panic 恢复
func (a *App) EnableRecovery() {
	if a.recovery == nil {
		a.recovery = core.NewRecovery(a)
	}
}

// SetPanicLog 设置 panic 日志文件
func (a *App) SetPanicLog(filename string) error {
	a.EnableRecovery()
	return a.recovery.EnablePanicLog(filename)
}

// AddPanicHandler 添加 panic 处理器
func (a *App) AddPanicHandler(handler core.PanicHandler) {
	a.EnableRecovery()
	a.recovery.AddHandler(handler)
}

// ============================================================================
// 渲染节流配置
// ============================================================================

// SetFPS 设置目标帧率
func (a *App) SetFPS(fps int) {
	a.throttler.SetFPS(fps)
}

// FPS 获取当前帧率
func (a *App) FPS() int {
	return a.throttler.FPS()
}

// ActualFPS 获取实际帧率
func (a *App) ActualFPS() float64 {
	return a.throttler.ActualFPS()
}

// EnableAdaptiveFPS 启用自适应帧率
func (a *App) EnableAdaptiveFPS(enable bool) {
	a.throttler.EnableAdaptive(enable)
}

// GetRenderStats 获取渲染统计信息
func (a *App) GetRenderStats() render.Stats {
	return a.throttler.Stats()
}

// ForceRender 强制下次渲染（跳过节流限制）
func (a *App) ForceRender() {
	a.throttler.ForceRender()
	a.dirty = true
}

// ============================================================================
// 主题系统配置
// ============================================================================

// InitTheme 初始化主题系统
// 如果未指定主题名称，则使用默认主题 "dark"
func (a *App) InitTheme(themeName string) error {
	mgr, err := theme.InitThemes(themeName)
	if err != nil {
		return fmt.Errorf("failed to initialize theme: %w", err)
	}
	a.themeMgr = mgr
	a.themeName = mgr.Current().Name
	a.themeEnabled = true
	return nil
}

// SetTheme 切换主题
func (a *App) SetTheme(name string) error {
	if a.themeMgr == nil {
		return errors.New("theme manager not initialized, call InitTheme first")
	}
	if err := a.themeMgr.Set(name); err != nil {
		return err
	}
	a.themeName = name
	a.dirty = true // 触发重绘
	return nil
}

// GetTheme 获取当前主题名称
func (a *App) GetTheme() string {
	return a.themeName
}

// ThemeManager 获取主题管理器
func (a *App) ThemeManager() *theme.Manager {
	return a.themeMgr
}

// IsThemeEnabled 检查主题系统是否启用
func (a *App) IsThemeEnabled() bool {
	return a.themeEnabled
}

// SetUserData 设置用户数据
func (a *App) SetUserData(key string, value interface{}) {
	a.userData[key] = value
}

// GetUserData 获取用户数据
func (a *App) GetUserData(key string) interface{} {
	return a.userData[key]
}

// ============================================================================
// 上下文管理
// ============================================================================

// Context 获取应用上下文
func (a *App) Context() context.Context {
	return a.contextMgr.Context()
}

// Shutdown 优雅关闭
func (a *App) Shutdown(timeout time.Duration) error {
	a.state = StateStopping
	return a.contextMgr.Shutdown(timeout)
}

// SetRoot 设置根组件
func (a *App) SetRoot(comp component.Node) {
	a.root = comp
	a.dirty = true

	// 使用 ComponentContext 注入运行时资源
	a.injectComponentContext(comp)
}

// createComponentContext 创建组件上下文
func (a *App) createComponentContext() *component.ComponentContext {
	ctx := component.NewComponentContext()
	ctx.SetDirtyCallback(func() {
		a.dirty = true
	})
	return ctx
}

// injectComponentContext 递归地为组件注入上下文
// 使用接口而非具体类型，遵循依赖倒置原则
func (a *App) injectComponentContext(node component.Node) {
	ctx := a.createComponentContext()

	// 1. 如果组件实现了 MountableWithContext，调用它
	if mountable, ok := node.(component.MountableWithContext); ok {
		// 获取父容器（如果有的话）
		var parent component.Container
		if p, ok := node.(interface{ GetParent() component.Container }); ok {
			parent = p.GetParent()
		}
		mountable.MountWithContext(parent, ctx)
		return
	}

	// 2. 兼容旧接口：通过辅助接口注入
	if dn, ok := node.(component.DirtyNotifiable); ok {
		dn.SetDirtyCallback(func() { a.dirty = true })
	}

	// 3. 递归处理子节点
	a.injectContextToChildren(node, ctx)
}

// injectContextToChildren 递归地为子节点注入上下文
func (a *App) injectContextToChildren(node component.Node, ctx *component.ComponentContext) {
	// 检查是否有子节点
	type childrenProvider interface {
		Children() []component.Node
	}

	if provider, ok := node.(childrenProvider); ok {
		for _, child := range provider.Children() {
			// 对子节点递归注入
			if mountable, ok := child.(component.MountableWithContext); ok {
				var parent component.Container
				if p, ok := child.(interface{ GetParent() component.Container }); ok {
					parent = p.GetParent()
				}
				mountable.MountWithContext(parent, ctx)
			} else if dn, ok := child.(component.DirtyNotifiable); ok {
				// 使用辅助接口设置 dirty callback
				dn.SetDirtyCallback(func() { a.dirty = true })
			}

			// 递归处理子节点的子节点
			a.injectContextToChildren(child, ctx)
		}
	}
}

// GetRoot 获取根组件
func (a *App) GetRoot() component.Node {
	return a.root
}

// OnKey 注册键盘事件处理
func (a *App) OnKey(key rune, handler func()) {
	a.keyMap.BindFunc(string(key), func(ev *frameworkevent.KeyEvent) {
		handler()
	})
}

// OnKeyCombo 注册快捷键处理
func (a *App) OnKeyCombo(combo string, handler func()) {
	a.keyMap.BindFunc(combo, func(ev *frameworkevent.KeyEvent) {
		handler()
	})
}

// OnEvent 注册事件处理
func (a *App) OnEvent(eventType frameworkevent.EventType, handler frameworkevent.EventHandler) func() {
	return a.router.Subscribe(eventType, handler)
}

// Init 初始化应用
func (a *App) Init() error {
	if a.state != StateCreated {
		return errors.New("app already initialized")
	}

	a.state = StateInitializing

	// 设置默认终端尺寸
	a.terminalWidth = 80
	a.terminalHeight = 24

	// 设置路由器
	a.setupRouter()

	// 创建并启动事件泵
	inputReader, err := platform.NewInputReader()
	if err != nil {
		return err
	}

	a.pump = frameworkevent.NewPump(inputReader)
	if err := a.pump.Start(); err != nil {
		return err
	}

	// 让根组件获得焦点
	if a.root != nil {
		if focusable, ok := a.root.(interface{ OnFocus() }); ok {
			focusable.OnFocus()
		}
	}

	a.state = StateRunning
	a.dirty = true

	return nil
}

// setupRouter 设置事件路由
func (a *App) setupRouter() {
	// 订阅退出事件
	a.router.Subscribe(frameworkevent.EventQuit, frameworkevent.EventHandlerFunc(func(ev frameworkevent.Event) bool {
		a.Quit()
		return true
	}))
}

// Run 运行应用
func (a *App) Run() error {
	// 启用 panic 恢复（如果配置）
	if a.recovery != nil {
		defer func() {
			if r := recover(); r != nil {
				a.recovery.Handle(r)
			}
		}()
	}

	if err := a.Init(); err != nil {
		return err
	}
	defer a.Close()

	// 主循环
	ticker := time.NewTicker(a.tickInterval)
	defer ticker.Stop()

	// 使用事件泵的通道
	eventChan := a.pump.Events()
	renderStartTime := time.Now()

	for a.state == StateRunning {
		select {
		case ev := <-eventChan:
			a.handleEvent(ev)

		case <-ticker.C:
			a.handleTick()

		case <-a.quit:
			a.state = StateStopping
			return nil
		case <-a.contextMgr.Context().Done():
			a.state = StateStopping
			return nil
		}

		// 渲染 - 使用节流器控制帧率
		if a.dirty && a.throttler.ShouldRender() {
			renderStartTime = time.Now()
			a.render()
			a.throttler.RecordFrameTime(time.Since(renderStartTime))
		}
	}

	return nil
}

// handleEvent 处理事件
func (a *App) handleEvent(ev frameworkevent.Event) {
	// 调试模式：记录事件
	if a.debugMode && a.debugRecorder != nil {
		a.debugRecorder.RecordEvent(ev)
	}

	// 通过事件过滤器处理
	if !a.eventFilter(ev) {
		// 事件被过滤器拦截
		return
	}

	// 路由事件
	a.router.Route(ev)

	// 窗口大小调整事件处理
	if ev.Type() == frameworkevent.EventResize {
		if resizeEv, ok := ev.(*frameworkevent.ResizeEvent); ok {
			a.Resize(resizeEv.NewWidth, resizeEv.NewHeight)
		}
		return
	}

	// 键盘事件处理
	if ev.Type() == frameworkevent.EventKeyPress {
		// 首先检查快捷键映射
		if keyEv, ok := ev.(*frameworkevent.KeyEvent); ok {
			if handler, found := a.keyMap.Lookup(keyEv); found {
				if handler.HandleEvent(ev) {
					a.dirty = true
					return
				}
			}
		}

		// 然后发送到根组件
		if a.root != nil {
			// 使用 event.Component 接口检查，而不是匿名接口
			// 这样可以避免类型别名导致的类型断言失败
			if handler, ok := a.root.(frameworkevent.Component); ok {
				if handler.HandleEvent(ev) {
					a.dirty = true
				}
			}
		}
		return
	}

	// 如果有目标组件，分发到组件
	if target := ev.Target(); target != nil {
		if handler, ok := target.(frameworkevent.Component); ok {
			if handler.HandleEvent(ev) {
				a.dirty = true
			}
		}
	}
}

// handleTick 处理定时器
// 光标闪烁现在由 TextInput.Paint 自己处理，不需要外部 Tick
func (a *App) handleTick() {
	// 定期触发重绘以支持光标闪烁
	// TextInput 会在 Paint 时自己检查时间并切换光标状态
	a.dirty = true
}

// render 渲染界面
func (a *App) render() {
	if a.root == nil {
		return
	}

	// 使用 V3 Paintable 接口渲染
	if paintable, ok := a.root.(component.Paintable); ok {
		buf := paint.NewBuffer(a.terminalWidth, a.terminalHeight)

		ctx := component.PaintContext{
			AvailableWidth:  a.terminalWidth,
			AvailableHeight: a.terminalHeight,
			X:               0,
			Y:               0,
		}

		paintable.Paint(ctx, buf)

		// 调试模式：记录渲染状态
		if a.debugMode && a.debugRecorder != nil {
			a.debugRecorder.RecordRender(buf)
		}

		// 将缓冲区内容输出到终端
		// 使用环境变量控制输出模式：
		// TUI_OUTPUT_MODE=direct  使用全量刷新（绕过差异比较）
		// TUI_OUTPUT_MODE=diff    使用差异比较优化（默认）
		outputMode := os.Getenv("TUI_OUTPUT_MODE")
		if outputMode == "direct" {
			a.outputBufferDirect(buf)
			// a.outputBuffer(buf)
		} else {
			// 默认使用差异比较优化
			a.outputBuffer(buf)
		}
	}

	a.dirty = false
}

// outputBuffer 输出缓冲区到终端（局部刷新优化版）
// 使用 output_diff.go 中的函数来处理差异比较和 ANSI 格式化
func (a *App) outputBuffer(buf *paint.Buffer) {
	// 调整 prevBuffer 大小（如果需要）
	a.prevBuffer = EnsurePrevBufferSize(a.prevBuffer, buf.Width, buf.Height)

	// 比较新旧 buffer，获取变化列表
	diffResult := CompareBuffers(buf, a.prevBuffer, a.lastCursorX, a.lastCursorY)

	// 更新光标位置
	a.lastCursorX = diffResult.CursorX
	a.lastCursorY = diffResult.CursorY

	// 如果没有变化，跳过输出
	if !diffResult.HasChanges {
		return
	}

	// 调试模式：记录输出
	if a.debugMode && a.debugRecorder != nil && os.Getenv("TUI_OUTPUT_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[OUTPUT] %d changes detected\n", len(diffResult.Changes))
	}

	// 排序变化（从上到下，从左到右）
	SortChanges(diffResult.Changes)

	// 格式化为 ANSI 输出
	output := FormatChangesAsANSI(buf, BufferDiffResult{
		Changes:    diffResult.Changes,
		CursorX:    diffResult.CursorX,
		CursorY:    diffResult.CursorY,
		HasChanges: true,
	}, a.firstRender)

	// 一次性输出
	fmt.Print(output)

	// 更新 prevBuffer - 在所有输出完成后更新
	UpdatePrevBuffer(a.prevBuffer, buf)

	// 清除首次渲染标记
	if a.firstRender {
		a.firstRender = false
	}
}

// outputBufferDirect 输出缓冲区到终端（全量刷新版）
// 每次都输出完整的屏幕内容，不使用局部刷新优化
func (a *App) outputBufferDirect(buf *paint.Buffer) {
	var output bytes.Buffer

	// 首次渲染时清屏
	if a.firstRender {
		output.WriteString("\x1b[2J")  // 清屏
		a.firstRender = false
	}

	// 隐藏终端光标
	output.WriteString("\x1b[?25l")

	// 调试模式：记录输出
	if a.debugMode && a.debugRecorder != nil && os.Getenv("TUI_OUTPUT_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[OUTPUT DIRECT] about to write %d cells to terminal\n", buf.Height*buf.Width)
	}

	// 移动光标到左上角
	output.WriteString("\x1b[1;1H")

	// 跟踪当前样式
	var currentStyle style.Style

	// 构建输出内容 - 输出所有单元格
	for y := 0; y < buf.Height; y++ {
		skipNextCell := false
		for x := 0; x < buf.Width; x++ {
			// 如果上一个字符是宽字符，跳过它占据的下一个单元格
			if skipNextCell {
				skipNextCell = false
				continue
			}

			cell := buf.Cells[y][x]

			// 跳过宽字符的填充单元格 (Width == 0)
			if cell.Width == 0 {
				continue
			}

			// 设置字符
			char := cell.Char
			if char == 0 {
				char = ' '
			}

			// 应用样式（如果改变）
			if cell.Style != currentStyle {
				if currentStyle != (style.Style{}) {
					output.WriteString("\x1b[0m")
				}
				if cell.Style != (style.Style{}) {
					output.WriteString(cell.Style.ToANSI())
				}
				currentStyle = cell.Style
			}

			output.WriteRune(char)
			// 如果是宽字符，标记跳过下一个单元格
			skipNextCell = (cell.Width == 2)
		}

		// 行末重置样式并换行（除了最后一行）
		if y < buf.Height-1 {
			if currentStyle != (style.Style{}) {
				output.WriteString("\x1b[0m")
				currentStyle = style.Style{}
			}
			output.WriteString("\r\n")
		}
	}

	// 重置样式
	if currentStyle != (style.Style{}) {
		output.WriteString("\x1b[0m")
	}

	// 移动光标到末尾（避免残留）
	output.WriteString(fmt.Sprintf("\x1b[%d;%dH", buf.Height, 1))

	// 一次性输出
	fmt.Print(output.String())
}

// Quit 退出应用
func (a *App) Quit() {
	select {
	case a.quit <- struct{}{}:
	default:
	}
}

// Close 关闭应用
func (a *App) Close() error {
	a.state = StateStopped

	// 让根组件失去焦点
	if a.root != nil {
		if focusable, ok := a.root.(interface{ OnBlur() }); ok {
			focusable.OnBlur()
		}
	}

	// 停止事件泵
	if a.pump != nil {
		a.pump.Stop()
	}

	// 调试模式：保存日志
	if a.debugMode && a.debugRecorder != nil {
		if err := a.debugRecorder.DumpToFile(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save debug log: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Debug log saved\n")
		}
	}

	// 显示终端光标
	a.ShowCursor()

	// 关闭 panic 恢复管理器
	if a.recovery != nil {
		a.recovery.Close()
	}

	return nil
}

// ============================================================================
// Terminal 接口实现（用于 Panic Recovery）
// ============================================================================

// SetNormalMode 恢复终端正常模式
func (a *App) SetNormalMode() {
	// TUI framework 使用事件泵，这里不需要额外操作
}

// ShowCursor 显示光标
func (a *App) ShowCursor() {
	fmt.Print("\x1b[?25h")
}

// ExitAltScreen 退出备用屏幕
func (a *App) ExitAltScreen() {
	fmt.Print("\x1b[?1049l")
}

// EnableEcho 启用回显
func (a *App) EnableEcho() {
	// 事件泵会处理回显
}

// Flush 刷新输出
func (a *App) Flush() {
	os.Stdout.Sync()
}

// GetState 获取应用状态
func (a *App) GetState() AppState {
	return a.state
}

// IsRunning 检查是否在运行
func (a *App) IsRunning() bool {
	return a.state == StateRunning
}

// SetTickInterval 设置定时器间隔
func (a *App) SetTickInterval(interval time.Duration) {
	a.tickInterval = interval
}

// GetSize 获取终端尺寸
func (a *App) GetSize() (width, height int) {
	return a.terminalWidth, a.terminalHeight
}

// Resize 调整尺寸
func (a *App) Resize(width, height int) {
	sizeChanged := a.terminalWidth != width || a.terminalHeight != height
	a.terminalWidth = width
	a.terminalHeight = height
	a.dirty = true

	// 尺寸变化时清屏，避免残留内容
	if sizeChanged && !a.firstRender {
		a.clearScreen()
	}
}

// clearScreen 清屏
func (a *App) clearScreen() {
	fmt.Print("\x1b[2J")  // 清屏
	fmt.Print("\x1b[H")   // 移动光标到左上角
}
