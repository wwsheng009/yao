package input

import (
	"github.com/yaoapp/yao/tui/framework/platform"
	"github.com/yaoapp/yao/tui/runtime/action"
)

// ==============================================================================
// KeyMap 输入转换系统 (V3)
// ==============================================================================
// KeyMap 负责将 Platform 的 RawInput 转换为语义化的 Action
// 这是 Runtime 的职责，Framework 和 Component 不应该直接处理 RawInput

// KeyMap 按键映射
type KeyMap struct {
	// 默认按键映射
	defaultMap map[platform.SpecialKey]action.ActionType

	// 上下文相关映射
	contextMaps map[string]map[platform.SpecialKey]action.ActionType

	// 当前上下文栈
	contextStack []string

	// 自定义绑定
	customBinds map[string]action.ActionType
}

// NewKeyMap 创建按键映射
func NewKeyMap() *KeyMap {
	km := &KeyMap{
		defaultMap:  make(map[platform.SpecialKey]action.ActionType),
		contextMaps: make(map[string]map[platform.SpecialKey]action.ActionType),
		contextStack: make([]string, 0),
		customBinds: make(map[string]action.ActionType),
	}

	// 初始化默认映射
	km.initDefaults()

	return km
}

// initDefaults 初始化默认按键映射
func (k *KeyMap) initDefaults() {
	// 导航
	k.defaultMap[platform.KeyUp] = action.ActionNavigateUp
	k.defaultMap[platform.KeyDown] = action.ActionNavigateDown
	k.defaultMap[platform.KeyLeft] = action.ActionNavigateLeft
	k.defaultMap[platform.KeyRight] = action.ActionNavigateRight
	k.defaultMap[platform.KeyHome] = action.ActionCursorHome
	k.defaultMap[platform.KeyEnd] = action.ActionCursorEnd
	k.defaultMap[platform.KeyPageUp] = action.ActionNavigatePageUp
	k.defaultMap[platform.KeyPageDown] = action.ActionNavigatePageDown

	// Vim 风格导航
	k.defaultMap[platform.KeyK] = action.ActionNavigateUp
	k.defaultMap[platform.KeyJ] = action.ActionNavigateDown
	k.defaultMap[platform.KeyH] = action.ActionNavigateLeft
	k.defaultMap[platform.KeyL] = action.ActionNavigateRight

	// Tab 导航
	k.defaultMap[platform.KeyTab] = action.ActionNavigateNext

	// 编辑
	k.defaultMap[platform.KeyBackspace] = action.ActionBackspace
	k.defaultMap[platform.KeyDelete] = action.ActionDeleteChar
	k.defaultMap[platform.KeyEnter] = action.ActionSubmit
	k.defaultMap[platform.KeyEscape] = action.ActionCancel

	// 功能键 F1-F12 可以映射到特定功能
	k.defaultMap[platform.KeyF1] = action.ActionHelp
	k.defaultMap[platform.KeyF5] = action.ActionRefresh
}

// Map 将 RawInput 转换为 Action
// 返回 nil 表示不需要处理此输入
func (k *KeyMap) Map(input platform.RawInput) *action.Action {
	// 只处理键盘按键
	if input.Type != platform.InputKeyPress {
		return nil
	}

	// 优先检查自定义绑定
	if key := k.makeKey(input); key != "" {
		if typ, ok := k.customBinds[key]; ok {
			return action.NewAction(typ).
				WithPayload(input)
		}
	}

	// 检查上下文相关映射
	if len(k.contextStack) > 0 {
		currentContext := k.contextStack[len(k.contextStack)-1]
		if ctxMap, ok := k.contextMaps[currentContext]; ok {
			if typ, ok := ctxMap[input.Special]; ok {
				return action.NewAction(typ).
					WithPayload(input)
			}
		}
	}

	// 使用默认映射
	if input.Special != platform.KeyUnknown {
		if typ, ok := k.defaultMap[input.Special]; ok {
			return action.NewAction(typ).
				WithPayload(input)
		}
	}

	// 普通字符输入
	if input.Key != 0 {
		return action.NewAction(action.ActionInputChar).
			WithPayload(input.Key)
	}

	return nil
}

