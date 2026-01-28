// Package testing 提供端到端测试支持
package testing

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/runtime/style"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// =============================================================================
// E2E 测试框架 - 模拟真实用户交互
// =============================================================================

// TestScenario 测试场景定义
type TestScenario struct {
	Name        string
	Description string
	Setup       func(*TestContext)
	Steps       []TestStep
	Timeout     time.Duration
}

// TestStep 测试步骤
type TestStep struct {
	Name        string
	Action      TestAction
	Assertions  []Assertion
	Wait        time.Duration
}

// TestAction 测试动作
type TestAction interface {
	Execute(*TestContext) error
}

// Assertion 断言
type Assertion func(*TestContext) error

// TestContext 测试上下文
type TestContext struct {
	// 组件树
	Root component.Node

	// 当前焦点组件
	Focused component.Node

	// 事件历史（用于调试）
	EventHistory []EventRecord

	// 状态快照历史
	StateHistory []StateSnapshot

	// 渲染缓冲区历史
	RenderHistory []RenderSnapshot

	// 输出缓冲区（用于调试）
	Output *bytes.Buffer

	// 测试选项
	Options TestOptions

	mu sync.RWMutex
}

// TestOptions 测试选项
type TestOptions struct {
	// 是否记录详细日志
	Verbose bool

	// 是否在失败时输出渲染缓冲区
	DebugRender bool

	// 是否记录事件历史
	RecordEvents bool

	// 自定义超时
	DefaultTimeout time.Duration
}

// EventRecord 事件记录
type EventRecord struct {
	Timestamp time.Time
	EventType string
	Source    string
	Target    string
	Details   map[string]interface{}
}

// StateSnapshot 状态快照
type StateSnapshot struct {
	Timestamp   time.Time
	ComponentID string
	Properties  map[string]interface{}
}

// RenderSnapshot 渲染快照
type RenderSnapshot struct {
	Timestamp time.Time
	Width     int
	Height    int
	Cells     [][]CellSnapshot
}

// CellSnapshot 单元格快照
type CellSnapshot struct {
	Char rune
	Style style.Style
}

// =============================================================================
// 预定义动作
// =============================================================================

// KeyPress 按键动作
type KeyPress struct {
	Key     rune
	Special event.SpecialKey
	Mods    event.Modifier
}

func (k KeyPress) Execute(ctx *TestContext) error {
	ctx.recordEvent("KeyPress", map[string]interface{}{
		"key":     string(k.Key),
		"special": specialKeyToString(k.Special),
	})

	// 检查根组件是否有 HandleEvent 方法
	if handler, ok := ctx.Root.(event.EventComponent); ok {
		ev := event.NewKeyEvent(event.Key{
			Rune: k.Key,
			Name: k.Special.String(),
		})
		if k.Special != event.KeyUnknown {
			ev.Special = k.Special
		}
		ev.Modifiers = k.Mods
		if !handler.HandleEvent(ev) {
			return fmt.Errorf("key press not handled: %v", k)
		}
		ctx.captureState()
		ctx.captureRender()
		return nil
	}
	return fmt.Errorf("root component does not support HandleEvent")
}

// FocusComponent 聚焦组件动作
type FocusComponent struct {
	ComponentID string
}

func (f FocusComponent) Execute(ctx *TestContext) error {
	ctx.recordEvent("Focus", map[string]interface{}{
		"target": f.ComponentID,
	})

	// 查找组件
	target := ctx.findComponent(f.ComponentID)
	if target == nil {
		return fmt.Errorf("component not found: %s", f.ComponentID)
	}

	// 调用 OnFocus
	if focusable, ok := target.(interface{ OnFocus() }); ok {
		focusable.OnFocus()
		ctx.mu.Lock()
		ctx.Focused = target
		ctx.mu.Unlock()
		ctx.captureState()
		ctx.captureRender()
		return nil
	}

	return fmt.Errorf("component does not support focus: %s", f.ComponentID)
}

// SetComponentValue 设置组件值动作
type SetComponentValue struct {
	ComponentID string
	Value       interface{}
}

func (s SetComponentValue) Execute(ctx *TestContext) error {
	ctx.recordEvent("SetValue", map[string]interface{}{
		"target": s.ComponentID,
		"value":  s.Value,
	})

	target := ctx.findComponent(s.ComponentID)
	if target == nil {
		return fmt.Errorf("component not found: %s", s.ComponentID)
	}

	// 尝试不同的值设置接口
	if setter, ok := target.(interface{ SetValue(string) }); ok {
		if str, ok := s.Value.(string); ok {
			setter.SetValue(str)
			ctx.captureState()
			ctx.captureRender()
			return nil
		}
	}

	if setter, ok := target.(interface{ SetValue(interface{}) error }); ok {
		if err := setter.SetValue(s.Value); err != nil {
			return err
		}
		ctx.captureState()
		ctx.captureRender()
		return nil
	}

	return fmt.Errorf("component does not support SetValue: %s", s.ComponentID)
}

