//go:build ignore
// +build ignore

// 应用示例：在 framework 中使用 binding 模块
//
// 本文件展示如何在 framework 组件中集成数据绑定功能
//
// 使用说明：
// 本文件仅用于文档说明，不会被编译。
// 如需测试示例代码，请将其复制到实际的测试文件中。

package main

import (
	"fmt"

	"github.com/yaoapp/yao/tui/framework/binding"
	cb "github.com/yaoapp/yao/tui/framework/component/binding"
	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/runtime/action"
	"github.com/yaoapp/yao/tui/runtime/paint"
	"github.com/yaoapp/yao/tui/runtime/style"
)

// ==============================================================================
// 示例 1: 创建支持数据绑定的 Label 组件
// ==============================================================================

// Label 支持数据绑定的文本标签组件
type Label struct {
	*cb.BaseBindable

	// 文本属性可以是静态值或数据绑定
	textProp binding.Prop[string]

	// 对齐方式
	alignProp binding.Prop[string]
}

// NewLabel 创建静态文本标签
func NewLabel(text string) *Label {
	return &Label{
		BaseBindable: cb.NewBaseBindable("label"),
		textProp:     binding.NewStatic(text),
		alignProp:    binding.NewStatic("left"),
	}
}

// NewLabelBinding 创建数据绑定文本标签
func NewLabelBinding(textPath string) *Label {
	return &Label{
		BaseBindable: cb.NewBaseBindable("label"),
		textProp:     binding.NewBinding[string](textPath),
		alignProp:    binding.NewStatic("left"),
	}
}

// SetText 设置文本（自动检测绑定语法）
func (l *Label) SetText(text string) *Label {
	l.textProp = binding.NewStringProp(text)
	return l
}

// SetAlign 设置对齐方式
func (l *Label) SetAlign(align string) *Label {
	l.alignProp = binding.NewStatic(align)
	return l
}

// Paint 实现 Paintable 接口
func (l *Label) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	if !l.IsVisible() {
		return
	}

	// 创建绑定上下文
	bindCtx := cb.CreateBindingContext(l)

	// 解析属性
	align := l.alignProp.Resolve(bindCtx)
	text := l.textProp.Resolve(bindCtx)

	// 简化渲染：实际使用时应该使用完整样式
	width := ctx.AvailableWidth
	if width <= 0 {
		return
	}

	// 根据对齐方式计算 x 位置
	textLen := len([]rune(text))
	var x int
	switch align {
	case "center":
		x = (width - textLen) / 2
	case "right":
		x = width - textLen
	default: // left
		x = 0
	}

	// 绘制文本（使用默认样式）
	for i, r := range text {
		if x+i < width {
			ctx.SetCell(x+i, 0, r, style.Style{})
		}
	}
}

// ==============================================================================
// 示例 2: 创建响应式列表组件
// ==============================================================================

// ReactiveList 支持数据绑定的列表组件
type ReactiveList struct {
	*cb.BaseBindable

	// 数据源绑定
	itemsProp binding.Prop[[]interface{}]

	// 每项的渲染模板
	itemTemplate func(binding.Context) string
}

// NewReactiveList 创建响应式列表
func NewReactiveList(itemsPath string) *ReactiveList {
	return &ReactiveList{
		BaseBindable: cb.NewBaseBindable("list"),
		itemsProp:    binding.NewBinding[[]interface{}](itemsPath),
		itemTemplate: func(ctx binding.Context) string {
		// 默认模板
		if name, ok := ctx.Get("name"); ok {
			return fmt.Sprintf("- %v", name)
		}
		return "-"
	},
	}
}

// SetItemTemplate 设置列表项模板
func (rl *ReactiveList) SetItemTemplate(fn func(binding.Context) string) *ReactiveList {
	rl.itemTemplate = fn
	return rl
}

