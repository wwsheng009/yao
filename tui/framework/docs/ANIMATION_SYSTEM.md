# Animation System Design (V3)

> **版本**: V3
> **核心原则**: 按需 Tick（On-demand Tick）
> **关键特性**: 节能、精确控制、可组合

## 概述

动画系统负责管理 UI 中的动态效果，如进度条加载、闪烁光标、平滑滚动等。V3 架构采用**按需 Tick** 模式，而不是传统的全局定时器。

### 为什么需要按需 Tick？

**传统全局 Tick 的问题**：
```go
// ❌ 全局 Tick - 浪费资源
func (app *App) Run() {
    ticker := time.NewTicker(16ms) // 60 FPS
    for {
        select {
        case <-ticker.C:
            app.Update()  // 即使没有动画也在运行
        case <-app.Events():
            app.HandleEvent()
        }
    }
}
```

**按需 Tick 的优势**：
```go
// ✅ 按需 Tick - 只在需要时运行
func (mgr *Manager) Start() {
    mgr.ticker = time.NewTicker(16ms)
    mgr.ticker.Stop()  // 默认停止
}

func (mgr *Manager) AddAnimation(anim Animation) {
    mgr.animations = append(mgr.animations, anim)
    if len(mgr.animations) == 1 {
        mgr.ticker.Start()  // 第一个动画时才启动
    }
}
```

## 设计目标

1. **节能**: 只在有活动动画时运行
2. **精确控制**: 支持暂停、恢复、取消
3. **可组合**: 多个动画可以并行或串行
4. **高性能**: 最小化渲染开销

## 架构概览

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Animation Flow                                   │
└─────────────────────────────────────────────────────────────────────────┘

    Component Request
          │
          ▼
┌─────────────────┐
│ Animation Manager│
│  - Register      │
│  - Start/Stop    │
└─────────────────┘
          │
          ▼ (if has active animations)
┌─────────────────┐
│  Ticker Loop    │ ←── On-demand (only when needed)
└─────────────────┘
          │
          ▼
┌─────────────────┐
│  Update Phase   │
│  - Progress     │
│  - Easing       │
└─────────────────┘
          │
          ▼
┌─────────────────┐
│  Mark Dirty     │
└─────────────────┘
          │
          ▼
┌─────────────────┐
│  Render Phase   │
└─────────────────┘
          │
          ▼ (if all animations done)
┌─────────────────┐
│  Stop Ticker    │
└─────────────────┘
```

## 核心类型定义

### 1. Animation 接口

```go
// 位于: tui/runtime/animation/animation.go

package animation

// Animation 动画接口
type Animation interface {
    // ID 获取动画 ID
    ID() string

    // Update 更新动画状态
    // 返回: (继续运行, 完成进度 0-1)
    Update(dt time.Duration) (running bool, progress float64)

    // Apply 应用动画到目标
    Apply(ctx *Context)

    // Duration 获取动画时长
    Duration() time.Duration

    // IsDone 是否完成
    IsDone() bool
}

// Context 动画上下文
type Context struct {
    // 时间
    Elapsed time.Duration

    // 进度 (0-1)
    Progress float64

    // 缓动函数
    Easing EasingFunction

    // 目标组件 ID
    TargetID string

    // 元数据
    Metadata map[string]interface{}
}

// EasingFunction 缓动函数
type EasingFunction func(float64) float64
```

### 2. 内置缓动函数

```go
// 位于: tui/runtime/animation/easing.go

package animation

import "math"

