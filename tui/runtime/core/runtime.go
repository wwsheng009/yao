package core

import (
	"context"
	"sync"
	"time"

	"github.com/yaoapp/yao/tui/runtime/platform"
	"github.com/yaoapp/yao/tui/runtime/action"
	"github.com/yaoapp/yao/tui/runtime/focus"
	"github.com/yaoapp/yao/tui/runtime/input"
	"github.com/yaoapp/yao/tui/runtime/layout"
	"github.com/yaoapp/yao/tui/runtime/paint"
	"github.com/yaoapp/yao/tui/runtime/state"
)

// ==============================================================================
// Runtime Core (V3)
// ==============================================================================
// Runtime 核心引擎，整合布局、绘制、状态、焦点、输入等子系统

// Runtime 运行时核心
type Runtime struct {
	mu sync.RWMutex

	// Platform 平台抽象（屏幕、输入、信号）
	platform platform.RuntimePlatform

	// LayoutEngine 布局引擎
	layoutEngine *layout.Engine

	// FocusManager 焦点管理器
	focusManager *focus.ManagerV3

	// StateTracker 状态追踪器
	stateTracker *state.Tracker

	// ActionDispatcher Action 分发器
	actionDispatcher *action.Dispatcher

	// KeyMap 输入映射
	keyMap *input.KeyMap

	// ContextManager 上下文管理器
	contextManager *ContextManager

	// Root 根节点
	root layout.Node

	// Buffer 绘制缓冲区
	buffer *paint.Buffer

	// DirtyTracker 脏区域跟踪器
	dirtyTracker *paint.DirtyTracker

	// Running 是否运行中
	running bool

	// WindowSize 窗口尺寸
	windowWidth  int
	windowHeight int
}

// NewRuntime 创建运行时
func NewRuntime(pf platform.RuntimePlatform) *Runtime {
	return &Runtime{
		platform:         pf,
		layoutEngine:     layout.NewEngine(),
		focusManager:     focus.NewManagerV3(),
		stateTracker:     state.NewTracker(),
		actionDispatcher: action.NewDispatcher(),
		keyMap:           input.NewKeyMap(),
		contextManager:   NewContextManager(context.Background()),
		dirtyTracker:     paint.NewDirtyTracker(),
		running:          false,
	}
}

// SetRoot 设置根节点
func (r *Runtime) SetRoot(node layout.Node) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.root = node

	// 更新可聚焦组件
	r.focusManager.UpdateFocusables(node)
}

// GetRoot 获取根节点
func (r *Runtime) GetRoot() layout.Node {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.root
}

// Start 启动运行时
func (r *Runtime) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.running {
		return nil
	}

	// 初始化平台
	if err := r.platform.Init(); err != nil {
		return err
	}

	// 获取窗口尺寸
	width, height := r.platform.Size()
	r.windowWidth = width
	r.windowHeight = height

	// 创建缓冲区
	r.buffer = paint.NewBuffer(width, height)

	// 初始化焦点域
	rootScope := focus.NewScope("root", "root")
	if r.root != nil {
		rootScope.SetFocusable(r.focusManager.CollectFocusable(r.root))
	}
	r.focusManager.PushScope(rootScope)

	r.running = true
	return nil
}

// Stop 停止运行时
func (r *Runtime) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return nil
	}

	r.running = false

	// 关闭平台
	return r.platform.Close()
}

// IsRunning 检查是否运行中
func (r *Runtime) IsRunning() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.running
}

// Update 更新状态
func (r *Runtime) Update() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running || r.root == nil {
		return nil
	}

	// 1. 记录更新前状态
	before := r.stateTracker.BeforeAction()

	// 2. 处理输入
	if err := r.processInput(); err != nil {
		return err
	}

	// 3. 记录更新后状态
	r.stateTracker.AfterAction(before)

	// 4. 布局
	if err := r.layout(); err != nil {
		return err
	}

	// 5. 标记脏区域
	r.markDirty()

	return nil
}

