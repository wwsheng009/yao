package event

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// ==============================================================================
// Event Filter System (V3)
// ==============================================================================
// 事件过滤器允许在事件处理流程中插入拦截器
// 用于日志记录、权限检查、消息转换等场景

// Filter 事件过滤器接口
// 返回 (event, proceed):
//   - event: 过滤后的事件（可能被修改）
//   - proceed: 是否继续传播
type Filter interface {
	// Filter 处理事件
	Filter(ctx *Context, event Event) (Event, bool)
}

// FilterFunc 函数式过滤器
type FilterFunc func(ctx *Context, event Event) (Event, bool)

// Filter 实现 Filter 接口
func (f FilterFunc) Filter(ctx *Context, event Event) (Event, bool) {
	return f(ctx, event)
}

// FilterChain 过滤器链
type FilterChain struct {
	mu      sync.RWMutex
	filters []Filter
}

// NewFilterChain 创建过滤器链
func NewFilterChain(filters ...Filter) *FilterChain {
	return &FilterChain{
		filters: filters,
	}
}

// Add 添加过滤器
func (c *FilterChain) Add(filter Filter) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.filters = append(c.filters, filter)
}

// Remove 移除过滤器（按类型）
func (c *FilterChain) Remove(filterType Filter) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, f := range c.filters {
		if fmt.Sprintf("%T", f) == fmt.Sprintf("%T", filterType) {
			c.filters = append(c.filters[:i], c.filters[i+1:]...)
			break
		}
	}
}

// Clear 清空所有过滤器
func (c *FilterChain) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.filters = c.filters[:0]
}

// Process 处理事件
// 返回 (event, proceed):
//   - event: 过滤后的事件
//   - proceed: 是否继续传播
func (c *FilterChain) Process(ctx *Context, event Event) (Event, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	current := event
	for _, filter := range c.filters {
		e, proceed := filter.Filter(ctx, current)
		if !proceed {
			return e, false // 被拦截
		}
		current = e
	}
	return current, true
}

// Count 返回过滤器数量
func (c *FilterChain) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.filters)
}

// ==============================================================================
// Built-in Filters
// ==============================================================================

// LoggingFilter 日志过滤器
type LoggingFilter struct {
	logger *log.Logger
	events []EventType // 只记录特定类型，空表示全部
}

// NewLoggingFilter 创建日志过滤器
func NewLoggingFilter(writer io.Writer) *LoggingFilter {
	return &LoggingFilter{
		logger: log.New(writer, "[EVENT] ", log.LstdFlags|log.Lmicroseconds),
	}
}

// NewLoggingFilterForTypes 创建只记录特定类型的日志过滤器
func NewLoggingFilterForTypes(writer io.Writer, events []EventType) *LoggingFilter {
	return &LoggingFilter{
		logger: log.New(writer, "[EVENT] ", log.LstdFlags|log.Lmicroseconds),
		events: events,
	}
}

// Filter 实现过滤器接口
func (f *LoggingFilter) Filter(ctx *Context, event Event) (Event, bool) {
	// 检查是否需要记录此类型
	if len(f.events) > 0 {
		found := false
		for _, et := range f.events {
			if et == event.Type() {
				found = true
				break
			}
		}
		if !found {
			return event, true
		}
	}

	f.logger.Printf("type=%d phase=%d target=%v",
		event.Type(), event.Phase(), event.Target())
	return event, true
}

// MetricsFilter 指标收集过滤器
type MetricsFilter struct {
	mu     sync.RWMutex
	counts map[EventType]int
	times  map[EventType]time.Duration
}

// NewMetricsFilter 创建指标过滤器
func NewMetricsFilter() *MetricsFilter {
	return &MetricsFilter{
		counts: make(map[EventType]int),
		times:  make(map[EventType]time.Duration),
	}
}

// Filter 实现过滤器接口
func (f *MetricsFilter) Filter(ctx *Context, event Event) (Event, bool) {
	start := time.Now()
	defer func() {
		f.mu.Lock()
		defer f.mu.Unlock()

		et := event.Type()
		f.counts[et]++
		f.times[et] += time.Since(start)
	}()

	return event, true
}

