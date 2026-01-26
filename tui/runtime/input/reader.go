package input

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
	"time"
)

// ==============================================================================
// Cancel Reader System (V3)
// ==============================================================================
// 跨平台的可取消输入读取器

var (
	// ErrCanceled 输入被取消
	ErrCanceled = errors.New("input canceled")
	// ErrTimeout 读取超时
	ErrTimeout = errors.New("input timeout")
)

// Reader 可取消的输入读取器接口
type Reader interface {
	io.Reader

	// Cancel 取消正在进行的读取操作
	Cancel() error

	// Close 关闭读取器
	Close() error
}

// CancelReader 可取消读取器
type CancelReader struct {
	r       io.Reader
	cancel  func()
	cancelChan chan struct{}
	closed  bool
	mu      sync.Mutex
}

// NewCancelReader 创建可取消读取器
func NewCancelReader(r io.Reader) *CancelReader {
	return &CancelReader{
		r:          r,
		cancelChan: make(chan struct{}),
	}
}

// Read 实现 io.Reader 接口
func (cr *CancelReader) Read(p []byte) (n int, err error) {
	// 首先检查是否已关闭
	cr.mu.Lock()
	if cr.closed {
		cr.mu.Unlock()
		return 0, io.ErrClosedPipe
	}
	cr.mu.Unlock()

	// 使用 select 实现可取消读取
	type result struct {
		n   int
		err error
	}
	resultChan := make(chan result, 1)

	go func() {
		n, err := cr.r.Read(p)
		resultChan <- result{n, err}
	}()

	select {
	case res := <-resultChan:
		return res.n, res.err
	case <-cr.cancelChan:
		return 0, ErrCanceled
	}
}

// Cancel 取消读取
func (cr *CancelReader) Cancel() error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if cr.closed {
		return nil
	}

	close(cr.cancelChan)

	// 调用平台特定的取消函数
	if cr.cancel != nil {
		cr.cancel()
	}

	return nil
}

// SetCancelFunc 设置取消函数
func (cr *CancelReader) SetCancelFunc(cancel func()) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.cancel = cancel
}

// Close 关闭读取器
func (cr *CancelReader) Close() error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if cr.closed {
		return nil
	}

	cr.closed = true

	// 安全地关闭通道
	select {
	case <-cr.cancelChan:
		// 已经关闭
	default:
		close(cr.cancelChan)
	}

	return nil
}

// ==============================================================================
// 终端输入读取器
// ==============================================================================

// TerminalInput 终端输入
type TerminalInput struct {
	reader Reader
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
}