// 缓动函数
var (
    // Linear 线性
    Linear EasingFunction = func(t float64) float64 {
        return t
    }

    // EaseInQuad 二次缓入
    EaseInQuad EasingFunction = func(t float64) float64 {
        return t * t
    }

    // EaseOutQuad 二次缓出
    EaseOutQuad EasingFunction = func(t float64) float64 {
        return t * (2 - t)
    }

    // EaseInOutQuad 二次缓入缓出
    EaseInOutQuad EasingFunction = func(t float64) float64 {
        if t < 0.5 {
            return 2 * t * t
        }
        return -1 + (4-2*t)*t
    }

    // EaseInCubic 三次缓入
    EaseInCubic EasingFunction = func(t float64) float64 {
        return t * t * t
    }

    // EaseOutCubic 三次缓出
    EaseOutCubic EasingFunction = func(t float64) float64 {
        t -= 1
        return t*t*t + 1
    }

    // EaseInOutCubic 三次缓入缓出
    EaseInOutCubic EasingFunction = func(t float64) float64 {
        if t < 0.5 {
            return 4 * t * t * t
        }
        t -= 1
        return 1 + 4*t*t*t
    }

    // EaseOutElastic 弹性缓出
    EaseOutElastic EasingFunction = func(t float64) float64 {
        const p = 0.3
        if t == 0 {
            return 0
        }
        if t == 1 {
            return 1
        }
        return math.Pow(2, -10*t) * math.Sin((t*10-p*4)/(2*math.Pi)) + 1
    }

    // EaseOutBounce 回弹缓出
    EaseOutBounce EasingFunction = func(t float64) float64 {
        const n1 = 7.5625
        const d1 = 2.75

        if t < 1/d1 {
            return n1 * t * t
        } else if t < 2/d1 {
            t -= 1.5 / d1
            return n1*t*t + 0.75
        } else if t < 2.5/d1 {
            t -= 2.25 / d1
            return n1*t*t + 0.9375
        } else {
            t -= 2.625 / d1
            return n1*t*t + 0.984375
        }
    }
)
```

### 3. Animation Manager

```go
// 位于: tui/runtime/animation/manager.go

package animation

// Manager 动画管理器
type Manager struct {
    mu sync.RWMutex

    // 活动动画列表
    animations []Animation

    // 已完成的动画（等待清理）
    completed []Animation

    // 定时器
    ticker *time.Ticker

    // 是否正在运行
    running bool

    // 动画变化监听器
    listeners []AnimationListener

    // 默认帧率
    frameRate time.Duration
}

// AnimationListener 动画监听器
type AnimationListener func(anim Animation, progress float64)

// NewManager 创建动画管理器
func NewManager() *Manager {
    return &Manager{
        animations: make([]Animation, 0),
        completed:  make([]Animation, 0),
        ticker:     time.NewTicker(time.Second / 60), // 60 FPS
        running:    false,
        listeners:  make([]AnimationListener, 0),
        frameRate:  time.Second / 60,
    }
}

// Add 添加动画
func (m *Manager) Add(anim Animation) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.animations = append(m.animations, anim)

    // 如果这是第一个动画，启动定时器
    if len(m.animations) == 1 && !m.running {
        m.start()
    }
}

// Remove 移除动画
func (m *Manager) Remove(id string) bool {
    m.mu.Lock()
    defer m.mu.Unlock()

    for i, anim := range m.animations {
        if anim.ID() == id {
            m.animations = append(m.animations[:i], m.animations[i+1:]...)

            // 如果没有动画了，停止定时器
            if len(m.animations) == 0 {
                m.stop()
            }
            return true
        }
    }
    return false
}

// Get 获取动画
func (m *Manager) Get(id string) (Animation, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    for _, anim := range m.animations {
        if anim.ID() == id {
            return anim, true
        }
    }
    return nil, false
}

// Update 更新所有动画
func (m *Manager) Update() {
    m.mu.Lock()
    defer m.mu.Unlock()

    if len(m.animations) == 0 {
        m.stop()
        return
    }

    dt := m.frameRate
    stillRunning := make([]Animation, 0)

    for _, anim := range m.animations {
        running, progress := anim.Update(dt)

        if running {
            stillRunning = append(stillRunning, anim)
        } else {
            m.completed = append(m.completed, anim)
        }

        // 通知监听器
        m.notify(anim, progress)
    }

    m.animations = stillRunning

    // 如果没有活动动画了，停止定时器
    if len(m.animations) == 0 {
        m.stop()
    }
}

// Apply 应用所有活动动画
func (m *Manager) Apply(ctx *Context) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    for _, anim := range m.animations {
        anim.Apply(ctx)
    }
}

// HasActive 是否有活动动画
func (m *Manager) HasActive() bool {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return len(m.animations) > 0
}

// GetActive 获取活动动画列表
func (m *Manager) GetActive() []Animation {
    m.mu.RLock()
    defer m.mu.RUnlock()

    result := make([]Animation, len(m.animations))
    copy(result, m.animations)
    return result
}

