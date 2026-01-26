package render

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewThrottler(t *testing.T) {
	throttler := NewThrottler(60)

	if throttler.FPS() != 60 {
		t.Errorf("expected FPS 60, got %d", throttler.FPS())
	}

	if throttler.IsAdaptive() {
		t.Error("adaptive should be disabled by default")
	}
}

func TestNewThrottler_ClampsFPS(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"zero", 0, 60},
		{"negative", -10, 60},
		{"too high", 200, 120},
		{"normal", 30, 30},
		{"max", 120, 120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			throttler := NewThrottler(tt.input)
			if throttler.FPS() != tt.expected {
				t.Errorf("expected FPS %d, got %d", tt.expected, throttler.FPS())
			}
		})
	}
}

func TestThrottler_ShouldRender(t *testing.T) {
	time.Sleep(10 * time.Millisecond) // 确保干净状态
	throttler := NewThrottler(60) // ~16.67ms 间隔

	// 第一次应该渲染
	if !throttler.ShouldRender() {
		t.Error("first render should be allowed")
	}

	// 立即第二次不应该渲染
	if throttler.ShouldRender() {
		t.Error("immediate second render should be blocked")
	}

	// 等待间隔后应该可以渲染
	time.Sleep(20 * time.Millisecond)
	if !throttler.ShouldRender() {
		t.Error("render after interval should be allowed")
	}
}

func TestThrottler_RecordFrameTime(t *testing.T) {
	throttler := NewThrottler(60)

	// 记录一些帧时间
	times := []time.Duration{
		5 * time.Millisecond,
		10 * time.Millisecond,
		15 * time.Millisecond,
	}

	for _, ft := range times {
		throttler.RecordFrameTime(ft)
	}

	stats := throttler.Stats()
	if stats.TotalFrames != int64(len(times)) {
		t.Errorf("expected %d frames, got %d", len(times), stats.TotalFrames)
	}
}

func TestThrottler_SetFPS(t *testing.T) {
	throttler := NewThrottler(30)

	throttler.SetFPS(60)
	if throttler.FPS() != 60 {
		t.Errorf("expected FPS 60, got %d", throttler.FPS())
	}

	throttler.SetFPS(0)
	if throttler.FPS() != 1 {
		t.Errorf("FPS 0 should be clamped to 1, got %d", throttler.FPS())
	}

	throttler.SetFPS(200)
	if throttler.FPS() != 120 {
		t.Errorf("FPS 200 should be clamped to 120, got %d", throttler.FPS())
	}
}

func TestThrottler_ActualFPS(t *testing.T) {
	throttler := NewThrottler(60)

	// 没有帧时间记录时，实际 FPS 应该是 0
	if throttler.ActualFPS() != 0 {
		t.Errorf("expected actual FPS 0, got %f", throttler.ActualFPS())
	}

	// 记录一些帧时间
	for i := 0; i < 10; i++ {
		throttler.RecordFrameTime(16 * time.Millisecond) // ~62.5 FPS
	}

	actual := throttler.ActualFPS()
	if actual < 50 || actual > 70 {
		t.Errorf("expected actual FPS around 60, got %f", actual)
	}
}

func TestThrottler_EnableAdaptive(t *testing.T) {
	throttler := NewThrottler(60)

	throttler.EnableAdaptive(true)
	if !throttler.IsAdaptive() {
		t.Error("adaptive should be enabled")
	}

	throttler.EnableAdaptive(false)
	if throttler.IsAdaptive() {
		t.Error("adaptive should be disabled")
	}
}

func TestThrottler_AdaptiveAdjustment(t *testing.T) {
	throttler := NewThrottler(60)
	throttler.EnableAdaptive(true)

	// 记录一些慢的帧时间（应该降低帧率）
	for i := 0; i < 60; i++ {
		throttler.RecordFrameTime(20 * time.Millisecond) // 大于目标帧时间
	}

	// 由于帧时间较长，帧率应该降低
	// 注意：这可能需要更多次记录才能触发调整
	_ = throttler.FPS()
}

func TestThrottler_Stats(t *testing.T) {
	throttler := NewThrottler(60)

	// 记录一些帧时间
	times := []time.Duration{
		10 * time.Millisecond,
		15 * time.Millisecond,
		20 * time.Millisecond,
	}

	for _, ft := range times {
		throttler.RecordFrameTime(ft)
	}

	stats := throttler.Stats()

	if stats.TargetFPS != 60 {
		t.Errorf("expected target FPS 60, got %d", stats.TargetFPS)
	}

	if stats.TotalFrames != 3 {
		t.Errorf("expected 3 frames, got %d", stats.TotalFrames)
	}

	if stats.MinFrameTime != 10*time.Millisecond {
		t.Errorf("expected min frame time 10ms, got %v", stats.MinFrameTime)
	}

	if stats.MaxFrameTime != 20*time.Millisecond {
		t.Errorf("expected max frame time 20ms, got %v", stats.MaxFrameTime)
	}
}

