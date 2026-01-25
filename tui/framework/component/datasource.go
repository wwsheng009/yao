package component

import "sync"

// ==============================================================================
// Data Source (V3)
// ==============================================================================
// 数据源接口，用于虚拟滚动组件

// DataSource 数据源接口
// 为 List、Table 等组件提供数据抽象
type DataSource interface {
	// Count 返回数据总数
	Count() int

	// Get 获取指定索引的数据
	Get(index int) interface{}
}

// SimpleDataSource 简单数据源
// 基于 slice 的内存数据源
type SimpleDataSource struct {
	mu    sync.RWMutex
	items []interface{}
}

// NewSimpleDataSource 创建简单数据源
func NewSimpleDataSource(items []interface{}) *SimpleDataSource {
	return &SimpleDataSource{
		items: items,
	}
}

// Count 返回数据总数
func (s *SimpleDataSource) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

// Get 获取指定索引的数据
func (s *SimpleDataSource) Get(index int) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if index >= 0 && index < len(s.items) {
		return s.items[index]
	}
	return nil
}

// Set 设置指定索引的数据
func (s *SimpleDataSource) Set(index int, item interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index >= 0 && index < len(s.items) {
		s.items[index] = item
	}
}

// Add 添加数据
func (s *SimpleDataSource) Add(item interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
}

// Remove 移除指定索引的数据
func (s *SimpleDataSource) Remove(index int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index >= 0 && index < len(s.items) {
		s.items = append(s.items[:index], s.items[index+1:]...)
	}
}

// Clear 清空数据
func (s *SimpleDataSource) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make([]interface{}, 0)
}

// Items 返回所有数据
func (s *SimpleDataSource) Items() []interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]interface{}, len(s.items))
	copy(result, s.items)
	return result
}

// ==============================================================================
// String Data Source
// ==============================================================================

// StringDataSource 字符串数据源
type StringDataSource struct {
	*SimpleDataSource
}

// NewStringDataSource 创建字符串数据源
func NewStringDataSource(items []string) *StringDataSource {
	wrapped := make([]interface{}, len(items))
	for i, item := range items {
		wrapped[i] = item
	}
	return &StringDataSource{
		SimpleDataSource: NewSimpleDataSource(wrapped),
	}
}

// GetString 获取字符串数据
func (s *StringDataSource) GetString(index int) string {
	item := s.Get(index)
	if str, ok := item.(string); ok {
		return str
	}
	return ""
}

// AddString 添加字符串
func (s *StringDataSource) AddString(item string) {
	s.Add(item)
}
