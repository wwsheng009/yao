# Stream Data System Design (V3)

> **优先级**: P1 (流式数据处理)
> **目标**: 支持实时流式数据显示
> **关键特性**: 流式订阅、增量更新、自动滚动、缓冲控制

## 概述

在 TUI 应用中，有些数据是持续产生的，如：
- 日志输出
- 实时监控数据
- AI 生成的响应流
- 网络请求响应
- 命令行输出

这些数据需要实时显示，而不是等待全部完成后才显示。

### 为什么需要流式数据系统？

**传统方式的问题**：
```go
// ❌ 等待所有数据
func ShowLog() {
    logs := fetchAllLogs()  // 阻塞，等待所有日志
    displayLogs(logs)
}

// 问题：
// - 用户等待时间长
// - 无法看到实时进度
// - 内存占用大
// - 无法提前终止
```

**流式处理的优势**：
```go
// ✅ 实时流式显示
func ShowLog() {
    stream := SubscribeToLogs()
    for chunk := range stream {
        displayChunk(chunk)  // 立即显示
    }
}

// 优势：
// - 实时反馈
// - 内存占用恒定
// - 可以提前终止
// - 支持自动滚动
```

## 设计目标

1. **实时显示**: 数据到达立即显示
2. **增量更新**: 只更新变化部分
3. **自动滚动**: 自动滚动到最新数据
4. **缓冲控制**: 防止内存溢出
5. **可取消**: 支持取消订阅
6. **组件集成**: 与现有组件无缝集成

## 核心类型定义

### 1. Stream 接口

```go
// 位于: tui/framework/stream/stream.go

package stream

import (
    "context"
)

// Stream 流接口
type Stream interface {
    // Subscribe 订阅流
    Subscribe() Subscription

    // Publish 发布数据
    Publish(data interface{}) error

    // Close 关闭流
    Close() error

    // IsClosed 是否已关闭
    IsClosed() bool
}

// Subscription 订阅
type Subscription interface {
    // Chan 获取数据通道
    Chan() <-interface{}

    // Unsubscribe 取消订阅
    Unsubscribe()

    // IsSubscribed 是否已订阅
    IsSubscribed() bool
}

// Chunk 数据块
type Chunk struct {
    Data  interface{}
    Index int64
    Time  time.Time
}

// StreamOption 流选项
type StreamOption func(*StreamConfig)

// StreamConfig 流配置
type StreamConfig struct {
    BufferSize    int           // 缓冲区大小
    MaxChunks     int           // 最大块数
    AutoScroll    bool          // 自动滚动
    OnData        func(Chunk)   // 数据回调
    OnError       func(error)   // 错误回调
    OnComplete    func()        // 完成回调
}
```

### 2. BasicStream 基础流

