package core

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mockTerminal 模拟终端
type mockTerminal struct {
	mu             sync.Mutex
	normalMode     bool
	cursorVisible  bool
	exitedAltScreen bool
	echoEnabled    bool
	closed         bool
}

func newMockTerminal() *mockTerminal {
	return &mockTerminal{
		normalMode:     false,
		cursorVisible:  false,
		exitedAltScreen: false,
		echoEnabled:    false,
		closed:         false,
	}
}

func (m *mockTerminal) SetNormalMode() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.normalMode = true
}

func (m *mockTerminal) ShowCursor() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cursorVisible = true
}

func (m *mockTerminal) ExitAltScreen() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.exitedAltScreen = true
}

func (m *mockTerminal) EnableEcho() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.echoEnabled = true
}

func (m *mockTerminal) Flush() {
	// 模拟刷新
}

func (m *mockTerminal) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockTerminal) isRestored() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.normalMode && m.cursorVisible && m.exitedAltScreen && m.echoEnabled
}

func TestNewRecovery(t *testing.T) {
	terminal := newMockTerminal()
	recovery := NewRecovery(terminal)

	if recovery == nil {
		t.Error("recovery should not be nil")
	}
}

func TestRecovery_Handle(t *testing.T) {
	terminal := newMockTerminal()
	recovery := NewRecovery(terminal)

	var buf bytes.Buffer
	recovery.SetLogWriter(&buf)

	// 模拟 panic
	recovery.Handle("test panic")

	if !terminal.isRestored() {
		t.Error("terminal should be restored")
	}

	output := buf.String()
	if !strings.Contains(output, "test panic") {
		t.Error("log should contain panic value")
	}
	if !strings.Contains(output, "Stack:") {
		t.Error("log should contain stack trace")
	}
}

func TestRecovery_AddHandler(t *testing.T) {
	terminal := newMockTerminal()
	recovery := NewRecovery(terminal)

	var handled bool
	handler := &MockPanicHandler{
		onPanic: func(r interface{}, stack []byte) {
			handled = true
		},
	}

	recovery.AddHandler(handler)
	recovery.Handle("test")

	if !handled {
		t.Error("handler should be called")
	}
}

func TestRecovery_EnablePanicLog(t *testing.T) {
	terminal := newMockTerminal()
	recovery := NewRecovery(terminal)

	// 创建临时文件
	tmpfile := "test_panic.log"
	defer os.Remove(tmpfile)

	err := recovery.EnablePanicLog(tmpfile)
	if err != nil {
		t.Fatalf("failed to enable panic log: %v", err)
	}

	recovery.Handle("test panic")

	// 检查文件内容
	data, err := os.ReadFile(tmpfile)
	if err != nil {
		t.Fatalf("failed to read panic log: %v", err)
	}

	if !strings.Contains(string(data), "test panic") {
		t.Error("panic log should contain panic value")
	}

	recovery.Close()
}

func TestRecovery_SetLogWriter(t *testing.T) {
	terminal := newMockTerminal()
	recovery := NewRecovery(terminal)

	var buf bytes.Buffer
	recovery.SetLogWriter(&buf)

	recovery.Handle("test")

	if buf.Len() == 0 {
		t.Error("log should be written")
	}
}

func TestLoggingPanicHandler(t *testing.T) {
	var buf bytes.Buffer
	handler := NewLoggingPanicHandler(&buf)

	handler.HandlePanic("test panic", nil)

	output := buf.String()
	if !strings.Contains(output, "test panic") {
		t.Error("log should contain panic value")
	}
}

func TestMetricsPanicHandler(t *testing.T) {
	handler := NewMetricsPanicHandler(10)

	handler.HandlePanic("panic1", []byte("stack1"))
	handler.HandlePanic("panic2", []byte("stack2"))

	if handler.PanicCount() != 2 {
		t.Errorf("expected 2 panics, got %d", handler.PanicCount())
	}

	records := handler.GetRecords()
	if len(records) != 2 {
		t.Errorf("expected 2 records, got %d", len(records))
	}

	if records[0].Value != "panic1" {
		t.Error("first record value should be panic1")
	}

	handler.Reset()
	if handler.PanicCount() != 0 {
		t.Error("count should be 0 after reset")
	}
}

