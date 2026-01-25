package component

import (
	"github.com/yaoapp/yao/tui/framework/event"
)

// Event 事件类型别名
type Event = event.Event

// EventHandler 事件处理器类型别名
type EventHandler = event.EventHandler

// Rect 矩形区域
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

// Contains 检查点是否在矩形内
func (r *Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.Width &&
		y >= r.Y && y < r.Y+r.Height
}

// Intersect 计算两个矩形的交集
func (r *Rect) Intersect(other *Rect) *Rect {
	if r == nil {
		return other
	}
	if other == nil {
		return r
	}

	x1 := maxInt(r.X, other.X)
	y1 := maxInt(r.Y, other.Y)
	x2 := minInt(r.X+r.Width, other.X+other.Width)
	y2 := minInt(r.Y+r.Height, other.Y+other.Height)

	if x1 >= x2 || y1 >= y2 {
		return nil
	}

	return &Rect{
		X:      x1,
		Y:      y1,
		Width:  x2 - x1,
		Height: y2 - y1,
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
