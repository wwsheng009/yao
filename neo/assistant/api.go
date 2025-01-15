package assistant

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yaoapp/gou/fs"
	"github.com/yaoapp/gou/process"
	chatctx "github.com/yaoapp/yao/neo/context"
	"github.com/yaoapp/yao/neo/message"
	chatMessage "github.com/yaoapp/yao/neo/message"
)

// Get get the assistant by id
func Get(id string) (*Assistant, error) {
	return LoadStore(id)
}

// GetByConnector get the assistant by connector
func GetByConnector(connector string, name string) (*Assistant, error) {
	id := "connector:" + connector

	assistant, exists := loaded.Get(id)
	if exists {
		return assistant, nil
	}

	data := map[string]interface{}{
		"assistant_id": id,
		"connector":    connector,
		"description":  "Default assistant for " + connector,
		"name":         name,
		"type":         "assistant",
	}

	assistant, err := loadMap(data)
	if err != nil {
		return nil, err
	}
	loaded.Put(assistant)
	return assistant, nil
}

// Execute implements the execute functionality
func (ast *Assistant) Execute(c *gin.Context, ctx chatctx.Context, input string, options map[string]interface{}) error {
	messages, err := ast.withHistory(ctx, input)
	if err != nil {
		return err
	}

	options = ast.withOptions(options)

	// Run init hook
	res, err := ast.HookInit(c, ctx, messages, options)
	if err != nil {
		return err
	}
	refAst := &ast
	// Switch to the new assistant if necessary
	if res.AssistantID != ctx.AssistantID {
		newAst, err := Get(res.AssistantID)
		if err != nil {
			return err
		}
		// *ast = *newAst
		refAst = &newAst
	}

	// Handle next action
	if res.Next != nil {
		return res.Next.Execute(c, ctx)
	}

	// Update options if provided
	if res.Options != nil {
		options = res.Options
	}

	// messages
	if res.Input != nil {
		messages = res.Input
	}

	// Only proceed with chat stream if no specific next action was handled
	return (*refAst).handleChatStream(c, ctx, messages, options)
}

// Execute the next action
func (next *NextAction) Execute(c *gin.Context, ctx chatctx.Context) error {
	switch next.Action {

	case "process":
		if next.Payload == nil {
			return fmt.Errorf("payload is required")
		}

		name, ok := next.Payload["name"].(string)
		if !ok {
			return fmt.Errorf("process name should be string")
		}

		args := []interface{}{}
		if v, ok := next.Payload["args"].([]interface{}); ok {
			args = v
		}

		// Add context and writer to args
		args = append(args, ctx, c.Writer)
		p, err := process.Of(name, args...)
		if err != nil {
			return fmt.Errorf("get process error: %s", err.Error())
		}

		err = p.Execute()
		if err != nil {
			return fmt.Errorf("execute process error: %s", err.Error())
		}
		defer p.Release()

		return nil

	case "assistant":
		if next.Payload == nil {
			return fmt.Errorf("payload is required")
		}

		// Get assistant id
		id, ok := next.Payload["assistant_id"].(string)
		if !ok {
			return fmt.Errorf("assistant id should be string")
		}

		// Get assistant
		assistant, err := Get(id)
		if err != nil {
			return fmt.Errorf("get assistant error: %s", err.Error())
		}

		// Input
		input, ok := next.Payload["input"].(string)
		if !ok {
			return fmt.Errorf("input should be string")
		}

		// Options
		options := map[string]interface{}{}
		if v, ok := next.Payload["options"].(map[string]interface{}); ok {
			options = v
		}
		return assistant.Execute(c, ctx, input, options)

	case "exit":
		return nil

	default:
		return fmt.Errorf("unknown action: %s", next.Action)
	}
}

// handleChatStream manages the streaming chat interaction with the AI
func (ast *Assistant) handleChatStream(c *gin.Context, ctx chatctx.Context, messages []message.Message, options map[string]interface{}) error {
	clientBreak := make(chan bool, 1)
	done := make(chan bool, 1)
	content := message.NewContent("text")

	// Chat with AI in background
	go func() {
		err := ast.streamChat(c, ctx, messages, options, clientBreak, done, content)
		if err != nil {
			chatMessage.New().Error(err).Done().Write(c.Writer)
		}

		ast.saveChatHistory(ctx, messages, content)
		done <- true
	}()

	// Wait for completion or client disconnect
	select {
	case <-done:
		return nil
	case <-c.Writer.CloseNotify():
		clientBreak <- true
		return nil
	}
}

