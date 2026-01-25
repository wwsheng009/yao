package component

import (
	"fmt"
	"sync"
)

// ==============================================================================
// Component Factory (V3)
// ==============================================================================
// 组件工厂，支持从 DSL/Spec 创建组件

// Factory 组件工厂
type Factory struct {
	mu       sync.RWMutex
	creators map[string]Creator
}

// Creator 组件创建器函数类型
type Creator func(spec *Spec) (Node, error)

// Spec 组件规范
// 用于 DSL 声明式定义组件
type Spec struct {
	// Type 组件类型
	Type string

	// ID 组件 ID
	ID string

	// Props 静态属性
	Props map[string]interface{}

	// Children 子组件规范
	Children []*Spec

	// Actions 事件处理
	Actions map[string]interface{}
}

// NewFactory 创建组件工厂
func NewFactory() *Factory {
	return &Factory{
		creators: make(map[string]Creator),
	}
}

// Register 注册组件创建器
func (f *Factory) Register(typ string, creator Creator) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.creators[typ] = creator
}

// Unregister 注销组件创建器
func (f *Factory) Unregister(typ string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.creators, typ)
}

// CreateFromSpec 从规范创建组件
func (f *Factory) CreateFromSpec(spec *Spec) (Node, error) {
	if spec == nil {
		return nil, fmt.Errorf("spec is nil")
	}

	f.mu.RLock()
	creator, ok := f.creators[spec.Type]
	f.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown component type: %s", spec.Type)
	}

	node, err := creator(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to create component %s: %w", spec.Type, err)
	}

	// 设置 ID
	if spec.ID != "" {
		if setIdtable, ok := node.(interface{ SetID(string) }); ok {
			setIdtable.SetID(spec.ID)
		}
	}

	// 处理子组件
	if len(spec.Children) > 0 {
		if container, ok := node.(Container); ok {
			for _, childSpec := range spec.Children {
				childNode, err := f.CreateFromSpec(childSpec)
				if err != nil {
					return nil, err
				}
				// childNode 已经是 Node 类型，直接添加
				container.Add(childNode)
			}
		}
	}

	return node, nil
}

// Create 从类型和属性创建组件
func (f *Factory) Create(typ string, props map[string]interface{}) (Node, error) {
	return f.CreateFromSpec(&Spec{
		Type:  typ,
		Props: props,
	})
}

// HasType 检查类型是否已注册
func (f *Factory) HasType(typ string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, ok := f.creators[typ]
	return ok
}

// GetTypes 获取所有已注册的类型
func (f *Factory) GetTypes() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	types := make([]string, 0, len(f.creators))
	for typ := range f.creators {
		types = append(types, typ)
	}
	return types
}

// Count 返回已注册类型数量
func (f *Factory) Count() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.creators)
}

// Clear 清空所有注册
func (f *Factory) Clear() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.creators = make(map[string]Creator)
}

// ============================================================================
// 全局工厂实例
// ============================================================================

var globalFactory = NewFactory()

// Register 注册组件到全局工厂
func Register(typ string, creator Creator) {
	globalFactory.Register(typ, creator)
}

// Unregister 从全局工厂注销组件
func Unregister(typ string) {
	globalFactory.Unregister(typ)
}

// CreateFromSpec 从规范创建组件（使用全局工厂）
func CreateFromSpec(spec *Spec) (Node, error) {
	return globalFactory.CreateFromSpec(spec)
}

// Create 从类型和属性创建组件（使用全局工厂）
func Create(typ string, props map[string]interface{}) (Node, error) {
	return globalFactory.Create(typ, props)
}

// HasType 检查类型是否已注册（使用全局工厂）
func HasType(typ string) bool {
	return globalFactory.HasType(typ)
}

// GetFactory 获取全局工厂
func GetFactory() *Factory {
	return globalFactory
}

// ============================================================================
// Spec 辅助方法
// ============================================================================

// NewSpec 创建组件规范
func NewSpec(typ string) *Spec {
	return &Spec{
		Type:     typ,
		Props:    make(map[string]interface{}),
		Children: make([]*Spec, 0),
		Actions:  make(map[string]interface{}),
	}
}

// WithID 设置 ID
func (s *Spec) WithID(id string) *Spec {
	s.ID = id
	return s
}

// WithProp 设置属性
func (s *Spec) WithProp(key string, value interface{}) *Spec {
	if s.Props == nil {
		s.Props = make(map[string]interface{})
	}
	s.Props[key] = value
	return s
}

// WithProps 批量设置属性
func (s *Spec) WithProps(props map[string]interface{}) *Spec {
	if s.Props == nil {
		s.Props = make(map[string]interface{})
	}
	for k, v := range props {
		s.Props[k] = v
	}
	return s
}

// WithChildren 设置子组件
func (s *Spec) WithChildren(children ...*Spec) *Spec {
	s.Children = append(s.Children, children...)
	return s
}

// WithChild 添加子组件
func (s *Spec) WithChild(child *Spec) *Spec {
	s.Children = append(s.Children, child)
	return s
}

// WithAction 设置动作处理
func (s *Spec) WithAction(name string, handler interface{}) *Spec {
	if s.Actions == nil {
		s.Actions = make(map[string]interface{})
	}
	s.Actions[name] = handler
	return s
}

// GetProp 获取属性
func (s *Spec) GetProp(key string) (interface{}, bool) {
	if s.Props == nil {
		return nil, false
	}
	v, ok := s.Props[key]
	return v, ok
}

// GetPropString 获取字符串属性
func (s *Spec) GetPropString(key string, defaultValue string) string {
	if v, ok := s.GetProp(key); ok {
		if str, ok := v.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetPropInt 获取整数属性
func (s *Spec) GetPropInt(key string, defaultValue int) int {
	if v, ok := s.GetProp(key); ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return defaultValue
}

// GetPropBool 获取布尔属性
func (s *Spec) GetPropBool(key string, defaultValue bool) bool {
	if v, ok := s.GetProp(key); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// Validate 验证规范
func (s *Spec) Validate() error {
	if s.Type == "" {
		return fmt.Errorf("spec type cannot be empty")
	}
	return nil
}

// Clone 克隆规范
func (s *Spec) Clone() *Spec {
	cloned := &Spec{
		Type:     s.Type,
		ID:       s.ID,
		Props:    make(map[string]interface{}),
		Children: make([]*Spec, 0, len(s.Children)),
		Actions:  make(map[string]interface{}),
	}

	// 克隆属性
	for k, v := range s.Props {
		cloned.Props[k] = v
	}

	// 克隆子组件
	for _, child := range s.Children {
		cloned.Children = append(cloned.Children, child.Clone())
	}

	// 克隆动作
	for k, v := range s.Actions {
		cloned.Actions[k] = v
	}

	return cloned
}
