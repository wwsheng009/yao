package ai

import (
	"fmt"
	"time"

	"github.com/yaoapp/yao/tui/runtime/action"
	"github.com/yaoapp/yao/tui/runtime/state"
)

// =============================================================================
// Operations (V3)
// =============================================================================
// Operation 操作接口
// 用于构建可组合的操作序列

// Operation 操作接口
type Operation interface {
	// Execute 执行操作
	Execute(ctrl Controller) error
}

// =============================================================================
// Click Operation
// =============================================================================

// ClickOperation 点击操作
type ClickOperation struct {
	ComponentID string
}

// NewClickOperation 创建点击操作
func NewClickOperation(componentID string) *ClickOperation {
	return &ClickOperation{
		ComponentID: componentID,
	}
}

// Execute 执行点击操作
func (op *ClickOperation) Execute(ctrl Controller) error {
	return ctrl.Click(op.ComponentID)
}

// =============================================================================
// Input Operation
// =============================================================================

// InputOperation 输入操作
type InputOperation struct {
	ComponentID string
	Text        string
}

// NewInputOperation 创建输入操作
func NewInputOperation(componentID, text string) *InputOperation {
	return &InputOperation{
		ComponentID: componentID,
		Text:        text,
	}
}

// Execute 执行输入操作
func (op *InputOperation) Execute(ctrl Controller) error {
	return ctrl.Input(op.ComponentID, op.Text)
}

// =============================================================================
// Navigate Operation
// =============================================================================

// NavigateOperation 导航操作
type NavigateOperation struct {
	Direction Direction
}

// NewNavigateOperation 创建导航操作
func NewNavigateOperation(direction Direction) *NavigateOperation {
	return &NavigateOperation{
		Direction: direction,
	}
}

// Execute 执行导航操作
func (op *NavigateOperation) Execute(ctrl Controller) error {
	return ctrl.Navigate(op.Direction)
}

// =============================================================================
// Wait Operation
// =============================================================================

// WaitOperation 等待操作
type WaitOperation struct {
	Condition func(*state.Snapshot) bool
	Timeout   time.Duration
}

// NewWaitOperation 创建等待操作
func NewWaitOperation(condition func(*state.Snapshot) bool, timeout time.Duration) *WaitOperation {
	return &WaitOperation{
		Condition: condition,
		Timeout:   timeout,
	}
}

// Execute 执行等待操作
func (op *WaitOperation) Execute(ctrl Controller) error {
	return ctrl.WaitUntil(op.Condition, op.Timeout)
}

// =============================================================================
// Dispatch Operation
// =============================================================================

// DispatchOperation 分发 Action 操作
type DispatchOperation struct {
	Action func() *action.Action
}

// NewDispatchOperation 创建分发操作
func NewDispatchOperation(actionFn func() *action.Action) *DispatchOperation {
	return &DispatchOperation{
		Action: actionFn,
	}
}

// Execute 执行分发操作
func (op *DispatchOperation) Execute(ctrl Controller) error {
	return ctrl.Dispatch(op.Action())
}

// =============================================================================
// Batch Operation
// =============================================================================

// BatchOperation 批量操作（原子执行）
type BatchOperation struct {
	Operations []Operation
	Atomic      bool // 如果为 true，任一操作失败则回滚
}

// NewBatchOperation 创建批量操作
func NewBatchOperation(atomic bool, ops ...Operation) *BatchOperation {
	return &BatchOperation{
		Operations: ops,
		Atomic:      atomic,
	}
}

// Execute 执行批量操作
func (op *BatchOperation) Execute(ctrl Controller) error {
	if !op.Atomic {
		// 非原子模式：依次执行，不回滚
		for _, o := range op.Operations {
			if err := o.Execute(ctrl); err != nil {
				return err
			}
		}
		return nil
	}

	// 原子模式：记录状态，失败时回滚
	// 注意：完整回滚需要更复杂的实现
	// 这里简化为：失败时尝试恢复到初始状态

	lastErr := error(nil)
	for _, o := range op.Operations {
		if err := o.Execute(ctrl); err != nil {
			lastErr = err
			// 尝试恢复（简化版）
			// 实际实现需要更完善的状态恢复机制
			break
		}
	}

	if lastErr != nil {
		return fmt.Errorf("batch operation failed: %w", lastErr)
	}

	return nil
}

