# 批量/顺序 Action 增强方案

## 概述

批量/顺序 Action 功能借鉴自 Bubble Tea 的 `Batch` 和 `Sequence` 命令，允许简化异步操作的编排，支持并发执行和顺序执行两种模式。

## 当前问题

### 异步操作编排复杂

当前 Yao TUI 的异步操作需要手动管理 goroutine 和同步：

```go
// 当前方式：繁琐的手动管理
func (c *Component) LoadMultipleData() {
    var wg sync.WaitGroup
    errors := make(chan error, 3)

    wg.Add(3)
    go func() { defer wg.Done(); errors <- c.loadUsers() }()
    go func() { defer wg.Done(); errors <- c.loadPosts() }()
    go func() { defer wg.Done(); errors <- c.loadComments() }()

    go func() {
        wg.Wait()
        close(errors)
        for err := range errors {
            if err != nil {
                c.Dispatch(NewErrorAction(err))
            }
        }
    }()
}
```

### 具体痛点

1. **并发控制困难** - 需要手动使用 WaitGroup、Channel
2. **错误处理分散** - 每个异步操作的错误处理逻辑重复
3. **顺序执行复杂** - 需要手动串行化异步操作
4. **组合能力弱** - 难以组合多个异步操作
5. **取消传播** - 取消操作难以传播到所有子任务

## 设计方案

### 核心接口

```go
// tui/runtime/action/composite.go

package action

import (
    "context"
    "sync"
)

// CompositeAction 复合 Action
type CompositeAction struct {
    mode     Mode
    actions  []Action
    callback Callback
}

// Mode 执行模式
type Mode int

const (
    // Sequential 顺序执行
    Sequential Mode = iota
    // Concurrent 并发执行
    Concurrent
)

// Callback 完成回调
type Callback func(results []ActionResult)

// Batch 并发执行多个 Action
func Batch(actions ...Action) Action {
    return &CompositeAction{
        mode:    Concurrent,
        actions: actions,
    }
}

// Sequence 顺序执行多个 Action
func Sequence(actions ...Action) Action {
    return &CompositeAction{
        mode:    Sequential,
        actions: actions,
    }
}

// WithCallback 设置完成回调
func (a *CompositeAction) WithCallback(cb Callback) Action {
    a.callback = cb
    return a
}

// Execute 执行复合 Action
func (a *CompositeAction) Execute(ctx *ActionContext) ActionResult {
    switch a.mode {
    case Concurrent:
        return a.executeConcurrent(ctx)
    case Sequential:
        return a.executeSequential(ctx)
    default:
        return ActionResultError{Err: ErrInvalidMode}
    }
}

// executeConcurrent 并发执行
func (a *CompositeAction) executeConcurrent(ctx *ActionContext) ActionResult {
    if len(a.actions) == 0 {
        return ActionResultOK
    }

    var wg sync.WaitGroup
    results := make([]ActionResult, len(a.actions))
    errors := make(chan error, len(a.actions))

    for i, action := range a.actions {
        wg.Add(1)
        go func(idx int, act Action) {
            defer wg.Done()
            result := act.Execute(ctx)
            results[idx] = result
            if err := result.Error(); err != nil {
                errors <- err
            }
        }(i, action)
    }

    wg.Wait()
    close(errors)

    // 收集错误
    var errs []error
    for err := range errors {
        errs = append(errs, err)
    }

    // 调用回调
    if a.callback != nil {
        a.callback(results)
    }

    if len(errs) > 0 {
        return ActionResultError{Err: &MultipleError{Errors: errs}}
    }

    return ActionResultOK
}

// executeSequential 顺序执行
func (a *CompositeAction) executeSequential(ctx *ActionContext) ActionResult {
    if len(a.actions) == 0 {
        return ActionResultOK
    }

    results := make([]ActionResult, 0, len(a.actions))

    for _, action := range a.actions {
        // 检查是否已取消
        if ctx.Err() != nil {
            return ActionResultCanceled
        }

        result := action.Execute(ctx)
        results = append(results, result)

        // 如果有错误且需要停止，可以中断执行
        // 这里选择继续执行，收集所有错误
    }

    // 调用回调
    if a.callback != nil {
        a.callback(results)
    }

    return ActionResultOK
}

// MultipleError 多错误包装
type MultipleError struct {
    Errors []error
}

func (e *MultipleError) Error() string {
    return "multiple errors occurred"
}

// CancelableAction 可取消的复合 Action
type CancelableAction struct {
    *CompositeAction
    cancel  context.CancelFunc
    canceled atomic.Bool
}

// Cancel 取消执行
func (a *CancelableAction) Cancel() {
    if a.canceled.CompareAndSwap(false, true) {
        if a.cancel != nil {
            a.cancel()
        }
    }
}
```

### 并发限制

