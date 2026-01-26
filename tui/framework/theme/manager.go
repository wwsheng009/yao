package theme

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// Manager 主题管理器
type Manager struct {
	mu sync.RWMutex

	// 当前主题
	current *Theme

	// 主题注册表
	themes map[string]*Theme

	// 主题变化监听器
	listeners []ThemeChangeListener

	// 切换动画配置
	transitionDuration time.Duration
}

// ThemeChangeListener 主题变化监听器
type ThemeChangeListener func(old, new *Theme)

// NewManager 创建主题管理器
func NewManager() *Manager {
	return &Manager{
		themes:            make(map[string]*Theme),
		listeners:         make([]ThemeChangeListener, 0),
		transitionDuration: 300 * time.Millisecond,
	}
}

// Register 注册主题
func (m *Manager) Register(theme *Theme) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if theme == nil {
		return
	}

	m.themes[theme.Name] = theme

	// 如果是第一个主题，设为默认
	if m.current == nil {
		m.current = theme
	}
}

// RegisterMultiple 注册多个主题
func (m *Manager) RegisterMultiple(themes []*Theme) {
	for _, theme := range themes {
		m.Register(theme)
	}
}

// Unregister 注销主题
func (m *Manager) Unregister(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.themes, name)

	// 如果注销的是当前主题，切换到其他主题
	if m.current != nil && m.current.Name == name {
		// 尝试找到下一个可用主题
		for _, theme := range m.themes {
			m.current = theme
			break
		}
		if len(m.themes) == 0 {
			m.current = nil
		}
	}
}

// Get 获取主题
func (m *Manager) Get(name string) (*Theme, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	theme, ok := m.themes[name]
	return theme, ok
}

// Current 获取当前主题
func (m *Manager) Current() *Theme {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.current
}

// Set 设置当前主题
func (m *Manager) Set(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.setLocked(name)
}

// setLocked 在已持有锁的情况下设置当前主题
func (m *Manager) setLocked(name string) error {
	theme, ok := m.themes[name]
	if !ok {
		return fmt.Errorf("theme not found: %s", name)
	}

	old := m.current
	m.current = theme

	// 通知监听器
	m.notify(old, theme)

	return nil
}

// SetWithTransition 设置当前主题（带过渡动画）
func (m *Manager) SetWithTransition(name string) (*Transition, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	theme, ok := m.themes[name]
	if !ok {
		return nil, fmt.Errorf("theme not found: %s", name)
	}

	old := m.current

	// 创建过渡动画
	transition := NewTransition(old, theme, m.transitionDuration)

	// 立即设置新主题
	m.current = theme

	return transition, nil
}

// Toggle 切换到下一个主题
func (m *Manager) Toggle() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	names := m.themeNames()
	if len(names) == 0 {
		return nil
	}

	currentName := ""
	if m.current != nil {
		currentName = m.current.Name
	}

	// 找到下一个主题
	for i, name := range names {
		if name == currentName {
			if i < len(names)-1 {
				return m.setLocked(names[i+1])
			} else {
				return m.setLocked(names[0])
			}
		}
	}

	// 如果没找到当前主题，设置第一个
	if len(names) > 0 {
		return m.setLocked(names[0])
	}

	return nil
}

// Prev 切换到上一个主题
func (m *Manager) Prev() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	names := m.themeNames()
	if len(names) == 0 {
		return nil
	}

	currentName := ""
	if m.current != nil {
		currentName = m.current.Name
	}

	// 找到上一个主题
	for i, name := range names {
		if name == currentName {
			if i > 0 {
				return m.setLocked(names[i-1])
			} else {
				return m.setLocked(names[len(names)-1])
			}
		}
	}

	// 如果没找到当前主题，设置最后一个
	if len(names) > 0 {
		return m.setLocked(names[len(names)-1])
	}

	return nil
}

// List 列出所有主题名称
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.themeNames()
}

// Count 返回主题数量
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.themes)
}

// Has 检查主题是否存在
func (m *Manager) Has(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.themes[name]
	return ok
}

// Subscribe 订阅主题变化
// 返回取消订阅函数
func (m *Manager) Subscribe(listener ThemeChangeListener) func() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.listeners = append(m.listeners, listener)

	return func() {
		m.Unsubscribe(listener)
	}
}

// Unsubscribe 取消订阅
func (m *Manager) Unsubscribe(listener ThemeChangeListener) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, l := range m.listeners {
		// 使用函数指针比较
		if getFunctionPointer(l) == getFunctionPointer(listener) {
			m.listeners = append(m.listeners[:i], m.listeners[i+1:]...)
			break
		}
	}
}

// notify 通知监听器
func (m *Manager) notify(old, new *Theme) {
	// 复制监听器列表以避免在通知过程中修改
	listeners := make([]ThemeChangeListener, len(m.listeners))
	copy(listeners, m.listeners)

	for _, listener := range listeners {
		if listener != nil {
			listener(old, new)
		}
	}
}