```go
// 位于: tui/framework/stream/basic_stream.go

package stream

import (
    "context"
    "sync"
    "sync/atomic"
)

// BasicStream 基础流实现
type BasicStream struct {
    config     StreamConfig
    subscribers map[int64]*basicSubscription
    nextSubID  int64
    closed     atomic.Bool
    mu         sync.RWMutex
}

// basicSubscription 基础订阅
type basicSubscription struct {
    id     int64
    ch     chan interface{}
    active atomic.Bool
    stream *BasicStream
}

// NewStream 创建流
func NewStream(opts ...StreamOption) Stream {
    config := StreamConfig{
        BufferSize: 100,
        MaxChunks:  1000,
        AutoScroll: true,
    }

    for _, opt := range opts {
        opt(&config)
    }

    return &BasicStream{
        config:      config,
        subscribers: make(map[int64]*basicSubscription),
    }
}

// WithBufferSize 设置缓冲区大小
func WithBufferSize(size int) StreamOption {
    return func(c *StreamConfig) {
        c.BufferSize = size
    }
}

// WithMaxChunks 设置最大块数
func WithMaxChunks(max int) StreamOption {
    return func(c *StreamConfig) {
        c.MaxChunks = max
    }
}

// WithAutoScroll 设置自动滚动
func WithAutoScroll(enable bool) StreamOption {
    return func(c *StreamConfig) {
        c.AutoScroll = enable
    }
}

// WithOnData 设置数据回调
func WithOnData(fn func(Chunk)) StreamOption {
    return func(c *StreamConfig) {
        c.OnData = fn
    }
}

// Subscribe 订阅流
func (s *BasicStream) Subscribe() Subscription {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.closed.Load() {
        return &closedSubscription{}
    }

    sub := &basicSubscription{
        id:     atomic.AddInt64(&s.nextSubID, 1),
        ch:     make(chan interface{}, s.config.BufferSize),
        active: atomic.Value{},
        stream: s,
    }
    sub.active.Store(true)

    s.subscribers[sub.id] = sub

    return sub
}

// Publish 发布数据
func (s *BasicStream) Publish(data interface{}) error {
    if s.closed.Load() {
        return ErrStreamClosed
    }

    s.mu.RLock()
    defer s.mu.RUnlock()

    // 发送给所有订阅者
    for _, sub := range s.subscribers {
        if sub.active.Load().(bool) {
            select {
            case sub.ch <- data:
            default:
                // 缓冲区满，丢弃数据
            }
        }
    }

    // 触发回调
    if s.config.OnData != nil {
        s.config.OnData(Chunk{
            Data: data,
            Time: time.Now(),
        })
    }

    return nil
}

// Close 关闭流
func (s *BasicStream) Close() error {
    if !s.closed.CompareAndSwap(false, true) {
        return nil // 已经关闭
    }

    s.mu.Lock()
    defer s.mu.Unlock()

    // 关闭所有订阅
    for _, sub := range s.subscribers {
        sub.Unsubscribe()
    }
    s.subscribers = make(map[int64]*basicSubscription)

    return nil
}

// IsClosed 是否已关闭
func (s *BasicStream) IsClosed() bool {
    return s.closed.Load()
}

// === basicSubscription 方法 ===

func (s *basicSubscription) Chan() <-interface{} {
    return s.ch
}

func (s *basicSubscription) Unsubscribe() {
    if !s.active.CompareAndSwap(true, false) {
        return // 已经取消
    }

    close(s.ch)

    s.stream.mu.Lock()
    delete(s.stream.subscribers, s.id)
    s.stream.mu.Unlock()
}

func (s *basicSubscription) IsSubscribed() bool {
    return s.active.Load().(bool)
}

// closedSubscription 已关闭的订阅
type closedSubscription struct{}

func (s *closedSubscription) Chan() <-interface{} {
    ch := make(chan interface{})
    close(ch)
    return ch
}

func (s *closedSubscription) Unsubscribe() {}

func (s *closedSubscription) IsSubscribed() bool {
    return false
}
```

### 3. StreamBuffer 流缓冲区