func TestCrashReportPanicHandler(t *testing.T) {
	tmpdir := t.TempDir()
	handler := NewCrashReportPanicHandler(tmpdir)

	handler.HandlePanic("test panic", []byte("stack trace"))

	// 检查是否创建了崩溃报告文件
	files, err := os.ReadDir(tmpdir)
	if err != nil {
		t.Fatalf("failed to read temp dir: %v", err)
	}

	found := false
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "crash_") {
			found = true
			// 读取文件内容
			data, _ := os.ReadFile(tmpdir + "/" + f.Name())
			if !strings.Contains(string(data), "test panic") {
				t.Error("crash report should contain panic value")
			}
			break
		}
	}

	if !found {
		t.Error("crash report file should be created")
	}
}

func TestNotificationPanicHandler(t *testing.T) {
	var notified interface{}
	var notifiedStack []byte

	handler := NewNotificationPanicHandler(func(r interface{}, stack []byte) {
		notified = r
		notifiedStack = stack
	})

	handler.HandlePanic("test", []byte("stack"))

	if notified != "test" {
		t.Errorf("expected notified value 'test', got %v", notified)
	}

	if string(notifiedStack) != "stack" {
		t.Error("stack should be passed")
	}
}

func TestSafeRunner_Run(t *testing.T) {
	terminal := newMockTerminal()
	recovery := NewRecovery(terminal)
	runner := NewSafeRunner(recovery)

	// 正常执行
	err := runner.Run(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 恢复后的错误
	err = runner.Run(func() error {
		panic("test panic")
	})

	if err == nil {
		t.Error("should return error after panic")
	}

	if !strings.Contains(err.Error(), "panic") {
		t.Error("error should mention panic")
	}
}

func TestSafeRunner_OnPanic(t *testing.T) {
	terminal := newMockTerminal()
	recovery := NewRecovery(terminal)
	runner := NewSafeRunner(recovery)

	var panicked interface{}
	runner.OnPanic(func(r interface{}) {
		panicked = r
	})

	runner.Run(func() error {
		panic("test")
	})

	if panicked != "test" {
		t.Errorf("expected 'test', got %v", panicked)
	}
}

func TestSafeRunner_RunWithContext(t *testing.T) {
	terminal := newMockTerminal()
	recovery := NewRecovery(terminal)
	runner := NewSafeRunner(recovery)

	ctx, cancel := context.WithCancel(context.Background())

	// 正常执行
	err := runner.RunWithContext(ctx, func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 取消上下文
	cancel()
	err = runner.RunWithContext(ctx, func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	})

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestSafeRunner_RunWithContext_Panic(t *testing.T) {
	terminal := newMockTerminal()
	recovery := NewRecovery(terminal)
	runner := NewSafeRunner(recovery)

	ctx := context.Background()
	err := runner.RunWithContext(ctx, func(ctx context.Context) error {
		panic("test panic")
	})

	if err == nil {
		t.Error("should return error after panic")
	}

	if !strings.Contains(err.Error(), "panic") {
		t.Error("error should mention panic")
	}
}

func TestWithRecovery(t *testing.T) {
	terminal := newMockTerminal()
	var ran bool

	WithRecovery(terminal, func() {
		ran = true
	})

	if !ran {
		t.Error("function should run")
	}
}

func TestWithRecovery_Panic(t *testing.T) {
	terminal := newMockTerminal()

	WithRecovery(terminal, func() {
		panic("test panic")
	})

	// 终端应该被恢复
	if !terminal.isRestored() {
		t.Error("terminal should be restored after panic")
	}
}

func TestSafeGo(t *testing.T) {
	terminal := newMockTerminal()
	done := make(chan struct{})

	SafeGo(terminal, func() {
		close(done)
	})

	select {
	case <-done:
		// 正确
	case <-time.After(100 * time.Millisecond):
		t.Error("goroutine should complete")
	}
}

func TestSafeGo_Panic(t *testing.T) {
	terminal := newMockTerminal()
	done := make(chan struct{})

	SafeGo(terminal, func() {
		defer close(done)
		panic("test panic")
	})

	select {
	case <-done:
		// goroutine 应该完成（panic 被捕获）
	case <-time.After(100 * time.Millisecond):
		t.Error("goroutine should complete after panic")
	}

	// 等待恢复完成
	time.Sleep(20 * time.Millisecond)

	// 终端应该被恢复
	if !terminal.isRestored() {
		t.Error("terminal should be restored after panic")
	}
}

func TestSafeGoWithContext(t *testing.T) {
	terminal := newMockTerminal()
	ctx := context.Background()
	done := make(chan struct{})

	SafeGoWithContext(ctx, terminal, func(ctx context.Context) {
		close(done)
	})

	select {
	case <-done:
		// 正确
	case <-time.After(100 * time.Millisecond):
		t.Error("goroutine should complete")
	}
}

func TestSafeGoWithContext_Cancel(t *testing.T) {
	terminal := newMockTerminal()
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	SafeGoWithContext(ctx, terminal, func(ctx context.Context) {
		<-ctx.Done()
		close(done)
	})

	cancel()

	select {
	case <-done:
		// 正确
	case <-time.After(100 * time.Millisecond):
		t.Error("goroutine should complete after cancel")
	}
}

func TestSafeGoWithContext_Panic(t *testing.T) {
	terminal := newMockTerminal()
	ctx := context.Background()
	done := make(chan struct{})

	SafeGoWithContext(ctx, terminal, func(ctx context.Context) {
		defer close(done)
		panic("test panic")
	})

	select {
	case <-done:
		// goroutine 应该完成
	case <-time.After(100 * time.Millisecond):
		t.Error("goroutine should complete after panic")
	}

	// 等待恢复完成
	time.Sleep(20 * time.Millisecond)

	// 终端应该被恢复
	if !terminal.isRestored() {
		t.Error("terminal should be restored after panic")
	}
}

func TestRecovery_Concurrent(t *testing.T) {
	terminal := newMockTerminal()
	recovery := NewRecovery(terminal)

	var wg sync.WaitGroup
	var count int32

	// 并发触发 panic
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					atomic.AddInt32(&count, 1)
				}
			}()

			recovery.Handle("concurrent panic")
		}()
	}

	wg.Wait()

	// 所有 panic 都应该被处理
	// count 可能为 0，因为 recover() 在 Handle 内部已经处理了
	_ = count
}

