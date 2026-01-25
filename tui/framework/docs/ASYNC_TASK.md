# Async Task System Design (V3)

> **优先级**: P1 (核心功能)
> **目标**: 支持长时间运行的计算任务
> **关键特性**: 任务管理、超时控制、进度反馈、取消机制

## 概述

在 TUI 应用中，有些操作需要较长时间才能完成，如：
- 从 API 获取数据
- 处理大文件
- 执行复杂计算
- 数据库查询

这些操作不能阻塞 UI 线程，否则会导致界面冻结。异步任务系统提供了一种非阻塞执行这些操作的方式。

### 为什么需要异步任务系统？

**没有异步任务系统的问题**：
```go
// ❌ 阻塞 UI
func (b *Button) HandleClick() {
    data := fetchFromAPI()  // 阻塞 5 秒！
    displayData(data)
}

// 问题：
// - UI 冻结
// - 用户无法操作
// - 无法显示进度
// - 无法取消操作
```

**有异步任务系统的优势**：
```go
// ✅ 非阻塞执行
func (b *Button) HandleClick() {
    task := async.NewTask("fetch", func() async.Result {
        return fetchFromAPI()
    }).WithTimeout(30 * time.Second).WithProgress()

    async.Submit(task)

    // 显示加载状态
    showLoading("Fetching data...")
}

// 优势：
// - UI 保持响应
// - 显示进度反馈
// - 可以取消操作
// - 超时自动处理
```

## 设计目标

1. **非阻塞**: 不阻塞 UI 线程
2. **可取消**: 支持任务取消
3. **超时控制**: 支持超时自动取消
4. **进度反馈**: 实时进度更新
5. **错误处理**: 完善的错误处理机制
6. **状态管理**: 清晰的任务状态追踪

## 核心类型定义

### 1. Task 接口

```go
// 位于: tui/framework/async/task.go

package async

import (
    "context"
    "time"
    "github.com/yaoapp/yao/tui/framework/result"
)

// Task 任务接口
type Task interface {
    // ID 获取任务 ID
    ID() string

    // Name 获取任务名称
    Name() string

    // Execute 执行任务
    Execute(ctx context.Context) result.Result[any]

    // Cancel 取消任务
    Cancel()

    // IsCanceled 是否已取消
    IsCanceled() bool

    // Status 获取任务状态
    Status() TaskStatus
}

// TaskStatus 任务状态
type TaskStatus int

const (
    StatusPending   TaskStatus = iota  // 等待中
    StatusRunning                      // 运行中
    StatusCompleted                    // 已完成
    StatusFailed                       // 已失败
    StatusCanceled                     // 已取消
    StatusTimeout                      // 已超时
)

// String 返回状态字符串
func (s TaskStatus) String() string {
    switch s {
    case StatusPending:
        return "pending"
    case StatusRunning:
        return "running"
    case StatusCompleted:
        return "completed"
    case StatusFailed:
        return "failed"
    case StatusCanceled:
        return "canceled"
    case StatusTimeout:
        return "timeout"
    default:
        return "unknown"
    }
}
```

### 2. BaseTask 实现