```go
// tui/runtime/action/pool.go

package action

import (
    "context"
    "sync"
)

// WorkerPool 工作池
type WorkerPool struct {
    maxWorkers int
    queue      chan Action
    wg         sync.WaitGroup
}

// NewWorkerPool 创建工作池
func NewWorkerPool(maxWorkers int) *WorkerPool {
    return &WorkerPool{
        maxWorkers: maxWorkers,
        queue:      make(chan Action, 100),
    }
}

// Start 启动工作池
func (p *WorkerPool) Start(ctx context.Context) {
    for i := 0; i < p.maxWorkers; i++ {
        p.wg.Add(1)
        go p.worker(ctx)
    }
}

// worker 工作协程
func (p *WorkerPool) worker(ctx context.Context) {
    defer p.wg.Done()
    for {
        select {
        case action := <-p.queue:
            action.Execute(NewActionContext(ctx, nil))
        case <-ctx.Done():
            return
        }
    }
}

// Submit 提交任务
func (p *WorkerPool) Submit(action Action) error {
    select {
    case p.queue <- action:
        return nil
    default:
        return ErrPoolFull
    }
}

// Stop 停止工作池
func (p *WorkerPool) Stop() {
    close(p.queue)
    p.wg.Wait()
}

// ParallelWithLimit 限制并发数的并行执行
func ParallelWithLimit(limit int, actions ...Action) Action {
    return &ParallelLimitedAction{
        actions: actions,
        limit:   limit,
    }
}

type ParallelLimitedAction struct {
    actions []Action
    limit   int
}

func (a *ParallelLimitedAction) Execute(ctx *ActionContext) ActionResult {
    pool := NewWorkerPool(a.limit)
    pool.Start(ctx.Context())
    defer pool.Stop()

    results := make([]ActionResult, len(a.actions))

    var wg sync.WaitGroup
    for i, action := range a.actions {
        wg.Add(1)
        go func(idx int, act Action) {
            defer wg.Done()
            results[idx] = act.Execute(ctx)
        }(i, action)
    }

    wg.Wait()
    return ActionResultOK
}
```

### 错误处理策略

```go
// tui/runtime/action/error_policy.go

package action

// ErrorPolicy 错误处理策略
type ErrorPolicy int

const (
    // StopOnError 遇到错误立即停止
    StopOnError ErrorPolicy = iota
    // ContinueOnError 遇到错误继续执行
    ContinueOnError
    // CollectErrors 收集所有错误
    CollectErrors
)

// WithErrorPolicy 应用错误策略
func (a *CompositeAction) WithErrorPolicy(policy ErrorPolicy) Action {
    // 根据策略修改执行逻辑
    return a
}
```

## 使用示例

### 1. 并发加载数据

```go
func (c *Component) LoadData() Action {
    return Batch(
        NewAction(func(ctx *ActionContext) ActionResult {
            users := fetchUsers()
            ctx.Dispatch(NewUsersLoadedAction(users))
            return ActionResultOK
        }),
        NewAction(func(ctx *ActionContext) ActionResult {
            posts := fetchPosts()
            ctx.Dispatch(NewPostsLoadedAction(posts))
            return ActionResultOK
        }),
        NewAction(func(ctx *ActionContext) ActionResult {
            comments := fetchComments()
            ctx.Dispatch(NewCommentsLoadedAction(comments))
            return ActionResultOK
        }),
    )
}
```

### 2. 顺序执行操作

```go
func (c *Component) SubmitForm() Action {
    return Sequence(
        NewAction(validateForm),
        NewAction(saveDraft),
        NewAction(uploadFiles),
        NewAction(submitForm),
        NewAction(showSuccessMessage),
    )
}
```

### 3. 带回调的复合操作

```go
func (c *Component) LoadWithCallback() Action {
    return Batch(
        c.loadUsers(),
        c.loadPosts(),
    ).WithCallback(func(results []ActionResult) {
        // 所有操作完成后执行
        c.Dispatch(NewLoadCompleteAction())
        c.HideLoadingIndicator()
    })
}
```

### 4. 并发限制

```go
func (c *Component) ProcessMany(items []Item) Action {
    actions := make([]Action, len(items))
    for i, item := range items {
        actions[i] = NewAction(func(ctx *ActionContext) ActionResult {
            return processItem(item)
        })
    }

    // 最多 10 个并发
    return ParallelWithLimit(10, actions...)
}
```

### 5. 错误处理

```go
func (c *Component) LoadWithErrorHandling() Action {
    composite := Batch(
        c.loadUsers(),
        c.loadPosts(),
    )

    composite.WithCallback(func(results []ActionResult) {
        for _, result := range results {
            if err := result.Error(); err != nil {
                c.Dispatch(NewErrorAction(err))
            }
        }
    })

    return composite
}
```

### 6. 可取消的操作

