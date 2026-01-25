# 恐慌恢复增强方案

## 概述

恐慌恢复功能借鉴自 Bubble Tea 的 panic 捕获和终端状态恢复机制，确保应用在发生不可恢复错误时能够正确清理资源并恢复终端状态。

## 当前问题

### 恐慌处理不完善

当前 Yao TUI 在发生 panic 时：

1. **终端状态未恢复** - 终端可能保持非正常状态（无回显、光标隐藏等）
2. **资源泄漏** - 文件句柄、goroutine 等资源未正确释放
3. **错误信息丢失** - panic 堆栈可能未正确记录
4. **调试困难** - 难以定位 panic 发生的位置

### 具体场景

```go
// 问题场景：panic 导致终端状态异常
func (c *Component) Handle(event Event) {
    // 如果这里发生 panic...
    data := make([]byte, -1) // panic: negative size
    // 终端可能无法恢复正常状态
}
```

## 设计方案

### 核心接口

```go
// tui/runtime/recovery.go

package runtime

import (
    "fmt"
    "io"
    "os"
    "runtime/debug"
    "sync"
)

// PanicHandler panic 处理器接口
type PanicHandler interface {
    HandlePanic(r interface{}, stack []byte)
}

// Recovery 恢复管理器
type Recovery struct {
    mu           sync.RWMutex
    handlers     []PanicHandler
    terminal     Terminal
    panicLogFile *os.File
}

// NewRecovery 创建恢复管理器
func NewRecovery(terminal Terminal) *Recovery {
    return &Recovery{
        terminal: terminal,
        handlers: make([]PanicHandler, 0),
    }
}

// AddHandler 添加 panic 处理器
func (r *Recovery) AddHandler(h PanicHandler) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.handlers = append(r.handlers, h)
}

// Handle 处理 panic
func (r *Recovery) Handle(panicValue interface{}) {
    stack := debug.Stack()

    // 1. 恢复终端状态
    r.restoreTerminal()

    // 2. 记录 panic
    r.logPanic(panicValue, stack)

    // 3. 调用处理器
    r.mu.RLock()
    for _, h := range r.handlers {
        h.HandlePanic(panicValue, stack)
    }
    r.mu.RUnlock()
}

// restoreTerminal 恢复终端状态
func (r *Recovery) restoreTerminal() {
    if r.terminal != nil {
        // 恢复正常模式
        r.terminal.SetNormalMode()

        // 显示光标
        r.terminal.ShowCursor()

        // 清除备用屏幕缓冲区
        r.terminal.ExitAltScreen()

        // 启用回显
        r.terminal.EnableEcho()

        // 刷新输出
        r.terminal.Flush()
    }
}

// logPanic 记录 panic
func (r *Recovery) logPanic(panicValue interface{}, stack []byte) {
    msg := fmt.Sprintf("\n\n=== PANIC ===\nValue: %v\n\nStack:\n%s\n\n",
        panicValue, stack)

    // 输出到 stderr
    fmt.Fprint(os.Stderr, msg)

    // 写入日志文件
    if r.panicLogFile != nil {
        r.panicLogFile.WriteString(msg)
    }
}

// EnablePanicLog 启用 panic 日志文件
func (r *Recovery) EnablePanicLog(filename string) error {
    f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        return err
    }
    r.panicLogFile = f
    return nil
}

// Close 关闭恢复管理器
func (r *Recovery) Close() error {
    if r.panicLogFile != nil {
        return r.panicLogFile.Close()
    }
    return nil
}
```

### 内置处理器