```go
// 位于: tui/framework/async/base_task.go

package async

import (
    "context"
    "fmt"
    "sync"
    "sync/atomic"
    "time"
)

// BaseTask 基础任务
type BaseTask struct {
    id       string
    name     string
    fn       TaskFunc
    timeout  time.Duration
    progress *Progress
    cancel   context.CancelFunc
    status   atomic.Value // TaskStatus
    result   chan result.Result[any]
    mu       sync.RWMutex
}

// TaskFunc 任务函数
type TaskFunc func(ctx context.Context, progress *Progress) result.Result[any]

// NewTask 创建任务
func NewTask(id, name string, fn TaskFunc) *BaseTask {
    return &BaseTask{
        id:       id,
        name:     name,
        fn:       fn,
        progress: NewProgress(),
        result:   make(chan result.Result[any], 1),
        status:   atomic.Value{},
    }
    t.status.Store(StatusPending)
}

// ID 获取任务 ID
func (t *BaseTask) ID() string {
    return t.id
}

// Name 获取任务名称
func (t *BaseTask) Name() string {
    return t.name
}

// WithTimeout 设置超时
func (t *BaseTask) WithTimeout(timeout time.Duration) *BaseTask {
    t.timeout = timeout
    return t
}

// WithProgress 设置进度
func (t *BaseTask) WithProgress() *BaseTask {
    return t // Progress 已默认创建
}

// Execute 执行任务
func (t *BaseTask) Execute(ctx context.Context) result.Result[any] {
    t.mu.Lock()

    // 创建带取消的上下文
    taskCtx, cancel := context.WithCancel(ctx)
    t.cancel = cancel

    // 设置超时
    if t.timeout > 0 {
        taskCtx, cancel = context.WithTimeout(taskCtx, t.timeout)
    }

    t.mu.Unlock()

    // 更新状态为运行中
    t.setStatus(StatusRunning)

    // 在 goroutine 中执行
    go func() {
        defer func() {
            if r := recover(); r != nil {
                t.result <- result.Err[any](fmt.Errorf("panic: %v", r))
            }
        }()

        res := t.fn(taskCtx, t.progress)
        t.result <- res
    }()

    // 等待结果
    select {
    case res := <-t.result:
        if res.IsOk() {
            t.setStatus(StatusCompleted)
        } else {
            t.setStatus(StatusFailed)
        }
        return res

    case <-taskCtx.Done():
        switch {
        case taskCtx.Err() == context.DeadlineExceeded:
            t.setStatus(StatusTimeout)
            return result.Err[any](ErrTimeout)
        case taskCtx.Err() == context.Canceled:
            t.setStatus(StatusCanceled)
            return result.Err[any](ErrCanceled)
        default:
            t.setStatus(StatusFailed)
            return result.Err[any](taskCtx.Err())
        }
    }
}

// Cancel 取消任务
func (t *BaseTask) Cancel() {
    t.mu.Lock()
    defer t.mu.Unlock()

    if t.cancel != nil {
        t.cancel()
    }
    t.setStatus(StatusCanceled)
}

// IsCanceled 是否已取消
func (t *BaseTask) IsCanceled() bool {
    return t.Status() == StatusCanceled
}

// Status 获取状态
func (t *BaseTask) Status() TaskStatus {
    return t.status.Load().(TaskStatus)
}

// setStatus 设置状态
func (t *BaseTask) setStatus(status TaskStatus) {
    t.status.Store(status)
}

// Progress 获取进度
func (t *BaseTask) Progress() *Progress {
    return t.progress
}
```

### 3. Progress 进度

```go
// 位于: tui/framework/async/progress.go

package async

import (
    "sync"
    "time"
)

// Progress 进度
type Progress struct {
    current int64
    total   int64
    message string
    mu      sync.RWMutex
    subscribers []chan ProgressUpdate
}

// ProgressUpdate 进度更新
type ProgressUpdate struct {
    Current int64
    Total   int64
    Percent float64
    Message string
    Time    time.Time
}

// NewProgress 创建进度
func NewProgress() *Progress {
    return &Progress{
        total:       100,
        subscribers: make([]chan ProgressUpdate, 0),
    }
}

// SetTotal 设置总数
func (p *Progress) SetTotal(total int64) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.total = total
    p.notify()
}

// SetCurrent 设置当前值
func (p *Progress) SetCurrent(current int64) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.current = current
    p.notify()
}

// Increment 增加
func (p *Progress) Increment(delta int64) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.current += delta
    p.notify()
}

// SetMessage 设置消息
func (p *Progress) SetMessage(message string) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.message = message
    p.notify()
}

// Current 获取当前值
func (p *Progress) Current() int64 {
    p.mu.RLock()
    defer p.mu.RUnlock()
    return p.current
}

// Total 获取总数
func (p *Progress) Total() int64 {
    p.mu.RLock()
    defer p.mu.RUnlock()
    return p.total
}

// Percent 获取百分比
func (p *Progress) Percent() float64 {
    p.mu.RLock()
    defer p.mu.RUnlock()

    if p.total == 0 {
        return 0
    }
    return float64(p.current) / float64(p.total) * 100
}

// Message 获取消息
func (p *Progress) Message() string {
    p.mu.RLock()
    defer p.mu.RUnlock()
    return p.message
}

// Subscribe 订阅进度更新
func (p *Progress) Subscribe() chan ProgressUpdate {
    p.mu.Lock()
    defer p.mu.Unlock()

    ch := make(chan ProgressUpdate, 100)
    p.subscribers = append(p.subscribers, ch)
    return ch
}

// Unsubscribe 取消订阅
func (p *Progress) Unsubscribe(ch chan ProgressUpdate) {
    p.mu.Lock()
    defer p.mu.Unlock()

    for i, subscriber := range p.subscribers {
        if subscriber == ch {
            p.subscribers = append(p.subscribers[:i], p.subscribers[i+1:]...)
            close(ch)
            break
        }
    }
}

// notify 通知订阅者
func (p *Progress) notify() {
    update := ProgressUpdate{
        Current: p.current,
        Total:   p.total,
        Percent: p.Percent(),
        Message: p.message,
        Time:    time.Now(),
    }

    for _, subscriber := range p.subscribers {
        select {
        case subscriber <- update:
        default:
            // 防止阻塞
        }
    }
}
```

