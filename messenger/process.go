package messenger

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yaoapp/gou/process"
	"github.com/yaoapp/kun/exception"
	messengertypes "github.com/yaoapp/yao/messenger/types"
)

func init() {
	process.RegisterGroup("messenger", map[string]process.Handler{
		"enabled":    ProcessEnabled,
		"send":       ProcessSend,
		"sendBatch":  ProcessSendBatch,
		"sendT":      ProcessSendWithTemplate,
		"sendTBatch": ProcessSendWithTemplateBatch,
	})
}

func CheckMessageInit() {
	if Instance == nil {
		exception.New("Messenger service not initialized", 500).Throw()
	}
}
func ProcessEnabled(process *process.Process) interface{} {
	return Instance == nil
}

// Common basic info extracted from process args
type sendInfo struct {
	channel     string
	messageType messengertypes.MessageType
	template    string
}

// 辅助函数：解析共有的前三个参数并修复逻辑 Bug
func parseBaseSendInfo(process *process.Process) (sendInfo, error) {
	process.ValidateArgNums(4)
	//channel id
	channel := process.ArgsString(0, "default")
	// 解析消息类型
	messageInput := strings.ToLower(process.ArgsString(1, "email"))
	var messageType messengertypes.MessageType

	switch messageInput {
	case "sms":
		messageType = messengertypes.MessageTypeSMS
	case "whatsapp":
		messageType = messengertypes.MessageTypeWhatsApp
	default:
		messageType = messengertypes.MessageTypeEmail
	}
	//template id
	template := process.ArgsString(2)

	return sendInfo{
		channel:     channel,
		messageType: messageType,
		template:    template,
	}, nil
}

// demo yao run messenger.sendt "default" "email" "zh-cn.invite_member" '::{"to":"to@example.com"}'
func ProcessSendWithTemplate(process *process.Process) interface{} {
	CheckMessageInit()
	info, err := parseBaseSendInfo(process)
	if err != nil {
		return err
	}

	templateData := map[string]interface{}{}
	value, ok := process.Args[3].(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to convert templateData to map[string]interface{}")
	}
	templateData = value

	if err := Instance.SendT(process.Context, info.channel, info.template, templateData, info.messageType); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

// demo yao run messenger.sendt "default" "email" "zh-cn.invite_member" '::[{"to":"to@example.com"},{"to":"to2@example.com"}]'
func ProcessSendWithTemplateBatch(process *process.Process) interface{} {
	CheckMessageInit()
	info, err := parseBaseSendInfo(process)
	if err != nil {
		return err
	}

	templateDatas := process.ArgsRecords(3) // 内部通常处理了长度判断
	dataList := make([]messengertypes.TemplateData, 0, len(templateDatas))
	for _, td := range templateDatas {
		dataList = append(dataList, messengertypes.TemplateData(td))
	}

	if err := Instance.SendTBatch(process.Context, info.channel, info.template, dataList, info.messageType); err != nil {
		return fmt.Errorf("failed to send message batch: %w", err)
	}
	return nil
}

func ProcessSend(process *process.Process) interface{} {
	CheckMessageInit()
	process.ValidateArgNums(2)
	channel := process.ArgsString(0, "default")
	message1 := process.Args[1]
	//convert interface to messengertypes.Message
	var message messengertypes.Message

	// Convert the interface{} to JSON bytes
	data, err := json.Marshal(message1)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	// Unmarshal JSON bytes into your specific struct
	if err := json.Unmarshal(data, &message); err != nil {
		return fmt.Errorf("failed to decode into Message struct: %w", err)
	}
	err = Instance.Send(process.Context, channel, &message)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func ProcessSendBatch(process *process.Process) interface{} {
	CheckMessageInit()
	process.ValidateArgNums(2)

	channel := process.ArgsString(0, "default")
	messagesArg := process.Args[1]
	var messages []messengertypes.Message
	data, err := json.Marshal(messagesArg)
	if err != nil {
		return fmt.Errorf("failed to marshal messages array: %w", err)
	}

	if err := json.Unmarshal(data, &messages); err != nil {
		return fmt.Errorf("failed to decode into Message slice: %w", err)
	}
	messagePointers := make([]*messengertypes.Message, len(messages))
	for i := range messages {
		messagePointers[i] = &messages[i]
	}
	err = Instance.SendBatch(process.Context, channel, messagePointers)
	if err != nil {
		return fmt.Errorf("failed to send message batch: %w", err)
	}
	return nil
}
