package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/core"
	"github.com/yaoapp/yao/tui/runtime"
	"github.com/yaoapp/yao/tui/runtime/dsl"
	"github.com/yaoapp/yao/tui/ui/components"
)

// ===========================================================================
// Runtime 集成 - Model 扩展方法
// ===========================================================================
//
// 这个文件包含了 Model 与 Runtime 集成的所有方法
// 设计原则：
// 1. 不修改现有方法，只添加新方法
// 2. 通过 UseRuntime 开关控制是否使用 Runtime
// 3. 保持与 Legacy 代码的兼容性
//

// ========== Runtime 初始化 ==========

// initializeRuntime 初始化 Runtime 引擎
// 这个方法在 Init() 中被调用（当 UseRuntime 为 true 时）
func (m *Model) initializeRuntime() {
	log.Trace("Model.initializeRuntime: initializing runtime engine")

	// 创建 Runtime 引擎
	m.RuntimeEngine = runtime.NewRuntime(m.Width, m.Height)

	// 创建 DSL 工厂用于创建原生 Runtime 组件
	factory := dsl.NewFactory()

	// 转换 DSL 树为 Runtime 树（使用原生组件）
	// Config.Layout.Children 包含顶层组件
	if m.Config != nil && len(m.Config.Layout.Children) > 0 {
		m.RuntimeRoot = m.createRuntimeLayoutTree(factory, &m.Config.Layout)
		log.Trace("Model.initializeRuntime: runtime root node created with ID=%s, type=%s",
			m.RuntimeRoot.ID, m.RuntimeRoot.Type)
	}

	// 初始化组件到 Components 映射（用于交互式组件）
	m.initNativeComponents(m.RuntimeRoot, factory)
}

// createRuntimeLayoutTree 使用 DSL 工厂创建 Runtime 布局树
func (m *Model) createRuntimeLayoutTree(factory *dsl.Factory, layout *Layout) *runtime.LayoutNode {
	if layout == nil {
		return nil
	}

	// 确定方向
	direction := runtime.DirectionColumn
	if layout.Direction == "horizontal" || layout.Direction == "row" {
		direction = runtime.DirectionRow
	}

	// 创建根节点（使用 Column 或 Row）
	var rootNode runtime.Component
	if direction == runtime.DirectionRow {
		rootNode = components.NewRow()
	} else {
		rootNode = components.NewColumn()
	}

	// 处理子组件
	var children []*runtime.LayoutNode
	for _, child := range layout.Children {
		childNode := m.createRuntimeComponentNode(factory, &child)
		if childNode != nil {
			children = append(children, childNode)
		}
	}

	// 转换为 LayoutNode (root node has no DSL Component, so pass nil)
	rootLayoutNode := m.componentToLayoutNode(rootNode, "root", direction, children, nil)

	return rootLayoutNode
}

// createRuntimeComponentNode 为单个 DSL 组件创建 Runtime LayoutNode
func (m *Model) createRuntimeComponentNode(factory *dsl.Factory, comp *Component) *runtime.LayoutNode {
	if comp == nil {
		return nil
	}

	// 绑定模板数据 - 将 {{key}} 替换为 Config.Data 中的实际值
	boundProps := m.bindPropsToData(comp.Props)

	// Process bind attribute - add __bind_data to props for components with bind
	// This ensures list/table components get their data during creation
	if comp.Bind != "" {
		m.StateMu.RLock()
		if bindValue, ok := m.State[comp.Bind]; ok {
			if boundProps == nil {
				boundProps = make(map[string]interface{})
			}
			boundProps["__bind_data"] = bindValue
			log.Trace("createRuntimeComponentNode: component %s has bind '%s', added __bind_data with %d items",
				comp.ID, comp.Bind, len(bindValue.([]interface{})))
		} else {
			log.Warn("createRuntimeComponentNode: component %s has bind '%s' but key not found in State",
				comp.ID, comp.Bind)
		}
		m.StateMu.RUnlock()
	}

	// 创建原生 Runtime 组件
	dslConfig := &dsl.ComponentConfig{
		ID:    comp.ID,
		Type:  comp.Type,
		Props: boundProps,
	}

	// Debug log for bind processing
	if comp.Bind != "" {
		log.Trace("Creating component %s (type: %s) with bind '%s'. Props has __bind_data: %v",
			comp.ID, comp.Type, comp.Bind, boundProps != nil && boundProps["__bind_data"] != nil)
	}

	nativeComponent, err := factory.Create(dslConfig)
	if err != nil {
		log.Warn("createRuntimeComponentNode: failed to create component %s: %v", comp.ID, err)
		// 创建一个占位符文本组件
		nativeComponent = components.NewTextComponent("[Error: " + comp.Type + "]")
	}

	// 递归处理子组件
	var children []*runtime.LayoutNode
	if comp.Children != nil {
		for _, child := range comp.Children {
			childNode := m.createRuntimeComponentNode(factory, &child)
			if childNode != nil {
				children = append(children, childNode)
			}
		}
	}

	// 确定方向
	direction := runtime.DirectionColumn
	if comp.Direction == "horizontal" || comp.Direction == "row" {
		direction = runtime.DirectionRow
	}

	// 转换为 LayoutNode，传入 DSL Component 以应用 Width/Height
	layoutNode := m.componentToLayoutNode(nativeComponent, comp.ID, direction, children, comp)

	return layoutNode
}