// Render 渲染到屏幕
func (r *Runtime) Render() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return nil
	}

	// 1. 清空缓冲区（或只清空脏区域）
	if r.dirtyTracker.IsAllDirty() {
		r.buffer = paint.NewBuffer(r.windowWidth, r.windowHeight)
	}

	// 2. 绘制组件
	if r.root != nil {
		r.paintNode(r.root, 0, 0)
	}

	// 3. 将缓冲区内容写入屏幕
	r.writeToScreen()

	// 4. 清除脏标记
	r.dirtyTracker.Clear()

	return nil
}

// ProcessInput 处理输入
func (r *Runtime) ProcessInput() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.processInput()
}

// processInput 处理输入（内部方法）
func (r *Runtime) processInput() error {
	// 从平台读取输入
	rawInput := r.platform.ReadInput()
	if rawInput == nil {
		return nil
	}

	// 处理不同类型的输入
	switch rawInput.Type {
	case platform.InputKeyPress:
		// 键盘输入：使用 KeyMap 转换
		act := r.keyMap.Map(*rawInput)
		if act == nil {
			return nil
		}

		// 设置目标为当前焦点
		focusID, _ := r.focusManager.GetFocused()
		act.Target = focusID

		// 分发 Action
		r.actionDispatcher.Dispatch(act)

	case platform.InputMouse:
		// 鼠标输入：需要命中测试
		r.processMouseInput(*rawInput)

	case platform.InputResize:
		// 窗口大小变化
		r.HandleWindowSize(rawInput.Width, rawInput.Height)
	}

	return nil
}

// processMouseInput 处理鼠标输入
func (r *Runtime) processMouseInput(input platform.RawInput) {
	if r.root == nil {
		return
	}

	// 使用命中测试找到目标组件
	// 注意：这里需要导入 event 包，暂时使用简单的坐标查找
	targetID := r.findComponentAt(input.MouseX, input.MouseY)

	// 创建鼠标 Action
	var act *action.Action
	switch input.MouseAction {
	case platform.MouseWheelUp:
		act = action.NewAction(action.ActionMouseWheelUp)
	case platform.MouseWheelDown:
		act = action.NewAction(action.ActionMouseWheelDown)
	case platform.MousePress:
		act = action.NewAction(action.ActionMouseClick)
	case platform.MouseRelease:
		act = action.NewAction(action.ActionMouseRelease)
	case platform.MouseMotion:
		act = action.NewAction(action.ActionMouseMotion)
	}

	if act != nil {
		act.Target = targetID
		act = act.WithPayload(MouseEventPayload{
			X:     input.MouseX,
			Y:     input.MouseY,
			Button: input.MouseButton,
		})
		r.actionDispatcher.Dispatch(act)
	}
}

// MouseEventPayload 鼠标事件负载
type MouseEventPayload struct {
	X      int
	Y      int
	Button platform.MouseButton
}

// findComponentAt 在指定坐标查找组件
func (r *Runtime) findComponentAt(x, y int) string {
	if r.root == nil {
		return ""
	}

	// 简单遍历查找（也可以使用 event.HitTest）
	return r.findComponentAtNode(r.root, x, y)
}

// findComponentAtNode 递归查找组件
func (r *Runtime) findComponentAtNode(node layout.Node, x, y int) string {
	if node == nil {
		return ""
	}

	nodeX, nodeY := node.GetPosition()
	nodeWidth, nodeHeight := node.GetSize()

	if nodeWidth <= 0 || nodeHeight <= 0 {
		return ""
	}

	// 检查点是否在节点内
	inBounds := x >= nodeX && x < nodeX+nodeWidth &&
		y >= nodeY && y < nodeY+nodeHeight

	if !inBounds {
		return ""
	}

	// 先检查子节点
	children := node.Children()
	for i := len(children) - 1; i >= 0; i-- {
		if childID := r.findComponentAtNode(children[i], x, y); childID != "" {
			return childID
		}
	}

	// 返回当前节点
	return node.ID()
}