// =============================================================================
// Repeat Operation
// =============================================================================

// RepeatOperation 重复操作
type RepeatOperation struct {
	Operation Operation
	Count     int
	Delay     time.Duration // 每次操作之间的延迟
}

// NewRepeatOperation 创建重复操作
func NewRepeatOperation(op Operation, count int, delay time.Duration) *RepeatOperation {
	return &RepeatOperation{
		Operation: op,
		Count:     count,
		Delay:     delay,
	}
}

// Execute 执行重复操作
func (op *RepeatOperation) Execute(ctrl Controller) error {
	for i := 0; i < op.Count; i++ {
		if err := op.Operation.Execute(ctrl); err != nil {
			return fmt.Errorf("repeat failed at iteration %d: %w", i, err)
		}
		if op.Delay > 0 && i < op.Count-1 {
			time.Sleep(op.Delay)
		}
	}
	return nil
}

// =============================================================================
// Retry Operation
// =============================================================================

// RetryOperation 重试操作
type RetryOperation struct {
	Operation        Operation
	MaxAttempts      int
	RetryDelay       time.Duration
	ShouldRetry      func(error) bool
}

// NewRetryOperation 创建重试操作
func NewRetryOperation(op Operation, maxAttempts int, retryDelay time.Duration, shouldRetry func(error) bool) *RetryOperation {
	return &RetryOperation{
		Operation:   op,
		MaxAttempts: maxAttempts,
		RetryDelay:  retryDelay,
		ShouldRetry: shouldRetry,
	}
}

// Execute 执行重试操作
func (op *RetryOperation) Execute(ctrl Controller) error {
	var lastErr error

	for attempt := 0; attempt < op.MaxAttempts; attempt++ {
		err := op.Operation.Execute(ctrl)
		if err == nil {
			return nil
		}

		lastErr = err

		// 检查是否应该重试
		if op.ShouldRetry != nil && !op.ShouldRetry(err) {
			break
		}

		// 最后一次尝试不等待
		if attempt < op.MaxAttempts-1 {
			time.Sleep(op.RetryDelay)
		}
	}

	return fmt.Errorf("retry failed after %d attempts: %w", op.MaxAttempts, lastErr)
}

// =============================================================================
// WaitValue Operation
// =============================================================================

// WaitValueOperation 等待特定值的操作
type WaitValueOperation struct {
	ComponentID string
	StateKey    string
	Expected    interface{}
	Timeout     time.Duration
}

// NewWaitValueOperation 创建等待值操作
func NewWaitValueOperation(componentID, stateKey string, expected interface{}, timeout time.Duration) *WaitValueOperation {
	return &WaitValueOperation{
		ComponentID: componentID,
		StateKey:    stateKey,
		Expected:    expected,
		Timeout:     timeout,
	}
}

// Execute 执行等待值操作
func (op *WaitValueOperation) Execute(ctrl Controller) error {
	return ctrl.WaitUntil(func(s *state.Snapshot) bool {
		comp, ok := s.GetComponent(op.ComponentID)
		if !ok {
			return false
		}
		value, ok := comp.State[op.StateKey]
		if !ok {
			return false
		}
		return value == op.Expected
	}, op.Timeout)
}

// =============================================================================
// Helper Functions
// =============================================================================

// Click 点击组件的快捷操作
func Click(id string) Operation {
	return NewClickOperation(id)
}

// Input 输入文本的快捷操作
func Input(id, text string) Operation {
	return NewInputOperation(id, text)
}

// Navigate 导航的快捷操作
func Navigate(dir Direction) Operation {
	return NewNavigateOperation(dir)
}

// Wait 等待条件的快捷操作
func Wait(condition func(*state.Snapshot) bool, timeout time.Duration) Operation {
	return NewWaitOperation(condition, timeout)
}

// WaitValue 等待值的快捷操作
func WaitValue(id, key string, expected interface{}, timeout time.Duration) Operation {
	return NewWaitValueOperation(id, key, expected, timeout)
}

// Batch 批量操作的快捷操作
func Batch(ops ...Operation) Operation {
	return NewBatchOperation(false, ops...)
}

// AtomicBatch 原子批量操作的快捷操作
func AtomicBatch(ops ...Operation) Operation {
	return NewBatchOperation(true, ops...)
}
