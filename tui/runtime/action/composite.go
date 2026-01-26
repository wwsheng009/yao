package action

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ==============================================================================
// Composite Action System (V3)
// ==============================================================================
// 支持批量并发执行和顺序执行多个 Action

// Mode 执行模式
type Mode int

const (
	// Sequential 顺序执行
	Sequential Mode = iota
	// Concurrent 并发执行
	Concurrent
)

// ActionResult Action 执行结果
type ActionResult struct {
	OK      bool
	Error   error
	Message string
	Data    interface{}
}

// OKAction 成功的 Action 结果
var OKAction = ActionResult{OK: true}

// ErrorAction 错误的 Action 结果
func ErrorAction(err error) ActionResult {
	return ActionResult{OK: false, Error: err}
}

// MessageAction 带消息的 Action 结果
func MessageAction(msg string) ActionResult {
	return ActionResult{OK: true, Message: msg}
}

// DataAction 带数据的 Action 结果
func DataAction(data interface{}) ActionResult {
	return ActionResult{OK: true, Data: data}
}

// CompositeAction 复合 Action
type CompositeAction struct {
	mode     Mode
	actions  []ActionHandler
	callback CallbackFunc
	mu       sync.Mutex
	canceled atomic.Bool
}

// ActionHandler Action 处理器接口
type ActionHandler interface {
	Execute(ctx context.Context) ActionResult
}

// ActionFunc 函数式 Action
type ActionFunc func(ctx context.Context) ActionResult

// Execute 实现 ActionHandler 接口
func (f ActionFunc) Execute(ctx context.Context) ActionResult {
	return f(ctx)
}

// CallbackFunc 完成回调函数类型
type CallbackFunc func(results []ActionResult)

// NewCompositeAction 创建复合 Action
func NewCompositeAction(mode Mode, actions ...ActionHandler) *CompositeAction {
	return &CompositeAction{
		mode:    mode,
		actions: actions,
	}
}

// WithCallback 设置完成回调
func (a *CompositeAction) WithCallback(callback CallbackFunc) *CompositeAction {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.callback = callback
	return a
}

// Execute 执行复合 Action
func (a *CompositeAction) Execute(ctx context.Context) ActionResult {
	if a.canceled.Load() {
		return ActionResult{OK: false, Error: fmt.Errorf("action canceled")}
	}

	switch a.mode {
	case Concurrent:
		return a.executeConcurrent(ctx)
	case Sequential:
		return a.executeSequential(ctx)
	default:
		return ActionResult{OK: false, Error: fmt.Errorf("unknown mode: %d", a.mode)}
	}
}