```go
// 位于: tui/framework/stream/buffer.go

package stream

// StreamBuffer 流缓冲区
type StreamBuffer struct {
    chunks    []Chunk
    maxChunks int
    autoScroll bool
    offset    int  // 滚动偏移
    mu        sync.RWMutex
}

// NewStreamBuffer 创建流缓冲区
func NewStreamBuffer(maxChunks int, autoScroll bool) *StreamBuffer {
    return &StreamBuffer{
        chunks:     make([]Chunk, 0, maxChunks),
        maxChunks:  maxChunks,
        autoScroll: autoScroll,
    }
}

// Append 添加数据块
func (b *StreamBuffer) Append(chunk Chunk) {
    b.mu.Lock()
    defer b.mu.Unlock()

    b.chunks = append(b.chunks, chunk)

    // 限制最大块数
    if len(b.chunks) > b.maxChunks {
        // 移除最旧的块
        b.chunks = b.chunks[1:]
        if b.offset > 0 {
            b.offset--
        }
    }

    // 自动滚动
    if b.autoScroll {
        b.offset = len(b.chunks)
    }
}

// GetVisible 获取可见块
func (b *StreamBuffer) GetVisible(height int) []Chunk {
    b.mu.RLock()
    defer b.mu.RUnlock()

    // 计算可见范围
    start := b.offset - height
    if start < 0 {
        start = 0
    }
    end := b.offset
    if end > len(b.chunks) {
        end = len(b.chunks)
    }

    if start >= end {
        return []Chunk{}
    }

    result := make([]Chunk, 0, end-start)
    for i := start; i < end; i++ {
        result = append(result, b.chunks[i])
    }

    return result
}

// Scroll 滚动
func (b *StreamBuffer) Scroll(delta int) {
    b.mu.Lock()
    defer b.mu.Unlock()

    b.offset += delta
    if b.offset < 0 {
        b.offset = 0
    }
    if b.offset > len(b.chunks) {
        b.offset = len(b.chunks)
    }
}

// ScrollTo 滚动到位置
func (b *StreamBuffer) ScrollTo(pos int) {
    b.mu.Lock()
    defer b.mu.Unlock()

    b.offset = pos
    if b.offset < 0 {
        b.offset = 0
    }
    if b.offset > len(b.chunks) {
        b.offset = len(b.chunks)
    }
}

// ScrollToEnd 滚动到末尾
func (b *StreamBuffer) ScrollToEnd() {
    b.mu.Lock()
    defer b.mu.Unlock()

    b.offset = len(b.chunks)
}

// Size 获取总块数
func (b *StreamBuffer) Size() int {
    b.mu.RLock()
    defer b.mu.RUnlock()

    return len(b.chunks)
}

// Offset 获取滚动偏移
func (b *StreamBuffer) Offset() int {
    b.mu.RLock()
    defer b.mu.RUnlock()

    return b.offset
}

// Clear 清空缓冲区
func (b *StreamBuffer) Clear() {
    b.mu.Lock()
    defer b.mu.Unlock()

    b.chunks = make([]Chunk, 0, b.maxChunks)
    b.offset = 0
}
```

### 4. StreamViewer 流查看器组件

