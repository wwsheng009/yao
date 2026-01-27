package layout

import "fmt"

// LayoutHelper contains helper functions for layout operations
type LayoutHelper struct {
	// engine *Engine // Store reference to engine if needed
}

// ensureStyle ensures that a node has a proper style object
func ensureStyle(node *LayoutNode) {
	if node.Style == nil {
		node.Style = &LayoutStyle{
			Direction: DirectionColumn,
		}
	}
}

// calculateMetrics calculates layout metrics for a node
func calculateMetrics(node *LayoutNode, width, height int) {
	if node.Style == nil {
		return
	}

	// Calculate available space after padding
	innerWidth := width
	innerHeight := height

	if node.Style.Padding != nil {
		innerWidth = max(0, innerWidth-node.Style.Padding.Left-node.Style.Padding.Right)
		innerHeight = max(0, innerHeight-node.Style.Padding.Top-node.Style.Padding.Bottom)
	}

	node.AvailableWidth = innerWidth
	node.AvailableHeight = innerHeight
}

// getProps retrieves props for a node
func getProps(node *LayoutNode, propsResolver PropsResolverFunc) map[string]interface{} {
	if propsResolver != nil {
		return propsResolver(node)
	}
	return node.Props
}

// PropsResolverFunc is a function type for resolving props from a node
type PropsResolverFunc func(node *LayoutNode) map[string]interface{}

// getDefaultHeight returns default height for a component type
func getDefaultHeight(componentType string, props map[string]interface{}) int {
	if componentType == "" {
		return 5
	}

	switch componentType {
	case "header":
		return 3
	case "text":
		if props != nil {
			if content, ok := props["content"].(string); ok && content != "" {
				return 1
			}
		}
		return 1
	case "list":
		if props != nil {
			if items, ok := props["items"].([]interface{}); ok {
				count := len(items)
				if count == 0 {
					return 5
				}
				if count > 20 {
					return 20
				}
				return count + 2
			}
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

// getDefaultWidth returns default width for a component type
func getDefaultWidth(componentType string, props map[string]interface{}) int {
	if componentType == "" {
		return 50
	}

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

// getComponentType extracts component type from node
func getComponentType(node *LayoutNode) string {
	componentType := node.ComponentType
	if componentType == "" && node.Component != nil && node.Component.Instance != nil {
		componentType = node.Component.Instance.GetComponentType()
	}
	return componentType
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