// Paint 实现绘制
func (rl *ReactiveList) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	if !rl.IsVisible() {
		return
	}

	// 创建绑定上下文
	bindCtx := cb.CreateBindingContext(rl)

	// 获取列表数据
	items := rl.itemsProp.Resolve(bindCtx)

	// 为每个列表项创建独立上下文
	contexts := binding.ListContext(bindCtx, items)

	// 渲染列表
	for i, itemCtx := range contexts {
		y := ctx.Y + i
		if y >= ctx.Y+ctx.AvailableHeight {
			break
		}

		text := rl.itemTemplate(itemCtx)
		for j, r := range text {
			ctx.SetCell(ctx.X+j, y, r, style.Style{})
		}
	}
}

// Measure 测量尺寸
func (rl *ReactiveList) Measure(maxWidth, maxHeight int) (width, height int) {
	bindCtx := cb.CreateBindingContext(rl)
	items := rl.itemsProp.Resolve(bindCtx)

	height = len(items)
	if maxHeight > 0 && height > maxHeight {
		height = maxHeight
	}

	width = maxWidth
	return width, height
}

// ==============================================================================
// 示例 3: 表单组件与 Store 集成
// ==============================================================================

// Form 表单组件，支持与 ReactiveStore 双向绑定
type Form struct {
	*cb.BaseBindable

	// 表单字段
	fields []FormField

	// 数据存储
	store *binding.ReactiveStore
}

type FormField struct {
	ID       string
	Label    string
	Input    component.Node
	BindPath string // Store 中的绑定路径
}

// NewForm 创建表单
func NewForm(store *binding.ReactiveStore) *Form {
	return &Form{
		BaseBindable: cb.NewBaseBindable("form"),
		store:        store,
		fields:       make([]FormField, 0),
	}
}

// AddField 添加表单字段
func (f *Form) AddField(id, label string, input component.Node, bindPath string) *Form {
	f.fields = append(f.fields, FormField{
		ID:       id,
		Label:    label,
		Input:    input,
		BindPath: bindPath,
	})
	return f
}

// Paint 绘制表单
func (f *Form) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	if !f.IsVisible() {
		return
	}

	// 绘制每个字段
	y := ctx.Y
	for _, field := range f.fields {
		// 绘制标签
		labelText := field.Label + ": "
		for i, r := range labelText {
			ctx.SetCell(ctx.X+i, y, r, style.Style{})
		}

		// 绘制输入组件
		if paintable, ok := field.Input.(interface{ Paint(component.PaintContext, *paint.Buffer) }); ok {
			// 创建子上下文用于输入组件
			inputCtx := component.PaintContext(*paint.NewPaintContext(buf, paint.Rect{
				X:      ctx.X + len(labelText),
				Y:      y,
				Width:  ctx.AvailableWidth - len(labelText),
				Height: 1,
			}))
			paintable.Paint(inputCtx, buf)
		}

		y++
	}
}

// Measure 测量表单尺寸
func (f *Form) Measure(maxWidth, maxHeight int) (width, height int) {
	height = len(f.fields)
	if maxHeight > 0 && height > maxHeight {
		height = maxHeight
	}
	width = maxWidth
	return width, height
}

// SyncFromStore 从 Store 同步数据到表单字段
func (f *Form) SyncFromStore() {
	bindCtx := f.store.ToContext()

	for _, field := range f.fields {
		if val, ok := bindCtx.Get(field.BindPath); ok {
			// 尝试设置输入组件的值
			if input, ok := field.Input.(interface{ SetValue(string) }); ok {
				input.SetValue(fmt.Sprintf("%v", val))
			}
		}
	}
}

// WatchStore 监听 Store 变化
func (f *Form) WatchStore() func() {
	cancels := make([]func(), 0)

	for _, field := range f.fields {
		cancel := f.store.Subscribe(field.BindPath, func(key string, old, new interface{}) {
			// Store 变化时同步到表单
			bindCtx := f.store.ToContext()
			if val, ok := bindCtx.Get(field.BindPath); ok {
				if input, ok := field.Input.(interface{ SetValue(string) }); ok {
					input.SetValue(fmt.Sprintf("%v", val))
				}
			}
			// 标记需要重绘
			f.MarkDirty()
		})
		cancels = append(cancels, cancel)
	}

	return func() {
		for _, cancel := range cancels {
			cancel()
		}
	}
}

// ==============================================================================
// 示例 4: 完整应用示例
// ==============================================================================

