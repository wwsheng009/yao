//go:build unix || linux || darwin || (freebsd && !windows)
// +build unix linux darwin freebsd,!windows

package platform

import (
	"bytes"
	"os"
	"syscall"
	"time"
	"unicode/utf8"
	"unsafe"
)

type unixInputReader struct {
	events   chan<- RawInput
	quit     chan struct{}
	original *syscall.Termios
}

func newInputReaderImpl() inputReaderImpl {
	return newUnixInputReader()
}

func newUnixInputReader() inputReaderImpl {
	return &unixInputReader{
		quit: make(chan struct{}),
	}
}

func (r *unixInputReader) Start(events chan<- RawInput) error {
	r.events = events

	if err := r.enableRawMode(); err != nil {
		return err
	}

	go r.readLoop()

	return nil
}

func (r *unixInputReader) Stop() error {
	close(r.quit)

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
				continue
			}

			if n > 0 {
				input := r.parseSequence(buf[:n])
				if input.Type != InputKeyPress || input.Key != 0 || input.Special != KeyUnknown {
					select {
					case r.events <- input:
					case <-r.quit:
						return
					}
				}
			}
		}
	}
}

func (r *unixInputReader) parseSequence(buf []byte) RawInput {
	now := time.Now()

	if len(buf) == 0 {
		return RawInput{Timestamp: now}
	}

	b := buf[0]

	// 控制字符
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

	// 可打印字符
	if b >= 32 && b <= 126 {
		return RawInput{
			Type:      InputKeyPress,
			Key:       rune(b),
			Timestamp: now,
		}
	}

	// CSI 序列: ESC [
	if len(buf) > 1 && b == 0x1b && buf[1] == '[' {
		if len(buf) >= 3 {
			final := buf[len(buf)-1]
			switch final {
			case 'A':
				return RawInput{Type: InputKeyPress, Special: KeyUp, Timestamp: now}
			case 'B':
				return RawInput{Type: InputKeyPress, Special: KeyDown, Timestamp: now}
			case 'C':
				return RawInput{Type: InputKeyPress, Special: KeyRight, Timestamp: now}
			case 'D':
				return RawInput{Type: InputKeyPress, Special: KeyLeft, Timestamp: now}
			}
		}
	}

	return RawInput{Type: InputKeyPress, Timestamp: now}
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