```go
// tui/runtime/recovery/handler.go

package recovery

import (
    "fmt"
    "os"
    "runtime"
)

// LoggingHandler 日志处理器
type LoggingHandler struct {
    writer io.Writer
}

func NewLoggingHandler(w io.Writer) *LoggingHandler {
    return &LoggingHandler{writer: w}
}

func (h *LoggingHandler) HandlePanic(r interface{}, stack []byte) {
    fmt.Fprintf(h.writer, "[PANIC] %v\n%s\n", r, stack)
}

// MetricsHandler 指标处理器
type MetricsHandler struct {
    panicCount int
}

func (h *MetricsHandler) HandlePanic(r interface{}, stack []byte) {
    h.panicCount++
}

func (h *MetricsHandler) PanicCount() int {
    return h.panicCount
}

// CrashReportHandler 崩溃报告处理器
type CrashReportHandler struct {
    reportDir string
}

func NewCrashReportHandler(dir string) *CrashReportHandler {
    return &CrashReportHandler{reportDir: dir}
}

func (h *CrashReportHandler) HandlePanic(r interface{}, stack []byte) {
    // 生成崩溃报告文件
    filename := fmt.Sprintf("%s/crash_%d.log", h.reportDir,
        time.Now().Unix())

    f, err := os.Create(filename)
    if err != nil {
        return
    }
    defer f.Close()

    f.WriteString(fmt.Sprintf("Panic: %v\n\n", r))
    f.WriteString("Stack:\n")
    f.Write(stack)
    f.WriteString("\n\nSystem Info:\n")
    f.WriteString(fmt.Sprintf("GOOS: %s\n", runtime.GOOS))
    f.WriteString(fmt.Sprintf("GOARCH: %s\n", runtime.GOARCH))
    f.WriteString(fmt.Sprintf("NumCPU: %d\n", runtime.NumCPU()))
    f.WriteString(fmt.Sprintf("Version: %s\n", runtime.Version()))
}

// NotificationHandler 通知处理器
type NotificationHandler struct {
    notifier func(panicValue interface{}, stack []byte)
}

func NewNotificationHandler(fn func(interface{}, []byte)) *NotificationHandler {
    return &NotificationHandler{notifier: fn}
}

func (h *NotificationHandler) HandlePanic(r interface{}, stack []byte) {
    if h.notifier != nil {
        h.notifier(r, stack)
    }
}
```

### App 集成

```go
// tui/runtime/app.go

package runtime

import (
    "fmt"
)

type App struct {
    recovery *Recovery
    // 现有字段...
}

func (a *App) Run() (err error) {
    // 设置 panic 恢复
    defer func() {
        if r := recover(); r != nil {
            if a.recovery != nil {
                a.recovery.Handle(r)
            }
            err = fmt.Errorf("panic: %v", r)
        }
    }()

    // 正常运行逻辑
    return a.run()
}

func (a *App) SetRecovery(rec *Recovery) {
    a.recovery = rec
}
```

### 组件级保护

```go
// tui/runtime/component/safe.go

package component

import (
    "fmt"
)

// SafeComponent 安全组件包装器
type SafeComponent struct {
    Component
    onPanic func(interface{})
}

func NewSafeComponent(c Component, onPanic func(interface{})) *SafeComponent {
    return &SafeComponent{
        Component: c,
        onPanic:   onPanic,
    }
}

func (s *SafeComponent) Handle(ctx *Context, event Event) bool {
    defer func() {
        if r := recover(); r != nil {
            if s.onPanic != nil {
                s.onPanic(r)
            }
        }
    }()

    return s.Component.Handle(ctx, event)
}

func (s *SafeComponent) Paint(ctx *PaintContext) error {
    defer func() {
        if r := recover(); r != nil {
            if s.onPanic != nil {
                s.onPanic(r)
            }
        }
    }()

    return s.Component.Paint(ctx)
}
```

## 使用示例

### 1. 基础使用

```go
func main() {
    app := NewApp()

    // 设置恢复管理器
    recovery := NewRecovery(app.Terminal())
    app.SetRecovery(recovery)

    app.Run()
}
```

### 2. 添加处理器

