package tui

import (
	"fmt"

	tuiruntime "github.com/yaoapp/yao/tui/tui/runtime"
)

// RuntimeAdapter 将 Model 的 DSL Component 转换为 Runtime 可用的 LayoutNode 格式
// 这是 Model 层和 Runtime 层之间的桥接器
type RuntimeAdapter struct {
	model *Model
}

// NewRuntimeAdapter 创建一个新的 Runtime 适配器
func NewRuntimeAdapter(model *Model) *RuntimeAdapter {
	return &RuntimeAdapter{model: model}
}

// ToRuntimeLayoutNode 将 DSL Component 转换为 tuiruntime.LayoutNode
// 这是递归转换，会处理整个组件树
func (a *RuntimeAdapter) ToRuntimeLayoutNode(comp *Component) *tuiruntime.LayoutNode {
	if comp == nil {
		return nil
	}

	// 构建基础的 tuiruntime.LayoutNode
	node := tuiruntime.NewLayoutNode(
		comp.ID,
		a.mapComponentType(comp.Type),
		a.mapStyle(comp),
	)

	// 设置 Position（绝对定位）
	node.Position = a.mapPosition(comp)

	// 设置 Props（用于调试，Runtime 不直接使用）
	node.Props = comp.Props

	// 尝试从 ComponentInstanceRegistry 获取组件实例
	// 首先检查 m.Components（交互式组件），然后检查 ComponentInstanceRegistry（所有组件）
	if comp.ID != "" && comp.ID != "root" {
		// First check m.Components (for interactive components)
		if instance, exists := a.model.Components[comp.ID]; exists {
			node.Component = tuiruntime.NewComponentRef(instance.ID, instance.Type, instance.Instance)
		} else {
			// Then check ComponentInstanceRegistry (for all components including non-interactive)
			if instance, exists := a.model.ComponentInstanceRegistry.Get(comp.ID); exists {
				node.Component = tuiruntime.NewComponentRef(instance.ID, instance.Type, instance.Instance)
			}
		}
	}

	// 递归处理子组件
	if comp.Children != nil {
		for _, child := range comp.Children {
			childNode := a.ToRuntimeLayoutNode(&child)
			if childNode != nil {
				node.AddChild(childNode)
			}
		}
	}

	return node
}

// mapComponentType 将 TUI 组件类型映射到 Runtime NodeType
func (a *RuntimeAdapter) mapComponentType(compType string) tuiruntime.NodeType {
	switch compType {
	case "text", "header", "footer", "static":
		return tuiruntime.NodeTypeText
	case "row", "flex", "hbox":
		return tuiruntime.NodeTypeRow
	case "column", "vbox":
		return tuiruntime.NodeTypeColumn
	case "viewport", "scroll":
		return tuiruntime.NodeTypeCustom // 使用 Custom 表示特殊组件
	default:
		return tuiruntime.NodeTypeCustom
	}
}

