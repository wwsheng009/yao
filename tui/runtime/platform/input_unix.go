//go:build unix || linux || darwin || (freebsd && !windows)
// +build unix linux darwin freebsd,!windows

package platform

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

type unixInputReader struct {
	events     chan<- RawInput
	quit       chan struct{}
	original   *syscall.Termios
	parseBuffer []byte // 解析缓冲区，用于多字节序列
}

func newInputReaderImpl() inputReaderImpl {
	return newUnixInputReader()
}

func newUnixInputReader() inputReaderImpl {
	return &unixInputReader{
		quit:       make(chan struct{}),
		parseBuffer: make([]byte, 0, 64),
	}
}

func (r *unixInputReader) Start(events chan<- RawInput) error {
	r.events = events

	if err := r.enableRawMode(); err != nil {
		return err
	}

	// 启用鼠标跟踪 (SGR 格式，支持所有鼠标事件)
	r.enableMouse()

	go r.readLoop()

	return nil
}

func (r *unixInputReader) Stop() error {
	close(r.quit)

	// 禁用鼠标跟踪
	r.disableMouse()

	if r.original != nil {
		r.restoreMode()
	}

	return nil
}

func (r *unixInputReader) ReadEvent() (RawInput, error) {
	buf := make([]byte, 1)
	_, err := os.Stdin.Read(buf)
	if err != nil {
		return RawInput{}, err
	}

	return r.parseInput(buf), nil
}

func (r *unixInputReader) readLoop() {
	buf := make([]byte, 1, 128)

	for {
		select {
		case <-r.quit:
			return
		default:
			os.Stdin.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			n, err := os.Stdin.Read(buf)
			if err != nil {
				// 继续尝试，超时是正常的
				continue
			}

			if n > 0 {
				// 将新数据追加到解析缓冲区
				r.parseBuffer = append(r.parseBuffer, buf[:n]...)

				// 尝试解析缓冲区中的数据
				for len(r.parseBuffer) > 0 {
					input, remaining := r.parseSequence(r.parseBuffer)
					// 发送所有有效事件（InputKeyPress=0 是有效值！）
					if input.Type >= 0 && input.Type <= InputSignal {
						select {
						case r.events <- input:
						case <-r.quit:
							return
						}
					}

					if remaining == nil {
						// 需要更多数据
						break
					}

					r.parseBuffer = remaining
				}
			}
		}
	}
}

func (r *unixInputReader) parseSequence(buf []byte) (RawInput, []byte) {
	now := time.Now()

	if len(buf) == 0 {
		return RawInput{Timestamp: now}, nil
	}

	b := buf[0]

	// 检测鼠标事件: ESC [ < M (SGR 格式) 或 ESC [ M (X10 格式)
	if b == 0x1b && len(buf) > 2 && buf[1] == '[' {
		if buf[2] == '<' {
			// SGR 格式鼠标事件: ESC [ < Cb ; Cx ; Cy M
			return r.parseSGRMouseEvent(buf, now)
		} else if buf[2] == 'M' {
			// X10 格式鼠标事件: ESC [ M Cb Cx Cy
			return r.parseX10MouseEvent(buf, now)
		}
	}

	// 控制字符
	switch b {
	case '\r', '\n':
		return RawInput{Type: InputKeyPress, Special: KeyEnter, Timestamp: now}, buf[1:]
	case '\t':
		return RawInput{Type: InputKeyPress, Special: KeyTab, Timestamp: now}, buf[1:]
	case 0x7f, 0x08:
		return RawInput{Type: InputKeyPress, Special: KeyBackspace, Timestamp: now}, buf[1:]
	case 0x1b:
		// ESC 序列，检查是否是 CSI 序列
		if len(buf) > 1 && buf[1] == '[' {
			if len(buf) >= 3 {
				final := buf[len(buf)-1]
				switch final {
				case 'A', 'B', 'C', 'D', 'E', 'F', 'H', '~':
					// 完整的 CSI 序列
					return r.parseCSISquence(buf, now), nil
				}
			}
			// 不完整的 CSI 序列，需要更多数据
			return RawInput{}, nil
		}
		// 单独的 ESC 键
		return RawInput{Type: InputKeyPress, Special: KeyEscape, Timestamp: now}, buf[1:]
	}

	// 可打印字符
	if b >= 32 && b <= 126 {
		return RawInput{
			Type:      InputKeyPress,
			Key:       rune(b),
			Timestamp: now,
		}, buf[1:]
	}

	// 未知字符，跳过
	return RawInput{Timestamp: now}, buf[1:]
}

func (r *unixInputReader) parseInput(buf []byte) RawInput {
	now := time.Now()
	if len(buf) == 0 {
		return RawInput{Timestamp: now}
	}

	b := buf[0]

	switch b {
	case '\r', '\n':
		return RawInput{Type: InputKeyPress, Special: KeyEnter, Timestamp: now}
	case '\t':
		return RawInput{Type: InputKeyPress, Special: KeyTab, Timestamp: now}
	case 0x7f, 0x08:
		return RawInput{Type: InputKeyPress, Special: KeyBackspace, Timestamp: now}
	case 0x1b:
		return RawInput{Type: InputKeyPress, Special: KeyEscape, Timestamp: now}
	}

	if b >= 32 && b <= 126 {
		return RawInput{
			Type:      InputKeyPress,
			Key:       rune(b),
			Timestamp: now,
		}
	}

	return RawInput{Type: InputKeyPress, Timestamp: now}
}

