package tui

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/yaoapp/yao/tui/components"
)

// TestListHeightInvestigation tests list height and viewport issues
func TestListHeightInvestigation(t *testing.T) {
	props := map[string]interface{}{
		"itemTemplate": "{{id}}. {{name}} - {{price}}",
		"height":       10,
		"width":        50,
	}

	items := []interface{}{
		map[string]interface{}{"id": 1, "name": "Apple", "price": "$1.99"},
		map[string]interface{}{"id": 2, "name": "Banana", "price": "$0.99"},
		map[string]interface{}{"id": 3, "name": "Cherry", "price": "$2.99"},
		map[string]interface{}{"id": 4, "name": "Date", "price": "$3.99"},
		map[string]interface{}{"id": 5, "name": "Elderberry", "price": "$4.99"},
	}
	props["__bind_data"] = items

	listProps := components.ParseListPropsWithBinding(props)

	bubbleItems := make([]list.Item, len(listProps.Items))
	for i, item := range listProps.Items {
		bubbleItems[i] = item
	}

	// Create list model with different heights
	testHeights := []int{5, 10, 15, 20}

	for _, height := range testHeights {
		fmt.Printf("\n=== Testing with height=%d ===\n", height)

		l := list.New(bubbleItems, list.NewDefaultDelegate(), 50, height)

		fmt.Printf("List model height: %d\n", l.Height())
		fmt.Printf("List model width: %d\n", l.Width())
		fmt.Printf("Total items: %d\n", len(bubbleItems))
		fmt.Printf("Current index: %d\n", l.Index())

		// Get the delegate's visible item count
		// The default delegate has Spacing()=0 and Height()=1
		// So visible items = list height minus title, status bar, pagination, help
		delegate := list.NewDefaultDelegate()
		fmt.Printf("Delegate height: %d, spacing: %d\n", delegate.Height(), delegate.Spacing())

		// Render
		rendered := l.View()
		fmt.Printf("Rendered view:\n%s\n", rendered)

		// Count how many items actually appear in the rendered output
		// Looking for lines that start with │
		itemCount := 0
		for _, line := range lines(rendered) {
			if len(line) > 0 && string(line[0]) == "│" {
				itemCount++
			}
		}
		fmt.Printf("Items actually rendered (lines starting with │): %d\n", itemCount)
	}
}

// lines splits a string into lines
func lines(s string) []string {
	result := []string{}
	current := ""
	for _, c := range s {
		if c == '\n' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