// CounterApp 计数器应用示例
type CounterApp struct {
	// 响应式存储
	store *binding.ReactiveStore

	// 组件
	label  *Label
	button *CounterButton
}

// CounterButton 计数按钮
type CounterButton struct {
	*cb.BaseBindable
	labelProp    binding.Prop[string]
	onClick      func()
}

func NewCounterButton(text string, onClick func()) *CounterButton {
	return &CounterButton{
		BaseBindable: cb.NewBaseBindable("button"),
		labelProp:    binding.NewStatic(text),
		onClick:      onClick,
	}
}

func (b *CounterButton) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	bindCtx := cb.CreateBindingContext(b)
	text := b.labelProp.Resolve(bindCtx)

	display := fmt.Sprintf("[ %s ]", text)
	for i, r := range display {
		ctx.SetCell(ctx.X+i, 0, r, style.Style{})
	}
}

func (b *CounterButton) HandleAction(a action.Action) bool {
	if a.Type == action.ActionMouseClick {
		if b.onClick != nil {
			b.onClick()
			return true
		}
	}
	return false
}

// NewCounterApp 创建计数器应用
func NewCounterApp() *CounterApp {
	// 创建响应式存储
	store := binding.NewReactiveStore()
	store.Set("count", 0)
	store.Set("label", "Count: 0")

	app := &CounterApp{
		store: store,
	}

	// 创建标签（数据绑定）
	app.label = NewLabelBinding("label")

	// 创建按钮
	app.button = NewCounterButton("Increment", func() {
		// 更新计数
		count, _ := store.Get("count")
		newCount := count.(int) + 1
		store.Set("count", newCount)

		// 更新标签文本
		store.Set("label", fmt.Sprintf("Count: %d", newCount))
	})

	// 监听变化，自动重绘
	store.Subscribe("label", func(key string, old, new interface{}) {
		if app.label != nil {
			app.label.MarkDirty()
		}
	})

	return app
}

// GetStore 获取存储（用于渲染时创建上下文）
func (a *CounterApp) GetStore() *binding.ReactiveStore {
	return a.store
}

// ==============================================================================
// 示例 5: DSL 解析与组件创建
// ==============================================================================

// ComponentFactory 从 DSL 配置创建组件
type ComponentFactory struct {
	store *binding.ReactiveStore
}

// NewComponentFactory 创建组件工厂
func NewComponentFactory(store *binding.ReactiveStore) *ComponentFactory {
	return &ComponentFactory{store: store}
}

// CreateFromDSL 从 DSL 创建组件
func (cf *ComponentFactory) CreateFromDSL(dsl map[string]interface{}) component.Node {
	typ, _ := dsl["type"].(string)

	switch typ {
	case "label":
		return cf.createLabel(dsl)
	case "list":
		return cf.createList(dsl)
	case "form":
		return cf.createForm(dsl)
	default:
		// 返回基础组件
		return component.NewBaseComponent(typ)
	}
}

// createLabel 创建 Label 组件
func (cf *ComponentFactory) createLabel(dsl map[string]interface{}) *Label {
	label := NewLabel("")

	// 解析属性
	if text, ok := dsl["text"]; ok {
		if textStr, ok := text.(string); ok {
			label.SetText(textStr)
		}
	}

	if align, ok := dsl["align"].(string); ok {
		label.SetAlign(align)
	}

	return label
}

// createList 创建 List 组件
func (cf *ComponentFactory) createList(dsl map[string]interface{}) *ReactiveList {
	var itemsPath string
	if path, ok := dsl["items"].(string); ok {
		itemsPath = path
	}

	list := NewReactiveList(itemsPath)

	// 设置模板
	if tmpl, ok := dsl["template"].(string); ok {
		list.SetItemTemplate(func(ctx binding.Context) string {
			// 简单的模板替换
			val, _ := ctx.Get(tmpl)
			return fmt.Sprintf("- %v", val)
		})
	}

	return list
}