// componentToLayoutNode 将原生 Runtime 组件转换为 LayoutNode
func (m *Model) componentToLayoutNode(comp runtime.Component, id string, direction runtime.Direction, children []*runtime.LayoutNode, dslComp *Component) *runtime.LayoutNode {
	// 根据组件类型创建对应的 LayoutNode
	var nodeType runtime.NodeType
	nodeStyle := runtime.NewStyle()

	switch comp.(type) {
	case *components.RowComponent:
		nodeType = runtime.NodeTypeRow
		// Extract properties from RowComponent
		if rowComp, ok := comp.(*components.RowComponent); ok {
			nodeStyle = runtime.NewStyle().
				WithDirection(runtime.DirectionRow).
				WithJustify(rowComp.GetJustify()).
				WithAlignItems(rowComp.GetAlign()).
				WithGap(rowComp.GetGap())
		} else {
			nodeStyle = runtime.NewStyle().WithDirection(runtime.DirectionRow)
		}
	case *components.ColumnComponent:
		nodeType = runtime.NodeTypeColumn
		// Extract properties from ColumnComponent
		if colComp, ok := comp.(*components.ColumnComponent); ok {
			nodeStyle = runtime.NewStyle().
				WithDirection(runtime.DirectionColumn).
				WithJustify(colComp.GetJustify()).
				WithAlignItems(colComp.GetAlign()).
				WithGap(colComp.GetGap())
		} else {
			nodeStyle = runtime.NewStyle().WithDirection(direction)
		}
	case *components.FlexComponent:
		nodeType = runtime.NodeTypeFlex
		nodeStyle = runtime.NewStyle().WithDirection(direction)
	case *components.InputComponent, *components.ButtonComponent, *components.TextComponent,
	     *components.HeaderComponent, *components.FooterComponent, *components.ListComponent,
	     *components.TableComponent, *components.FormComponent, *components.TextareaComponent,
	     *components.ProgressComponent, *components.SpinnerComponent:
		// 叶子组件
		nodeType = runtime.NodeTypeCustom
	default:
		nodeType = runtime.NodeTypeCustom
	}

	// Apply Width/Height from DSL Component if specified
	if dslComp != nil {
		if dslComp.Width != nil {
			if w, ok := dslComp.Width.(int); ok {
				nodeStyle.Width = w
			} else if w, ok := dslComp.Width.(float64); ok {
				nodeStyle.Width = int(w)
			} else if wStr, ok := dslComp.Width.(string); ok && wStr == "flex" {
				// "flex" means flex-grow: 1
				nodeStyle.FlexGrow = 1
			}
		}
		if dslComp.Height != nil {
			if h, ok := dslComp.Height.(int); ok {
				nodeStyle.Height = h
			} else if h, ok := dslComp.Height.(float64); ok {
				nodeStyle.Height = int(h)
			} else if hStr, ok := dslComp.Height.(string); ok && hStr == "flex" {
				// "flex" means flex-grow: 1
				nodeStyle.FlexGrow = 1
			}
		}

		// Parse justify property from Props
		if dslComp.Props != nil {
			if justifyVal, ok := dslComp.Props["justify"]; ok {
				if justifyStr, ok := justifyVal.(string); ok {
					nodeStyle.Justify = mapJustifyString(justifyStr)
				}
			}
			if alignVal, ok := dslComp.Props["alignItems"]; ok {
				if alignStr, ok := alignVal.(string); ok {
					nodeStyle.AlignItems = mapAlignString(alignStr)
				}
			}
		}
	}

	node := runtime.NewLayoutNode(id, nodeType, nodeStyle)

	// Apply Position from DSL Component if specified
	if dslComp != nil && dslComp.Style != nil {
		// Check for position type
		if posStr, ok := dslComp.Style["position"].(string); ok {
			switch posStr {
			case "absolute":
				node.Position.Type = runtime.PositionAbsolute
			case "relative":
				node.Position.Type = runtime.PositionRelative
			}
		}

		// Parse offsets for absolute positioning
		if node.Position.Type == runtime.PositionAbsolute {
			if top, ok := dslComp.Style["top"].(int); ok {
				node.Position.Top = &top
			} else if topF, ok := dslComp.Style["top"].(float64); ok {
				top := int(topF)
				node.Position.Top = &top
			}
			if left, ok := dslComp.Style["left"].(int); ok {
				node.Position.Left = &left
			} else if leftF, ok := dslComp.Style["left"].(float64); ok {
				left := int(leftF)
				node.Position.Left = &left
			}
			if right, ok := dslComp.Style["right"].(int); ok {
				node.Position.Right = &right
			} else if rightF, ok := dslComp.Style["right"].(float64); ok {
				right := int(rightF)
				node.Position.Right = &right
			}
			if bottom, ok := dslComp.Style["bottom"].(int); ok {
				node.Position.Bottom = &bottom
			} else if bottomF, ok := dslComp.Style["bottom"].(float64); ok {
				bottom := int(bottomF)
				node.Position.Bottom = &bottom
			}
		}
	}

	// Properly add children with parent references
	for _, child := range children {
		if child != nil {
			node.AddChild(child)
		}
	}

	// Set component's internal children field for container components
	// This is needed for the component's Measure method to work correctly
	switch c := comp.(type) {
	case *components.RowComponent:
		c.WithChildren(children...)
	case *components.ColumnComponent:
		c.WithChildren(children...)
	case *components.FlexComponent:
		c.WithChildren(children...)
	}

	// 包装为 ComponentInstance - 使用包装器确保实现了 core.ComponentInterface
	wrapper := &NativeComponentWrapper{Component: comp}
	node.Component = &core.ComponentInstance{
		Instance: wrapper,
		Type:     getComponentTypeString(comp),
	}

	return node
}

