package binding

import (
	"context"
	"sync"
	"time"

	"github.com/yaoapp/yao/tui/runtime/priority"
)

// Notifier is a function that notifies of state changes.
type Notifier func(key string, oldValue, newValue interface{})

// ReactiveStore is a reactive data store with dependency tracking.
//
// ReactiveStore allows components to subscribe to state changes and
// automatically notifies them when relevant data is modified.
//
// Example:
//
//	store := NewReactiveStore()
//	store.Set("user.name", "Alice")
//
//	// Subscribe to changes
//	unsubscribe := store.Subscribe("user.name", func(key string, old, new interface{}) {
//	    fmt.Printf("%s changed from %v to %v\n", key, old, new)
//	})
//
//	store.Set("user.name", "Bob")  // Triggers notification
//	unsubscribe()                  // Stop listening
type ReactiveStore struct {
	mu        sync.RWMutex
	data      map[string]interface{}
	observers map[string][]Notifier
	global    []Notifier
	enabled   bool
	// Batch update support
	batching     bool
	pendingKeys  map[string]interface{}
	pendingNotifs []notification

	// Dependency graph for automatic dirty marking
	depGraph      *DependencyGraph
	dirtyCallback DirtyCallback
}

// notification represents a pending notification during batching.
type notification struct {
	key       string
	oldValue  interface{}
	newValue  interface{}
}

// NewReactiveStore creates a new reactive store.
func NewReactiveStore() *ReactiveStore {
	return &ReactiveStore{
		data:      make(map[string]interface{}),
		observers: make(map[string][]Notifier),
		global:    make([]Notifier, 0),
		enabled:   true,
		depGraph:  NewDependencyGraph(),
	}
}

// Get retrieves a value by path.
func (s *ReactiveStore) Get(path string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.data[path]
	return val, ok
}

// Set sets a value at the given path.
//
// If batching is enabled, the notification is deferred until EndBatch is called.
func (s *ReactiveStore) Set(path string, value interface{}) {
	s.SetWithZone(path, value, priority.ZoneData)
}

// SetWithZone sets a value at the given path with a specific zone.
// The zone determines the priority of the update and is used to mark
// dependent nodes dirty with the appropriate priority level.
func (s *ReactiveStore) SetWithZone(path string, value interface{}, zone priority.StateZone) {
	s.mu.Lock()

	oldValue, exists := s.data[path]
	originalOld := oldValue
	s.data[path] = value

	// Get dependents BEFORE releasing lock
	dependents := s.depGraph.GetDependents(path)

	if s.batching {
		if s.pendingKeys == nil {
			s.pendingKeys = make(map[string]interface{})
		}
		s.pendingKeys[path] = value

		// For batching, track the original old value and only add notification once per key
		// Check if we already have a pending notification for this key
		hasPending := false
		for i, notif := range s.pendingNotifs {
			if notif.key == path {
				// Update existing notification
				s.pendingNotifs[i].newValue = value
				hasPending = true
				break
			}
		}
		if !hasPending && (exists || oldValue != value) {
			s.pendingNotifs = append(s.pendingNotifs, notification{
				key:      path,
				oldValue: originalOld,
				newValue: value,
			})
		}
		s.mu.Unlock()
		return
	}

	changed := !exists || oldValue != value
	s.mu.Unlock()

	if changed && s.enabled {
		// Notify observers
		s.notify(path, oldValue, value)

		// Mark dependent nodes dirty via callback
		for _, nodeID := range dependents {
			if s.dirtyCallback != nil {
				s.dirtyCallback(nodeID, zone)
			}
		}
	}
}

// Delete removes a value at the given path.
func (s *ReactiveStore) Delete(path string) {
	s.mu.Lock()

	oldValue, exists := s.data[path]
	delete(s.data, path)

	s.mu.Unlock()

	if exists && s.enabled {
		s.notify(path, oldValue, nil)
	}
}

// Has checks if a value exists at the given path.
func (s *ReactiveStore) Has(path string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[path]
	return ok
}

// GetAll returns a copy of all data in the store.
func (s *ReactiveStore) GetAll() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]interface{}, len(s.data))
	for k, v := range s.data {
		result[k] = v
	}
	return result
}

// SetAll replaces all data in the store.
func (s *ReactiveStore) SetAll(data map[string]interface{}) {
	s.mu.Lock()
	oldData := s.data
	s.data = make(map[string]interface{}, len(data))
	for k, v := range data {
		s.data[k] = v
	}
	s.mu.Unlock()

	if s.enabled {
		for k, v := range data {
			oldVal, had := oldData[k]
			if !had || oldVal != v {
				s.notify(k, oldVal, v)
			}
		}
	}
}

