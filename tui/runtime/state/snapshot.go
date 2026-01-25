package state

import "time"

// ==============================================================================
// State Snapshot (V3)
// ==============================================================================
// 所有状态必须是：
// 1. 可枚举 - 可以完整列出所有状态
// 2. 可快照 - 可以保存和恢复状态
// 3. 可追溯 - 每个状态变化都能追溯到 Action

// Snapshot 状态快照
// 这是在某个时间点的完整 UI 状态
type Snapshot struct {
	// 时间戳
	Timestamp time.Time

	// 焦点路径
	FocusPath FocusPath

	// 组件状态
	Components map[string]ComponentState

	// Modal 栈
	Modals []ModalState

	// 脏区域
	Dirty DirtyRegion

	// 元数据
	Metadata map[string]interface{}
}

// ComponentState 组件状态
type ComponentState struct {
	// 组件标识
	ID string

	// 组件类型
	Type string

	// 静态属性（配置）
	Props map[string]interface{}

	// 动态状态（运行时）
	State map[string]interface{}

	// 布局位置
	Rect Rect

	// 可见性
	Visible bool

	// 可交互性
	Disabled bool
}

// Rect 矩形区域
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

// ModalState Modal 状态
type ModalState struct {
	ID        string
	Type      ModalType
	Focus     string
	Open      bool
	Closable  bool
}

// ModalType Modal 类型
type ModalType int

const (
	ModalDialog ModalType = iota
	ModalAlert
	ModalConfirm
	ModalMenu
)

// DirtyRegion 脏区域
type DirtyRegion struct {
	// 脏单元格列表
	Cells []CellRef

	// 脏矩形列表（优化渲染）
	Rects []Rect
}

// CellRef 单元格引用
type CellRef struct {
	X int
	Y int
}

// FocusPath 焦点路径
// 使用路径而不是 bool 或索引，支持深层嵌套和 Modal
type FocusPath []string

// NewSnapshot 创建状态快照
func NewSnapshot() *Snapshot {
	return &Snapshot{
		Timestamp:  time.Now(),
		FocusPath:  make(FocusPath, 0),
		Components: make(map[string]ComponentState),
		Modals:     make([]ModalState, 0),
		Dirty:      DirtyRegion{},
		Metadata:   make(map[string]interface{}),
	}
}

// Clone 克隆快照
func (s *Snapshot) Clone() *Snapshot {
	components := make(map[string]ComponentState)
	for id, comp := range s.Components {
		components[id] = comp.Clone()
	}

	modals := make([]ModalState, len(s.Modals))
	copy(modals, s.Modals)

	focusPath := make(FocusPath, len(s.FocusPath))
	copy(focusPath, s.FocusPath)

	return &Snapshot{
		Timestamp:  s.Timestamp,
		FocusPath:  focusPath,
		Components: components,
		Modals:     modals,
		Dirty:      s.Dirty,
		Metadata:   copyMap(s.Metadata),
	}
}

// GetComponent 获取组件状态
func (s *Snapshot) GetComponent(id string) (ComponentState, bool) {
	comp, ok := s.Components[id]
	return comp, ok
}

// SetComponent 设置组件状态
func (s *Snapshot) SetComponent(comp ComponentState) {
	s.Components[comp.ID] = comp
}

// GetFocus 获取焦点
func (s *Snapshot) GetFocus() FocusPath {
	return s.FocusPath
}

// SetFocus 设置焦点
func (s *Snapshot) SetFocus(path FocusPath) {
	s.FocusPath = path
}

// Equal 比较两个快照是否相等
func (s *Snapshot) Equal(other *Snapshot) bool {
	if s == nil && other == nil {
		return true
	}
	if s == nil || other == nil {
		return false
	}

	// 比较焦点路径
	if !s.FocusPath.Equals(other.FocusPath) {
		return false
	}

	// 比较组件数量
	if len(s.Components) != len(other.Components) {
		return false
	}

	// 比较每个组件状态
	for id, comp := range s.Components {
		otherComp, ok := other.Components[id]
		if !ok || !comp.Equal(otherComp) {
			return false
		}
	}

	// 比较 Modal
	if len(s.Modals) != len(other.Modals) {
		return false
	}
	for i, modal := range s.Modals {
		if i >= len(other.Modals) || modal != other.Modals[i] {
			return false
		}
	}

	return true
}

// Clone 克隆组件状态
func (c *ComponentState) Clone() ComponentState {
	return ComponentState{
		ID:       c.ID,
		Type:     c.Type,
		Props:    copyMap(c.Props),
		State:    copyMap(c.State),
		Rect:     c.Rect,
		Visible:  c.Visible,
		Disabled: c.Disabled,
	}
}

// copyMap 复制 map
func copyMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}

// Equals 比较组件状态
func (c *ComponentState) Equal(other ComponentState) bool {
	if c.ID != other.ID || c.Type != other.Type {
		return false
	}
	if c.Visible != other.Visible || c.Disabled != other.Disabled {
		return false
	}
	// 简化比较，实际可能需要更深入的比较
	return true
}

// FocusPath 方法

// String 返回焦点路径的字符串表示
func (f FocusPath) String() string {
	if len(f) == 0 {
		return ""
	}
	result := ""
	for i, part := range f {
		if i > 0 {
			result += "."
		}
		result += part
	}
	return result
}

// Equals 比较两个焦点路径是否相等
func (f FocusPath) Equals(other FocusPath) bool {
	if len(f) != len(other) {
		return false
	}
	for i := range f {
		if f[i] != other[i] {
			return false
		}
	}
	return true
}

// Clone 克隆焦点路径
func (f FocusPath) Clone() FocusPath {
	result := make(FocusPath, len(f))
	copy(result, f)
	return result
}

// Append 追加路径段
func (f FocusPath) Append(parts ...string) FocusPath {
	result := make(FocusPath, len(f)+len(parts))
	copy(result, f)
	copy(result[len(f):], parts)
	return result
}

// Parent 返回父路径
func (f FocusPath) Parent() FocusPath {
	if len(f) == 0 {
		return f
	}
	if len(f) == 1 {
		return FocusPath{}
	}
	return f[:len(f)-1]
}

// Current 返回当前焦点 ID
func (f FocusPath) Current() string {
	if len(f) == 0 {
		return ""
	}
	return f[len(f)-1]
}

// IsEmpty 检查是否为空
func (f FocusPath) IsEmpty() bool {
	return len(f) == 0
}

// Push 推入新的路径段
func (f *FocusPath) Push(part string) {
	*f = append(*f, part)
}

// Pop 弹出最后一个路径段
func (f *FocusPath) Pop() string {
	if len(*f) == 0 {
		return ""
	}
	last := len(*f) - 1
	part := (*f)[last]
	*f = (*f)[:last]
	return part
}
