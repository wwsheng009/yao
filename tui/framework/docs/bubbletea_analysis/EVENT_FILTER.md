# 事件过滤拦截器增强方案

## 概述

事件过滤拦截器（Event Filter Interceptor）借鉴自 Bubble Tea 的 `WithFilter` 功能，允许在事件处理流程中插入拦截器，用于日志记录、权限检查、消息转换等场景。

## 当前问题

### 缺少统一的事件拦截机制

目前的 Yao TUI 事件系统直接从输入源分发到组件，缺少中间拦截层：

```
当前流程:
Input → Dispatch → Capture → Target → Bubble

理想流程:
Input → Filter(s) → Dispatch → Capture → Target → Bubble
              ↑
         可插入拦截器
```

### 具体痛点

1. **调试困难** - 无法统一追踪所有事件
2. **权限控制** - 无法在事件到达组件前进行权限验证
3. **事件转换** - 无法统一修改事件内容
4. **审计日志** - 无法记录所有事件流

## 设计方案

### 核心接口

```go
// tui/runtime/event/filter.go

package event

import (
    "context"
    "fmt"
    "time"
)

// Filter 事件过滤器接口
// 返回 (event, proceed):
//   - event: 过滤后的事件（可能被修改）
//   - proceed: 是否继续传播
type Filter interface {
    Filter(ctx *Context, event Event) (Event, bool)
}

// FilterFunc 函数式过滤器
type FilterFunc func(ctx *Context, event Event) (Event, bool)

func (f FilterFunc) Filter(ctx *Context, event Event) (Event, bool) {
    return f(ctx, event)
}

// FilterChain 过滤器链
type FilterChain struct {
    filters []Filter
}

func NewFilterChain(filters ...Filter) *FilterChain {
    return &FilterChain{filters: filters}
}

// Add 添加过滤器
func (c *FilterChain) Add(filter Filter) {
    c.filters = append(c.filters, filter)
}

// Process 处理事件
func (c *FilterChain) Process(ctx *Context, event Event) (Event, bool) {
    current := event
    for _, filter := range c.filters {
        e, proceed := filter.Filter(ctx, current)
        if !proceed {
            return e, false // 被拦截
        }
        current = e
    }
    return current, true
}
```

### 内置过滤器

```go
// tui/runtime/event/filter/builtin.go

package filter

import (
    "context"
    "log"
    "os"
    "time"
)

// LoggingFilter 日志过滤器
type LoggingFilter struct {
    logger *log.Logger
    events []EventType // 只记录特定类型
}

func NewLoggingFilter(writer io.Writer) *LoggingFilter {
    return &LoggingFilter{
        logger: log.New(writer, "[EVENT] ", log.LstdFlags|log.Lmicroseconds),
    }
}

func (f *LoggingFilter) Filter(ctx *Context, event Event) (Event, bool) {
    f.logger.Printf("%s %+v", event.Type(), event)
    return event, true
}

// MetricsFilter 指标收集过滤器
type MetricsFilter struct {
    counts map[EventType]int
    times  map[EventType]time.Duration
}

func NewMetricsFilter() *MetricsFilter {
    return &MetricsFilter{
        counts: make(map[EventType]int),
        times:  make(map[EventType]time.Duration),
    }
}

func (f *MetricsFilter) Filter(ctx *Context, event Event) (Event, bool) {
    start := time.Now()
    defer func() {
        f.times[event.Type()] += time.Since(start)
        f.counts[event.Type()]++
    }()
    return event, true
}

// RateLimitFilter 频率限制过滤器
type RateLimitFilter struct {
    limits map[EventType]RateLimit
}

type RateLimit struct {
    Interval time.Duration
    MaxCount int
    current  int
    lastTime time.Time
}

func (f *RateLimitFilter) Filter(ctx *Context, event Event) (Event, bool) {
    limit, exists := f.limits[event.Type()]
    if !exists {
        return event, true
    }

    now := time.Now()
    if now.Sub(limit.lastTime) >= limit.Interval {
        limit.current = 0
        limit.lastTime = now
    }

    if limit.current >= limit.MaxCount {
        return event, false // 超出限制，拦截
    }

    limit.current++
    return event, true
}

// TransformFilter 事件转换过滤器
type TransformFilter struct {
    transformer func(Event) Event
}

func NewTransformFilter(fn func(Event) Event) *TransformFilter {
    return &TransformFilter{transformer: fn}
}

func (f *TransformFilter) Filter(ctx *Context, event Event) (Event, bool) {
    return f.transformer(event), true
}

// PermissionFilter 权限控制过滤器
type PermissionFilter struct {
    checker func(ctx *Context, event Event) bool
}

func (f *PermissionFilter) Filter(ctx *Context, event Event) (Event, bool) {
    if !f.checker(ctx, event) {
        return event, false // 无权限，拦截
    }
    return event, true
}
```

### 集成到 Dispatcher