// makeKey 构建映射键
func (k *KeyMap) makeKey(input platform.RawInput) string {
	if input.Special != platform.KeyUnknown {
		// 特殊键 + 修饰键
		modStr := ""
		if input.Modifiers&platform.ModCtrl != 0 {
			modStr += "C-"
		}
		if input.Modifiers&platform.ModAlt != 0 {
			modStr += "A-"
		}
		if input.Modifiers&platform.ModShift != 0 {
			modStr += "S-"
		}
		return modStr + specialKeyToString(input.Special)
	}

	// 普通键 + 修饰键
	if input.Key != 0 {
		modStr := ""
		if input.Modifiers&platform.ModCtrl != 0 {
			modStr += "C-"
		}
		if input.Modifiers&platform.ModAlt != 0 {
			modStr += "A-"
		}
		if input.Modifiers&platform.ModShift != 0 {
			modStr += "S-"
		}
		return modStr + string(input.Key)
	}

	return ""
}

// Bind 绑定自定义按键
// combo 格式: "C-c" (Ctrl+C), "A-x" (Alt+X), "C-S-a" (Ctrl+Shift+A)
func (k *KeyMap) Bind(combo string, actionType action.ActionType) {
	k.customBinds[combo] = actionType
}

// Unbind 解除绑定
func (k *KeyMap) Unbind(combo string) {
	delete(k.customBinds, combo)
}

// BindContext 绑定上下文相关映射
func (k *KeyMap) BindContext(context string, bindings map[platform.SpecialKey]action.ActionType) {
	k.contextMaps[context] = bindings
}

// PushContext 推入上下文
func (k *KeyMap) PushContext(context string) {
	k.contextStack = append(k.contextStack, context)
}

// PopContext 弹出上下文
func (k *KeyMap) PopContext() {
	if len(k.contextStack) > 0 {
		k.contextStack = k.contextStack[:len(k.contextStack)-1]
	}
}

// ClearContext 清空上下文栈
func (k *KeyMap) ClearContext() {
	k.contextStack = k.contextStack[:0]
}

// specialKeyToString 特殊键转字符串
func specialKeyToString(key platform.SpecialKey) string {
	switch key {
	case platform.KeyEscape:
		return "Esc"
	case platform.KeyEnter:
		return "Enter"
	case platform.KeyTab:
		return "Tab"
	case platform.KeyBackspace:
		return "Backspace"
	case platform.KeyDelete:
		return "Delete"
	case platform.KeyInsert:
		return "Insert"
	case platform.KeyUp:
		return "Up"
	case platform.KeyDown:
		return "Down"
	case platform.KeyLeft:
		return "Left"
	case platform.KeyRight:
		return "Right"
	case platform.KeyHome:
		return "Home"
	case platform.KeyEnd:
		return "End"
	case platform.KeyPageUp:
		return "PageUp"
	case platform.KeyPageDown:
		return "PageDown"
	case platform.KeyF1:
		return "F1"
	case platform.KeyF2:
		return "F2"
	case platform.KeyF3:
		return "F3"
	case platform.KeyF4:
		return "F4"
	case platform.KeyF5:
		return "F5"
	case platform.KeyF6:
		return "F6"
	case platform.KeyF7:
		return "F7"
	case platform.KeyF8:
		return "F8"
	case platform.KeyF9:
		return "F9"
	case platform.KeyF10:
		return "F10"
	case platform.KeyF11:
		return "F11"
	case platform.KeyF12:
		return "F12"
	case platform.KeySpace:
		return "Space"
	case platform.KeyK:
		return "k"
	case platform.KeyJ:
		return "j"
	case platform.KeyH:
		return "h"
	case platform.KeyL:
		return "l"
	default:
		return "Unknown"
	}
}
