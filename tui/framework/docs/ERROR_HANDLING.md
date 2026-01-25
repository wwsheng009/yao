# Error Handling System Design (V3)

> **优先级**: P0 (核心功能)
> **目标**: 可靠的错误传播和恢复机制
> **关键特性**: Result 类型、错误边界、异步错误、可恢复性

## 概述

在异步、事件驱动的 TUI 框架中，错误处理是一个复杂但关键的问题。传统的 Go 错误处理方式在回调、异步操作、事件传播等场景下存在不足。

### 为什么需要专门的错误处理系统？

**传统方式的问题**：
```go
// ❌ 传统错误处理：在回调中难以传播
func (b *Button) OnClick(fn func()) {
    b.onClick = func() {
        fn()  // 如果 fn panic 了怎么办？
    }
}

// ❌ 错误丢失
func (b *Button) HandleAction(a *action.Action) bool {
    go b.processAsync()  // 如果 processAsync 出错，错误去哪了？
    return true
}

// ❌ 错误无法恢复
func (app *App) Run() {
    if err := app.runtime.Start(); err != nil {
        log.Fatal(err)  // 只能退出，无法恢复
    }
}
```

**V3 错误处理的优势**：
```go
// ✅ Result 类型：显式错误处理
type Result[T any] struct {
    Value T
    Error error
}

// ✅ 错误边界：捕获并恢复
func (b *Boundary) Catch(err error) RecoveryAction {
    if errors.Is(err, RecoverableError) {
        return RecoveryActionRetry
    }
    return RecoveryActionFallback
}

// ✅ 异步错误：通过事件传递
func (b *Button) processAsync() {
    result := <-b.taskChan
    if result.Error != nil {
        b.dispatcher.Dispatch(action.NewAction(action.ActionError).
            WithPayload(result.Error))
    }
}
```

## 设计目标

1. **类型安全**: Result 类型提供编译时检查
2. **错误传播**: 清晰的错误传播路径
3. **可恢复性**: 支持错误恢复和降级
4. **可观测性**: 完整的错误记录和追踪
5. **无恐慌**: 优雅降级，避免 panic

## 核心类型定义

### 1. Result 类型

```go
// 位于: tui/framework/result/result.go

package result

// Result 结果类型，显式表示可能失败的操作
type Result[T any] struct {
    value T
    err   error
}

// Ok 创建成功结果
func Ok[T any](value T) Result[T] {
    return Result[T]{value: value, err: nil}
}

// Err 创建错误结果
func Err[T any](err error) Result[T] {
    var zero T
    return Result[T]{value: zero, err: err}
}

// IsOk 是否成功
func (r Result[T]) IsOk() bool {
    return r.err == nil
}

// IsErr 是否失败
func (r Result[T]) IsErr() bool {
    return r.err != nil
}

// Unwrap 解包结果，panic 如果是错误
func (r Result[T]) Unwrap() T {
    if r.err != nil {
        panic(r.err)
    }
    return r.value
}

// UnwrapOr 解包结果，失败时返回默认值
func (r Result[T]) UnwrapOr(defaultValue T) T {
    if r.err != nil {
        return defaultValue
    }
    return r.value
}

// UnwrapOrElse 解包结果，失败时调用函数
func (r Result[T]) UnwrapOrElse(fn func(error) T) T {
    if r.err != nil {
        return fn(r.err)
    }
    return r.value
}

// Map 映射成功值
func (r Result[T]) Map(fn func(T) T) Result[T] {
    if r.err != nil {
        return r
    }
    return Ok(fn(r.value))
}

// MapErr 映射错误
func (r Result[T]) MapErr(fn func(error) error) Result[T] {
    if r.err != nil {
        return Err[T](fn(r.err))
    }
    return r
}

// AndThen 链式调用，仅在成功时执行
func (r Result[T]) AndThen(fn func(T) Result[T]) Result[T] {
    if r.err != nil {
        return r
    }
    return fn(r.value)
}

// OrElse 链式调用，仅在失败时执行
func (r Result[T]) OrElse(fn(error) Result[T]) Result[T] {
    if r.err != nil {
        return fn(r.err)
    }
    return r
}

// Error 获取错误
func (r Result[T]) Error() error {
    return r.err
}

// Value 获取值（不 panic）
func (r Result[T]) Value() T {
    return r.value
}
```

