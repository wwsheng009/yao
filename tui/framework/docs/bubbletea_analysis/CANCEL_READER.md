# 跨平台输入取消增强方案

## 概述

跨平台输入取消功能借鉴自 Bubble Tea 的 `cancelreader` 实现，提供统一的输入取消机制，支持优雅地中断阻塞的输入读取操作。

## 当前问题

### 输入取消不一致

不同平台对输入取消的支持不一致：

1. **Unix/Linux** - 可以使用 select/poll/epoll
2. **Windows** - 需要使用 WaitForMultipleObjects
3. **跨平台** - 需要统一的抽象接口

### 具体场景

```go
// 问题场景：无法取消的阻塞读取
func (a *App) ReadInput() {
    for {
        // 这个调用会阻塞，无法从外部中断
        event := readEvent()
        a.Dispatch(event)
    }
}
```

## 设计方案

### 核心接口

```go
// tui/runtime/input/reader.go

package input

import (
    "io"
    "os"
)

// Reader 可取消的输入读取器接口
type Reader interface {
    io.Reader

    // Cancel 取消正在进行的读取操作
    Cancel() error

    // SetCancelFunc 设置取消函数
    SetCancelFunc(cancel func())

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

func (cr *CancelReader) Read(p []byte) (n int, err error) {
    cr.mu.Lock()
    defer cr.mu.Unlock()

    if cr.closed {
        return 0, io.ErrClosedPipe
    }

    // 使用 select 实现可取消读取
    done := make(chan struct{})
    result := make(chan readResult, 1)

    go func() {
        n, err := cr.r.Read(p)
        result <- readResult{n, err}
        close(done)
    }()

    select {
    case <-cr.cancelChan:
        // 被取消
        return 0, ErrCanceled
    case res := <-result:
        return res.n, res.err
    }
}

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

func (cr *CancelReader) SetCancelFunc(cancel func()) {
    cr.mu.Lock()
    defer cr.mu.Unlock()
    cr.cancel = cancel
}

func (cr *CancelReader) Close() error {
    cr.mu.Lock()
    defer cr.mu.Unlock()

    if cr.closed {
        return nil
    }

    cr.closed = true
    close(cr.cancelChan)
    return nil
}

type readResult struct {
    n   int
    err error
}
```

### Unix 实现

```go
// tui/runtime/input/unix.go

// +build darwin linux freebsd netbsd openbsd

package input

import (
    "os"
    "syscall"
)

// NewStdinReader 为 Unix 创建 stdin 读取器
func NewStdinReader() (Reader, error) {
    // 设置非阻塞模式
    fd := int(os.Stdin.Fd())
    flags, err := syscall.FcntlInt(uintptr(fd), syscall.F_GETFL, 0)
    if err != nil {
        return nil, err
    }

    // 设置为非阻塞
    _, err = syscall.FcntlInt(uintptr(fd), syscall.F_SETFL, flags|syscall.O_NONBLOCK)
    if err != nil {
        return nil, err
    }

    cr := NewCancelReader(os.Stdin)

    // Unix 特定的取消函数
    cr.SetCancelFunc(func() {
        // 在 Unix 上，非阻塞模式下取消不需要特殊操作
        // 读取会自动返回 EAGAIN
    })

    return cr, nil
}

// WaitForInput 等待输入（带超时和取消）
func WaitForInput(r Reader, timeout time.Duration) error {
    fd := int(os.Stdin.Fd())

    // 创建文件描述符集合
    fds := &syscall.FdSet{}
    fds.Set(fd)

    // 计算超时
    var tv *syscall.Timeval
    if timeout > 0 {
        tv = &syscall.Timeval{
            Sec:  timeout.Seconds(),
            Usec: timeout.Microseconds() % 1000000,
        }
    }

    // 使用 select 等待
    _, err := syscall.Select(fd+1, fds, nil, nil, tv)
    return err
}
```

### Windows 实现

```go
// tui/runtime/input/windows.go

// +build windows

package input

import (
    "os"
    "syscall"
    "unsafe"
)

var (
    kernel32 = syscall.NewLazyDLL("kernel32.dll")

    procCancelIoEx = kernel32.NewProc("CancelIoEx")
    procCreateEvent = kernel32.NewProc("CreateEventW")
)

// NewStdinReader 为 Windows 创建 stdin 读取器
func NewStdinReader() (Reader, error) {
    cr := NewCancelReader(os.Stdin)

    // Windows 特定的取消函数
    cr.SetCancelFunc(func() {
        handle := syscall.Handle(os.Stdin.Fd())
        procCancelIoEx.Call(uintptr(handle), 0)
    })

    return cr, nil
}

// WaitForInput 等待输入（Windows 版本）
func WaitForInput(r Reader, timeout time.Duration) error {
    handle := syscall.Handle(os.Stdin.Fd())

    // 创建事件对象
    event, _, err := procCreateEvent.Call(
        0, 0, 1, 0,
        uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("INPUT_EVENT"))),
    )
    if event == 0 {
        return err
    }

    // 等待输入或超时
    var msec uint32
    if timeout > 0 {
        msec = uint32(timeout.Milliseconds())
    } else {
        msec = syscall.INFINITE
    }

    // WaitForMultipleObjects
    _, err = syscall.WaitForMultipleObjects(
        1,
        &handle,
        false,
        msec,
    )

    return err
}
```

### 终端输入适配器

