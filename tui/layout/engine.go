package layout

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/reflow/wrap"
	"github.com/yaoapp/yao/tui/core"
)

type Engine struct {
	config *LayoutConfig
	root   *LayoutNode
	window WindowSize
	theme  map[string]interface{}
}

func NewEngine(config *LayoutConfig) *Engine {
	if config.WindowSize == nil {
		config.WindowSize = &WindowSize{Width: 80, Height: 24}
	}
	// Ensure non-zero dimensions to avoid layout issues during initialization
	if config.WindowSize.Width == 0 {
		config.WindowSize.Width = 80
	}
	if config.WindowSize.Height == 0 {
		config.WindowSize.Height = 24
	}
	return &Engine{
		config: config,
		root:   config.Root,
		window: *config.WindowSize,
		theme:  config.Theme,
	}
}

func (e *Engine) UpdateWindowSize(width, height int) {
	e.window.Width = width
	e.window.Height = height
	e.MarkDirty(e.root)
}

func (e *Engine) MarkDirty(node *LayoutNode) {
	if node == nil {
		return
	}
	if node.ID == "root" {
		node.Dirty = true
		return
	}
}

func (e *Engine) Layout() *LayoutResult {
	if e.root == nil {
		return &LayoutResult{}
	}

	result := &LayoutResult{}

	// ✅ 阶段1：约束传递
	e.passConstraints(e.root, e.window.Width, e.window.Height)

	// 阶段2：子节点响应并计算实际 Bound
	e.layoutNode(e.root, 0, 0, e.window.Width, e.window.Height, result)

	// 阶段3：通知组件其实际分配的大小
	e.notifyComponentSizes(result.Nodes)

	return result
}

// passConstraints 传递约束给节点树
func (e *Engine) passConstraints(node *LayoutNode, maxWidth, maxHeight int) {
	if node == nil {
		return
	}

	// 计算内部可用空间（减去 padding）
	innerWidth := maxWidth
	innerHeight := maxHeight

	if node.Style.Padding != nil {
		innerWidth = max(0, innerWidth-node.Style.Padding.Left-node.Style.Padding.Right)
		innerHeight = max(0, innerHeight-node.Style.Padding.Top-node.Style.Padding.Bottom)
	}

	// 设置节点的可用约束
	node.AvailableWidth = innerWidth
	node.AvailableHeight = innerHeight

	// 对于没有子节点的叶子节点，调用 Measure
	if len(node.Children) == 0 && node.Component != nil && node.Component.Instance != nil {
		if measurable, ok := node.Component.Instance.(core.Measurable); ok {
			node.PreferredWidth, node.PreferredHeight = measurable.Measure(innerWidth, innerHeight)
		}
	}

	// 递归传递约束给子节点
	for _, child := range node.Children {
		e.passConstraints(child, innerWidth, innerHeight)
	}
}

// notifyComponentSizes 通知所有组件其实际分配的尺寸
func (e *Engine) notifyComponentSizes(nodes []*LayoutNode) {
	for _, node := range nodes {
		// 只通知有组件实例的节点
		if node.Component == nil || node.Component.Instance == nil {
			continue
		}

		// 尝试调用 SetSize 方法
		// 方案 A：如果 ComponentInterface 有 SetSize (通过类型断言)
		if component, ok := node.Component.Instance.(interface{ SetSize(w, h int) }); ok {
			component.SetSize(node.Bound.Width, node.Bound.Height)
			continue
		}

		// 兜底：尝试调用 SetWidth/SetHeight（向后兼容）
		if setter, ok := node.Component.Instance.(interface{ SetWidth(w int) }); ok {
			setter.SetWidth(node.Bound.Width)
		}
		if setter, ok := node.Component.Instance.(interface{ SetHeight(h int) }); ok {
			setter.SetHeight(node.Bound.Height)
		}
	}
}

