# 帧率限制增强方案

## 概述

帧率限制功能借鉴自 Bubble Tea 的渲染器 FPS 限制机制，防止过度渲染导致的性能浪费和 CPU 占用过高。

## 当前问题

### 无限制的渲染循环

当前 Yao TUI 在事件驱动时每次都触发完整重绘，可能导致：

1. **CPU 浪费** - 高频事件（如鼠标移动）导致过度渲染
2. **性能抖动** - 渲染耗时变化导致帧率不稳定
3. **电量消耗** - 移动设备上电池续航降低
4. **终端卡顿** - 某些终端无法处理高频更新

### 具体场景

```go
// 问题场景：鼠标移动事件触发过多重绘
func (c *Container) HandleMouseMove(event MouseEvent) {
    c.mousePos = event.Position
    c.MarkDirty()        // 每次鼠标移动都重绘
    c.RequestRender()    // 可能达到每秒数百次
}
```

## 设计方案

### 核心接口

```go
// tui/runtime/render/throttle.go

package render

import (
    "sync"
    "time"
)

// Throttler 帧率限制器
type Throttler struct {
    mu           sync.Mutex
    targetFPS    int
    minInterval  time.Duration
    lastRender   time.Time
    pendingCount int

    // 帧时间统计
    frameTimes   []time.Duration
    frameIndex   int

    // 动态调整
    adaptive     bool
    targetFrameTime time.Duration
}

// NewThrottler 创建限制器
func NewThrottler(fps int) *Throttler {
    if fps <= 0 {
        fps = 60 // 默认 60 FPS
    }

    interval := time.Second / time.Duration(fps)

    return &Throttler{
        targetFPS:       fps,
        minInterval:     interval,
        frameTimes:      make([]time.Duration, 60), // 保存最近 60 帧
        adaptive:        false,
        targetFrameTime: interval * 90 / 100, // 目标渲染时间小于帧间隔
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
    return true
}

// RecordFrameTime 记录实际渲染时间
func (t *Throttler) RecordFrameTime(d time.Duration) {
    t.mu.Lock()
    defer t.mu.Unlock()

    t.frameTimes[t.frameIndex] = d
    t.frameIndex = (t.frameIndex + 1) % len(t.frameTimes)

    // 自适应调整
    if t.adaptive {
        t.adjust()
    }
}

// adjust 根据渲染性能动态调整
func (t *Throttler) adjust() {
    // 计算平均渲染时间
    var sum time.Duration
    for _, ft := range t.frameTimes {
        sum += ft
    }
    avg := sum / time.Duration(len(t.frameTimes))

    // 如果渲染时间过长，降低帧率
    if avg > t.targetFrameTime {
        newFPS := t.targetFPS * 8 / 10 // 降低 20%
        if newFPS >= 30 { // 最低 30 FPS
            t.SetFPS(newFPS)
        }
    } else if avg < t.targetFrameTime/2 {
        // 渲染很快，可以提高帧率
        newFPS := t.targetFPS * 12 / 10 // 提高 20%
        if newFPS <= 120 { // 最高 120 FPS
            t.SetFPS(newFPS)
        }
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

    if count == 0 {
        return 0
    }

    avg := sum / time.Duration(count)
    if avg == 0 {
        return 0
    }

    return float64(time.Second) / float64(avg)
}

// EnableAdaptive 启用自适应帧率
func (t *Throttler) EnableAdaptive(enable bool) {
    t.mu.Lock()
    defer t.mu.Unlock()
    t.adaptive = enable
}

// Stats 返回统计数据
type Stats struct {
    TargetFPS    int
    ActualFPS    float64
    PendingCount int
    AvgFrameTime time.Duration
}

func (t *Throttler) Stats() Stats {
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

    return Stats{
        TargetFPS:    t.targetFPS,
        ActualFPS:    t.ActualFPS(),
        PendingCount: t.pendingCount,
        AvgFrameTime: sum / time.Duration(count),
    }
}
```

### 集成到渲染循环

```go
// tui/runtime/render/renderer.go

package render

import (
    "context"
    "time"
)

type Renderer struct {
    throttler *Throttler
    // 现有字段...
}

func NewRenderer() *Renderer {
    return &Renderer{
        throttler: NewThrottler(60), // 默认 60 FPS
    }
}

// SetFPS 设置渲染帧率
func (r *Renderer) SetFPS(fps int) {
    r.throttler.SetFPS(fps)
}

// Render 渲染（带节流）
func (r *Renderer) Render(ctx context.Context, root Component) error {
    // 检查是否应该渲染
    if !r.throttler.ShouldRender() {
        return nil
    }

    start := time.Now()

    // 执行实际渲染
    err := r.doRender(ctx, root)

    // 记录渲染时间
    r.throttler.RecordFrameTime(time.Since(start))

    return err
}

// doRender 实际渲染逻辑
func (r *Renderer) doRender(ctx context.Context, root Component) error {
    // 现有渲染逻辑...
    return nil
}
```

### 智能渲染策略

