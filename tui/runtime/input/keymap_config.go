package input

import (
	"fmt"
	"os"
	"strings"

	"github.com/yaoapp/yao/tui/runtime/platform"
	"github.com/yaoapp/yao/tui/runtime/action"
	"gopkg.in/yaml.v3"
)

// ==============================================================================
// KeyMap Configuration Loader (V3)
// ==============================================================================
// 支持从 YAML 文件加载按键映射配置

// KeyMapConfig 配置结构
type KeyMapConfig struct {
	Navigation map[string][]string `yaml:"navigation"`
	Editing    map[string][]string `yaml:"editing"`
	Form       map[string][]string `yaml:"form"`
	System     map[string][]string `yaml:"system"`
	Scroll     map[string][]string `yaml:"scroll"`
	Contexts   map[string]map[string][]string `yaml:"contexts"`
}

// LoadKeyMap 从配置文件加载 KeyMap
// 如果文件不存在或解析失败，返回默认的 KeyMap
func LoadKeyMap(path string) (*KeyMap, error) {
	km := NewKeyMap()

	// 尝试读取配置文件
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，返回默认 KeyMap
			return km, nil
		}
		return nil, fmt.Errorf("failed to read keymap file: %w", err)
	}

	// 解析 YAML
	var config KeyMapConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse keymap file: %w", err)
	}

	// 应用配置
	applyConfig(km, &config)

	return km, nil
}

// LoadKeyMapFromString 从 YAML 字符串加载 KeyMap
func LoadKeyMapFromString(yamlStr string) (*KeyMap, error) {
	km := NewKeyMap()

	var config KeyMapConfig
	if err := yaml.Unmarshal([]byte(yamlStr), &config); err != nil {
		return nil, fmt.Errorf("failed to parse keymap config: %w", err)
	}

	applyConfig(km, &config)

	return km, nil
}

// applyConfig 应用配置到 KeyMap
func applyConfig(km *KeyMap, config *KeyMapConfig) {
	// 应用导航映射
	if config.Navigation != nil {
		applyActionMappings(km, config.Navigation, navigationActionMap)
	}

	// 应用编辑映射
	if config.Editing != nil {
		applyActionMappings(km, config.Editing, editingActionMap)
	}

	// 应用表单映射
	if config.Form != nil {
		applyActionMappings(km, config.Form, formActionMap)
	}

	// 应用系统映射
	if config.System != nil {
		applyActionMappings(km, config.System, systemActionMap)
	}

	// 应用滚动映射
	if config.Scroll != nil {
		applyActionMappings(km, config.Scroll, scrollActionMap)
	}

	// 应用上下文映射
	for ctx, mappings := range config.Contexts {
		applyContextMappings(km, ctx, mappings)
	}
}

// applyActionMappings 应用动作映射
func applyActionMappings(km *KeyMap, config map[string][]string, actionMap map[string]action.ActionType) {
	for name, keys := range config {
		if actionType, ok := actionMap[name]; ok {
			for _, key := range keys {
				km.Bind(key, actionType)
			}
		}
	}
}

// applyContextMappings 应用上下文映射
func applyContextMappings(km *KeyMap, context string, mappings map[string][]string) {
	// 合并所有动作类型映射
	allActions := make(map[string]action.ActionType)
	for k, v := range navigationActionMap {
		allActions[k] = v
	}
	for k, v := range editingActionMap {
		allActions[k] = v
	}
	for k, v := range formActionMap {
		allActions[k] = v
	}
	for k, v := range systemActionMap {
		allActions[k] = v
	}
	for k, v := range scrollActionMap {
		allActions[k] = v
	}

	// 为上下文创建特殊键映射
	ctxMap := make(map[platform.SpecialKey]action.ActionType)

	for name, keys := range mappings {
		if actionType, ok := allActions[name]; ok {
			for _, key := range keys {
				if special := parseSpecialKey(key); special != platform.KeyUnknown {
					ctxMap[special] = actionType
				}
			}
		}
	}

	if len(ctxMap) > 0 {
		km.BindContext(context, ctxMap)
	}
}

// ==============================================================================
// Action 类型映射表
// ==============================================================================

var navigationActionMap = map[string]action.ActionType{
	"next":      action.ActionNavigateNext,
	"prev":      action.ActionNavigatePrev,
	"up":        action.ActionNavigateUp,
	"down":      action.ActionNavigateDown,
	"left":      action.ActionNavigateLeft,
	"right":     action.ActionNavigateRight,
	"first":     action.ActionNavigateFirst,
	"last":      action.ActionNavigateLast,
	"page_up":   action.ActionNavigatePageUp,
	"page_down": action.ActionNavigatePageDown,
}

var editingActionMap = map[string]action.ActionType{
	"delete_char":  action.ActionDeleteChar,
	"delete_word":  action.ActionDeleteWord,
	"delete_line":  action.ActionDeleteLine,
	"backspace":    action.ActionBackspace,
	"select_all":   action.ActionSelectAll,
	"select_word":  action.ActionSelectWord,
	"select_line":  action.ActionSelectLine,
	"cursor_home":  action.ActionCursorHome,
	"cursor_end":   action.ActionCursorEnd,
	"cursor_left":  action.ActionCursorLeft,
	"cursor_right": action.ActionCursorRight,
	"clear":        action.ActionClear,
}