// getComponentTypeString 获取组件类型字符串
func getComponentTypeString(comp runtime.Component) string {
	switch comp.(type) {
	case *components.InputComponent:
		return "input"
	case *components.ButtonComponent:
		return "button"
	case *components.TextComponent:
		return "text"
	case *components.HeaderComponent:
		return "header"
	case *components.FooterComponent:
		return "footer"
	case *components.ListComponent:
		return "list"
	case *components.TableComponent:
		return "table"
	case *components.FormComponent:
		return "form"
	case *components.TextareaComponent:
		return "textarea"
	case *components.ProgressComponent:
		return "progress"
	case *components.SpinnerComponent:
		return "spinner"
	case *components.RowComponent:
		return "row"
	case *components.ColumnComponent:
		return "column"
	case *components.FlexComponent:
		return "flex"
	default:
		return "custom"
	}
}

// bindPropsToData binds template expressions in props to Config.Data
// This resolves {{key}} expressions to actual values
func (m *Model) bindPropsToData(props map[string]interface{}) map[string]interface{} {
	if props == nil {
		return nil
	}

	// Use the existing bindData function from expression_resolver.go
	bound := m.bindData(props)

	// Convert back to map[string]interface{}
	if result, ok := bound.(map[string]interface{}); ok {
		return result
	}
	return props
}

// NativeComponentWrapper 包装原生 Runtime 组件使其实现 core.ComponentInterface
// 同时也暴露 runtime.Measurable 接口
type NativeComponentWrapper struct {
	Component runtime.Component
}

// Init implements core.ComponentInterface
func (w *NativeComponentWrapper) Init() tea.Cmd {
	return nil
}

// UpdateMsg implements core.ComponentInterface
func (w *NativeComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	return w, nil, core.Ignored
}

// View implements core.ComponentInterface
func (w *NativeComponentWrapper) View() string {
	if c, ok := w.Component.(interface{ View() string }); ok {
		return c.View()
	}
	return ""
}

// Render implements core.ComponentInterface
func (w *NativeComponentWrapper) Render(config core.RenderConfig) (string, error) {
	if c, ok := w.Component.(interface{ Render(core.RenderConfig) (string, error) }); ok {
		return c.Render(config)
	}
	return w.View(), nil
}

// UpdateRenderConfig implements core.ComponentInterface
// This forwards the config to native Runtime components, handling data binding
func (w *NativeComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	if config.Data == nil {
		return nil
	}

	// Type-assert Data to map[string]interface{}
	dataMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return nil
	}

	// Handle data binding for list/table components
	// Check for __bind_data which contains the bound state data
	if bindData, ok := dataMap["__bind_data"]; ok {
		switch comp := w.Component.(type) {
		case *components.ListComponent:
			// List component: bind data should be the items array
			if items, ok := bindData.([]interface{}); ok {
				factory := dsl.NewFactory()
				parsedItems := factory.ParseListItems(items)
				comp.WithItems(parsedItems)
			}
		case *components.TableComponent:
			// Table component: bind data can be rows array or data array
			if rows, ok := bindData.([][]interface{}); ok {
				comp.WithData(rows)
			} else if data, ok := bindData.([]interface{}); ok {
				// Convert []interface{} to [][]interface{}
				// This handles both row arrays and object arrays
				converted := make([][]interface{}, 0, len(data))
				cols := comp.GetColumns()

				for _, rowIntf := range data {
					// Check if it's already a row ([]interface{})
					if rowSlice, ok := rowIntf.([]interface{}); ok {
						converted = append(converted, rowSlice)
					} else if objMap, ok := rowIntf.(map[string]interface{}); ok {
						// Convert object map to row using column keys
						row := make([]interface{}, len(cols))
						for i, col := range cols {
							if val, exists := objMap[col.Key]; exists {
								row[i] = val
							} else {
								row[i] = ""
							}
						}
						converted = append(converted, row)
					}
				}
				comp.WithData(converted)
			}
		}
	}

	// Handle other props
	if title, ok := dataMap["title"].(string); ok {
		if comp, ok := w.Component.(*components.ListComponent); ok {
			comp.WithTitle(title)
		}
		if comp, ok := w.Component.(*components.FormComponent); ok {
			comp.WithTitle(title)
		}
	}

	return nil
}

// Cleanup implements core.ComponentInterface
func (w *NativeComponentWrapper) Cleanup() {
	// Nothing to clean up
}

// GetID implements core.ComponentInterface
func (w *NativeComponentWrapper) GetID() string {
	if c, ok := w.Component.(interface{ GetID() string }); ok {
		return c.GetID()
	}
	return ""
}

// SetFocus implements core.ComponentInterface
func (w *NativeComponentWrapper) SetFocus(focused bool) {
	if c, ok := w.Component.(interface{ SetFocus(bool) }); ok {
		c.SetFocus(focused)
	}
}

// GetFocus implements core.ComponentInterface
func (w *NativeComponentWrapper) GetFocus() bool {
	if c, ok := w.Component.(interface{ GetFocus() bool }); ok {
		return c.GetFocus()
	}
	return false
}

