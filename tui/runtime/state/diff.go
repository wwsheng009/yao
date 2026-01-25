package state

import "reflect"

// ==============================================================================
// State Diff (V3)
// ==============================================================================

// Diff 状态差异
type Diff struct {
	// 变化的组件
	ChangedComponents []string

	// 变化的字段
	ChangedFields map[string][]string

	// 焦点变化
	FocusChanged bool

	// 新增/删除的组件
	AddedComponents   []string
	RemovedComponents []string
}

// ComputeDiff 计算两个快照的差异
func ComputeDiff(before, after *Snapshot) *Diff {
	diff := &Diff{
		ChangedComponents: make([]string, 0),
		ChangedFields:     make(map[string][]string),
	AddedComponents:   make([]string, 0),
		RemovedComponents: make([]string, 0),
	}

	// 检查焦点变化
	diff.FocusChanged = !before.FocusPath.Equals(after.FocusPath)

	// 检查新增组件
	for id := range after.Components {
		if _, ok := before.Components[id]; !ok {
			diff.AddedComponents = append(diff.AddedComponents, id)
		}
	}

	// 检查删除组件
	for id := range before.Components {
		if _, ok := after.Components[id]; !ok {
			diff.RemovedComponents = append(diff.RemovedComponents, id)
		}
	}

	// 检查组件状态变化
	for id, afterComp := range after.Components {
		beforeComp, ok := before.Components[id]
		if !ok {
			continue
		}

		// 比较状态字段
		changed := compareState(beforeComp.State, afterComp.State)
		if len(changed) > 0 {
			diff.ChangedComponents = append(diff.ChangedComponents, id)
			diff.ChangedFields[id] = changed
		}
	}

	return diff
}

// compareState 比较状态字段
func compareState(before, after map[string]interface{}) []string {
	changed := make([]string, 0)

	// 检查新增或修改的字段
	for key, afterVal := range after {
		beforeVal, ok := before[key]
		if !ok || !reflect.DeepEqual(beforeVal, afterVal) {
			changed = append(changed, key)
		}
	}

	// 检查删除的字段
	for key := range before {
		if _, ok := after[key]; !ok {
			changed = append(changed, key)
		}
	}

	return changed
}

// String 返回差异的字符串表示
func (d *Diff) String() string {
	result := "StateDiff{"

	if d.FocusChanged {
		result += " FocusChanged"
	}

	if len(d.ChangedComponents) > 0 {
		result += " Changed:" + stringSliceToString(d.ChangedComponents)
	}

	if len(d.AddedComponents) > 0 {
		result += " Added:" + stringSliceToString(d.AddedComponents)
	}

	if len(d.RemovedComponents) > 0 {
		result += " Removed:" + stringSliceToString(d.RemovedComponents)
	}

	result += " }"
	return result
}

// stringSliceToString 字符串切片转字符串
func stringSliceToString(arr []string) string {
	if len(arr) == 0 {
		return "[]"
	}
	result := "["
	for i, s := range arr {
		if i > 0 {
			result += ","
		}
		result += `"` + s + `"`
	}
	result += "]"
	return result
}

// HasChanges 检查是否有任何变化
func (d *Diff) HasChanges() bool {
	return d.FocusChanged ||
		len(d.ChangedComponents) > 0 ||
		len(d.AddedComponents) > 0 ||
		len(d.RemovedComponents) > 0
}

// GetComponentChanges 获取特定组件的变化字段
func (d *Diff) GetComponentChanges(id string) []string {
	return d.ChangedFields[id]
}

// IsComponentChanged 检查组件是否变化
func (d *Diff) IsComponentChanged(id string) bool {
	for _, cid := range d.ChangedComponents {
		if cid == id {
			return true
		}
	}
	return false
}

// IsComponentAdded 检查组件是否新增
func (d *Diff) IsComponentAdded(id string) bool {
	for _, cid := range d.AddedComponents {
		if cid == id {
			return true
		}
	}
	return false
}

// IsComponentRemoved 检查组件是否删除
func (d *Diff) IsComponentRemoved(id string) bool {
	for _, cid := range d.RemovedComponents {
		if cid == id {
			return true
		}
	}
	return false
}

// Merge 合并差异到基础快照
func (d *Diff) Merge(base *Snapshot) *Snapshot {
	result := base.Clone()

	// 应用新增组件
	for _, id := range d.AddedComponents {
		if comp, ok := base.Components[id]; ok {
			result.Components[id] = comp
		}
	}

	// 应用删除组件
	for _, id := range d.RemovedComponents {
		delete(result.Components, id)
	}

	// 应用状态变化
	for _, id := range d.ChangedComponents {
		if comp, ok := base.Components[id]; ok {
			result.Components[id] = comp
		}
	}

	// 应用焦点变化
	if d.FocusChanged {
		result.FocusPath = base.FocusPath.Clone()
	}

	return result
}
