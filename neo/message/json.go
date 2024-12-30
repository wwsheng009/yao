package message

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/yaoapp/gou/helper"
	"github.com/yaoapp/kun/exception"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/kun/maps"
	"github.com/yaoapp/yao/openai"
)

// JSON the JSON message
type JSON struct{ *Message }

// New create a new JSON message
func New() *JSON {
	return &JSON{makeMessage()}
}

// NewOpenAI create a new JSON message
func NewOpenAI(data []byte) *JSON {

	if data == nil || len(data) == 0 {
		return nil
	}

	msg := makeMessage()
	text := string(data)
	data = []byte(strings.TrimPrefix(text, "data: "))
	switch {
	case strings.Contains(text, `"delta":{`) && strings.Contains(text, `"content":`):
		var message openai.Message
		err := jsoniter.Unmarshal(data, &message)
		if err != nil {
			msg.Text = err.Error()
			return &JSON{msg}
		}

		if len(message.Choices) > 0 {
			msg.Text = message.Choices[0].Delta.Content
		}
		break

	case strings.Contains(text, `[DONE]`):
		msg.Done = true
		break

	case strings.Contains(text, `"finish_reason":"stop"`):
		msg.Done = true
		break

	default:

		str := string(data)
		// Remove "data: " and "
		str = strings.TrimPrefix(str, "data: ")
		str = strings.Trim(str, "\"")
		msg.Type = "error"
		msg.Text = str
	}

	return &JSON{msg}
}

func (json *JSON) String() string {
	if json.Message == nil {
		return ""
	}
	return json.Message.Text
}

// Text set the text
func (json *JSON) Text(text string) *JSON {

	json.Message.Text = text
	if json.Message.Data != nil {
		replaced := helper.Bind(text, json.Message.Data)
		if replacedText, ok := replaced.(string); ok {
			json.Message.Text = replacedText
		}
	}

	return json
}

// Error set the error
func (json *JSON) Error(message interface{}) *JSON {
	json.Message.Type = "error"
	if err, ok := message.(error); ok {
		json.Message.Text = err.Error()
	} else if msg, ok := message.(string); ok {
		json.Message.Text = msg
	} else {
		json.Message.Text = fmt.Sprintf("%v", message)
	}
	return json
}

// Map set from map
func (json *JSON) Map(msg map[string]interface{}) *JSON {
	if msg == nil {
		return json
	}

	if text, ok := msg["text"].(string); ok {
		json.Message.Text = text
	}

	if typ, ok := msg["type"].(string); ok {
		json.Message.Text = typ
	}

	if done, ok := msg["done"].(bool); ok {
		json.Message.Done = done
	}

	if confirm, ok := msg["confirm"].(bool); ok {
		json.Message.Confirm = confirm
	}

	if command, ok := msg["command"].(map[string]interface{}); ok {
		json.Message.Command = &Command{}
		if id, ok := command["id"].(string); ok {
			json.Message.Command.ID = id
		}
		if name, ok := command["name"].(string); ok {
			json.Message.Command.Name = name
		}
		if request, ok := command["request"].(string); ok {
			json.Message.Command.Reqeust = request
		}
	}

	if actions, ok := msg["actions"].([]interface{}); ok {
		for _, action := range actions {
			if v, ok := action.(map[string]interface{}); ok {
				action := Action{}
				if name, ok := v["name"].(string); ok {
					action.Name = name
				}
				if t, ok := v["type"].(string); ok {
					action.Type = t
				}
				if payload, ok := v["payload"].(map[string]interface{}); ok {
					action.Payload = payload
				}

				if next, ok := v["next"].(string); ok {
					action.Next = next
				}
				json.Message.Actions = append(json.Message.Actions, action)
			}
		}
	}

	if data, ok := msg["data"].(map[string]interface{}); ok {
		json.Message.Data = data
	}

	return json
}

// Done set the done
func (json *JSON) Done() *JSON {
	json.Message.Done = true
	return json
}

// Confirm set the confirm
func (json *JSON) Confirm() *JSON {
	json.Message.Confirm = true
	return json
}

// Command set the command
func (json *JSON) Command(name, id, request string) *JSON {
	json.Message.Command = &Command{
		ID:      id,
		Name:    name,
		Reqeust: request,
	}
	return json
}

// Action set the action
func (json *JSON) Action(name string, t string, payload interface{}, next string) *JSON {

	if json.Message.Data != nil {
		payload = helper.Bind(payload, json.Message.Data)
	}

	json.Message.Actions = append(json.Message.Actions, Action{
		Name:    name,
		Type:    t,
		Payload: payload,
		Next:    next,
	})
	return json
}

// Bind replace with data
func (json *JSON) Bind(data map[string]interface{}) *JSON {
	if data == nil {
		return json
	}

	json.Message.Data = maps.Of(data).Dot()
	return json
}

// IsDone check if the message is done
func (json *JSON) IsDone() bool {
	return json.Message.Done
}

// Write the message
func (json *JSON) Write(w gin.ResponseWriter) bool {

	defer func() {
		if r := recover(); r != nil {
			message := "Write Response Exception: (if clinet close the connection, it's normal) \n  %s\n\n"
			color.Red(message, r)
		}
	}()

	data, err := jsoniter.Marshal(json.Message)
	if err != nil {
		log.Error("%s", err.Error())
		return false
	}

	data = append([]byte("data: "), data...)
	data = append(data, []byte("\n\n")...)

	_, err = w.Write(data)
	if err != nil {
		color.Red("Write JSON Message Error: %s", err.Error())
		return false
	}
	w.Flush()
	return true
}

// Append the message
func (json *JSON) Append(content []byte) []byte {
	return append(content, []byte(json.Message.Text)...)
}

func (json *JSON) writeError(w gin.ResponseWriter, message string) {
	data := []byte(`{"text":"` + strings.Trim(exception.New(message, 500).Message, "\"") + `","type":"error"}`)
	if json.Message.Done {
		data = []byte(`{"text":"` + strings.Trim(exception.New(message, 500).Message, "\"") + `","type":"error", "done":true}`)
	}
	data = append([]byte("data: "), data...)
	data = append(data, []byte("\n\n")...)
	_, err := w.Write(data)
	if err != nil {
		color.Red("Write JSON Message Error: %s", message)
	}
	w.Flush()
}
