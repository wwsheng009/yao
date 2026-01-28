package binding

import (
	"strings"
	"sync"
)

// Scope is a chainable data context that supports inheritance.
//
// Scope implements the Context interface with a linked-list structure,
// allowing child scopes to inherit data from parent scopes while
// maintaining their own local data.
//
// Example:
//
//	root := NewRootScope(map[string]interface{}{
//	    "app": map[string]interface{}{
//	        "name": "MyApp",
//	        "version": "1.0",
//	    },
//	})
//
//	// Create a child scope with additional data
//	userScope := root.New(map[string]interface{}{
//	    "user": map[string]interface{}{
//	        "name": "Alice",
//	        "email": "alice@example.com",
//	    },
//	})
//
//	// Child scope can access both user and app data
//	name := userScope.Get("user.name")  // "Alice"
// appName := userScope.Get("app.name") // "MyApp"
type Scope struct {
	mu     sync.RWMutex
	data   map[string]interface{}
	parent *Scope
}

// NewRootScope creates a new root scope with the given data.
func NewRootScope(data map[string]interface{}) *Scope {
	if data == nil {
		data = make(map[string]interface{})
	}
	return &Scope{
		data:   data,
		parent: nil,
	}
}

// New creates a new child scope with additional data.
func (s *Scope) New(data map[string]interface{}) Context {
	if data == nil {
		data = make(map[string]interface{})
	}
	return &Scope{
		data:   data,
		parent: s,
	}
}

// NewScope creates a new child scope (convenience method).
func NewScope(parent *Scope, data map[string]interface{}) *Scope {
	if data == nil {
		data = make(map[string]interface{})
	}
	return &Scope{
		data:   data,
		parent: parent,
	}
}

// Get retrieves a value by path.
//
// Path can be:
//   - A simple key: "name"
//   - A nested path: "user.address.city"
//   - Special variables: "$index", "$item", "$parent"
//
// The search starts in the current scope and traverses up the parent chain.
func (s *Scope) Get(path string) (interface{}, bool) {
	// Handle special variables
	if strings.HasPrefix(path, "$") {
		return s.getSpecial(path)
	}

	// Split path into parts
	parts := splitPath(path)
	if len(parts) == 0 {
		return nil, false
	}

	// Try to get from current scope first
	if val, ok := s.getFromScope(s, parts); ok {
		return val, true
	}

	// If not found and we have a parent, try parent
	if s.parent != nil {
		// For nested paths, the root key should be found in parent
		return s.parent.Get(path)
	}

	return nil, false
}

// getFromScope attempts to get a value from a specific scope using path parts.
func (s *Scope) getFromScope(scope *Scope, parts []string) (interface{}, bool) {
	if len(parts) == 0 {
		return nil, false
	}

	scope.mu.RLock()
	defer scope.mu.RUnlock()

	// Get first part from current scope data
	val, ok := scope.data[parts[0]]
	if !ok {
		return nil, false
	}

	// Navigate through nested structure
	current := val
	for i := 1; i < len(parts); i++ {
		current = navigateTo(current, parts[i])
		if current == nil {
			return nil, false
		}
	}

	return current, true
}

// getSpecial handles special variables like $index, $item, $parent.
func (s *Scope) getSpecial(path string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	switch path {
	case "$index":
		return s.data["$index"], true
	case "$item":
		return s.data["$item"], true
	case "$parent":
		if s.parent != nil {
			return s.parent, true
		}
		return nil, false
	case "$root":
		return s.Root(), true
	default:
		// Check if it's a nested path starting with $
		if strings.HasPrefix(path, "$parent.") {
			if s.parent != nil {
				return s.parent.Get(strings.TrimPrefix(path, "$parent."))
			}
			return nil, false
		}
		// Check in data for keys starting with $
		if val, ok := s.data[path]; ok {
			return val, true
		}
		return nil, false
	}
}