### 4. TaskManager 任务管理器

```go
// 位于: tui/framework/async/manager.go

package async

import (
    "context"
    "fmt"
    "sync"
    "time"
)

// TaskManager 任务管理器
type TaskManager struct {
    tasks   map[string]Task
    results chan TaskResult
    mu      sync.RWMutex
    ctx     context.Context
    cancel  context.CancelFunc
}

// TaskResult 任务结果
type TaskResult struct {
    TaskID  string
    Name    string
    Status  TaskStatus
    Result  result.Result[any]
    Time    time.Time
}

// NewTaskManager 创建任务管理器
func NewTaskManager() *TaskManager {
    ctx, cancel := context.WithCancel(context.Background())
    return &TaskManager{
        tasks:   make(map[string]Task),
        results: make(chan TaskResult, 100),
        ctx:     ctx,
        cancel:  cancel,
    }
}

// Submit 提交任务
func (m *TaskManager) Submit(task Task) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    if _, exists := m.tasks[task.ID()]; exists {
        return fmt.Errorf("task %s already exists", task.ID())
    }

    m.tasks[task.ID()] = task

    // 异步执行
    go func(t Task) {
        res := t.Execute(m.ctx)

        // 发送结果
        m.results <- TaskResult{
            TaskID: t.ID(),
            Name:   t.Name(),
            Status: t.Status(),
            Result: res,
            Time:   time.Now(),
        }

        // 清理完成的任务
        m.mu.Lock()
        delete(m.tasks, t.ID())
        m.mu.Unlock()
    }(task)

    return nil
}

// Cancel 取消任务
func (m *TaskManager) Cancel(taskID string) error {
    m.mu.RLock()
    task, exists := m.tasks[taskID]
    m.mu.RUnlock()

    if !exists {
        return fmt.Errorf("task %s not found", taskID)
    }

    task.Cancel()
    return nil
}

// Get 获取任务
func (m *TaskManager) Get(taskID string) (Task, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    task, exists := m.tasks[taskID]
    if !exists {
        return nil, fmt.Errorf("task %s not found", taskID)
    }

    return task, nil
}

// List 列出所有任务
func (m *TaskManager) List() []Task {
    m.mu.RLock()
    defer m.mu.RUnlock()

    tasks := make([]Task, 0, len(m.tasks))
    for _, task := range m.tasks {
        tasks = append(tasks, task)
    }
    return tasks
}

// Results 获取结果通道
func (m *TaskManager) Results() <-chan TaskResult {
    return m.results
}

// Shutdown 关闭管理器
func (m *TaskManager) Shutdown() {
    m.cancel()

    // 等待所有任务完成或超时
    deadline := time.After(5 * time.Second)
    for len(m.tasks) > 0 {
        select {
        case <-deadline:
            // 超时，强制取消所有任务
            m.mu.Lock()
            for _, task := range m.tasks {
                task.Cancel()
            }
            m.mu.Unlock()
            return
        case <-time.After(100 * time.Millisecond):
            // 继续等待
        }
    }
}
```

### 5. UI 集成