```go
// 位于: tui/framework/component/stream_viewer.go

package component

import (
    "github.com/yaoapp/yao/tui/framework/stream"
    "github.com/yaoapp/yao/tui/runtime"
)

// StreamViewer 流查看器
type StreamViewer struct {
    BaseComponent
    *Measurable
    *ThemeHolder

    // 流
    stream      stream.Stream
    subscription stream.Subscription

    // 缓冲区
    buffer *stream.StreamBuffer

    // 配置
    formatter func(interface{}) string
    lineLimit int

    // 状态
    active bool
}

// NewStreamViewer 创建流查看器
func NewStreamViewer(s stream.Stream) *StreamViewer {
    viewer := &StreamViewer{
        stream:     s,
        buffer:     stream.NewStreamBuffer(1000, true),
        formatter:  defaultFormatter,
        lineLimit:  1000,
        active:     false,
    }

    viewer.Measurable = NewMeasurable()
    viewer.ThemeHolder = NewThemeHolder(nil)

    return viewer
}

// SetFormatter 设置格式化函数
func (v *StreamViewer) SetFormatter(fn func(interface{}) string) {
    v.formatter = fn
}

// SetLineLimit 设置行数限制
func (v *StreamViewer) SetLineLimit(limit int) {
    v.lineLimit = limit
    v.buffer = stream.NewStreamBuffer(limit, true)
}

// Start 开始接收数据
func (v *StreamViewer) Start() {
    if v.active {
        return
    }

    v.subscription = v.stream.Subscribe()
    v.active = true

    // 启动接收 goroutine
    go v.receive()
}

// Stop 停止接收数据
func (v *StreamViewer) Stop() {
    if !v.active {
        return
    }

    v.active = false
    if v.subscription != nil {
        v.subscription.Unsubscribe()
    }
}

// receive 接收数据
func (v *StreamViewer) receive() {
    for data := range v.subscription.Chan() {
        chunk := stream.Chunk{
            Data: data,
            Time: time.Now(),
        }

        v.buffer.Append(chunk)
        v.MarkDirty()
    }
}

// HandleAction 处理 Action
func (v *StreamViewer) HandleAction(a *action.Action) bool {
    switch a.Type {
    case action.ActionNavigateUp:
        v.buffer.Scroll(-1)
        v.MarkDirty()
        return true

    case action.ActionNavigateDown:
        v.buffer.Scroll(1)
        v.MarkDirty()
        return true

    case action.ActionNavigatePageUp:
        v.buffer.Scroll(-20)
        v.MarkDirty()
        return true

    case action.ActionNavigatePageDown:
        v.buffer.Scroll(20)
        v.MarkDirty()
        return true

    case action.ActionNavigateFirst:
        v.buffer.ScrollTo(0)
        v.MarkDirty()
        return true

    case action.ActionNavigateLast:
        v.buffer.ScrollToEnd()
        v.MarkDirty()
        return true

    case action.ActionCancel:
        v.Stop()
        return true

    default:
        return false
    }
}

// Paint 绘制
func (v *StreamViewer) Paint(ctx PaintContext, buf *runtime.CellBuffer) {
    theme := v.GetTheme()
    bounds := v.Bounds()

    // 获取可见块
    height := bounds.Height
    visible := v.buffer.GetVisible(height)

    // 绘制每一行
    y := bounds.Y
    for _, chunk := range visible {
        text := v.formatter(chunk.Data)

        // 绘制时间戳
        if !chunk.Time.IsZero() {
            timestamp := chunk.Time.Format("15:04:05")
            buf.DrawText(bounds.X, y, timestamp+" ", theme.GetStyle("stream.timestamp"))
        }

        // 绘制内容
        buf.DrawText(bounds.X+9, y, text, theme.GetStyle("stream.content"))

        y++
    }

    // 绘制空行（填充剩余空间）
    for y < bounds.Y+bounds.Height {
        buf.DrawText(bounds.X, y, "~", theme.GetStyle("stream.empty"))
        y++
    }
}

// Measure 测量尺寸
func (v *StreamViewer) Measure(maxWidth, maxHeight int) (width, height int) {
    return maxWidth, maxHeight
}

func defaultFormatter(data interface{}) string {
    return fmt.Sprintf("%v", data)
}
```

### 5. 流式数据源

```go
// 位于: tui/framework/stream/sources.go

package stream

import (
    "bufio"
    "context"
    "io"
    "os/exec"
)

// ReaderStream 从 io.Reader 创建流
func ReaderStream(ctx context.Context, r io.Reader) Stream {
    s := NewStream()

    go func() {
        scanner := bufio.NewScanner(r)
        for scanner.Scan() {
            select {
            case <-ctx.Done():
                s.Close()
                return
            default:
                s.Publish(scanner.Text())
            }
        }
        s.Close()
    }()

    return s
}

// CommandStream 从命令输出创建流
func CommandStream(ctx context.Context, cmd *exec.Cmd) (Stream, error) {
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return nil, err
    }

    if err := cmd.Start(); err != nil {
        return nil, err
    }

    return ReaderStream(ctx, stdout), nil
}

// ChannelStream 从 channel 创建流
func ChannelStream(ctx context.Context, ch <-chan interface{}) Stream {
    s := NewStream()

    go func() {
        for {
            select {
            case <-ctx.Done():
                s.Close()
                return
            case data, ok := <-ch:
                if !ok {
                    s.Close()
                    return
                }
                s.Publish(data)
            }
        }
    }()

    return s
}

// GeneratorStream 从生成函数创建流
func GeneratorStream(ctx context.Context, gen func() (interface{}, bool)) Stream {
    s := NewStream()

    go func() {
        for {
            select {
            case <-ctx.Done():
                s.Close()
                return
            default:
                data, ok := gen()
                if !ok {
                    s.Close()
                    return
                }
                s.Publish(data)
            }
        }
    }()

    return s
}

// FileAsStream 从文件逐行读取创建流
func FileAsStream(ctx context.Context, path string) (Stream, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }

    return ReaderStream(ctx, f), nil
}
```