// StopAll 停止所有动画
func (m *Manager) StopAll() {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.animations = m.animations[:0]
    m.stop()
}

// PauseAll 暂停所有动画
func (m *Manager) PauseAll() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.stop()
    m.running = false
}

// ResumeAll 恢复所有动画
func (m *Manager) ResumeAll() {
    m.mu.Lock()
    defer m.mu.Unlock()

    if len(m.animations) > 0 && !m.running {
        m.start()
    }
}

// Tick 获取定时器通道（用于 Runtime 集成）
func (m *Manager) Tick() <-chan time.Time {
    return m.ticker.C
}

// Subscribe 订阅动画变化
func (m *Manager) Subscribe(listener AnimationListener) func() {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.listeners = append(m.listeners, listener)

    return func() {
        m.Unsubscribe(listener)
    }
}

// Unsubscribe 取消订阅
func (m *Manager) Unsubscribe(listener AnimationListener) {
    m.mu.Lock()
    defer m.mu.Unlock()

    for i, l := range m.listeners {
        if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
            m.listeners = append(m.listeners[:i], m.listeners[i+1:]...)
            break
        }
    }
}

// start 启动定时器
func (m *Manager) start() {
    if !m.running {
        m.running = true
        // ticker 已经在 NewManager 中创建，只需要重置
        m.ticker.Reset(m.frameRate)
    }
}

// stop 停止定时器
func (m *Manager) stop() {
    if m.running {
        m.running = false
        m.ticker.Stop()
    }
}

// notify 通知监听器
func (m *Manager) notify(anim Animation, progress float64) {
    for _, listener := range m.listeners {
        listener(anim, progress)
    }
}

// Cleanup 清理已完成的动画
func (m *Manager) Cleanup() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.completed = m.completed[:0]
}
```

## 具体动画实现

### 1. Transition 动画

```go
// 位于: tui/runtime/animation/transition.go

package animation

// Transition 过渡动画
type Transition struct {
    id         string
    duration   time.Duration
    elapsed    time.Duration
    from       interface{}
    to         interface{}
    current    interface{}
    easing     EasingFunction
    targetID   string
    done       bool
}

// TransitionConfig 过渡动画配置
type TransitionConfig struct {
    ID       string
    Duration time.Duration
    From     interface{}
    To       interface{}
    Easing   EasingFunction
    TargetID string
}

// NewTransition 创建过渡动画
func NewTransition(config TransitionConfig) *Transition {
    if config.ID == "" {
        config.ID = generateID()
    }
    if config.Duration == 0 {
        config.Duration = 300 * time.Millisecond
    }
    if config.Easing == nil {
        config.Easing = EaseInOutQuad
    }

    return &Transition{
        id:       config.ID,
        duration: config.Duration,
        from:     config.From,
        to:       config.To,
        current:  config.From,
        easing:   config.Easing,
        targetID: config.TargetID,
    }
}

// ID 实现 Animation 接口
func (t *Transition) ID() string {
    return t.id
}

// Update 实现 Animation 接口
func (t *Transition) Update(dt time.Duration) (bool, float64) {
    if t.done {
        return false, 1.0
    }

    t.elapsed += dt
    if t.elapsed >= t.duration {
        t.elapsed = t.duration
        t.current = t.to
        t.done = true
        return false, 1.0
    }

    progress := float64(t.elapsed) / float64(t.duration)
    eased := t.easing(progress)

    // 插值计算
    t.current = t.interpolate(t.from, t.to, eased)

    return true, progress
}

// Apply 实现 Animation 接口
func (t *Transition) Apply(ctx *Context) {
    ctx.Progress = float64(t.elapsed) / float64(t.duration)
    ctx.Easing = t.easing
    ctx.TargetID = t.targetID
}

// Duration 实现 Animation 接口
func (t *Transition) Duration() time.Duration {
    return t.duration
}

// IsDone 实现 Animation 接口
func (t *Transition) IsDone() bool {
    return t.done
}

// Current 获取当前值
func (t *Transition) Current() interface{} {
    return t.current
}

