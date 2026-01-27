package component

// ComponentContext 组件上下文
//
// ComponentContext 在组件挂载时传递给组件，提供运行时资源访问。
// 这避免了 App 直接依赖具体组件类型，符合依赖倒置原则。
type ComponentContext struct {
	// dirtyCallback 脏标记回调函数
	dirtyCallback func()
}

// NewComponentContext 创建组件上下文
func NewComponentContext() *ComponentContext {
	return &ComponentContext{}
}

// SetDirtyCallback 设置脏标记回调
func (c *ComponentContext) SetDirtyCallback(fn func()) {
	c.dirtyCallback = fn
}

// GetDirtyCallback 获取脏标记回调
func (c *ComponentContext) GetDirtyCallback() func() {
	return c.dirtyCallback
}

// MarkDirty 触发脏标记（如果有设置回调）
func (c *ComponentContext) MarkDirty() {
	if c.dirtyCallback != nil {
		c.dirtyCallback()
	}
}

// MountableWithContext is defined in capabilities.go.

// ==============================================================================
// Helper Interfaces - For type-safe context resource access
// ==============================================================================

// DirtyNotifiable 需要脏标记通知的组件接口
type DirtyNotifiable interface {
	Node
	// SetDirtyCallback 设置脏标记回调
	SetDirtyCallback(fn func())
}