```go
func (c *Component) LongRunningOperation() *CancelableAction {
    base := &CompositeAction{
        mode: Concurrent,
        actions: []Action{
            c.processPart1(),
            c.processPart2(),
            c.processPart3(),
        },
    }

    ctx, cancel := context.WithCancel(context.Background())
    return &CancelableAction{
        CompositeAction: base,
        cancel:         cancel,
        canceled:       atomic.Bool{},
    }
}

// 使用
func (c *Component) Start() {
    c.currentOp = c.LongRunningOperation()
    c.currentOp.Execute(ctx)
}

func (c *Component) Cancel() {
    if c.currentOp != nil {
        c.currentOp.Cancel()
    }
}
```

## 实施计划

### Phase 1: 核心接口 (Week 1)

- [ ] 实现 `CompositeAction`
- [ ] 实现 `Batch` 和 `Sequence`
- [ ] 单元测试

### Phase 2: 高级特性 (Week 2)

- [ ] 实现 `WorkerPool`
- [ ] 实现 `ParallelWithLimit`
- [ ] 实现 `CancelableAction`
- [ ] 集成测试

### Phase 3: 错误处理 (Week 2)

- [ ] 实现 `ErrorPolicy`
- [ ] 实现 `MultipleError`
- [ ] 错误恢复机制

### Phase 4: 文档和示例 (Week 3)

- [ ] API 文档
- [ ] 使用示例
- [ ] 性能基准测试

## 文件结构

```
tui/runtime/action/
├── composite.go           # 复合 Action
├── pool.go                # 工作池
├── error_policy.go        # 错误策略
├── cancelable.go          # 可取消 Action
├── batch.go               # Batch 快捷方法
├── sequence.go            # Sequence 快捷方法
└── composite_test.go      # 测试
```

## 测试策略

```go
func TestBatch(t *testing.T) {
    results := make([]int, 3)
    actions := []Action{
        NewAction(func(ctx *ActionContext) ActionResult {
            results[0] = 1
            return ActionResultOK
        }),
        NewAction(func(ctx *ActionContext) ActionResult {
            results[1] = 2
            return ActionResultOK
        }),
        NewAction(func(ctx *ActionContext) ActionResult {
            results[2] = 3
            return ActionResultOK
        }),
    }

    batch := Batch(actions...)
    result := batch.Execute(nil)

    assert.Equal(t, ActionResultOK, result)
    assert.Equal(t, []int{1, 2, 3}, results)
}

func TestSequence(t *testing.T) {
    order := make([]int, 0)
    actions := []Action{
        NewAction(func(ctx *ActionContext) ActionResult {
            order = append(order, 1)
            return ActionResultOK
        }),
        NewAction(func(ctx *ActionContext) ActionResult {
            order = append(order, 2)
            return ActionResultOK
        }),
    }

    seq := Sequence(actions...)
    seq.Execute(nil)

    assert.Equal(t, []int{1, 2}, order)
}

func TestParallelWithLimit(t *testing.T) {
    var concurrent int32
    var maxConcurrent int32

    actions := make([]Action, 10)
    for i := range actions {
        actions[i] = NewAction(func(ctx *ActionContext) ActionResult {
            n := atomic.AddInt32(&concurrent, 1)
            for {
                m := atomic.LoadInt32(&maxConcurrent)
                if n <= m || atomic.CompareAndSwapInt32(&maxConcurrent, m, n) {
                    break
                }
            }
            time.Sleep(50 * time.Millisecond)
            atomic.AddInt32(&concurrent, -1)
            return ActionResultOK
        })
    }

    limit := ParallelWithLimit(3, actions...)
    limit.Execute(nil)

    assert.LessOrEqual(t, maxConcurrent, int32(3))
}
```

## 性能考虑

1. **零抽象成本** - Batch/Sequence 只是语法糖
2. **Goroutine 池** - 避免频繁创建销毁
3. **内存复用** - 结果预分配
4. **上下文传递** - 支持取消传播

## 与 Bubble Tea 的差异

| 特性 | Bubble Tea | Yao TUI Composite Action |
|------|-----------|-------------------------|
| **抽象层级** | Cmd 层 | Action 层 |
| **并发控制** | 无限制 | WorkerPool 限制 |
| **错误处理** | 简单 | 多种策略 |
| **取消支持** | context | context + Cancelable |
| **回调** | 无 | WithCallback |
| **可组合性** | 可嵌套 | 可嵌套 |

## 向后兼容

```go
// 现有单 Action 代码无需修改
func (c *Component) LoadData() Action {
    return NewAction(func(ctx *ActionContext) ActionResult {
        // ...
        return ActionResultOK
    })
}

// 可以渐进式迁移到复合 Action
func (c *Component) LoadMultiple() Action {
    return Batch(
        c.loadData1(),
        c.loadData2(),
    )
}
```
