package core

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestContextManager_Context(t *testing.T) {
	m := NewContextManager(context.Background())

	ctx := m.Context()
	if ctx == nil {
		t.Fatal("context should not be nil")
	}
}

func TestContextManager_WithValue(t *testing.T) {
	m := NewContextManager(context.Background())

	m.WithValue(KeyRequestID, "test-123")
	m.WithValue(KeyUser, "alice")

	if m.String(KeyRequestID) != "test-123" {
		t.Errorf("expected request ID 'test-123', got '%s'", m.String(KeyRequestID))
	}

	if m.String(KeyUser) != "alice" {
		t.Errorf("expected user 'alice', got '%s'", m.String(KeyUser))
	}
}

func TestContextManager_WithContext(t *testing.T) {
	m := NewContextManager(context.Background())

	oldCtx := m.Context()

	// 设置新上下文
	newCtx, cancel := context.WithCancel(context.Background())
	m.WithContext(newCtx)

	// 上下文应该改变
	if m.Context() == oldCtx {
		t.Error("context should have changed")
	}

	// 旧上下文应该被取消
	select {
	case <-oldCtx.Done():
		// 正确
	default:
		t.Error("old context should be canceled")
	}

	cancel()
}

func TestContextManager_Shutdown(t *testing.T) {
	m := NewContextManager(context.Background())

	// 启动一些 goroutine
	var count int32
	for i := 0; i < 5; i++ {
		m.Go(func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					atomic.AddInt32(&count, 1)
					return nil
				case <-time.After(10 * time.Millisecond):
					// 模拟工作
				}
			}
		})
	}

	// 关闭
	err := m.Shutdown(1 * time.Second)
	if err != nil {
		t.Errorf("shutdown failed: %v", err)
	}

	// 所有 goroutine 应该完成
	if count != 5 {
		t.Errorf("expected 5 goroutines to complete, got %d", count)
	}

	// 上下文应该被取消
	if m.Err() == nil {
		t.Error("context should be canceled after shutdown")
	}
}

func TestContextManager_ShutdownTimeout(t *testing.T) {
	m := NewContextManager(context.Background())

	// 启动一个永不结束的 goroutine
	m.Go(func(ctx context.Context) error {
		<-ctx.Done()
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	// 设置很短的超时
	err := m.Shutdown(10 * time.Millisecond)
	if err == nil {
		t.Error("shutdown should timeout")
	}

	if err.Error() != "shutdown timeout" {
		t.Errorf("expected timeout error, got: %v", err)
	}
}

func TestContextManager_Done(t *testing.T) {
	m := NewContextManager(context.Background())

	done := make(chan struct{})
	go func() {
		<-m.Done()
		close(done)
	}()

	m.Shutdown()

	select {
	case <-done:
		// 正确
	case <-time.After(100 * time.Millisecond):
		t.Error("Done channel should be closed")
	}
}

func TestContextManager_WithTimeout(t *testing.T) {
	m := NewContextManager(context.Background())

	ctx, cancel := m.WithTimeout(10 * time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
	case <-time.After(100 * time.Millisecond):
		t.Error("context should timeout")
	}
}

func TestContextManager_Go(t *testing.T) {
	m := NewContextManager(context.Background())

	var ran bool
	m.Go(func(ctx context.Context) error {
		ran = true
		return nil
	})

	// 等待 goroutine 完成
	time.Sleep(50 * time.Millisecond)

	if !ran {
		t.Error("goroutine should have run")
	}

	// 关闭以确保 goroutine 正确退出
	m.Shutdown()
}

func TestContextManager_Go_Cancel(t *testing.T) {
	m := NewContextManager(context.Background())

	var canceled bool
	m.Go(func(ctx context.Context) error {
		<-ctx.Done()
		canceled = true
		return nil
	})

	// 取消上下文
	m.cancel()

	// 等待 goroutine 完成
	m.Shutdown()

	if !canceled {
		t.Error("goroutine should have been canceled")
	}
}

func TestContextManager_AddWaiter(t *testing.T) {
	m := NewContextManager(context.Background())

	// AddWaiter 和 DoneWaiter 的正确使用方式
	m.AddWaiter()
	done := make(chan struct{})
	go func() {
		time.Sleep(10 * time.Millisecond)
		m.DoneWaiter()
		close(done)
	}()

	select {
	case <-done:
		// 正确
	case <-time.After(100 * time.Millisecond):
		t.Error("waiter should have completed")
	}

	// Shutdown 应该立即返回（没有其他 goroutine）
	if err := m.Shutdown(10 * time.Millisecond); err != nil {
		t.Errorf("shutdown failed: %v", err)
	}
}

func TestContextValueKey(t *testing.T) {
	tests := []struct {
		name  string
		key   ContextKey
		value interface{}
	}{
		{"app", KeyApp, "my-app"},
		{"user", KeyUser, "user123"},
		{"request_id", KeyRequestID, "req-456"},
		{"session_id", KeySessionID, "sess-789"},
	}

	m := NewContextManager(context.Background())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.WithValue(tt.key, tt.value)
			if m.Value(tt.key) != tt.value {
				t.Errorf("expected %v, got %v", tt.value, m.Value(tt.key))
			}
		})
	}
}

func TestContextManager_TypeHelpers(t *testing.T) {
	m := NewContextManager(context.Background())

	m.WithValue(KeyRequestID, "test")
	m.WithValue(KeyUser, "alice")
	m.WithValue("int_key", 42)
	m.WithValue("bool_key", true)

	// String
	if m.String(KeyRequestID) != "test" {
		t.Errorf("expected 'test', got '%s'", m.String(KeyRequestID))
	}

	// Int
	if m.Int("int_key") != 42 {
		t.Errorf("expected 42, got %d", m.Int("int_key"))
	}

	// Bool
	if !m.Bool("bool_key") {
		t.Error("expected true")
	}

	// Missing key returns default values
	if m.String("missing") != "" {
		t.Error("missing string should be empty")
	}
	if m.Int("missing") != 0 {
		t.Error("missing int should be 0")
	}
	if m.Bool("missing") != false {
		t.Error("missing bool should be false")
	}
}