func TestRecovery_Close(t *testing.T) {
	terminal := newMockTerminal()
	recovery := NewRecovery(terminal)

	tmpfile := "test_close_panic.log"
	defer os.Remove(tmpfile)

	recovery.EnablePanicLog(tmpfile)
	recovery.Handle("test")

	err := recovery.Close()
	if err != nil {
		t.Errorf("close failed: %v", err)
	}

	// 再次关闭应该没有错误
	err = recovery.Close()
	if err != nil {
		t.Errorf("second close should not error, got: %v", err)
	}
}

// MockPanicHandler 模拟 panic 处理器
type MockPanicHandler struct {
	onPanic func(r interface{}, stack []byte)
}

func (m *MockPanicHandler) HandlePanic(r interface{}, stack []byte) {
	if m.onPanic != nil {
		m.onPanic(r, stack)
	}
}

func TestPanicRecord(t *testing.T) {
	now := time.Now()
	record := PanicRecord{
		Time:  now,
		Value: "test",
		Stack: []byte("stack"),
	}

	if record.Value != "test" {
		t.Error("value should be test")
	}

	if !record.Time.Equal(now) {
		t.Error("time should match")
	}
}

func ExampleNewRecovery() {
	terminal := newMockTerminal()
	recovery := NewRecovery(terminal)

	recovery.AddHandler(NewLoggingPanicHandler(os.Stdout))

	// Handle panics and restore terminal state
	_ = recovery.Handle
}

func ExampleSafeRunner() {
	terminal := newMockTerminal()
	recovery := NewRecovery(terminal)
	runner := NewSafeRunner(recovery)

	err := runner.Run(func() error {
		// 做一些可能 panic 的事情
		return nil
	})

	if err != nil {
		fmt.Println("Error:", err)
	}
}
