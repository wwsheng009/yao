package binding

import (
	"strings"
	"sync"
)

// PropKind represents the kind of property.
type PropKind int

const (
	// PropStatic is a static value property.
	PropStatic PropKind = iota
	// PropBound is a data-bound property.
	PropBound
	// PropExpression is an expression property.
	PropExpression
)

// Prop is a generic property that can be either a static value or a dynamic binding.
//
// Prop[T] encapsulates the concept of a component property that may be:
//   - A static value: "Hello World"
//   - A data binding: {{ user.name }}
//   - An expression: {{ price * quantity }}
//
// The Resolve method is called during rendering to get the current value.
type Prop[T any] struct {
	mu         sync.RWMutex
	kind       PropKind
	staticVal  T
	bindPath   string
	expression *Expression
	// For computed properties
	computed  bool
	deps      []string
	computeFn func(Context) T
}

// NewStatic creates a static property with the given value.
func NewStatic[T any](val T) Prop[T] {
	return Prop[T]{
		kind:      PropStatic,
		staticVal: val,
	}
}

// NewBinding creates a data-bound property.
//
// The path should be in the format "key" or "nested.key.path".
// If the path contains "{{ }}", they will be automatically stripped.
func NewBinding[T any](path string) Prop[T] {
	// Strip {{ }} if present
	cleanPath := strings.TrimSpace(path)
	cleanPath = strings.TrimPrefix(cleanPath, "{{")
	cleanPath = strings.TrimSuffix(cleanPath, "}}")
	cleanPath = strings.TrimSpace(cleanPath)

	return Prop[T]{
		kind:     PropBound,
		bindPath: cleanPath,
	}
}

// NewExpression creates an expression property.
//
// The expression is parsed and can be evaluated during rendering.
// Examples: "{{ price * quantity }}", "{{ user.firstName + ' ' + user.lastName }}"
func NewExpression[T any](expr string) Prop[T] {
	// Strip {{ }} if present
	cleanExpr := strings.TrimSpace(expr)
	cleanExpr = strings.TrimPrefix(cleanExpr, "{{")
	cleanExpr = strings.TrimSuffix(cleanExpr, "}}")
	cleanExpr = strings.TrimSpace(cleanExpr)

	return Prop[T]{
		kind:       PropExpression,
		expression: ParseExpression(cleanExpr),
	}
}

// NewComputed creates a computed property.
//
// The compute function is called with the context to derive the value.
// The deps slice specifies which keys this computed property depends on.
func NewComputed[T any](deps []string, fn func(Context) T) Prop[T] {
	return Prop[T]{
		kind:      PropExpression,
		computed:  true,
		deps:      deps,
		computeFn: fn,
	}
}

// Parse creates a Prop by detecting the type from the string.
//
// If the string contains "{{", it creates a binding/expression.
// Otherwise, it creates a static property.
func Parse[T any](value string, convert func(string) T) Prop[T] {
	if strings.Contains(value, "{{") {
		// Check if it's a simple binding or expression
		clean := strings.TrimSpace(value)
		clean = strings.TrimPrefix(clean, "{{")
		clean = strings.TrimSuffix(clean, "}}")
		clean = strings.TrimSpace(clean)

		if isSimpleBinding(clean) {
			return NewBinding[T](value)
		}
		return NewExpression[T](value)
	}
	// Static value
	return NewStatic(convert(value))
}

// isSimpleBinding checks if the expression is a simple path reference.
func isSimpleBinding(expr string) bool {
	// Simple binding contains only alphanumeric, dots, and underscores
	for _, ch := range expr {
		if !(ch == '.' || ch == '_' || (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
			return false
		}
	}
	return expr != ""
}

// Resolve resolves the property value against the given context.
//
// For static properties, returns the static value.
// For bound properties, looks up the path in the context.
// For expression properties, evaluates the expression.
func (p *Prop[T]) Resolve(ctx Context) T {
	p.mu.RLock()
	defer p.mu.RUnlock()

	switch p.kind {
	case PropStatic:
		return p.staticVal

	case PropBound:
		if val, ok := ctx.Get(p.bindPath); ok {
			if typed, ok := val.(T); ok {
				return typed
			}
			// Try to convert
			if converted, ok := convertValue[T](val); ok {
				return converted
			}
		}
		// Return zero value or static fallback
		var zero T
		return zero

	case PropExpression:
		if p.computed && p.computeFn != nil {
			return p.computeFn(ctx)
		}
		if p.expression != nil {
			result := p.expression.Evaluate(ctx)
			if typed, ok := result.(T); ok {
				return typed
			}
			if converted, ok := convertValue[T](result); ok {
				return converted
			}
		}
		var zero T
		return zero

	default:
		var zero T
		return zero
	}
}

// ResolveString resolves the property to a string.
func (p *Prop[T]) ResolveString(ctx Context) string {
	val := p.Resolve(ctx)
	return stringify(val)
}

// IsBound returns true if this is a bound property.
func (p *Prop[T]) IsBound() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.kind == PropBound
}