func (r *unixInputReader) enableRawMode() error {
	fd := int(os.Stdin.Fd())

	var termios syscall.Termios
	if err := ioctl(fd, syscall.TCGETS, uintptr(unsafe.Pointer(&termios))); err != nil {
		return err
	}

	r.original = &termios

	termios.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP |
		syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	termios.Oflag &^= syscall.OPOST
	termios.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	termios.Cflag &^= syscall.CSIZE | syscall.PARENB
	termios.Cflag |= syscall.CS8
	termios.Cc[syscall.VMIN] = 1
	termios.Cc[syscall.VTIME] = 0

	return ioctl(fd, syscall.TCSETS, uintptr(unsafe.Pointer(&termios)))
}

func (r *unixInputReader) restoreMode() error {
	if r.original == nil {
		return nil
	}

	fd := int(os.Stdin.Fd())
	return ioctl(fd, syscall.TCSETS, uintptr(unsafe.Pointer(r.original)))
}

func ioctl(fd, cmd, arg uintptr) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, arg)
	if err != 0 {
		return err
	}
	return nil
}

// restoreTerminalImpl Unix 终端恢复实现
func restoreTerminalImpl() {
	fd := int(os.Stdin.Fd())

	var termios syscall.Termios
	if err := ioctl(fd, syscall.TCGETS, uintptr(unsafe.Pointer(&termios))); err != nil {
		return
	}

	// 恢复到规范模式
	termios.Lflag |= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	termios.Iflag |= syscall.ICRNL | syscall.IXON
	termios.Oflag |= syscall.OPOST

	ioctl(fd, syscall.TCSETS, uintptr(unsafe.Pointer(&termios)))
}

// ==============================================================================
// 鼠标事件支持
// ==============================================================================

// enableMouse 启用鼠标跟踪 (SGR 格式)
func (r *unixInputReader) enableMouse() {
	// SGR 1006: 启用扩展的鼠标报告格式，支持所有事件
	// \x1b[?1000h - 启用鼠标跟踪 (按键和释放)
	// \x1b[?1002h - 启用按钮事件跟踪
	// \x1b[?1003h - 启用所有鼠标事件跟踪
	// \x1b[?1006h - 启用 SGR 扩展模式
	os.Stdout.WriteString("\x1b[?1000h")
	os.Stdout.WriteString("\x1b[?1002h")
	os.Stdout.WriteString("\x1b[?1003h")
	os.Stdout.WriteString("\x1b[?1006h")
}

// disableMouse 禁用鼠标跟踪
func (r *unixInputReader) disableMouse() {
	os.Stdout.WriteString("\x1b[?1003l")
	os.Stdout.WriteString("\x1b[?1002l")
	os.Stdout.WriteString("\x1b[?1000l")
	os.Stdout.WriteString("\x1b[?1006l")
}

// parseSGRMouseEvent 解析 SGR 格式鼠标事件
// 格式: ESC [ < Cb ; Cx ; Cy M(m)
// Cb: 按钮码 (bitmask)
// Cx, Cy: 坐标 (1-based)
// M: 按下, m: 释放
func (r *unixInputReader) parseSGRMouseEvent(buf []byte, now time.Time) (RawInput, []byte) {
	// 查找序列结束位置 (M 或 m)
	end := -1
	for i := 3; i < len(buf); i++ {
		if buf[i] == 'M' || buf[i] == 'm' {
			end = i
			break
		}
	}

	if end == -1 {
		// 序列不完整，需要更多数据
		return RawInput{}, nil
	}

	// 提取参数部分: ESC [ < ... M/m
	params := string(buf[3:end])
	// 分离三个参数: Cb;Cx;Cy
	var cb, cx, cy int
	if _, err := fmt.Sscanf(params, "%d;%d;%d", &cb, &cx, &cy); err != nil {
		return RawInput{Timestamp: now}, buf[end+1:]
	}

	input := RawInput{
		Type:      InputMouse,
		Timestamp: now,
		MouseX:    cx - 1, // 转换为 0-based
		MouseY:    cy - 1,
	}

	// 解析按钮码 (bitmask)
	// bit 0: 左键
	// bit 1: 中键
	// bit 2: 右键
	// bit 3-4: 释放时移动
	// bit 5-6: 滚轮
	buttonCode := cb & 0x43
	release := buf[end] == 'm'

	switch buttonCode {
	case 0:
		input.MouseButton = MouseLeft
	case 1:
		input.MouseButton = MouseMiddle
	case 2:
		input.MouseButton = MouseRight
	case 32:
		input.MouseButton = MouseNone
		input.MouseAction = MouseWheelUp
	case 33:
		input.MouseButton = MouseNone
		input.MouseAction = MouseWheelDown
	default:
		input.MouseButton = MouseNone
	}

	// 确定动作类型
	switch {
	case buttonCode == 32 || buttonCode == 33:
		// 滚轮，已在上面设置
	case buttonCode >= 64:
		// 鼠标移动 (带按钮状态)
		input.MouseAction = MouseMotion
		// 更新按钮状态
		if buttonCode&0x01 != 0 {
			input.MouseButton = MouseLeft
		} else if buttonCode&0x02 != 0 {
			input.MouseButton = MouseMiddle
		} else if buttonCode&0x04 != 0 {
			input.MouseButton = MouseRight
		}
	case release:
		input.MouseAction = MouseRelease
	default:
		input.MouseAction = MousePress
	}

	return input, buf[end+1:]
}

