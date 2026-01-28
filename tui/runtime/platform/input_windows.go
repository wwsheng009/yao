//go:build windows
// +build windows

package platform

import (
	"sync"
	"syscall"
	"time"
	"unsafe"
)

type windowsInputReader struct {
	events       chan<- RawInput
	quit         chan struct{}
	quitOnce     sync.Once
	mu           sync.Mutex
	originalMode uint32
	lastWidth    int
	lastHeight   int
}

func newInputReaderImpl() inputReaderImpl {
	return newWindowsInputReader()
}

func newWindowsInputReader() inputReaderImpl {
	return &windowsInputReader{
		quit: make(chan struct{}),
	}
}

func (r *windowsInputReader) Start(events chan<- RawInput) error {
	r.mu.Lock()
	// Recreate quit channel if it was closed
	select {
	case <-r.quit:
		r.quit = make(chan struct{})
		r.quitOnce = sync.Once{}
	default:
		// Channel still open, reuse it
	}
	r.mu.Unlock()

	r.events = events

	handle, _, err := procGetStdHandle.Call(STD_INPUT_HANDLE)
	if handle == 0 {
		return err
	}

	r.originalMode = r.getConsoleMode(handle)

	// 启用鼠标输入和窗口输入
	mode := r.originalMode | ENABLE_WINDOW_INPUT | ENABLE_EXTENDED_FLAGS | ENABLE_MOUSE_INPUT
	mode &^= ENABLE_VIRTUAL_TERMINAL_INPUT
	r.setConsoleMode(handle, mode)

	// 初始化当前窗口大小
	r.updateWindowSize(handle)

	go r.readLoop(handle)

	return nil
}

func (r *windowsInputReader) Stop() error {
	r.quitOnce.Do(func() {
		close(r.quit)
	})

	handle, _, _ := procGetStdHandle.Call(STD_INPUT_HANDLE)
	if handle != 0 {
		r.setConsoleMode(handle, r.originalMode)
	}

	return nil
}

func (r *windowsInputReader) ReadEvent() (RawInput, error) {
	handle, _, err := procGetStdHandle.Call(STD_INPUT_HANDLE)
	if handle == 0 {
		return RawInput{}, err
	}

	record, err := r.readSingleRecord(handle)
	if err != nil {
		return RawInput{}, err
	}

	return r.parseRecord(record), nil
}