// GetComponentType implements core.ComponentInterface
func (w *NativeComponentWrapper) GetComponentType() string {
	return getComponentTypeString(w.Component)
}

// GetStateChanges implements core.ComponentInterface
func (w *NativeComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	return nil, false
}

// GetSubscribedMessageTypes implements core.ComponentInterface
func (w *NativeComponentWrapper) GetSubscribedMessageTypes() []string {
	return nil
}

// Measure implements runtime.Measurable interface
// This delegates to the wrapped native component's Measure method
func (w *NativeComponentWrapper) Measure(c runtime.BoxConstraints) runtime.Size {
	if measurable, ok := w.Component.(interface{ Measure(runtime.BoxConstraints) runtime.Size }); ok {
		return measurable.Measure(c)
	}
	return runtime.Size{Width: 0, Height: 0}
}

// initNativeComponents 初始化原生组件到 Components 映射
// 这是用于事件处理和焦点管理
func (m *Model) initNativeComponents(node *runtime.LayoutNode, factory *dsl.Factory) {
	if node == nil {
		return
	}

	// 如果有组件实例，添加到 Components 映射
	if node.Component != nil && node.Component.Instance != nil {
		if m.Components == nil {
			m.Components = make(map[string]*core.ComponentInstance)
		}

		// 只添加交互式组件（使用现有的 isInteractiveComponent 函数）
		if isInteractiveComponent(node.Component.Type) {
			m.Components[node.ID] = node.Component
			// Also add to ComponentInstanceRegistry for consistency
			m.ComponentInstanceRegistry.UpdateComponent(node.ID, node.Component.Instance)
		}
	}

	// 递归处理子节点
	for _, child := range node.Children {
		m.initNativeComponents(child, factory)
	}
}

// ========== Runtime 渲染 ==========

// renderWithRuntime 使用 Runtime 渲染布局
// 这个方法替代 renderLayout() 当 UseRuntime 为 true 时
func (m *Model) renderWithRuntime() string {
	if m.RuntimeEngine == nil || m.RuntimeRoot == nil {
		log.Error("Model.renderWithRuntime: runtime not initialized")
		return "Runtime not initialized"
	}

	// 创建约束
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  m.Width,
		MinHeight: 0,
		MaxHeight: m.Height,
	}

	// Phase 1 & 2: Measure + Layout
	result := m.RuntimeEngine.Layout(m.RuntimeRoot, constraints)

	// IMPORTANT: Update component configs with bound data before rendering
	// This resolves the 'bind' property and populates '__bind_data' for list/table components
	m.updateComponentConfigsForRender(result)

	// Phase 3: Render
	frame := m.RuntimeEngine.Render(result)

	// 更新焦点列表（基于几何位置）
	m.updateFocusListFromRuntime(result)

	// 转换为字符串输出
	return frame.String()
}

// updateFocusListFromRuntime 从 LayoutResult 更新可聚焦组件列表
// 这是几何优先的焦点管理
func (m *Model) updateFocusListFromRuntime(result runtime.LayoutResult) {
	// 根据 LayoutBox 的位置排序可聚焦组件
	// 排序顺序：Y（行）优先，然后 X（列）
	type focusableItem struct {
		id   string
		x, y int
	}

	var focusables []focusableItem

	for _, box := range result.Boxes {
		if box.Node != nil && box.Node.Component != nil {
			// 检查是否是可聚焦组件
			if m.isComponentFocusable(box.Node.ID) {
				focusables = append(focusables, focusableItem{
					id: box.Node.ID,
					x:  box.X,
					y:  box.Y,
				})
			}
		}
	}

	// 按位置排序
	for i := 0; i < len(focusables); i++ {
		for j := i + 1; j < len(focusables); j++ {
			if focusables[i].y > focusables[j].y ||
				(focusables[i].y == focusables[j].y && focusables[i].x > focusables[j].x) {
				focusables[i], focusables[j] = focusables[j], focusables[i]
			}
		}
	}

	// 更新 Model 的焦点列表
	m.runtimeFocusList = make([]string, len(focusables))
	for i, item := range focusables {
		m.runtimeFocusList[i] = item.id
	}
}

// updateComponentConfigsForRender 更新组件配置以绑定数据
// 这解决了 Runtime 模式下 list/table 组件数据绑定问题
func (m *Model) updateComponentConfigsForRender(result runtime.LayoutResult) {
	for _, box := range result.Boxes {
		if box.Node == nil || box.Node.Component == nil || box.Node.Component.Instance == nil {
			continue
		}

		compID := box.Node.ID
		compInstance := box.Node.Component.Instance

		// 找到原始组件配置
		compConfig := m.findComponentConfig(compID)
		if compConfig == nil {
			continue
		}

		// 解析属性（处理 bind 属性）
		props := m.resolveProps(compConfig)

		// 创建渲染配置并更新组件
		config := core.RenderConfig{
			Data:   props,
			Width:  box.W,
			Height: box.H,
		}

		// 更新组件配置（这会解析 __bind_data）
		compInstance.UpdateRenderConfig(config)
	}
}

// isComponentFocusable 检查组件是否可聚焦
func (m *Model) isComponentFocusable(compID string) bool {
	// 检查组件实例
	if comp, exists := m.Components[compID]; exists {
		// 检查组件类型
		registry := GetGlobalRegistry()
		return registry.IsFocusable(ComponentType(comp.Type))
	}
	return false
}

