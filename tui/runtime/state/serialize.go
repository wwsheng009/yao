package state

import (
	"encoding/json"
	"os"
)

// ==============================================================================
// State Serialization (V3)
// ==============================================================================

// Serialize 序列化状态为 JSON
func (s *Snapshot) Serialize() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// Deserialize 从 JSON 反序列化状态
func Deserialize(data []byte) (*Snapshot, error) {
	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// SaveToFile 保存状态到文件
func (s *Snapshot) SaveToFile(path string) error {
	data, err := s.Serialize()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadFromFile 从文件加载状态
func LoadFromFile(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Deserialize(data)
}

// SaveToFileAtomic 原子性保存状态
// 使用临时文件 + 重命名确保原子性
func (s *Snapshot) SaveToFileAtomic(path string) error {
	data, err := s.Serialize()
	if err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	// 原子性重命名
	return os.Rename(tmpPath, path)
}

// LoadFromFileAtomic 原子性加载状态
// 读取临时文件，验证后再替换
func LoadFromFileAtomic(path string) (*Snapshot, error) {
	// 先读取临时文件
	tmpPath := path + ".tmp"
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, err
	}

	// 验证可以反序列化
	snapshot, err := Deserialize(data)
	if err != nil {
		// 删除无效的临时文件
		os.Remove(tmpPath)
		return nil, err
	}

	// 原子性重命名
	if err := os.Rename(tmpPath, path); err != nil {
		return nil, err
	}

	return snapshot, nil
}

// ExportToMap 导出为 map（用于 AI 查询）
func (s *Snapshot) ExportToMap() map[string]interface{} {
	result := make(map[string]interface{})

	// 基本信息
	result["timestamp"] = s.Timestamp
	result["focusPath"] = s.FocusPath.String()

	// 组件状态
	components := make(map[string]interface{})
	for id, comp := range s.Components {
		components[id] = map[string]interface{}{
			"type":     comp.Type,
			"props":    comp.Props,
			"state":    comp.State,
			"rect":     map[string]int{
				"x":      comp.Rect.X,
				"y":      comp.Rect.Y,
				"width":  comp.Rect.Width,
				"height": comp.Rect.Height,
			},
			"visible":  comp.Visible,
			"disabled": comp.Disabled,
		}
	}
	result["components"] = components

	// Modal 状态
	modals := make([]map[string]interface{}, len(s.Modals))
	for i, modal := range s.Modals {
		modals[i] = map[string]interface{}{
			"id":        modal.ID,
			"type":      int(modal.Type),
			"focus":     modal.Focus,
			"open":      modal.Open,
			"closable":  modal.Closable,
		}
	}
	result["modals"] = modals

	// 脏区域
	if len(s.Dirty.Cells) > 0 || len(s.Dirty.Rects) > 0 {
		dirty := make(map[string]interface{})
		if len(s.Dirty.Cells) > 0 {
			cells := make([]map[string]int, len(s.Dirty.Cells))
			for i, cell := range s.Dirty.Cells {
				cells[i] = map[string]int{"x": cell.X, "y": cell.Y}
			}
			dirty["cells"] = cells
		}
		if len(s.Dirty.Rects) > 0 {
			rects := make([]map[string]interface{}, len(s.Dirty.Rects))
			for i, rect := range s.Dirty.Rects {
				rects[i] = map[string]interface{}{
					"x":      rect.X,
					"y":      rect.Y,
					"width":  rect.Width,
					"height": rect.Height,
				}
			}
			dirty["rects"] = rects
		}
		result["dirty"] = dirty
	}

	// 元数据
	if len(s.Metadata) > 0 {
		result["metadata"] = s.Metadata
	}

	return result
}

// ImportFromMap 从 map 导入状态（用于 AI 操作）
func ImportFromMap(data map[string]interface{}) (*Snapshot, error) {
	snapshot := NewSnapshot()

	// 时间戳
	if ts, ok := data["timestamp"].(string); ok {
		// TODO: 解析时间戳
		_ = ts
	}

	// 焦点路径
	if focusPath, ok := data["focusPath"].(string); ok {
		parts := parseFocusPath(focusPath)
		snapshot.FocusPath = FocusPath(parts)
	}

	// 组件状态
	if components, ok := data["components"].(map[string]interface{}); ok {
		for id, compData := range components {
			comp, ok := compData.(map[string]interface{})
			if !ok {
				continue
			}

			componentState := ComponentState{ID: id}

			if typ, ok := comp["type"].(string); ok {
				componentState.Type = typ
			}
			if props, ok := comp["props"].(map[string]interface{}); ok {
				componentState.Props = props
			}
			if state, ok := comp["state"].(map[string]interface{}); ok {
				componentState.State = state
			}
			if rect, ok := comp["rect"].(map[string]int); ok {
				componentState.Rect = Rect{
					X:      rect["x"],
					Y:      rect["y"],
					Width:  rect["width"],
					Height: rect["height"],
				}
			}
			if visible, ok := comp["visible"].(bool); ok {
				componentState.Visible = visible
			}
			if disabled, ok := comp["disabled"].(bool); ok {
				componentState.Disabled = disabled
			}

			snapshot.Components[id] = componentState
		}
	}

	// Modal 状态
	if modals, ok := data["modals"].([]interface{}); ok {
		for _, modalData := range modals {
			modal, ok := modalData.(map[string]interface{})
			if !ok {
				continue
			}

			modalState := ModalState{}

			if id, ok := modal["id"].(string); ok {
				modalState.ID = id
			}
			if typ, ok := modal["type"].(float64); ok {
				modalState.Type = ModalType(typ)
			}
			if focus, ok := modal["focus"].(string); ok {
				modalState.Focus = focus
			}
			if open, ok := modal["open"].(bool); ok {
				modalState.Open = open
			}
			if closable, ok := modal["closable"].(bool); ok {
				modalState.Closable = closable
			}

			snapshot.Modals = append(snapshot.Modals, modalState)
		}
	}

	// 元数据
	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		snapshot.Metadata = metadata
	}

	return snapshot, nil
}

// parseFocusPath 解析焦点路径字符串
func parseFocusPath(path string) []string {
	if path == "" {
		return []string{}
	}
	// 简单按 "." 分割
	result := []string{}
	current := ""
	for _, ch := range path {
		if ch == '.' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
