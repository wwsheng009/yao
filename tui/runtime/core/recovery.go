package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// ==============================================================================
// Panic Recovery System (V3)
// ==============================================================================
// 恐慌恢复管理器，确保应用在发生不可恢复错误时能够正确清理资源

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
	logWriter    io.Writer
}

// Terminal 终端接口
type Terminal interface {
	SetNormalMode()
	ShowCursor()
	ExitAltScreen()
	EnableEcho()
	Flush()
	Close() error
}

// NewRecovery 创建恢复管理器
func NewRecovery(terminal Terminal) *Recovery {
	return &Recovery{
		terminal:  terminal,
		handlers: make([]PanicHandler, 0),
		logWriter: os.Stderr,
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

		// 关闭终端
		r.terminal.Close()
	}
}

// logPanic 记录 panic
func (r *Recovery) logPanic(panicValue interface{}, stack []byte) {
	msg := fmt.Sprintf("\n\n=== PANIC ===\nValue: %v\n\nStack:\n%s\n\n",
		panicValue, stack)

	// 输出到日志写入器
	if r.logWriter != nil {
		r.logWriter.Write([]byte(msg))
	}

	// 写入日志文件
	if r.panicLogFile != nil {
		r.panicLogFile.WriteString(msg)
		r.panicLogFile.Sync()
	}
}

// EnablePanicLog 启用 panic 日志文件
func (r *Recovery) EnablePanicLog(filename string) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	if r.panicLogFile != nil {
		r.panicLogFile.Close()
	}

	r.panicLogFile = f
	return nil
}

// SetLogWriter 设置日志写入器
func (r *Recovery) SetLogWriter(w io.Writer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logWriter = w
}

// Close 关闭恢复管理器
func (r *Recovery) Close() error {
	if r.panicLogFile != nil {
		err := r.panicLogFile.Close()
		r.panicLogFile = nil
		return err
	}
	return nil
}

// ==============================================================================
// 内置处理器
// ==============================================================================

// LoggingPanicHandler 日志处理器
type LoggingPanicHandler struct {
	writer io.Writer
	prefix string
}

// NewLoggingPanicHandler 创建日志处理器
func NewLoggingPanicHandler(w io.Writer) *LoggingPanicHandler {
	return &LoggingPanicHandler{
		writer: w,
		prefix: "[PANIC] ",
	}
}

// HandlePanic 处理 panic
func (h *LoggingPanicHandler) HandlePanic(r interface{}, stack []byte) {
	msg := fmt.Sprintf("%sPanic: %v\n%s\n", h.prefix, r, stack)
	if h.writer != nil {
		h.writer.Write([]byte(msg))
	}
}

// MetricsPanicHandler 指标处理器
type MetricsPanicHandler struct {
	mu        sync.Mutex
	panicCount int
	lastPanic  time.Time
	panics     []PanicRecord
	maxRecords int
}

// PanicRecord panic 记录
type PanicRecord struct {
	Time  time.Time
	Value interface{}
	Stack []byte
}

// NewMetricsPanicHandler 创建指标处理器
func NewMetricsPanicHandler(maxRecords int) *MetricsPanicHandler {
	return &MetricsPanicHandler{
		panics:     make([]PanicRecord, 0, maxRecords),
		maxRecords: maxRecords,
	}
}

// HandlePanic 处理 panic
func (h *MetricsPanicHandler) HandlePanic(r interface{}, stack []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.panicCount++
	h.lastPanic = time.Now()

	record := PanicRecord{
		Time:  time.Now(),
		Value: r,
		Stack: stack,
	}

	// 保留最近的记录
	if len(h.panics) >= h.maxRecords {
		h.panics = h.panics[1:]
	}
	h.panics = append(h.panics, record)
}

// PanicCount 返回 panic 次数
func (h *MetricsPanicHandler) PanicCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.panicCount
}

