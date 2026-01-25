# 上下文取消支持增强方案

## 概述

上下文取消支持借鉴自 Bubble Tea 的 `WithContext` 功能，允许通过 `context.Context` 优雅地控制应用生命周期，实现超时控制、优雅关闭等功能。

## 当前问题

### 缺少统一的上下文管理

Yao TUI 目前缺少对 Go 标准 `context.Context` 的原生支持，导致：

1. **无法优雅关闭** - 终端信号处理不够完善
2. **无法超时控制** - 长时间运行的操作无法主动取消
3. **无法级联取消** - 组件无法感知应用关闭状态
4. **无法传递元数据** - 请求追踪、用户上下文等无法传递

### 具体痛点

```go
// 当前问题示例：无法取消的异步操作
func (c *Component) LoadData() {
    go func() {
        // 即使应用关闭，这个 goroutine 仍会继续
        data := fetchFromAPI() // 可能会 hang 很久
        c.Dispatch(NewDataLoadedAction(data))
    }()
}
```

## 设计方案

### 核心接口

```go
// tui/runtime/context.go

package runtime

import (
    "context"
    "sync"
    "time"
)

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

// App 应用主结构
type App struct {
    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup

    // 现有字段...
}

// NewContext 创建带上下文的应用
func NewApp(parentCtx context.Context) *App {
    ctx, cancel := context.WithCancel(parentCtx)
    return &App{
        ctx:    ctx,
        cancel: cancel,
    }
}

// Context 返回应用上下文
func (a *App) Context() context.Context {
    return a.ctx
}

// WithContext 设置父上下文
func (a *App) WithContext(parent context.Context) {
    a.cancel() // 取消旧的上下文
    a.ctx, a.cancel = context.WithCancel(parent)
}

// Shutdown 优雅关闭应用
func (a *App) Shutdown(timeout ...time.Duration) error {
    a.cancel() // 取消上下文

    if len(timeout) > 0 {
        done := make(chan struct{})
        go func() {
            a.wg.Wait()
            close(done)
        }()

        select {
        case <-done:
            return nil
        case <-time.After(timeout[0]):
            return ErrShutdownTimeout
        }
    } else {
        a.wg.Wait()
    }

    return nil
}

// Go 在应用上下文中启动 goroutine
func (a *App) Go(f func(ctx context.Context) error) {
    a.wg.Add(1)
    go func() {
        defer a.wg.Done()
        if err := f(a.ctx); err != nil && a.ctx.Err() == nil {
            // 只有非取消错误才记录
            a.handleError(err)
        }
    }()
}

// Done 返回取消通道
func (a *App) Done() <-chan struct{} {
    return a.ctx.Done()
}

// Err 返回取消原因
func (a *App) Err() error {
    return a.ctx.Err()
}
```

### 组件级上下文支持

```go
// tui/runtime/component/context.go

package component

import (
    "context"
)

// ContextAware 支持上下文的组件接口
type ContextAware interface {
    Component
    WithContext(ctx context.Context)
    Context() context.Context
}

// BaseComponent 带上下文的基础组件
type BaseComponent struct {
    ctx context.Context
    // 现有字段...
}

func NewBaseComponent() *BaseComponent {
    return &BaseComponent{
        ctx: context.Background(),
    }
}

func (c *BaseComponent) WithContext(ctx context.Context) {
    c.ctx = ctx
}

func (c *BaseComponent) Context() context.Context {
    return c.ctx
}

// ContextValue 获取上下文值
func (c *BaseComponent) ContextValue(key interface{}) interface{} {
    return c.ctx.Value(key)
}
```

### Action 上下文支持

```go
// tui/runtime/action/context.go

package action

import (
    "context"
)

// ActionContext 执行上下文
type ActionContext struct {
    context.Context
    app *runtime.App
}

// NewActionContext 创建执行上下文
func NewActionContext(ctx context.Context, app *runtime.App) *ActionContext {
    return &ActionContext{
        Context: ctx,
        app:     app,
    }
}

// App 返回应用实例
func (c *ActionContext) App() *runtime.App {
    return c.app
}

// Action 带上下文的 Action 接口
type Action interface {
    Execute(ctx *ActionContext) ActionResult
}

// ActionFunc 函数式 Action
type ActionFunc func(ctx *ActionContext) ActionResult

func (f ActionFunc) Execute(ctx *ActionContext) ActionResult {
    return f(ctx)
}
```

### 信号处理

```go
// tui/runtime/signal.go

package runtime

import (
    "context"
    "os"
    "os/signal"
    "syscall"
)

// SignalHandler 信号处理器
type SignalHandler struct {
    ctx    context.Context
    cancel context.CancelFunc
}

// NewSignalHandler 创建信号处理器
func NewSignalHandler(ctx context.Context, cancel context.CancelFunc) *SignalHandler {
    return &SignalHandler{
        ctx:    ctx,
        cancel: cancel,
    }
}

// Handle 启动信号处理
func (h *SignalHandler) Handle() {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan,
        syscall.SIGINT,  // Ctrl+C
        syscall.SIGTERM, // kill
        syscall.SIGHUP,  // 终端关闭
    )

    go func() {
        for {
            select {
            case sig := <-sigChan:
                h.handleSignal(sig)
            case <-h.ctx.Done():
                signal.Stop(sigChan)
                return
            }
        }
    }()
}

func (h *SignalHandler) handleSignal(sig os.Signal) {
    switch sig {
    case syscall.SIGINT, syscall.SIGTERM:
        // 优雅关闭
        h.cancel()
    case syscall.SIGHUP:
        // 可选：重新加载配置
        h.cancel()
    }
}
```