// Set sets a value at the given path in the current scope.
func (s *Scope) Set(path string, value interface{}) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	parts := splitPath(path)
	if len(parts) == 0 {
		return false
	}

	if len(parts) == 1 {
		s.data[parts[0]] = value
		return true
	}

	// Navigate to parent of target, creating intermediate maps as needed
	current := s.data
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		next, ok := current[part]
		if !ok {
			// Create intermediate map
			if i == len(parts)-2 {
				// Last level before target
				current[part] = make(map[string]interface{})
			} else {
				current[part] = make(map[string]interface{})
			}
			next = current[part]
		}
		if m, ok := next.(map[string]interface{}); ok {
			current = m
		} else {
			// Path exists but is not a map
			return false
		}
	}

	// Set the final value
	current[parts[len(parts)-1]] = value
	return true
}

// Has checks if a value exists at the given path.
func (s *Scope) Has(path string) bool {
	_, ok := s.Get(path)
	return ok
}

// Parent returns the parent scope.
func (s *Scope) Parent() Context {
	if s.parent != nil {
		return s.parent
	}
	return nil
}

// Root returns the root scope of the chain.
func (s *Scope) Root() Context {
	current := s
	for current.parent != nil {
		current = current.parent
	}
	return current
}

// Data returns a copy of the scope's data.
func (s *Scope) Data() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]interface{}, len(s.data))
	for k, v := range s.data {
		result[k] = v
	}
	return result
}

// Merge merges data into the current scope.
func (s *Scope) Merge(data map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range data {
		s.data[k] = v
	}
}

// Clear clears all data in the current scope.
func (s *Scope) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]interface{})
}

// Keys returns all keys in the current scope.
func (s *Scope) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

// Len returns the number of items in the current scope.
func (s *Scope) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// IsRoot returns true if this is a root scope (no parent).
func (s *Scope) IsRoot() bool {
	return s.parent == nil
}

// Path returns the full path from root to this scope.
func (s *Scope) Path() []string {
	if s.parent == nil {
		return []string{}
	}
	return append(s.parent.Path(), "<scope>")
}

// Clone creates a shallow copy of the scope.
func (s *Scope) Clone() *Scope {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dataCopy := make(map[string]interface{}, len(s.data))
	for k, v := range s.data {
		dataCopy[k] = v
	}

	return &Scope{
		data:   dataCopy,
		parent: s.parent,
	}
}

// splitPath splits a dot-notation path into parts.
func splitPath(path string) []string {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}

	// Handle bracket notation: user["address"]["city"]
	// For now, just support dot notation
	parts := strings.Split(path, ".")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}

// navigateTo navigates through a nested structure to find a key.
func navigateTo(current interface{}, key string) interface{} {
	switch val := current.(type) {
	case map[string]interface{}:
		return val[key]
	case map[interface{}]interface{}:
		return val[key]
	default:
		return nil
	}
}

// ScopeStack manages a stack of scopes.
type ScopeStack struct {
	mu     sync.RWMutex
	scopes []*Scope
}

// NewScopeStack creates a new scope stack.
func NewScopeStack() *ScopeStack {
	return &ScopeStack{
		scopes: make([]*Scope, 0, 10),
	}
}

// Push pushes a new scope onto the stack.
func (s *ScopeStack) Push(scope *Scope) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scopes = append(s.scopes, scope)
}

// Pop pops the top scope from the stack.
func (s *ScopeStack) Pop() *Scope {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.scopes) == 0 {
		return nil
	}

	last := len(s.scopes) - 1
	scope := s.scopes[last]
	s.scopes = s.scopes[:last]
	return scope
}

// Peek returns the top scope without removing it.
func (s *ScopeStack) Peek() *Scope {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.scopes) == 0 {
		return nil
	}
	return s.scopes[len(s.scopes)-1]
}

// Depth returns the current stack depth.
func (s *ScopeStack) Depth() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.scopes)
}

// ListContext creates a list of child scopes, one for each item.
//
// This is useful for rendering list components where each row needs
// its own scope with item-specific data.
func ListContext(parent Context, items []interface{}) []Context {
	result := make([]Context, len(items))

	for i, item := range items {
		itemData := map[string]interface{}{
			"$index": i,
			"$item":  item,
		}

		// Flatten item properties if it's a map
		if m, ok := item.(map[string]interface{}); ok {
			for k, v := range m {
				itemData[k] = v
			}
		}

		result[i] = parent.New(itemData)
	}

	return result
}

// WithIndex creates a new scope with index information.
func WithIndex(parent Context, index int, item interface{}) Context {
	return parent.New(map[string]interface{}{
		"$index": index,
		"$item":  item,
	})
}