// Clear removes all data from the store.
func (s *ReactiveStore) Clear() {
	s.mu.Lock()
	oldData := s.data
	s.data = make(map[string]interface{})
	s.mu.Unlock()

	if s.enabled {
		for k, v := range oldData {
			s.notify(k, v, nil)
		}
	}
}

// Subscribe registers a notifier for a specific key.
//
// The returned function can be called to unsubscribe.
func (s *ReactiveStore) Subscribe(path string, notifier Notifier) func() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.observers[path] = append(s.observers[path], notifier)

	return func() {
		s.Unsubscribe(path, notifier)
	}
}

// SubscribeGlobal registers a global notifier for all changes.
func (s *ReactiveStore) SubscribeGlobal(notifier Notifier) func() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.global = append(s.global, notifier)

	return func() {
		s.UnsubscribeGlobal(notifier)
	}
}

// Unsubscribe removes a notifier for a specific key.
func (s *ReactiveStore) Unsubscribe(path string, notifier Notifier) {
	s.mu.Lock()
	defer s.mu.Unlock()

	notifiers := s.observers[path]
	for i, n := range notifiers {
		if &n == &notifier || getFunctionPointer(n) == getFunctionPointer(notifier) {
			s.observers[path] = append(notifiers[:i], notifiers[i+1:]...)
			break
		}
	}
}

// UnsubscribeGlobal removes a global notifier.
func (s *ReactiveStore) UnsubscribeGlobal(notifier Notifier) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, n := range s.global {
		if &n == &notifier || getFunctionPointer(n) == getFunctionPointer(notifier) {
			s.global = append(s.global[:i], s.global[i+1:]...)
			break
		}
	}
}

// Enable enables notifications from the store.
func (s *ReactiveStore) Enable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = true
}

// Disable disables notifications from the store.
func (s *ReactiveStore) Disable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = false
}

// IsEnabled returns whether notifications are enabled.
func (s *ReactiveStore) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// BeginBatch starts batching updates.
//
// While batching, notifications are deferred until EndBatch is called.
func (s *ReactiveStore) BeginBatch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.batching = true
	s.pendingKeys = make(map[string]interface{})
	s.pendingNotifs = make([]notification, 0)
}

// EndBatch ends batching and fires all pending notifications.
func (s *ReactiveStore) EndBatch() {
	s.mu.Lock()

	// Collect pending notifications
	pending := s.pendingNotifs
	s.batching = false
	s.pendingKeys = nil
	s.pendingNotifs = nil

	s.mu.Unlock()

	// Fire notifications outside the lock
	for _, notif := range pending {
		if s.enabled {
			s.notify(notif.key, notif.oldValue, notif.newValue)
		}
	}
}

// notify notifies observers of a change.
func (s *ReactiveStore) notify(key string, oldValue, newValue interface{}) {
	s.mu.RLock()

	// Key-specific observers
	observers := make([]Notifier, len(s.observers[key]))
	copy(observers, s.observers[key])

	// Global observers
	global := make([]Notifier, len(s.global))
	copy(global, s.global)

	s.mu.RUnlock()

	// Notify key-specific observers
	for _, observer := range observers {
		observer(key, oldValue, newValue)
	}

	// Notify global observers
	for _, observer := range global {
		observer(key, oldValue, newValue)
	}
}

// Size returns the number of items in the store.
func (s *ReactiveStore) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// Keys returns all keys in the store.
func (s *ReactiveStore) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

// ToContext creates a Context from the store's data.
func (s *ReactiveStore) ToContext() Context {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a root scope with all store data
	data := make(map[string]interface{}, len(s.data))
	for k, v := range s.data {
		data[k] = v
	}
	return NewRootScope(data)
}

// getFunctionPointer is a helper for comparing functions.
// Note: This is a simplified implementation.
func getFunctionPointer(f Notifier) uintptr {
	// In Go, you can't directly get function pointers for comparison
	// This is a placeholder that returns 0
	return 0
}

// StoreObserver wraps a component for store notifications.
type StoreObserver struct {
	store    *ReactiveStore
	key      string
	callback func()
	cancel   func()
}

// NewStoreObserver creates a new store observer.
func NewStoreObserver(store *ReactiveStore, key string, callback func()) *StoreObserver {
	cancel := store.Subscribe(key, func(k string, old, new interface{}) {
		callback()
	})
	return &StoreObserver{
		store:    store,
		key:      key,
		callback: callback,
		cancel:   cancel,
	}
}