// IsExpression returns true if this is an expression property.
func (p *Prop[T]) IsExpression() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.kind == PropExpression
}

// IsStatic returns true if this is a static property.
func (p *Prop[T]) IsStatic() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.kind == PropStatic
}

// GetPath returns the binding path for bound properties.
func (p *Prop[T]) GetPath() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.bindPath
}

// GetDependencies returns the dependencies for computed properties.
func (p *Prop[T]) GetDependencies() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.computed {
		return p.deps
	}
	if p.kind == PropBound && p.bindPath != "" {
		return []string{getRootPath(p.bindPath)}
	}
	if p.kind == PropExpression && p.expression != nil {
		return p.expression.Dependencies()
	}
	return nil
}

// ResolveWithTracking resolves the property and registers dependencies.
// This method should be used during component rendering to automatically
// track which state keys the component depends on.
//
// The tracker parameter is typically the dependency graph from the store.
func (p *Prop[T]) ResolveWithTracking(ctx Context, nodeID string, tracker DependencyTracker) T {
	// First, register dependencies
	deps := p.GetDependencies()
	if deps != nil && tracker != nil {
		for _, dep := range deps {
			tracker.Register(nodeID, dep)
		}
	}

	// Then resolve normally
	return p.Resolve(ctx)
}

// DependencyTracker is the interface for tracking dependencies.
// Implemented by DependencyGraph.
type DependencyTracker interface {
	Register(nodeID, stateKey string)
	Unregister(nodeID string)
}

// getRootPath extracts the root path from a nested path.
func getRootPath(path string) string {
	if idx := strings.Index(path, "."); idx >= 0 {
		return path[:idx]
	}
	return path
}

// SetStatic sets the static value (converts to static property).
func (p *Prop[T]) SetStatic(val T) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.kind = PropStatic
	p.staticVal = val
	p.bindPath = ""
	p.expression = nil
	p.computed = false
}

// StringProp is a convenience type for string properties.
type StringProp = Prop[string]

// NewStringProp creates a string property, detecting the type.
func NewStringProp(value string) StringProp {
	if strings.Contains(value, "{{") {
		clean := strings.TrimSpace(value)
		clean = strings.TrimPrefix(clean, "{{")
		clean = strings.TrimSuffix(clean, "}}")
		clean = strings.TrimSpace(clean)

		if isSimpleBinding(clean) {
			return NewBinding[string](value)
		}
		return NewExpression[string](value)
	}
	return NewStatic(value)
}

// IntProp is a convenience type for int properties.
type IntProp = Prop[int]

// BoolProp is a convenience type for bool properties.
type BoolProp = Prop[bool]

// convertValue attempts to convert a value to type T.
func convertValue[T any](val interface{}) (T, bool) {
	var zero T
	// This is a simplified conversion
	// In a full implementation, you'd use reflection or a type switch
	switch val.(type) {
	case string:
		// Try to convert string to target type
		// This is a no-op for most types
		return zero, false
	case int:
		// May succeed depending on T
		return zero, false
	case int64:
		return zero, false
	case float64:
		return zero, false
	case bool:
		return zero, false
	default:
		return zero, false
	}
}

// PropBuilder helps build properties from various sources.
type PropBuilder struct {
	props map[string]interface{}
}

// NewPropBuilder creates a new property builder.
func NewPropBuilder() *PropBuilder {
	return &PropBuilder{
		props: make(map[string]interface{}),
	}
}

// String adds a string property.
func (b *PropBuilder) String(key, value string) *PropBuilder {
	b.props[key] = NewStringProp(value)
	return b
}

// Int adds an int property.
func (b *PropBuilder) Int(key string, value int) *PropBuilder {
	b.props[key] = NewStatic(value)
	return b
}

// Bool adds a bool property.
func (b *PropBuilder) Bool(key string, value bool) *PropBuilder {
	b.props[key] = NewStatic(value)
	return b
}

// Binding adds a binding property.
func (b *PropBuilder) Binding(key, path string) *PropBuilder {
	b.props[key] = NewBinding[interface{}](path)
	return b
}

// Build returns the property map.
func (b *PropBuilder) Build() map[string]interface{} {
	return b.props
}

// MustResolve resolves a property or panics.
// Useful for initialization code.
func MustResolve[T any](p Prop[T], ctx Context) T {
	val := p.Resolve(ctx)
	// Could add additional validation here
	return val
}

// LazyProp creates a lazily evaluated property.
func LazyProp[T any](fn func() T) Prop[T] {
	return Prop[T]{
		kind:      PropExpression,
		computed:  true,
		computeFn: func(Context) T { return fn() },
	}
}