// Wait 等待动作
type Wait struct {
	Duration time.Duration
}

func (w Wait) Execute(ctx *TestContext) error {
	ctx.recordEvent("Wait", map[string]interface{}{
		"duration": w.Duration.String(),
	})
	time.Sleep(w.Duration)
	ctx.captureState()
	return nil
}

// =============================================================================
// 预定义断言
// =============================================================================

// ComponentValue 断言组件值
func ComponentValue(componentID string, expected interface{}) Assertion {
	return func(ctx *TestContext) error {
		target := ctx.findComponent(componentID)
		if target == nil {
			return fmt.Errorf("component not found: %s", componentID)
		}

		if getter, ok := target.(interface{ GetValue() string }); ok {
			actual := getter.GetValue()
			expectedStr := fmt.Sprint(expected)
			if actual != expectedStr {
				return fmt.Errorf("value mismatch for %s: expected '%s', got '%s'",
					componentID, expectedStr, actual)
			}
			return nil
		}

		return fmt.Errorf("component does not support GetValue: %s", componentID)
	}
}

// ComponentFocused 断言组件是否聚焦
func ComponentFocused(componentID string) Assertion {
	return func(ctx *TestContext) error {
		ctx.mu.RLock()
		focused := ctx.Focused
		ctx.mu.RUnlock()

		if focused == nil {
			return fmt.Errorf("no component is focused")
		}

		if focused.ID() != componentID {
			return fmt.Errorf("focused component mismatch: expected %s, got %s",
				componentID, focused.ID())
		}
		return nil
	}
}

// ComponentVisible 断言组件是否可见
func ComponentVisible(componentID string, visible bool) Assertion {
	return func(ctx *TestContext) error {
		target := ctx.findComponent(componentID)
		if target == nil {
			return fmt.Errorf("component not found: %s", componentID)
		}

		if v, ok := target.(interface{ IsVisible() bool }); ok {
			if v.IsVisible() != visible {
				return fmt.Errorf("visibility mismatch for %s: expected %v, got %v",
					componentID, visible, v.IsVisible())
			}
			return nil
		}

		return fmt.Errorf("component does not support IsVisible: %s", componentID)
	}
}

// RenderContains 断言渲染内容包含指定文本
func RenderContains(text string) Assertion {
	return func(ctx *TestContext) error {
		if len(ctx.RenderHistory) == 0 {
			return fmt.Errorf("no render snapshots available")
		}

		lastRender := ctx.RenderHistory[len(ctx.RenderHistory)-1]
		content := ctx.renderToString(lastRender)

		if !strings.Contains(content, text) {
			return fmt.Errorf("render does not contain '%s'\nActual content:\n%s",
				text, content)
		}
		return nil
	}
}

// CursorPosition 断言光标位置
func CursorPosition(componentID string, expectedPos int) Assertion {
	return func(ctx *TestContext) error {
		target := ctx.findComponent(componentID)
		if target == nil {
			return fmt.Errorf("component not found: %s", componentID)
		}

		if cursor, ok := target.(interface{ GetCursor() int }); ok {
			actual := cursor.GetCursor()
			if actual != expectedPos {
				return fmt.Errorf("cursor position mismatch for %s: expected %d, got %d",
					componentID, expectedPos, actual)
			}
			return nil
		}

		return fmt.Errorf("component does not support GetCursor: %s", componentID)
	}
}

// =============================================================================
// TestContext 方法
// =============================================================================

// NewTestContext 创建测试上下文
func NewTestContext(root component.Node, opts TestOptions) *TestContext {
	return &TestContext{
		Root:          root,
		Output:        &bytes.Buffer{},
		Options:       opts,
		EventHistory:  make([]EventRecord, 0),
		StateHistory:  make([]StateSnapshot, 0),
		RenderHistory: make([]RenderSnapshot, 0),
	}
}