```go
// tui/runtime/render/smart.go

package render

import (
    "context"
)

// SmartRenderer 智能渲染器
type SmartRenderer struct {
    *Renderer
    strategy RenderStrategy
}

type RenderStrategy int

const (
    // StrategyAlways 总是渲染
    StrategyAlways RenderStrategy = iota
    // StrategyOnDirty 仅在脏标记时渲染
    StrategyOnDirty
    // StrategyAdaptive 自适应渲染
    StrategyAdaptive
)

func NewSmartRenderer() *SmartRenderer {
    return &SmartRenderer{
        Renderer: NewRenderer(),
        strategy: StrategyOnDirty,
    }
}

// Render 智能渲染
func (r *SmartRenderer) Render(ctx context.Context, root Component) error {
    switch r.strategy {
    case StrategyAlways:
        return r.doRender(ctx, root)

    case StrategyOnDirty:
        if root.IsDirty() {
            return r.doRender(ctx, root)
        }
        return nil

    case StrategyAdaptive:
        // 结合脏标记和帧率限制
        if root.IsDirty() && r.throttler.ShouldRender() {
            return r.doRender(ctx, root)
        }
        return nil
    }

    return nil
}

// SetStrategy 设置渲染策略
func (r *SmartRenderer) SetStrategy(s RenderStrategy) {
    r.strategy = s
}
```

## 使用示例

### 1. 基础使用

```go
app := NewApp()
app.Renderer().SetFPS(60) // 限制 60 FPS
```

### 2. 自适应帧率

```go
renderer := NewRenderer()
renderer.throttler.EnableAdaptive(true)

// 在高负载场景自动降低帧率
```

### 3. 监控渲染性能

```go
// 定期输出统计信息
go func() {
    ticker := time.NewTicker(5 * time.Second)
    for range ticker.C {
        stats := app.Renderer().Throttler().Stats()
        log.Printf("FPS: target=%d actual=%.1f avg=%v",
            stats.TargetFPS,
            stats.ActualFPS,
            stats.AvgFrameTime,
        )
    }
}()
```

### 4. 按场景切换帧率

```go
func (a *App) SetScene(scene Scene) {
    switch scene {
    case SceneIdle:
        a.renderer.SetFPS(30) // 空闲时降低帧率
    case SceneAnimation:
        a.renderer.SetFPS(60) // 动画时正常帧率
    case SceneGame:
        a.renderer.SetFPS(120) // 游戏时高帧率
    }
}
```

### 5. 调试模式

```go
func (r *Renderer) EnableDebugInfo(enable bool) {
    if enable {
        // 在屏幕右上角显示 FPS
        r.debugOverlay = &DebugOverlay{
            throttler: r.throttler,
        }
    } else {
        r.debugOverlay = nil
    }
}
```

## 实施计划

### Phase 1: 核心节流器 (Week 1)

- [ ] 实现 `Throttler`
- [ ] 实现 `ShouldRender()`
- [ ] 实现 `RecordFrameTime()`
- [ ] 单元测试

### Phase 2: 渲染器集成 (Week 1)

- [ ] 集成到 `Renderer`
- [ ] 实现 `SetFPS()`
- [ ] 集成测试

### Phase 3: 智能渲染 (Week 2)

- [ ] 实现 `SmartRenderer`
- [ ] 实现渲染策略
- [ ] 性能测试

### Phase 4: 调试工具 (Week 2)

- [ ] 实现 FPS 叠加层
- [ ] 实现统计输出
- [ ] 文档和示例

## 文件结构

```
tui/runtime/render/
├── throttle.go           # 节流器
├── smart.go              # 智能渲染器
├── fps.go                # FPS 计算
├── debug.go              # 调试叠加层
└── throttle_test.go      # 测试
```

## 测试策略

```go
func TestThrottler(t *testing.T) {
    throttler := NewThrottler(60) // 60 FPS = ~16.67ms 间隔

    assert.True(t, throttler.ShouldRender())
    time.Sleep(10 * time.Millisecond)
    assert.False(t, throttler.ShouldRender())
    time.Sleep(10 * time.Millisecond)
    assert.True(t, throttler.ShouldRender())
}

func TestFPSAdjustment(t *testing.T) {
    throttler := NewThrottler(60)

    // 记录一些较慢的帧
    for i := 0; i < 60; i++ {
        throttler.RecordFrameTime(20 * time.Millisecond)
    }

    // 应该降低帧率
    throttler.EnableAdaptive(true)
    throttler.adjust()

    assert.Less(t, throttler.FPS(), 60)
}

func TestActualFPS(t *testing.T) {
    throttler := NewThrottler(60)

    start := time.Now()
    count := 0

    for time.Since(start) < time.Second {
        if throttler.ShouldRender() {
            count++
            throttler.RecordFrameTime(5 * time.Millisecond)
        }
        time.Sleep(time.Millisecond)
    }

    actualFPS := throttler.ActualFPS()
    assert.InDelta(t, 60, actualFPS, 5)
}
```

## 性能考虑

1. **零开销** - 不渲染时开销极小
2. **统计采样** - 只采样最近帧，避免内存增长
3. **锁优化** - 使用细粒度锁保护关键区域

## 推荐配置

| 场景 | 推荐帧率 | 说明 |
|------|---------|------|
| 普通应用 | 30-60 | 平衡性能和流畅度 |
| 动画效果 | 60 | 保证流畅 |
| 游戏/实时 | 60-120 | 低延迟 |
| 静态表单 | 15-30 | 节省资源 |
| 调试模式 | Unlimted | 实时反馈 |

## 与 Bubble Tea 的差异

| 特性 | Bubble Tea | Yao TUI Throttler |
|------|-----------|------------------|
| **默认 FPS** | 60 | 60 |
| **最大 FPS** | 120 | 120 |
| **自适应** | 无 | 有 |
| **统计信息** | 无 | 详细统计 |
| **渲染策略** | 总是渲染 | 支持按需渲染 |
| **调试信息** | 无 | 内置叠加层 |