// interpolate 插值
func (t *Transition) interpolate(from, to interface{}, progress float64) interface{} {
    switch v := from.(type) {
    case int:
        toInt := to.(int)
        return int(float64(v) + float64(toInt-v)*progress)
    case float64:
        toFloat := to.(float64)
        return v + (toFloat-v)*progress
    case string:
        // 字符串渐变（逐字符）
        return t.interpolateString(v, to.(string), progress)
    default:
        return to
    }
}

// interpolateString 字符串插值
func (t *Transition) interpolateString(from, to string, progress float64) string {
    // 计算应该显示多少个字符
    fromLen := len(from)
    toLen := len(to)
    targetLen := int(float64(fromLen) + float64(toLen-fromLen)*progress)

    if targetLen <= toLen {
        return to[:targetLen]
    }
    return to
}

// generateID 生成唯一 ID
func generateID() string {
    return fmt.Sprintf("anim-%d", time.Now().UnixNano())
}
```

### 2. Loop 动画

```go
// 位于: tui/runtime/animation/loop.go

package animation

// Loop 循环动画
type Loop struct {
    id         string
    duration   time.Duration
    elapsed    time.Duration
    iterations int
    maxIter    int
    done       bool
}

// LoopConfig 循环动画配置
type LoopConfig struct {
    ID        string
    Duration  time.Duration
    Iterations int // 0 = 无限循环
}

// NewLoop 创建循环动画
func NewLoop(config LoopConfig) *Loop {
    if config.ID == "" {
        config.ID = generateID()
    }
    if config.Duration == 0 {
        config.Duration = 1000 * time.Millisecond
    }

    return &Loop{
        id:       config.ID,
        duration: config.Duration,
        maxIter:  config.Iterations,
    }
}

// ID 实现 Animation 接口
func (l *Loop) ID() string {
    return l.id
}

// Update 实现 Animation 接口
func (l *Loop) Update(dt time.Duration) (bool, float64) {
    l.elapsed += dt

    // 计算当前循环次数
    cycle := int(l.elapsed / l.duration)

    if l.maxIter > 0 && cycle >= l.maxIter {
        l.done = true
        return false, 1.0
    }

    // 计算当前循环内的进度
    progress := float64(l.elapsed%l.duration) / float64(l.duration)

    return true, progress
}

// Apply 实现 Animation 接口
func (l *Loop) Apply(ctx *Context) {
    ctx.Progress = float64(l.elapsed%l.duration) / float64(l.duration)
    ctx.Elapsed = l.elapsed
}

// Duration 实现 Animation 接口
func (l *Loop) Duration() time.Duration {
    if l.maxIter > 0 {
        return l.duration * time.Duration(l.maxIter)
    }
    return l.duration
}

// IsDone 实现 Animation 接口
func (l *Loop) IsDone() bool {
    return l.done
}

// GetCycle 获取当前循环次数
func (l *Loop) GetCycle() int {
    return int(l.elapsed / l.duration)
}
```

### 3. Sequence 动画

```go
// 位于: tui/runtime/animation/sequence.go

package animation

// Sequence 串行动画
type Sequence struct {
    id         string
    animations []Animation
    current    int
    done       bool
}

// SequenceConfig 串行动画配置
type SequenceConfig struct {
    ID         string
    Animations []Animation
}

// NewSequence 创建串行动画
func NewSequence(config SequenceConfig) *Sequence {
    if config.ID == "" {
        config.ID = generateID()
    }

    return &Sequence{
        id:         config.ID,
        animations: config.Animations,
        current:    0,
    }
}

// ID 实现 Animation 接口
func (s *Sequence) ID() string {
    return s.id
}

// Update 实现 Animation 接口
func (s *Sequence) Update(dt time.Duration) (bool, float64) {
    if s.done || s.current >= len(s.animations) {
        s.done = true
        return false, 1.0
    }

    // 更新当前动画
    running, _ := s.animations[s.current].Update(dt)

    if !running {
        // 当前动画完成，移动到下一个
        s.current++
        if s.current >= len(s.animations) {
            s.done = true
            return false, 1.0
        }
    }

    // 计算总体进度
    progress := float64(s.current) / float64(len(s.animations))
    return true, progress
}

// Apply 实现 Animation 接口
func (s *Sequence) Apply(ctx *Context) {
    if s.current < len(s.animations) {
        s.animations[s.current].Apply(ctx)
    }
}