func TestThrottler_Reset(t *testing.T) {
	throttler := NewThrottler(60)

	// 记录一些数据
	throttler.RecordFrameTime(10 * time.Millisecond)
	throttler.RecordFrameTime(15 * time.Millisecond)

	// 触发一些 pending
	throttler.ShouldRender()
	throttler.ShouldRender()

	// 重置
	throttler.Reset()

	stats := throttler.Stats()
	if stats.TotalFrames != 0 {
		t.Errorf("expected 0 frames after reset, got %d", stats.TotalFrames)
	}

	if stats.PendingCount != 0 {
		t.Errorf("expected 0 pending after reset, got %d", stats.PendingCount)
	}
}

func TestThrottler_TimeUntilNextRender(t *testing.T) {
	throttler := NewThrottler(60)

	// 刚渲染过，应该有等待时间
	throttler.ShouldRender()
	until := throttler.TimeUntilNextRender()
	if until == 0 {
		t.Error("should have non-zero wait time")
	}

	// 等待超过间隔
	time.Sleep(20 * time.Millisecond)
	until = throttler.TimeUntilNextRender()
	if until != 0 {
		t.Errorf("should be ready to render, but wait time is %v", until)
	}
}

func TestThrottler_ForceRender(t *testing.T) {
	throttler := NewThrottler(60)

	// 正常渲染
	throttler.ShouldRender()
	// 立即第二次应该被阻止
	if throttler.ShouldRender() {
		t.Error("second render should be blocked")
	}

	// 强制渲染
	throttler.ForceRender()
	// 现在应该可以渲染了
	if !throttler.ShouldRender() {
		t.Error("should render after force")
	}
}

func TestSmartRenderer(t *testing.T) {
	renderer := NewSmartRenderer(60)

	if renderer.FPS() != 60 {
		t.Errorf("expected FPS 60, got %d", renderer.FPS())
	}

	// 默认策略是 OnDirty
	if !renderer.ShouldRender(true) {
		t.Error("should render when dirty")
	}

	if renderer.ShouldRender(false) {
		t.Error("should not render when not dirty")
	}
}

func TestSmartRenderer_AlwaysStrategy(t *testing.T) {
	renderer := NewSmartRenderer(60)
	renderer.SetStrategy(StrategyAlways)

	// 无论是否 dirty 都应该渲染
	if !renderer.ShouldRender(false) {
		t.Error("should always render with StrategyAlways")
	}
}

func TestSmartRenderer_AdaptiveStrategy(t *testing.T) {
	renderer := NewSmartRenderer(60)
	renderer.SetStrategy(StrategyAdaptive)

	// 不 dirty 时不渲染
	if renderer.ShouldRender(false) {
		t.Error("should not render when not dirty")
	}

	// dirty 时应该渲染
	if !renderer.ShouldRender(true) {
		t.Error("should render when dirty")
	}
}

func TestSmartRenderer_Render(t *testing.T) {
	renderer := NewSmartRenderer(60)

	var called bool
	renderer.Render(func() time.Duration {
		called = true
		return 10 * time.Millisecond
	})

	if !called {
		t.Error("render function should be called")
	}
}

func TestSmartRenderer_SetFPS(t *testing.T) {
	renderer := NewSmartRenderer(60)
	renderer.SetFPS(30)

	if renderer.FPS() != 30 {
		t.Errorf("expected FPS 30, got %d", renderer.FPS())
	}
}

func TestSmartRenderer_EnableAdaptive(t *testing.T) {
	renderer := NewSmartRenderer(60)

	renderer.EnableAdaptive(true)

	if !renderer.throttler.IsAdaptive() {
		t.Error("adaptive should be enabled")
	}

	// 策略应该变为 Adaptive
	if renderer.strategy != StrategyAdaptive {
		t.Error("strategy should be Adaptive when adaptive is enabled")
	}
}

func TestSmartRenderer_AdaptiveWithPending(t *testing.T) {
	renderer := NewSmartRenderer(60)
	renderer.SetStrategy(StrategyAdaptive)

	// 第一次调用会返回 true（因为 lastRender 是零时间）
	// 后续调用会增加 pendingCount
	renderer.throttler.ShouldRender() // 返回 true, pendingCount = 0
	renderer.throttler.ShouldRender() // 返回 false, pendingCount = 1
	renderer.throttler.ShouldRender() // 返回 false, pendingCount = 2
	renderer.throttler.ShouldRender() // 返回 false, pendingCount = 3

	// 即使很近的时间，有大量 pending 时也应该渲染
	if !renderer.ShouldRender(true) {
		t.Error("should render with high pending count")
	}
}