// executeConcurrent 并发执行
func (a *CompositeAction) executeConcurrent(ctx context.Context) ActionResult {
	if len(a.actions) == 0 {
		return OKAction
	}

	var wg sync.WaitGroup
	results := make([]ActionResult, len(a.actions))
	errors := make(chan error, len(a.actions))

	for i, action := range a.actions {
		if a.canceled.Load() {
			break
		}

		wg.Add(1)
		go func(idx int, act ActionHandler) {
			defer wg.Done()

			// 检查上下文是否已取消
			if ctx.Err() != nil {
				results[idx] = ActionResult{OK: false, Error: ctx.Err()}
				return
			}

			result := act.Execute(ctx)
			results[idx] = result

			if err := result.Error; err != nil {
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
	a.mu.Lock()
	if a.callback != nil {
		a.callback(results)
	}
	a.mu.Unlock()

	if len(errs) > 0 {
		return ActionResult{
			OK:    false,
			Error: &MultipleError{Errors: errs},
		}
	}

	return OKAction
}

// executeSequential 顺序执行
func (a *CompositeAction) executeSequential(ctx context.Context) ActionResult {
	if len(a.actions) == 0 {
		return OKAction
	}

	results := make([]ActionResult, 0, len(a.actions))

	for _, action := range a.actions {
		// 检查是否已取消
		if a.canceled.Load() || ctx.Err() != nil {
			return ActionResult{OK: false, Error: fmt.Errorf("action canceled")}
		}

		result := action.Execute(ctx)
		results = append(results, result)

		// 如果有致命错误，中断执行
		if result.Error != nil && !isRecoverable(result.Error) {
			break
		}
	}

	// 调用回调
	a.mu.Lock()
	if a.callback != nil {
		a.callback(results)
	}
	a.mu.Unlock()

	return OKAction
}

// Cancel 取消执行
func (a *CompositeAction) Cancel() {
	a.canceled.Store(true)
}

// IsCanceled 检查是否已取消
func (a *CompositeAction) IsCanceled() bool {
	return a.canceled.Load()
}

// MultipleError 多错误包装
type MultipleError struct {
	Errors []error
}

func (e *MultipleError) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("%d errors occurred, first: %v", len(e.Errors), e.Errors[0])
}

// Unwrap 返回第一个错误
func (e *MultipleError) Unwrap() error {
	if len(e.Errors) > 0 {
		return e.Errors[0]
	}
	return nil
}

// ==============================================================================
// 便捷函数
// ==============================================================================

// Batch 并发执行多个 Action
func Batch(actions ...ActionHandler) *CompositeAction {
	return NewCompositeAction(Concurrent, actions...)
}

// BatchWithCallback 并发执行多个 Action，带回调
func BatchWithCallback(callback CallbackFunc, actions ...ActionHandler) *CompositeAction {
	return NewCompositeAction(Concurrent, actions...).WithCallback(callback)
}

// Sequence 顺序执行多个 Action
func Sequence(actions ...ActionHandler) *CompositeAction {
	return NewCompositeAction(Sequential, actions...)
}

// SequenceWithCallback 顺序执行多个 Action，带回调
func SequenceWithCallback(callback CallbackFunc, actions ...ActionHandler) *CompositeAction {
	return NewCompositeAction(Sequential, actions...).WithCallback(callback)
}

// ==============================================================================
// 辅助函数
// ==============================================================================

// isRecoverable 判断错误是否可恢复
func isRecoverable(err error) bool {
	if err == nil {
		return true
	}
	// 上下文取消错误是可恢复的
	if err == context.Canceled || err == context.DeadlineExceeded {
		return true
	}
	return false
}

// ActionFromActionFunc 将旧的 Action 函数转换为新的 ActionHandler
// 用于向后兼容
func ActionFromActionFunc(fn func() error) ActionHandler {
	return ActionFunc(func(ctx context.Context) ActionResult {
		if err := fn(); err != nil {
			return ErrorAction(err)
		}
		return OKAction
	})
}

// ActionFromActionResultFunc 将返回 ActionResult 的函数转换为 ActionHandler
func ActionFromActionResultFunc(fn func() ActionResult) ActionHandler {
	return ActionFunc(func(ctx context.Context) ActionResult {
		return fn()
	})
}

// ==============================================================================
// WorkerPool 工作池
// ==============================================================================

// WorkerPool 工作池，用于限制并发数量
type WorkerPool struct {
	maxWorkers int
	queue      chan ActionHandler
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewWorkerPool 创建工作池
func NewWorkerPool(maxWorkers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		maxWorkers: maxWorkers,
		queue:      make(chan ActionHandler, 100),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start 启动工作池
func (p *WorkerPool) Start() {
	for i := 0; i < p.maxWorkers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// worker 工作协程
func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		case action, ok := <-p.queue:
			if !ok {
				return
			}
			action.Execute(p.ctx)
		}
	}
}

// Submit 提交任务
func (p *WorkerPool) Submit(action ActionHandler) error {
	select {
	case p.queue <- action:
		return nil
	default:
		return fmt.Errorf("worker pool queue is full")
	}
}

// SubmitWithTimeout 带超时提交任务
func (p *WorkerPool) SubmitWithTimeout(action ActionHandler, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case p.queue <- action:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("submit timeout")
	}
}

// Stop 停止工作池
func (p *WorkerPool) Stop() {
	p.cancel()
	close(p.queue)
	p.wg.Wait()
}

// ParallelWithLimit 限制并发数的并行执行
func ParallelWithLimit(limit int, actions ...ActionHandler) ActionHandler {
	if len(actions) == 0 {
		return ActionFunc(func(ctx context.Context) ActionResult {
			return OKAction
		})
	}

	return ActionFunc(func(ctx context.Context) ActionResult {
		if len(actions) <= limit {
			// 如果动作数不超过限制，直接并发执行
			results := make([]ActionResult, len(actions))
			var wg sync.WaitGroup
			var mu sync.Mutex
			var hasError bool

			for i, action := range actions {
				if ctx.Err() != nil {
					break
				}

				wg.Add(1)
				go func(idx int, act ActionHandler) {
					defer wg.Done()

					result := act.Execute(ctx)
					mu.Lock()
					results[idx] = result
					if result.Error != nil {
						hasError = true
					}
					mu.Unlock()
				}(i, action)
			}

			wg.Wait()

			if hasError {
				return ActionResult{OK: false, Error: fmt.Errorf("some actions failed")}
			}
			return OKAction
		}

		// 使用信号量限制并发数
		sem := make(chan struct{}, limit)
		results := make([]ActionResult, len(actions))
		var wg sync.WaitGroup
		var mu sync.Mutex
		var hasError bool

		for i, action := range actions {
			if ctx.Err() != nil {
				break
			}

			wg.Add(1)
			go func(idx int, act ActionHandler) {
				defer wg.Done()

				sem <- struct{}{}        // 获取信号量
				defer func() { <-sem }() // 释放信号量

				result := act.Execute(ctx)
				mu.Lock()
				results[idx] = result
				if result.Error != nil {
					hasError = true
				}
				mu.Unlock()
			}(i, action)
		}

		wg.Wait()

		if hasError {
			return ActionResult{OK: false, Error: fmt.Errorf("some actions failed")}
		}

		return OKAction
	})
}

// RetryAction 重试 Action
type RetryAction struct {
	action     ActionHandler
	maxRetries int
	delay      time.Duration
}

// NewRetryAction 创建重试 Action
func NewRetryAction(action ActionHandler, maxRetries int, delay time.Duration) *RetryAction {
	return &RetryAction{
		action:     action,
		maxRetries: maxRetries,
		delay:      delay,
	}
}

// Execute 执行带重试的 Action
func (r *RetryAction) Execute(ctx context.Context) ActionResult {
	var lastErr error

	for i := 0; i <= r.maxRetries; i++ {
		if i > 0 && r.delay > 0 {
			select {
			case <-ctx.Done():
				return ActionResult{OK: false, Error: ctx.Err()}
			case <-time.After(r.delay):
			}
		}

		result := r.action.Execute(ctx)
		if result.OK {
			return result
		}

		lastErr = result.Error
	}

	return ActionResult{OK: false, Error: fmt.Errorf("after %d retries: %w", r.maxRetries, lastErr)}
}

// TimeoutAction 超时 Action
type TimeoutAction struct {
	action ActionHandler
	timeout time.Duration
}

// NewTimeoutAction 创建超时 Action
func NewTimeoutAction(action ActionHandler, timeout time.Duration) *TimeoutAction {
	return &TimeoutAction{
		action: action,
		timeout: timeout,
	}
}

// Execute 执行带超时的 Action
func (t *TimeoutAction) Execute(ctx context.Context) ActionResult {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	resultChan := make(chan ActionResult, 1)
	go func() {
		resultChan <- t.action.Execute(ctx)
	}()

	select {
	case result := <-resultChan:
		return result
	case <-ctx.Done():
		return ActionResult{OK: false, Error: fmt.Errorf("action timeout after %v", t.timeout)}
	}
}

// FallbackAction 带回退的 Action
type FallbackAction struct {
	primary   ActionHandler
	secondary ActionHandler
}

// NewFallbackAction 创建带回退的 Action
func NewFallbackAction(primary, secondary ActionHandler) *FallbackAction {
	return &FallbackAction{
		primary:   primary,
		secondary: secondary,
	}
}

// Execute 执行带回退的 Action
func (f *FallbackAction) Execute(ctx context.Context) ActionResult {
	result := f.primary.Execute(ctx)
	if !result.OK {
		// 主 Action 失败，执行回退 Action
		return f.secondary.Execute(ctx)
	}
	return result
}

// LazyAction 延迟执行的 Action
type LazyAction struct {
	factory func() ActionHandler
}

// NewLazyAction 创建延迟 Action
func NewLazyAction(factory func() ActionHandler) *LazyAction {
	return &LazyAction{factory: factory}
}

// Execute 执行延迟 Action
func (l *LazyAction) Execute(ctx context.Context) ActionResult {
	action := l.factory()
	return action.Execute(ctx)
}
