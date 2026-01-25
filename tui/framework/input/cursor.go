package input

import (
	"sync"
	"time"
)

// ============================================================================
// 全局光标闪烁管理器
// ============================================================================

var (
	cursorMutex     sync.RWMutex
	focusedInputs   []*TextInput
	lastBlinkCheck  time.Time
	dirtyCallback   func()
)

// RegisterCursor 注册光标到全局管理器
func RegisterCursor(txt *TextInput) {
	cursorMutex.Lock()
	defer cursorMutex.Unlock()

	// 检查是否已注册
	for _, existing := range focusedInputs {
		if existing == txt {
			return
		}
	}
	focusedInputs = append(focusedInputs, txt)
}

// UnregisterCursor 从全局管理器注销光标
func UnregisterCursor(txt *TextInput) {
	cursorMutex.Lock()
	defer cursorMutex.Unlock()

	for i, existing := range focusedInputs {
		if existing == txt {
			focusedInputs = append(focusedInputs[:i], focusedInputs[i+1:]...)
			return
		}
	}
}

// SetDirtyCallback 设置脏标记回调
func SetDirtyCallback(fn func()) {
	cursorMutex.Lock()
	defer cursorMutex.Unlock()
	dirtyCallback = fn
}

// CursorBlinkTick 更新所有光标的闪烁状态
// 返回是否需要重新渲染
func CursorBlinkTick() bool {
	cursorMutex.RLock()
	inputs := make([]*TextInput, len(focusedInputs))
	copy(inputs, focusedInputs)
	callback := dirtyCallback
	cursorMutex.RUnlock()

	now := time.Now()
	needsRedraw := false

	// 每500ms更新一次光标状态
	if now.Sub(lastBlinkCheck) >= 500*time.Millisecond {
		for _, txt := range inputs {
			if txt.UpdateCursorBlink() {
				needsRedraw = true
			}
		}
		lastBlinkCheck = now

		// 只在光标状态改变时触发回调
		if needsRedraw && callback != nil {
			callback()
		}
	}

	return needsRedraw
}
