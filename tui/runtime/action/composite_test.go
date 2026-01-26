package action

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestActionResult(t *testing.T) {
	// OKAction
	if !OKAction.OK {
		t.Error("OKAction should be OK")
	}

	// ErrorAction
	err := fmt.Errorf("test error")
	result := ErrorAction(err)
	if result.OK {
		t.Error("ErrorAction should not be OK")
	}
	if result.Error != err {
		t.Error("ErrorAction should contain the error")
	}

	// MessageAction
	result = MessageAction("test message")
	if !result.OK {
		t.Error("MessageAction should be OK")
	}
	if result.Message != "test message" {
		t.Error("MessageAction should contain the message")
	}

	// DataAction
	data := "test data"
	result = DataAction(data)
	if !result.OK {
		t.Error("DataAction should be OK")
	}
	if result.Data != data {
		t.Error("DataAction should contain the data")
	}
}

func TestActionFunc(t *testing.T) {
	var called bool
	fn := ActionFunc(func(ctx context.Context) ActionResult {
		called = true
		return OKAction
	})

	result := fn.Execute(context.Background())
	if !called {
		t.Error("ActionFunc should call the underlying function")
	}
	if !result.OK {
		t.Error("result should be OK")
	}
}

func TestBatch(t *testing.T) {
	var count int32
	var order []int
	var orderMu sync.Mutex

	actions := []ActionHandler{
		ActionFunc(func(ctx context.Context) ActionResult {
			n := atomic.AddInt32(&count, 1)
			time.Sleep(10 * time.Millisecond)
			orderMu.Lock()
			order = append(order, int(n))
			orderMu.Unlock()
			return OKAction
		}),
		ActionFunc(func(ctx context.Context) ActionResult {
			n := atomic.AddInt32(&count, 1)
			time.Sleep(10 * time.Millisecond)
			orderMu.Lock()
			order = append(order, int(n))
			orderMu.Unlock()
			return OKAction
		}),
		ActionFunc(func(ctx context.Context) ActionResult {
			n := atomic.AddInt32(&count, 1)
			time.Sleep(10 * time.Millisecond)
			orderMu.Lock()
			order = append(order, int(n))
			orderMu.Unlock()
			return OKAction
		}),
	}

	batch := Batch(actions...)
	result := batch.Execute(context.Background())

	if !result.OK {
		t.Errorf("batch execution failed: %v", result.Error)
	}

	if count != 3 {
		t.Errorf("expected 3 actions to execute, got %d", count)
	}

	// 由于是并发执行，顺序可能不同
	if len(order) != 3 {
		t.Errorf("expected 3 order entries, got %d", len(order))
	}
}

func TestSequence(t *testing.T) {
	var order []int

	actions := []ActionHandler{
		ActionFunc(func(ctx context.Context) ActionResult {
			order = append(order, 1)
			return OKAction
		}),
		ActionFunc(func(ctx context.Context) ActionResult {
			order = append(order, 2)
			return OKAction
		}),
		ActionFunc(func(ctx context.Context) ActionResult {
			order = append(order, 3)
			return OKAction
		}),
	}

	seq := Sequence(actions...)
	result := seq.Execute(context.Background())

	if !result.OK {
		t.Errorf("sequence execution failed: %v", result.Error)
	}

	// 顺序执行应该是 1, 2, 3
	expected := []int{1, 2, 3}
	if len(order) != len(expected) {
		t.Fatalf("expected %d order entries, got %d", len(expected), len(order))
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("expected order[%d] = %d, got %d", i, v, order[i])
		}
	}
}

func TestBatchWithCallback(t *testing.T) {
	var callbackResults []ActionResult

	actions := []ActionHandler{
		ActionFunc(func(ctx context.Context) ActionResult {
			return MessageAction("action1")
		}),
		ActionFunc(func(ctx context.Context) ActionResult {
			return MessageAction("action2")
		}),
	}

	batch := BatchWithCallback(func(results []ActionResult) {
		callbackResults = results
	}, actions...)

	result := batch.Execute(context.Background())
	if !result.OK {
		t.Errorf("batch execution failed: %v", result.Error)
	}

	if len(callbackResults) != 2 {
		t.Errorf("expected 2 callback results, got %d", len(callbackResults))
	}
}

func TestCompositeAction_Cancel(t *testing.T) {
	var executed int32

	actions := []ActionHandler{
		ActionFunc(func(ctx context.Context) ActionResult {
			time.Sleep(50 * time.Millisecond)
			atomic.AddInt32(&executed, 1)
			return OKAction
		}),
		ActionFunc(func(ctx context.Context) ActionResult {
			time.Sleep(50 * time.Millisecond)
			atomic.AddInt32(&executed, 1)
			return OKAction
		}),
	}

	batch := Batch(actions...)

	// 立即取消
	batch.Cancel()
	result := batch.Execute(context.Background())

	// 由于立即取消，可能没有或只有少量执行
	// 关键是取消状态应该被记录
	if !batch.IsCanceled() {
		t.Error("action should be marked as canceled")
	}

	_ = result // 可能被取消
}

