package tui

import (
	"fmt"
	"testing"

	"github.com/yaoapp/yao/tui/tui/component"
)

// TestListPropsResolution tests props resolution with debugging
func TestListPropsResolution(t *testing.T) {
	config := &Config{
		Name: "List Test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "list",
					ID:   "testList",
					Bind: "fruits",
					Props: map[string]interface{}{
						"height":       10,
						"width":        50,
						"itemTemplate": "{{id}}. {{name}} - {{price}}",
					},
				},
			},
		},
		Data: map[string]interface{}{
			"fruits": []interface{}{
				map[string]interface{}{"id": 1, "name": "Apple", "price": "$1.99"},
				map[string]interface{}{"id": 2, "name": "Banana", "price": "$0.99"},
			},
		},
	}

	model := NewModel(config, nil)
	model.InitializeComponents()

	// Get the component from the layout layout
	child := &model.Config.Layout.Children[0]

	// Check what props are resolved
	props := model.resolveProps(child)

	fmt.Printf("=== Resolved Props ===\n")
	fmt.Printf("Has __bind_data: %v\n", props["__bind_data"] != nil)
	if bindData, ok := props["__bind_data"]; ok {
		fmt.Printf("__bind_data type: %T\n", bindData)
		if items, ok := bindData.([]interface{}); ok {
			fmt.Printf("__bind_data length: %d\n", len(items))
			for i, item := range items {
				fmt.Printf("  Item %d: %v\n", i, item)
			}
		}
	}
	fmt.Printf("itemTemplate: %v\n", props["itemTemplate"])

	// Now test ParseListPropsWithBinding with these props
	listProps := component.ParseListPropsWithBinding(props)

	fmt.Printf("\n=== Parsed List Props ===\n")
	fmt.Printf("Items count: %d\n", len(listProps.Items))
	for i, item := range listProps.Items {
		fmt.Printf("Item %d: Title()='%s'\n", i, item.Title())
	}
}