// createForm 创建 Form 组件
func (cf *ComponentFactory) createForm(dsl map[string]interface{}) *Form {
	form := NewForm(cf.store)

	// 解析字段定义
	if fields, ok := dsl["fields"].([]interface{}); ok {
		for _, fieldDef := range fields {
			if fieldMap, ok := fieldDef.(map[string]interface{}); ok {
				id, _ := fieldMap["id"].(string)
				label, _ := fieldMap["label"].(string)
				bindPath, _ := fieldMap["bind"].(string)

				// 创建输入组件
				input := component.NewBaseComponent("input")

				form.AddField(id, label, input, bindPath)
			}
		}
	}

	return form
}

// ==============================================================================
// 使用示例
// ==============================================================================

// ExampleBasicUsage 基础使用示例
func ExampleBasicUsage() {
	// 1. 创建响应式存储
	store := binding.NewReactiveStore()
	store.Set("username", "alice")
	store.Set("status", "online")

	// 2. 创建数据上下文
	ctx := store.ToContext()

	// 3. 创建绑定属性
	_ = NewLabelBinding("status")
	_ = NewLabelBinding("username")

	// 4. 解析属性（实际使用时通过组件的 Paint 方法获取）
	// 这里直接从上下文获取值
	status, _ := ctx.Get("status")
	username, _ := ctx.Get("username")

	fmt.Printf("Status: %v, Username: %v\n", status, username)
}

// ExampleScopeChain 作用链示例
func ExampleScopeChain() {
	// 全局作用域
	global := binding.NewRootScope(map[string]interface{}{
		"app": map[string]interface{}{
			"name": "MyApp",
			"version": "1.0",
		},
	})

	// 用户作用域（继承全局）
	userScope := global.New(map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
		},
	})

	// 组件作用域（继承用户）
	componentScope := userScope.(*binding.Scope).New(map[string]interface{}{
		"component": "InputLabel",
	})

	// 可以访问所有层级的数据
	appName, _ := componentScope.Get("app.name")    // "MyApp"
	userName, _ := componentScope.Get("user.name")  // "Alice"
	compName, _ := componentScope.Get("component")    // "InputLabel"

	fmt.Printf("App: %s, User: %s, Component: %s\n", appName, userName, compName)
}

// ExampleListRendering 列表示例
func ExampleListRendering() {
	store := binding.NewReactiveStore()
	store.Set("items", []interface{}{
		map[string]interface{}{"id": 1, "name": "Task 1", "done": true},
		map[string]interface{}{"id": 2, "name": "Task 2", "done": false},
	})

	ctx := store.ToContext()
	items, _ := ctx.Get("items")

	// 为每个列表项创建独立上下文
	contexts := binding.ListContext(ctx, items.([]interface{}))

	for _, itemCtx := range contexts {
		index, _ := itemCtx.Get("$index")
		name, _ := itemCtx.Get("name")
		done, _ := itemCtx.Get("done")

		fmt.Printf("[%v] %v (done: %v)\n", index, name, done)
	}
}

// ExampleReactiveUpdate 响应式更新示例
func ExampleReactiveUpdate() {
	store := binding.NewReactiveStore()
	store.Set("count", 0)

	// 订阅变化
	cancel := store.Subscribe("count", func(key string, old, new interface{}) {
		fmt.Printf("Count changed: %v → %v\n", old, new)
	})

	// 更新数据
	store.Set("count", 1)  // 输出: Count changed: 0 → 1
	store.Set("count", 2)  // 输出: Count changed: 1 → 2

	cancel()
	store.Set("count", 3)  // 无输出（已取消订阅）
}

// ExampleDSLParsing DSL 解析示例
func ExampleDSLParsing() {
	factory := NewComponentFactory(binding.NewReactiveStore())

	// DSL 配置
	dslConfig := map[string]interface{}{
		"type":  "label",
		"text":  "{{ app.title }}",
		"align": "center",
	}

	// 创建组件
	_ = factory.createLabel(dslConfig)

	// 设置应用数据
	factory.store.Set("app", map[string]interface{}{
		"title": "Dashboard",
	})

	// 解析属性
	ctx := factory.store.ToContext()
	// 实际使用时，组件会在 Paint 方法中解析绑定
	// 这里直接从上下文获取值
	title, _ := ctx.Get("app.title")

	fmt.Printf("Label text: %v\n", title) // "Dashboard"
}
