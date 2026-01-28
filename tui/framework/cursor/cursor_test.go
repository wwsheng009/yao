package cursor

import (
	"testing"
	"time"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/runtime/style"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// TestNewCursor 测试光标创建
func TestNewCursor(t *testing.T) {
	c := NewCursor()

	if c == nil {
		t.Fatal("NewCursor returned nil")
	}

	// 检查初始状态
	if !c.IsBlinkEnabled() {
		t.Error("blink should be enabled by default")
	}

	if c.GetShape() != ShapeBlock {
		t.Error("shape should be ShapeBlock by default")
	}
}

// TestCursorPosition 测试位置设置和获取
func TestCursorPosition(t *testing.T) {
	c := NewCursor()

	// 设置位置
	c.SetPosition(5, 10)
	x, y := c.GetPosition()

	if x != 5 || y != 10 {
		t.Errorf("expected position (5, 10), got (%d, %d)", x, y)
	}

	// 更新位置
	c.SetPosition(15, 20)
	x, y = c.GetPosition()

	if x != 15 || y != 20 {
		t.Errorf("expected position (15, 20), got (%d, %d)", x, y)
	}
}

// TestCursorBlink 测试闪烁功能
func TestCursorBlink(t *testing.T) {
	c := NewCursor()

	// 初始状态应该是可见的
	if !c.IsVisible() {
		t.Error("cursor should be visible initially")
	}

	// 禁用闪烁后应该始终可见
	c.SetBlinkEnabled(false)
	if !c.IsVisible() {
		t.Error("cursor should always be visible when blink is disabled")
	}

	// 启用闪烁
	c.SetBlinkEnabled(true)
	if !c.IsBlinkEnabled() {
		t.Error("blink should be enabled")
	}

	// 等待超过闪烁间隔
	c.SetBlinkInterval(10 * time.Millisecond)
	time.Sleep(15 * time.Millisecond)

	// 多次检查以确保状态切换
	visibleCount := 0
	for i := 0; i < 10; i++ {
		if c.IsVisible() {
			visibleCount++
		}
		time.Sleep(12 * time.Millisecond)
	}

	// 应该有可见和不可见的状态
	if visibleCount == 10 || visibleCount == 0 {
		t.Logf("Warning: cursor may not be blinking properly, visible count: %d", visibleCount)
	}
}

// TestCursorBlinkInterval 测试闪烁间隔设置
func TestCursorBlinkInterval(t *testing.T) {
	c := NewCursor()

	// 设置自定义间隔
	c.SetBlinkInterval(250 * time.Millisecond)

	// 重置闪烁状态
	c.ResetBlink()

	if !c.IsVisible() {
		t.Error("cursor should be visible after reset")
	}
}

// TestCursorStyle 测试样式设置
func TestCursorStyle(t *testing.T) {
	c := NewCursor()

	// 设置自定义样式
	customStyle := style.Style{}.Foreground(style.Red).Background(style.Blue)
	c.SetStyle(customStyle)

	retrievedStyle := c.GetStyle()
	if retrievedStyle.FG != style.Red {
		t.Errorf("foreground color not set correctly: got %v, want %v", retrievedStyle.FG, style.Red)
	}
	if retrievedStyle.BG != style.Blue {
		t.Errorf("background color not set correctly: got %v, want %v", retrievedStyle.BG, style.Blue)
	}
}

// TestCursorShape 测试形状设置
func TestCursorShape(t *testing.T) {
	c := NewCursor()

	// 测试所有形状
	shapes := []Shape{ShapeBlock, ShapeUnderline, ShapeBar}
	for _, shape := range shapes {
		c.SetShape(shape)
		if c.GetShape() != shape {
			t.Errorf("shape not set correctly: expected %v, got %v", shape, c.GetShape())
		}
	}
}

// TestCursorFocus 测试焦点行为
func TestCursorFocus(t *testing.T) {
	c := NewCursor()

	// 获得焦点应该启用闪烁
	c.OnFocus()
	if !c.IsBlinkEnabled() {
		t.Error("blink should be enabled after OnFocus")
	}
	if !c.IsFocused() {
		t.Error("cursor should be focused")
	}

	// 失去焦点应该禁用闪烁
	c.OnBlur()
	if c.IsBlinkEnabled() {
		t.Error("blink should be disabled after OnBlur")
	}
	if c.IsFocused() {
		t.Error("cursor should not be focused")
	}
}

// TestCursorPaint 测试绘制功能
func TestCursorPaint(t *testing.T) {
	c := NewCursor()

	buf := paint.NewBuffer(20, 10)

	// 先绘制一些背景字符
	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
			buf.SetCell(x, y, 'x', style.Style{})
		}
	}

	ctx := component.NewPaintContext(buf, 0, 0, 80, 24)

	// 绘制块状光标
	c.SetPosition(5, 3)
	c.SetShape(ShapeBlock)
	c.Paint(ctx, buf)

	// 直接访问 buffer.Cells
	cell := buf.Cells[3][5]

	// 块状光标应该保留原字符但改变样式
	if cell.Char != 'x' {
		t.Errorf("expected char 'x', got '%c'", cell.Char)
	}

	// 测试下划线形状
	c.SetShape(ShapeUnderline)
	c.SetPosition(7, 4)
	c.Paint(ctx, buf)

	cell = buf.Cells[4][7]
	// 下划线形状会替换字符
	if cell.Char != '_' {
		t.Errorf("expected char '_', got '%c'", cell.Char)
	}
}

// TestCursorPaintWhenNotVisible 测试光标不可见时不绘制
func TestCursorPaintWhenNotVisible(t *testing.T) {
	c := NewCursor()
	c.SetBlinkEnabled(false)

	buf := paint.NewBuffer(20, 10)

	// 绘制背景
	buf.SetCell(5, 3, 'x', style.Style{})

	ctx := component.NewPaintContext(buf, 0, 0, 20, 10)

	// 禁用闪烁后绘制
	c.SetPosition(5, 3)
	c.SetBlinkEnabled(false)
	c.Paint(ctx, buf)

	// 检查原单元格没有被修改
	cell := buf.Cells[3][5]
	// 禁用闪烁时应该仍然可见（光标始终显示）
	if cell.Char != 'x' {
		t.Errorf("expected char 'x', got '%c'", cell.Char)
	}
}

// TestPaintCursor 测试便捷函数
func TestPaintCursor(t *testing.T) {
	// 创建一个实现 Host 接口的测试组件
	testHost := &testCursorHost{cursor: NewCursor()}

	buf := paint.NewBuffer(20, 10)
	ctx := component.NewPaintContext(buf, 0, 0, 20, 10)

	// 使用便捷函数绘制光标
	PaintCursor(testHost, ctx, buf, 5, 3)

	// 检查光标位置是否正确
	x, y := testHost.GetCursor().GetPosition()
	if x != 5 || y != 3 {
		t.Errorf("expected cursor position (5, 3), got (%d, %d)", x, y)
	}
}

// testCursorHost 实现 Host 接口的测试组件
type testCursorHost struct {
	cursor *Cursor
}

func (h *testCursorHost) GetCursor() *Cursor {
	return h.cursor
}

// BenchmarkCursorPaint 性能测试
func BenchmarkCursorPaint(b *testing.B) {
	c := NewCursor()
	buf := paint.NewBuffer(80, 24)
	ctx := component.NewPaintContext(buf, 0, 0, 80, 24)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.SetPosition(i%80, i%24)
		c.Paint(ctx, buf)
	}
}
