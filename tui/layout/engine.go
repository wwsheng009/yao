package layout

import (
	"fmt"
	"math"
	"strings"

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
	e.layoutNode(e.root, 0, 0, e.window.Width, e.window.Height, result)

	return result
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

	for _, child := range node.Children {
		info := e.measureChild(child, config, width, height)
		allChildren = append(allChildren, info)
		if info.Grow.Value > 0 {
			growSum += info.Grow.Value
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

	// 确保可用空间不为负
	if availableSpace < 0 {
		availableSpace = 0
	}

	// 为 flex 子元素分配空间
	for _, info := range allChildren {
		if info.Grow.Value > 0 && growSum > 0 {
			extra := int(float64(availableSpace) * (info.Grow.Value / growSum))
			info.Size = extra
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
	if size != nil && size.Value != nil {
		switch v := size.Value.(type) {
		case float64:
			info.Size = int(v)
		case int:
			info.Size = v
		case string:
			if v == "flex" {
				info.Grow.Value = 1
				info.Size = 0 // 初始为0，稍后分配
			}
		default:
			// 未知类型，使用默认测量
			if config.Direction == DirectionRow {
				info.Size = e.measureChildWidth(child, parentHeight)
			} else {
				info.Size = e.measureChildHeight(child, parentWidth)
			}
		}
	} else {
		// size 为 nil 或 size.Value 为 nil，使用默认测量
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
		renderConfig := core.RenderConfig{
			Width:  200,
			Height: height,
			Data:   props,
		}

		if node.Component.Instance.GetComponentType() == "text" {
			if props != nil {
				if content, ok := props["content"].(string); ok {
					return len(content)
				}
			}
		}

		content, err := node.Component.Instance.Render(renderConfig)
		if err == nil {
			lines := strings.Split(content, "\n")
			maxWidth := 0
			for _, line := range lines {
				w := len(line)
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
					// Simple estimation: content length / width
					// This is a fallback if Render() above failed or wasn't called
					if width > 0 {
						return int(math.Ceil(float64(len(content)) / float64(width)))
					}
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
