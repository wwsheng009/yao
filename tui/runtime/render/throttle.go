package render

import (
	"sync"
	"time"
)

// ==============================================================================
// Frame Throttling System (V3)
// ==============================================================================
// 帧率限制器，防止过度渲染导致的性能浪费

// Throttler 帧率限制器
type Throttler struct {
	mu               sync.Mutex
	targetFPS        int
	minInterval      time.Duration
	lastRender       time.Time
	pendingCount     int
	frameTimes       []time.Duration
	frameIndex       int
	adaptive         bool
	targetFrameTime  time.Duration
	lastStatsUpdate  time.Time
	totalFrames      int64
	totalRenderTime  time.Duration
}

// NewThrottler 创建限制器
func NewThrottler(fps int) *Throttler {
	if fps <= 0 {
		fps = 60 // 默认 60 FPS
	}
	if fps > 120 {
		fps = 120 // 最高 120 FPS
	}

	interval := time.Second / time.Duration(fps)

	return &Throttler{
		targetFPS:       fps,
		minInterval:     interval,
		frameTimes:      make([]time.Duration, 60), // 保存最近 60 帧
		adaptive:        false,
		targetFrameTime: interval * 90 / 100, // 目标渲染时间小于帧间隔
		lastRender:      time.Time{}, // 初始化为零时间，使得首次调用总是返回 true
		lastStatsUpdate: time.Now(),
	}
}

// ShouldRender 检查是否应该渲染
func (t *Throttler) ShouldRender() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(t.lastRender)

	if elapsed < t.minInterval {
		t.pendingCount++
		return false
	}

	t.lastRender = now
	t.pendingCount = 0
	return true
}

// RecordFrameTime 记录实际渲染时间
func (t *Throttler) RecordFrameTime(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.frameTimes[t.frameIndex] = d
	t.frameIndex = (t.frameIndex + 1) % len(t.frameTimes)

	t.totalFrames++
	t.totalRenderTime += d

	// 自适应调整
	if t.adaptive {
		t.adjust()
	}
}

// adjust 根据渲染性能动态调整
func (t *Throttler) adjust() {
	now := time.Now()

	// 每秒最多调整一次
	if now.Sub(t.lastStatsUpdate) < time.Second {
		return
	}
	t.lastStatsUpdate = now

	// 计算平均渲染时间
	var sum time.Duration
	count := 0
	for _, ft := range t.frameTimes {
		if ft > 0 {
			sum += ft
			count++
		}
	}

	if count == 0 {
		return
	}

	avg := sum / time.Duration(count)

	// 如果渲染时间过长，降低帧率
	if avg > t.targetFrameTime && t.targetFPS > 30 {
		t.targetFPS = t.targetFPS * 8 / 10 // 降低 20%
		if t.targetFPS < 30 {
			t.targetFPS = 30
		}
		t.minInterval = time.Second / time.Duration(t.targetFPS)
	} else if avg < t.targetFrameTime/2 && t.targetFPS < 120 {
		// 渲染很快，可以提高帧率
		t.targetFPS = t.targetFPS * 12 / 10 // 提高 20%
		if t.targetFPS > 120 {
			t.targetFPS = 120
		}
		t.minInterval = time.Second / time.Duration(t.targetFPS)
	}
}

// SetFPS 设置目标帧率
func (t *Throttler) SetFPS(fps int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if fps < 1 {
		fps = 1
	}
	if fps > 120 {
		fps = 120
	}

	t.targetFPS = fps
	t.minInterval = time.Second / time.Duration(fps)
}

// FPS 返回当前帧率
func (t *Throttler) FPS() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.targetFPS
}

// ActualFPS 返回实际帧率（基于最近渲染时间）
func (t *Throttler) ActualFPS() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	var sum time.Duration
	count := 0
	for _, ft := range t.frameTimes {
		if ft > 0 {
			sum += ft
			count++
		}
	}

	if count == 0 || sum == 0 {
		return 0
	}

	avg := sum / time.Duration(count)
	return float64(time.Second) / float64(avg)
}

// EnableAdaptive 启用自适应帧率
func (t *Throttler) EnableAdaptive(enable bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.adaptive = enable
}

// IsAdaptive 检查是否启用自适应
func (t *Throttler) IsAdaptive() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.adaptive
}

// Stats 返回统计数据
type Stats struct {
	TargetFPS     int
	ActualFPS     float64
	PendingCount  int
	AvgFrameTime  time.Duration
	MinFrameTime  time.Duration
	MaxFrameTime  time.Duration
	TotalFrames   int64
	TotalTime     time.Duration
	Adaptive      bool
}

// Stats 返回统计信息
func (t *Throttler) Stats() Stats {
	t.mu.Lock()
	defer t.mu.Unlock()

	var sum time.Duration
	minTime := time.Duration(-1)
	maxTime := time.Duration(0)
	count := 0

	for _, ft := range t.frameTimes {
		if ft > 0 {
			sum += ft
			count++
			if minTime < 0 || ft < minTime {
				minTime = ft
			}
			if ft > maxTime {
				maxTime = ft
			}
		}
	}

	var avg time.Duration
	if count > 0 {
		avg = sum / time.Duration(count)
	}

	actualFPS := 0.0
	if avg > 0 {
		actualFPS = float64(time.Second) / float64(avg)
	}

	return Stats{
		TargetFPS:     t.targetFPS,
		ActualFPS:     actualFPS,
		PendingCount:  t.pendingCount,
		AvgFrameTime:  avg,
		MinFrameTime:  minTime,
		MaxFrameTime:  maxTime,
		TotalFrames:   t.totalFrames,
		TotalTime:     t.totalRenderTime,
		Adaptive:      t.adaptive,
	}
}