```go
// tui/runtime/event/dispatch.go

type Dispatcher struct {
    // 现有字段...
    filterChain *FilterChain
}

func (d *Dispatcher) AddFilter(filter Filter) {
    if d.filterChain == nil {
        d.filterChain = NewFilterChain()
    }
    d.filterChain.Add(filter)
}

func (d *Dispatcher) Dispatch(ctx *Context, event Event) {
    // 应用过滤器
    if d.filterChain != nil {
        filtered, proceed := d.filterChain.Process(ctx, event)
        if !proceed {
            return // 事件被拦截
        }
        event = filtered
    }

    // 继续原有逻辑
    d.dispatch(ctx, event)
}
```

## 使用示例

### 1. 调试日志

```go
package main

import (
    "os"
    "github.com/yaoapp/yao/tui/runtime/event"
)

func main() {
    app := NewApp()

    // 添加日志过滤器
    app.Dispatcher.AddFilter(filter.NewLoggingFilter(os.Stderr))

    app.Run()
}
```

### 2. 事件转换

```go
// 将 Ctrl+Q 转换为退出事件
transform := filter.NewTransformFilter(func(e event.Event) event.Event {
    if ke, ok := e.(*event.KeyEvent); ok {
        if ke.Ctrl && ke.Key == 'q' {
            return &event.QuitEvent{}
        }
    }
    return e
})

app.Dispatcher.AddFilter(transform)
```

### 3. 权限控制

```go
// 某些操作需要管理员权限
permFilter := &filter.PermissionFilter{
    checker: func(ctx *event.Context, e event.Event) bool {
        if _, ok := e.(*event.AdminActionEvent); ok {
            return ctx.User().IsAdmin()
        }
        return true
    },
}

app.Dispatcher.AddFilter(permFilter)
```

### 4. 频率限制

```go
// 限制某些高频事件
rateLimit := &filter.RateLimitFilter{
    limits: map[event.EventType]filter.RateLimit{
        event.TypeMouseMove: {
            Interval: time.Second,
            MaxCount: 60, // 每秒最多 60 个鼠标移动事件
        },
    },
}

app.Dispatcher.AddFilter(rateLimit)
```

### 5. 组合使用

```go
// 创建过滤器链
chain := event.NewFilterChain(
    filter.NewMetricsFilter(),
    filter.NewLoggingFilter(os.Stderr),
    &filter.PermissionFilter{checker: permissionChecker},
)

app.Dispatcher.SetFilterChain(chain)
```

## 实施计划

### Phase 1: 核心接口 (Week 1)

- [ ] 实现 `Filter` 接口
- [ ] 实现 `FilterChain`
- [ ] 集成到 `Dispatcher`
- [ ] 单元测试

### Phase 2: 内置过滤器 (Week 2)

- [ ] `LoggingFilter`
- [ ] `MetricsFilter`
- [ ] `TransformFilter`
- [ ] 集成测试

### Phase 3: 高级过滤器 (Week 3)

- [ ] `RateLimitFilter`
- [ ] `PermissionFilter`
- [ ] 性能测试

### Phase 4: 文档和示例 (Week 4)

- [ ] API 文档
- [ ] 使用示例
- [ ] 最佳实践

## 文件结构

```
tui/runtime/event/
├── filter.go              # 核心接口
├── filter/
│   ├── builtin.go         # 内置过滤器
│   ├── logging.go         # 日志过滤器
│   ├── metrics.go         # 指标过滤器
│   ├── transform.go       # 转换过滤器
│   ├── ratelimit.go       # 频率限制过滤器
│   └── permission.go      # 权限过滤器
└── filter_test.go         # 测试
```

## 测试策略

```go
func TestFilterChain(t *testing.T) {
    chain := NewFilterChain()

    // 测试拦截
    chain.Add(FilterFunc(func(ctx *Context, e Event) (Event, bool) {
        return e, false
    }))

    _, proceed := chain.Process(nil, &KeyEvent{})
    assert.False(t, proceed)
}

func TestLoggingFilter(t *testing.T) {
    buf := &bytes.Buffer{}
    filter := NewLoggingFilter(buf)

    ctx := &Context{}
    event := &KeyEvent{Key: 'a'}

    filtered, proceed := filter.Filter(ctx, event)

    assert.True(t, proceed)
    assert.Contains(t, buf.String(), "KeyEvent")
}
```

## 性能考虑

1. **惰性应用** - 只在有过滤器时才执行
2. **快速路径** - 空过滤器链直接跳过
3. **性能监控** - 使用 MetricsFilter 监控过滤器开销

## 与 Bubble Tea 的差异

| 特性 | Bubble Tea | Yao TUI Event Filter |
|------|-----------|----------------------|
| **拦截点** | Update 前 | Dispatch 前 |
| **多过滤器** | 单一 Filter | FilterChain |
| **内置过滤器** | 无 | Logging/Metrics 等 |
| **事件修改** | 是 | 是 |
| **停止传播** | 返回 nil | return false |