### 2. 组件错误处理

```go
// 位于: tui/framework/component/error_handler.go

package component

import "github.com/yaoapp/yao/tui/framework/result"

// ErrorHandler 组件错误处理器
type ErrorHandler struct {
    // 错误回调
    onError []func(error)

    // 错误边界
    boundary *ErrorBoundary

    // 错误状态
    lastError error
    errorCount int
}

// ErrorBoundary 错误边界
type ErrorBoundary struct {
    // 恢复策略
    strategy RecoveryStrategy

    // 降级组件
    fallback Component

    // 重试配置
    maxRetries    int
    retryDelay    time.Duration
    retryBackoff  float64
}

// RecoveryStrategy 恢复策略
type RecoveryStrategy int

const (
    // RecoverPropagate 传播错误（向上层传递）
    RecoverPropagate RecoveryStrategy = iota

    // RecoverRetry 重试操作
    RecoverRetry

    // RecoverFallback 使用降级组件
    RecoverFallback

    // RecoverIgnore 忽略错误
    RecoverIgnore

    // RecoverRestart 重启组件
    RecoverRestart
)

// HandleError 处理错误
func (h *ErrorHandler) HandleError(err error) result.Result[bool] {
    if err == nil {
        return result.Ok(true)
    }

    h.lastError = err
    h.errorCount++

    // 触发错误回调
    for _, cb := range h.onError {
        // 安全调用，防止回调本身 panic
        func() {
            defer func() {
                if r := recover(); r != nil {
                    log.Errorf("Error callback panic: %v", r)
                }
            }()
            cb(err)
        }()
    }

    // 如果有错误边界，使用边界处理
    if h.boundary != nil {
        action := h.boundary.Recover(err)
        switch action {
        case RecoverRetry:
            return result.Err[bool](fmt.Errorf("retry requested"))
        case RecoverFallback:
            return result.Ok(false) // 使用降级
        case RecoverIgnore:
            return result.Ok(true)  // 忽略错误
        case RecoverRestart:
            return result.Err[bool](ErrRestart)
        default:
            return result.Err[bool](err)
        }
    }

    return result.Err[bool](err)
}

// Recover 恢复错误
func (b *ErrorBoundary) Recover(err error) RecoveryStrategy {
    // 检查错误类型
    if IsRecoverable(err) {
        return RecoverRetry
    }

    // 检查是否有降级组件
    if b.fallback != nil {
        return RecoverFallback
    }

    return RecoverPropagate
}

// OnError 注册错误回调
func (h *ErrorHandler) OnError(fn func(error)) {
    h.onError = append(h.onError, fn)
}

// SetBoundary 设置错误边界
func (h *ErrorHandler) SetBoundary(boundary *ErrorBoundary) {
    h.boundary = boundary
}

// GetLastError 获取最后一个错误
func (h *ErrorHandler) GetLastError() error {
    return h.lastError
}

// GetErrorCount 获取错误计数
func (h *ErrorHandler) GetErrorCount() int {
    return h.errorCount
}

// Reset 重置错误状态
func (h *ErrorHandler) Reset() {
    h.lastError = nil
    h.errorCount = 0
}
```

### 3. Action 错误处理