// layout 执行布局
func (r *Runtime) layout() error {
	if r.root == nil {
		return nil
	}

	constraints := layout.NewConstraints(
		0, r.windowWidth,
		0, r.windowHeight,
	)

	result := r.layoutEngine.Layout([]layout.Node{r.root}, constraints)

	// 更新根节点尺寸
	if len(result.Boxes) > 0 {
		rootBox := result.Boxes[0]
		r.root.SetSize(rootBox.Width, rootBox.Height)
		r.root.SetPosition(rootBox.X, rootBox.Y)
	}

	return nil
}

// paintNode 绘制节点
func (r *Runtime) paintNode(node layout.Node, x, y int) {
	// 创建绘制上下文
	bounds := paint.Rect{
		X:      x,
		Y:      y,
		Width:  node.GetWidth(),
		Height: node.GetHeight(),
	}
	ctx := paint.NewPaintContext(r.buffer, bounds)

	// 检查是否有焦点
	focusPath := r.focusManager.GetFocusPath()
	ctx = ctx.WithFocusPath(focusPath)
	ctx.Focused = focusPath.Current() == node.ID()

	// 如果节点实现了 Paintable 接口，绘制它
	type PaintableNode interface {
		Paint(paint.PaintContext, *paint.Buffer)
	}
	if paintable, ok := node.(PaintableNode); ok {
		paintable.Paint(*ctx, r.buffer)
	}

	// 递归绘制子节点
	for _, child := range node.Children() {
		childX, childY := child.GetPosition()
		r.paintNode(child, childX, childY)
	}
}

// markDirty 标记脏区域
func (r *Runtime) markDirty() {
	// 简化实现：每次都标记全部为脏
	// 实际实现应该比较前后状态差异
	r.dirtyTracker.MarkAll()
}

// writeToScreen 将缓冲区写入屏幕
func (r *Runtime) writeToScreen() {
	// 获取脏区域
	if r.dirtyTracker.IsAllDirty() {
		// 全屏重绘
		r.writeFullBuffer()
	} else {
		// 部分重绘
		r.writeDirtyRegions()
	}
}

// writeFullBuffer 写入完整缓冲区
func (r *Runtime) writeFullBuffer() {
	// 简化实现：逐行写入
	for y := 0; y < r.buffer.Height; y++ {
		row := make([]rune, r.buffer.Width)
		for x := 0; x < r.buffer.Width; x++ {
			row[x] = r.buffer.Cells[y][x].Char
		}
		// 写入平台
		r.platform.WriteString(string(row))
	}
}

// writeDirtyRegions 写入脏区域
func (r *Runtime) writeDirtyRegions() {
	rects := r.dirtyTracker.GetDirtyRects()
	for _, rect := range rects {
		for y := rect.Y; y < rect.Y+rect.Height; y++ {
			if y < 0 || y >= r.buffer.Height {
				continue
			}
			for x := rect.X; x < rect.X+rect.Width; x++ {
				if x < 0 || x >= r.buffer.Width {
					continue
				}
				// 写入单个单元格
				char := r.buffer.Cells[y][x].Char
				r.platform.WriteString(string(char))
			}
		}
	}
}

// HandleWindowSize 处理窗口大小变化
func (r *Runtime) HandleWindowSize(width, height int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.windowWidth = width
	r.windowHeight = height

	// 重新创建缓冲区
	r.buffer = paint.NewBuffer(width, height)

	// 标记全部为脏
	r.dirtyTracker.MarkAll()
}

// GetState 获取当前状态
func (r *Runtime) GetState() *state.Snapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stateTracker.Current()
}

// GetFocusManager 获取焦点管理器
func (r *Runtime) GetFocusManager() *focus.ManagerV3 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.focusManager
}

// GetActionDispatcher 获取 Action 分发器
func (r *Runtime) GetActionDispatcher() *action.Dispatcher {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.actionDispatcher
}