func (r *windowsInputReader) readLoop(handle uintptr) {
	pollTicker := time.NewTicker(250 * time.Millisecond)
	defer pollTicker.Stop()

	for {
		select {
		case <-r.quit:
			return

		case <-pollTicker.C:
			// 定期检查窗口大小变化（Windows resize 事件不可靠）
			r.updateWindowSize(handle)

		default:
			var count uint32
			procGetNumberOfConsoleInputEvents.Call(handle, uintptr(unsafe.Pointer(&count)))

			if count == 0 {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			record, err := r.readSingleRecord(handle)
			if err != nil {
				continue
			}

			input := r.parseRecord(record)
			// 发送所有有效事件（包括鼠标事件）
			if input.Type != 0 {
				select {
				case r.events <- input:
				case <-r.quit:
					return
				}
			}
		}
	}
}

func (r *windowsInputReader) readSingleRecord(handle uintptr) (*INPUT_RECORD, error) {
	var record INPUT_RECORD
	var count uint32

	ret, _, err := procReadConsoleInput.Call(
		handle,
		uintptr(unsafe.Pointer(&record)),
		1,
		uintptr(unsafe.Pointer(&count)),
	)

	if ret == 0 {
		return nil, err
	}

	return &record, nil
}

func (r *windowsInputReader) parseRecord(record *INPUT_RECORD) RawInput {
	now := time.Now()

	switch record.EventType {
	case KEY_EVENT:
		return r.parseKeyEvent(record, now)

	case MOUSE_EVENT:
		return r.parseMouseEvent(record, now)

	case WINDOW_BUFFER_SIZE_EVENT:
		// 获取控制台屏幕缓冲区信息
		handle, _, _ := procGetStdHandle.Call(STD_OUTPUT_HANDLE)
		if handle != 0 {
			var info CONSOLE_SCREEN_BUFFER_INFO
			procGetConsoleScreenBufferInfo.Call(handle, uintptr(unsafe.Pointer(&info)))

			// 使用 srWindow 获取实际可见窗口大小（而不是 dwSize 缓冲区大小）
			// srWindow.Right - srWindow.Left + 1 = 宽度
			// srWindow.Bottom - srWindow.Top + 1 = 高度
			width := int(info.srWindow.Right - info.srWindow.Left + 1)
			height := int(info.srWindow.Bottom - info.srWindow.Top + 1)

			return RawInput{
				Type:      InputResize,
				Timestamp: now,
				Width:     width,
				Height:    height,
			}
		}
		return RawInput{Timestamp: now}

	default:
		return RawInput{Timestamp: now}
	}
}

func (r *windowsInputReader) parseKeyEvent(record *INPUT_RECORD, now time.Time) RawInput {
	keyEvent := (*KEY_EVENT_RECORD)(unsafe.Pointer(&record.Event[0]))

	if keyEvent.KeyDown == 0 {
		return RawInput{Timestamp: now}
	}

	input := RawInput{
		Type:      InputKeyPress,
		Timestamp: now,
	}

	input.Special = SpecialKey(r.virtualKeyToSpecial(keyEvent.VirtualKeyCode))

	if keyEvent.ControlKeyState&0x0008 != 0 {
		input.Modifiers |= ModShift
	}
	if keyEvent.ControlKeyState&0x0004 != 0 {
		input.Modifiers |= ModCtrl
	}
	if keyEvent.ControlKeyState&0x0002 != 0 {
		input.Modifiers |= ModAlt
	}

	if input.Special == KeyUnknown && keyEvent.UChar > 0 {
		input.Key = rune(keyEvent.UChar)
	}

	return input
}

// parseMouseEvent 解析鼠标事件
func (r *windowsInputReader) parseMouseEvent(record *INPUT_RECORD, now time.Time) RawInput {
	mouseEvent := (*MOUSE_EVENT_RECORD)(unsafe.Pointer(&record.Event[0]))

	input := RawInput{
		Type:      InputMouse,
		Timestamp: now,
		MouseX:    int(mouseEvent.MousePosition.X),
		MouseY:    int(mouseEvent.MousePosition.Y),
	}

	// 确定鼠标按钮
	buttonState := mouseEvent.ButtonState
	if buttonState&FROM_LEFT_1ST_BUTTON_PRESSED != 0 {
		input.MouseButton = MouseLeft
	} else if buttonState&RIGHTMOST_BUTTON_PRESSED != 0 {
		input.MouseButton = MouseRight
	} else if buttonState&FROM_LEFT_2ND_BUTTON_PRESSED != 0 {
		input.MouseButton = MouseMiddle
	} else {
		input.MouseButton = MouseNone
	}

	// 确定鼠标动作类型
	eventFlags := mouseEvent.EventFlags
	if eventFlags&MOUSE_WHEELED != 0 {
		input.MouseAction = MouseWheelUp
	} else if eventFlags&MOUSE_HWHEELED != 0 {
		input.MouseAction = MouseWheelDown
	} else if eventFlags&DOUBLE_CLICK != 0 {
		// 双击事件 - 需要在上层处理
		input.MouseAction = MousePress
		// 设置双击标记
		input.Modifiers |= ModShift // 临时使用 Shift 位表示双击
	} else if eventFlags&MOUSE_MOVED != 0 {
		if buttonState != 0 {
			input.MouseAction = MousePress // 拖动
		} else {
			input.MouseAction = MouseMotion
		}
	} else {
		// 普通点击
		if buttonState != 0 {
			input.MouseAction = MousePress
		} else {
			input.MouseAction = MouseRelease
		}
	}

	return input
}

func (r *windowsInputReader) virtualKeyToSpecial(vk uint16) int {
	switch vk {
	case 0x08:
		return int(KeyBackspace)
	case 0x09:
		return int(KeyTab)
	case 0x0D:
		return int(KeyEnter)
	case 0x1B:
		return int(KeyEscape)
	case 0x21:
		return int(KeyPageUp)
	case 0x22:
		return int(KeyPageDown)
	case 0x23:
		return int(KeyEnd)
	case 0x24:
		return int(KeyHome)
	case 0x25:
		return int(KeyLeft)
	case 0x26:
		return int(KeyUp)
	case 0x27:
		return int(KeyRight)
	case 0x28:
		return int(KeyDown)
	case 0x2D:
		return int(KeyInsert)
	case 0x2E:
		return int(KeyDelete)
	case 0x70:
		return int(KeyF1)
	case 0x71:
		return int(KeyF2)
	case 0x72:
		return int(KeyF3)
	case 0x73:
		return int(KeyF4)
	case 0x74:
		return int(KeyF5)
	case 0x75:
		return int(KeyF6)
	case 0x76:
		return int(KeyF7)
	case 0x77:
		return int(KeyF8)
	case 0x78:
		return int(KeyF9)
	case 0x79:
		return int(KeyF10)
	case 0x7A:
		return int(KeyF11)
	case 0x7B:
		return int(KeyF12)
	default:
		return int(KeyUnknown)
	}
}

func (r *windowsInputReader) getConsoleMode(handle uintptr) uint32 {
	var mode uint32
	procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))
	return mode
}