func (e *Engine) layoutNode(
	node *LayoutNode,
	x, y, width, height int,
	result *LayoutResult,
) {
	if node == nil {
		return
	}

	e.ensureStyle(node)
	e.calculateMetrics(node, width, height)
	node.Bound = Rect{X: x, Y: y, Width: width, Height: height}

	result.Nodes = append(result.Nodes, node)
	if node.Dirty {
		result.Dirties = append(result.Dirties, node)
	}

	if len(node.Children) == 0 {
		return
	}

	innerX := x
	innerY := y
	innerWidth := width
	innerHeight := height

	if node.Style.Padding != nil {
		innerX += node.Style.Padding.Left
		innerY += node.Style.Padding.Top
		innerWidth = max(0, innerWidth-node.Style.Padding.Left-node.Style.Padding.Right)
		innerHeight = max(0, innerHeight-node.Style.Padding.Top-node.Style.Padding.Bottom)
	}

	switch node.Type {
	case LayoutFlex:
		e.layoutFlex(node, innerX, innerY, innerWidth, innerHeight, result)
	case LayoutGrid:
		e.layoutGrid(node, innerX, innerY, innerWidth, innerHeight, result)
	case LayoutAbsolute:
		e.layoutAbsolute(node, x, y, width, height, result)
	}
}

func (e *Engine) layoutFlex(
	node *LayoutNode,
	x, y, width, height int,
	result *LayoutResult,
) {
	if len(node.Children) == 0 {
		return
	}

	config := &FlexConfig{
		Direction:  node.Style.Direction,
		AlignItems: node.Style.AlignItems,
		Justify:    node.Style.Justify,
		Wrap:       node.Style.Wrap,
		Gap:        node.Style.Gap,
	}

	// 收集所有子元素信息，保持原始顺序
	var allChildren []*flexChildInfo
	var totalFixedSize int
	var growSum float64
	var shrinkSum float64  // 新增: 计算 shrink 总和

	for _, child := range node.Children {
		info := e.measureChild(child, config, width, height)
		allChildren = append(allChildren, info)

		if info.Grow.Value > 0 {
			growSum += info.Grow.Value
		} else if info.Shrink.Value > 0 {
			shrinkSum += info.Shrink.Value
		} else {
			totalFixedSize += info.Size
		}
	}

	totalGap := node.Style.Gap * (len(node.Children) - 1)

	// 根据布局方向选择正确的可用空间维度
	var containerSize int
	if config.Direction == DirectionRow {
		containerSize = width
	} else {
		containerSize = height
	}
	availableSpace := containerSize - totalFixedSize - totalGap

	// ✅ 新增：处理空间不足的情况（Shrink）
	if availableSpace < 0 && shrinkSum > 0 {
		// 按照收缩比例减少子元素大小
		for _, info := range allChildren {
			if info.Shrink.Value > 0 {
				shrinkAmount := int(float64(-availableSpace) * (info.Shrink.Value / shrinkSum))
				info.Size = max(0, info.Size - shrinkAmount)
			}
		}
	} else if availableSpace > 0 && growSum > 0 {
		// 处理空间充足的情况（Grow）
		for _, info := range allChildren {
			if info.Grow.Value > 0 {
				extra := int(float64(availableSpace) * (info.Grow.Value / growSum))
				info.Size = extra
			}
		}
	}

	e.distributeFlexChildrenOrdered(
		allChildren, config, x, y, width, height, result,
	)
}

type flexChildInfo struct {
	Node   *LayoutNode
	Size   int
	Grow   Grow
	Shrink Grow
	Basis  int
}

