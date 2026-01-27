package tui

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/tui/component"
)

// TestListRenderWithFullFlow tests the full flow from parsing to rendering
func TestListRenderWithFullFlow(t *testing.T) {
	// Parse props (this works correctly)
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
	}
	props["__bind_data"] = items

	listProps := component.ParseListPropsWithBinding(props)

	fmt.Printf("=== After ParseListPropsWithBinding ===\n")
	for i, li := range listProps.Items {
		fmt.Printf("Item %d: Title()='%s', Description()='%s', FilterValue()='%s'\n",
			i, li.Title(), li.Description(), li.FilterValue())
	}

	// Convert to Bubble Tea list items
	bubbleItems := make([]list.Item, len(listProps.Items))
	for i, item := range listProps.Items {
		bubbleItems[i] = item
	}

	// Create list model with custom delegate
	delegate := &component.ListItemDelegate{}
	l := list.New(bubbleItems, delegate, 50, 10)

	// Check what the delegate sees
	visibleCount := len(bubbleItems)
	if visibleCount > l.Height()-3 { // Approximation of visible items
		visibleCount = l.Height() - 3
	}
	for i := 0; i < visibleCount; i++ {
		it := l.Index() + i
		if it >= len(bubbleItems) {
			break
		}
		item := bubbleItems[it]
		if listItem, ok := item.(component.ListItem); ok {
			fmt.Printf("Visible item %d (index %d): Title()='%s'\n", i, it, listItem.Title())
		}
	}

	// Render
	rendered := l.View()
	fmt.Printf("\n=== List Model View ===\n%s\n", rendered)

	assert.Contains(t, rendered, "Apple", "List should contain 'Apple'")
	assert.Contains(t, rendered, "Banana", "List should contain 'Banana'")
}
