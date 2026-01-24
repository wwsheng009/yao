package screen

import (
	"github.com/yaoapp/yao/tui/framework/style"
)

// Manager 屏幕管理器
type Manager struct {
	// 终端操作通过低级系统调用实现
	width  int
	height int

	// 双缓冲
	front *Buffer
	back  *Buffer

	// 光标
	cursor       Cursor
	cursorVisible bool
}

// NewManager 创建屏幕管理器
func NewManager(width, height int) *Manager {
	return &Manager{
		width:  width,
		height: height,
		cursor: Cursor{},
	}
}

// Init 初始化屏幕
func (m *Manager) Init() error {
	m.front = NewBuffer(m.width, m.height)
	m.back = NewBuffer(m.width, m.height)

	// 进入备用屏幕
	print("\x1b[?1049h")

	// 启用原始模式
	// TODO: 实现 Unix 原始模式

	// 隐藏光标
	m.hideCursor()

	return nil
}

// Close 关闭屏幕
func (m *Manager) Close() error {
	// 显示光标
	m.showCursor()

	// 退出原始模式
	// TODO: 实现 Unix 原始模式退出

	// 退出备用屏幕
	print("\x1b[?1049l")

	return nil
}

// Render 渲染缓冲区
func (m *Manager) Render(buf *Buffer) error {
	// 计算差异
	diff := m.diff(m.front, buf)

	// 输出变更
	m.drawChanges(diff)

	// 更新前缓冲
	m.front = buf

	return nil
}

// diff 计算缓冲区差异
func (m *Manager) diff(old, new *Buffer) []Change {
	var changes []Change

	minW := minInt(old.width, new.width)
	minH := minInt(old.height, new.height)

	for y := 0; y < minH; y++ {
		for x := 0; x < minW; x++ {
			oldCell := old.cells[y][x]
			newCell := new.cells[y][x]

			if oldCell.Char != newCell.Char || oldCell.Style != newCell.Style {
				changes = append(changes, Change{
					X:     x,
					Y:     y,
					Char:  newCell.Char,
					Style: newCell.Style,
				})
			}
		}
	}

	return changes
}

// drawChanges 绘制变更
func (m *Manager) drawChanges(changes []Change) {
	// TODO: 优化输出，批量处理相同样式
	for _, c := range changes {
		m.moveCursor(c.X, c.Y)
		ansi := c.Style.ToANSI()
		if ansi != "" {
			print(ansi)
		}
		if c.Char == 0 {
			print(" ")
		} else {
			print(string(c.Char))
		}
	}
	// 重置样式
	print("\x1b[0m")
}

// GetSize 获取屏幕尺寸
func (m *Manager) GetSize() (width, height int) {
	return m.width, m.height
}

// SetSize 设置屏幕尺寸
func (m *Manager) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.back = NewBuffer(width, height)
}

// SetCursor 设置光标位置
func (m *Manager) SetCursor(x, y int) {
	m.cursor.X = x
	m.cursor.Y = y
	if m.cursorVisible {
		m.moveCursor(x, y)
	}
}

// GetCursor 获取光标位置
func (m *Manager) GetCursor() (x, y int) {
	return m.cursor.X, m.cursor.Y
}

// SetCursorVisible 设置光标可见性
func (m *Manager) SetCursorVisible(visible bool) {
	if visible != m.cursorVisible {
		if visible {
			m.showCursor()
		} else {
			m.hideCursor()
		}
		m.cursorVisible = visible
	}
}

// hideCursor 隐藏光标
func (m *Manager) hideCursor() {
	print("\x1b[?25l")
}

// showCursor 显示光标
func (m *Manager) showCursor() {
	print("\x1b[?25h")
}

// moveCursor 移动光标
func (m *Manager) moveCursor(x, y int) {
	print("\x1b[")
	print(itoa(y+1))
	print(";")
	print(itoa(x+1))
	print("H")
}

// Cursor 光标结构
type Cursor struct {
	X int
	Y int
}

// Change 缓冲区变更
type Change struct {
	X     int
	Y     int
	Char  rune
	Style style.Style
}

// Buffer 渲染缓冲区
type Buffer struct {
	width  int
	height int
	cells  [][]Cell
}

// Cell 缓冲区单元格
type Cell struct {
	Char  rune
	Style style.Style
}

// NewBuffer 创建缓冲区
func NewBuffer(width, height int) *Buffer {
	b := &Buffer{
		width:  width,
		height: height,
		cells:  make([][]Cell, height),
	}

	for y := 0; y < height; y++ {
		b.cells[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			b.cells[y][x] = Cell{Char: ' '}
		}
	}

	return b
}

// GetSize 获取缓冲区尺寸
func (b *Buffer) GetSize() (width, height int) {
	return b.width, b.height
}

// SetCell 设置单元格
func (b *Buffer) SetCell(x, y int, char rune, s style.Style) {
	if x >= 0 && x < b.width && y >= 0 && y < b.height {
		b.cells[y][x] = Cell{Char: char, Style: s}
	}
}

// SetLine 设置一行
func (b *Buffer) SetLine(y int, text string, s style.Style) {
	if y < 0 || y >= b.height {
		return
	}

	runes := []rune(text)
	for x, r := range runes {
		if x < b.width {
			b.cells[y][x] = Cell{Char: r, Style: s}
		}
	}
}

// Fill 填充区域
func (b *Buffer) Fill(x, y, width, height int, char rune, s style.Style) {
	for py := y; py < y+height && py < b.height; py++ {
		for px := x; px < x+width && px < b.width; px++ {
			b.cells[py][px] = Cell{Char: char, Style: s}
		}
	}
}

// Clear 清空缓冲区
func (b *Buffer) Clear() {
	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			b.cells[y][x] = Cell{Char: ' '}
		}
	}
}

// GetCell 获取单元格
func (b *Buffer) GetCell(x, y int) Cell {
	if x >= 0 && x < b.width && y >= 0 && y < b.height {
		return b.cells[y][x]
	}
	return Cell{}
}

// String 返回缓冲区的字符串表示
func (b *Buffer) String() string {
	var lines []string
	for y := 0; y < b.height; y++ {
		var line string
		currentStyle := style.Style{}
		for x := 0; x < b.width; x++ {
			cell := b.cells[y][x]
			if cell.Style != currentStyle {
				if currentStyle != (style.Style{}) {
					line += "\x1b[0m"
				}
				line += cell.Style.ToANSI()
				currentStyle = cell.Style
			}
			if cell.Char == 0 {
				line += " "
			} else {
				line += string(cell.Char)
			}
		}
		if currentStyle != (style.Style{}) {
			line += "\x1b[0m"
		}
		lines = append(lines, line)
	}
	return joinLines(lines)
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\r\n"
		}
		result += line
	}
	return result
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 11)
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