func (e *Engine) measureChild(child *LayoutNode, config *FlexConfig, parentWidth, parentHeight int) *flexChildInfo {
	info := &flexChildInfo{
		Node: child,
		Grow: Grow{Value: 0},
	}

	if child.Style == nil {
		e.ensureStyle(child)
	}

	var size *Size
	if config.Direction == DirectionRow {
		size = child.Style.Width
	} else {
		size = child.Style.Height
	}

	// 检查 size 是否有有效值
	isStyleSet := false
	if size != nil && size.Value != nil {
		switch v := size.Value.(type) {
		case float64:
			info.Size = int(v)
			isStyleSet = true
		case int:
			info.Size = v
			isStyleSet = true
		case string:
			if v == "flex" {
				info.Grow.Value = 1
				info.Size = 0 // 初始为0，稍后分配
				isStyleSet = true
			}
		}
	}

	// 如果没有样式定义，检查是否实现 Measurable 接口
	if !isStyleSet && child.Component != nil && child.Component.Instance != nil {
		if measurable, ok := child.Component.Instance.(core.Measurable); ok {
			// 使用组件提供的测量结果
			measuredWidth, measuredHeight := measurable.Measure(parentWidth, parentHeight)

			if config.Direction == DirectionRow {
				info.Size = measuredWidth
			} else {
				info.Size = measuredHeight
			}
			isStyleSet = true
		}
	}

	// 如果仍未确定尺寸，使用默认测量逻辑
	if !isStyleSet {
		if config.Direction == DirectionRow {
			info.Size = e.measureChildWidth(child, parentHeight)
		} else {
			info.Size = e.measureChildHeight(child, parentWidth)
		}
	}

	// Apply minimum size based on direction
	if config.Direction == DirectionRow && info.Size < child.Style.MinWidth {
		info.Size = child.Style.MinWidth
	} else if config.Direction == DirectionColumn && info.Size < child.Style.MinHeight {
		info.Size = child.Style.MinHeight
	}

	return info
}

func (e *Engine) distributeFlexChildren(
	fixedChildren, flexibleChildren []*flexChildInfo,
	config *FlexConfig, x, y, width, height int,
	totalGap int,
	result *LayoutResult,
) {
	allChildren := append(append([]*flexChildInfo{}, fixedChildren...), flexibleChildren...)

	offset := 0

	switch config.Direction {
	case DirectionRow:
		switch config.Justify {
		case JustifyCenter:
			totalWidth := totalGap
			for _, child := range allChildren {
				totalWidth += child.Size
			}
			offset = (width - totalWidth) / 2
		case JustifyEnd:
			totalWidth := totalGap
			for _, child := range allChildren {
				totalWidth += child.Size
			}
			offset = width - totalWidth
		case JustifySpaceBetween:
			if len(allChildren) > 1 {
				availableSpace := width - totalGap
				for _, child := range allChildren {
					availableSpace -= child.Size
				}
				totalGap = int(float64(availableSpace) / float64(len(allChildren)-1))
			}
		case JustifySpaceAround:
			totalWidth := totalGap
			for _, child := range allChildren {
				totalWidth += child.Size
			}
			space := (width - totalWidth) / (2 * len(allChildren))
			offset += space
			totalGap = space * 2
		case JustifyStart:
			offset = 0
		}

		for _, childInfo := range allChildren {
			childX := x + offset
			childWidth := childInfo.Size
			childHeight := height

			switch config.AlignItems {
			case AlignCenter:
				measureHeight := e.measureChildHeight(childInfo.Node, childWidth)
				if measureHeight < height {
					childY := y + (height-measureHeight)/2
					childHeight = measureHeight
					e.layoutNode(childInfo.Node, childX, childY, childWidth, childHeight, result)
				} else {
					e.layoutNode(childInfo.Node, childX, y, childWidth, childHeight, result)
				}
			case AlignEnd:
				measureHeight := e.measureChildHeight(childInfo.Node, childWidth)
				if measureHeight < height {
					childY := y + (height - measureHeight)
					childHeight = measureHeight
					e.layoutNode(childInfo.Node, childX, childY, childWidth, childHeight, result)
				} else {
					e.layoutNode(childInfo.Node, childX, y, childWidth, childHeight, result)
				}
			case AlignStretch:
				e.layoutNode(childInfo.Node, childX, y, childWidth, childHeight, result)
			default:
				e.layoutNode(childInfo.Node, childX, y, childWidth, childHeight, result)
			}

			offset += childWidth + nodeStyleGap(nodeFromInfo(childInfo))
		}

	case DirectionColumn:
		switch config.Justify {
		case JustifyCenter:
			totalHeight := totalGap
			for _, child := range allChildren {
				totalHeight += child.Size
			}
			offset = (height - totalHeight) / 2
		case JustifyEnd:
			totalHeight := totalGap
			for _, child := range allChildren {
				totalHeight += child.Size
			}
			offset = height - totalHeight
		case JustifyStart:
			offset = 0
		}

		for _, childInfo := range allChildren {
			childY := y + offset
			childHeight := childInfo.Size
			childWidth := width

			switch config.AlignItems {
			case AlignCenter:
				measureWidth := e.measureChildWidth(childInfo.Node, childHeight)
				if measureWidth < width {
					childX := x + (width-measureWidth)/2
					childWidth = measureWidth
					e.layoutNode(childInfo.Node, childX, childY, childWidth, childHeight, result)
				} else {
					e.layoutNode(childInfo.Node, x, childY, childWidth, childHeight, result)
				}
			case AlignEnd:
				measureWidth := e.measureChildWidth(childInfo.Node, childHeight)
				if measureWidth < width {
					childX := x + (width - measureWidth)
					childWidth = measureWidth
					e.layoutNode(childInfo.Node, childX, childY, childWidth, childHeight, result)
				} else {
					e.layoutNode(childInfo.Node, x, childY, childWidth, childHeight, result)
				}
			case AlignStretch:
				e.layoutNode(childInfo.Node, x, childY, childWidth, childHeight, result)
			default:
				e.layoutNode(childInfo.Node, x, childY, childWidth, childHeight, result)
			}

			offset += childHeight + nodeStyleGap(nodeFromInfo(childInfo))
		}
	}
}

