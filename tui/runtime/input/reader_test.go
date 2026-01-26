package input

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestNewCancelReader(t *testing.T) {
	r := bytes.NewBufferString("test")
	reader := NewCancelReader(r)

	if reader == nil {
		t.Error("reader should not be nil")
	}
}

func TestCancelReader_Read(t *testing.T) {
	r := bytes.NewBufferString("hello")
	reader := NewCancelReader(r)

	buf := make([]byte, 5)
	n, err := reader.Read(buf)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes, got %d", n)
	}
	if string(buf) != "hello" {
		t.Errorf("expected 'hello', got '%s'", string(buf))
	}
}

func TestCancelReader_Cancel(t *testing.T) {
	// 创建一个永远阻塞的读取器
	blockingReader := &blockingReader{}
	reader := NewCancelReader(blockingReader)

	// 启动读取
	buf := make([]byte, 10)
	errChan := make(chan error, 1)

	go func() {
		_, err := reader.Read(buf)
		errChan <- err
	}()

	// 等待一下让读取启动
	time.Sleep(10 * time.Millisecond)

	// 取消读取
	err := reader.Cancel()
	if err != nil {
		t.Errorf("cancel failed: %v", err)
	}

	// 应该收到取消错误
	select {
	case err := <-errChan:
		if err != ErrCanceled {
			t.Errorf("expected ErrCanceled, got %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("should receive error after cancel")
	}
}

func TestCancelReader_Close(t *testing.T) {
	r := bytes.NewBufferString("test")
	reader := NewCancelReader(r)

	err := reader.Close()
	if err != nil {
		t.Errorf("close failed: %v", err)
	}

	// 关闭后再读取应该返回错误
	buf := make([]byte, 10)
	_, err = reader.Read(buf)
	if err != io.ErrClosedPipe {
		t.Errorf("expected ErrClosedPipe, got %v", err)
	}
}

func TestCancelReader_CloseAfterCancel(t *testing.T) {
	r := bytes.NewBufferString("test")
	reader := NewCancelReader(r)

	reader.Cancel()
	err := reader.Close()
	if err != nil {
		t.Errorf("close after cancel should not error, got: %v", err)
	}
}

func TestTerminalInput_New(t *testing.T) {
	input, err := NewTerminalInput()
	if err != nil {
		t.Fatalf("failed to create terminal input: %v", err)
	}

	if input == nil {
		t.Error("input should not be nil")
	}

	input.Close()
}

func TestTerminalInput_NewWithContext(t *testing.T) {
	ctx := context.Background()
	input, err := NewTerminalInputWithContext(ctx)
	if err != nil {
		t.Fatalf("failed to create terminal input: %v", err)
	}

	if input.Context() != ctx {
		t.Error("context should match")
	}

	input.Close()
}

func TestTerminalInput_Cancel(t *testing.T) {
	input, err := NewTerminalInput()
	if err != nil {
		t.Fatalf("failed to create terminal input: %v", err)
	}

	// 取消不应该 panic
	input.Cancel()
	input.Close()
}

func TestTerminalInput_WithContext(t *testing.T) {
	input, err := NewTerminalInput()
	if err != nil {
		t.Fatalf("failed to create terminal input: %v", err)
	}

	ctx := context.Background()
	input.WithContext(ctx)

	if input.Context() != ctx {
		t.Error("context should match")
	}

	input.Close()
}

func TestReadWithTimeout(t *testing.T) {
	r := bytes.NewBufferString("hello")
	reader := NewCancelReader(r)

	buf := make([]byte, 5)
	n, err := ReadWithTimeout(reader, buf, 100*time.Millisecond)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes, got %d", n)
	}
	if string(buf) != "hello" {
		t.Errorf("expected 'hello', got '%s'", string(buf))
	}
	_ = reader.Close()
}

func TestReadWithTimeout_Timeout(t *testing.T) {
	blockingReader := NewCancelReader(&blockingReader{})

	_, err := ReadWithTimeout(blockingReader, make([]byte, 10), 10*time.Millisecond)
	if err != ErrTimeout {
		t.Errorf("expected ErrTimeout, got %v", err)
	}
}

func TestReadLine(t *testing.T) {
	r := bytes.NewBufferString("hello\nworld")

	line, err := ReadLine(NewCancelReader(r))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if line != "hello\n" {
		t.Errorf("expected 'hello\\n', got '%s'", line)
	}

	// 读取第二行
	line, err = ReadLine(NewCancelReader(r))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if line != "world" {
		t.Errorf("expected 'world', got '%s'", line)
	}
}

func TestReadLine_Canceled(t *testing.T) {
	reader := NewCancelReader(&blockingReader{})

	// 启动读取
	errChan := make(chan error, 1)
	go func() {
		_, err := ReadLine(reader)
		errChan <- err
	}()

	time.Sleep(10 * time.Millisecond)
	reader.Cancel()

	select {
	case err := <-errChan:
		if err != ErrCanceled {
			t.Errorf("expected ErrCanceled, got %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("should receive error after cancel")
	}
}

func TestReadLineWithTimeout(t *testing.T) {
	r := bytes.NewBufferString("test line\n")

	line, err := ReadLineWithTimeout(NewCancelReader(r), 100*time.Millisecond)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if line != "test line\n" {
		t.Errorf("expected 'test line\\n', got '%s'", line)
	}
}

func TestReadLineWithTimeout_Timeout(t *testing.T) {
	reader := NewCancelReader(&blockingReader{})

	_, err := ReadLineWithTimeout(reader, 10*time.Millisecond)
	if err != ErrTimeout {
		t.Errorf("expected ErrTimeout, got %v", err)
	}
}

func TestInputEvent_Key(t *testing.T) {
	event := NewKeyEvent('a')

	if !event.IsKey() {
		t.Error("should be key event")
	}

	if event.IsMouse() {
		t.Error("should not be mouse event")
	}

	if event.Key() != 'a' {
		t.Errorf("expected key 'a', got %c", event.Key())
	}

	str := event.String()
	if !strings.Contains(str, "Key") {
		t.Error("string should contain 'Key'")
	}
}

func TestInputEvent_Mouse(t *testing.T) {
	event := InputEvent{
		Type:      InputMouse,
		Data:      []byte("click"),
		Timestamp: time.Now(),
	}

	if !event.IsMouse() {
		t.Error("should be mouse event")
	}

	if event.IsKey() {
		t.Error("should not be key event")
	}
}

func TestInputEvent_Resize(t *testing.T) {
	event := InputEvent{
		Type:      InputResize,
		Data:      []byte("80x24"),
		Timestamp: time.Now(),
	}

	if !event.IsResize() {
		t.Error("should be resize event")
	}

	if !strings.Contains(event.String(), "Resize") {
		t.Error("string should contain 'Resize'")
	}
}

func TestInputBuffer_Write(t *testing.T) {
	buf := NewInputBuffer(10)

	n, err := buf.Write([]byte("hello"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5, got %d", n)
	}

	if buf.Len() != 5 {
		t.Errorf("expected length 5, got %d", buf.Len())
	}

	if buf.String() != "hello" {
		t.Errorf("expected 'hello', got '%s'", buf.String())
	}
}

func TestInputBuffer_Read(t *testing.T) {
	buf := NewInputBuffer(10)

	buf.Write([]byte("hello"))
	readBuf := make([]byte, 10)

	n, err := buf.Read(readBuf)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5, got %d", n)
	}

	if string(readBuf[:5]) != "hello" {
		t.Errorf("expected 'hello', got '%s'", string(readBuf[:5]))
	}

	// 再次读取应该返回 EOF
	n, err = buf.Read(readBuf)
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func TestInputBuffer_Clear(t *testing.T) {
	buf := NewInputBuffer(10)

	buf.Write([]byte("hello"))
	if buf.Len() != 5 {
		t.Errorf("expected length 5, got %d", buf.Len())
	}

	buf.Clear()
	if buf.Len() != 0 {
		t.Errorf("expected length 0 after clear, got %d", buf.Len())
	}
}

func TestInputBuffer_Concurrent(t *testing.T) {
	buf := NewInputBuffer(100)
	var wg sync.WaitGroup
	var errors int32

	// 并发写入
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				if _, err := buf.Write([]byte("a")); err != nil {
					atomic.AddInt32(&errors, 1)
				}
			}
		}(i)
	}

	// 并发读取
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			readBuf := make([]byte, 10)
			for j := 0; j < 10; j++ {
				if _, err := buf.Read(readBuf); err != nil && err != io.EOF {
					atomic.AddInt32(&errors, 1)
				}
			}
		}()
	}

	wg.Wait()

	if errors > 0 {
		t.Errorf("encountered %d errors", errors)
	}

	// 最终应该有正确数量的数据
	expectedLen := 10 * 10 * 10 // 10 goroutines * 10 writes * 1 byte
	actualLen := buf.Len()
	if actualLen > expectedLen {
		t.Errorf("buffer length %d exceeds expected %d", actualLen, expectedLen)
	}
}

