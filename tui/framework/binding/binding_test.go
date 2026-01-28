package binding

import (
	"testing"
)

func TestScope_Get(t *testing.T) {
	root := NewRootScope(map[string]interface{}{
		"app": map[string]interface{}{
			"name": "TestApp",
			"version": "1.0.0",
		},
		"user": map[string]interface{}{
			"name": "Alice",
			"email": "alice@example.com",
		},
	})

	tests := []struct {
		name     string
		path     string
		want     interface{}
		wantOk   bool
		compareFunc func(got, want interface{}) bool
	}{
		{"simple key", "app", map[string]interface{}{"name": "TestApp", "version": "1.0.0"}, true, compareMaps},
		{"nested path", "app.name", "TestApp", true, nil},
		{"nested path deep", "user.email", "alice@example.com", true, nil},
		{"missing key", "missing", nil, false, nil},
		{"missing nested", "app.missing", nil, false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := root.Get(tt.path)
			if ok != tt.wantOk {
				t.Errorf("Get() ok = %v, want %v", ok, tt.wantOk)
				return
			}
			if ok && tt.compareFunc != nil {
				if !tt.compareFunc(got, tt.want) {
					t.Errorf("Get() = %v, want %v", got, tt.want)
				}
			} else if ok && got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to compare map values
func compareMaps(got, want interface{}) bool {
	gotMap, ok1 := got.(map[string]interface{})
	wantMap, ok2 := want.(map[string]interface{})
	if !ok1 || !ok2 {
		return false
	}
	if len(gotMap) != len(wantMap) {
		return false
	}
	for k, v := range wantMap {
		if gotMap[k] != v {
			return false
		}
	}
	return true
}

func TestScope_Chain(t *testing.T) {
	root := NewRootScope(map[string]interface{}{
		"global": "root-value",
	})

	child := root.New(map[string]interface{}{
		"local": "child-value",
	}).(*Scope)

	tests := []struct {
		name   string
		scope  *Scope
		path   string
		want   string
		wantOk bool
	}{
		{"child local", child, "local", "child-value", true},
		{"child inherits", child, "global", "root-value", true},
		{"root local", root, "local", "", false},
		{"root global", root, "global", "root-value", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.scope.Get(tt.path)
			if ok != tt.wantOk {
				t.Errorf("Get() ok = %v, want %v", ok, tt.wantOk)
				return
			}
			if ok && got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScope_Set(t *testing.T) {
	scope := NewRootScope(nil)

	tests := []struct {
		name   string
		key    string
		value  interface{}
		wantOk bool
	}{
		{"simple", "key", "value", true},
		{"nested", "a.b.c", "deep", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ok := scope.Set(tt.key, tt.value); ok != tt.wantOk {
				t.Errorf("Set() ok = %v, want %v", ok, tt.wantOk)
			}
			if tt.wantOk {
				got, ok := scope.Get(tt.key)
				if !ok || got != tt.value {
					t.Errorf("After Set(), Get() = %v, %v; want %v", got, ok, tt.value)
				}
			}
		})
	}
}

func TestProp_Static(t *testing.T) {
	prop := NewStatic("hello")

	ctx := NewRootScope(nil)
	got := prop.Resolve(ctx)

	if got != "hello" {
		t.Errorf("Static Prop.Resolve() = %v, want hello", got)
	}
}

func TestProp_Binding(t *testing.T) {
	prop := NewBinding[string]("user.name")

	ctx := NewRootScope(map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
		},
	})

	got := prop.Resolve(ctx)
	if got != "Alice" {
		t.Errorf("Binding Prop.Resolve() = %v, want Alice", got)
	}
}

func TestProp_Parse(t *testing.T) {
	tests := []struct {
		name  string
		value string
		isBound bool
	}{
		{"static", "hello", false},
		{"binding", "{{ user.name }}", true},
		{"binding spaced", "{{user.name}}", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop := NewStringProp(tt.value)
			if prop.IsBound() != tt.isBound {
				t.Errorf("Parse(%q).IsBound() = %v, want %v", tt.value, prop.IsBound(), tt.isBound)
			}
		})
	}
}

func TestReactiveStore_Basic(t *testing.T) {
	store := NewReactiveStore()

	// Test Set and Get
	store.Set("key1", "value1")
	if val, ok := store.Get("key1"); !ok || val != "value1" {
		t.Errorf("After Set(), Get() = %v, %v; want value1, true", val, ok)
	}

	// Test Has
	if !store.Has("key1") {
		t.Error("Has() = false, want true")
	}
}

func TestReactiveStore_Subscribe(t *testing.T) {
	store := NewReactiveStore()

	called := false
	var notifiedKey string
	var notifiedNew interface{}

	cancel := store.Subscribe("key1", func(key string, old, new interface{}) {
		called = true
		notifiedKey = key
		_ = old // Not used in this test
		notifiedNew = new
	})

	store.Set("key1", "value1")

	if !called {
		t.Error("Subscribe callback was not called")
	}
	if notifiedKey != "key1" {
		t.Errorf("Notified key = %v, want key1", notifiedKey)
	}
	if notifiedNew != "value1" {
		t.Errorf("Notified new = %v, want value1", notifiedNew)
	}

	// Test cancel
	called = false
	cancel()
	store.Set("key1", "value2")

	if called {
		t.Error("Callback was called after cancel")
	}
}

func TestReactiveStore_Batch(t *testing.T) {
	store := NewReactiveStore()

	callCount := 0
	store.Subscribe("key1", func(key string, old, new interface{}) {
		callCount++
	})

	store.BeginBatch()
	store.Set("key1", "value1")
	store.Set("key1", "value2")
	store.Set("key1", "value3")
	store.EndBatch()

	if callCount != 1 {
		t.Errorf("Batch updates triggered %d notifications, want 1", callCount)
	}
}

func TestExpression_Basic(t *testing.T) {
	expr := ParseExpression("1 + 2")
	ctx := NewRootScope(nil)
	result := expr.Evaluate(ctx)

	if num, ok := result.(float64); !ok || num != 3 {
		t.Errorf("Expression.Evaluate() = %v, want 3", result)
	}
}

func TestExpression_Variable(t *testing.T) {
	expr := ParseExpression("price * quantity")
	ctx := NewRootScope(map[string]interface{}{
		"price": 10,
		"quantity": 5,
	})
	result := expr.Evaluate(ctx)

	if num, ok := result.(float64); !ok || num != 50 {
		t.Errorf("Expression.Evaluate() = %v, want 50", result)
	}
}

func TestExpression_Dependencies(t *testing.T) {
	expr := ParseExpression("price * quantity + tax")
	deps := expr.Dependencies()

	if len(deps) < 3 {
		t.Errorf("Dependencies() = %v, want at least 3", deps)
	}
}

func TestListContext(t *testing.T) {
	items := []interface{}{
		map[string]interface{}{"id": 1, "name": "Item 1"},
		map[string]interface{}{"id": 2, "name": "Item 2"},
		map[string]interface{}{"id": 3, "name": "Item 3"},
	}

	parent := NewRootScope(nil)
	contexts := ListContext(parent, items)

	if len(contexts) != len(items) {
		t.Fatalf("ListContext() returned %d contexts, want %d", len(contexts), len(items))
	}

	// Check first context
	idx, ok := contexts[0].Get("$index")
	if !ok || idx != 0 {
		t.Errorf("Context 0 $index = %v, %v; want 0, true", idx, ok)
	}

	name, ok := contexts[0].Get("name")
	if !ok || name != "Item 1" {
		t.Errorf("Context 0 name = %v, %v; want Item 1, true", name, ok)
	}
}

func TestScope_SpecialVariables(t *testing.T) {
	scope := NewRootScope(map[string]interface{}{
		"$index": 5,
		"$item": "test-item",
	})

	tests := []struct {
		path string
		want interface{}
		wantOk bool
	}{
		{"$index", 5, true},
		{"$item", "test-item", true},
		{"$parent", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got, ok := scope.Get(tt.path)
			if ok != tt.wantOk {
				t.Errorf("Get(%q) ok = %v, want %v", tt.path, ok, tt.wantOk)
				return
			}
			if ok && got != tt.want {
				t.Errorf("Get(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestComputed_Basic(t *testing.T) {
	store := NewReactiveStore()
	store.Set("price", 10)
	store.Set("quantity", 3)

	computed := NewStoreComputed(store, []string{"price", "quantity"}, func() interface{} {
		price, _ := store.Get("price")
		quantity, _ := store.Get("quantity")
		return price.(int) * quantity.(int)
	})

	val := computed.Value()
	if val != 30 {
		t.Errorf("Computed.Value() = %v, want 30", val)
	}

	// Update dependency
	store.Set("quantity", 5)

	val = computed.Value()
	if val != 50 {
		t.Errorf("After update, Computed.Value() = %v, want 50", val)
	}

	computed.Dispose()
}

func TestPropBuilder(t *testing.T) {
	props := NewPropBuilder().
		String("title", "Hello").
		Int("count", 42).
		Bool("enabled", true).
		Build()

	if len(props) != 3 {
		t.Errorf("PropBuilder.Build() returned %d props, want 3", len(props))
	}

	titleVal, ok := props["title"]
	if !ok {
		t.Error("PropBuilder.String() - title not found")
		return
	}

	// Try to convert to Prop[string]
	title, ok := titleVal.(Prop[string])
	if !ok {
		t.Errorf("PropBuilder.String() - title is %T, not Prop[string]", titleVal)
		return
	}

	if !title.IsStatic() {
		t.Error("PropBuilder.String() - title should be static")
	}
	if title.Resolve(nil) != "Hello" {
		t.Errorf("PropBuilder.String() - title.Resolve() = %v, want Hello", title.Resolve(nil))
	}
}