// distributeFlexChildrenOrdered 按原始顺序布局子元素
func (e *Engine) distributeFlexChildrenOrdered(
	children []*flexChildInfo,
	config *FlexConfig, x, y, width, height int,
	result *LayoutResult,
) {
	if len(children) == 0 {
		return
	}

	offset := 0
	gap := config.Gap

	switch config.Direction {
	case DirectionRow:
		for i, childInfo := range children {
			childX := x + offset
			childWidth := childInfo.Size
			childHeight := height

			e.layoutNode(childInfo.Node, childX, y, childWidth, childHeight, result)

			offset += childWidth
			if i < len(children)-1 {
				offset += gap
			}
		}

	case DirectionColumn:
		for i, childInfo := range children {
			childY := y + offset
			childHeight := childInfo.Size
			childWidth := width

			e.layoutNode(childInfo.Node, x, childY, childWidth, childHeight, result)

			offset += childHeight
			if i < len(children)-1 {
				offset += gap
			}
		}
	}
}

func (e *Engine) layoutGrid(
	node *LayoutNode,
	x, y, width, height int,
	result *LayoutResult,
) {
	if len(node.Children) == 0 {
		return
	}

	columns := 2
	rows := math.Ceil(float64(len(node.Children)) / float64(columns))

	colWidth := width / columns
	rowHeight := int(height / int(rows))

	colGap := node.Style.Gap
	rowGap := node.Style.Gap

	for idx, child := range node.Children {
		col := idx % columns
		row := int(idx / columns)

		childX := x + col*colWidth + col*colGap
		childY := y + row*rowHeight + row*rowGap

		lastRow := int(rows) - 1
		isLastRow := row == lastRow

		childWidth := colWidth
		childHeight := rowHeight

		if col == columns-1 {
			childWidth = width - childX
		}
		if isLastRow {
			childHeight = max(0, height-childY)
		}

		e.layoutNode(child, childX, childY, childWidth, childHeight, result)
	}
}

func (e *Engine) layoutAbsolute(
	node *LayoutNode,
	x, y, width, height int,
	result *LayoutResult,
) {
	for _, child := range node.Children {
		childX := x
		childY := y
		childWidth := width
		childHeight := height

		style := child.Style
		if style != nil {
			if style.Position == PositionAbsolute {
				if style.Left > 0 {
					childX = x + style.Left
				} else if style.Right > 0 {
					childX = x + width - style.Right
				}
				if style.Top > 0 {
					childY = y + style.Top
				} else if style.Bottom > 0 {
					childY = y + height - style.Bottom
				}
			}

			if style.Width != nil {
				switch v := style.Width.Value.(type) {
				case int:
					childWidth = v
				case float64:
					childWidth = int(v)
				}
			}
			if style.Height != nil {
				switch v := style.Height.Value.(type) {
				case int:
					childHeight = v
				case float64:
					childHeight = int(v)
				}
			}
		}

		e.layoutNode(child, childX, childY, childWidth, childHeight, result)
	}
}