```go
// 位于: tui/framework/action/error_handling.go

package action

import "github.com/yaoapp/yao/tui/framework/result"

// ErrorAction 错误 Action 类型
const (
    ActionError      ActionType = "error"
    ActionPanic      ActionType = "panic"
    ActionTimeout    ActionType = "timeout"
    ActionValidation ActionType = "validation"
)

// ErrorPayload 错误负载
type ErrorPayload struct {
    Error   error
    Source  string      // 错误来源
    Context interface{} // 额外上下文
    Time    time.Time
}

// NewErrorAction 创建错误 Action
func NewErrorAction(err error, source string) *Action {
    return NewAction(ActionError).
        WithPayload(ErrorPayload{
            Error:  err,
            Source: source,
            Time:   time.Now(),
        })
}

// DispatcherWithError 带错误处理的派发器
type DispatcherWithError struct {
    *Dispatcher
    errorHandler ErrorHandlerFunc
}

// ErrorHandlerFunc 错误处理函数
type ErrorHandlerFunc func(a *Action, err error) bool

// DispatchWithError 派发 Action 并处理错误
func (d *DispatcherWithError) DispatchWithError(a *Action) result.Result[bool] {
    handled := d.Dispatcher.Dispatch(a)

    // 检查是否有子 Action 产生的错误
    if errors := a.GetErrors(); len(errors) > 0 {
        for _, err := range errors {
            if d.errorHandler != nil {
                if !d.errorHandler(a, err) {
                    return result.Err[bool](err)
                }
            } else {
                return result.Err[bool](err)
            }
        }
    }

    return result.Ok(handled)
}

// AttachError 附加错误到 Action
func (a *Action) AttachError(err error) {
    a.mutex.Lock()
    defer a.mutex.Unlock()

    if a.errors == nil {
        a.errors = make([]error, 0)
    }
    a.errors = append(a.errors, err)
}

// GetErrors 获取 Action 的所有错误
func (a *Action) GetErrors() []error {
    a.mutex.RLock()
    defer a.mutex.RUnlock()

    if a.errors == nil {
        return make([]error, 0)
    }
    result := make([]error, len(a.errors))
    copy(result, a.errors)
    return result
}

// HasErrors 是否有错误
func (a *Action) HasErrors() bool {
    return len(a.GetErrors()) > 0
}
```

### 4. 异步任务错误处理

```go
// 位于: tui/framework/async/task.go

package async

import "github.com/yaoapp/yao/tui/framework/result"

// Task 异步任务
type Task struct {
    id       string
    fn       func() result.Result[any]
    resultCh chan result.Result[any]
    timeout  time.Duration
    cancel   chan struct{}
}

// NewTask 创建异步任务
func NewTask(id string, fn func() result.Result[any]) *Task {
    return &Task{
        id:       id,
        fn:       fn,
        resultCh: make(chan result.Result[any], 1),
        cancel:   make(chan struct{}),
    }
}

// WithTimeout 设置超时
func (t *Task) WithTimeout(timeout time.Duration) *Task {
    t.timeout = timeout
    return t
}

// Execute 执行任务
func (t *Task) Execute() result.Result[any] {
    // 启动 goroutine 执行
    go func() {
        defer func() {
            if r := recover(); r != nil {
                t.resultCh <- result.Err[any]( fmt.Errorf("panic: %v", r))
            }
        }()

        result := t.fn()
        t.resultCh <- result
    }()

    // 等待结果或超时
    if t.timeout > 0 {
        select {
        case result := <-t.resultCh:
            return result
        case <-time.After(t.timeout):
            t.Cancel()
            return result.Err[any](ErrTimeout)
        case <-t.cancel:
            return result.Err[any](ErrCanceled)
        }
    }

    result := <-t.resultCh
    return result
}

// Cancel 取消任务
func (t *Task) Cancel() {
    close(t.cancel)
}

// TaskManager 任务管理器
type TaskManager struct {
    tasks   map[string]*Task
    results chan TaskResult
    mutex   sync.RWMutex
}

// TaskResult 任务结果
type TaskResult struct {
    TaskID string
    Result result.Result[any]
}

// NewTaskManager 创建任务管理器
func NewTaskManager() *TaskManager {
    return &TaskManager{
        tasks:   make(map[string]*Task),
        results: make(chan TaskResult, 100),
    }
}

// Submit 提交任务
func (m *TaskManager) Submit(task *Task) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    if _, exists := m.tasks[task.id]; exists {
        return fmt.Errorf("task %s already exists", task.id)
    }

    m.tasks[task.id] = task

    // 异步执行并收集结果
    go func() {
        result := task.Execute()
        m.results <- TaskResult{
            TaskID: task.id,
            Result: result,
        }

        // 清理完成的任务
        m.mutex.Lock()
        delete(m.tasks, task.id)
        m.mutex.Unlock()
    }()

    return nil
}

// Cancel 取消任务
func (m *TaskManager) Cancel(taskID string) error {
    m.mutex.RLock()
    task, exists := m.tasks[taskID]
    m.mutex.RUnlock()

    if !exists {
        return fmt.Errorf("task %s not found", taskID)
    }

    task.Cancel()
    return nil
}

// Results 获取结果通道
func (m *TaskManager) Results() <-chan TaskResult {
    return m.results
}
```

### 5. 装饰器错误处理