// RegisterActionTarget 注册 Action 目标
func (r *Runtime) RegisterActionTarget(target action.Target) {
	r.actionDispatcher.Register(target)
}

// UnregisterActionTarget 注销 Action 目标
func (r *Runtime) UnregisterActionTarget(id string) {
	r.actionDispatcher.Unregister(id)
}

// SubscribeGlobalAction 订阅全局 Action
func (r *Runtime) SubscribeGlobalAction(actionType action.ActionType, handler action.Handler) func() {
	return r.actionDispatcher.Subscribe(actionType, handler)
}

// SetDefaultActionHandler 设置默认 Action 处理器
func (r *Runtime) SetDefaultActionHandler(handler action.Handler) {
	r.actionDispatcher.SetDefaultHandler(handler)
}

// PushFocusScope 推入焦点域
func (r *Runtime) PushFocusScope(scope *focus.Scope) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.focusManager.PushScope(scope)
}

// PopFocusScope 弹出焦点域
func (r *Runtime) PopFocusScope() *focus.Scope {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.focusManager.PopScope()
}

// FocusNext 移动焦点到下一个
func (r *Runtime) FocusNext() (string, bool) {
	return r.focusManager.FocusNext()
}

// FocusPrev 移动焦点到上一个
func (r *Runtime) FocusPrev() (string, bool) {
	return r.focusManager.FocusPrev()
}

// FocusSpecific 设置焦点到指定组件
func (r *Runtime) FocusSpecific(id string) bool {
	return r.focusManager.FocusSpecific(id)
}

// GetFocused 获取当前聚焦的组件ID
func (r *Runtime) GetFocused() (string, bool) {
	return r.focusManager.GetFocused()
}

// GetFocusPath 获取焦点路径
func (r *Runtime) GetFocusPath() state.FocusPath {
	return r.focusManager.GetFocusPath()
}

// Invalidate 作废布局缓存
func (r *Runtime) Invalidate() {
	r.layoutEngine.Invalidate()
	r.dirtyTracker.MarkAll()
}

// InvalidateNode 作废特定节点的布局缓存
func (r *Runtime) InvalidateNode(id string) {
	r.layoutEngine.InvalidateNode(id)
}

// Run 主运行循环
func (r *Runtime) Run() error {
	if err := r.Start(); err != nil {
		return err
	}

	defer r.Stop()

	for r.IsRunning() {
		// 检查上下文是否已取消
		select {
		case <-r.Context().Done():
			return r.Context().Err()
		default:
		}

		if err := r.Update(); err != nil {
			return err
		}
		if err := r.Render(); err != nil {
			return err
		}
	}

	return nil
}

// ==============================================================================
// Context Support Methods
// ==============================================================================

// Context 返回运行时上下文
func (r *Runtime) Context() context.Context {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.contextManager.Context()
}

// WithContext 设置父上下文
func (r *Runtime) WithContext(ctx context.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.contextManager = NewContextManager(ctx)
}

// Shutdown 优雅关闭运行时
func (r *Runtime) Shutdown(timeout ...time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 停止运行
	r.running = false

	// 关闭平台
	if err := r.platform.Close(); err != nil {
		return err
	}

	// 关闭上下文管理器
	return r.contextManager.Shutdown(timeout...)
}

// Go 在运行时上下文中启动 goroutine
func (r *Runtime) Go(f func(context.Context) error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	r.contextManager.Go(f)
}

// ContextValue 获取上下文值
func (r *Runtime) ContextValue(key ContextKey) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.contextManager.Value(key)
}

// SetContextValue 设置上下文值
func (r *Runtime) SetContextValue(key ContextKey, value interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.contextManager.WithValue(key, value)
}

// IsCanceled 检查上下文是否已取消
func (r *Runtime) IsCanceled() bool {
	select {
	case <-r.Context().Done():
		return true
	default:
		return false
	}
}