// streamChat handles the streaming chat interaction
func (ast *Assistant) streamChat(
	c *gin.Context,
	ctx chatctx.Context,
	messages []message.Message,
	options map[string]interface{},
	clientBreak chan bool,
	done chan bool,
	content *message.Content) error {

	return ast.Chat(c.Request.Context(), messages, options, func(data []byte) int {
		select {
		case <-clientBreak:
			return 0 // break

		default:
			msg := chatMessage.NewOpenAI(data)
			if msg == nil {
				return 1 // continue
			}

			// Handle error
			if msg.Type == "error" {
				value := msg.String()
				res, hookErr := ast.HookFail(c, ctx, messages, content.String(), fmt.Errorf("%s", value))
				if hookErr == nil && res != nil && (res.Output != "" || res.Error != "") {
					value = res.Output
					if res.Error != "" {
						value = res.Error
					}
				}
				chatMessage.New().Error(value).Done().Write(c.Writer)
				return 0 // break
			}

			// Handle tool call
			if msg.Type == "tool_calls" {
				content.SetType("function") // Set type to function
				// Set id
				if id, ok := msg.Props["id"].(string); ok && id != "" {
					content.SetID(id)
				}

				// Set name
				if name, ok := msg.Props["name"].(string); ok && name != "" {
					content.SetName(name)
				}
			}

			// Append content and send message
			value := msg.String()
			content.Append(value)
			if value != "" {
				// Handle stream
				res, err := ast.HookStream(c, ctx, messages, content.String(), msg.Type == "tool_calls")
				if err == nil && res != nil {
					if res.Output != "" {
						value = res.Output
					}

					if res.Next != nil {
						err = res.Next.Execute(c, ctx)
						if err != nil {
							chatMessage.New().Error(err.Error()).Done().Write(c.Writer)
						}

						done <- true
						return 0 // break
					}

					if res.Silent {
						return 1 // continue
					}
				}

				chatMessage.New().
					Map(map[string]interface{}{
						"text": value,
						"done": msg.IsDone,
					}).
					Write(c.Writer)
			}

			// Complete the stream
			if msg.IsDone {
				// if value == "" {
				// 	msg.Write(c.Writer)
				// }

				// Call HookDone
				content.SetStatus(message.ContentStatusDone)
				res, hookErr := ast.HookDone(c, ctx, messages, content.String(), msg.Type == "tool_calls")
				if hookErr == nil && res != nil {
					if res.Output != "" {
						chatMessage.New().
							Map(map[string]interface{}{
								"text": res.Output,
								"done": true,
							}).
							Write(c.Writer)
					}

					if res.Next != nil {
						err := res.Next.Execute(c, ctx)
						if err != nil {
							chatMessage.New().Error(err.Error()).Done().Write(c.Writer)
						}
						done <- true
						return 0 // break
					}

				} else if value != "" {
					chatMessage.New().
						Map(map[string]interface{}{
							"text": value,
							"done": true,
						}).
						Write(c.Writer)
				}

				done <- true
				return 0 // break
			}

			return 1 // continue
		}
	})
}

// saveChatHistory saves the chat history if storage is available
func (ast *Assistant) saveChatHistory(ctx chatctx.Context, messages []message.Message, content *message.Content) {
	if len(content.Bytes) > 0 && ctx.Sid != "" && len(messages) > 0 {
		storage.SaveHistory(
			ctx.Sid,
			[]map[string]interface{}{
				{"role": "user", "content": messages[len(messages)-1].Content(), "name": ctx.Sid},
				{"role": "assistant", "content": content.String(), "name": ctx.Sid},
			},
			ctx.ChatID,
			nil,
		)
	}
}

func (ast *Assistant) withOptions(options map[string]interface{}) map[string]interface{} {
	if options == nil {
		options = map[string]interface{}{}
	}

	if ast.Options != nil {
		for key, value := range ast.Options {
			options[key] = value
		}
	}

	// Add functions
	if ast.Functions != nil {
		options["tools"] = ast.Functions
		if options["tool_choice"] == nil {
			options["tool_choice"] = "auto"
		}
	}

	return options
}

func (ast *Assistant) withPrompts(messages []message.Message) []message.Message {
	if ast.Prompts != nil {
		for _, prompt := range ast.Prompts {
			name := ast.Name
			if prompt.Name != "" {
				name = prompt.Name
			}
			messages = append(messages, *message.New().Map(map[string]interface{}{"role": prompt.Role, "content": prompt.Content, "name": name}))
		}
	}
	return messages
}

