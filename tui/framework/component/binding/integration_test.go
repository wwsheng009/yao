package componentbinding

import (
	"testing"

	"github.com/yaoapp/yao/tui/framework/binding"
	"github.com/yaoapp/yao/tui/framework/component"
)

func TestParseDSLProps(t *testing.T) {
	dslProps := map[string]interface{}{
		"title":       "{{ user.name }}",
		"count":       "{{ item.count }}",
		"static":      "Hello",
		"number":      42,
		"enabled":     true,
		"description": "This is a test",
	}

	result := ParseDSLProps(dslProps)

	// Check that strings are converted to props
	if len(result) != 6 {
		t.Errorf("ParseDSLProps() returned %d props, want 6", len(result))
	}

	// Check binding detection
	title, ok := result["title"].(binding.Prop[string])
	if !ok {
		t.Error("title should be a binding.Prop[string]")
	} else if !title.IsBound() {
		t.Error("title should be bound")
	}

	// Check static detection
	static, ok := result["static"].(binding.Prop[string])
	if !ok {
		t.Error("static should be a binding.Prop[string]")
	} else if !static.IsStatic() {
		t.Error("static should be static")
	}
}

func TestCreateBindingContext(t *testing.T) {
	// Create a component with StateHolder
	bindable := NewBaseBindable("test")
	bindable.SetStateValue("key1", "value1")
	bindable.SetProp("prop1", "propValue1")

	ctx := CreateBindingContext(bindable)

	// Context should have access to values
	if val, ok := ctx.Get("key1"); !ok || val != "value1" {
		t.Errorf("Context should have key1=value1, got %v, %v", val, ok)
	}
}

func TestBindableComponent(t *testing.T) {
	bindable := NewBaseBindable("test")

	// Test SetBinding
	bindable.SetBinding("title", binding.NewStatic("Hello"))
	prop, ok := bindable.GetBinding("title")
	if !ok {
		t.Error("GetBinding() should return the prop")
	}
	if typed, ok := prop.(binding.Prop[string]); ok {
		if typed.Resolve(nil) != "Hello" {
			t.Error("Prop should resolve to Hello")
		}
	}

	// Test BindAll
	bindable.BindAll(map[string]interface{}{
		"name": binding.NewBinding[string]("user.name"),
		"age":  binding.NewStatic(25),
	})

	// Test ResolveBindings
	scope := binding.NewRootScope(map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
		},
	})
	resolved := bindable.ResolveBindings(scope)
	if resolved["title"] != "Hello" {
		t.Errorf("title should resolve to Hello, got %v", resolved["title"])
	}
	if resolved["name"] != "Alice" {
		t.Errorf("name should resolve to Alice, got %v", resolved["name"])
	}
}

func TestCreateListContexts(t *testing.T) {
	items := []interface{}{
		map[string]interface{}{"id": 1, "name": "Item 1"},
		map[string]interface{}{"id": 2, "name": "Item 2"},
	}

	parent := binding.NewRootScope(map[string]interface{}{
		"global": "root-value",
	})

	contexts := CreateListContexts(parent, items, nil)

	if len(contexts) != 2 {
		t.Fatalf("CreateListContexts() returned %d contexts, want 2", len(contexts))
	}

	// Check first context
	if idx, ok := contexts[0].Get("$index"); !ok || idx != 0 {
		t.Errorf("Context 0 $index = %v, %v; want 0, true", idx, ok)
	}

	// Check inherited value
	if val, ok := contexts[0].Get("global"); !ok || val != "root-value" {
		t.Errorf("Context should inherit global, got %v, %v", val, ok)
	}

	// Check item properties
	if name, ok := contexts[0].Get("name"); !ok || name != "Item 1" {
		t.Errorf("Context 0 name = %v, %v; want Item 1, true", name, ok)
	}
}

func TestPaintContext(t *testing.T) {
	ctx := binding.NewRootScope(map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
		},
	})

	// Create a mock base paint context
	baseCtx := component.PaintContext{}

	paintCtx := NewPaintContext(baseCtx, ctx)

	// Check that data is accessible
	if val, ok := paintCtx.Data.Get("user.name"); !ok || val != "Alice" {
		t.Errorf("PaintContext.Data should have user.name=Alice, got %v, %v", val, ok)
	}

	// Check WithData
	newCtx := paintCtx.WithData(binding.NewRootScope(map[string]interface{}{
		"key": "value",
	}))

	if val, ok := newCtx.Data.Get("key"); !ok || val != "value" {
		t.Errorf("WithData() should set new context, got %v, %v", val, ok)
	}
}
