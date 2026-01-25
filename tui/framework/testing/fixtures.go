package testing

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yaoapp/yao/tui/runtime"
	"github.com/yaoapp/yao/tui/runtime/action"
	"github.com/yaoapp/yao/tui/runtime/ai"
	"github.com/yaoapp/yao/tui/runtime/focus"
	"github.com/yaoapp/yao/tui/runtime/state"
)

// =============================================================================
// Fixtures (V3)
// =============================================================================
// Fixtures 测试夹具工具
// 用于加载测试应用和组件

// Fixture 测试夹具
type Fixture struct {
	root         *runtime.LayoutNode
	dispatcher   *action.Dispatcher
	tracker      *state.Tracker
	focusMgr     *focus.Manager
	controller    *ai.RuntimeController
	teardownFuncs []func()
	cleanupFuncs  []func() // 资源清理函数
}

// NewFixture 创建测试夹具
func NewFixture() *Fixture {
	dispatcher := action.NewDispatcher()
	tracker := state.NewTracker()
	focusMgr := focus.NewManager(nil)

	controller := ai.NewRuntimeController(dispatcher, tracker, focusMgr)

	return &Fixture{
		dispatcher:   dispatcher,
		tracker:      tracker,
		focusMgr:     focusMgr,
		controller:    controller,
		teardownFuncs: make([]func(), 0),
		cleanupFuncs:  make([]func(), 0),
	}
}

// =============================================================================
// Getter 方法
// =============================================================================

// Root 获取根节点
func (f *Fixture) Root() *runtime.LayoutNode {
	return f.root
}

// Dispatcher 获取 Action 分发器
func (f *Fixture) Dispatcher() *action.Dispatcher {
	return f.dispatcher
}

// Tracker 获取状态追踪器
func (f *Fixture) Tracker() *state.Tracker {
	return f.tracker
}

// FocusManager 获取焦点管理器
func (f *Fixture) FocusManager() *focus.Manager {
	return f.focusMgr
}

// Controller 获取 AI 控制器
func (f *Fixture) Controller() *ai.RuntimeController {
	return f.controller
}

// =============================================================================
// 生命周期管理
// =============================================================================

// Setup 设置夹具
func (f *Fixture) Setup() error {
	// 默认实现为空，子类可以覆盖
	return nil
}

// Teardown 清理夹具
func (f *Fixture) Teardown() {
	// 执行清理函数（倒序）
	for i := len(f.teardownFuncs) - 1; i >= 0; i-- {
		f.teardownFuncs[i]()
	}
	f.teardownFuncs = nil

	// 执行资源清理函数（倒序）
	for i := len(f.cleanupFuncs) - 1; i >= 0; i-- {
		f.cleanupFuncs[i]()
	}
	f.cleanupFuncs = nil

	// 清理焦点管理器
	if f.focusMgr != nil {
		f.focusMgr.Clear()
	}
}

// Defer 注册清理函数
func (f *Fixture) Defer(fn func()) {
	f.teardownFuncs = append(f.teardownFuncs, fn)
}

// DeferCleanup 注册资源清理函数
func (f *Fixture) DeferCleanup(fn func()) {
	f.cleanupFuncs = append(f.cleanupFuncs, fn)
}

// =============================================================================
// 快捷操作
// =============================================================================

// MustDispatch 分发 Action，失败则 panic
func (f *Fixture) MustDispatch(a *action.Action) {
	if !f.dispatcher.Dispatch(a) {
		panic(fmt.Sprintf("action not handled: %s", a.Type))
	}
}

// Snapshot 获取当前状态快照
func (f *Fixture) Snapshot() *state.Snapshot {
	return f.tracker.Current()
}

// RefreshFocusables 刷新可聚焦组件
func (f *Fixture) RefreshFocusables() {
	if f.focusMgr != nil {
		f.focusMgr.RefreshFocusables()
	}
}

// =============================================================================
// 应用加载
// =============================================================================

// LoadApp 从文件加载测试应用
func LoadApp(appPath string) (*Fixture, error) {
	// 检查文件是否存在
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("app file not found: %s", appPath)
	}

	fixture := NewFixture()

	// 这里应该解析应用文件并创建组件树
	// 具体实现依赖于应用文件的格式

	return fixture, nil
}

// LoadAppFromDir 从目录加载测试应用
func LoadAppFromDir(dir string) (*Fixture, error) {
	appFiles, err := filepath.Glob(filepath.Join(dir, "*.yao"))
	if err != nil {
		return nil, fmt.Errorf("failed to find app files: %w", err)
	}

	if len(appFiles) == 0 {
		return nil, fmt.Errorf("no app files found in directory: %s", dir)
	}

	// 使用找到的第一个应用文件
	return LoadApp(appFiles[0])
}