func (e *Engine) ensureStyle(node *LayoutNode) bool {
	if node.Style == nil {
		node.Style = &LayoutStyle{
			Direction:  DirectionColumn,
			AlignItems: AlignStart,
			Justify:    JustifyStart,
			Wrap:       false,
			Gap:        0,
			Width:      NewSize(nil),
			Height:     NewSize(nil),
			Position:   PositionRelative,
		}
		return true
	}
	return false
}

func (e *Engine) calculateMetrics(node *LayoutNode, width, height int) {
	node.Metrics = &LayoutMetrics{
		ContentWidth:  width,
		ContentHeight: height,
		PaddingWidth:  0,
		PaddingHeight: 0,
		MarginWidth:   0,
		MarginHeight:  0,
		BorderWidth:   0,
		BorderHeight:  0,
		TotalWidth:    width,
		TotalHeight:   height,
	}

	if node.Style.Padding != nil {
		node.Metrics.PaddingWidth = node.Style.Padding.Left + node.Style.Padding.Right
		node.Metrics.PaddingHeight = node.Style.Padding.Top + node.Style.Padding.Bottom
		node.Metrics.ContentWidth = max(0, width-node.Metrics.PaddingWidth)
		node.Metrics.ContentHeight = max(0, height-node.Metrics.PaddingHeight)
	}

	node.Metrics.TotalWidth = node.Metrics.ContentWidth + node.Metrics.PaddingWidth
	node.Metrics.TotalHeight = node.Metrics.ContentHeight + node.Metrics.PaddingHeight
}

func (e *Engine) getProps(node *LayoutNode) map[string]interface{} {
	if e.config.PropsResolver != nil {
		return e.config.PropsResolver(node)
	}
	return node.Props
}

func (e *Engine) measureChildWidth(node *LayoutNode, height int) int {
	if node.Style != nil && node.Style.Width != nil {
		switch v := node.Style.Width.Value.(type) {
		case string:
			if v == "flex" {
				return 0 // Flex 的宽度将在 distributeFlexChildren 中计算
			}
		case int:
			if v > 0 {
				return v
			}
		case float64:
			if v > 0 {
				return int(v)
			}
		}
	}

	if node.Component != nil && node.Component.Instance != nil {
		if config := node.Component.LastConfig; config.Width > 0 {
			return config.Width
		}

		if config := node.Component.LastConfig; config.Height > 0 {
			height = config.Height
		}

		props := e.getProps(node)

		// ✅ 对于 text 组件，使用 runewidth 计算中文宽度
		if node.Component.Instance.GetComponentType() == "text" {
			if props != nil {
				if content, ok := props["content"].(string); ok {
					// 剥离 ANSI 转义符
					stripped := ansi.Strip(content)
					// 计算视觉宽度（中文算2个字符宽度）
					return runewidth.StringWidth(stripped)
				}
			}
		}

		renderConfig := core.RenderConfig{
			Width:  200,
			Height: height,
			Data:   props,
		}

		content, err := node.Component.Instance.Render(renderConfig)
		if err == nil {
			lines := strings.Split(content, "\n")
			maxWidth := 0
			for _, line := range lines {
				// ✅ 剥离 ANSI 转义符
				stripped := ansi.Strip(line)
				// ✅ 使用 runewidth 计算视觉宽度
				w := runewidth.StringWidth(stripped)
				if w > maxWidth {
					maxWidth = w
				}
			}
			if maxWidth > 0 && maxWidth < 200 {
				return maxWidth
			}
		}

		// Try to get component type from instance or node
		componentType := node.ComponentType
		if componentType == "" && node.Component != nil && node.Component.Instance != nil {
			componentType = node.Component.Instance.GetComponentType()
		}

		if componentType != "" {
			switch componentType {
			case "header":
				return 80
			case "text":
				return 40
			case "list":
				return 80
			case "input":
				return 40
			case "button":
				return 20
			default:
				return 50
			}
		}
	} else if node.ComponentType != "" {
		switch node.ComponentType {
		case "header":
			return 80
		case "text":
			return 40
		case "list":
			return 80
		case "input":
			return 40
		case "button":
			return 20
		default:
			return 50
		}
	}

	return 20
}

