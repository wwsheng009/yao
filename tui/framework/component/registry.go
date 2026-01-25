package component

import (
	"fmt"
	"sync"
)

// Registry 组件注册表
type Registry struct {
	mu        sync.RWMutex
	factories map[string]FactoryFunc
}

// FactoryFunc 组件工厂函数
type FactoryFunc func() Component

// NewRegistry 创建组件注册表
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]FactoryFunc),
	}
}

// Register 注册组件
func (r *Registry) Register(name string, factory FactoryFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[name] = factory
}

// Unregister 注销组件
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.factories, name)
}

// Create 创建组件
func (r *Registry) Create(name string) (Component, error) {
	r.mu.RLock()
	factory, ok := r.factories[name]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown component type: %s", name)
	}

	return factory(), nil
}

// CreateOrpanic 创建组件，失败时 panic
func (r *Registry) CreateOrpanic(name string) Component {
	comp, err := r.Create(name)
	if err != nil {
		panic(err)
	}
	return comp
}

// Has 检查组件是否已注册
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	_, ok := r.factories[name]
	r.mu.RUnlock()
	return ok
}

// List 列出所有已注册的组件
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// 全局默认注册表 (V2 兼容层)
var defaultRegistry = NewRegistry()

// RegisterV2 注册组件到默认注册表 (V2 兼容函数)
func RegisterV2(name string, factory FactoryFunc) {
	defaultRegistry.Register(name, factory)
}

// CreateV2 从默认注册表创建组件 (V2 兼容函数)
func CreateV2(name string) (Component, error) {
	return defaultRegistry.Create(name)
}