```go
// 位于: tui/framework/component/error_decorator.go

package component

// ErrorDecorator 错误装饰器
type ErrorDecorator struct {
    component Component
    handler   *ErrorHandler
}

// NewErrorDecorator 创建错误装饰器
func NewErrorDecorator(comp Component) *ErrorDecorator {
    return &ErrorDecorator{
        component: comp,
        handler:   NewErrorHandler(),
    }
}

// Paint 绘制组件（带错误保护）
func (d *ErrorDecorator) Paint(ctx PaintContext, buf *runtime.CellBuffer) {
    defer func() {
        if r := recover(); r != nil {
            err := fmt.Errorf("paint panic: %v", r)
            d.handler.HandleError(err)
            d.paintError(ctx, buf, err)
        }
    }()

    d.component.Paint(ctx, buf)
}

// HandleAction 处理 Action（带错误保护）
func (d *ErrorDecorator) HandleAction(a *action.Action) bool {
    defer func() {
        if r := recover(); r != nil {
            err := fmt.Errorf("handle action panic: %v", r)
            d.handler.HandleError(err)
            a.AttachError(err)
        }
    }()

    return d.component.HandleAction(a)
}

// paintError 绘制错误状态
func (d *ErrorDecorator) paintError(ctx PaintContext, buf *runtime.CellBuffer, err error) {
    // 显示错误占位符
    errorStyle := runtime.CellStyle{
        Foreground: runtime.ColorRed,
        Background: runtime.ColorDefault,
    }

    message := "Error: " + err.Error()
    for i, ch := range message {
        if i >= buf.Width {
            break
        }
        buf.SetCell(i, 0, ch, errorStyle)
    }
}

// OnError 注册错误回调
func (d *ErrorDecorator) OnError(fn func(error)) {
    d.handler.OnError(fn)
}

// SetBoundary 设置错误边界
func (d *ErrorDecorator) SetBoundary(boundary *ErrorBoundary) {
    d.handler.SetBoundary(boundary)
}
```

## 错误类型定义

```go
// 位于: tui/framework/errors/errors.go

package errors

import (
    stderrors "errors"
)

// 标准错误
var (
    ErrTimeout   = stderrors.New("timeout")
    ErrCanceled  = stderrors.New("canceled")
    ErrRestart   = stderrors.New("restart required")
    ErrInvalid   = stderrors.New("invalid input")
    ErrNotFound  = stderrors.New("not found")
    ErrDenied    = stderrors.New("access denied")
    ErrBusy      = stderrors.New("resource busy")
    ErrLimit     = stderrors.New("limit exceeded")
    ErrClosed    = stderrors.New("closed")
    ErrFormat    = stderrors.New("format error")
)

// TUIError TUI 框架错误接口
type TUIError interface {
    error
    // Code 错误代码
    Code() string
    // Recoverable 是否可恢复
    Recoverable() bool
    // Temporary 是否临时错误
    Temporary() bool
    // Cause 原因
    Cause() error
}

// BaseError 基础错误
type BaseError struct {
    code        string
    message     string
    cause       error
    recoverable bool
    temporary   bool
    stack       []string
}

func (e *BaseError) Error() string {
    if e.cause != nil {
        return e.code + ": " + e.message + ": " + e.cause.Error()
    }
    return e.code + ": " + e.message
}

func (e *BaseError) Code() string {
    return e.code
}

func (e *BaseError) Recoverable() bool {
    return e.recoverable
}

func (e *BaseError) Temporary() bool {
    return e.temporary
}

func (e *BaseError) Cause() error {
    return e.cause
}

// Unwrap 解包错误
func (e *BaseError) Unwrap() error {
    return e.cause
}

// New 创建新错误
func New(code, message string) *BaseError {
    return &BaseError{
        code:    code,
        message: message,
    }
}

// Wrap 包装错误
func Wrap(err error, code, message string) *BaseError {
    return &BaseError{
        code:    code,
        message: message,
        cause:   err,
    }
}

// Recoverable 设置可恢复
func (e *BaseError) Recoverable() *BaseError {
    e.recoverable = true
    return e
}

// Temporary 设置临时
func (e *BaseError) Temporary() *BaseError {
    e.temporary = true
    return e
}

// IsRecoverable 检查是否可恢复
func IsRecoverable(err error) bool {
    var tuiErr TUIError
    if stderrors.As(err, &tuiErr) {
        return tuiErr.Recoverable()
    }
    return false
}

// IsTemporary 检查是否临时
func IsTemporary(err error) bool {
    var tuiErr TUIError
    if stderrors.As(err, &tuiErr) {
        return tuiErr.Temporary()
    }
    return false
}
```

