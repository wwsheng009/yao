package event

// EventHandler 事件处理器接口
type EventHandler interface {
	HandleEvent(Event) bool
}

// EventHandlerFunc 事件处理器函数
type EventHandlerFunc func(Event) bool

func (f EventHandlerFunc) HandleEvent(ev Event) bool {
	return f(ev)
}

// Router 事件路由器
type Router struct {
	globalHandlers  map[EventType][]EventHandler
	captureHandlers []EventHandler
}

// NewRouter 创建事件路由器
func NewRouter() *Router {
	return &Router{
		globalHandlers: make(map[EventType][]EventHandler),
	}
}

// Subscribe 订阅事件
func (r *Router) Subscribe(eventType EventType, handler EventHandler) func() {
	r.globalHandlers[eventType] = append(r.globalHandlers[eventType], handler)
	return func() {
		r.Unsubscribe(eventType, handler)
	}
}

// Unsubscribe 取消订阅
func (r *Router) Unsubscribe(eventType EventType, handler EventHandler) {
	handlers := r.globalHandlers[eventType]
	for i, h := range handlers {
		if h == handler {
			r.globalHandlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Route 路由事件
func (r *Router) Route(ev Event) {
	// 捕获阶段
	for _, handler := range r.captureHandlers {
		if handler.HandleEvent(ev) {
			if ev.IsPropagationStopped() {
				return
			}
		}
	}

	// 全局处理器
	if handlers, ok := r.globalHandlers[ev.Type()]; ok {
		for _, handler := range handlers {
			if handler.HandleEvent(ev) {
				if ev.IsPropagationStopped() {
					return
				}
			}
		}
	}

	// 目标阶段
	if target := ev.Target(); target != nil {
		target.HandleEvent(ev)
	}
}

// KeyMap 快捷键映射
type KeyMap struct {
	bindings map[string]EventHandler
}

// NewKeyMap 创建快捷键映射
func NewKeyMap() *KeyMap {
	return &KeyMap{
		bindings: make(map[string]EventHandler),
	}
}

// BindFunc 绑定快捷键到函数
func (k *KeyMap) BindFunc(combo string, handler func(*KeyEvent)) error {
	// TODO: 解析快捷键
	k.bindings[combo] = EventHandlerFunc(func(ev Event) bool {
		if keyEv, ok := ev.(*KeyEvent); ok {
			handler(keyEv)
			return true
		}
		return false
	})
	return nil
}

// Lookup 查找快捷键处理器
func (k *KeyMap) Lookup(ev *KeyEvent) (EventHandler, bool) {
	// 优先按字符键查找
	if ev.Key > 0 {
		if handler, ok := k.bindings[string(ev.Key)]; ok {
			return handler, true
		}
	}

	// 按特殊键查找
	if ev.Special != KeyUnknown {
		specialName := k.specialKeyName(ev.Special)
		if handler, ok := k.bindings[specialName]; ok {
			return handler, true
		}
	}

	return nil, false
}

// specialKeyName 获取特殊键名称
func (k *KeyMap) specialKeyName(key SpecialKey) string {
	names := map[SpecialKey]string{
		KeyEscape:   "escape",
		KeyEnter:    "enter",
		KeyTab:      "tab",
		KeyBackspace: "backspace",
		KeyDelete:   "delete",
		KeyInsert:   "insert",
		KeyUp:       "up",
		KeyDown:     "down",
		KeyLeft:     "left",
		KeyRight:    "right",
		KeyHome:     "home",
		KeyEnd:      "end",
		KeyPageUp:   "pageup",
		KeyPageDown: "pagedown",
		KeyF1:       "f1",
		KeyF2:       "f2",
		KeyF3:       "f3",
		KeyF4:       "f4",
		KeyF5:       "f5",
		KeyF6:       "f6",
		KeyF7:       "f7",
		KeyF8:       "f8",
		KeyF9:       "f9",
		KeyF10:      "f10",
		KeyF11:      "f11",
		KeyF12:      "f12",
		KeySpace:    "space",
	}
	if name, ok := names[key]; ok {
		return name
	}
	return ""
}

// MousePressEvent 鼠标按下事件
type MousePressEvent struct {
	BaseEvent
	X      int
	Y      int
	Button MouseButton
}

// MouseButton 鼠标按钮
type MouseButton int

const (
	MouseNone MouseButton = iota
	MouseLeft
	MouseMiddle
	MouseRight
)

// ResizeEvent 窗口大小改变事件
type ResizeEvent struct {
	BaseEvent
	OldWidth  int
	OldHeight int
	NewWidth  int
	NewHeight int
}

// CloseEvent 关闭事件
type CloseEvent struct {
	BaseEvent
}