var formActionMap = map[string]action.ActionType{
	"submit":   action.ActionSubmit,
	"cancel":   action.ActionCancel,
	"validate": action.ActionValidate,
	"reset":    action.ActionReset,
}

var systemActionMap = map[string]action.ActionType{
	"quit":    action.ActionQuit,
	"help":    action.ActionHelp,
	"search":  action.ActionSearch,
	"refresh": action.ActionRefresh,
	"copy":    action.ActionCopy,
	"paste":   action.ActionPaste,
	"undo":    action.ActionUndo,
	"redo":    action.ActionRedo,
}

var scrollActionMap = map[string]action.ActionType{
	"up":     action.ActionScrollUp,
	"down":   action.ActionScrollDown,
	"left":   action.ActionScrollLeft,
	"right":  action.ActionScrollRight,
	"zoom_in":  action.ActionZoomIn,
	"zoom_out": action.ActionZoomOut,
}

// ==============================================================================
// 按键解析辅助函数
// ==============================================================================

// parseSpecialKey 解析特殊键字符串
func parseSpecialKey(key string) platform.SpecialKey {
	switch key {
	case "Esc", "Escape":
		return platform.KeyEscape
	case "Enter":
		return platform.KeyEnter
	case "Tab":
		return platform.KeyTab
	case "Backspace":
		return platform.KeyBackspace
	case "Delete":
		return platform.KeyDelete
	case "Insert":
		return platform.KeyInsert
	case "Up":
		return platform.KeyUp
	case "Down":
		return platform.KeyDown
	case "Left":
		return platform.KeyLeft
	case "Right":
		return platform.KeyRight
	case "Home":
		return platform.KeyHome
	case "End":
		return platform.KeyEnd
	case "PageUp", "PgUp":
		return platform.KeyPageUp
	case "PageDown", "PgDn":
		return platform.KeyPageDown
	case "F1":
		return platform.KeyF1
	case "F2":
		return platform.KeyF2
	case "F3":
		return platform.KeyF3
	case "F4":
		return platform.KeyF4
	case "F5":
		return platform.KeyF5
	case "F6":
		return platform.KeyF6
	case "F7":
		return platform.KeyF7
	case "F8":
		return platform.KeyF8
	case "F9":
		return platform.KeyF9
	case "F10":
		return platform.KeyF10
	case "F11":
		return platform.KeyF11
	case "F12":
		return platform.KeyF12
	case "Space":
		return platform.KeySpace
	case "k", "K":
		return platform.KeyK
	case "j", "J":
		return platform.KeyJ
	case "h", "H":
		return platform.KeyH
	case "l", "L":
		return platform.KeyL
	default:
		return platform.KeyUnknown
	}
}

// parseKeyBinding 解析按键绑定字符串，返回 combo 格式
// 例如: "C-c" -> "C-c", "Ctrl+C" -> "C-c"
func parseKeyBinding(keyStr string) string {
	// 处理各种格式
	keyStr = strings.TrimSpace(keyStr)

	// 替换常见前缀
	replacer := strings.NewReplacer(
		"Ctrl+", "C-",
		"ctrl+", "C-",
		"Alt+", "A-",
		"alt+", "A-",
		"Shift+", "S-",
		"shift+", "S-",
		"Control+", "C-",
		"control+", "C-",
	)

	result := replacer.Replace(keyStr)

	// 处理全大写形式
	if strings.HasPrefix(result, "C-") || strings.HasPrefix(result, "A-") || strings.HasPrefix(result, "S-") {
		// 保持格式
	}

	return result
}

// ==============================================================================
// 默认配置
// ==============================================================================

// DefaultKeyMapConfigYAML 返回默认配置的 YAML 字符串
func DefaultKeyMapConfigYAML() string {
	return `
# KeyMap 配置文件
# 格式: Ctrl+C 表示按住 Ctrl 键同时按 C 键

navigation:
  next: ["Tab", "Down"]
  prev: ["Shift+Tab", "Up"]
  up: "Up"
  down: "Down"
  left: "Left"
  right: "Right"
  first: "Home"
  last: "End"
  page_up: "PageUp"
  page_down: "PageDown"

editing:
  delete_char: ["Backspace", "Delete"]
  delete_word: "Ctrl+W"
  delete_line: "Ctrl+K"
  backspace: "Backspace"
  select_all: "Ctrl+A"
  cursor_home: "Home"
  cursor_end: "End"
  cursor_left: "Left"
  cursor_right: "Right"
  clear: "Ctrl+U"

form:
  submit: ["Enter", "Ctrl+Enter"]
  cancel: "Escape"
  validate: "Ctrl+V"
  reset: "Ctrl+R"

system:
  quit: ["Ctrl+C", "Ctrl+Q"]
  help: "F1"
  search: "Ctrl+F"
  refresh: "F5"
  copy: "Ctrl+C"
  paste: "Ctrl+V"
  undo: "Ctrl+Z"
  redo: "Ctrl+Y"

scroll:
  up: "Up"
  down: "Down"
  left: "Left"
  right: "Right"
  zoom_in: "Ctrl+Plus"
  zoom_out: "Ctrl+Minus"

# 上下文特定映射
contexts:
  modal:
    submit: ["Enter", "Ctrl+Enter"]
    cancel: "Escape"

  text_input:
    left: "Ctrl+B"
    right: "Ctrl+F"
    home: "Ctrl+A"
    end: "Ctrl+E"
`
}