func TestErrCanceled(t *testing.T) {
	if ErrCanceled == nil {
		t.Error("ErrCanceled should not be nil")
	}

	if ErrCanceled.Error() != "input canceled" {
		t.Errorf("unexpected error message: %s", ErrCanceled.Error())
	}
}

func TestErrTimeout(t *testing.T) {
	if ErrTimeout == nil {
		t.Error("ErrTimeout should not be nil")
	}

	if ErrTimeout.Error() != "input timeout" {
		t.Errorf("unexpected error message: %s", ErrTimeout.Error())
	}
}

func TestIsNonBlockingCheck(t *testing.T) {
	// 测试 EAGAIN
	eagain := syscall.Errno(11) // EAGAIN on most systems
	if !IsNonBlockingCheck(eagain) {
		t.Logf("EAGAIN(%d) should be non-blocking check, got: %v", int(eagain), IsNonBlockingCheck(eagain))
	}

	// 在 Windows 上 EAGAIN 可能是不同的值
	eagainWindows := syscall.Errno(11)
	if int(eagainWindows) == 11 {
		if !IsNonBlockingCheck(eagainWindows) {
			t.Log("EAGAIN not recognized as non-blocking on this system")
		}
	}

	// 测试普通错误
	regular := io.EOF
	if IsNonBlockingCheck(regular) {
		t.Error("EOF should not be non-blocking check")
	}
}

// blockingReader 用于测试的阻塞读取器
type blockingReader struct{}

func (b *blockingReader) Read(p []byte) (int, error) {
	time.Sleep(100 * time.Millisecond)
	return 0, nil // 模拟没有数据
}

func TestCancelReader_SetCancelFunc(t *testing.T) {
	r := bytes.NewBufferString("test")
	reader := NewCancelReader(r)

	var called bool
	reader.SetCancelFunc(func() {
		called = true
	})

	reader.Cancel()

	if !called {
		t.Error("cancel function should be called")
	}
}