// Reset 重置统计信息
func (t *Throttler) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.frameTimes = make([]time.Duration, 60)
	t.frameIndex = 0
	t.totalFrames = 0
	t.totalRenderTime = 0
	t.pendingCount = 0
	t.lastStatsUpdate = time.Now()
}

// TimeUntilNextRender 返回距离下次可渲染的时间
func (t *Throttler) TimeUntilNextRender() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()

	elapsed := time.Now().Sub(t.lastRender)
	if elapsed >= t.minInterval {
		return 0
	}
	return t.minInterval - elapsed
}

// ForceRender 强制下次渲染（跳过限制）
func (t *Throttler) ForceRender() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastRender = time.Time{} // 重置为零时间，使得下次调用立即返回 true
}

// SetTargetFrameTime 设置目标帧渲染时间
func (t *Throttler) SetTargetFrameTime(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.targetFrameTime = d
}

// ==============================================================================
// 渲染策略
// ==============================================================================

// RenderStrategy 渲染策略
type RenderStrategy int

const (
	// StrategyAlways 总是渲染
	StrategyAlways RenderStrategy = iota
	// StrategyOnDirty 仅在脏标记时渲染
	StrategyOnDirty
	// StrategyAdaptive 自适应渲染
	StrategyAdaptive
)

// SmartRenderer 智能渲染器
type SmartRenderer struct {
	throttler *Throttler
	strategy  RenderStrategy
}

// NewSmartRenderer 创建智能渲染器
func NewSmartRenderer(fps int) *SmartRenderer {
	return &SmartRenderer{
		throttler: NewThrottler(fps),
		strategy:  StrategyOnDirty,
	}
}

// ShouldRender 检查是否应该渲染
func (r *SmartRenderer) ShouldRender(isDirty bool) bool {
	switch r.strategy {
	case StrategyAlways:
		return r.throttler.ShouldRender()
	case StrategyOnDirty:
		return isDirty && r.throttler.ShouldRender()
	case StrategyAdaptive:
		// 结合脏标记和帧率限制
		if !isDirty {
			return false
		}
		// 如果有大量待渲染帧，允许更高帧率
		if r.throttler.pendingCount > 2 {
			return true
		}
		return r.throttler.ShouldRender()
	default:
		return r.throttler.ShouldRender()
	}
}

// Render 渲染（带记录）
func (r *SmartRenderer) Render(fn func() time.Duration) {
	if r.throttler.ShouldRender() {
		start := time.Now()
		fn()
		r.throttler.RecordFrameTime(time.Since(start))
	}
}

// SetStrategy 设置渲染策略
func (r *SmartRenderer) SetStrategy(s RenderStrategy) {
	r.strategy = s
}

// SetFPS 设置帧率
func (r *SmartRenderer) SetFPS(fps int) {
	r.throttler.SetFPS(fps)
}

// FPS 返回当前帧率
func (r *SmartRenderer) FPS() int {
	return r.throttler.FPS()
}

// ActualFPS 返回实际帧率
func (r *SmartRenderer) ActualFPS() float64 {
	return r.throttler.ActualFPS()
}

// Stats 返回统计信息
func (r *SmartRenderer) Stats() Stats {
	return r.throttler.Stats()
}

// EnableAdaptive 启用自适应帧率
func (r *SmartRenderer) EnableAdaptive(enable bool) {
	r.throttler.EnableAdaptive(enable)
	if enable && r.strategy != StrategyAdaptive {
		r.strategy = StrategyAdaptive
	}
}

// Throttler 返回底层节流器
func (r *SmartRenderer) Throttler() *Throttler {
	return r.throttler
}

// ==============================================================================
// 辅助函数
// ==============================================================================

// FPSToInterval 将 FPS 转换为时间间隔
func FPSToInterval(fps int) time.Duration {
	if fps <= 0 {
		fps = 60
	}
	return time.Second / time.Duration(fps)
}

// IntervalToFPS 将时间间隔转换为 FPS
func IntervalToFPS(interval time.Duration) int {
	if interval <= 0 {
		return 60
	}
	fps := int(float64(time.Second) / float64(interval))
	if fps > 120 {
		fps = 120
	}
	if fps < 1 {
		fps = 1
	}
	return fps
}

// RecommendFPS 根据场景推荐帧率
func RecommendFPS(scene string) int {
	switch scene {
	case "idle", "static":
		return 20 // 空闲或静态场景，低帧率
	case "animation", "transitions":
		return 60 // 动画场景
	case "game", "realtime":
		return 120 // 游戏或实时场景
	case "form", "input":
		return 30 // 表单输入
	default:
		return 60 // 默认
	}
}

// CommonPresets 常用预设
var CommonPresets = map[string]int{
	"max":      120, // 最大帧率
	"high":     60,  // 高帧率
	"medium":   30,  // 中等帧率
	"low":      15,  // 低帧率
	"min":      1,   // 最小帧率
	"default":  60,  // 默认帧率
	"animation": 60,  // 动画
	"game":     120, // 游戏
	"static":   15,  // 静态内容
}

// PresetFPS 获取预设帧率
func PresetFPS(name string) int {
	if fps, ok := CommonPresets[name]; ok {
		return fps
	}
	return 60 // 默认
}

// AvailablePresets 返回可用的预设名称
func AvailablePresets() []string {
	names := make([]string, 0, len(CommonPresets))
	for name := range CommonPresets {
		names = append(names, name)
	}
	return names
}