## 使用示例

### 示例 1：基础 Result 使用

```go
// ✅ 使用 Result 类型
func ParseInt(s string) result.Result[int] {
    val, err := strconv.Atoi(s)
    if err != nil {
        return result.Err[int](err)
    }
    return result.Ok(val)
}

// 链式调用
result := ParseInt("42").
    Map(func(x int) int { return x * 2 }).
    AndThen(func(x int) result.Result[int] {
        if x > 100 {
            return result.Err[int](errors.ErrLimit)
        }
        return result.Ok(x)
    })

if result.IsErr() {
    // 处理错误
}
```

### 示例 2：组件错误处理

```go
// ✅ 创建带错误处理的组件
func NewSafeButton() *Button {
    btn := &Button{
        BaseComponent: NewBaseComponent(),
    }

    // 添加错误处理器
    handler := NewErrorHandler()
    handler.OnError(func(err error) {
        log.Errorf("Button error: %v", err)
    })

    // 设置错误边界
    handler.SetBoundary(&ErrorBoundary{
        strategy: RecoverFallback,
        fallback: NewErrorPlaceholder(),
    })

    btn.errorHandler = handler
    return btn
}

// 在 HandleAction 中处理错误
func (b *Button) HandleAction(a *action.Action) bool {
    if a.Type == action.ActionClick {
        result := b.doClick()
        if result.IsErr() {
            b.errorHandler.HandleError(result.Error())
            a.AttachError(result.Error())
            return false
        }
    }
    return true
}
```

### 示例 3：异步任务错误处理

```go
// ✅ 创建异步任务
func LongRunningTask() result.Result[string] {
    // 模拟可能失败的操作
    if someCondition {
        return result.Err[string](errors.ErrTimeout)
    }
    return result.Ok("success")
}

// 提交任务
task := async.NewTask("task-1", LongRunningTask).
    WithTimeout(30 * time.Second)

manager := async.NewTaskManager()
manager.Submit(task)

// 监听结果
go func() {
    for result := range manager.Results() {
        if result.Result.IsErr() {
            // 处理错误
            log.Errorf("Task %s failed: %v",
                result.TaskID, result.Result.Error())
        } else {
            // 处理成功结果
            log.Infof("Task %s succeeded: %v",
                result.TaskID, result.Result.Value())
        }
    }
}()
```

### 示例 4：表单验证错误

```go
// ✅ 表单验证
type FormValidator struct {
    validators map[string]func(interface{}) result.Result[any]
}

func (v *FormValidator) Validate(field string, value interface{}) result.Result[any] {
    if validator, ok := v.validators[field]; ok {
        return validator(value)
    }
    return result.Ok(value)
}

// 使用
validator := &FormValidator{
    validators: map[string]func(interface{}) result.Result[any]{
        "email": func(v interface{}) result.Result[any] {
            email := v.(string)
            if !strings.Contains(email, "@") {
                return result.Err[any](errors.New("invalid email"))
            }
            return result.Ok(email)
        },
        "age": func(v interface{}) result.Result[any] {
            age := v.(int)
            if age < 0 || age > 150 {
                return result.Err[any](errors.New("invalid age"))
            }
            return result.Ok(age)
        },
    },
}

result := validator.Validate("email", "invalid")
if result.IsErr() {
    // 显示验证错误
}
```

### 示例 5：装饰器保护组件

```go
// ✅ 使用装饰器保护可能出错的组件
unsafeComp := NewComplexComponent()
safeComp := NewErrorDecorator(unsafeComp)

// 配置错误处理
safeComp.OnError(func(err error) {
    // 发送错误到监控系统
    monitoring.RecordError(err)
})

safeComp.SetBoundary(&ErrorBoundary{
    strategy: RecoverFallback,
    fallback: NewErrorPlaceholder(),
})

// 使用组件
app.Mount(safeComp)
```

## 与 Action 系统集成