// Duration 实现 Animation 接口
func (s *Sequence) Duration() time.Duration {
    var total time.Duration
    for _, anim := range s.animations {
        total += anim.Duration()
    }
    return total
}

// IsDone 实现 Animation 接口
func (s *Sequence) IsDone() bool {
    return s.done
}
```

### 4. Parallel 动画

```go
// 位于: tui/runtime/animation/parallel.go

package animation

// Parallel 并行动画
type Parallel struct {
    id         string
    animations []Animation
    done       bool
}

// ParallelConfig 并行动画配置
type ParallelConfig struct {
    ID         string
    Animations []Animation
}

// NewParallel 创建并行动画
func NewParallel(config ParallelConfig) *Parallel {
    if config.ID == "" {
        config.ID = generateID()
    }

    return &Parallel{
        id:         config.ID,
        animations: config.Animations,
    }
}

// ID 实现 Animation 接口
func (s *Parallel) ID() string {
    return s.id
}

// Update 实现 Animation 接口
func (s *Parallel) Update(dt time.Duration) (bool, float64) {
    if s.done {
        return false, 1.0
    }

    allDone := true
    totalProgress := 0.0

    for _, anim := range s.animations {
        running, progress := anim.Update(dt)
        if running {
            allDone = false
        }
        totalProgress += progress
    }

    if allDone {
        s.done = true
        return false, 1.0
    }

    // 平均进度
    progress := totalProgress / float64(len(s.animations))
    return true, progress
}

// Apply 实现 Animation 接口
func (s *Parallel) Apply(ctx *Context) {
    for _, anim := range s.animations {
        anim.Apply(ctx)
    }
}

// Duration 实现 Animation 接口
func (s *Parallel) Duration() time.Duration {
    var max time.Duration
    for _, anim := range s.animations {
        if d := anim.Duration(); d > max {
            max = d
        }
    }
    return max
}

// IsDone 实现 Animation 接口
func (s *Parallel) IsDone() bool {
    return s.done
}
```

## 组件集成

### TextInput 光标闪烁

```go
// 位于: tui/framework/component/textinput.go

package component

// TextInput 文本输入组件
type TextInput struct {
    BaseComponent

    // 光标闪烁动画
    cursorAnimation *animation.Loop
}

// Mount 挂载组件
func (t *TextInput) Mount(rt Runtime) error {
    // 创建光标闪烁动画
    t.cursorAnimation = animation.NewLoop(animation.LoopConfig{
        ID:       t.ID() + "-cursor",
        Duration: 500 * time.Millisecond,
    })

    // 订阅动画更新
    t.cursorAnimation.Subscribe(func(anim animation.Animation, progress float64) {
        t.MarkDirty()
    })

    // 添加到动画管理器
    rt.AnimationManager().Add(t.cursorAnimation)

    return nil
}

// Unmount 卸载组件
func (t *TextInput) Unmount(rt Runtime) error {
    // 移除动画
    rt.AnimationManager().Remove(t.cursorAnimation.ID())
    return nil
}

// Paint 绘制组件
func (t *TextInput) Paint(ctx PaintContext, buf *runtime.CellBuffer) {
    // 获取光标闪烁状态
    cursorVisible := t.shouldShowCursor()

    // 绘制光标
    if cursorVisible && t.HasFocus() {
        buf.SetCell(t.x+t.cursor, t.y, ' ', cursorStyle)
    }
}

// shouldShowCursor 是否显示光标
func (t *TextInput) shouldShowCursor() bool {
    if t.cursorAnimation == nil {
        return true
    }
    // 使用动画进度决定是否显示（闪烁效果）
    cycle := t.cursorAnimation.GetCycle()
    return cycle%2 == 0
}
```

### ProgressBar 进度条动画

```go
// 位于: tui/framework/component/progressbar.go

package component

// ProgressBar 进度条组件
type ProgressBar struct {
    BaseComponent
    *StateHolder

    // 进度动画
    progressAnimation *animation.Transition
}

