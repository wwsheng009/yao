// Package componentbinding provides integration between the binding package
// and the TUI component system. It extends components with data binding capabilities
// while keeping the core binding package independent of the framework.
package componentbinding

import (
	"github.com/yaoapp/yao/tui/framework/binding"
	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// PaintContext extends the component PaintContext with binding support.
type PaintContext struct {
	component.PaintContext
	Data binding.Context
}

// NewPaintContext creates a new paint context with binding support.
func NewPaintContext(base component.PaintContext, data binding.Context) PaintContext {
	return PaintContext{
		PaintContext: base,
		Data:         data,
	}
}

// WithData returns a new paint context with the given data context.
func (c PaintContext) WithData(data binding.Context) PaintContext {
	return PaintContext{
		PaintContext: c.PaintContext,
		Data:         data,
	}
}

// BindableComponent is a component that supports data binding.
type BindableComponent interface {
	component.Node
	// SetBinding sets a property binding
	SetBinding(key string, prop interface{})
	// GetBinding gets a property binding
	GetBinding(key string) (interface{}, bool)
	// BindAll sets multiple bindings
	BindAll(props map[string]interface{})
	// ResolveBindings resolves all bindings against a context
	ResolveBindings(ctx binding.Context) map[string]interface{}
}

// BaseBindable is a base implementation of BindableComponent.
type BaseBindable struct {
	*component.BaseComponent
	*component.StateHolder
	bindings map[string]interface{}
}

// NewBaseBindable creates a new base bindable component.
func NewBaseBindable(typ string) *BaseBindable {
	return &BaseBindable{
		BaseComponent: component.NewBaseComponent(typ),
		StateHolder:   component.NewStateHolder(),
		bindings:      make(map[string]interface{}),
	}
}

// SetBinding sets a property binding.
func (b *BaseBindable) SetBinding(key string, prop interface{}) {
	if b.bindings == nil {
		b.bindings = make(map[string]interface{})
	}
	b.bindings[key] = prop
}

// GetBinding gets a property binding.
func (b *BaseBindable) GetBinding(key string) (interface{}, bool) {
	if b.bindings == nil {
		return nil, false
	}
	prop, ok := b.bindings[key]
	return prop, ok
}

// BindAll sets multiple bindings.
func (b *BaseBindable) BindAll(props map[string]interface{}) {
	if b.bindings == nil {
		b.bindings = make(map[string]interface{})
	}
	for k, v := range props {
		b.bindings[k] = v
	}
}

// ResolveBindings resolves all bindings against a context.
func (b *BaseBindable) ResolveBindings(ctx binding.Context) map[string]interface{} {
	result := make(map[string]interface{})

	for key, prop := range b.bindings {
		switch p := prop.(type) {
		case binding.Prop[string]:
			result[key] = p.Resolve(ctx)
		case *binding.Prop[string]:
			result[key] = p.Resolve(ctx)
		case binding.Prop[int]:
			result[key] = p.Resolve(ctx)
		case *binding.Prop[int]:
			result[key] = p.Resolve(ctx)
		case binding.Prop[bool]:
			result[key] = p.Resolve(ctx)
		case *binding.Prop[bool]:
			result[key] = p.Resolve(ctx)
		default:
			// Static value
			result[key] = prop
		}
	}

	return result
}

// PaintWithBindings is a helper for painting with data binding.
func PaintWithBindings(comp BindableComponent, baseCtx component.PaintContext, buf *paint.Buffer, paintFunc func(PaintContext, *paint.Buffer)) {
	// Create binding context from component state
	ctx := CreateBindingContext(comp)

	// Resolve bindings
	_ = comp.ResolveBindings(ctx)

	// Create paint context with data
	paintCtx := NewPaintContext(baseCtx, ctx)

	// Call the paint function
	paintFunc(paintCtx, buf)
}

// CreateBindingContext creates a binding context from a component.
func CreateBindingContext(comp component.Node) binding.Context {
	// Start with root scope
	scope := binding.NewRootScope(make(map[string]interface{}))

	// Add component state if available
	if holder, ok := comp.(interface{ GetState() map[string]interface{} }); ok {
		state := holder.GetState()
		for k, v := range state {
			scope.Set(k, v)
		}
	}

	// Add component props if available
	if holder, ok := comp.(interface{ GetProps() map[string]interface{} }); ok {
		props := holder.GetProps()
		for k, v := range props {
			scope.Set(k, v)
		}
	}

	return scope
}

// ComponentWithProps helps create components with property bindings from DSL.
type ComponentWithProps struct {
	component component.Node
	props     map[string]interface{}
}

// NewComponentWithProps creates a component with properties.
func NewComponentWithProps(comp component.Node, props map[string]interface{}) *ComponentWithProps {
	return &ComponentWithProps{
		component: comp,
		props:     props,
	}
}

// ApplyProps applies properties to a component from a DSL map.
func ApplyProps(comp component.Node, props map[string]interface{}) {
	// Try to set bindings if component supports it
	if bindable, ok := comp.(BindableComponent); ok {
		bindable.BindAll(props)
	}

	// Try to set state holder props
	if holder, ok := comp.(interface{ SetProp(string, interface{}) }); ok {
		for k, v := range props {
			holder.SetProp(k, v)
		}
	}
}

// ParseDSLProps parses DSL properties into binding props.
func ParseDSLProps(dslProps map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range dslProps {
		switch val := v.(type) {
		case string:
			// Auto-detect binding syntax
			result[k] = binding.NewStringProp(val)
		case int:
			result[k] = binding.NewStatic(val)
		case bool:
			result[k] = binding.NewStatic(val)
		default:
			result[k] = v
		}
	}

	return result
}

// CreateListContexts creates contexts for list rendering.
func CreateListContexts(parent binding.Context, items []interface{}, props map[string]interface{}) []binding.Context {
	if props == nil {
		props = make(map[string]interface{})
	}

	// Add $item and $index props to each context
	contexts := make([]binding.Context, len(items))
	for i, item := range items {
		itemScope := parent.New(map[string]interface{}{
			"$index": i,
			"$item":  item,
		})

		// Flatten item properties if it's a map
		if m, ok := item.(map[string]interface{}); ok {
			for k, v := range m {
				itemScope.(*binding.Scope).Set(k, v)
			}
		}

		// Add additional props
		for k, v := range props {
			itemScope.(*binding.Scope).Set(k, v)
		}

		contexts[i] = itemScope
	}

	return contexts
}

// StoreToComponentAdapter integrates a ReactiveStore with a component.
type StoreToComponentAdapter struct {
	store    *binding.ReactiveStore
	component component.Node
	bindings map[string]string // prop -> store key
}

// NewStoreToComponentAdapter creates a new store adapter.
func NewStoreToComponentAdapter(store *binding.ReactiveStore, comp component.Node, bindings map[string]string) *StoreToComponentAdapter {
	return &StoreToComponentAdapter{
		store:    store,
		component: comp,
		bindings: bindings,
	}
}

// Sync syncs store values to component properties.
func (a *StoreToComponentAdapter) Sync() {
	ctx := a.store.ToContext()

	for prop, key := range a.bindings {
		if val, ok := ctx.Get(key); ok {
			if holder, ok := a.component.(interface{ SetProp(string, interface{}) }); ok {
				holder.SetProp(prop, val)
			}
		}
	}
}

// Watch watches the store and updates the component.
func (a *StoreToComponentAdapter) Watch(callback func()) func() {
	var cancels []func()

	for prop, key := range a.bindings {
		cancel := a.store.Subscribe(key, func(k string, old, new interface{}) {
			if holder, ok := a.component.(interface{ SetProp(string, interface{}) }); ok {
				holder.SetProp(prop, new)
			}
			if callback != nil {
				callback()
			}
		})
		cancels = append(cancels, cancel)
	}

	return func() {
		for _, cancel := range cancels {
			cancel()
		}
	}
}

// TwoWayBinding creates a two-way binding between store and component.
type TwoWayBinding struct {
	store       *binding.ReactiveStore
	component   component.Node
	prop        string
	key         string
	cancelStore func()
	cancelWatch func()
}

// NewTwoWayBinding creates a two-way binding.
func NewTwoWayBinding(store *binding.ReactiveStore, comp component.Node, prop, key string) *TwoWayBinding {
	twb := &TwoWayBinding{
		store:     store,
		component: comp,
		prop:      prop,
		key:       key,
	}

	// Watch store -> component
	twb.cancelStore = store.Subscribe(key, func(k string, old, new interface{}) {
		if holder, ok := comp.(interface{ SetProp(string, interface{}) }); ok {
			holder.SetProp(prop, new)
		}
		if holder, ok := comp.(interface{ MarkDirty() }); ok {
			holder.MarkDirty()
		}
	})

	return twb
}

// Dispose releases the binding.
func (b *TwoWayBinding) Dispose() {
	if b.cancelStore != nil {
		b.cancelStore()
	}
	if b.cancelWatch != nil {
		b.cancelWatch()
	}
}