### 6. AI 响应流

```go
// 位于: tui/framework/stream/ai_stream.go

package stream

import (
    "context"
    "encoding/json"
)

// AIStream AI 响应流
type AIStream struct {
    *BasicStream
}

// AIChunk AI 数据块
type AIChunk struct {
    Type    string      // "text", "error", "done"
    Content interface{}
}

// NewAIStream 创建 AI 流
func NewAIStream() *AIStream {
    return &AIStream{
        BasicStream: NewStream().(*BasicStream),
    }
}

// FromSSE 从 Server-Sent Events 创建流
func FromSSE(ctx context.Context, url string) (*AIStream, error) {
    s := NewAIStream()

    go func() {
        resp, err := http.Get(url)
        if err != nil {
            s.Publish(AIChunk{Type: "error", Content: err.Error()})
            s.Close()
            return
        }
        defer resp.Body.Close()

        scanner := bufio.NewScanner(resp.Body)
        for scanner.Scan() {
            select {
            case <-ctx.Done():
                s.Close()
                return
            default:
                line := scanner.Text()
                if strings.HasPrefix(line, "data: ") {
                    data := strings.TrimPrefix(line, "data: ")

                    var chunk AIChunk
                    if err := json.Unmarshal([]byte(data), &chunk); err == nil {
                        s.Publish(chunk)
                    }
                }
            }
        }
        s.Close()
    }()

    return s, nil
}

// FromOpenAIStream 从 OpenAI 流式响应创建流
func FromOpenAIStream(ctx context.Context, client *openai.Client, req openai.ChatCompletionRequest) (*AIStream, error) {
    s := NewAIStream()

    stream, err := client.CreateChatCompletionStream(ctx, req)
    if err != nil {
        return nil, err
    }

    go func() {
        defer stream.Close()
        defer s.Close()

        for {
            select {
            case <-ctx.Done():
                return
            default:
                resp, err := stream.Recv()
                if err != nil {
                    if err == io.EOF {
                        s.Publish(AIChunk{Type: "done"})
                    } else {
                        s.Publish(AIChunk{Type: "error", Content: err.Error()})
                    }
                    return
                }

                for _, choice := range resp.Choices {
                    if len(choice.Delta.Content) > 0 {
                        s.Publish(AIChunk{
                            Type:    "text",
                            Content: choice.Delta.Content,
                        })
                    }
                }
            }
        }
    }()

    return s, nil
}
```

## 使用示例

### 示例 1：基础流式显示

```go
// ✅ 创建流式显示
stream := stream.NewStream()

viewer := component.NewStreamViewer(stream)
viewer.SetBounds(runtime.Rect{X: 0, Y: 0, Width: 80, Height: 24})
viewer.Start()

app.Mount(viewer)

// 发布数据
go func() {
    for i := 0; i < 100; i++ {
        stream.Publish(fmt.Sprintf("Line %d", i))
        time.Sleep(100 * time.Millisecond)
    }
    stream.Close()
}()
```

### 示例 2：命令行输出流

```go
// ✅ 实时显示命令输出
cmd := exec.Command("tail", "-f", "/var/log/syslog")

logStream, err := stream.CommandStream(context.Background(), cmd)
if err != nil {
    log.Fatal(err)
}

viewer := component.NewStreamViewer(logStream)
viewer.Start()
app.Mount(viewer)
```

### 示例 3：AI 响应流

```go
// ✅ 显示 AI 流式响应
aiStream, err := stream.FromOpenAIStream(
    context.Background(),
    client,
    openai.ChatCompletionRequest{
        Model:     openai.GPT4,
        Messages:  []openai.ChatCompletionMessage{{Role: "user", Content: "Hello"}},
        Stream:    true,
    },
)

if err != nil {
    log.Fatal(err)
}

viewer := component.NewStreamViewer(aiStream)
viewer.SetFormatter(func(data interface{}) string {
    if chunk, ok := data.(stream.AIChunk); ok {
        if chunk.Type == "text" {
            return chunk.Content.(string)
        }
    }
    return ""
})
viewer.Start()
app.Mount(viewer)
```

