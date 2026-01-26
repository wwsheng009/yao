package framework

import (
	"context"
	"testing"
	"time"

	frameworkevent "github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/runtime/core"
	"github.com/yaoapp/yao/tui/runtime/render"
)

// TestApp_Throttler 测试 App 的节流器集成
func TestApp_Throttler(t *testing.T) {
	app := NewApp()

	// 测试默认帧率
	if fps := app.throttler.FPS(); fps != 60 {
		t.Errorf("expected FPS 60, got %d", fps)
	}

	// 测试设置帧率
	app.SetFPS(30)
	if fps := app.FPS(); fps != 30 {
		t.Errorf("expected FPS 30, got %d", fps)
	}

	// 测试统计信息
	stats := app.GetRenderStats()
	if stats.TargetFPS != 30 {
		t.Errorf("expected TargetFPS 30, got %d", stats.TargetFPS)
	}
}

// TestApp_ContextManager 测试 App 的上下文管理器集成
func TestApp_ContextManager(t *testing.T) {
	app := NewApp()

	// 测试获取上下文
	ctx := app.Context()
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	// 测试上下文取消
	app.Shutdown(100 * time.Millisecond)
	select {
	case <-ctx.Done():
		// 上下文已取消，预期行为
	case <-time.After(200 * time.Millisecond):
		t.Fatal("context should be cancelled after Shutdown")
	}
}

// TestApp_Recovery 测试 App 的 panic 恢复集成
func TestApp_Recovery(t *testing.T) {
	app := NewApp()

	// 启用恢复
	app.EnableRecovery()

	if app.recovery == nil {
		t.Fatal("expected recovery manager to be initialized")
	}

	// 测试添加处理器 - 创建一个简单的测试处理器
	app.AddPanicHandler(&testPanicHandler{})
}

// TestApp_EventFilter 测试 App 的事件过滤器集成
func TestApp_EventFilter(t *testing.T) {
	app := NewApp()
	filterCalled := false

	// 设置过滤器
	app.SetEventFilter(func(ev frameworkevent.Event) bool {
		filterCalled = true
		return true
	})

	// 清除过滤器
	app.ClearEventFilter()

	// 设置一个拦截所有事件的过滤器
	app.SetEventFilter(func(ev frameworkevent.Event) bool {
		return false
	})
	_ = filterCalled // 避免未使用变量警告
}

// TestApp_ForceRender 测试强制渲染
func TestApp_ForceRender(t *testing.T) {
	app := NewApp()

	// 强制渲染应该设置 dirty 标记
	app.ForceRender()
	if !app.dirty {
		t.Error("expected dirty to be true after ForceRender")
	}
}

// TestApp_AdaptiveFPS 测试自适应帧率
func TestApp_AdaptiveFPS(t *testing.T) {
	app := NewApp()

	// 启用自适应帧率
	app.EnableAdaptiveFPS(true)

	stats := app.GetRenderStats()
	if !stats.Adaptive {
		t.Error("expected adaptive mode to be enabled")
	}

	// 禁用自适应帧率
	app.EnableAdaptiveFPS(false)

	stats = app.GetRenderStats()
	if stats.Adaptive {
		t.Error("expected adaptive mode to be disabled")
	}
}

// TestApp_GracefulShutdown 测试优雅关闭
func TestApp_GracefulShutdown(t *testing.T) {
	app := NewApp()

	// 获取上下文
	ctx := app.Context()

	// 在另一个 goroutine 中触发关闭
	go func() {
		time.Sleep(50 * time.Millisecond)
		app.Shutdown(100 * time.Millisecond)
	}()

	// 等待上下文取消
	select {
	case <-ctx.Done():
		// 预期行为
	case <-time.After(200 * time.Millisecond):
		t.Fatal("context should be cancelled after Shutdown")
	}
}

// TestThrottler_Behavior 详细测试节流器行为
func TestThrottler_Behavior(t *testing.T) {
	throttler := render.NewThrottler(60)

	// 首次调用应该返回 true
	if !throttler.ShouldRender() {
		t.Error("first render should be allowed")
	}

	// 立即再次调用应该返回 false（间隔太短）
	if throttler.ShouldRender() {
		t.Error("immediate second render should be throttled")
	}

	// 等待超过最小间隔后应该允许渲染
	time.Sleep(20 * time.Millisecond)
	if !throttler.ShouldRender() {
		t.Error("render after interval should be allowed")
	}
}

// TestContextManager_Integration 测试上下文管理器集成
func TestContextManager_Integration(t *testing.T) {
	ctx := context.Background()
	mgr := core.NewContextManager(ctx)

	// 测试上下文传播
	appCtx := mgr.Context()
	if appCtx == nil {
		t.Fatal("expected non-nil context from manager")
	}

	// 测试优雅关闭
	done := make(chan error, 1)
	go func() {
		done <- mgr.Shutdown(100 * time.Millisecond)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("shutdown failed: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("shutdown timed out")
	}
}

// testPanicHandler 测试用的 panic 处理器
type testPanicHandler struct{}

func (h *testPanicHandler) HandlePanic(r interface{}, stack []byte) {
	// 测试实现，什么都不做
}