// NewTerminalInput 创建终端输入
func NewTerminalInput() (*TerminalInput, error) {
	stdin := os.Stdin
	reader := NewCancelReader(stdin)

	ctx, cancel := context.WithCancel(context.Background())

	return &TerminalInput{
		reader: reader,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// NewTerminalInputWithContext 创建带上下文的终端输入
func NewTerminalInputWithContext(ctx context.Context) (*TerminalInput, error) {
	stdin := os.Stdin
	reader := NewCancelReader(stdin)

	return &TerminalInput{
		reader: reader,
		ctx:    ctx,
		cancel: func() {}, // 不拥有取消权限
	}, nil
}

// Read 读取输入（可取消）
func (ti *TerminalInput) Read(p []byte) (n int, err error) {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	select {
	case <-ti.ctx.Done():
		return 0, ti.ctx.Err()
	default:
		return ti.reader.Read(p)
	}
}

// ReadEvent 读取事件（带超时）
func (ti *TerminalInput) ReadEvent(timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ti.ctx, timeout)
	defer cancel()

	buf := make([]byte, 1024)

	type result struct {
		data []byte
		err  error
	}
	resultChan := make(chan result, 1)

	go func() {
		n, err := ti.reader.Read(buf)
		if n > 0 {
			resultChan <- result{buf[:n], err}
		} else {
			resultChan <- result{nil, err}
		}
	}()

	select {
	case res := <-resultChan:
		return res.data, res.err
	case <-ctx.Done():
		ti.reader.Cancel()
		return nil, ErrTimeout
	}
}

// Cancel 取消读取
func (ti *TerminalInput) Cancel() {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	ti.cancel()
	ti.reader.Cancel()
}

// Close 关闭输入
func (ti *TerminalInput) Close() error {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	ti.cancel()
	return ti.reader.Close()
}

// Context 返回上下文
func (ti *TerminalInput) Context() context.Context {
	ti.mu.Lock()
	defer ti.mu.Unlock()
	return ti.ctx
}

// WithContext 设置上下文
func (ti *TerminalInput) WithContext(ctx context.Context) {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	ti.cancel()
	ti.ctx = ctx
	ti.cancel = func() {} // 不拥有取消权限
}

// ==============================================================================
// 平台特定实现
// ==============================================================================

// PlatformReader 平台特定的读取器接口
type PlatformReader interface {
 Reader
	SetNonBlocking() error
	SetRawMode() error
	RestoreMode() error
}

// NewStdinReader 为当前平台创建 stdin 读取器
func NewStdinReader() (Reader, error) {
	return NewCancelReader(os.Stdin), nil
}

// IsNonBlockingCheck 检查错误是否为非阻塞模式的正常返回
func IsNonBlockingCheck(err error) bool {
	if err == nil {
		return false
	}

	// 在非阻塞模式下，没有数据可读时会返回 EAGAIN
	if errno, ok := err.(syscall.Errno); ok {
		return errno == syscall.EAGAIN || errno == syscall.EWOULDBLOCK
	}

	return false
}

// ==============================================================================
// 辅助函数
// ==============================================================================

// ReadWithTimeout 带超时读取
func ReadWithTimeout(r Reader, buf []byte, timeout time.Duration) (n int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	type result struct {
		n   int
		err error
	}
	resultChan := make(chan result, 1)

	go func() {
		n, err := r.Read(buf)
		resultChan <- result{n, err}
	}()

	select {
	case res := <-resultChan:
		return res.n, res.err
	case <-ctx.Done():
		if canceler, ok := r.(interface{ Cancel() error }); ok {
			canceler.Cancel()
		}
		return 0, ErrTimeout
	}
}

// ReadLine 读取一行（带取消支持）
func ReadLine(r Reader) (string, error) {
	var buf []byte
	single := make([]byte, 1)

	for {
		n, err := r.Read(single)
		if err != nil {
			if err == ErrCanceled {
				return "", err
			}
			// EOF 时，如果有数据则返回数据，否则返回 EOF
			if len(buf) > 0 {
				return string(buf), nil
			}
			return "", err
		}

		if n == 0 {
			break
		}

		buf = append(buf, single[0])

		if single[0] == '\n' {
			break
		}
	}

	return string(buf), nil
}

// ReadLineWithTimeout 带超时读取一行
func ReadLineWithTimeout(r Reader, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	type result struct {
		line string
		err  error
	}
	resultChan := make(chan result, 1)

	go func() {
		line, err := ReadLine(r)
		resultChan <- result{line, err}
	}()

	select {
	case res := <-resultChan:
		return res.line, res.err
	case <-ctx.Done():
		if canceler, ok := r.(interface{ Cancel() error }); ok {
			canceler.Cancel()
		}
		return "", ErrTimeout
	}
}

// ==============================================================================
// 输入事件
// ==============================================================================

// InputEvent 输入事件
type InputEvent struct {
	Type      InputEventType
	Data      []byte
	Timestamp time.Time
}

// InputEventType 输入事件类型
type InputEventType int

const (
	InputKey      InputEventType = iota // 按键
	InputMouse                            // 鼠标
	InputResize                           // 窗口大小变化
	InputPaste                            // 粘贴
	InputUnknown                          // 未知
)

// NewKeyEvent 创建按键事件
func NewKeyEvent(key rune) InputEvent {
	return InputEvent{
		Type:      InputKey,
		Data:      []byte(string(key)),
		Timestamp: time.Now(),
	}
}

// String 返回事件字符串表示
func (e InputEvent) String() string {
	switch e.Type {
	case InputKey:
		return fmt.Sprintf("Key: %s", string(e.Data))
	case InputMouse:
		return fmt.Sprintf("Mouse: %v", e.Data)
	case InputResize:
		return fmt.Sprintf("Resize: %v", e.Data)
	case InputPaste:
		return fmt.Sprintf("Paste: %s", string(e.Data))
	default:
		return "Unknown"
	}
}

// IsKey 检查是否是按键事件
func (e InputEvent) IsKey() bool {
	return e.Type == InputKey
}

// IsMouse 检查是否是鼠标事件
func (e InputEvent) IsMouse() bool {
	return e.Type == InputMouse
}

// IsResize 检查是否是调整大小事件
func (e InputEvent) IsResize() bool {
	return e.Type == InputResize
}

// Key 返回按键字符（如果是按键事件）
func (e InputEvent) Key() rune {
	if e.Type != InputKey || len(e.Data) == 0 {
		return 0
	}
	runes := []rune(string(e.Data))
	if len(runes) > 0 {
		return runes[0]
	}
	return 0
}

// ==============================================================================
// 输入缓冲区
// ==============================================================================

// InputBuffer 输入缓冲区
type InputBuffer struct {
	mu   sync.Mutex
	buf  []byte
	pos  int
	size int
}

// NewInputBuffer 创建输入缓冲区
func NewInputBuffer(size int) *InputBuffer {
	return &InputBuffer{
		buf:  make([]byte, size),
		pos:  0,
		size: size,
	}
}

// Write 写入数据
func (b *InputBuffer) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	n := copy(b.buf[b.pos:], data)
	b.pos += n
	return n, nil
}

// Read 读取数据
func (b *InputBuffer) Read(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.pos == 0 {
		return 0, io.EOF
	}

	n := copy(p, b.buf[:b.pos])

	// 移动剩余数据
	copy(b.buf, b.buf[n:b.pos])
	b.pos -= n

	return n, nil
}

// Bytes 返回缓冲区内容
func (b *InputBuffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()

	result := make([]byte, b.pos)
	copy(result, b.buf[:b.pos])
	return result
}

// Len 返回缓冲区长度
func (b *InputBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.pos
}

// Clear 清空缓冲区
func (b *InputBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pos = 0
}

// String 返回字符串内容
func (b *InputBuffer) String() string {
	return string(b.Bytes())
}
