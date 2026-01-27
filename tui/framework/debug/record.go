// Package debug provides debugging and recording utilities for TUI framework
package debug

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// Recorder 记录TUI运行时的状态变化
// 用于诊断光标闪烁、输入显示等问题
type Recorder struct {
	mu            sync.Mutex
	enabled       bool
	outputFile    *os.File
	buffer        *bytes.Buffer
	eventLog      []EventRecord
	renderLog     []RenderRecord
	stateSnapshot []StateRecord
	startTime     time.Time
}

// EventRecord 事件记录
type EventRecord struct {
	Time    time.Time
	Type    string
	Details string
}

// RenderRecord 渲染记录
type RenderRecord struct {
	Time     time.Time
	Width    int
	Height   int
	Content  string     // 前1000字符预览
	Cursor   CursorInfo // 光标位置信息
	Cells    int        // 总单元格数
	NonEmpty int        // 非空单元格数
}

// StateRecord 组件状态记录
type StateRecord struct {
	Time          time.Time
	ComponentID   string
	Value         string
	Cursor        int
	Focused       bool
	Visible       bool
	CursorVisible bool
}

// CursorInfo 光标信息
type CursorInfo struct {
	Row       int
	Col       int
	Char      rune
	Found     bool
	HasStyle  bool
	IsReverse bool
}

// NewRecorder 创建记录器
func NewRecorder(filename string) (*Recorder, error) {
	if filename == "" {
		filename = fmt.Sprintf("tui_debug_%s.log", time.Now().Format("20060102_150405"))
	}

	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return &Recorder{
		enabled:    true,
		outputFile: file,
		buffer:     &bytes.Buffer{},
		eventLog:   make([]EventRecord, 0),
		renderLog:  make([]RenderRecord, 0),
		startTime:  time.Now(),
	}, nil
}

// Enable 启用记录
func (r *Recorder) Enable() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enabled = true
	r.log("=== Recording Started ===")
}

// Disable 禁用记录
func (r *Recorder) Disable() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enabled = false
	r.log("=== Recording Paused ===")
}