// GetCounts 获取事件计数
func (f *MetricsFilter) GetCounts() map[EventType]int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	counts := make(map[EventType]int, len(f.counts))
	for k, v := range f.counts {
		counts[k] = v
	}
	return counts
}

// GetTimes 获取事件处理时间
func (f *MetricsFilter) GetTimes() map[EventType]time.Duration {
	f.mu.RLock()
	defer f.mu.RUnlock()

	times := make(map[EventType]time.Duration, len(f.times))
	for k, v := range f.times {
		times[k] = v
	}
	return times
}

// Reset 重置指标
func (f *MetricsFilter) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.counts = make(map[EventType]int)
	f.times = make(map[EventType]time.Duration)
}

// Stats 返回统计信息
func (f *MetricsFilter) Stats() map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	totalEvents := 0
	for _, count := range f.counts {
		totalEvents += count
	}

	var totalTime time.Duration
	for _, t := range f.times {
		totalTime += t
	}

	return map[string]interface{}{
		"total_events": totalEvents,
		"total_time":   totalTime.String(),
		"types":        len(f.counts),
	}
}

// RateLimitFilter 频率限制过滤器
type RateLimitFilter struct {
	mu     sync.RWMutex
	limits map[EventType]*RateLimit
}

// RateLimit 频率限制配置
type RateLimit struct {
	Interval time.Duration
	MaxCount int
	current  int
	lastTime time.Time
	mu       sync.Mutex
}

// NewRateLimit 创建频率限制配置
func NewRateLimit(interval time.Duration, maxCount int) *RateLimit {
	return &RateLimit{
		Interval: interval,
		MaxCount: maxCount,
		lastTime: time.Now(),
	}
}

// NewRateLimitFilter 创建频率限制过滤器
func NewRateLimitFilter() *RateLimitFilter {
	return &RateLimitFilter{
		limits: make(map[EventType]*RateLimit),
	}
}

// SetLimit 设置特定事件的频率限制
func (f *RateLimitFilter) SetLimit(eventType EventType, limit *RateLimit) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.limits[eventType] = limit
}

// RemoveLimit 移除频率限制
func (f *RateLimitFilter) RemoveLimit(eventType EventType) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.limits, eventType)
}

// Filter 实现过滤器接口
func (f *RateLimitFilter) Filter(ctx *Context, event Event) (Event, bool) {
	f.mu.RLock()
	limit, exists := f.limits[event.Type()]
	f.mu.RUnlock()

	if !exists {
		return event, true
	}

	limit.mu.Lock()
	defer limit.mu.Unlock()

	now := time.Now()
	if now.Sub(limit.lastTime) >= limit.Interval {
		limit.current = 0
		limit.lastTime = now
	}

	if limit.current >= limit.MaxCount {
		return event, false // 超出限制，拦截
	}

	limit.current++
	return event, true
}

// TransformFilter 事件转换过滤器
type TransformFilter struct {
	transformer func(Event) Event
}

// NewTransformFilter 创建转换过滤器
func NewTransformFilter(transformer func(Event) Event) *TransformFilter {
	return &TransformFilter{transformer: transformer}
}

// Filter 实现过滤器接口
func (f *TransformFilter) Filter(ctx *Context, event Event) (Event, bool) {
	if f.transformer != nil {
		return f.transformer(event), true
	}
	return event, true
}

// PermissionFilter 权限控制过滤器
type PermissionFilter struct {
	checker func(ctx *Context, event Event) bool
}

// NewPermissionFilter 创建权限过滤器
func NewPermissionFilter(checker func(ctx *Context, event Event) bool) *PermissionFilter {
	return &PermissionFilter{checker: checker}
}

// Filter 实现过滤器接口
func (f *PermissionFilter) Filter(ctx *Context, event Event) (Event, bool) {
	if f.checker != nil {
		if !f.checker(ctx, event) {
			return event, false // 无权限，拦截
		}
	}
	return event, true
}

// ContextFilter 上下文过滤器 - 在上下文取消时拦截事件
type ContextFilter struct{}

// NewContextFilter 创建上下文过滤器
func NewContextFilter() *ContextFilter {
	return &ContextFilter{}
}