// mapStyle 将 DSL Component 映射到 tuiruntime.Style
// 注意：大部分样式属性来自 Props，而不是 Component 的直接字段
func (a *RuntimeAdapter) mapStyle(comp *Component) tuiruntime.Style {
	style := tuiruntime.NewStyle()

	// 处理 Props 中的样式属性
	if comp.Props != nil {
		// Width
		if w, ok := propInt(comp.Props, "width"); ok {
			style = style.WithWidth(w)
		} else if propStringEquals(comp.Props, "width", "flex") {
			style = style.WithFlexGrow(1.0)
		}

		// Height
		if h, ok := propInt(comp.Props, "height"); ok {
			style = style.WithHeight(h)
		} else if propStringEquals(comp.Props, "height", "flex") {
			style = style.WithFlexGrow(1.0)
		}

		// FlexGrow
		if fg, ok := propFloat(comp.Props, "flexGrow"); ok {
			style = style.WithFlexGrow(fg)
		} else if fg, ok := propFloat(comp.Props, "flex"); ok {
			style = style.WithFlexGrow(fg)
		}

		// Direction
		if d, ok := propString(comp.Props, "direction"); ok {
			style = style.WithDirection(mapDirection(d))
		}

		// Padding
		if padding, ok := propIntArray(comp.Props, "padding"); ok && len(padding) >= 4 {
			style = style.WithPadding(tuiruntime.Insets{
				Top:    padding[0],
				Right:  padding[1],
				Bottom: padding[2],
				Left:   padding[3],
			})
		} else if p, ok := propInt(comp.Props, "padding"); ok {
			// 单值 padding
			style = style.WithPadding(tuiruntime.Insets{
				Top:    p,
				Right:  p,
				Bottom: p,
				Left:   p,
			})
		}

		// Border
		if border, ok := propIntArray(comp.Props, "border"); ok && len(border) >= 4 {
			style = style.WithBorder(tuiruntime.Insets{
				Top:    border[0],
				Right:  border[1],
				Bottom: border[2],
				Left:   border[3],
			})
		} else if b, ok := propInt(comp.Props, "border"); ok {
			style = style.WithBorder(tuiruntime.Insets{
				Top:    b,
				Right:  b,
				Bottom: b,
				Left:   b,
			})
		}

		// Gap
		if g, ok := propInt(comp.Props, "gap"); ok {
			style = style.WithGap(g)
		}

		// AlignItems
		if align, ok := propString(comp.Props, "alignItems"); ok {
			style = style.WithAlignItems(mapAlign(align))
		}

		// JustifyContent
		if justify, ok := propString(comp.Props, "justifyContent"); ok {
			style = style.WithJustify(mapJustify(justify))
		}

		// ZIndex
		if z, ok := propInt(comp.Props, "zIndex"); ok {
			style = style.WithZIndex(z)
		}

		// Overflow
		if overflow, ok := propString(comp.Props, "overflow"); ok {
			style = style.WithOverflow(mapOverflow(overflow))
		}
	}

	// 从 Component 的直接字段处理（用于兼容）
	// Width
	if comp.Width != nil {
		width, ok := toInt(comp.Width)
		if ok {
			style = style.WithWidth(width)
		} else if propStringEqualsInterface(comp.Width, "flex") {
			style = style.WithFlexGrow(1.0)
		}
	}

	// Height
	if comp.Height != nil {
		height, ok := toInt(comp.Height)
		if ok {
			style = style.WithHeight(height)
		} else if propStringEqualsInterface(comp.Height, "flex") {
			style = style.WithFlexGrow(1.0)
		}
	}

	// Direction
	if comp.Direction != "" {
		style = style.WithDirection(mapDirection(comp.Direction))
	}

	return style
}

// ========== 辅助函数：从 Props 中提取值 ==========

func propInt(props map[string]interface{}, key string) (int, bool) {
	if props == nil {
		return 0, false
	}
	val, exists := props[key]
	if !exists {
		return 0, false
	}

	switch v := val.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	case float32:
		return int(v), true
	default:
		return 0, false
	}
}

func propFloat(props map[string]interface{}, key string) (float64, bool) {
	if props == nil {
		return 0, false
	}
	val, exists := props[key]
	if !exists {
		return 0, false
	}

	switch v := val.(type) {
	case int:
		return float64(v), true
	case float64:
		return v, true
	case float32:
		return float64(v), true
	default:
		return 0, false
	}
}

func propString(props map[string]interface{}, key string) (string, bool) {
	if props == nil {
		return "", false
	}
	val, exists := props[key]
	if !exists {
		return "", false
	}
	if str, ok := val.(string); ok {
		return str, true
	}
	return "", false
}

func propStringEquals(props map[string]interface{}, key string, value string) bool {
	s, ok := propString(props, key)
	return ok && s == value
}

func propStringEqualsInterface(val interface{}, value string) bool {
	if s, ok := val.(string); ok {
		return s == value
	}
	return false
}

func propIntArray(props map[string]interface{}, key string) ([]int, bool) {
	if props == nil {
		return nil, false
	}
	val, exists := props[key]
	if !exists {
		return nil, false
	}

	switch v := val.(type) {
	case []int:
		return v, true
	case []interface{}:
		result := make([]int, 0)
		for _, item := range v {
			if i, ok := item.(int); ok {
				result = append(result, i)
			}
		}
		return result, len(result) > 0
	case []string:
		result := make([]int, 0)
		for _, s := range v {
			var i int
			_, err := fmt.Sscanf(s, "%d", &i)
			if err == nil {
				result = append(result, i)
			}
		}
		return result, len(result) > 0
	default:
		return nil, false
	}
}

// toInt 将 interface{} 转换为 int
func toInt(v interface{}) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case float64:
		return int(val), true
	case float32:
		return int(val), true
	case string:
		if val == "flex" {
			return 0, false
		}
		var result int
		_, err := fmt.Sscanf(val, "%d", &result)
		return result, err == nil
	default:
		return 0, false
	}
}

// ========== 类型映射函数 ==========