### 示例 4：多路复用

```go
// ✅ 合并多个流
merged := stream.NewMergedStream()

// 添加多个流
merged.Add(logStream)
merged.Add(metricStream)
merged.Add(eventStream)

viewer := component.NewStreamViewer(merged)
viewer.SetFormatter(func(data interface{}) string {
    // 根据数据类型格式化
    switch v := data.(type) {
    case LogEntry:
        return fmt.Sprintf("[LOG] %s %s", v.Time, v.Message)
    case Metric:
        return fmt.Sprintf("[METRIC] %s: %v", v.Name, v.Value)
    case Event:
        return fmt.Sprintf("[EVENT] %s", v.Name)
    default:
        return fmt.Sprintf("%v", v)
    }
})
viewer.Start()
app.Mount(viewer)
```

### 示例 5：过滤和转换

```go
// ✅ 流处理管道
original := stream.NewStream()

// 只保留错误级别
filtered := stream.Filter(original, func(data interface{}) bool {
    if log, ok := data.(LogEntry); ok {
        return log.Level == "ERROR"
    }
    return false
})

// 转换格式
transformed := stream.Map(filtered, func(data interface{}) interface{} {
    if log, ok := data.(LogEntry); ok {
        return fmt.Sprintf("❌ %s: %s", log.Time, log.Message)
    }
    return data
})

viewer := component.NewStreamViewer(transformed)
viewer.Start()
app.Mount(viewer)
```

## 与 Action 系统集成

```go
// 位于: tui/framework/action/stream_actions.go

package action

const (
    // 流式相关 Action
    ActionStreamStart   ActionType = "stream.start"
    ActionStreamStop    ActionType = "stream.stop"
    ActionStreamPause   ActionType = "stream.pause"
    ActionStreamResume  ActionType = "stream.resume"
    ActionStreamClear   ActionType = "stream.clear"
    ActionStreamScroll  ActionType = "stream.scroll"
)
```

## 测试

```go
// 位于: tui/framework/stream/stream_test.go

func TestStreamPublish(t *testing.T) {
    s := stream.NewStream()
    sub := s.Subscribe()
    defer sub.Unsubscribe()

    s.Publish("test")

    select {
    case data := <-sub.Chan():
        assert.Equal(t, "test", data)
    case <-time.After(time.Second):
        t.Fatal("timeout")
    }
}

func TestStreamBuffer(t *testing.T) {
    buf := stream.NewStreamBuffer(10, true)

    for i := 0; i < 20; i++ {
        buf.Append(stream.Chunk{Data: i})
    }

    // 应该只保留最后 10 个
    assert.Equal(t, 10, buf.Size())
}

func TestStreamViewer(t *testing.T) {
    s := stream.NewStream()
    viewer := component.NewStreamViewer(s)
    viewer.Start()

    // 发布数据
    for i := 0; i < 5; i++ {
        s.Publish(fmt.Sprintf("Line %d", i))
    }

    // 等待处理
    time.Sleep(100 * time.Millisecond)

    assert.Equal(t, 5, viewer.Buffer().Size())

    viewer.Stop()
}
```

## 总结

流式数据系统提供：

1. **实时显示**: 数据到达立即显示
2. **增量更新**: 只更新变化部分
3. **自动滚动**: 自动滚动到最新数据
4. **缓冲控制**: 防止内存溢出
5. **多数据源**: 支持多种数据源
6. **流处理**: 支持过滤、转换等操作

## 相关文档

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 架构概览
- [COMPONENTS.md](./COMPONENTS.md) - 组件系统
- [ASYNC_TASK.md](./ASYNC_TASK.md) - 异步任务
- [ACTION_SYSTEM.md](./ACTION_SYSTEM.md) - Action 系统