// RunScenario 运行测试场景
func (ctx *TestContext) RunScenario(scenario TestScenario) error {
	if ctx.Options.Verbose {
		ctx.Output.WriteString(fmt.Sprintf("\n=== Running Scenario: %s ===\n", scenario.Name))
		ctx.Output.WriteString(fmt.Sprintf("Description: %s\n\n", scenario.Description))
	}

	// 执行设置
	if scenario.Setup != nil {
		scenario.Setup(ctx)
	}

	// 执行步骤
	for i, step := range scenario.Steps {
		if ctx.Options.Verbose {
			ctx.Output.WriteString(fmt.Sprintf("Step %d: %s\n", i+1, step.Name))
		}

		// 执行动作（如果有）
		if step.Action != nil {
			if err := step.Action.Execute(ctx); err != nil {
				return fmt.Errorf("step %d failed: %w", i+1, err)
			}
		}

		// 等待
		if step.Wait > 0 {
			time.Sleep(step.Wait)
		}

		// 执行断言
		for j, assert := range step.Assertions {
			if err := assert(ctx); err != nil {
				ctx.dumpDebugInfo()
				return fmt.Errorf("step %d, assertion %d failed: %w", i+1, j+1, err)
			}
		}

		if ctx.Options.Verbose {
			ctx.Output.WriteString(fmt.Sprintf("  ✓ Pass\n"))
		}
	}

	if ctx.Options.Verbose {
		ctx.Output.WriteString(fmt.Sprintf("\n=== Scenario Passed ===\n\n"))
	}
	return nil
}

// findComponent 查找组件（递归）
func (ctx *TestContext) findComponent(id string) component.Node {
	return ctx.findComponentRecursive(ctx.Root, id)
}

func (ctx *TestContext) findComponentRecursive(node component.Node, id string) component.Node {
	if node.ID() == id {
		return node
	}

	// 检查是否有子节点（标准 Children 方法）
	if childrenProvider, ok := node.(interface{ Children() []component.Node }); ok {
		for _, child := range childrenProvider.Children() {
			if found := ctx.findComponentRecursive(child, id); found != nil {
				return found
			}
		}
	}

	// 特殊处理 Form 类型（通过反射或接口访问嵌套的 Input）
	// 检查是否有 GetFields 方法
	if getFields, ok := node.(interface {
		GetFields() map[string]interface{}
	}); ok {
		for _, field := range getFields.GetFields() {
			// 检查字段是否有 Input 属性
			if fieldWithInput, ok := field.(interface {
				Input() component.Node
			}); ok {
				if fieldWithInput.Input() != nil && fieldWithInput.Input().ID() == id {
					return fieldWithInput.Input()
				}
			}
		}
	}

	return nil
}

// recordEvent 记录事件
func (ctx *TestContext) recordEvent(eventType string, details map[string]interface{}) {
	if !ctx.Options.RecordEvents {
		return
	}

	ctx.mu.Lock()
	ctx.EventHistory = append(ctx.EventHistory, EventRecord{
		Timestamp: time.Now(),
		EventType: eventType,
		Details:   details,
	})
	ctx.mu.Unlock()
}

// captureState 捕获状态快照
func (ctx *TestContext) captureState() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	ctx.captureStateRecursive(ctx.Root, "")
}

func (ctx *TestContext) captureStateRecursive(node component.Node, path string) {
	currentPath := path + "/" + node.ID()

	// 捕获基本状态
	snapshot := StateSnapshot{
		Timestamp:   time.Now(),
		ComponentID: currentPath,
		Properties:  make(map[string]interface{}),
	}

	// 尝试获取常见状态
	if v, ok := node.(interface{ GetValue() string }); ok {
		snapshot.Properties["value"] = v.GetValue()
	}
	if v, ok := node.(interface{ GetCursor() int }); ok {
		snapshot.Properties["cursor"] = v.GetCursor()
	}
	if v, ok := node.(interface{ IsFocused() bool }); ok {
		snapshot.Properties["focused"] = v.IsFocused()
	}
	if v, ok := node.(interface{ IsVisible() bool }); ok {
		snapshot.Properties["visible"] = v.IsVisible()
	}

	ctx.StateHistory = append(ctx.StateHistory, snapshot)

	// 递归捕获子节点
	if childrenProvider, ok := node.(interface{ Children() []component.Node }); ok {
		for _, child := range childrenProvider.Children() {
			ctx.captureStateRecursive(child, currentPath)
		}
	}
}

// captureRender 捕获渲染快照
func (ctx *TestContext) captureRender() {
	if paintable, ok := ctx.Root.(component.Paintable); ok {
		// 创建一个合理大小的缓冲区
		buf := paint.NewBuffer(80, 24)

		paintCtx := component.NewPaintContext(buf, 0, 0, 80, 24)

		paintable.Paint(paintCtx, buf)

		// 捕获快照
		snapshot := RenderSnapshot{
			Timestamp: time.Now(),
			Width:     buf.Width,
			Height:    buf.Height,
			Cells:     make([][]CellSnapshot, buf.Height),
		}

		for y := 0; y < buf.Height; y++ {
			snapshot.Cells[y] = make([]CellSnapshot, buf.Width)
			for x := 0; x < buf.Width; x++ {
				cell := buf.Cells[y][x]
				snapshot.Cells[y][x] = CellSnapshot{
					Char:  cell.Char,
					Style: cell.Style,
				}
			}
		}

		ctx.mu.Lock()
		ctx.RenderHistory = append(ctx.RenderHistory, snapshot)
		ctx.mu.Unlock()
	}
}