// ========== Runtime 事件处理 ==========

// handleKeyPressWithRuntime 处理键盘事件（Runtime 模式）
func (m *Model) handleKeyPressWithRuntime(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 1. 检查全局快捷键
	if m.Config.Bindings != nil {
		// 检查是否有匹配的绑定
		for key, action := range m.Config.Bindings {
			if m.keyMatches(msg, key) {
				return m, m.executeAction(&action)
			}
		}
	}

	// 2. 处理焦点切换键
	switch msg.Type {
	case tea.KeyTab:
		return m, m.runtimeFocusNext()
	case tea.KeyShiftTab:
		return m, m.runtimeFocusPrev()
	case tea.KeyEscape:
		// ESC 清除焦点
		return m, m.clearFocus()
	}

	// 3. 转发到焦点组件
	if m.CurrentFocus != "" {
		updatedModel, cmd, _ := m.dispatchMessageToComponent(m.CurrentFocus, msg)
		return updatedModel, cmd
	}

	return m, nil
}

// keyMatches 检查按键是否匹配绑定键
func (m *Model) keyMatches(msg tea.KeyMsg, key string) bool {
	// 简单的按键匹配逻辑
	// 可以扩展支持更多按键格式
	keyMap := map[string]tea.KeyType{
		"ctrl+c":  tea.KeyCtrlC,
		"ctrl+z":  tea.KeyCtrlZ,
		"enter":   tea.KeyEnter,
		"tab":     tea.KeyTab,
		"esc":     tea.KeyEscape,
		"space":   tea.KeySpace,
		"up":      tea.KeyUp,
		"down":    tea.KeyDown,
		"left":    tea.KeyLeft,
		"right":   tea.KeyRight,
	}

	if kt, ok := keyMap[key]; ok {
		return msg.Type == kt
	}

	// 直接匹配字符
	if len(key) == 1 && msg.Type == tea.KeyRunes {
		return []rune(key)[0] == msg.Runes[0]
	}

	return false
}

// runtimeFocusNext 移动焦点到下一个可聚焦组件
// 基于几何位置（从左到右，从上到下）
func (m *Model) runtimeFocusNext() tea.Cmd {
	if len(m.runtimeFocusList) == 0 {
		return nil
	}

	// 找到当前焦点位置
	currentIndex := -1
	for i, id := range m.runtimeFocusList {
		if id == m.CurrentFocus {
			currentIndex = i
			break
		}
	}

	// 移动到下一个
	var nextFocus string
	tabCycles := m.Config.TabCycles
	if !tabCycles {
		tabCycles = true
	}

	if currentIndex >= 0 && currentIndex < len(m.runtimeFocusList)-1 {
		nextFocus = m.runtimeFocusList[currentIndex+1]
	} else if currentIndex == len(m.runtimeFocusList)-1 {
		if tabCycles {
			nextFocus = m.runtimeFocusList[0]
		} else {
			return nil
		}
	} else {
		nextFocus = m.runtimeFocusList[0]
	}

	return m.setFocus(nextFocus)
}

// runtimeFocusPrev 移动焦点到上一个可聚焦组件
func (m *Model) runtimeFocusPrev() tea.Cmd {
	if len(m.runtimeFocusList) == 0 {
		return nil
	}

	// 找到当前焦点位置
	currentIndex := -1
	for i, id := range m.runtimeFocusList {
		if id == m.CurrentFocus {
			currentIndex = i
			break
		}
	}

	// 移动到上一个
	tabCycles := m.Config.TabCycles
	if !tabCycles {
		tabCycles = true
	}

	var prevFocus string
	if currentIndex > 0 {
		prevFocus = m.runtimeFocusList[currentIndex-1]
	} else if currentIndex == 0 {
		if tabCycles {
			prevFocus = m.runtimeFocusList[len(m.runtimeFocusList)-1]
		} else {
			return nil
		}
	} else {
		prevFocus = m.runtimeFocusList[len(m.runtimeFocusList)-1]
	}

	return m.setFocus(prevFocus)
}

// ========== Runtime 窗口尺寸更新 ==========

// updateRuntimeWindowSize 更新 Runtime 窗口尺寸
// 这个方法在处理 tea.WindowSizeMsg 时被调用
func (m *Model) updateRuntimeWindowSize(width, height int) {
	if m.RuntimeEngine != nil {
		m.RuntimeEngine.UpdateDimensions(width, height)

		// 重新计算布局
		if m.RuntimeRoot != nil {
			constraints := runtime.BoxConstraints{
				MinWidth:  0,
				MaxWidth:  width,
				MinHeight: 0,
				MaxHeight: height,
			}
			m.RuntimeEngine.Layout(m.RuntimeRoot, constraints)
		}
	}
}

// ========== 状态同步 ==========

