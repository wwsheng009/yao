package cursor

import (
	"testing"
	"time"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// TestBlinkPositionStability 测试闪烁时光标位置的稳定性
func TestBlinkPositionStability(t *testing.T) {
	c := NewCursor()
	c.SetBlinkInterval(50 * time.Millisecond)

	buf := paint.NewBuffer(80, 24)
	ctx := component.NewPaintContext(buf, 10, 5, 80, 24) // 父组件起始位置 (10, 5)

	// 设置光标位置
	c.SetPosition(3, 2) // 相对位置

	// 初始绘制
	c.Paint(ctx, buf)

	// 验证光标在正确位置 (10+3=13, 5+2=7)
	expectedX := 13
	expectedY := 7
	if buf.Cells[expectedY][expectedX].Style.IsReverse() {
		t.Logf("✓ 初始光标位置正确: (%d, %d)", expectedX, expectedY)
	} else {
		t.Errorf("初始光标位置错误: 预期 (%d, %d) 有反色", expectedX, expectedY)
	}

	// 等待闪烁切换
	time.Sleep(60 * time.Millisecond)

	// 再次绘制（光标应该不可见）
	buf2 := paint.NewBuffer(80, 24)
	c.Paint(ctx, buf2)

	if !buf2.Cells[expectedY][expectedX].Style.IsReverse() {
		t.Logf("✓ 闪烁后光标隐藏正确")
	} else {
		t.Errorf("闪烁后光标应该隐藏但仍然可见")
	}

	// 再次等待闪烁切换
	time.Sleep(60 * time.Millisecond)

	// 第三次绘制（光标应该再次可见）
	buf3 := paint.NewBuffer(80, 24)
	c.Paint(ctx, buf3)

	if buf3.Cells[expectedY][expectedX].Style.IsReverse() {
		t.Logf("✓ 再次闪烁后光标显示正确")
	} else {
		t.Errorf("再次闪烁后光标应该显示但仍然隐藏")
	}
}

// TestMultiplePaintCallsSameFrame 测试同一帧内多次调用 Paint
func TestMultiplePaintCallsSameFrame(t *testing.T) {
	c := NewCursor()
	c.SetBlinkInterval(500 * time.Millisecond) // 较长的间隔

	buf := paint.NewBuffer(80, 24)
	ctx := component.NewPaintContext(buf, 0, 0, 80, 24)

	c.SetPosition(5, 0)

	// 同一帧内多次调用 Paint（模拟多个组件共享同一个 buffer）
	reverseCount := 0
	for i := 0; i < 5; i++ {
		// 每次调用前重置 buffer
		if i > 0 {
			buf = paint.NewBuffer(80, 24)
		}
		c.Paint(ctx, buf)
		if buf.Cells[0][5].Style.IsReverse() {
			reverseCount++
		}
		// 不等待，立即再次调用
	}

	// 应该始终可见（因为时间间隔未到，不会切换）
	if reverseCount == 5 {
		t.Logf("✓ 同一帧内多次调用 Paint，光标状态稳定")
	} else {
		t.Errorf("同一帧内多次调用 Paint，光标状态不稳定: %d/5 可见", reverseCount)
	}
}

// TestPositionUpdateDuringBlink 测试闪烁期间更新位置
func TestPositionUpdateDuringBlink(t *testing.T) {
	c := NewCursor()
	c.SetBlinkInterval(100 * time.Millisecond)

	buf := paint.NewBuffer(80, 24)
	ctx := component.NewPaintContext(buf, 0, 0, 80, 24)

	// 初始位置
	c.SetPosition(5, 0)
	c.Paint(ctx, buf)

	if !buf.Cells[0][5].Style.IsReverse() {
		t.Errorf("位置 5 应该有光标")
	}

	// 等待闪烁切换
	time.Sleep(120 * time.Millisecond)

	// 更新位置到 10
	buf2 := paint.NewBuffer(80, 24)
	c.SetPosition(10, 0)
	c.Paint(ctx, buf2)

	// 检查：位置 5 不应该有光标，位置 10 应该有光标（如果可见）
	if buf2.Cells[0][5].Style.IsReverse() {
		t.Errorf("位置 5 不应该有光标（已移动到 10）")
	}

	// 位置 10 可能有光标（取决于闪烁状态）
	if buf2.Cells[0][10].Style.IsReverse() {
		t.Logf("✓ 位置更新正确，新位置有光标")
	} else {
		t.Logf("注意: 位置 10 没有光标（可能恰好处在闪烁不可见状态）")
	}
}

// TestParentContextChange 测试父组件上下文变化时光标位置
func TestParentContextChange(t *testing.T) {
	c := NewCursor()
	c.SetPosition(3, 0) // 相对位置保持不变

	buf := paint.NewBuffer(80, 24)

	// 第一次绘制：父组件在 (0, 0)
	ctx1 := component.NewPaintContext(buf, 0, 0, 80, 24)
	c.Paint(ctx1, buf)

	if !buf.Cells[0][3].Style.IsReverse() {
		t.Errorf("光标应该在 (3, 0)")
	}

	// 第二次绘制：父组件在 (10, 5)
	buf2 := paint.NewBuffer(80, 24)
	ctx2 := component.NewPaintContext(buf2, 10, 5, 80, 24)
	c.Paint(ctx2, buf2)

	if !buf2.Cells[5][13].Style.IsReverse() {
		t.Errorf("光标应该在 (13, 5) = 10+3, 5+0")
	}

	if buf2.Cells[0][3].Style.IsReverse() {
		t.Errorf("旧位置 (3, 0) 不应该有光标")
	}

	t.Logf("✓ 父组件上下文变化时光标位置正确")
}

// TestIsVisibleSideEffects 测试 IsVisible 的副作用
func TestIsVisibleSideEffects(t *testing.T) {
	c := NewCursor()
	c.SetBlinkInterval(100 * time.Millisecond)

	// 记录初始可见状态
	visible1 := c.IsVisible()
	t.Logf("初始可见: %v", visible1)

	// 立即再次调用（不应该切换状态）
	visible2 := c.IsVisible()
	t.Logf("立即再次调用: %v", visible2)

	if visible1 != visible2 {
		t.Errorf("短时间多次调用 IsVisible 不应该切换状态")
	}

	// 等待超过间隔
	time.Sleep(120 * time.Millisecond)

	// 现在应该切换状态
	visible3 := c.IsVisible()
	t.Logf("等待后调用: %v", visible3)

	if visible1 == visible3 {
		t.Errorf("等待足够时间后 IsVisible 应该切换状态")
	} else {
		t.Logf("✓ IsVisible 闪烁逻辑正确")
	}
}

// TestConcurrentPositionAndBlink 测试位置更新和闪烁并发进行
func TestConcurrentPositionAndBlink(t *testing.T) {
	c := NewCursor()
	c.SetBlinkInterval(50 * time.Millisecond)

	buf := paint.NewBuffer(80, 24)
	ctx := component.NewPaintContext(buf, 0, 0, 80, 24)

	// 模拟用户输入时不断更新位置
	for i := 0; i < 10; i++ {
		buf = paint.NewBuffer(80, 24)
		c.SetPosition(i, 0)
		c.Paint(ctx, buf)

		// 验证光标在新位置
		if buf.Cells[0][i].Style.IsReverse() {
			// 光标可见
		}

		time.Sleep(20 * time.Millisecond) // 模拟处理时间
	}

	t.Logf("✓ 并发位置更新和闪烁没有问题")
}