```go
// ✅ 错误转换为 Action
func (h *ErrorHandler) DispatchError(err error, dispatcher *action.Dispatcher) {
    // 根据错误类型发送不同的 Action
    switch {
    case errors.Is(err, ErrTimeout):
        dispatcher.Dispatch(action.NewAction(action.ActionTimeout).
            WithPayload(err))
    case IsRecoverable(err):
        dispatcher.Dispatch(action.NewAction(action.ActionError).
            WithPayload(err))
    default:
        dispatcher.Dispatch(action.NewAction(action.ActionPanic).
            WithPayload(err))
    }
}

// ✅ 在 HandleAction 中传播错误
func (c *Component) HandleAction(a *action.Action) bool {
    result := c.processAction(a)
    if result.IsErr() {
        // 附加错误到 Action
        a.AttachError(result.Error())

        // 如果有错误处理器，使用它
        if c.errorHandler != nil {
            c.errorHandler.DispatchError(result.Error(), c.dispatcher)
        }

        return false
    }
    return true
}
```

## 错误恢复策略

```go
// RecoveryHandler 恢复处理器
type RecoveryHandler struct {
    strategies map[error.Type]RecoveryStrategy
    defaultStrategy RecoveryStrategy
}

// Recover 恢复错误
func (h *RecoveryHandler) Recover(err error) RecoveryAction {
    // 确定错误类型
    errType := ClassifyError(err)

    // 获取策略
    strategy, ok := h.strategies[errType]
    if !ok {
        strategy = h.defaultStrategy
    }

    // 应用策略
    switch strategy {
    case RecoverRetry:
        return h.retry(err)
    case RecoverFallback:
        return h.fallback(err)
    case RecoverRestart:
        return h.restart(err)
    default:
        return RecoveryActionPropagate
    }
}

// retry 重试
func (h *RecoveryHandler) retry(err error) RecoveryAction {
    if !IsTemporary(err) {
        return RecoveryActionPropagate
    }
    return RecoveryActionRetry
}

// fallback 降级
func (h *RecoveryHandler) fallback(err error) RecoveryAction {
    return RecoveryActionFallback
}

// restart 重启
func (h *RecoveryHandler) restart(err error) RecoveryAction {
    if IsRecoverable(err) {
        return RecoveryActionRestart
    }
    return RecoveryActionPropagate
}
```

## 测试

```go
// 位于: tui/framework/result/result_test.go

func TestResultOk(t *testing.T) {
    result := result.Ok(42)

    assert.True(t, result.IsOk())
    assert.False(t, result.IsErr())
    assert.Equal(t, 42, result.Unwrap())
}

func TestResultErr(t *testing.T) {
    err := errors.New("test error")
    result := result.Err[int](err)

    assert.False(t, result.IsOk())
    assert.True(t, result.IsErr())
    assert.Equal(t, err, result.Error())
}

func TestResultMap(t *testing.T) {
    result := result.Ok(42).
        Map(func(x int) int { return x * 2 })

    assert.Equal(t, 84, result.Unwrap())
}

func TestResultAndThen(t *testing.T) {
    result := result.Ok(42).
        AndThen(func(x int) result.Result[int] {
            if x > 100 {
                return result.Err[int](errors.ErrLimit)
            }
            return result.Ok(x)
        })

    assert.True(t, result.IsOk())
    assert.Equal(t, 42, result.Unwrap())
}

func TestErrorHandler(t *testing.T) {
    handler := NewErrorHandler()

    errorCount := 0
    handler.OnError(func(err error) {
        errorCount++
    })

    handler.HandleError(errors.New("test"))

    assert.Equal(t, 1, errorCount)
    assert.Equal(t, 1, handler.GetErrorCount())
}
```

## 总结

错误处理系统提供：

1. **Result 类型**: 类型安全的错误处理
2. **错误边界**: 组件级别的错误捕获和恢复
3. **异步错误**: 完整的异步任务错误处理
4. **装饰器保护**: 透明的错误处理包装
5. **Action 集成**: 错误通过 Action 系统传播
6. **可恢复性**: 支持重试、降级、重启等策略

## 相关文档

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 架构概览
- [ACTION_SYSTEM.md](./ACTION_SYSTEM.md) - Action 系统
- [COMPONENTS.md](./COMPONENTS.md) - 组件系统
- [STATE_MANAGEMENT.md](./STATE_MANAGEMENT.md) - 状态管理
- [ARCHITECTURE_INVARIANTS.md](./ARCHITECTURE_INVARIANTS.md) - 架构不变量