// syncStateToRuntime 将 Model 的状态同步到 Runtime
// 当 State 发生变化时调用此方法
//
// 此方法执行以下操作：
// 1. 清除 Props 缓存
// 2. 重新解析所有组件的 Props（基于新的 State）
// 3. 更新所有组件实例的 RenderConfig
// 4. 标记 Runtime 需要重新渲染
func (m *Model) syncStateToRuntime() {
	if !m.UseRuntime {
		return
	}

	// For native Runtime components, traverse the RuntimeRoot tree
	// and update components with fresh template bindings
	if m.RuntimeRoot != nil {
		m.syncRuntimeNode(m.RuntimeRoot)
		return
	}

	// Legacy fallback: Use ComponentInstanceRegistry
	// 1. 清除 Props 缓存
	if m.propsCache != nil {
		m.propsCache.Clear()
	}

	// 2. 获取所有组件实例
	allInstances := m.ComponentInstanceRegistry.GetAll()

	// 3. 更新每个组件实例的配置
	for compID, compInstance := range allInstances {
		// 查找组件配置
		compConfig := m.findComponentConfig(compID)
		if compConfig == nil {
			log.Trace("syncStateToRuntime: Component config not found for %s, skipping", compID)
			continue
		}

		// 使用当前 State 重新解析 Props
		freshProps := m.resolveProps(compConfig)

		// 创建新的 RenderConfig
		newConfig := core.RenderConfig{
			Data:   freshProps,
			Width:  m.Width,
			Height: m.Height,
		}

		// 更新组件实例的配置
		// 使用已有的 updateComponentInstanceConfig 函数，它会：
		// - 检查配置是否真的改变了
		// - 调用组件的 UpdateRenderConfig 方法
		// - 更新 LastConfig
		updated := updateComponentInstanceConfig(compInstance, newConfig, compID)
		if updated {
			log.Trace("syncStateToRuntime: Updated config for component %s", compID)
		}
	}
}

// syncRuntimeNode recursively syncs state to Runtime nodes and their components
func (m *Model) syncRuntimeNode(node *runtime.LayoutNode) {
	if node == nil {
		return
	}

	// Update this node's component if it's a native component
	if node.Component != nil && node.Component.Instance != nil {
		// Try to update native Runtime components with fresh template bindings
		if wrapper, ok := node.Component.Instance.(*NativeComponentWrapper); ok {
			m.updateNativeComponent(wrapper, node.ID)
		}
	}

	// Recursively update children
	for _, child := range node.Children {
		m.syncRuntimeNode(child)
	}
}