// LastPanic 返回上次 panic 时间
func (h *MetricsPanicHandler) LastPanic() time.Time {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.lastPanic
}

// GetRecords 返回 panic 记录
func (h *MetricsPanicHandler) GetRecords() []PanicRecord {
	h.mu.Lock()
	defer h.mu.Unlock()

	records := make([]PanicRecord, len(h.panics))
	copy(records, h.panics)
	return records
}

// Reset 重置指标
func (h *MetricsPanicHandler) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.panicCount = 0
	h.lastPanic = time.Time{}
	h.panics = make([]PanicRecord, 0, h.maxRecords)
}

// CrashReportPanicHandler 崩溃报告处理器
type CrashReportPanicHandler struct {
	reportDir string
}

// NewCrashReportPanicHandler 创建崩溃报告处理器
func NewCrashReportPanicHandler(dir string) *CrashReportPanicHandler {
	return &CrashReportPanicHandler{reportDir: dir}
}

// HandlePanic 处理 panic
func (h *CrashReportPanicHandler) HandlePanic(r interface{}, stack []byte) {
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
	f.WriteString(fmt.Sprintf("Time: %s\n", time.Now().Format(time.RFC3339)))
}

// NotificationPanicHandler 通知处理器
type NotificationPanicHandler struct {
	notifier func(panicValue interface{}, stack []byte)
}

// NewNotificationPanicHandler 创建通知处理器
func NewNotificationPanicHandler(fn func(interface{}, []byte)) *NotificationPanicHandler {
	return &NotificationPanicHandler{notifier: fn}
}

// HandlePanic 处理 panic
func (h *NotificationPanicHandler) HandlePanic(r interface{}, stack []byte) {
	if h.notifier != nil {
		h.notifier(r, stack)
	}
}

// ==============================================================================
// SafeRunner 安全运行器
// ==============================================================================

// SafeFunc 安全函数类型
type SafeFunc func() error

// SafeRunner 安全运行器
type SafeRunner struct {
	recovery *Recovery
	onPanic  func(interface{})
}

// NewSafeRunner 创建安全运行器
func NewSafeRunner(recovery *Recovery) *SafeRunner {
	return &SafeRunner{
		recovery: recovery,
	}
}

// Run 安全运行函数
func (s *SafeRunner) Run(fn SafeFunc) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if s.recovery != nil {
				s.recovery.Handle(r)
			}
			if s.onPanic != nil {
				s.onPanic(r)
			}
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	return fn()
}

// OnPanic 设置 panic 回调
func (s *SafeRunner) OnPanic(fn func(interface{})) {
	s.onPanic = fn
}

// RunWithContext 在上下文中安全运行函数
func (s *SafeRunner) RunWithContext(ctx context.Context, fn func(context.Context) error) error {
	done := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				if s.recovery != nil {
					s.recovery.Handle(r)
				}
				done <- fmt.Errorf("panic: %v", r)
			}
		}()

		done <- fn(ctx)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ==============================================================================
// 便捷函数
// ==============================================================================

// RecoverPanic 捕获 panic 并处理
func RecoverPanic(terminal Terminal) {
	if r := recover(); r != nil {
		recovery := NewRecovery(terminal)
		recovery.Handle(r)
	}
}

// WithRecovery 在函数中添加 panic 恢复
func WithRecovery(terminal Terminal, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			recovery := NewRecovery(terminal)
			recovery.Handle(r)
		}
	}()

	fn()
}

// SafeGo 安全启动 goroutine
func SafeGo(terminal Terminal, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				recovery := NewRecovery(terminal)
				recovery.Handle(r)
			}
		}()

		fn()
	}()
}

// SafeGoWithContext 安全启动带上下文的 goroutine
func SafeGoWithContext(ctx context.Context, terminal Terminal, fn func(context.Context)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				recovery := NewRecovery(terminal)
				recovery.Handle(r)
			}
		}()

		fn(ctx)
	}()
}