// parseX10MouseEvent 解析 X10 格式鼠标事件
// 格式: ESC [ M Cb Cx Cy
// Cb: 按钮码 + 32
// Cx, Cy: 坐标 + 32
func (r *unixInputReader) parseX10MouseEvent(buf []byte, now time.Time) (RawInput, []byte) {
	// X10 格式固定长度: ESC [ M Cb Cx Cy = 6 字节
	if len(buf) < 6 {
		return RawInput{}, nil
	}

	cb := int(buf[3] - 32)
	cx := int(buf[4] - 33) // 转换为 0-based
	cy := int(buf[5] - 33)

	input := RawInput{
		Type:      InputMouse,
		Timestamp: now,
		MouseX:    cx,
		MouseY:    cy,
	}

	// 解析按钮
	button := cb & 0x03
	modifiers := cb & 0x1c

	switch button {
	case 0:
		input.MouseButton = MouseLeft
	case 1:
		input.MouseButton = MouseMiddle
	case 2:
		input.MouseButton = MouseRight
	case 3:
		input.MouseButton = MouseNone // 释放
	}

	// 检查是否是滚轮事件 (modifiers = 0x40 或 0x41)
	if modifiers == 0x40 {
		input.MouseButton = MouseNone
		input.MouseAction = MouseWheelUp
	} else if modifiers == 0x41 {
		input.MouseButton = MouseNone
		input.MouseAction = MouseWheelDown
	} else {
		if button == 3 || (cb&0x40) != 0 {
			input.MouseAction = MouseRelease
		} else {
			input.MouseAction = MousePress
		}
	}

	return input, buf[6:]
}

// parseCSISquence 解析 CSI 序列 (光标键等)
func (r *unixInputReader) parseCSISquence(buf []byte, now time.Time) RawInput {
	// 查找序列结束字符
	end := -1
	for i := 2; i < len(buf); i++ {
		c := buf[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z' || c == '~') {
			end = i
			break
		}
	}

	if end == -1 {
		return RawInput{Timestamp: now}
	}

	final := buf[end]
	switch final {
	case 'A':
		return RawInput{Type: InputKeyPress, Special: KeyUp, Timestamp: now}
	case 'B':
		return RawInput{Type: InputKeyPress, Special: KeyDown, Timestamp: now}
	case 'C':
		return RawInput{Type: InputKeyPress, Special: KeyRight, Timestamp: now}
	case 'D':
		return RawInput{Type: InputKeyPress, Special: KeyLeft, Timestamp: now}
	case 'F':
		return RawInput{Type: InputKeyPress, Special: KeyEnd, Timestamp: now}
	case 'H':
		return RawInput{Type: InputKeyPress, Special: KeyHome, Timestamp: now}
	case '~':
		// 解析参数
		params := string(buf[2:end])
		switch params {
		case "1", "7":
			return RawInput{Type: InputKeyPress, Special: KeyHome, Timestamp: now}
		case "2":
			return RawInput{Type: InputKeyPress, Special: KeyInsert, Timestamp: now}
		case "3":
			return RawInput{Type: InputKeyPress, Special: KeyDelete, Timestamp: now}
		case "4":
			return RawInput{Type: InputKeyPress, Special: KeyEnd, Timestamp: now}
		case "5":
			return RawInput{Type: InputKeyPress, Special: KeyPageUp, Timestamp: now}
		case "6":
			return RawInput{Type: InputKeyPress, Special: KeyPageDown, Timestamp: now}
		case "15":
			return RawInput{Type: InputKeyPress, Special: KeyF5, Timestamp: now}
		case "17":
			return RawInput{Type: InputKeyPress, Special: KeyF6, Timestamp: now}
		case "18":
			return RawInput{Type: InputKeyPress, Special: KeyF7, Timestamp: now}
		case "19":
			return RawInput{Type: InputKeyPress, Special: KeyF8, Timestamp: now}
		case "20":
			return RawInput{Type: InputKeyPress, Special: KeyF9, Timestamp: now}
		case "21":
			return RawInput{Type: InputKeyPress, Special: KeyF10, Timestamp: now}
		case "23":
			return RawInput{Type: InputKeyPress, Special: KeyF11, Timestamp: now}
		case "24":
			return RawInput{Type: InputKeyPress, Special: KeyF12, Timestamp: now}
		}
	}

	return RawInput{Timestamp: now}
}
