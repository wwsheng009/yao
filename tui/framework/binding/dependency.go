package binding

import (
	"sync"

	"github.com/yaoapp/yao/tui/runtime/priority"
)

// DependencyGraph tracks state → component dependencies
type DependencyGraph struct {
	mu sync.RWMutex
	// deps maps state key → list of dependent node IDs
	deps map[string][]string
	// reverse maps node ID → list of state keys it depends on
	reverse map[string][]string
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		deps:    make(map[string][]string),
		reverse: make(map[string][]string),
	}
}

// Register registers a dependency: nodeID depends on stateKey
func (g *DependencyGraph) Register(nodeID, stateKey string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Check if already registered
	for _, id := range g.deps[stateKey] {
		if id == nodeID {
			return
		}
	}

	// Add to deps map
	g.deps[stateKey] = append(g.deps[stateKey], nodeID)

	// Add to reverse map
	g.reverse[nodeID] = append(g.reverse[nodeID], stateKey)
}

// Unregister removes all dependencies for a node
func (g *DependencyGraph) Unregister(nodeID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Get all keys this node depends on
	keys, ok := g.reverse[nodeID]
	if !ok {
		return
	}

	// Remove from deps map
	for _, key := range keys {
		nodes := g.deps[key]
		for i, id := range nodes {
			if id == nodeID {
				g.deps[key] = append(nodes[:i], nodes[i+1:]...)
				break
			}
		}
	}

	// Remove from reverse map
	delete(g.reverse, nodeID)
}

// GetDependents returns all node IDs that depend on the given state key
func (g *DependencyGraph) GetDependents(stateKey string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	deps, ok := g.deps[stateKey]
	if !ok {
		return nil
	}

	result := make([]string, len(deps))
	copy(result, deps)
	return result
}

// GetDependencies returns all state keys that a node depends on
func (g *DependencyGraph) GetDependencies(nodeID string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	deps, ok := g.reverse[nodeID]
	if !ok {
		return nil
	}

	result := make([]string, len(deps))
	copy(result, deps)
	return result
}

// Clear clears all dependencies
func (g *DependencyGraph) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.deps = make(map[string][]string)
	g.reverse = make(map[string][]string)
}

// Size returns the number of registered dependencies
func (g *DependencyGraph) Size() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	count := 0
	for _, nodes := range g.deps {
		count += len(nodes)
	}
	return count
}

// DirtyCallback is called when a dependent node should be marked dirty
// nodeID is the ID of the node to mark dirty
// zone is the logical zone where the state change originated
type DirtyCallback func(nodeID string, zone priority.StateZone)