```go
// tui/runtime/input/terminal.go

package input

import (
    "context"
    "time"
)

// TerminalInput 终端输入
type TerminalInput struct {
    reader Reader
    ctx    context.Context
    cancel context.CancelFunc
}

// NewTerminalInput 创建终端输入
func NewTerminalInput() (*TerminalInput, error) {
    reader, err := NewStdinReader()
    if err != nil {
        return nil, err
    }

    ctx, cancel := context.WithCancel(context.Background())

    return &TerminalInput{
        reader: reader,
        ctx:    ctx,
        cancel: cancel,
    }, nil
}

// ReadEvent 读取事件（可取消）
func (ti *TerminalInput) ReadEvent() (Event, error) {
    buf := make([]byte, 1024)

    for {
        select {
        case <-ti.ctx.Done():
            return nil, ErrCanceled

        default:
            n, err := ti.reader.Read(buf)
            if err != nil {
                return nil, err
            }

            event, err := ParseEvent(buf[:n])
            if err != nil {
                return nil, err
            }

            return event, nil
        }
    }
}

// ReadEventWithTimeout 带超时读取
func (ti *TerminalInput) ReadEventWithTimeout(timeout time.Duration) (Event, error) {
    ctx, cancel := context.WithTimeout(ti.ctx, timeout)
    defer cancel()

    eventChan := make(chan Event, 1)
    errChan := make(chan error, 1)

    go func() {
        event, err := ti.ReadEvent()
        if err != nil {
            errChan <- err
            return
        }
        eventChan <- event
    }()

    select {
    case event := <-eventChan:
        return event, nil
    case err := <-errChan:
        return nil, err
    case <-ctx.Done():
        ti.Cancel()
        return nil, ErrTimeout
    }
}

// Cancel 取消读取
func (ti *TerminalInput) Cancel() {
    ti.cancel()
    ti.reader.Cancel()
}

// Close 关闭输入
func (ti *TerminalInput) Close() error {
    ti.cancel()
    return ti.reader.Close()
}
```

## 使用示例

### 1. 基础使用

```go
func (a *App) Run() error {
    input, err := NewTerminalInput()
    if err != nil {
        return err
    }
    defer input.Close()

    for {
        event, err := input.ReadEvent()
        if err != nil {
            if err == ErrCanceled {
                break
            }
            return err
        }

        a.Dispatch(event)
    }

    return nil
}
```

### 2. 优雅关闭

```go
func (a *App) Shutdown() {
    // 取消输入读取
    if a.input != nil {
        a.input.Cancel()
    }

    // 等待清理
    time.Sleep(100 * time.Millisecond)
}
```

### 3. 超时读取

```go
func (a *App) ReadWithTimeout(timeout time.Duration) (Event, error) {
    return a.input.ReadEventWithTimeout(timeout)
}
```

### 4. 上下文集成

```go
func (a *App) RunWithContext(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            a.input.Cancel()
            return ctx.Err()

        default:
            event, err := a.input.ReadEvent()
            if err != nil {
                return err
            }
            a.Dispatch(event)
        }
    }
}
```

## 实施计划

### Phase 1: 核心接口 (Week 1)

- [ ] 实现 `Reader` 接口
- [ ] 实现 `CancelReader`
- [ ] 单元测试

### Phase 2: 平台实现 (Week 2)

- [ ] Unix 实现
- [ ] Windows 实现
- [ ] 集成测试

### Phase 3: 终端适配器 (Week 2)

- [ ] 实现 `TerminalInput`
- [ ] 实现超时读取
- [ ] 上下文集成

### Phase 4: 文档和示例 (Week 3)

- [ ] API 文档
- [ ] 平台差异说明
- [ ] 使用示例

## 文件结构

```
tui/runtime/input/
├── reader.go               # 核心接口
├── cancel_reader.go        # 可取消读取器
├── terminal.go             # 终端输入
├── unix.go                 # Unix 实现
├── windows.go              # Windows 实现
├── reader_test.go          # 测试
└── README.md               # 文档
```

## 测试策略

```go
func TestCancelReader(t *testing.T) {
    r, w := io.Pipe()
    cr := NewCancelReader(r)

    // 测试取消
    go func() {
        time.Sleep(100 * time.Millisecond)
        cr.Cancel()
    }()

    buf := make([]byte, 1024)
    _, err := cr.Read(buf)

    assert.Equal(t, ErrCanceled, err)
}

func TestTerminalInput(t *testing.T) {
    ti, err := NewTerminalInput()
    require.NoError(t, err)
    defer ti.Close()

    // 测试超时读取
    _, err = ti.ReadEventWithTimeout(100 * time.Millisecond)
    assert.Equal(t, ErrTimeout, err)
}
```

## 性能考虑

1. **零拷贝** - 直接读取到缓冲区
2. **最小锁竞争** - 细粒度锁保护
3. **高效取消** - 使用通道而非互斥锁

## 与 Bubble Tea 的差异

| 特性 | Bubble Tea cancelreader | Yao TUI Input |
|------|------------------------|---------------|
| **抽象层级** | 独立包 | 集成到 runtime |
| **上下文支持** | 无 | 原生支持 |
| **超时读取** | 无 | 内置 |
| **事件解析** | 分离 | 集成 |
| **终端管理** | 分离 | 集成 |
| **平台支持** | Unix/Windows | Unix/Windows |

## 平台差异

| 平台 | 实现方式 | 特点 |
|------|---------|------|
| Linux | epoll/select | 高效 |
| macOS | kqueue/select | 兼容性好 |
| Windows | WaitForMultipleObjects | 需要特殊处理 |
| FreeBSD | kqueue/select | 类似 macOS |

## 向后兼容

```go
// 现有代码可以渐进式迁移
// 旧代码
func (a *App) ReadInput() Event {
    return parseEvent(readRawInput())
}

// 新代码
func (a *App) ReadInput() (Event, error) {
    return a.input.ReadEvent()
}
```