func (r *windowsInputReader) setConsoleMode(handle uintptr, mode uint32) {
	procSetConsoleMode.Call(handle, uintptr(mode))
}

// updateWindowSize 检查并发送窗口大小变化事件
func (r *windowsInputReader) updateWindowSize(handle uintptr) {
	outHandle, _, _ := procGetStdHandle.Call(STD_OUTPUT_HANDLE)
	if outHandle == 0 {
		return
	}

	var info CONSOLE_SCREEN_BUFFER_INFO
	procGetConsoleScreenBufferInfo.Call(outHandle, uintptr(unsafe.Pointer(&info)))

	width := int(info.srWindow.Right - info.srWindow.Left + 1)
	height := int(info.srWindow.Bottom - info.srWindow.Top + 1)

	// 检查大小是否变化
	if width != r.lastWidth || height != r.lastHeight {
		r.lastWidth = width
		r.lastHeight = height

		// 发送大小变化事件
		select {
		case r.events <- RawInput{
			Type:      InputResize,
			Timestamp: time.Now(),
			Width:     width,
			Height:    height,
		}:
		case <-r.quit:
		}
	}
}

const (
	ENABLE_VIRTUAL_TERMINAL_INPUT = 0x0200
	ENABLE_WINDOW_INPUT            = 0x0008
	ENABLE_MOUSE_INPUT             = 0x0010
	ENABLE_EXTENDED_FLAGS          = 0x0080

	STD_INPUT_HANDLE  = ^uintptr(10 - 1) // -10 as unsigned
	STD_OUTPUT_HANDLE = ^uintptr(11 - 1) // -11 as unsigned

	KEY_EVENT                  = 0x0001
	MOUSE_EVENT                = 0x0002
	WINDOW_BUFFER_SIZE_EVENT   = 0x0004
)

type INPUT_RECORD struct {
	EventType uint16
	Padding   uint16
	Event     [16]byte
}

type KEY_EVENT_RECORD struct {
	KeyDown         int32
	RepeatCount     uint16
	VirtualKeyCode  uint16
	VirtualScanCode uint16
	UChar           uint16
	ControlKeyState uint32
}

// MOUSE_EVENT_RECORD 鼠标事件记录
type MOUSE_EVENT_RECORD struct {
	MousePosition    COORD
	ButtonState      uint32
	ControlKeyState  uint32
	EventFlags       uint32
}

// 鼠标按钮状态掩码
const (
	FROM_LEFT_1ST_BUTTON_PRESSED = 0x0001
	RIGHTMOST_BUTTON_PRESSED     = 0x0002
	FROM_LEFT_2ND_BUTTON_PRESSED = 0x0004
	FROM_LEFT_3RD_BUTTON_PRESSED = 0x0008
	FROM_LEFT_4TH_BUTTON_PRESSED = 0x0010
)

// 鼠标事件标志
const (
	DOUBLE_CLICK = 0x0002
	MOUSE_MOVED   = 0x0001
	MOUSE_WHEELED = 0x0004
	MOUSE_HWHEELED = 0x0008
)

// CONSOLE_SCREEN_BUFFER_INFO 控制台屏幕缓冲区信息
type CONSOLE_SCREEN_BUFFER_INFO struct {
	dwSize              COORD
	dwCursorPosition   COORD
	wAttributes        uint16
	srWindow           SMALL_RECT
	dwMaximumWindowSize COORD
}

type COORD struct {
	X int16
	Y int16
}

type SMALL_RECT struct {
	Left   int16
	Top    int16
	Right  int16
	Bottom int16
}

var (
	kernel32                         = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode               = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode               = kernel32.NewProc("SetConsoleMode")
	procGetStdHandle                 = kernel32.NewProc("GetStdHandle")
	procReadConsoleInput             = kernel32.NewProc("ReadConsoleInputW")
	procGetNumberOfConsoleInputEvents = kernel32.NewProc("GetNumberOfConsoleInputEvents")
	procGetConsoleScreenBufferInfo  = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

// restoreTerminalImpl Windows 终端恢复实现
func restoreTerminalImpl() {
	handle, _, _ := procGetStdHandle.Call(STD_INPUT_HANDLE)
	if handle != 0 {
		// 恢复到默认控制台模式
		defaultMode := uint32(ENABLE_EXTENDED_FLAGS | ENABLE_WINDOW_INPUT)
		procSetConsoleMode.Call(handle, uintptr(defaultMode))
	}

	// 同时恢复输出模式
	outHandle, _, _ := procGetStdHandle.Call(STD_OUTPUT_HANDLE)
	if outHandle != 0 {
		// 启用虚拟终端处理
		procSetConsoleMode.Call(outHandle, uintptr(ENABLE_VIRTUAL_TERMINAL_PROCESSING))
	}
}

const (
	ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004
)