```go
recovery := NewRecovery(terminal)

// 添加日志处理器
recovery.AddHandler(NewLoggingHandler(os.Stderr))

// 添加崩溃报告处理器
recovery.AddHandler(NewCrashReportHandler("./crashes"))

// 添加自定义通知
recovery.AddHandler(NewNotificationHandler(func(panicValue interface{}, stack []byte) {
    // 发送通知（如 Sentry、Slack 等）
    sendCrashReport(panicValue, stack)
}))

app.SetRecovery(recovery)
```

### 3. 启用崩溃日志

```go
recovery := NewRecovery(terminal)
recovery.EnablePanicLog("./panics.log")
app.SetRecovery(recovery)
```

### 4. 组件级保护

```go
func NewSafeButton(text string) *Button {
    btn := NewButton(text)

    // 包装为安全组件
    return NewSafeComponent(btn, func(panicValue interface{}) {
        log.Printf("Button panic: %v", panicValue)
        // 可以显示错误提示
        ShowError(fmt.Sprintf("Button error: %v", panicValue))
    })
}
```

### 5. 测试中的 panic 处理

```go
func TestComponent(t *testing.T) {
    comp := NewComponent()

    // 测试组件是否正确处理 panic
    defer func() {
        if r := recover(); r != nil {
            t.Logf("Caught panic: %v", r)
        }
    }()

    comp.Handle(testContext, testEvent)
}
```

## 实施计划

### Phase 1: 核心恢复器 (Week 1)

- [ ] 实现 `Recovery`
- [ ] 实现 `restoreTerminal()`
- [ ] 实现 `logPanic()`
- [ ] 单元测试

### Phase 2: 内置处理器 (Week 1)

- [ ] 实现 `LoggingHandler`
- [ ] 实现 `MetricsHandler`
- [ ] 实现 `CrashReportHandler`
- [ ] 实现 `NotificationHandler`

### Phase 3: App 集成 (Week 2)

- [ ] 集成到 `App.Run()`
- [ ] 实现组件级保护
- [ ] 集成测试

### Phase 4: 文档和示例 (Week 2)

- [ ] API 文档
- [ ] 使用示例
- [ ] 最佳实践

## 文件结构

```
tui/runtime/
├── recovery.go              # 核心恢复管理器
├── recovery/
│   ├── handler.go           # 内置处理器
│   ├── terminal.go          # 终端恢复
│   ├── log.go              # 日志记录
│   └── crash_report.go     # 崩溃报告
├── component/
│   └── safe.go             # 安全组件包装器
└── recovery_test.go        # 测试
```

## 测试策略

```go
func TestRecovery(t *testing.T) {
    terminal := &MockTerminal{}
    recovery := NewRecovery(terminal)

    // 模拟 panic
    recovery.Handle("test panic")

    // 验证终端恢复
    assert.True(t, terminal.Restored)
}

func TestSafeComponent(t *testing.T) {
    panicCalled := false

    base := &MockComponent{
        HandleFunc: func(ctx *Context, event Event) bool {
            panic("test panic")
        },
    }

    safe := NewSafeComponent(base, func(v interface{}) {
        panicCalled = true
        assert.Equal(t, "test panic", v)
    })

    safe.Handle(nil, nil)

    assert.True(t, panicCalled)
}
```

## 最佳实践

1. **总是设置恢复器** - 在生产环境必须启用
2. **记录崩溃信息** - 便于后续分析
3. **保持终端可用** - 恢复后终端应能正常使用
4. **组件级保护** - 对关键组件使用安全包装器
5. **监控 panic** - 使用指标收集器监控 panic 频率

## 与 Bubble Tea 的差异

| 特性 | Bubble Tea | Yao TUI Recovery |
|------|-----------|-----------------|
| **恢复层级** | Program 级 | App + Component 级 |
| **处理器** | 无 | 可扩展 Handler |
| **崩溃报告** | 无 | 内置报告生成 |
| **指标收集** | 无 | 内置指标 |
| **日志文件** | 无 | 支持 |
| **通知机制** | 无 | 可扩展 |