## 使用示例

### 1. 基础使用

```go
package main

func main() {
    app := NewApp(context.Background())

    // 启动应用
    app.Run()
}

func (a *App) Run() error {
    // 创建信号处理器
    signalHandler := NewSignalHandler(a.ctx, a.cancel)
    signalHandler.Handle()

    // 主循环
    for {
        select {
        case event := <-a.events:
            a.Dispatch(event)
        case <-a.ctx.Done():
            a.Cleanup()
            return a.ctx.Err() // 退出原因
        }
    }
}
```

### 2. 可取消的异步操作

```go
func (c *Component) LoadData() {
    c.app.Go(func(ctx context.Context) error {
        // 可以被取消的 HTTP 请求
        req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            if ctx.Err() == context.Canceled {
                return nil // 正常取消，不报错
            }
            return err
        }
        defer resp.Body.Close()

        data := parseResponse(resp)
        c.Dispatch(NewDataLoadedAction(data))
        return nil
    })
}
```

### 3. 超时控制

```go
func (c *Component) FetchWithTimeout() {
    ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
    defer cancel()

    c.app.Go(func(ctx context.Context) error {
        req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
        resp, err := http.DefaultClient.Do(req)
        // 处理响应...
        return err
    })
}
```

### 4. 优雅关闭

```go
func (a *App) Shutdown() {
    // 取消所有操作
    a.cancel()

    // 等待所有 goroutine 完成（最多 5 秒）
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    done := make(chan struct{})
    go func() {
        a.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        log.Println("所有操作已完成")
    case <-ctx.Done():
        log.Println("关闭超时，强制退出")
    }

    // 恢复终端状态
    a.restoreTerminal()
}
```

### 5. 传递元数据

```go
func main() {
    // 创建带元数据的上下文
    ctx := context.WithValue(context.Background(), KeyRequestID, uuid.New())
    ctx = context.WithValue(ctx, KeyUser, &User{ID: 123, Name: "Alice"})

    app := NewApp(ctx)
    app.Run()
}

// 在组件中访问
func (c *Component) GetUser() *User {
    if user, ok := c.Context().Value(KeyUser).(*User); ok {
        return user
    }
    return nil
}
```

## 实施计划

### Phase 1: 核心上下文 (Week 1)

- [ ] 实现 `App.Context()` 和 `App.WithContext()`
- [ ] 实现 `App.Go()` goroutine 管理
- [ ] 实现 `App.Shutdown()`
- [ ] 单元测试

### Phase 2: 组件支持 (Week 1)

- [ ] 实现 `ContextAware` 接口
- [ ] 更新 `BaseComponent`
- [ ] 组件上下文传递

### Phase 3: 信号处理 (Week 2)

- [ ] 实现 `SignalHandler`
- [ ] 集成到 `App.Run()`
- [ ] 测试各种信号场景

### Phase 4: Action 集成 (Week 2)

- [ ] 实现 `ActionContext`
- [ ] 更新 Action 接口
- [ ] 确保向后兼容

### Phase 5: 文档和示例 (Week 3)

- [ ] API 文档
- [ ] 使用示例
- [ ] 迁移指南

## 文件结构

```
tui/runtime/
├── context.go              # 核心上下文支持
├── signal.go               # 信号处理
├── app.go                  # App 上下文方法
├── component/
│   └── context.go          # 组件上下文支持
├── action/
│   └── context.go          # Action 上下文支持
└── context_test.go         # 测试
```

## 测试策略

```go
func TestAppCancellation(t *testing.T) {
    app := NewApp(context.Background())

    ran := false
    app.Go(func(ctx context.Context) error {
        time.Sleep(100 * time.Millisecond)
        ran = true
        return nil
    })

    app.Shutdown()
    assert.True(t, ran, "goroutine should complete")
}

func TestAppTimeout(t *testing.T) {
    app := NewApp(context.Background())

    app.Go(func(ctx context.Context) error {
        select {
        case <-time.After(10 * time.Second):
            return nil
        case <-ctx.Done():
            return ctx.Err()
        }
    })

    app.cancel()
    err := app.Shutdown(100 * time.Millisecond)
    assert.Equal(t, context.Canceled, err)
}

func TestComponentContext(t *testing.T) {
    ctx := context.WithValue(context.Background(), "key", "value")
    comp := NewBaseComponent()
    comp.WithContext(ctx)

    assert.Equal(t, "value", comp.ContextValue("key"))
}
```

## 性能考虑

1. **轻量级上下文** - Go 的 context.Context 非常轻量
2. **选择性传递** - 只在需要时传递上下文
3. **避免泄漏** - 通过 `App.Go()` 确保所有 goroutine 正确退出

## 向后兼容

```go
// 向后兼容：没有上下文的旧代码仍可工作
type OldAction struct {
    // ...
}

func (a *OldAction) Execute(ctx *ActionContext) ActionResult {
    // 如果不需要上下文，忽略即可
    return a.DoWork()
}
```

## 与 Bubble Tea 的差异

| 特性 | Bubble Tea | Yao TUI Context |
|------|-----------|-----------------|
| **上下文传递** | Program 级别 | App + Component + Action 级别 |
| **Goroutine 管理** | Cmd 模式 | App.Go() 方法 |
| **信号处理** | 内置 | 可选 SignalHandler |
| **关闭超时** | 无 | 支持 |
| **元数据传递** | context.Value | context.Value + 便捷方法 |