```go
// 位于: tui/framework/component/async_indicator.go

package component

import (
    "fmt"
    "time"

    "github.com/yaoapp/yao/tui/framework/async"
    "github.com/yaoapp/yao/tui/runtime"
)

// AsyncIndicator 异步任务指示器
type AsyncIndicator struct {
    BaseComponent
    *Measurable

    manager   *async.TaskManager
    activeTasks map[string]*TaskStatus
    mu        sync.RWMutex
}

// TaskStatus 任务状态显示
type TaskStatus struct {
    Name     string
    Status   async.TaskStatus
    Progress float64
    Message  string
}

// NewAsyncIndicator 创建异步任务指示器
func NewAsyncIndicator(manager *async.TaskManager) *AsyncIndicator {
    ind := &AsyncIndicator{
        manager:     manager,
        activeTasks: make(map[string]*TaskStatus),
    }

    ind.Measurable = NewMeasurable()

    // 监听任务结果
    go ind.watchResults()

    return ind
}

// watchResults 监听任务结果
func (ind *AsyncIndicator) watchResults() {
    for result := range ind.manager.Results() {
        ind.mu.Lock()

        if result.Status == async.StatusRunning {
            // 任务正在运行，显示进度
            if task := ind.getTask(result.TaskID); task != nil {
                ind.activeTasks[result.TaskID] = &TaskStatus{
                    Name:     result.Name,
                    Status:   result.Status,
                    Progress: task.Progress().Percent(),
                    Message:  task.Progress().Message(),
                }
            }
        } else {
            // 任务完成，移除显示
            delete(ind.activeTasks, result.TaskID)
        }

        ind.mu.Unlock()
        ind.MarkDirty()
    }
}

// getTask 获取任务
func (ind *AsyncIndicator) getTask(id string) async.Task {
    task, _ := ind.manager.Get(id)
    return task
}

// Paint 绘制指示器
func (ind *AsyncIndicator) Paint(ctx PaintContext, buf *runtime.CellBuffer) {
    ind.mu.RLock()
    defer ind.mu.RUnlock()

    if len(ind.activeTasks) == 0 {
        return
    }

    y := 0
    for _, status := range ind.activeTasks {
        // 绘制任务状态
        line := ind.formatStatus(status)
        buf.DrawText(0, y, line, ind.getStyle(status.Status))
        y++
    }
}

// formatStatus 格式化状态
func (ind *AsyncIndicator) formatStatus(status *TaskStatus) string {
    switch status.Status {
    case async.StatusRunning:
        if status.Progress > 0 {
            return fmt.Sprintf("⟳ %s: %.0f%% - %s",
                status.Name, status.Progress, status.Message)
        }
        return fmt.Sprintf("⟳ %s: %s", status.Name, status.Message)
    case async.StatusCompleted:
        return fmt.Sprintf("✓ %s: completed", status.Name)
    case async.StatusFailed:
        return fmt.Sprintf("✗ %s: failed", status.Name)
    case async.StatusCanceled:
        return fmt.Sprintf("⊘ %s: canceled", status.Name)
    case async.StatusTimeout:
        return fmt.Sprintf("⏱ %s: timeout", status.Name)
    default:
        return status.Name
    }
}

// getStyle 获取样式
func (ind *AsyncIndicator) getStyle(status async.TaskStatus) runtime.Style {
    theme := ind.GetTheme()

    switch status {
    case async.StatusRunning:
        return theme.GetStyle("async.running")
    case async.StatusCompleted:
        return theme.GetStyle("async.completed")
    case async.StatusFailed:
        return theme.GetStyle("async.failed")
    case async.StatusCanceled:
        return theme.GetStyle("async.canceled")
    case async.StatusTimeout:
        return theme.GetStyle("async.timeout")
    default:
        return theme.GetStyle("async.default")
    }
}

// Measure 测量尺寸
func (ind *AsyncIndicator) Measure(maxWidth, maxHeight int) (width, height int) {
    ind.mu.RLock()
    taskCount := len(ind.activeTasks)
    ind.mu.RUnlock()

    // 每个任务一行
    return maxWidth, taskCount
}
```

## 使用示例

