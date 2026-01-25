package component

import (
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// ==============================================================================
// Capability Interfaces (V3)
// ==============================================================================
// 组件通过组合这些能力接口来声明其功能，而不是实现一个大的 Component 接口
// 这是 React / Flutter / Jetpack Compose 都采用的成熟模式

// Node 基础节点接口 - 所有组件的最小接口
type Node interface {
	// ID 返回组件唯一标识符
	ID() string

	// Type 返回组件类型
	Type() string
}

// Mountable 可挂载接口 - 组件可以挂载到容器
type Mountable interface {
	Node

	// Mount 挂载到父容器
	Mount(parent Container)

	// Unmount 从父容器卸载
	Unmount()

	// IsMounted 检查是否已挂载
	IsMounted() bool
}

// Measurable 可测量接口 - 组件可以报告其理想尺寸
type Measurable interface {
	Node

	// Measure 根据约束计算理想尺寸
	// maxWidth/maxHeight 是父容器提供的最大约束
	// 返回组件期望的理想尺寸
	Measure(maxWidth, maxHeight int) (width, height int)

	// GetSize 获取当前分配的尺寸
	GetSize() (width, height int)
}

// Paintable 可绘制接口 (V3: 不返回 string)
// 组件直接绘制到 CellBuffer，而不是返回渲染字符串
type Paintable interface {
	Node

	// Paint 将组件绘制到缓冲区
	// ctx 包含绘制上下文（位置、偏移、裁剪区域等）
	// buf 是虚拟画布，组件在此绘制其内容
	Paint(ctx PaintContext, buf *paint.Buffer)
}

// PaintContext 绘制上下文
type PaintContext struct {
	// 可用尺寸
	AvailableWidth  int
	AvailableHeight int

	// 组件位置 (相对于父组件)
	X int
	Y int

	// 滚动偏移
	OffsetX int
	OffsetY int

	// Z-index 层级
	ZIndex int

	// 裁剪区域
	ClipRect *Rect
}

// ActionTarget 可处理 Action 的接口 (V3: 不处理 KeyEvent)
// 组件只处理语义化的 Action，不处理原始按键
type ActionTarget interface {
	Node

	// HandleAction 处理语义化 Action
	// 返回 true 表示已处理，false 表示继续传递
	HandleAction(a Action) bool
}

// Action 语义化 Action (将在 runtime/action 包中完整定义)
type Action interface {
	// Type 返回 Action 类型
	Type() ActionType

	// Payload 返回 Action 携带的数据
	Payload() interface{}

	// Source 返回 Action 来源
	Source() string

	// Target 返回 Action 目标
	Target() string
}

// ActionType Action 类型
type ActionType string

// Focusable 可聚焦接口 (V3: 返回 FocusID)
type Focusable interface {
	Node

	// FocusID 返回焦点标识符
	// 用于焦点系统定位和管理
	FocusID() string

	// OnFocus 获得焦点时调用
	OnFocus()

	// OnBlur 失去焦点时调用
	OnBlur()
}

// Scrollable 可滚动接口
type Scrollable interface {
	Node

	// ScrollTo 滚动到指定位置
	ScrollTo(x, y int)

	// ScrollBy 相对滚动
	ScrollBy(dx, dy int)

	// GetScrollPosition 获取当前滚动位置
	GetScrollPosition() (x, y int)
}

// Validatable 可验证接口
type Validatable interface {
	Node

	// Validate 验证组件状态
	Validate() error

	// IsValid 检查是否有效
	IsValid() bool
}

// Container 容器接口 - 定义在 container.go 中

// ==============================================================================
// 组合接口 - 常用能力组合
// ==============================================================================
// 这些接口定义了组件的能力组合
// 注意：BaseComponent 结构体在 base.go 中定义

// ComponentNode 基础组件节点组合
// 大多数静态组件（如 Text）应该实现这个接口
type ComponentNode interface {
	Node
	Mountable
	Measurable
	Paintable
}

// InteractiveComponent 交互组件组合
// 可交互组件（如 Button、Input）应该实现这个接口
type InteractiveComponent interface {
	ComponentNode
	ActionTarget
	Focusable
}

// ValidatableComponent 可验证组件组合
// 需要验证的组件（如 Form、Input）应该实现这个接口
type ValidatableComponent interface {
	ComponentNode
	Validatable
}

// ContainerComponent 容器组件组合
// 容器组件（如 Box、Flex）应该实现这个接口
type ContainerComponent interface {
	Container
	Measurable
	Paintable
}

// ============================================================================
// Common Types (V3)
// ============================================================================

// TextAlign 文本对齐
type TextAlign int

const (
	AlignLeft TextAlign = iota
	AlignCenter
	AlignRight
)