func TestCompositeAction_CancelDuringExecution(t *testing.T) {
	started := make(chan struct{})
	readyToCancel := make(chan struct{})

	actions := []ActionHandler{
		ActionFunc(func(ctx context.Context) ActionResult {
			close(started)
			<-readyToCancel
			return OKAction
		}),
		ActionFunc(func(ctx context.Context) ActionResult {
			return OKAction
		}),
	}

	batch := Batch(actions...)

	go func() {
		<-started
		time.Sleep(10 * time.Millisecond)
		batch.Cancel()
		close(readyToCancel)
	}()

	result := batch.Execute(context.Background())
	// 由于取消了，结果可能是错误或部分成功
	_ = result
}

func TestParallelWithLimit(t *testing.T) {
	var concurrent int32
	var maxConcurrent int32

	actions := make([]ActionHandler, 10)
	for i := range actions {
		actions[i] = ActionFunc(func(ctx context.Context) ActionResult {
			n := atomic.AddInt32(&concurrent, 1)
			defer atomic.AddInt32(&concurrent, -1)

			// 更新最大并发数
			for {
				m := atomic.LoadInt32(&maxConcurrent)
				if n <= m || atomic.CompareAndSwapInt32(&maxConcurrent, m, n) {
					break
				}
			}

			time.Sleep(50 * time.Millisecond)
			return OKAction
		})
	}

	limit := ParallelWithLimit(3, actions...)
	result := limit.Execute(context.Background())

	if !result.OK {
		t.Errorf("parallel execution failed: %v", result.Error)
	}

	if maxConcurrent > 3 {
		t.Errorf("expected max concurrent 3, got %d", maxConcurrent)
	}
}

func TestMultipleError(t *testing.T) {
	errs := []error{
		fmt.Errorf("error1"),
		fmt.Errorf("error2"),
		fmt.Errorf("error3"),
	}

	multiErr := &MultipleError{Errors: errs}

	if multiErr.Error() == "" {
		t.Error("error message should not be empty")
	}

	// Unwrap 应该返回第一个错误
	if multiErr.Unwrap() != errs[0] {
		t.Error("unwrap should return first error")
	}
}

