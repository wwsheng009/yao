package framework

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/input"
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

	// 配置
	tickInterval time.Duration
}

// NewApp 创建新应用
func NewApp() *App {
	return &App{
		router:       event.NewRouter(),
		keyMap:       event.NewKeyMap(),
		quit:         make(chan struct{}),
		tickInterval: 16 * time.Millisecond, // ~60fps
		firstRender:  true,
	}
}

// SetRoot 设置根组件
func (a *App) SetRoot(comp component.Node) {
	a.root = comp
	a.dirty = true
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

	// 设置光标闪烁的脏回调
	input.SetDirtyCallback(func() {
		a.dirty = true
	})

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
func (a *App) handleTick() {
	// 更新光标闪烁
	input.CursorBlinkTick()
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

		// 将缓冲区内容输出到终端
		a.outputBuffer(buf)
	}

	a.dirty = false
}

// outputBuffer 输出缓冲区到终端
func (a *App) outputBuffer(buf *paint.Buffer) {
	var output bytes.Buffer

	// 首次渲染时清屏，后续只移动光标到左上角
	if a.firstRender {
		output.WriteString("\x1b[2J")  // 清屏
		a.firstRender = false
	}
	// 移动光标到左上角，隐藏终端光标
	output.WriteString("\x1b[H\x1b[?25l")

	// 跟踪当前样式，避免重复输出
	var currentStyle style.Style

	// 构建输出内容
	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y][x]
			char := cell.Char
			if char == 0 {
				char = ' '
			}

			// 检查样式是否改变
			if cell.Style != currentStyle {
				// 先重置之前的样式
				if currentStyle != (style.Style{}) {
					output.WriteString("\x1b[0m")
				}
				// 应用新样式
				if cell.Style != (style.Style{}) {
					output.WriteString(cell.Style.ToANSI())
				}
				currentStyle = cell.Style
			}

			output.WriteRune(char)
		}
		// 重置样式
		if currentStyle != (style.Style{}) {
			output.WriteString("\x1b[0m")
			currentStyle = style.Style{}
		}
		if y < buf.Height-1 {
			output.WriteString("\r\n")
		}
	}

	// 一次性输出整个内容
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
