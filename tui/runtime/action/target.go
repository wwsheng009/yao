package action

// ==============================================================================
// Target Interface (V3)
// ==============================================================================
// Target 定义了接收 Action 的目标接口
// 任何组件实现此接口即可成为 Action 的分发目标

// Target Action 目标接口
// 组件实现此接口以接收 Action
type Target interface {
	// ID 返回组件 ID
	ID() string

	// HandleAction 处理 Action
	// 返回 true 表示已处理，false 表示继续传递
	HandleAction(a *Action) bool
}

// ==============================================================================
// Target Adapter
// ==============================================================================

// TargetFunc 函数式 Target，用于快速测试或简单处理器
type TargetFunc struct {
	id      string
	handler func(*Action) bool
}

// NewTargetFunc 创建函数式 Target
func NewTargetFunc(id string, handler func(*Action) bool) *TargetFunc {
	return &TargetFunc{
		id:      id,
		handler: handler,
	}
}

// ID 返回 Target ID
func (t *TargetFunc) ID() string {
	return t.id
}

// HandleAction 处理 Action
func (t *TargetFunc) HandleAction(a *Action) bool {
	if t.handler != nil {
		return t.handler(a)
	}
	return false
}

// ==============================================================================
// Target Chain (责任链模式)
// ==============================================================================

// TargetChain 链式 Target，按顺序尝试多个 Target
type TargetChain struct {
	id      string
	targets []Target
}

// NewTargetChain 创建链式 Target
func NewTargetChain(id string, targets ...Target) *TargetChain {
	return &TargetChain{
		id:      id,
		targets: targets,
	}
}

// ID 返回链 ID
func (c *TargetChain) ID() string {
	return c.id
}

// HandleAction 按顺序尝试每个 Target
func (c *TargetChain) HandleAction(a *Action) bool {
	for _, target := range c.targets {
		if target.HandleAction(a) {
			return true
		}
	}
	return false
}

// AddTarget 添加 Target 到链
func (c *TargetChain) AddTarget(target Target) {
	c.targets = append(c.targets, target)
}

// ==============================================================================
// NoOp Target
// ==============================================================================

// NoOpTarget 空操作 Target，用于测试或占位
type NoOpTarget struct {
	id string
}

// NewNoOpTarget 创建空操作 Target
func NewNoOpTarget(id string) *NoOpTarget {
	return &NoOpTarget{id: id}
}

// ID 返回 Target ID
func (n *NoOpTarget) ID() string {
	return n.id
}

// HandleAction 什么都不做，返回 false
func (n *NoOpTarget) HandleAction(a *Action) bool {
	return false
}