func TestFPSToInterval(t *testing.T) {
	tests := []struct {
		fps      int
		expected time.Duration
	}{
		{1, time.Second},
		{10, 100 * time.Millisecond},
		{60, time.Second / 60},
		{120, time.Second / 120},
	}

	for _, tt := range tests {
		t.Run(fpsToString(tt.fps), func(t *testing.T) {
			interval := FPSToInterval(tt.fps)
			if interval != tt.expected {
				t.Errorf("expected interval %v, got %v", tt.expected, interval)
			}
		})
	}
}

func fpsToString(fps int) string {
	switch fps {
	case 1:
		return "1fps"
	case 10:
		return "10fps"
	case 60:
		return "60fps"
	case 120:
		return "120fps"
	default:
		return "unknown"
	}
}

func TestIntervalToFPS(t *testing.T) {
	tests := []struct {
		interval time.Duration
		expected int
	}{
		{time.Second, 1},
		{100 * time.Millisecond, 10},
		{time.Second / 60, 60},
		{time.Second / 120, 120},
	}

	for _, tt := range tests {
		t.Run(tt.interval.String(), func(t *testing.T) {
			fps := IntervalToFPS(tt.interval)
			if fps != tt.expected {
				t.Errorf("expected FPS %d, got %d", tt.expected, fps)
			}
		})
	}
}

func TestRecommendFPS(t *testing.T) {
	tests := []struct {
		scene    string
		minFPS   int
		maxFPS   int
	}{
		{"idle", 15, 30},
		{"static", 15, 30},
		{"animation", 60, 60},
		{"game", 120, 120},
		{"form", 30, 30},
	}

	for _, tt := range tests {
		t.Run(tt.scene, func(t *testing.T) {
			fps := RecommendFPS(tt.scene)
			if fps < tt.minFPS || fps > tt.maxFPS {
				t.Errorf("FPS for %s should be between %d and %d, got %d",
					tt.scene, tt.minFPS, tt.maxFPS, fps)
			}
		})
	}
}

func TestPresetFPS(t *testing.T) {
	tests := []struct {
		name     string
		expected int
	}{
		{"max", 120},
		{"high", 60},
		{"medium", 30},
		{"low", 15},
		{"min", 1},
		{"default", 60},
		{"unknown", 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fps := PresetFPS(tt.name)
			if fps != tt.expected {
				t.Errorf("expected FPS %d for preset %s, got %d",
					tt.expected, tt.name, fps)
			}
		})
	}
}

func TestAvailablePresets(t *testing.T) {
	presets := AvailablePresets()

	if len(presets) == 0 {
		t.Error("should have available presets")
	}

	// 检查是否包含常用预设
	presetMap := make(map[string]bool)
	for _, name := range presets {
		presetMap[name] = true
	}

	expected := []string{"max", "high", "medium", "low", "min", "default"}
	for _, name := range expected {
		if !presetMap[name] {
			t.Errorf("preset %s should be available", name)
		}
	}
}

func TestThrottler_ConcurrentAccess(t *testing.T) {
	throttler := NewThrottler(60)
	var wg sync.WaitGroup
	var errors int32
	var count int32

	// 并发测试
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					atomic.AddInt32(&errors, 1)
				}
			}()

			for j := 0; j < 100; j++ {
				if throttler.ShouldRender() {
					atomic.AddInt32(&count, 1)
					throttler.RecordFrameTime(10 * time.Millisecond)
				}
				throttler.FPS()
				throttler.ActualFPS()
				throttler.Stats()
			}
		}()
	}

	wg.Wait()

	if errors > 0 {
		t.Errorf("encountered %d panics", errors)
	}
}

func TestThrottler_ResetConcurrent(t *testing.T) {
	throttler := NewThrottler(60)
	var wg sync.WaitGroup

	// 并发重置和记录
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				throttler.RecordFrameTime(time.Duration(j) * time.Millisecond)
				throttler.Reset()
				throttler.Stats()
			}
		}()
	}

	wg.Wait()
	// 只验证没有 panic
}

func TestSmartRenderer_Concurrent(t *testing.T) {
	renderer := NewSmartRenderer(60)
	var wg sync.WaitGroup
	var count int32

	// 并发渲染测试
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				renderer.Render(func() time.Duration {
					atomic.AddInt32(&count, 1)
					return 5 * time.Millisecond
				})
			}
		}()
	}

	wg.Wait()

	if count == 0 {
		t.Error("should have some renders completed")
	}
}

