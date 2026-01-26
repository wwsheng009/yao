package event

import (
	"bytes"
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestFilterChain_Process(t *testing.T) {
	chain := NewFilterChain()

	// 测试空链
	event := NewBaseEvent(EventKeyPress)
	ctx := NewContext()

	filtered, proceed := chain.Process(ctx, event)
	if !proceed {
		t.Error("empty chain should allow event to proceed")
	}
	if filtered != event {
		t.Error("event should be unchanged")
	}
}

func TestFilterChain_Add(t *testing.T) {
	chain := NewFilterChain()
	filter := &Passthrough{}

	chain.Add(filter)
	if chain.Count() != 1 {
		t.Errorf("expected 1 filter, got %d", chain.Count())
	}

	chain.Add(filter)
	if chain.Count() != 2 {
		t.Errorf("expected 2 filters, got %d", chain.Count())
	}
}

func TestFilterChain_Clear(t *testing.T) {
	chain := NewFilterChain()
	chain.Add(&Passthrough{})
	chain.Add(&Passthrough{})

	chain.Clear()
	if chain.Count() != 0 {
		t.Errorf("expected 0 filters after clear, got %d", chain.Count())
	}
}

func TestPassthroughFilter(t *testing.T) {
	filter := NewPassthrough()
	event := NewBaseEvent(EventKeyPress)
	ctx := NewContext()

	filtered, proceed := filter.Filter(ctx, event)
	if !proceed {
		t.Error("Passthrough should allow all events")
	}
	if filtered != event {
		t.Error("event should be unchanged")
	}
}

func TestDiscardFilter(t *testing.T) {
	filter := NewDiscard()
	event := NewBaseEvent(EventKeyPress)
	ctx := NewContext()

	_, proceed := filter.Filter(ctx, event)
	if proceed {
		t.Error("Discard should block all events")
	}
}

func TestLoggingFilter(t *testing.T) {
	buf := &bytes.Buffer{}
	filter := NewLoggingFilter(buf)
	event := NewBaseEvent(EventKeyPress)
	ctx := NewContext()

	filtered, proceed := filter.Filter(ctx, event)
	if !proceed {
		t.Error("LoggingFilter should allow all events")
	}
	if filtered != event {
		t.Error("event should be unchanged")
	}

	output := buf.String()
	if !strings.Contains(output, "type=") {
		t.Error("log should contain event type")
	}
}

func TestLoggingFilterForTypes(t *testing.T) {
	buf := &bytes.Buffer{}
	filter := NewLoggingFilterForTypes(buf, []EventType{EventKeyPress})

	event1 := NewBaseEvent(EventKeyPress)
	event2 := NewBaseEvent(EventKeyRelease)
	ctx := NewContext()

	// 应该记录 EventKeyPress
	filter.Filter(ctx, event1)
	if buf.Len() == 0 {
		t.Error("should log EventKeyPress")
	}

	buf.Reset()

	// 不应该记录 EventKeyRelease
	filter.Filter(ctx, event2)
	if buf.Len() > 0 {
		t.Error("should not log EventKeyRelease")
	}
}

func TestMetricsFilter(t *testing.T) {
	filter := NewMetricsFilter()

	event1 := NewBaseEvent(EventKeyPress)
	event2 := NewBaseEvent(EventKeyRelease)
	ctx := NewContext()

	filter.Filter(ctx, event1)
	filter.Filter(ctx, event2)

	counts := filter.GetCounts()
	if counts[EventKeyPress] != 1 {
		t.Errorf("expected 1 EventKeyPress, got %d", counts[EventKeyPress])
	}
	if counts[EventKeyRelease] != 1 {
		t.Errorf("expected 1 EventKeyRelease, got %d", counts[EventKeyRelease])
	}

	stats := filter.Stats()
	if stats["total_events"] != 2 {
		t.Errorf("expected 2 total events, got %v", stats["total_events"])
	}

	filter.Reset()
	if len(filter.GetCounts()) != 0 {
		t.Error("counts should be empty after reset")
	}
}

func TestRateLimitFilter(t *testing.T) {
	filter := NewRateLimitFilter()

	// 每秒最多 2 个事件
	limit := NewRateLimit(time.Second, 2)
	filter.SetLimit(EventKeyPress, limit)

	event := NewBaseEvent(EventKeyPress)
	ctx := NewContext()

	// 前 2 个事件应该通过
	for i := 0; i < 2; i++ {
		_, proceed := filter.Filter(ctx, event)
		if !proceed {
			t.Errorf("event %d should proceed", i)
		}
	}

	// 第 3 个事件应该被拦截
	_, proceed := filter.Filter(ctx, event)
	if proceed {
		t.Error("third event should be rate limited")
	}
}

func TestRateLimitFilter_Reset(t *testing.T) {
	filter := NewRateLimitFilter()

	// 每 100ms 最多 1 个事件
	limit := NewRateLimit(100*time.Millisecond, 1)
	filter.SetLimit(EventKeyPress, limit)

	event := NewBaseEvent(EventKeyPress)
	ctx := NewContext()

	// 第一个事件通过
	_, proceed := filter.Filter(ctx, event)
	if !proceed {
		t.Error("first event should proceed")
	}

	// 第二个事件被拦截
	_, proceed = filter.Filter(ctx, event)
	if proceed {
		t.Error("second event should be rate limited")
	}

	// 等待限制重置
	time.Sleep(150 * time.Millisecond)

	// 现在应该可以通过了
	_, proceed = filter.Filter(ctx, event)
	if !proceed {
		t.Error("event should proceed after reset")
	}
}

func TestTransformFilter(t *testing.T) {
	called := false
	transformer := func(e Event) Event {
		called = true
		// 修改事件类型
		base := e.(*BaseEvent)
		base.eventType = EventKeyRelease
		return base
	}

	filter := NewTransformFilter(transformer)
	event := NewBaseEvent(EventKeyPress)
	ctx := NewContext()

	filtered, proceed := filter.Filter(ctx, event)
	if !proceed {
		t.Error("TransformFilter should allow event to proceed")
	}
	if !called {
		t.Error("transformer should be called")
	}
	if filtered.Type() != EventKeyRelease {
		t.Errorf("event type should be changed to %d, got %d", EventKeyRelease, filtered.Type())
	}
}

func TestPermissionFilter(t *testing.T) {
	checker := func(ctx *Context, event Event) bool {
		return event.Type() == EventKeyPress
	}

	filter := NewPermissionFilter(checker)
	ctx := NewContext()

	// 允许的事件
	event1 := NewBaseEvent(EventKeyPress)
	_, proceed := filter.Filter(ctx, event1)
	if !proceed {
		t.Error("EventKeyPress should be allowed")
	}

	// 拒绝的事件
	event2 := NewBaseEvent(EventKeyRelease)
	_, proceed = filter.Filter(ctx, event2)
	if proceed {
		t.Error("EventKeyRelease should be blocked")
	}
}

func TestContextFilter(t *testing.T) {
	filter := NewContextFilter()

	// 正常上下文
	ctx1 := NewContext()
	event := NewBaseEvent(EventKeyPress)

	_, proceed := filter.Filter(ctx1, event)
	if !proceed {
		t.Error("event should proceed with normal context")
	}

	// 已取消的上下文
	ctx2, cancel := context.WithCancel(context.Background())
	cancel()

	ctx2context := NewContextWithContext(ctx2)
	_, proceed = filter.Filter(ctx2context, event)
	if proceed {
		t.Error("event should be blocked with canceled context")
	}
}

func TestOnceFilter(t *testing.T) {
	var count int32
	filter := NewOnceFilter(FilterFunc(func(ctx *Context, event Event) (Event, bool) {
		atomic.AddInt32(&count, 1)
		return event, true
	}))

	event := NewBaseEvent(EventKeyPress)
	ctx := NewContext()

	// 第一次执行
	filter.Filter(ctx, event)
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	// 第二次不执行
	filter.Filter(ctx, event)
	if count != 1 {
		t.Errorf("expected count still 1, got %d", count)
	}

	// 重置后可以再次执行
	filter.Reset()
	filter.Filter(ctx, event)
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

func TestConditionalFilter(t *testing.T) {
	var called bool
	condition := func(ctx *Context, event Event) bool {
		return event.Type() == EventKeyPress
	}

	filter := NewConditionalFilter(condition, FilterFunc(func(ctx *Context, event Event) (Event, bool) {
		called = true
		return event, true
	}))

	ctx := NewContext()

	// 满足条件
	event1 := NewBaseEvent(EventKeyPress)
	filter.Filter(ctx, event1)
	if !called {
		t.Error("filter should be called when condition is met")
	}

	called = false

	// 不满足条件
	event2 := NewBaseEvent(EventKeyRelease)
	filter.Filter(ctx, event2)
	if called {
		t.Error("filter should not be called when condition is not met")
	}
}

func TestComposeFilter(t *testing.T) {
	var order []int
	filter1 := FilterFunc(func(ctx *Context, event Event) (Event, bool) {
		order = append(order, 1)
		return event, true
	})
	filter2 := FilterFunc(func(ctx *Context, event Event) (Event, bool) {
		order = append(order, 2)
		return event, true
	})

	filter := NewComposeFilter(filter1, filter2)

	event := NewBaseEvent(EventKeyPress)
	ctx := NewContext()

	filter.Filter(ctx, event)

	if len(order) != 2 {
		t.Fatalf("expected 2 filters called, got %d", len(order))
	}
	if order[0] != 1 || order[1] != 2 {
		t.Errorf("filters should be called in order, got %v", order)
	}
}

func TestComposeFilter_Block(t *testing.T) {
	var secondCalled bool
	filter1 := FilterFunc(func(ctx *Context, event Event) (Event, bool) {
		return event, false // 拦截
	})
	filter2 := FilterFunc(func(ctx *Context, event Event) (Event, bool) {
		secondCalled = true
		return event, true
	})

	filter := NewComposeFilter(filter1, filter2)

	event := NewBaseEvent(EventKeyPress)
	ctx := NewContext()

	_, proceed := filter.Filter(ctx, event)
	if proceed {
		t.Error("event should be blocked")
	}
	if secondCalled {
		t.Error("second filter should not be called when first blocks")
	}
}

func TestAnyFilter(t *testing.T) {
	filter1 := FilterFunc(func(ctx *Context, event Event) (Event, bool) {
		return event, false // 拦截
	})
	filter2 := FilterFunc(func(ctx *Context, event Event) (Event, bool) {
		return event, true // 通过
	})

	filter := NewAnyFilter(filter1, filter2)

	event := NewBaseEvent(EventKeyPress)
	ctx := NewContext()

	// AnyFilter 只有在所有过滤器都拦截时才拦截
	_, proceed := filter.Filter(ctx, event)
	if !proceed {
		t.Error("event should proceed when at least one filter allows")
	}
}

func TestAnyFilter_AllBlock(t *testing.T) {
	filter1 := FilterFunc(func(ctx *Context, event Event) (Event, bool) {
		return event, false // 拦截
	})
	filter2 := FilterFunc(func(ctx *Context, event Event) (Event, bool) {
		return event, false // 拦截
	})

	filter := NewAnyFilter(filter1, filter2)

	event := NewBaseEvent(EventKeyPress)
	ctx := NewContext()

	_, proceed := filter.Filter(ctx, event)
	if proceed {
		t.Error("event should be blocked when all filters block")
	}
}

func TestContext_WithValue(t *testing.T) {
	ctx := NewContext()
	ctx.Set("key1", "value1")
	ctx.Set("key2", 42)

	if ctx.Get("key1") != "value1" {
		t.Error("key1 should be value1")
	}
	if ctx.Get("key2") != 42 {
		t.Error("key2 should be 42")
	}
}

func TestContext_WithChain(t *testing.T) {
	ctx := NewContext().
		With("key1", "value1").
		With("key2", "value2").
		WithSource("test-source")

	if ctx.Get("key1") != "value1" {
		t.Error("key1 should be value1")
	}
	if ctx.Source() != "test-source" {
		t.Error("source should be test-source")
	}
}

func TestContext_Clone(t *testing.T) {
	ctx := NewContext()
	ctx.Set("key", "value")
	ctx.source = "source1"

	cloned := ctx.Clone()
	cloned.Set("key", "changed")
	cloned.source = "source2"

	// 原始值不应该改变
	if ctx.Get("key") != "value" {
		t.Error("original context should not be affected")
	}
	if ctx.source != "source1" {
		t.Error("original source should not be affected")
	}

	// 克隆的值应该改变
	if cloned.Get("key") != "changed" {
		t.Error("cloned context should have new value")
	}
	if cloned.source != "source2" {
		t.Error("cloned source should be changed")
	}
}

func TestContext_WithGoContext(t *testing.T) {
	goCtx := context.WithValue(context.Background(), "user", "alice")
	ctx := NewContextWithContext(goCtx)

	if ctx.Context().Value("user") != "alice" {
		t.Error("Go context value should be preserved")
	}
}

func TestFilterFunc(t *testing.T) {
	var called bool
	fn := FilterFunc(func(ctx *Context, event Event) (Event, bool) {
		called = true
		return event, true
	})

	fn.Filter(NewContext(), NewBaseEvent(EventKeyPress))
	if !called {
		t.Error("FilterFunc should call the underlying function")
	}
}

func TestLogHelpers(t *testing.T) {
	// LogToStdout
	filter1 := LogToStdout()
	if filter1 == nil {
		t.Error("LogToStdout should return a filter")
	}

	// LogToStderr
	filter2 := LogToStderr()
	if filter2 == nil {
		t.Error("LogToStderr should return a filter")
	}
}

func TestFilterChain_Remove(t *testing.T) {
	chain := NewFilterChain()
	passthrough := NewPassthrough()
	discard := NewDiscard()

	chain.Add(passthrough)
	chain.Add(discard)

	if chain.Count() != 2 {
		t.Errorf("expected 2 filters, got %d", chain.Count())
	}

	// 移除 Passthrough
	chain.Remove(passthrough)

	if chain.Count() != 1 {
		t.Errorf("expected 1 filter after removing Passthrough, got %d", chain.Count())
	}
}
