package framework

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/debug"
	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/platform"
	"github.com/yaoapp/yao/tui/framework/style"
	"github.com/yaoapp/yao/tui/runtime/paint"
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
	router *event.Router
	keyMap *event.KeyMap
	pump   *event.Pump

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
	debugMode      bool
	debugRecorder  *debug.Recorder
	debugLogFile   string
}

// NewApp 创建新应用
func NewApp() *App {
	return &App{
		router:       event.NewRouter(),
		keyMap:       event.NewKeyMap(),
		quit:         make(chan struct{}),
		tickInterval: 16 * time.Millisecond, // ~60fps
		firstRender:  true,
		debugMode:    os.Getenv("TUI_DEBUG") == "true",
		debugLogFile: os.Getenv("TUI_DEBUG_LOG"),
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
	a.keyMap.BindFunc(string(key), func(ev *event.KeyEvent) {
		handler()
	})
}

// OnKeyCombo 注册快捷键处理
func (a *App) OnKeyCombo(combo string, handler func()) {
	a.keyMap.BindFunc(combo, func(ev *event.KeyEvent) {
		handler()
	})
}

// OnEvent 注册事件处理
func (a *App) OnEvent(eventType event.EventType, handler event.EventHandler) func() {
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

	a.pump = event.NewPump(inputReader)
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
	a.router.Subscribe(event.EventClose, event.EventHandlerFunc(func(ev event.Event) bool {
		a.Quit()
		return true
	}))
}

// Run 运行应用
func (a *App) Run() error {
	if err := a.Init(); err != nil {
		return err
	}
	defer a.Close()

	// 主循环
	ticker := time.NewTicker(a.tickInterval)
	defer ticker.Stop()

	// 使用事件泵的通道
	eventChan := a.pump.Events()

	for a.state == StateRunning {
		select {
		case ev := <-eventChan:
			a.handleEvent(ev)

		case <-ticker.C:
			a.handleTick()

		case <-a.quit:
			a.state = StateStopping
			return nil
		}

		// 渲染
		if a.dirty {
			a.render()
		}
	}

	return nil
}

// handleEvent 处理事件
func (a *App) handleEvent(ev event.Event) {
	// 调试模式：记录事件
	if a.debugMode && a.debugRecorder != nil {
		a.debugRecorder.RecordEvent(ev)
	}

	// 路由事件
	a.router.Route(ev)

	// 键盘事件发送到根组件
	if ev.Type() == event.EventKeyPress {
		if a.root != nil {
			// 使用 duck typing 检查是否有 HandleEvent 方法
			if handler, ok := a.root.(interface{ HandleEvent(event.Event) bool }); ok {
				if handler.HandleEvent(ev) {
					a.dirty = true
				}
			}
		}
		return
	}

	// 如果有目标组件，分发到组件
	if target := ev.Target(); target != nil {
		if handler, ok := target.(interface{ HandleEvent(event.Event) bool }); ok {
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
		} else {
			// 默认使用差异比较优化
			a.outputBuffer(buf)
		}
	}

	a.dirty = false
}

// outputBuffer 输出缓冲区到终端（局部刷新优化版）
func (a *App) outputBuffer(buf *paint.Buffer) {
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
		fmt.Fprintf(os.Stderr, "[OUTPUT] about to write %d bytes to terminal\n", output.Len())
	}

	// 跟踪当前样式和位置
	var currentStyle style.Style
	var lastX, lastY int = 0, 0

	// 调整 prevBuffer 大小（如果需要）
	if a.prevBuffer == nil || len(a.prevBuffer) != buf.Height || len(a.prevBuffer[0]) != buf.Width {
		a.prevBuffer = make([][]paint.Cell, buf.Height)
		for y := 0; y < buf.Height; y++ {
			a.prevBuffer[y] = make([]paint.Cell, buf.Width)
		}
	}

	// 扫描缓冲区，查找光标位置（有反转样式的单元格）
	// 同时收集所有变化的单元格
	type cellChange struct {
		x, y  int
		force bool // 是否强制输出（用于光标）
	}
	changes := make([]cellChange, 0)
	currentCursorX, currentCursorY := -1, -1

	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			newCell := buf.Cells[y][x]
			oldCell := a.prevBuffer[y][x]

			// 检测光标位置（有反转样式的单元格）
			if newCell.Style.IsReverse() {
				currentCursorX = x
				currentCursorY = y
			}

			// 检查单元格是否改变
			cellChanged := newCell.Char != oldCell.Char || newCell.Style != oldCell.Style

			if cellChanged {
				changes = append(changes, cellChange{x: x, y: y, force: false})
			}
		}
	}

	// 如果光标位置改变，强制刷新新旧光标位置
	if currentCursorX != a.lastCursorX || currentCursorY != a.lastCursorY {
		// 旧光标位置需要刷新（清除反转样式）
		if a.lastCursorX >= 0 && a.lastCursorY >= 0 {
			changes = append(changes, cellChange{x: a.lastCursorX, y: a.lastCursorY, force: true})
		}
		// 新光标位置需要刷新（确保反转样式生效）
		if currentCursorX >= 0 && currentCursorY >= 0 {
			changes = append(changes, cellChange{x: currentCursorX, y: currentCursorY, force: true})
		}
		a.lastCursorX = currentCursorX
		a.lastCursorY = currentCursorY
	}

	// 如果没有变化，跳过输出
	if len(changes) == 0 {
		return
	}

	// 按位置排序变化（从上到下，从左到右）以优化输出
	for i := 0; i < len(changes)-1; i++ {
		for j := i + 1; j < len(changes); j++ {
			if changes[j].y < changes[i].y || (changes[j].y == changes[i].y && changes[j].x < changes[i].x) {
				changes[i], changes[j] = changes[j], changes[i]
			}
		}
	}

	// 构建输出内容
	currentY := -1
	for _, change := range changes {
		x, y := change.x, change.y
		newCell := buf.Cells[y][x]

		// 如果换行了，移动到新行的开头
		if y != currentY {
			output.WriteString(fmt.Sprintf("\x1b[%d;%dH", y+1, x+1))
			currentY = y
			lastX, lastY = x, y
		} else if x != lastX || y != lastY {
			// 同一行内，使用相对移动
			if x > lastX {
				output.WriteString(strings.Repeat("\x1b[C", x-lastX))
			} else if x < lastX {
				// 需要向左移动，使用绝对定位
				output.WriteString(fmt.Sprintf("\x1b[%d;%dH", y+1, x+1))
			}
			lastX, lastY = x, y
		}

		// 设置字符
		char := newCell.Char
		if char == 0 {
			char = ' '
		}

		// 应用样式（如果改变）
		if newCell.Style != currentStyle {
			if currentStyle != (style.Style{}) {
				output.WriteString("\x1b[0m")
			}
			if newCell.Style != (style.Style{}) {
				output.WriteString(newCell.Style.ToANSI())
			}
			currentStyle = newCell.Style
		}

		output.WriteRune(char)
		lastX++
	}

	// 更新 prevBuffer - 在所有输出完成后更新
	for y := 0; y < buf.Height; y++ {
		copy(a.prevBuffer[y], buf.Cells[y])
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
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y][x]

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
	fmt.Print("\x1b[?25h")

	return nil
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
	a.terminalWidth = width
	a.terminalHeight = height
	a.dirty = true
}