// mapDirection 将布局方向映射到 Runtime Direction
func mapDirection(d string) tuiruntime.Direction {
	switch d {
	case "row", "horizontal", "hbox":
		return tuiruntime.DirectionRow
	case "column", "vertical", "vbox":
		return tuiruntime.DirectionColumn
	default:
		return tuiruntime.DirectionColumn
	}
}

// mapAlign 将对齐方式映射到 Runtime Align
func mapAlign(align string) tuiruntime.Align {
	switch align {
	case "start", "left", "top":
		return tuiruntime.AlignStart
	case "center", "middle":
		return tuiruntime.AlignCenter
	case "end", "right", "bottom":
		return tuiruntime.AlignEnd
	case "stretch":
		return tuiruntime.AlignStretch
	default:
		return tuiruntime.AlignStart
	}
}

// mapJustify 将主轴对齐映射到 Runtime Justify
func mapJustify(justify string) tuiruntime.Justify {
	switch justify {
	case "start", "left", "top":
		return tuiruntime.JustifyStart
	case "center", "middle":
		return tuiruntime.JustifyCenter
	case "end", "right", "bottom":
		return tuiruntime.JustifyEnd
	case "space-between":
		return tuiruntime.JustifySpaceBetween
	case "space-around":
		return tuiruntime.JustifySpaceAround
	case "space-evenly":
		return tuiruntime.JustifySpaceEvenly
	default:
		return tuiruntime.JustifyStart
	}
}

// mapOverflow 将溢出处理映射到 Runtime Overflow
func mapOverflow(overflow string) tuiruntime.Overflow {
	switch overflow {
	case "visible":
		return tuiruntime.OverflowVisible
	case "hidden":
		return tuiruntime.OverflowHidden
	case "scroll":
		return tuiruntime.OverflowScroll
	default:
		return tuiruntime.OverflowVisible
	}
}

// mapPosition 将位置属性映射到 Runtime Position
func (a *RuntimeAdapter) mapPosition(comp *Component) tuiruntime.Position {
	position := tuiruntime.NewPosition() // Default: relative

	// Check for position type in component-level Style field
	if comp.Style != nil {
		if posStr, ok := propString(comp.Style, "position"); ok {
			switch posStr {
			case "absolute":
				position.Type = tuiruntime.PositionAbsolute
			case "relative":
				position.Type = tuiruntime.PositionRelative
			}
		}

		// Parse offsets from component-level Style field
		if position.Type == tuiruntime.PositionAbsolute {
			if top, ok := propInt(comp.Style, "top"); ok {
				position.Top = &top
			}
			if left, ok := propInt(comp.Style, "left"); ok {
				position.Left = &left
			}
			if right, ok := propInt(comp.Style, "right"); ok {
				position.Right = &right
			}
			if bottom, ok := propInt(comp.Style, "bottom"); ok {
				position.Bottom = &bottom
			}
		}
	}

	// Also check for position in props (for backwards compatibility)
	if comp.Props != nil {
		if posStr, ok := propString(comp.Props, "position"); ok {
			switch posStr {
			case "absolute":
				position.Type = tuiruntime.PositionAbsolute
			case "relative":
				position.Type = tuiruntime.PositionRelative
			}
		}

		// Check for position type in style field (nested in props)
		if style, ok := comp.Props["style"].(map[string]interface{}); ok {
			if posStr, ok := style["position"].(string); ok && posStr == "absolute" {
				position.Type = tuiruntime.PositionAbsolute
			}

			// Parse offsets from style field
			if position.Type == tuiruntime.PositionAbsolute {
				if top, ok := propInt(style, "top"); ok {
					position.Top = &top
				}
				if left, ok := propInt(style, "left"); ok {
					position.Left = &left
				}
				if right, ok := propInt(style, "right"); ok {
					position.Right = &right
				}
				if bottom, ok := propInt(style, "bottom"); ok {
					position.Bottom = &bottom
				}
			}
		}

		// Parse offsets from props (direct children)
		if position.Type == tuiruntime.PositionAbsolute {
			if top, ok := propInt(comp.Props, "top"); ok {
				position.Top = &top
			}
			if left, ok := propInt(comp.Props, "left"); ok {
				position.Left = &left
			}
			if right, ok := propInt(comp.Props, "right"); ok {
				position.Right = &right
			}
			if bottom, ok := propInt(comp.Props, "bottom"); ok {
				position.Bottom = &bottom
			}
		}
	}

	return position
}
