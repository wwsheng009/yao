package tui

import (
	"reflect"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// MessageSubscriptionManager 管理组件消息订阅
type MessageSubscriptionManager struct {
	// subscriptions 存储消息类型到订阅组件 ID 的映射
	sync.RWMutex
	subscriptions map[string][]string // message type -> list of component IDs

	// componentSubscriptions 存储组件 ID 到其订阅的消息类型
	componentSubscriptions map[string][]string // component ID -> list of message types
}

// NewMessageSubscriptionManager 创建新的消息订阅管理器
func NewMessageSubscriptionManager() *MessageSubscriptionManager {
	return &MessageSubscriptionManager{
		subscriptions:          make(map[string][]string),
		componentSubscriptions: make(map[string][]string),
	}
}

// Subscribe 订阅组件到指定消息类型
func (m *MessageSubscriptionManager) Subscribe(componentID string, messageTypes []string) {
	if len(messageTypes) == 0 {
		return
	}

	m.Lock()
	defer m.Unlock()

	// 存储组件的消息订阅
	m.componentSubscriptions[componentID] = messageTypes

	// 为每个消息类型添加组件订阅
	for _, msgType := range messageTypes {
		// 查找是否已存在该消息类型的订阅
		found := false
		for _, id := range m.subscriptions[msgType] {
			if id == componentID {
				found = true
				break
			}
		}

		// 如果不存在，则添加订阅
		if !found {
			m.subscriptions[msgType] = append(m.subscriptions[msgType], componentID)
		}
	}
}

// Unsubscribe 取消组件的所有订阅
func (m *MessageSubscriptionManager) Unsubscribe(componentID string) {
	m.Lock()
	defer m.Unlock()

	// 获取组件订阅的消息类型
	messageTypes, exists := m.componentSubscriptions[componentID]
	if !exists {
		return
	}

	// 从每个消息类型中移除该组件
	for _, msgType := range messageTypes {
		ids := m.subscriptions[msgType]
		var newIDs []string
		for _, id := range ids {
			if id != componentID {
				newIDs = append(newIDs, id)
			}
		}
		m.subscriptions[msgType] = newIDs
	}

	// 删除组件的订阅记录
	delete(m.componentSubscriptions, componentID)
}

// GetSubscribers 获取订阅了指定消息类型的所有组件 ID
func (m *MessageSubscriptionManager) GetSubscribers(messageType string) []string {
	m.RLock()
	defer m.RUnlock()

	subs, exists := m.subscriptions[messageType]
	if !exists || len(subs) == 0 {
		return nil
	}

	// 返回副本
	result := make([]string, len(subs))
	copy(result, subs)
	return result
}

// GetAllSubscribedComponents 获取所有有订阅的组件 ID
func (m *MessageSubscriptionManager) GetAllSubscribedComponents() []string {
	m.RLock()
	defer m.RUnlock()

	result := make([]string, 0, len(m.componentSubscriptions))
	for id := range m.componentSubscriptions {
		result = append(result, id)
	}
	return result
}

// GetComponentSubscriptions 获取组件订阅的所有消息类型
func (m *MessageSubscriptionManager) GetComponentSubscriptions(componentID string) []string {
	m.RLock()
	defer m.RUnlock()

	subs, exists := m.componentSubscriptions[componentID]
	if !exists || len(subs) == 0 {
		return nil
	}

	// 返回副本
	result := make([]string, len(subs))
	copy(result, subs)
	return result
}

// Clear 清除所有订阅
func (m *MessageSubscriptionManager) Clear() {
	m.Lock()
	defer m.Unlock()

	m.subscriptions = make(map[string][]string)
	m.componentSubscriptions = make(map[string][]string)
}

// getMessageType 获取消息的类型字符串
func getMessageType(msg tea.Msg) string {
	if msg == nil {
		return "nil"
	}

	// 使用反射获取消息类型
	msgValue := reflect.ValueOf(msg)
	msgType := msgValue.Type()

	// 处理指针类型
	if msgType.Kind() == reflect.Ptr {
		msgType = msgType.Elem()
	}

	// 返回类型名称
	return msgType.String()
}

// 常用消息类型常量
const (
	MsgTypeKeyDown    = "tea.KeyMsg"
	MsgTypeMouseDown  = "tea.MouseMsg"
	MsgTypeWindowSize = "tea.WindowSizeMsg"
	MsgTypeQuit       = "tea.QuitMsg"
	MsgTypeFocusMsg   = "FocusMsg"
	MsgTypeActionMsg  = "ActionMsg"
	MsgTypeErrorMsg   = "Error"
	MsgTypeInfoMsg    = "Info"
	MsgTypeSuccessMsg = "Success"
	MsgTypeWarningMsg = "Warning"
)

// GetMessageTypeString 将常见消息类型转换为常量字符串
func GetMessageTypeString(msg tea.Msg) string {
	switch msg.(type) {
	case tea.KeyMsg:
		return MsgTypeKeyDown
	case tea.MouseMsg:
		return MsgTypeMouseDown
	case tea.WindowSizeMsg:
		return MsgTypeWindowSize
	case tea.QuitMsg:
		return MsgTypeQuit
	default:
		return getMessageType(msg)
	}
}
