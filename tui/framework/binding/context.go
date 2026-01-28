// Package binding provides data binding and reactive state management for TUI components.
//
// This package implements the "compile-time binding" pattern described in the
// state management design document, supporting:
//
//   - Static and dynamic property binding ({{ variable }} syntax)
//   - Scope chain for state inheritance
//   - Reactive store with fine-grained dependency tracking
//   - Expression parsing and evaluation
package binding

// Context defines the data context interface for property resolution.
//
// A Context provides data access for property binding resolution. Components
// use Context to resolve dynamic properties like {{ user.name }}.
type Context interface {
	// Get retrieves a value by path. Returns the value and true if found,
	// nil and false otherwise.
	//
	// Path can be:
	//   - A simple key: "name"
	//   - A nested path: "user.address.city"
	//   - A special variable: "$index", "$item"
	Get(path string) (interface{}, bool)

	// Set sets a value at the given path. Returns true if successful.
	Set(path string, value interface{}) bool

	// Has checks if a value exists at the given path.
	Has(path string) bool

	// Parent returns the parent context for scope chain traversal.
	// Returns nil if this is the root context.
	Parent() Context

	// Root returns the root context of the scope chain.
	Root() Context

	// New creates a new child scope with additional data.
	// The child scope inherits all data from this context.
	New(data map[string]interface{}) Context
}

// Resolver is a function that resolves a value from a context.
type Resolver func(ctx Context) interface{}

// ValueAccessor provides typed access to context values.
type ValueAccessor interface {
	// String returns the value as a string.
	String(path string, defaultValue string) string

	// Int returns the value as an integer.
	Int(path string, defaultValue int) int

	// Int64 returns the value as an int64.
	Int64(path string, defaultValue int64) int64

	// Float64 returns the value as a float64.
	Float64(path string, defaultValue float64) float64

	// Bool returns the value as a boolean.
	Bool(path string, defaultValue bool) bool

	// Map returns the value as a map.
	Map(path string) map[string]interface{}

	// Slice returns the value as a slice.
	Slice(path string) []interface{}
}

// NewValueAccessor creates a new ValueAccessor for the given context.
func NewValueAccessor(ctx Context) ValueAccessor {
	return &valueAccessor{ctx: ctx}
}

// valueAccessor implements ValueAccessor.
type valueAccessor struct {
	ctx Context
}

func (a *valueAccessor) String(path string, defaultValue string) string {
	if v, ok := a.ctx.Get(path); ok {
		if s, ok := v.(string); ok {
			return s
		}
		// Try to convert other types to string
		return stringify(v)
	}
	return defaultValue
}

func (a *valueAccessor) Int(path string, defaultValue int) int {
	if v, ok := a.ctx.Get(path); ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return defaultValue
}

func (a *valueAccessor) Int64(path string, defaultValue int64) int64 {
	if v, ok := a.ctx.Get(path); ok {
		switch val := v.(type) {
		case int:
			return int64(val)
		case int64:
			return val
		case float64:
			return int64(val)
		}
	}
	return defaultValue
}

func (a *valueAccessor) Float64(path string, defaultValue float64) float64 {
	if v, ok := a.ctx.Get(path); ok {
		switch val := v.(type) {
		case int:
			return float64(val)
		case int64:
			return float64(val)
		case float64:
			return val
		}
	}
	return defaultValue
}

func (a *valueAccessor) Bool(path string, defaultValue bool) bool {
	if v, ok := a.ctx.Get(path); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func (a *valueAccessor) Map(path string) map[string]interface{} {
	if v, ok := a.ctx.Get(path); ok {
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

func (a *valueAccessor) Slice(path string) []interface{} {
	if v, ok := a.ctx.Get(path); ok {
		if s, ok := v.([]interface{}); ok {
			return s
		}
	}
	return nil
}

// stringify converts various types to string.
func stringify(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int, int64, float64:
		return formatNumber(val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case nil:
		return ""
	default:
		return ""
	}
}

func formatNumber(v interface{}) string {
	// Simple number formatting
	switch val := v.(type) {
	case int:
		return itoa(val)
	case int64:
		return i64toa(val)
	case float64:
		return f64toa(val)
	default:
		return ""
	}
}

// Simple implementations to avoid fmt dependency
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf []byte
	for i > 0 {
		buf = append(buf, byte('0'+i%10))
		i /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

func i64toa(i int64) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf []byte
	for i > 0 {
		buf = append(buf, byte('0'+i%10))
		i /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

func f64toa(f float64) string {
	// Simplified float to string
	i := int64(f)
	return i64toa(i)
}