// Filter 实现过滤器接口
// 如果上下文已取消，拦截所有事件
func (f *ContextFilter) Filter(ctx *Context, event Event) (Event, bool) {
	if ctx == nil || ctx.context == nil {
		return event, true
	}

	select {
	case <-ctx.context.Done():
		return event, false // 上下文已取消，拦截事件
	default:
		return event, true
	}
}

// OnceFilter 只执行一次的过滤器
type OnceFilter struct {
	mu      sync.Mutex
	called  bool
	wrapped Filter
}

// NewOnceFilter 创建一次性过滤器
func NewOnceFilter(filter Filter) *OnceFilter {
	return &OnceFilter{
		wrapped: filter,
	}
}

// Filter 实现过滤器接口
func (f *OnceFilter) Filter(ctx *Context, event Event) (Event, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.called {
		return event, true // 已经执行过，跳过
	}

	f.called = true
	return f.wrapped.Filter(ctx, event)
}

// Reset 重置过滤器，允许再次执行
func (f *OnceFilter) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.called = false
}

// ConditionalFilter 条件过滤器
type ConditionalFilter struct {
	condition func(ctx *Context, event Event) bool
	filter    Filter
}

// NewConditionalFilter 创建条件过滤器
func NewConditionalFilter(condition func(ctx *Context, event Event) bool, filter Filter) *ConditionalFilter {
	return &ConditionalFilter{
		condition: condition,
		filter:    filter,
	}
}

// Filter 实现过滤器接口
func (f *ConditionalFilter) Filter(ctx *Context, event Event) (Event, bool) {
	if f.condition != nil && f.condition(ctx, event) {
		return f.filter.Filter(ctx, event)
	}
	return event, true
}

// ComposeFilter 组合多个过滤器（逻辑与）
// 所有过滤器都通过时才继续传播
type ComposeFilter struct {
	filters []Filter
}

// NewComposeFilter 创建组合过滤器
func NewComposeFilter(filters ...Filter) *ComposeFilter {
	return &ComposeFilter{filters: filters}
}

// Filter 实现过滤器接口
func (f *ComposeFilter) Filter(ctx *Context, event Event) (Event, bool) {
	current := event
	for _, filter := range f.filters {
		e, proceed := filter.Filter(ctx, current)
		if !proceed {
			return e, false
		}
		current = e
	}
	return current, true
}

// AnyFilter 任意过滤器（逻辑或）
// 任一过滤器拦截时才拦截
type AnyFilter struct {
	filters []Filter
}

// NewAnyFilter 创建任意过滤器
func NewAnyFilter(filters ...Filter) *AnyFilter {
	return &AnyFilter{filters: filters}
}

// Filter 实现过滤器接口
func (f *AnyFilter) Filter(ctx *Context, event Event) (Event, bool) {
	// 收集所有过滤器的结果
	var lastEvent Event
	allBlocked := true

	for _, filter := range f.filters {
		e, proceed := filter.Filter(ctx, event)
		lastEvent = e
		if proceed {
			allBlocked = false
		}
	}

	if allBlocked {
		return lastEvent, false
	}
	return event, true
}

// ==============================================================================
// Helper Functions
// ==============================================================================

// LogToFile 创建日志到文件的过滤器
func LogToFile(filename string) (*LoggingFilter, error) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return NewLoggingFilter(f), nil
}

// LogToStdout 创建日志到标准输出的过滤器
func LogToStdout() *LoggingFilter {
	return NewLoggingFilter(os.Stdout)
}

// LogToStderr 创建日志到标准错误的过滤器
func LogToStderr() *LoggingFilter {
	return NewLoggingFilter(os.Stderr)
}

// Discard 丢弃所有事件的过滤器
type Discard struct{}

// NewDiscard 创建丢弃过滤器
func NewDiscard() *Discard {
	return &Discard{}
}

// Filter 实现过滤器接口 - 拦截所有事件
func (d *Discard) Filter(ctx *Context, event Event) (Event, bool) {
	return event, false // 拦截所有事件
}

// Passthrough 透传所有事件的过滤器
type Passthrough struct{}

// NewPassthrough 创建透传过滤器
func NewPassthrough() *Passthrough {
	return &Passthrough{}
}

// Filter 实现过滤器接口 - 透传所有事件
func (p *Passthrough) Filter(ctx *Context, event Event) (Event, bool) {
	return event, true // 透传所有事件
}
