package framework

import (
	"errors"
	"time"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/screen"
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
	// 屏幕管理
	screen *screen.Manager

	// 组件树
	root component.Component

	// 事件
	events chan event.Event
	router *event.Router
	keyMap *event.KeyMap

	// 生命周期
	state AppState
	quit  chan struct{}
	dirty bool

	// 配置
	tickInterval time.Duration
}

// NewApp 创建新应用
func NewApp() *App {
	return &App{
		events:       make(chan event.Event, 100),
		router:       event.NewRouter(),
		keyMap:       event.NewKeyMap(),
		quit:         make(chan struct{}),
		tickInterval: 16 * time.Millisecond, // ~60fps
	}
}

// SetRoot 设置根组件
func (a *App) SetRoot(comp component.Component) {
	a.root = comp
	a.dirty = true
}

// GetRoot 获取根组件
func (a *App) GetRoot() component.Component {
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

	// 初始化屏幕管理器 (默认尺寸)
	a.screen = screen.NewManager(80, 24)
	if err := a.screen.Init(); err != nil {
		return err
	}

	// 设置路由器
	a.setupRouter()

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

	for a.state == StateRunning {
		select {
		case ev := <-a.events:
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

	// 如果有目标组件，分发到组件
	if target := ev.Target(); target != nil {
		target.HandleEvent(ev)
		a.dirty = true
	}
}

// handleTick 处理定时器
func (a *App) handleTick() {
	// 处理动画等周期性任务
}

// render 渲染界面
func (a *App) render() {
	if a.root == nil {
		return
	}

	// 创建缓冲区
	width, height := a.screen.GetSize()
	buf := screen.NewBuffer(width, height)

	// 渲染根组件
	ctx := component.NewRenderContext(width, height)
	content := a.root.Render(ctx)

	// 将内容写入缓冲区
	lines := splitLines(content)
	for y, line := range lines {
		if y < height {
			buf.SetLine(y, line, a.root.GetStyle())
		}
	}

	// 输出到屏幕
	a.screen.Render(buf)

	a.dirty = false
}

func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}

	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
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

	// 关闭屏幕
	if a.screen != nil {
		return a.screen.Close()
	}

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
	if a.screen != nil {
		return a.screen.GetSize()
	}
	return 80, 24
}

// Resize 调整尺寸
func (a *App) Resize(width, height int) {
	if a.screen != nil {
		a.screen.SetSize(width, height)
	}
	a.dirty = true
}