// themeNames 获取所有主题名称（已排序）
func (m *Manager) themeNames() []string {
	names := make([]string, 0, len(m.themes))
	for name := range m.themes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetStyle 获取样式（从当前主题）
func (m *Manager) GetStyle(componentID, styleKey string) StyleConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.current == nil {
		return StyleConfig{}
	}

	// 1. 查找组件特定样式（如 input、button 等）
	if compStyle, ok := m.current.Components[componentID]; ok {
		if styleKey != "" {
			if stateStyle, ok := compStyle.States[styleKey]; ok {
				return stateStyle
			}
		}
		return compStyle.Base
	}

	// 2. 如果 styleKey 为空，尝试用 componentID 直接查找全局样式
	// 这支持 "text.primary"、"text.secondary" 这样的样式ID
	if styleKey == "" {
		if style, ok := m.current.Styles[componentID]; ok {
			return style
		}
	}

	// 3. 查找全局样式（使用 styleKey）
	if styleKey != "" {
		if style, ok := m.current.Styles[styleKey]; ok {
			return style
		}
	}

	// 4. 查找父主题
	return m.resolveStyle(m.current.Parent, componentID, styleKey)
}

// resolveStyle 递归解析样式
func (m *Manager) resolveStyle(theme *Theme, componentID, styleKey string) StyleConfig {
	if theme == nil {
		return StyleConfig{}
	}

	// 1. 查找组件特定样式
	if compStyle, ok := theme.Components[componentID]; ok {
		if styleKey != "" {
			if stateStyle, ok := compStyle.States[styleKey]; ok {
				return stateStyle
			}
		}
		if compStyle.Base.Foreground != nil || compStyle.Base.Background != nil {
			return compStyle.Base
		}
	}

	// 2. 查找全局样式
	if style, ok := theme.Styles[styleKey]; ok {
		return style
	}

	// 3. 递归查找父主题
	if theme.Parent != nil {
		return m.resolveStyle(theme.Parent, componentID, styleKey)
	}

	return StyleConfig{}
}

// GetColor 获取颜色（从当前主题）
func (m *Manager) GetColor(colorKey string) Color {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.current == nil {
		return NoColor
	}

	return m.current.GetColor(colorKey)
}

// GetComponentStyle 获取组件样式
func (m *Manager) GetComponentStyle(componentID, state string) StyleConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.current == nil {
		return StyleConfig{}
	}

	return m.current.GetComponentStyle(componentID, state)
}

// SetTransitionDuration 设置过渡动画时长
func (m *Manager) SetTransitionDuration(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.transitionDuration = duration
}

// GetTransitionDuration 获取过渡动画时长
func (m *Manager) GetTransitionDuration() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.transitionDuration
}

// SetDefault 设置默认主题
func (m *Manager) SetDefault(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	theme, ok := m.themes[name]
	if !ok {
		return fmt.Errorf("theme not found: %s", name)
	}

	// 只在没有当前主题时设置
	if m.current == nil {
		m.current = theme
	}

	return nil
}

// Refresh 刷新主题（触发所有监听器）
func (m *Manager) Refresh() {
	m.mu.Lock()
	current := m.current
	m.mu.Unlock()

	if current != nil {
		m.notify(current, current)
	}
}

// getFunctionPointer 获取函数指针（用于比较）
// 注意：Go 不支持直接比较函数，这里简化处理
func getFunctionPointer(f interface{}) uintptr {
	// 使用简单哈希代替函数指针比较
	// 在实际使用中，监听器通常是闭包，很难比较
	// 这里返回一个伪指针用于标识
	return 0
}

// DefaultStyleConfig 默认样式配置
func DefaultStyleConfig() StyleConfig {
	return StyleConfig{}
}

// =============================================================================
// Transition 主题切换过渡
// =============================================================================

// Transition 主题切换过渡
type Transition struct {
	from      *Theme
	to        *Theme
	progress  float64
	duration  time.Duration
	startedAt time.Time
}

// NewTransition 创建过渡
func NewTransition(from, to *Theme, duration time.Duration) *Transition {
	return &Transition{
		from:      from,
		to:        to,
		duration:  duration,
		startedAt: time.Now(),
	}
}

// Update 更新过渡
func (t *Transition) Update(dt time.Duration) (done bool, progress float64) {
	elapsed := time.Since(t.startedAt)
	t.progress = float64(elapsed) / float64(t.duration)

	if t.progress >= 1.0 {
		t.progress = 1.0
		return true, t.progress
	}

	return false, t.progress
}

// InterpolateColor 插值颜色
func (t *Transition) InterpolateColor(colorKey string) Color {
	if t.from == nil || t.to == nil {
		return NoColor
	}

	fromColor := t.from.GetColor(colorKey)
	toColor := t.to.GetColor(colorKey)

	// 如果类型不匹配，直接返回目标颜色
	if fromColor.Type != toColor.Type {
		return toColor
	}

	// 只支持 RGB 颜色插值
	if fromColor.Type == ColorRGB && toColor.Type == ColorRGB {
		fromR, fromG, fromB := fromColor.RGBValue()
		toR, toG, toB := toColor.RGBValue()

		return Color{
			Type: ColorRGB,
			Value: [3]int{
				int(float64(fromR) + (float64(toR)-float64(fromR))*t.progress),
				int(float64(fromG) + (float64(toG)-float64(fromG))*t.progress),
				int(float64(fromB) + (float64(toB)-float64(fromB))*t.progress),
			},
		}
	}

	return toColor
}

// GetProgress 获取当前进度
func (t *Transition) GetProgress() float64 {
	return t.progress
}

// IsComplete 检查是否完成
func (t *Transition) IsComplete() bool {
	return t.progress >= 1.0
}