### 示例 1：简单异步任务

```go
// ✅ 创建并提交任务
manager := async.NewTaskManager()

task := async.NewTask("fetch-user", "Fetch User Data",
    func(ctx context.Context, progress *async.Progress) result.Result[any] {
        // 模拟 API 请求
        progress.SetMessage("Connecting to API...")
        time.Sleep(1 * time.Second)

        progress.SetMessage("Fetching data...")
        progress.SetCurrent(30)
        time.Sleep(1 * time.Second)

        progress.SetMessage("Processing...")
        progress.SetCurrent(70)
        time.Sleep(1 * time.Second)

        progress.SetCurrent(100)

        // 返回结果
        return result.Ok(map[string]interface{}{
            "name":  "John Doe",
            "email": "john@example.com",
        })
    })

manager.Submit(task)
```

### 示例 2：带超时和取消

```go
// ✅ 超时和取消
task := async.NewTask("long-task", "Long Running Task",
    func(ctx context.Context, progress *async.Progress) result.Result[any] {
        for i := 0; i <= 100; i += 10 {
            select {
            case <-ctx.Done():
                // 任务被取消
                return result.Err[any](ctx.Err())
            default:
                progress.SetCurrent(int64(i))
                time.Sleep(500 * time.Millisecond)
            }
        }
        return result.Ok("completed")
    }).
    WithTimeout(30 * time.Second).  // 30 秒超时
    WithProgress()

manager.Submit(task)

// 稍后取消
go func() {
    time.Sleep(5 * time.Second)
    manager.Cancel("long-task")
}()
```

### 示例 3：进度显示

```go
// ✅ 显示进度
task := async.NewTask("download", "Download File",
    func(ctx context.Context, progress *async.Progress) result.Result[any] {
        progress.SetTotal(1000)

        for i := int64(0); i < 1000; i++ {
            select {
            case <-ctx.Done():
                return result.Err[any](ctx.Err())
            default:
                // 下载一块数据
                time.Sleep(10 * time.Millisecond)
                progress.SetCurrent(i + 1)

                if i%100 == 0 {
                    progress.SetMessage(fmt.Sprintf("Downloaded %d KB", i/10))
                }
            }
        }

        return result.Ok("download complete")
    })

// 订阅进度
progressCh := task.Progress().Subscribe()
go func() {
    for update := range progressCh {
        fmt.Printf("Progress: %.0f%% - %s\n",
            update.Percent, update.Message)
    }
    defer task.Progress().Unsubscribe(progressCh)
}()

manager.Submit(task)
```

### 示例 4：与 Action 系统集成

```go
// ✅ 集成到 Action 系统
type DataLoader struct {
    manager *async.TaskManager
}

func (l *DataLoader) HandleAction(a *action.Action) bool {
    switch a.Type {
    case action.ActionLoad:
        // 启动异步加载
        task := async.NewTask("load-data", "Loading Data",
            func(ctx context.Context, progress *async.Progress) result.Result[any] {
                data, err := l.loadFromAPI(ctx, progress)
                if err != nil {
                    return result.Err[any](err)
                }
                return result.Ok(data)
            }).WithProgress()

        l.manager.Submit(task)

        // 发送状态更新 Action
        l.dispatcher.Dispatch(action.NewAction(action.ActionStatus).
            WithPayload("Loading..."))

        return true

    case action.ActionCancel:
        // 取消任务
        if taskID, ok := a.Payload.(string); ok {
            l.manager.Cancel(taskID)
        }
        return true
    }

    return false
}
```

### 示例 5：错误处理

```go
// ✅ 完整的错误处理
task := async.NewTask("risky-task", "Risky Task",
    func(ctx context.Context, progress *async.Progress) result.Result[any] {
        // 可能失败的操作
        data, err := someOperation()
        if err != nil {
            return result.Err[any](fmt.Errorf("operation failed: %w", err))
        }

        return result.Ok(data)
    })

// 监听结果
go func() {
    for result := range manager.Results() {
        if result.Result.IsErr() {
            // 处理错误
            switch {
            case errors.Is(result.Result.Error(), async.ErrTimeout):
                showError("Operation timed out. Please try again.")
            case errors.Is(result.Result.Error(), async.ErrCanceled):
                showInfo("Operation was canceled.")
            default:
                showError("Operation failed: " + result.Result.Error().Error())
            }
        } else {
            // 处理成功结果
            showData(result.Result.Value())
        }
    }
}()

manager.Submit(task)
```

