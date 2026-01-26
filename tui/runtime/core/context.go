package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ==============================================================================
// Context Support (V3)
// ==============================================================================
// 添加上下文支持，实现优雅关闭和超时控制

// ContextKey 上下文键类型
type ContextKey string

const (
	// KeyApp 应用实例
	KeyApp ContextKey = "app"
	// KeyUser 用户信息
	KeyUser ContextKey = "user"
	// KeyRequestID 请求追踪 ID
	KeyRequestID ContextKey = "request_id"
	// KeySessionID 会话 ID
	KeySessionID ContextKey = "session_id"
)

// ContextAware 支持上下文的接口
type ContextAware interface {
	// Context 返回当前上下文
	Context() context.Context

	// WithContext 设置上下文
	WithContext(ctx context.Context)
}

// ContextManager 上下文管理器
type ContextManager struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewContextManager 创建上下文管理器
func NewContextManager(parent context.Context) *ContextManager {
	ctx, cancel := context.WithCancel(parent)
	return &ContextManager{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Context 返回当前上下文
func (m *ContextManager) Context() context.Context {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ctx
}

// WithContext 设置父上下文
func (m *ContextManager) WithContext(parent context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 取消旧的上下文
	m.cancel()

	// 创建新的上下文
	m.ctx, m.cancel = context.WithCancel(parent)
}

// Shutdown 优雅关闭，等待所有操作完成
func (m *ContextManager) Shutdown(timeout ...time.Duration) error {
	m.mu.Lock()
	m.cancel()
	m.mu.Unlock()

	if len(timeout) > 0 {
		done := make(chan struct{})
		go func() {
			m.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			return nil
		case <-time.After(timeout[0]):
			return fmt.Errorf("shutdown timeout")
		}
	} else {
		m.wg.Wait()
	}

	return nil
}

// Go 在上下文中启动 goroutine
// 上下文取消时，goroutine 会自动退出
func (m *ContextManager) Go(f func(context.Context) error) {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := f(m.ctx); err != nil && m.ctx.Err() == nil {
			// 只有非取消错误才记录
			fmt.Printf("goroutine error: %v\n", err)
		}
	}()
}

// Done 返回取消通道
func (m *ContextManager) Done() <-chan struct{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ctx.Done()
}

// Err 返回取消原因
func (m *ContextManager) Err() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ctx.Err()
}

// WithValue 设置上下文值
func (m *ContextManager) WithValue(key ContextKey, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ctx = context.WithValue(m.ctx, key, value)
}

// Value 获取上下文值
func (m *ContextManager) Value(key ContextKey) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ctx.Value(key)
}

// WithTimeout 设置超时上下文
func (m *ContextManager) WithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	m.mu.RLock()
	ctx := m.ctx
	m.mu.RUnlock()
	return context.WithTimeout(ctx, timeout)
}

// WithDeadline 设置截止时间上下文
func (m *ContextManager) WithDeadline(deadline time.Time) (context.Context, context.CancelFunc) {
	m.mu.RLock()
	ctx := m.ctx
	m.mu.RUnlock()
	return context.WithDeadline(ctx, deadline)
}

// AddWaiter 增加等待计数（手动管理）
func (m *ContextManager) AddWaiter() {
	m.wg.Add(1)
}

// DoneWaiter 减少等待计数（手动管理）
func (m *ContextManager) DoneWaiter() {
	m.wg.Done()
}

// Extend 添加扩展上下文的便捷方法

// String 获取字符串值
func (m *ContextManager) String(key ContextKey) string {
	if v, ok := m.Value(key).(string); ok {
		return v
	}
	return ""
}

// Int 获取整数值
func (m *ContextManager) Int(key ContextKey) int {
	if v, ok := m.Value(key).(int); ok {
		return v
	}
	return 0
}

// Bool 获取布尔值
func (m *ContextManager) Bool(key ContextKey) bool {
	if v, ok := m.Value(key).(bool); ok {
		return v
	}
	return false
}