func (e *Engine) measureChildHeight(node *LayoutNode, width int) int {
	if node.Style != nil && node.Style.Height != nil {
		switch v := node.Style.Height.Value.(type) {
		case string:
			if v == "flex" {
				return 0 // Flex 的高度将在 distributeFlexChildren 中计算
			}
		case int:
			if v > 0 {
				return v
			}
		case float64:
			if v > 0 {
				return int(v)
			}
		}
	}

	props := e.getProps(node)

	if node.Component != nil && node.Component.Instance != nil {
		if config := node.Component.LastConfig; config.Height > 0 {
			return config.Height
		}

		if config := node.Component.LastConfig; config.Width > 0 {
			width = config.Width
		}

		renderConfig := core.RenderConfig{
			Width:  width,
			Height: 1000,
			Data:   props,
		}
		content, err := node.Component.Instance.Render(renderConfig)
		if err == nil {
			// ✅ 改进：使用 runewidth 计算行高（考虑中文换行）
			lines := strings.Split(content, "\n")
			lineCount := len(lines)
			if lineCount > 0 && lineCount < 1000 {
				return lineCount
			}
		}
	}

	// Try to determine component type
	componentType := node.ComponentType
	if componentType == "" && node.Component != nil && node.Component.Instance != nil {
		componentType = node.Component.Instance.GetComponentType()
	}

	if componentType != "" {
		switch componentType {
		case "header":
			return 3
		case "text":
			// Try to estimate text height based on content and width
			if props != nil {
				if content, ok := props["content"].(string); ok {
					// Use reflow/wrap for accurate height calculation with ANSI support
					if width > 0 {
						wrapped := wrap.String(content, width)
						return strings.Count(wrapped, "\n") + 1
					}
					// Fallback if width is unknown (should assume 1 line or max width)
					return 1
				}
			}
			return 1
		case "list":
			// Try to estimate list height based on items count
			if props != nil {
				// Check for items array
				if items, ok := props["items"].([]interface{}); ok {
					// Limit default height to something reasonable
					count := len(items)
					if count == 0 {
						return 5 // Empty list default
					}
					if count > 20 {
						return 20 // Max default height
					}
					return count + 2 // +2 for title/border
				}
				// Check for bound data
				if bindData, ok := props["__bind_data"].([]interface{}); ok {
					count := len(bindData)
					if count == 0 {
						return 5
					}
					if count > 20 {
						return 20
					}
					return count + 2
				}
			}
			return 10
		case "input":
			return 1
		case "button":
			return 1
		default:
			return 5
		}
	}

	return 1
}

func nodeFromInfo(info *flexChildInfo) *LayoutNode {
	return info.Node
}

func nodeStyleGap(node *LayoutNode) int {
	if node != nil && node.Style != nil {
		return node.Style.Gap
	}
	return 0
}

func FindNodeByID(root *LayoutNode, id string) *LayoutNode {
	if root == nil {
		return nil
	}
	if root.ID == id {
		return root
	}
	for _, child := range root.Children {
		if found := FindNodeByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

func GetNodePath(root *LayoutNode, id string) []*LayoutNode {
	var path []*LayoutNode
	findPath(root, id, &path)
	return path
}

func findPath(node *LayoutNode, id string, path *[]*LayoutNode) bool {
	if node == nil {
		return false
	}
	*path = append(*path, node)
	if node.ID == id {
		return true
	}
	for _, child := range node.Children {
		if findPath(child, id, path) {
			return true
		}
	}
	*path = (*path)[:len(*path)-1]
	return false
}

func ValidateLayoutTree(node *LayoutNode, parent *LayoutNode) error {
	if node == nil {
		return nil
	}

	if node.Parent != parent {
		return fmt.Errorf("node '%s' has incorrect parent", node.ID)
	}

	for i, child := range node.Children {
		if child.Parent != node {
			return fmt.Errorf("child %d of node '%s' has incorrect parent", i, node.ID)
		}
		if err := ValidateLayoutTree(child, node); err != nil {
			return err
		}
	}

	return nil
}