## 与 Action 系统集成

```go
// 位于: tui/framework/action/async_actions.go

package action

const (
    // 异步相关 Action
    ActionAsyncStart   ActionType = "async.start"
    ActionAsyncCancel  ActionType = "async.cancel"
    ActionAsyncProgress ActionType = "async.progress"
    ActionAsyncComplete ActionType = "async.complete"
    ActionAsyncError    ActionType = "async.error"
)

// AsyncStartPayload 异步启动负载
type AsyncStartPayload struct {
    TaskID   string
    TaskName string
    Timeout  time.Duration
}

// AsyncCancelPayload 异步取消负载
type AsyncCancelPayload struct {
    TaskID string
}

// AsyncCompletePayload 异步完成负载
type AsyncCompletePayload struct {
    TaskID string
    Result any
}

// AsyncErrorPayload 异步错误负载
type AsyncErrorPayload struct {
    TaskID string
    Error  error
}

// AsyncProgressPayload 异步进度负载
type AsyncProgressPayload struct {
    TaskID   string
    Progress float64
    Message  string
}
```

## 主题样式

```css
/* 主题中定义异步任务样式 */
async:
  running:
    foreground: blue
    background: default
  completed:
    foreground: green
    background: default
  failed:
    foreground: red
    background: default
  canceled:
    foreground: yellow
    background: default
  timeout:
    foreground: magenta
    background: default
```

## 测试

```go
// 位于: tui/framework/async/manager_test.go

func TestTaskManager(t *testing.T) {
    manager := async.NewTaskManager()

    task := async.NewTask("test", "Test Task",
        func(ctx context.Context, progress *async.Progress) result.Result[any] {
            return result.Ok("success")
        })

    err := manager.Submit(task)
    assert.NoError(t, err)

    // 等待结果
    select {
    case result := <-manager.Results():
        assert.True(t, result.Result.IsOk())
        assert.Equal(t, "success", result.Result.Value())
    case <-time.After(5 * time.Second):
        t.Fatal("timeout waiting for result")
    }
}

func TestTaskCancel(t *testing.T) {
    manager := async.NewTaskManager()

    task := async.NewTask("long", "Long Task",
        func(ctx context.Context, progress *async.Progress) result.Result[any] {
            select {
            case <-time.After(10 * time.Second):
                return result.Ok("completed")
            case <-ctx.Done():
                return result.Err[any](ctx.Err())
            }
        })

    manager.Submit(task)

    // 立即取消
    err := manager.Cancel("long")
    assert.NoError(t, err)

    result := <-manager.Results()
    assert.Equal(t, async.StatusCanceled, result.Status)
}

func TestTaskTimeout(t *testing.T) {
    manager := async.NewTaskManager()

    task := async.NewTask("slow", "Slow Task",
        func(ctx context.Context, progress *async.Progress) result.Result[any] {
            time.Sleep(5 * time.Second)
            return result.Ok("completed")
        }).WithTimeout(100 * time.Millisecond)

    manager.Submit(task)

    result := <-manager.Results()
    assert.Equal(t, async.StatusTimeout, result.Status)
}
```

## 总结

异步任务系统提供：

1. **非阻塞执行**: 不阻塞 UI 线程
2. **超时控制**: 支持任务超时自动取消
3. **取消机制**: 支持主动取消任务
4. **进度反馈**: 实时进度更新
5. **错误处理**: 完善的错误处理机制
6. **UI 集成**: 与 Action 系统无缝集成

## 相关文档

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 架构概览
- [ACTION_SYSTEM.md](./ACTION_SYSTEM.md) - Action 系统
- [ERROR_HANDLING.md](./ERROR_HANDLING.md) - 错误处理
- [STATE_MANAGEMENT.md](./STATE_MANAGEMENT.md) - 状态管理
- [THEME_SYSTEM.md](./THEME_SYSTEM.md) - 主题系统