// updateNativeComponent updates a native Runtime component with fresh template bindings
func (m *Model) updateNativeComponent(wrapper *NativeComponentWrapper, compID string) {
	// Find the component config from the original DSL
	compConfig := m.findComponentConfig(compID)
	if compConfig == nil {
		// Try to find from Config layout
		compConfig = m.findComponentConfigInLayout(m.Config.Layout.Children, compID)
		if compConfig == nil {
			return
		}
	}

	// Re-bind props to current state
	freshProps := m.bindPropsToData(compConfig.Props)

	// Update the native component based on its type
	switch comp := wrapper.Component.(type) {
	case *components.TextComponent:
		if content, ok := freshProps["content"].(string); ok {
			comp.WithContent(content)
		} else if text, ok := freshProps["text"].(string); ok {
			comp.WithContent(text)
		}
	case *components.HeaderComponent:
		if title, ok := freshProps["title"].(string); ok {
			comp.WithContent(title)
		} else if content, ok := freshProps["content"].(string); ok {
			comp.WithContent(content)
		}
	case *components.FooterComponent:
		if text, ok := freshProps["text"].(string); ok {
			comp.WithContent(text)
		} else if content, ok := freshProps["content"].(string); ok {
			comp.WithContent(content)
		}
	case *components.InputComponent:
		if value, ok := freshProps["value"].(string); ok {
			comp.WithValue(value)
		}
		if placeholder, ok := freshProps["placeholder"].(string); ok {
			comp.WithPlaceholder(placeholder)
		}
	case *components.ListComponent:
		// Update list items if they changed
		if itemsData, ok := freshProps["items"]; ok {
			factory := dsl.NewFactory()
			items := factory.ParseListItems(itemsData)
			comp.WithItems(items)
		}
		// Update title
		if title, ok := freshProps["title"].(string); ok {
			comp.WithTitle(title)
		}
	case *components.TableComponent:
		// Update table columns if they changed
		if columnsData, ok := freshProps["columns"]; ok {
			factory := dsl.NewFactory()
			columns := factory.ParseTableColumns(columnsData)
			comp.WithColumns(columns)
		}
		// Update table rows/data if they changed
		if rowsData, ok := freshProps["rows"]; ok {
			factory := dsl.NewFactory()
			rows := factory.ParseTableRows(rowsData)
			comp.WithRows(rows)
		}
		if dataData, ok := freshProps["data"]; ok {
			if dataArray, ok := dataData.([][]interface{}); ok {
				comp.WithData(dataArray)
			} else if dataArray, ok := dataData.([]interface{}); ok {
				converted := make([][]interface{}, 0, len(dataArray))
				for _, rowIntf := range dataArray {
					if rowSlice, ok := rowIntf.([]interface{}); ok {
						converted = append(converted, rowSlice)
					}
				}
				comp.WithData(converted)
			}
		}
		// Update dimensions
		if width, ok := freshProps["width"]; ok {
			if w, ok := width.(int); ok {
				comp.WithWidth(w)
			} else if w, ok := width.(float64); ok {
				comp.WithWidth(int(w))
			}
		}
		if height, ok := freshProps["height"]; ok {
			if h, ok := height.(int); ok {
				comp.WithHeight(h)
			} else if h, ok := height.(float64); ok {
				comp.WithHeight(int(h))
			}
		}

	case *components.FormComponent:
	// Update form fields if they changed
	if fieldsData, ok := freshProps["fields"]; ok {
		factory := dsl.NewFactory()
		fields := factory.ParseFormFields(fieldsData)
		comp.WithFields(fields)
	}

	// Update title and description
	if title, ok := freshProps["title"]; ok {
		if titleStr, ok := title.(string); ok {
			comp.WithTitle(titleStr)
		}
	}
	if description, ok := freshProps["description"]; ok {
		if descStr, ok := description.(string); ok {
			comp.WithDescription(descStr)
		}
	}

	// Update button labels
	if submitLabel, ok := freshProps["submitLabel"]; ok {
		if labelStr, ok := submitLabel.(string); ok {
			comp.WithSubmitLabel(labelStr)
		}
	}
	if cancelLabel, ok := freshProps["cancelLabel"]; ok {
		if labelStr, ok := cancelLabel.(string); ok {
			comp.WithCancelLabel(labelStr)
		}
	}

	// Update initial values
	if valuesData, ok := freshProps["values"]; ok {
		if valuesMap, ok := valuesData.(map[string]interface{}); ok {
			for key, value := range valuesMap {
				if valueStr, ok := value.(string); ok {
					comp.SetFieldValue(key, valueStr)
				}
			}
		}
	}

	// Update dimensions
	if width, ok := freshProps["width"]; ok {
		if w, ok := width.(int); ok {
			comp.WithWidth(w)
		} else if w, ok := width.(float64); ok {
			comp.WithWidth(int(w))
		}
	}
	if height, ok := freshProps["height"]; ok {
		if h, ok := height.(int); ok {
			comp.WithHeight(h)
		} else if h, ok := height.(float64); ok {
			comp.WithHeight(int(h))
		}
	}

case *components.TextareaComponent:
	// Update textarea value if it changed
	if value, ok := freshProps["value"].(string); ok {
		comp.SetValue(value)
	}

	// Update placeholder
	if placeholder, ok := freshProps["placeholder"].(string); ok {
		comp.WithPlaceholder(placeholder)
	}

	// Update prompt
	if prompt, ok := freshProps["prompt"].(string); ok {
		comp.WithPrompt(prompt)
	}

	// Update charLimit
	if charLimit, ok := freshProps["charLimit"].(int); ok {
		comp.WithCharLimit(charLimit)
	} else if charLimit, ok := freshProps["charLimit"].(float64); ok {
		comp.WithCharLimit(int(charLimit))
	}

	// Update showLineNumbers
	if showLineNumbers, ok := freshProps["showLineNumbers"].(bool); ok {
		comp.WithShowLineNumbers(showLineNumbers)
	}

	// Update enterSubmits
	if enterSubmits, ok := freshProps["enterSubmits"].(bool); ok {
		comp.WithEnterSubmits(enterSubmits)
	}

	// Update dimensions
	if width, ok := freshProps["width"].(int); ok {
		comp.WithWidth(width)
	} else if width, ok := freshProps["width"].(float64); ok {
		comp.WithWidth(int(width))
	}
	if height, ok := freshProps["height"].(int); ok {
		comp.WithHeight(height)
	} else if height, ok := freshProps["height"].(float64); ok {
		comp.WithHeight(int(height))
	}
	if maxHeight, ok := freshProps["maxHeight"].(int); ok {
		comp.WithMaxHeight(maxHeight)
	} else if maxHeight, ok := freshProps["maxHeight"].(float64); ok {
		comp.WithMaxHeight(int(maxHeight))
	}

case *components.ProgressComponent:
	// Update percent
	if percent, ok := freshProps["percent"].(float64); ok {
		comp.SetPercent(percent)
	} else if percent, ok := freshProps["percent"].(int); ok {
		comp.SetPercent(float64(percent))
	}

	// Update label
	if label, ok := freshProps["label"].(string); ok {
		comp.WithLabel(label)
	}

	// Update showPercentage
	if showPercentage, ok := freshProps["showPercentage"].(bool); ok {
		comp.WithShowPercentage(showPercentage)
	}

	// Update filledChar
	if filledChar, ok := freshProps["filledChar"].(string); ok {
		comp.WithFilledChar(filledChar)
	}

	// Update emptyChar
	if emptyChar, ok := freshProps["emptyChar"].(string); ok {
		comp.WithEmptyChar(emptyChar)
	}

	// Update fullColor
	if fullColor, ok := freshProps["fullColor"].(string); ok {
		comp.WithFullColor(fullColor)
	} else if color, ok := freshProps["color"].(string); ok {
		comp.WithFullColor(color)
	}

	// Update emptyColor
	if emptyColor, ok := freshProps["emptyColor"].(string); ok {
		comp.WithEmptyColor(emptyColor)
	}

	// Update dimensions
	if width, ok := freshProps["width"].(int); ok {
		comp.WithWidth(width)
	} else if width, ok := freshProps["width"].(float64); ok {
		comp.WithWidth(int(width))
	}
	if height, ok := freshProps["height"].(int); ok {
		comp.WithHeight(height)
	} else if height, ok := freshProps["height"].(float64); ok {
		comp.WithHeight(int(height))
	}

case *components.SpinnerComponent:
	// Update style
	if style, ok := freshProps["style"].(string); ok {
		comp.WithStyle(style)
	}

	// Update label
	if label, ok := freshProps["label"].(string); ok {
		comp.WithLabel(label)
	}

	// Update labelPosition
	if labelPosition, ok := freshProps["labelPosition"].(string); ok {
		comp.WithLabelPosition(labelPosition)
	}

	// Update running
	if running, ok := freshProps["running"].(bool); ok {
		comp.WithRunning(running)
	}

	// Update color
	if color, ok := freshProps["color"].(string); ok {
		comp.WithColor(color)
	}

	// Update background
	if background, ok := freshProps["background"].(string); ok {
		comp.WithBackground(background)
	}

	// Update speed
	if speed, ok := freshProps["speed"].(int); ok {
		comp.WithSpeed(speed)
	} else if speed, ok := freshProps["speed"].(float64); ok {
		comp.WithSpeed(int(speed))
	}

	// Update dimensions
	if width, ok := freshProps["width"].(int); ok {
		comp.WithWidth(width)
	} else if width, ok := freshProps["width"].(float64); ok {
		comp.WithWidth(int(width))
	}
	if height, ok := freshProps["height"].(int); ok {
		comp.WithHeight(height)
	} else if height, ok := freshProps["height"].(float64); ok {
		comp.WithHeight(int(height))
	}
}
}