// SetProgress 设置进度
func (p *ProgressBar) SetProgress(value float64) {
    if value < 0 {
        value = 0
    }
    if value > 1 {
        value = 1
    }

    // 创建过渡动画
    oldProgress := p.GetProgress()
    p.progressAnimation = animation.NewTransition(animation.TransitionConfig{
        ID:       p.ID() + "-progress",
        Duration: 300 * time.Millisecond,
        From:     oldProgress,
        To:       value,
        Easing:   animation.EaseOutQuad,
        TargetID: p.ID(),
    })

    // 添加到动画管理器
    p.rt.AnimationManager().Add(p.progressAnimation)
}

// GetProgress 获取当前进度
func (p *ProgressBar) GetProgress() float64 {
    if p.progressAnimation != nil && !p.progressAnimation.IsDone() {
        if val, ok := p.progressAnimation.Current().(float64); ok {
            return val
        }
    }
    if val, ok := p.GetStateValue("progress").(float64); ok {
        return val
    }
    return 0
}

// Paint 绘制组件
func (p *ProgressBar) Paint(ctx PaintContext, buf *runtime.CellBuffer) {
    progress := p.GetProgress()
    width := int(float64(p.width) * progress)

    // 绘制进度条
    for i := 0; i < width; i++ {
        buf.SetCell(p.x+i, p.y, ' ', progressStyle)
    }
}
```

## 与 Runtime 集成

```go
// 位于: tui/runtime/runtime.go

package runtime

type Runtime struct {
    // ...
    animationManager *animation.Manager
}

// Run 主循环
func (r *Runtime) Run() error {
    for {
        select {
        case <-r.animationManager.Tick():
            // 动画更新
            r.animationManager.Update()

            // 标记需要重新渲染
            r.dirty.MarkAll()

            // 渲染
            r.Render()

        case event := <-r.Events():
            // 处理事件
            r.HandleEvent(event)
        }
    }
}
```

## 测试

```go
// 位于: tui/runtime/animation/transition_test.go

package animation

func TestTransition(t *testing.T) {
    trans := NewTransition(TransitionConfig{
        Duration: 100 * time.Millisecond,
        From:     0,
        To:       100,
    })

    // 初始状态
    assert.Equal(t, 0, trans.Current())
    assert.True(t, trans.IsRunning())

    // 50% 进度
    running, _ := trans.Update(50 * time.Millisecond)
    assert.True(t, running)
    assert.Equal(t, 50, trans.Current())

    // 完成
    running, progress := trans.Update(50 * time.Millisecond)
    assert.False(t, running)
    assert.Equal(t, 1.0, progress)
    assert.Equal(t, 100, trans.Current())
    assert.True(t, trans.IsDone())
}

func TestLoop(t *testing.T) {
    loop := NewLoop(LoopConfig{
        Duration: 100 * time.Millisecond,
    })

    // 第一个循环
    running, progress := loop.Update(50 * time.Millisecond)
    assert.True(t, running)
    assert.Equal(t, 0.5, progress)
    assert.Equal(t, 0, loop.GetCycle())

    // 第二个循环
    running, progress = loop.Update(100 * time.Millisecond)
    assert.True(t, running)
    assert.Equal(t, 0.5, progress)
    assert.Equal(t, 1, loop.GetCycle())
}

func TestManager(t *testing.T) {
    mgr := NewManager()

    trans := NewTransition(TransitionConfig{
        Duration: 100 * time.Millisecond,
        From:     0,
        To:       100,
    })

    mgr.Add(trans)

    assert.True(t, mgr.HasActive())
    assert.True(t, mgr.running)

    // 更新到完成
    for i := 0; i < 10; i++ {
        mgr.Update()
    }

    assert.False(t, mgr.HasActive())
    assert.False(t, mgr.running)
}
```

## 总结

动画系统提供：

1. **按需 Tick**: 只在有活动动画时运行定时器
2. **精确控制**: 支持暂停、恢复、取消
3. **可组合**: 支持串行和并行动画
4. **内置缓动**: 提供常用缓动函数
5. **节能**: 无活动动画时停止 tick
6. **易集成**: 组件可以轻松添加动画

## 相关文档

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 架构概览
- [RENDERING.md](./RENDERING.md) - 渲染系统
- [COMPONENTS.md](./COMPONENTS.md) - 组件系统
- [STATE_MANAGEMENT.md](./STATE_MANAGEMENT.md) - 状态管理