// Dispose stops observing the store.
func (o *StoreObserver) Dispose() {
	if o.cancel != nil {
		o.cancel()
		o.cancel = nil
	}
}

// StoreComputed creates a computed value that updates when dependencies change.
type StoreComputed struct {
	store *ReactiveStore
	deps  []string
	compute func() interface{}
	value interface{}
	cancel func()
}

// NewStoreComputed creates a new computed value.
func NewStoreComputed(store *ReactiveStore, deps []string, compute func() interface{}) *StoreComputed {
	c := &StoreComputed{
		store:   store,
		deps:    deps,
		compute: compute,
		value:   compute(),
	}

	// Subscribe to all dependencies
	var cancels []func()
	for _, dep := range deps {
		cancel := store.Subscribe(dep, func(key string, old, new interface{}) {
			c.value = c.compute()
		})
		cancels = append(cancels, cancel)
	}

	c.cancel = func() {
		for _, cancel := range cancels {
			cancel()
		}
	}

	return c
}

// Value returns the current computed value.
func (c *StoreComputed) Value() interface{} {
	return c.value
}

// Dispose stops updating the computed value.
func (c *StoreComputed) Dispose() {
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
}

// Watcher watches for changes and triggers async actions.
type Watcher struct {
	store   *ReactiveStore
	key     string
	cancel  func()
	debounce time.Duration
	lastRun time.Time
	pending bool
}

// WatchOptions configures a watcher.
type WatchOptions struct {
	Debounce time.Duration
}

// NewWatcher creates a new watcher.
func NewWatcher(store *ReactiveStore, key string, handler func(key string, old, new interface{}), opts ...WatchOptions) *Watcher {
	w := &Watcher{
		store: store,
		key:   key,
	}

	if len(opts) > 0 {
		w.debounce = opts[0].Debounce
	}

	w.cancel = store.Subscribe(key, func(k string, old, new interface{}) {
		if w.debounce > 0 {
			w.scheduleDebounced(handler, k, old, new)
		} else {
			handler(k, old, new)
		}
	})

	return w
}

// scheduleDebounced schedules a debounced handler call.
func (w *Watcher) scheduleDebounced(handler func(string, interface{}, interface{}), key string, old, new interface{}) {
	w.pending = true
	go func() {
		time.Sleep(w.debounce)
		w.store.mu.Lock()
		doNotify := w.pending
		w.pending = false
		w.store.mu.Unlock()

		if doNotify {
			handler(key, old, new)
		}
	}()
}

// Dispose stops watching.
func (w *Watcher) Dispose() {
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
	}
}

// StoreContext creates a context-aware store.
type StoreContext struct {
	store *ReactiveStore
	ctx   context.Context
}

// NewStoreContext creates a new store context.
func NewStoreContext(ctx context.Context, store *ReactiveStore) *StoreContext {
	return &StoreContext{
		store: store,
		ctx:   ctx,
	}
}

// Context returns the underlying context.
func (sc *StoreContext) Context() context.Context {
	return sc.ctx
}

// Done returns when the context is cancelled.
func (sc *StoreContext) Done() <-chan struct{} {
	return sc.ctx.Done()
}

// AutoDispose automatically disposes resources when context is done.
func (sc *StoreContext) AutoDispose() {
	go func() {
		<-sc.ctx.Done()
		sc.store.Clear()
	}()
}

// =============================================================================
// Dependency Graph Integration
// =============================================================================

// SetDirtyCallback sets the callback to be invoked when dependent nodes should be marked dirty
func (s *ReactiveStore) SetDirtyCallback(cb DirtyCallback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dirtyCallback = cb
}

// RegisterDependency registers a node's dependency on a state key
// When the state at the given key changes, the node will be marked dirty
func (s *ReactiveStore) RegisterDependency(nodeID, stateKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.depGraph.Register(nodeID, stateKey)
}

// UnregisterDependencies removes all dependencies for a given node
func (s *ReactiveStore) UnregisterDependencies(nodeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.depGraph.Unregister(nodeID)
}

// GetDependents returns all node IDs that depend on the given state key
func (s *ReactiveStore) GetDependents(stateKey string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.depGraph.GetDependents(stateKey)
}

// GetDependencies returns all state keys that a node depends on
func (s *ReactiveStore) GetDependencies(nodeID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.depGraph.GetDependencies(nodeID)
}

// DependencyCount returns the total number of registered dependencies
func (s *ReactiveStore) DependencyCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.depGraph.Size()
}