// RecordEvent 记录事件
func (r *Recorder) RecordEvent(ev event.Event) {
	if !r.enabled {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	record := EventRecord{
		Time: time.Now(),
		Type: fmt.Sprintf("%d", ev.Type()),
	}

	if keyEv, ok := ev.(*event.KeyEvent); ok {
		record.Details = fmt.Sprintf("Key: '%c', Special: %v", keyEv.Key.Rune, keyEv.Special)
	}

	r.eventLog = append(r.eventLog, record)
	r.log("EVENT: %s - %s", record.Type, record.Details)
}

// RecordRender 记录渲染状态
func (r *Recorder) RecordRender(buf *paint.Buffer) {
	if !r.enabled || buf == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	// 分析缓冲区内容
	preview := r.captureBufferPreview(buf, 15) // 前15行预览
	cursor := r.findCursor(buf)
	nonEmpty := r.countNonEmptyCells(buf)

	record := RenderRecord{
		Time:     time.Now(),
		Width:    buf.Width,
		Height:   buf.Height,
		Content:  preview,
		Cursor:   cursor,
		Cells:    buf.Width * buf.Height,
		NonEmpty: nonEmpty,
	}

	r.renderLog = append(r.renderLog, record)

	r.log("RENDER: %dx%d, cursor at (%d,%d), %d/%d cells used",
		buf.Width, buf.Height, cursor.Row, cursor.Col, nonEmpty, record.Cells)

	// 如果找到光标，记录详细信息
	if cursor.Found {
		r.log("  CURSOR: row=%d col=%d char='%c' (0x%02x) reverse=%v",
			cursor.Row, cursor.Col, cursor.Char, cursor.Char, cursor.IsReverse)

		// 检查光标周围的内容
		if cursor.Row < buf.Height && cursor.Col+5 < buf.Width {
			surrounding := ""
			for dx := -2; dx <= 2; dx++ {
				xc := cursor.Col + dx
				if xc >= 0 && xc < buf.Width {
					ch := buf.Cells[cursor.Row][xc].Char
					if ch == 0 {
						ch = ' '
					}
					surrounding += string(ch)
				}
			}
			r.log("  CURSOR SURROUNDING: '[%s]' (cursor at center)", surrounding)
		}
	}

	// 记录缓冲区预览
	r.log("  BUFFER PREVIEW:\n%s", preview)
}

// RecordTextInputState 记录TextInput组件状态
func (r *Recorder) RecordTextInputState(id, value string, cursor int, focused, visible, cursorVisible bool) {
	if !r.enabled {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	record := StateRecord{
		Time:          time.Now(),
		ComponentID:   id,
		Value:         value,
		Cursor:        cursor,
		Focused:       focused,
		Visible:       visible,
		CursorVisible: cursorVisible,
	}

	r.stateSnapshot = append(r.stateSnapshot, record)

	r.log("STATE[%s]: value='%s' cursor=%d focused=%v visible=%v cursorVisible=%v",
		id, value, cursor, focused, visible, cursorVisible)
}

// captureBufferPreview 捕获缓冲区预览
func (r *Recorder) captureBufferPreview(buf *paint.Buffer, maxLines int) string {
	var preview bytes.Buffer
	lines := min(maxLines, buf.Height)

	for y := 0; y < lines; y++ {
		preview.WriteString(fmt.Sprintf("%2d: ", y))
		for x := 0; x < min(80, buf.Width); x++ {
			cell := buf.Cells[y][x]
			char := cell.Char
			if char == 0 {
				char = ' '
			}

			// 标记光标位置
			if cell.Style.IsReverse() {
				preview.WriteString(fmt.Sprintf("\033[7m%c\033[0m", char)) // 反白显示
			} else if cell.Style.IsBold() {
				preview.WriteString(fmt.Sprintf("\033[1m%c\033[0m", char)) // 粗体
			} else {
				preview.WriteRune(char)
			}
		}
		preview.WriteString("\n")
	}

	return preview.String()
}

// findCursor 在缓冲区中查找光标位置
func (r *Recorder) findCursor(buf *paint.Buffer) CursorInfo {
	cursor := CursorInfo{Found: false}

	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y][x]
			if cell.Style.IsReverse() {
				cursor.Found = true
				cursor.Row = y
				cursor.Col = x
				cursor.Char = cell.Char
				cursor.HasStyle = true
				cursor.IsReverse = true
				return cursor
			}
		}
	}

	return cursor
}

// countNonEmptyCells 统计非空单元格
func (r *Recorder) countNonEmptyCells(buf *paint.Buffer) int {
	count := 0
	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			if buf.Cells[y][x].Char != 0 {
				count++
			}
		}
	}
	return count
}

// DumpToFile 将所有记录写入文件
func (r *Recorder) DumpToFile() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.log("\n=== Recording Summary ===")
	r.log("Duration: %v", time.Since(r.startTime))
	r.log("Total Events: %d", len(r.eventLog))
	r.log("Total Renders: %d", len(r.renderLog))
	r.log("Total States: %d", len(r.stateSnapshot))

	// 写入缓冲区内容到文件
	if _, err := r.buffer.WriteTo(r.outputFile); err != nil {
		return err
	}

	return r.outputFile.Close()
}

// GetOutput 获取记录内容
func (r *Recorder) GetOutput() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.buffer.String()
}

// log 内部日志
func (r *Recorder) log(format string, args ...interface{}) {
	timestamp := time.Since(r.startTime).Milliseconds()
	msg := fmt.Sprintf("[%8dms] %s\n", timestamp, fmt.Sprintf(format, args...))
	r.buffer.WriteString(msg)
	if r.outputFile != nil {
		r.outputFile.WriteString(msg)
	}
}

// min 返回较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GlobalRecorder 全局记录器实例
var GlobalRecorder *Recorder

// InitGlobalRecorder 初始化全局记录器
func InitGlobalRecorder(filename string) error {
	recorder, err := NewRecorder(filename)
	if err != nil {
		return err
	}
	GlobalRecorder = recorder
	GlobalRecorder.Enable()
	return nil
}

// CloseGlobalRecorder 关闭全局记录器
func CloseGlobalRecorder() error {
	if GlobalRecorder != nil {
		return GlobalRecorder.DumpToFile()
	}
	return nil
}