// dumpDebugInfo 输出调试信息
func (ctx *TestContext) dumpDebugInfo() {
	if !ctx.Options.DebugRender {
		return
	}

	ctx.Output.WriteString("\n=== Debug Information ===\n\n")

	// 输出事件历史
	if len(ctx.EventHistory) > 0 {
		ctx.Output.WriteString("Event History:\n")
		for i, ev := range ctx.EventHistory {
			ctx.Output.WriteString(fmt.Sprintf("  %d. [%s] %s %v\n",
				i+1, ev.Timestamp.Format("15:04:05.000"), ev.EventType, ev.Details))
		}
		ctx.Output.WriteString("\n")
	}

	// 输出状态历史
	if len(ctx.StateHistory) > 0 {
		ctx.Output.WriteString("State History (last snapshot):\n")
		last := ctx.StateHistory[len(ctx.StateHistory)-1]
		ctx.Output.WriteString(fmt.Sprintf("  Component: %s\n", last.ComponentID))
		for k, v := range last.Properties {
			ctx.Output.WriteString(fmt.Sprintf("    %s: %v\n", k, v))
		}
		ctx.Output.WriteString("\n")
	}

	// 输出渲染快照
	if len(ctx.RenderHistory) > 0 {
		ctx.Output.WriteString("Render Snapshot:\n")
		last := ctx.RenderHistory[len(ctx.RenderHistory)-1]
		ctx.Output.WriteString(ctx.renderToString(last))
		ctx.Output.WriteString("\n")
	}
}

// renderToString 将渲染快照转换为字符串
func (ctx *TestContext) renderToString(snapshot RenderSnapshot) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("Dimensions: %dx%d\n", snapshot.Width, snapshot.Height))
	buf.WriteString("┌" + strings.Repeat("─", snapshot.Width) + "┐\n")

	for y := 0; y < snapshot.Height; y++ {
		buf.WriteString("│")
		for x := 0; x < snapshot.Width; x++ {
			cell := snapshot.Cells[y][x]
			if cell.Char == 0 {
				buf.WriteString(" ")
			} else {
				buf.WriteString(string(cell.Char))
			}
		}
		buf.WriteString("│\n")
	}

	buf.WriteString("└" + strings.Repeat("─", snapshot.Width) + "┘\n")

	// 输出带有样式标记的版本
	buf.WriteString("\nWith Style Markers:\n")
	buf.WriteString("  [R] = Reverse (cursor)\n")
	buf.WriteString("  [B] = Bold\n")
	buf.WriteString("  •••••••••••••••\n")

	for y := 0; y < min(snapshot.Height, 10); y++ {  // 限制输出行数
		buf.WriteString(fmt.Sprintf("%2d│", y))
		for x := 0; x < min(snapshot.Width, 80); x++ {  // 限制输出列数
			cell := snapshot.Cells[y][x]
			char := cell.Char
			if char == 0 {
				char = ' '
			}

			// 添加样式标记
			if cell.Style.IsReverse() {
				buf.WriteString(fmt.Sprintf("\033[7m%c\033[0m", char))  // 反白显示
			} else if cell.Style.IsBold() {
				buf.WriteString(fmt.Sprintf("\033[1m%c\033[0m", char))  // 粗体显示
			} else {
				buf.WriteString(string(char))
			}
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

// GetOutput 获取测试输出
func (ctx *TestContext) GetOutput() string {
	return ctx.Output.String()
}

// WriteTo 实现 io.Writer
func (ctx *TestContext) WriteTo(w io.Writer) (int64, error) {
	return io.Copy(w, ctx.Output)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// =============================================================================
// 辅助函数
// =============================================================================

// specialKeyToString 返回 SpecialKey 的字符串表示
func specialKeyToString(s event.SpecialKey) string {
	switch s {
	case event.KeyUnknown:
		return "Unknown"
	case event.KeyEscape:
		return "Escape"
	case event.KeyEnter:
		return "Enter"
	case event.KeyBackspace:
		return "Backspace"
	case event.KeyTab:
		return "Tab"
	case event.KeyUp:
		return "Up"
	case event.KeyDown:
		return "Down"
	case event.KeyLeft:
		return "Left"
	case event.KeyRight:
		return "Right"
	case event.KeyHome:
		return "Home"
	case event.KeyEnd:
		return "End"
	case event.KeyDelete:
		return "Delete"
	default:
		return fmt.Sprintf("Special(%d)", int(s))
	}
}