// findComponentConfigInLayout searches for a component config in the layout tree
func (m *Model) findComponentConfigInLayout(children []Component, compID string) *Component {
	for i := range children {
		if children[i].ID == compID {
			return &children[i]
		}
		if children[i].Children != nil {
			if found := m.findComponentConfigInLayout(children[i].Children, compID); found != nil {
				return found
			}
		}
	}
	return nil
}

// ========== 运行时切换 ==========

// SwitchToRuntime 切换到 Runtime 引擎
// 可以在运行时动态切换（主要用于调试）
func (m *Model) SwitchToRuntime() {
	if !m.UseRuntime {
		m.UseRuntime = true
		m.initializeRuntime()
		log.Info("Switched to Runtime engine")
	}
}

// SwitchToLegacy 切换到 Legacy 引擎
// 可以在运行时动态切换（主要用于调试）
func (m *Model) SwitchToLegacy() {
	if m.UseRuntime {
		m.UseRuntime = false
		m.initializeLegacy()
		log.Info("Switched to Legacy engine")
	}
}

// initializeLegacy 初始化 Legacy 布局引擎
// 保留用于兼容性
func (m *Model) initializeLegacy() {
	log.Trace("Model.initializeLegacy: initializing legacy layout engine")
	// Legacy 引擎初始化逻辑...
}

// ========== 调试方法 ==========

// GetRuntimeDebugInfo 获取 Runtime 调试信息
func (m *Model) GetRuntimeDebugInfo() map[string]interface{} {
	info := make(map[string]interface{})

	info["UseRuntime"] = m.UseRuntime
	info["RuntimeEngine"] = m.RuntimeEngine != nil
	info["RuntimeRoot"] = m.RuntimeRoot != nil

	if m.RuntimeEngine != nil {
		info["RuntimeWidth"] = m.RuntimeEngine.GetWidth()
		info["RuntimeHeight"] = m.RuntimeEngine.GetHeight()
	}

	if m.RuntimeRoot != nil {
		info["RootNodeID"] = m.RuntimeRoot.ID
		info["RootNodeType"] = m.RuntimeRoot.Type
	}

	info["FocusList"] = m.runtimeFocusList
	info["CurrentFocus"] = m.CurrentFocus

	return info
}

// ========== Runtime 接口实现 ==========

// GetLayoutResult 获取当前布局结果
func (m *Model) GetLayoutResult() runtime.LayoutResult {
	if m.RuntimeEngine != nil && m.RuntimeRoot != nil {
		constraints := runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  m.Width,
			MinHeight: 0,
			MaxHeight: m.Height,
		}
		return m.RuntimeEngine.Layout(m.RuntimeRoot, constraints)
	}
	return runtime.LayoutResult{}
}

// GetFrame 获取当前渲染帧
func (m *Model) GetFrame() runtime.Frame {
	if m.RuntimeEngine != nil {
		result := m.GetLayoutResult()
		return m.RuntimeEngine.Render(result)
	}
	return runtime.Frame{}
}

// ResolvePropsForRuntime 为 Runtime 组件包装器提供 Props 解析
// 这是 PropsResolver 接口的实现
func (m *Model) ResolvePropsForRuntime(compID string) map[string]interface{} {
	compConfig := m.findComponentConfig(compID)
	if compConfig == nil {
		return make(map[string]interface{})
	}

	return m.resolveProps(compConfig)
}

// mapJustifyString maps justify string to runtime.Justify
func mapJustifyString(justify string) runtime.Justify {
	switch justify {
	case "start", "left", "top":
		return runtime.JustifyStart
	case "center", "middle":
		return runtime.JustifyCenter
	case "end", "right", "bottom":
		return runtime.JustifyEnd
	case "space-between":
		return runtime.JustifySpaceBetween
	case "space-around":
		return runtime.JustifySpaceAround
	case "space-evenly":
		return runtime.JustifySpaceEvenly
	default:
		return runtime.JustifyStart
	}
}

// mapAlignString maps align string to runtime.Align
func mapAlignString(align string) runtime.Align {
	switch align {
	case "start", "left", "top":
		return runtime.AlignStart
	case "center", "middle":
		return runtime.AlignCenter
	case "end", "right", "bottom":
		return runtime.AlignEnd
	case "stretch":
		return runtime.AlignStretch
	default:
		return runtime.AlignStart
	}
}
