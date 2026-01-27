package tui

import (
	"sync"
)

// modelStore stores active TUI models by ID
var modelStore = struct {
	sync.RWMutex
	models map[string]*Model
}{
	models: make(map[string]*Model),
}

// RegisterModel registers a model instance with a unique ID
func RegisterModel(id string, model *Model) {
	modelStore.Lock()
	defer modelStore.Unlock()
	modelStore.models[id] = model
}

// UnregisterModel removes a model instance by ID
func UnregisterModel(id string) {
	modelStore.Lock()
	defer modelStore.Unlock()
	delete(modelStore.models, id)
}

// GetModel retrieves a model instance by ID
func GetModel(id string) *Model {
	modelStore.RLock()
	defer modelStore.RUnlock()
	return modelStore.models[id]
}

// ListModelIDs returns all registered model IDs
func ListModelIDs() []string {
	modelStore.RLock()
	defer modelStore.RUnlock()
	
	ids := make([]string, 0, len(modelStore.models))
	for id := range modelStore.models {
		ids = append(ids, id)
	}
	return ids
}