func (ast *Assistant) withHistory(ctx chatctx.Context, input string) ([]message.Message, error) {
	messages := []message.Message{}
	messages = ast.withPrompts(messages)
	if storage != nil {
		history, err := storage.GetHistory(ctx.Sid, ctx.ChatID)
		if err != nil {
			return nil, err
		}

		// Add history messages
		for _, h := range history {
			messages = append(messages, *message.New().Map(h))
		}
	}

	// Add user message
	messages = append(messages, *message.New().Map(map[string]interface{}{"role": "user", "content": input, "name": ctx.Sid}))
	return messages, nil
}

// Chat implements the chat functionality
func (ast *Assistant) Chat(ctx context.Context, messages []message.Message, option map[string]interface{}, cb func(data []byte) int) error {
	if ast.openai == nil {
		return fmt.Errorf("openai is not initialized")
	}

	requestMessages, err := ast.requestMessages(ctx, messages)
	if err != nil {
		return fmt.Errorf("request messages error: %s", err.Error())
	}

	_, ext := ast.openai.ChatCompletionsWith(ctx, requestMessages, option, cb)
	if ext != nil {
		return fmt.Errorf("openai chat completions with error: %s", ext.Message)
	}

	return nil
}

func (ast *Assistant) requestMessages(ctx context.Context, messages []message.Message) ([]map[string]interface{}, error) {
	newMessages := []map[string]interface{}{}
	// With Prompts
	if ast.Prompts != nil {
		for _, prompt := range ast.Prompts {
			msg := map[string]interface{}{
				"role":    prompt.Role,
				"content": prompt.Content,
			}

			name := ast.Name
			if prompt.Name != "" {
				name = prompt.Name
			}

			msg["name"] = name
			newMessages = append(newMessages, msg)
		}
	}

	length := len(messages)
	for index, message := range messages {
		role := message.Role
		if role == "" {
			continue
			// return nil, fmt.Errorf("role must be string")
		}

		content := message.Text
		if content == "" {
			continue
			// return nil, fmt.Errorf("content must be string")
		}

		newMessage := map[string]interface{}{
			"role":    role,
			"content": content,
		}

		if name := message.Name; name != "" {
			newMessage["name"] = name
		}

		// Special handling for user messages with JSON content last message
		if role == "user" && index == length-1 {
			content = strings.TrimSpace(content)
			msg, err := chatMessage.NewString(content)
			if err != nil {
				return nil, fmt.Errorf("new string error: %s", err.Error())
			}

			newMessage["content"] = msg.Text
			if message.Attachments != nil {
				content, err := ast.withAttachments(ctx, &message)
				if err != nil {
					return nil, fmt.Errorf("with attachments error: %s", err.Error())
				}
				newMessage["content"] = content
			}
		}

		newMessages = append(newMessages, newMessage)
	}
	return newMessages, nil
}

func (ast *Assistant) withAttachments(ctx context.Context, msg *chatMessage.Message) ([]map[string]interface{}, error) {
	contents := []map[string]interface{}{{"type": "text", "text": msg.Text}}
	images := []string{}
	for _, attachment := range msg.Attachments {
		if strings.HasPrefix(attachment.ContentType, "image/") {
			images = append(images, attachment.FileID)
		}
	}

	if len(images) == 0 {
		return contents, nil
	}

	for _, image := range images {
		bytes64, err := ast.ReadBase64(ctx, image)
		if err != nil {
			return nil, fmt.Errorf("read base64 error: %s", err.Error())
		}

		contents = append(contents, map[string]interface{}{
			"type": "image_url",
			"image_url": map[string]string{
				"url": fmt.Sprintf("data:image/jpeg;base64,%s", bytes64),
			},
		})
	}

	return contents, nil
}

// ReadBase64 implements base64 file reading functionality
func (ast *Assistant) ReadBase64(ctx context.Context, fileID string) (string, error) {
	data, err := fs.Get("data")
	if err != nil {
		return "", fmt.Errorf("get filesystem error: %s", err.Error())
	}

	exists, err := data.Exists(fileID)
	if err != nil {
		return "", fmt.Errorf("check file error: %s", err.Error())
	}
	if !exists {
		return "", fmt.Errorf("file %s not found", fileID)
	}

	content, err := data.ReadFile(fileID)
	if err != nil {
		return "", fmt.Errorf("read file error: %s", err.Error())
	}

	return base64.StdEncoding.EncodeToString(content), nil
}
