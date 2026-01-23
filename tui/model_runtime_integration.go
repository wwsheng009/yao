package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/core"
	"github.com/yaoapp/yao/tui/runtime"
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

	// 创建适配器
	adapter := NewRuntimeAdapter(m)

	// 转换 DSL 树为 Runtime 树
	// Config.Layout.Children 包含顶层组件
	if m.Config != nil && len(m.Config.Layout.Children) > 0 {
		// 创建虚拟根节点包含所有顶层组件
		rootComponent := Component{
			ID:   "root",
			Type: "column",
		}

		// 复制所有子组件
		rootComponent.Children = m.Config.Layout.Children
		rootComponent.Direction = m.Config.Layout.Direction

		m.RuntimeRoot = adapter.ToRuntimeLayoutNode(&rootComponent)
		log.Trace("Model.initializeRuntime: runtime root node created with ID=%s, type=%s",
			m.RuntimeRoot.ID, m.RuntimeRoot.Type)
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

	// 4. 标记 Runtime 需要重新渲染
	if m.RuntimeEngine != nil {
		m.RuntimeEngine.MarkFullRender()
	}
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