// =============================================================================
// 测试数据构建
// =============================================================================

// NewTestRootNode 创建测试根节点
func NewTestRootNode() *runtime.LayoutNode {
	return &runtime.LayoutNode{
		ID:   "root",
		Type: "root",
		Children: make([]*runtime.LayoutNode, 0),
	}
}

// NewTestNode 创建测试节点
func NewTestNode(id, typ string) *runtime.LayoutNode {
	return &runtime.LayoutNode{
		ID:   id,
		Type: runtime.NodeType(typ),
		Children: make([]*runtime.LayoutNode, 0),
	}
}

// AddChild 添加子节点
func AddChild(parent, child *runtime.LayoutNode) *runtime.LayoutNode {
	parent.Children = append(parent.Children, child)
	return parent
}

// =============================================================================
// 常用测试场景
// =============================================================================

// SimpleFormFixture 创建简单表单测试夹具
func SimpleFormFixture() *Fixture {
	fixture := NewFixture()

	root := NewTestRootNode()
	form := NewTestNode("form", "Form")
	username := NewTestNode("username", "TextInput")
	password := NewTestNode("password", "TextInput")
	submit := NewTestNode("submit", "Button")

	AddChild(form, username)
	AddChild(form, password)
	AddChild(form, submit)
	AddChild(root, form)

	fixture.root = root

	return fixture
}

// SimpleListFixture 创建简单列表测试夹具
func SimpleListFixture() *Fixture {
	fixture := NewFixture()

	root := NewTestRootNode()
	list := NewTestNode("list", "List")

	// 添加列表项
	for i := 0; i < 10; i++ {
		item := NewTestNode(fmt.Sprintf("item-%d", i), "ListItem")
		AddChild(list, item)
	}

	AddChild(root, list)
	fixture.root = root

	return fixture
}

// SimpleTableFixture 创建简单表格测试夹具
func SimpleTableFixture() *Fixture {
	fixture := NewFixture()

	root := NewTestRootNode()
	table := NewTestNode("table", "Table")

	// 添加列
	for i := 0; i < 3; i++ {
		col := NewTestNode(fmt.Sprintf("col-%d", i), "TableColumn")
		AddChild(table, col)
	}

	// 添加行
	for i := 0; i < 5; i++ {
		row := NewTestNode(fmt.Sprintf("row-%d", i), "TableRow")
		AddChild(table, row)
	}

	AddChild(root, table)
	fixture.root = root

	return fixture
}

// ModalDialogFixture 创建模态对话框测试夹具
func ModalDialogFixture() *Fixture {
	fixture := NewFixture()

	root := NewTestRootNode()
	mainContent := NewTestNode("main", "Box")
	dialog := NewTestNode("dialog", "Dialog")
	dialogOverlay := NewTestNode("dialog-overlay", "Overlay")

	AddChild(dialogOverlay, dialog)
	AddChild(root, mainContent)
	AddChild(root, dialogOverlay)

	fixture.root = root

	return fixture
}

// =============================================================================
// 状态构建
// =============================================================================

// WithComponentState 添加组件状态
func (f *Fixture) WithComponentState(id, typ string, stateFn func(*state.ComponentState)) *Fixture {
	snapshot := f.tracker.Current()
	comp := state.ComponentState{
		ID:    id,
		Type:  typ,
		State: make(map[string]interface{}),
		Props: make(map[string]interface{}),
		Rect:  state.Rect{X: 0, Y: 0, Width: 10, Height: 1},
	}

	if stateFn != nil {
		stateFn(&comp)
	}

	snapshot.SetComponent(comp)
	return f
}

// WithFocus 设置焦点
func (f *Fixture) WithFocus(path ...string) *Fixture {
	snapshot := f.tracker.Current()
	snapshot.SetFocus(path)
	return f
}

// =============================================================================
// 预定义测试数据
// =============================================================================

// TestUser 测试用户数据
var TestUser = map[string]interface{}{
	"username": "testuser@example.com",
	"password": "secret123",
	"name":     "Test User",
}

// TestConfig 测试配置数据
var TestConfig = map[string]interface{}{
	"timeout":    30,
	"max_retries": 3,
	"debug":      false,
}

// TestListItems 测试列表项
var TestListItems = []string{
	"Item 1",
	"Item 2",
	"Item 3",
	"Item 4",
	"Item 5",
}

// TestTableData 测试表格数据
var TestTableData = [][]string{
	{"Name", "Email", "Role"},
	{"Alice", "alice@example.com", "Admin"},
	{"Bob", "bob@example.com", "User"},
	{"Charlie", "charlie@example.com", "User"},
}