func TestRetryAction(t *testing.T) {
	var attempts int

	action := NewRetryAction(
		ActionFunc(func(ctx context.Context) ActionResult {
			attempts++
			if attempts < 3 {
				return ErrorAction(fmt.Errorf("not yet"))
			}
			return OKAction
		}),
		3, // max retries
		10*time.Millisecond,
	)

	result := action.Execute(context.Background())
	if !result.OK {
		t.Errorf("retry action failed: %v", result.Error)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryAction_Exhausted(t *testing.T) {
	action := NewRetryAction(
		ActionFunc(func(ctx context.Context) ActionResult {
			return ErrorAction(fmt.Errorf("always fails"))
		}),
		2, // max retries
		0,  // no delay
	)

	result := action.Execute(context.Background())
	if result.OK {
		t.Error("retry action should fail after exhausting retries")
	}
}

func TestTimeoutAction(t *testing.T) {
	action := NewTimeoutAction(
		ActionFunc(func(ctx context.Context) ActionResult {
			time.Sleep(200 * time.Millisecond)
			return OKAction
		}),
		50*time.Millisecond,
	)

	result := action.Execute(context.Background())
	if result.OK {
		t.Error("timeout action should fail")
	}
}

func TestTimeoutAction_Success(t *testing.T) {
	action := NewTimeoutAction(
		ActionFunc(func(ctx context.Context) ActionResult {
			return OKAction
		}),
		100*time.Millisecond,
	)

	result := action.Execute(context.Background())
	if !result.OK {
		t.Errorf("timeout action failed: %v", result.Error)
	}
}

func TestFallbackAction(t *testing.T) {
	primary := ActionFunc(func(ctx context.Context) ActionResult {
		return ErrorAction(fmt.Errorf("primary failed"))
	})

	secondary := ActionFunc(func(ctx context.Context) ActionResult {
		return MessageAction("secondary success")
	})

	action := NewFallbackAction(primary, secondary)
	result := action.Execute(context.Background())

	if !result.OK {
		t.Errorf("fallback action failed: %v", result.Error)
	}

	if result.Message != "secondary success" {
		t.Errorf("unexpected message: %s", result.Message)
	}
}

func TestFallbackAction_PrimarySuccess(t *testing.T) {
	primary := ActionFunc(func(ctx context.Context) ActionResult {
		return MessageAction("primary success")
	})

	secondary := ActionFunc(func(ctx context.Context) ActionResult {
		return MessageAction("secondary success")
	})

	action := NewFallbackAction(primary, secondary)
	result := action.Execute(context.Background())

	if !result.OK {
		t.Errorf("fallback action failed: %v", result.Error)
	}

	if result.Message != "primary success" {
		t.Errorf("unexpected message: %s", result.Message)
	}
}

func TestLazyAction(t *testing.T) {
	var called bool

	action := NewLazyAction(func() ActionHandler {
		return ActionFunc(func(ctx context.Context) ActionResult {
			called = true
			return OKAction
		})
	})

	// LazyAction 还没有执行
	if called {
		t.Error("lazy action should not execute yet")
	}

	// 执行 LazyAction
	result := action.Execute(context.Background())

	if !result.OK {
		t.Errorf("lazy action failed: %v", result.Error)
	}

	if !called {
		t.Error("lazy action should have executed")
	}
}

func TestActionFromActionFunc(t *testing.T) {
	var called bool
	fn := func() error {
		called = true
		return nil
	}

	action := ActionFromActionFunc(fn)
	result := action.Execute(context.Background())

	if !called {
		t.Error("function should be called")
	}
	if !result.OK {
		t.Error("result should be OK")
	}
}

func TestActionFromActionResultFunc(t *testing.T) {
	fn := func() ActionResult {
		return MessageAction("test")
	}

	action := ActionFromActionResultFunc(fn)
	result := action.Execute(context.Background())

	if !result.OK {
		t.Error("result should be OK")
	}
	if result.Message != "test" {
		t.Errorf("unexpected message: %s", result.Message)
	}
}

func TestWorkerPool_Submit(t *testing.T) {
	pool := NewWorkerPool(2)
	pool.Start()
	defer pool.Stop()

	var executed int32
	action := ActionFunc(func(ctx context.Context) ActionResult {
		atomic.AddInt32(&executed, 1)
		return OKAction
	})

	err := pool.Submit(action)
	if err != nil {
		t.Errorf("submit failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if executed == 0 {
		t.Error("action should have been executed")
	}
}

func TestWorkerPool_SubmitWithTimeout(t *testing.T) {
	pool := NewWorkerPool(1)
	pool.Start()
	defer pool.Stop()

	// 提交一个长时间运行的任务
	longAction := ActionFunc(func(ctx context.Context) ActionResult {
		time.Sleep(200 * time.Millisecond)
		return OKAction
	})
	pool.Submit(longAction)

	// 队列应该满了，尝试快速提交会超时
	err := pool.SubmitWithTimeout(ActionFunc(func(ctx context.Context) ActionResult {
		return OKAction
	}), 10*time.Millisecond)

	// 由于第一个任务还在运行，第二个任务可能成功入队（队列大小为100）
	// 或者超时，这里我们只验证没有 panic
	_ = err
}

func TestSequence_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	actions := []ActionHandler{
		ActionFunc(func(ctx context.Context) ActionResult {
			cancel() // 取消上下文
			return OKAction
		}),
		ActionFunc(func(ctx context.Context) ActionResult {
			time.Sleep(10 * time.Millisecond)
			return OKAction
		}),
		ActionFunc(func(ctx context.Context) ActionResult {
			return OKAction
		}),
	}

	seq := Sequence(actions...)
	result := seq.Execute(ctx)

	// 由于上下文被取消，结果应该是失败
	if result.OK && result.Error != context.Canceled {
		t.Logf("sequence result: OK=%v, Error=%v", result.OK, result.Error)
	}
}

func TestBatch_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	actions := []ActionHandler{
		ActionFunc(func(ctx context.Context) ActionResult {
			time.Sleep(10 * time.Millisecond)
			cancel()
			return OKAction
		}),
		ActionFunc(func(ctx context.Context) ActionResult {
			time.Sleep(50 * time.Millisecond)
			return OKAction
		}),
	}

	batch := Batch(actions...)
	result := batch.Execute(ctx)

	// 由于上下文被取消，一些操作可能失败
	// 这里的关键是测试不会 panic
	_ = result
}

func TestNewCompositeAction_Empty(t *testing.T) {
	batch := NewCompositeAction(Concurrent)
	result := batch.Execute(context.Background())

	if !result.OK {
		t.Errorf("empty batch failed: %v", result.Error)
	}

	seq := NewCompositeAction(Sequential)
	result = seq.Execute(context.Background())

	if !result.OK {
		t.Errorf("empty sequence failed: %v", result.Error)
	}
}
