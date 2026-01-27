package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/tui/component"
)

// TestListComponentWithItemTemplate tests that the list component properly handles itemTemplate
func TestListComponentWithItemTemplate(t *testing.T) {
	props := map[string]interface{}{
		"itemTemplate": "{{id}}. {{name}} - {{price}}",
		"height":       12,
		"width":        50,
	}

	items := []interface{}{
		map[string]interface{}{"id": 1, "name": "Apple", "price": "$1.99"},
		map[string]interface{}{"id": 2, "name": "Banana", "price": "$0.99"},
		map[string]interface{}{"id": 3, "name": "Cherry", "price": "$2.99"},
	}
	props["__bind_data"] = items

	listProps := component.ParseListPropsWithBinding(props)

	// Verify items are parsed correctly
	assert.Equal(t, 3, len(listProps.Items))

	// Verify first item title uses the template
	assert.Equal(t, "1. Apple - $1.99", listProps.Items[0].Title())
	assert.Equal(t, "2. Banana - $0.99", listProps.Items[1].Title())
	assert.Equal(t, "3. Cherry - $2.99", listProps.Items[2].Title())

	// Verify height and width
	assert.Equal(t, 12, listProps.Height)
	assert.Equal(t, 50, listProps.Width)
}

// TestListComponentWithFallback tests fallback logic for item title
func TestListComponentWithFallback(t *testing.T) {
	props := map[string]interface{}{}

	items := []interface{}{
		map[string]interface{}{"id": 1, "name": "Test Item", "value": "custom"},
	}
	props["__bind_data"] = items

	listProps := component.ParseListPropsWithBinding(props)

	// Should use 'name' field as fallback
	assert.Equal(t, "Test Item", listProps.Items[0].Title())
	assert.Equal(t, "custom", listProps.Items[0].Value)
}

// TestListComponentWithTitleField tests that explicit 'title' field is prioritized
func TestListComponentWithTitleField(t *testing.T) {
	props := map[string]interface{}{
		"itemTemplate": "{{id}}. {{name}}",
	}

	items := []interface{}{
		map[string]interface{}{"id": 1, "name": "Apple", "title": "Custom Title"},
	}
	props["__bind_data"] = items

	listProps := component.ParseListPropsWithBinding(props)

	// Should prioritize 'title' field over template
	assert.Equal(t, "Custom Title", listProps.Items[0].Title())
}
