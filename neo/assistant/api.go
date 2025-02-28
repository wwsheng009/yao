package assistant

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/yaoapp/gou/fs"
	"github.com/yaoapp/kun/log"
	chatctx "github.com/yaoapp/yao/neo/context"
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
func (ast *Assistant) Execute(c *gin.Context, ctx chatctx.Context, input interface{}, options map[string]interface{}, callback ...interface{}) error {
	contents := chatMessage.NewContents()
	messages, err := ast.withHistory(ctx, input)
	if err != nil {
		return err
	}
	return ast.execute(c, ctx, messages, options, contents, callback...)
}

// Execute implements the execute functionality
func (ast *Assistant) execute(c *gin.Context, ctx chatctx.Context, userInput interface{}, userOptions map[string]interface{}, contents *chatMessage.Contents, callback ...interface{}) error {

	var input []chatMessage.Message

	switch v := userInput.(type) {
	case string:
		input = []chatMessage.Message{{Role: "user", Text: v}}

	case []interface{}:
		raw, err := jsoniter.Marshal(v)
		if err != nil {
			return fmt.Errorf("marshal input error: %s", err.Error())
		}
		err = jsoniter.Unmarshal(raw, &input)
		if err != nil {
			return fmt.Errorf("unmarshal input error: %s", err.Error())
		}

	case []chatMessage.Message:
		input = v
	}

	if contents == nil {
		contents = chatMessage.NewContents()
	}
	options := ast.withOptions(userOptions)

	// Add RAG and Version support
	ctx.RAG = rag != nil
	ctx.Version = ast.vision

	// Run init hook
	res, err := ast.HookCreate(c, ctx, input, options, contents)
	if err != nil {
		chatMessage.New().
			Assistant(ast.ID, ast.Name, ast.Avatar).
			Error(err).
			Done().
			Write(c.Writer)
		return err
	}
	refAst := &ast

	// Update options if provided
	if res != nil && res.Options != nil {
		options = res.Options
	}

	// messages
	if res != nil && res.Input != nil {
		input = res.Input
	}

	// Handle next action
	// It's not used, return the new assistant_id and chat_id
	// if res != nil && res.Next != nil {
	// 	return res.Next.Execute(c, ctx, contents)
	// }

	// Switch to the new assistant if necessary
	if res != nil && res.AssistantID != "" && res.AssistantID != ctx.AssistantID {
		newAst, err := Get(res.AssistantID)
		if err != nil {
			chatMessage.New().
				Assistant(ast.ID, ast.Name, ast.Avatar).
				Error(err).
				Done().
				Write(c.Writer)
			return err
		}
		refAst = &newAst

		// Reset Message Contents
		last := input[len(input)-1]
		input, err = newAst.withHistory(ctx, last)
		if err != nil {
			return err
		}

		// Reset options
		options = newAst.withOptions(userOptions)

		// Update options if provided
		if res.Options != nil {
			options = res.Options
		}

		// Update assistant id
		ctx.AssistantID = res.AssistantID
		return newAst.handleChatStream(c, ctx, input, options, contents, callback...)
	}

	// Only proceed with chat stream if no specific next action was handled
	return (*refAst).handleChatStream(c, ctx, input, options, contents, callback...)
}

// Execute the next action
func (next *NextAction) Execute(c *gin.Context, ctx chatctx.Context, contents *chatMessage.Contents, callback ...interface{}) error {
	switch next.Action {

	// It's not used, because the process could be executed in the hook script
	// It may remove in the future
	// case "process":
	// 	if next.Payload == nil {
	// 		return fmt.Errorf("payload is required")
	// 	}

	// 	name, ok := next.Payload["name"].(string)
	// 	if !ok {
	// 		return fmt.Errorf("process name should be string")
	// 	}

	// 	args := []interface{}{}
	// 	if v, ok := next.Payload["args"].([]interface{}); ok {
	// 		args = v
	// 	}

	// 	// Add context and writer to args
	// 	args = append(args, ctx, c.Writer)
	// 	p, err := process.Of(name, args...)
	// 	if err != nil {
	// 		return fmt.Errorf("get process error: %s", err.Error())
	// 	}

	// 	err = p.Execute()
	// 	if err != nil {
	// 		return fmt.Errorf("execute process error: %s", err.Error())
	// 	}
	// 	defer p.Release()

	// 	return nil

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
		input := chatMessage.Message{}
		_, has := next.Payload["input"]
		if !has {
			return fmt.Errorf("input is required")
		}

		// Retry mode
		retry := false
		_, has = next.Payload["retry"]
		if has {
			retry = next.Payload["retry"].(bool)
			ctx.Retry = retry
		}

		switch v := next.Payload["input"].(type) {
		case string:
			messages := chatMessage.Message{}
			err := jsoniter.UnmarshalFromString(v, &messages)
			if err != nil {
				return fmt.Errorf("unmarshal input error: %s", err.Error())
			}
			input = messages

		case map[string]interface{}:
			msg, err := chatMessage.NewMap(v)
			if err != nil {
				return fmt.Errorf("unmarshal input error: %s", err.Error())
			}
			input = *msg

		case *chatMessage.Message:
			input = *v

		case chatMessage.Message:
			input = v

		default:
			return fmt.Errorf("input should be string or []chatMessage.Message")
		}

		// Options
		options := map[string]interface{}{}
		if v, ok := next.Payload["options"].(map[string]interface{}); ok {
			options = v
		}

		input.Hidden = true                    // not show in the history
		if input.Name == "" && ctx.Sid != "" { // add user id to the input
			input.Name = ctx.Sid
		}

		messages, err := assistant.withHistory(ctx, input)
		if err != nil {
			return fmt.Errorf("with history error: %s", err.Error())
		}

		// Create a new Text
		// Send loading message and mark as new
		if !ctx.Silent {
			msg := chatMessage.New().Map(map[string]interface{}{
				"new":   true,
				"role":  "assistant",
				"type":  "loading",
				"props": map[string]interface{}{"placeholder": "Calling " + assistant.Name},
			})
			msg.Assistant(assistant.ID, assistant.Name, assistant.Avatar)
			msg.Write(c.Writer)
		}
		newContents := chatMessage.NewContents()

		// Update the context id
		ctx.AssistantID = assistant.ID
		return assistant.execute(c, ctx, messages, options, newContents, callback...)

	case "exit":
		return nil

	default:
		return fmt.Errorf("unknown action: %s", next.Action)
	}
}

// GetPlaceholder returns the placeholder of the assistant
func (ast *Assistant) GetPlaceholder() *Placeholder {
	return ast.Placeholder
}

// Call implements the call functionality
func (ast *Assistant) Call(c *gin.Context, payload APIPayload) (interface{}, error) {
	scriptCtx, err := ast.Script.NewContext(payload.Sid, nil)
	if err != nil {
		return nil, err
	}
	defer scriptCtx.Close()
	ctx := c.Request.Context()

	method := fmt.Sprintf("%sAPI", payload.Name)

	// Check if the method exists
	if !scriptCtx.Global().Has(method) {
		return nil, fmt.Errorf(HookErrorMethodNotFound)
	}

	return scriptCtx.CallWith(ctx, method, payload.Payload)
}

// handleChatStream manages the streaming chat interaction with the AI
func (ast *Assistant) handleChatStream(c *gin.Context, ctx chatctx.Context, messages []chatMessage.Message, options map[string]interface{}, contents *chatMessage.Contents, callback ...interface{}) error {
	clientBreak := make(chan bool, 1)
	done := make(chan bool, 1)

	// Chat with AI in background
	go func() {
		err := ast.streamChat(c, ctx, messages, options, clientBreak, done, contents, false, callback...)
		if err != nil {
			chatMessage.New().Error(err).Done().Write(c.Writer)
		}
		count := 0
		for {
			if len(contents.Data) == 0 || contents.Current < 0 || contents.Current > (len(contents.Data)-1) {
				break
			}
			currentLine := contents.Data[contents.Current]

			if currentLine.Type != "tool" ||
				currentLine.Props["result"] == nil {
				break
			}
			result, ok := currentLine.Props["result"].(string)
			if !ok {
				break
			}
			id, ok := currentLine.Props["id"].(string)
			if !ok {
				break
			}
			fname, ok := currentLine.Props["function"].(string)
			if !ok {
				break
			}
			// doubao function calling model call
			if currentLine.Props["arguments"] != nil &&
				currentLine.Props["tool_calls_native"] != nil {

				args := ""
				arguments, ok := currentLine.Props["arguments"]
				if !ok {
					break
				}
				if args, ok = arguments.(string); !ok {
					args, err = jsoniter.MarshalToString(arguments)
					if err != nil {
						chatMessage.New().Error(err).Done().Write(c.Writer)
						break
					}
				}

				var asstMmsg chatMessage.Message
				asstMmsg.Role = "assistant"
				asstMmsg.Name = fname
				asstMmsg.Text = ""
				asstMmsg.ToolCalls = []chatMessage.FunctionCall{
					{
						Index: 0,
						ID:    id,
						Type:  "function",
						Function: chatMessage.FCAttributes{
							Name:      fname,
							Arguments: args,
						},
					},
				}
				asstMmsg.Hidden = true
				messages = append(messages, asstMmsg)
				var msg chatMessage.Message
				msg.Role = "tool"
				msg.ToolCallId = id
				msg.Text = result
				msg.Hidden = true
				messages = append(messages, msg)
			} else {
				// deepseek like call
				var msg chatMessage.Message

				msg.Role = "assistant"
				msg.Text = contents.JSON()
				// msg.Hidden = true 不需要隐藏，作为历史记录给用户参考，并不会再次发送给AI

				messages = append(messages, msg)
				msg.Role = "user"
				msg.Text = "function call [" + id + "] result :" + result + "\""
				msg.Hidden = true //作为中间结果，发送给AI后，不需要显示给用户
				messages = append(messages, msg)
			}
			contents = chatMessage.NewContents()
			err := ast.streamChat(c, ctx, messages, options, clientBreak, done, contents, true, callback...)
			if err != nil {
				chatMessage.New().Error(err).Done().Write(c.Writer)
				break
			}
			count++

			if count > 5 {
				contents = chatMessage.NewContents()
				contents.AppendText([]byte("Too many function calls"))
				err = fmt.Errorf("too many function calls")
				chatMessage.New().Error(err).Done().Write(c.Writer)
				break
			}
		}
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
	messages []chatMessage.Message,
	options map[string]interface{},
	clientBreak chan bool,
	done chan bool,
	contents *chatMessage.Contents,
	recall bool,
	callback ...interface{},
) error {

	var cb interface{}
	if len(callback) > 0 {
		cb = callback[0]
	}

	errorRaw := ""
	isFirst := true
	isFirstThink := true
	isThinking := false

	toolsCount := 0
	currentMessageID := ""
	isRecall := recall
	err := ast.Chat(c.Request.Context(), messages, options, func(data []byte) int {
		select {
		case <-clientBreak:
			return 0 // break

		default:
			msg := chatMessage.NewOpenAI(data, isThinking)
			if msg == nil {
				return 1 // continue
			}

			if msg.Pending {
				errorRaw += msg.Text
				return 1 // continue
			}

			// Retry mode
			msg.Retry = ctx.Retry   // Retry mode
			msg.Silent = ctx.Silent // Silent mode

			// Handle error
			if msg.Type == "error" {
				value := msg.String()
				res, hookErr := ast.HookFail(c, ctx, messages, fmt.Errorf("%s", value), contents)
				if hookErr == nil && res != nil && (res.Output != "" || res.Error != "") {
					value = res.Output
					if res.Error != "" {
						value = res.Error
					}
				}
				newMsg := chatMessage.New().Error(value).Done()
				newMsg.Retry = ctx.Retry
				newMsg.Silent = ctx.Silent
				newMsg.Callback(cb).Write(c.Writer)
				return 0 // break
			}

			// for api reasoning_content response
			if msg.Type == "think" {
				if isFirstThink {
					msg.Text = "<think>\n" + msg.Text // add the think begin tag
					isFirstThink = false
					isThinking = true
				}
			}

			// for api reasoning_content response
			if isThinking && msg.Type != "think" {
				// add the think close tag
				end := chatMessage.New().Map(map[string]interface{}{"text": "\n</think>\n", "type": "think", "delta": true})
				end.ID = currentMessageID
				end.Retry = ctx.Retry
				end.Silent = ctx.Silent

				end.Callback(cb).Write(c.Writer)
				end.AppendTo(contents)
				contents.UpdateType("think", map[string]interface{}{"text": contents.Text()}, currentMessageID)
				isThinking = false

				// Clear the token and make a new line
				contents.NewText([]byte{}, currentMessageID)
				contents.ClearToken()
			}

			// for native tool_calls response, keep the first tool_calls_native message
			if msg.Type == "tool_calls_native" {

				if toolsCount > 1 {
					msg.Text = "" // clear the text
					msg.Type = "text"
					msg.IsNew = false
					return 1 // continue
				}

				if msg.IsBeginTool {

					if toolsCount == 1 {
						msg.IsNew = false
						msg.Text = "\n</tool>\n" // add the tool_calls close tag
					}

					if toolsCount == 0 {
						msg.Text = "\n<tool>\n" + msg.Text // add the tool_calls begin tag
					}

					toolsCount++

				}

				if msg.IsEndTool {
					msg.Text = msg.Text + "\n</tool>\n" // add the tool_calls close tag
				}
			}

			delta := msg.String()
			// Chunk the delta
			if delta != "" {

				msg.AppendTo(contents) // Append content and send message

				// Scan the tokens
				callback := func(token string, id string, begin bool, text string, tails string) {
					currentMessageID = id
					msg.ID = id
					msg.Type = token
					msg.Text = "" // clear the text
					if msg.Props == nil {
						msg.Props = map[string]interface{}{"text": text} // Update props
					} else {
						msg.Props["text"] = text
					}

					// End of the token clear the text
					if begin {
						return
					}

					// New message with the tails
					if tails != "" {
						newMsg, err := chatMessage.NewString(tails, id)
						if err != nil {
							return
						}
						messages = append(messages, *newMsg)
					}
				}
				contents.ScanTokens(currentMessageID, callback)

				// if isTool && msg.IsDone {
				// 	contents.ScanTokens(currentMessageID, callback)
				// 	isTool = false
				// }

				// Handle stream
				// The stream hook is not used, because there's no need to handle the stream output
				// if some thing need to be handled in future, we can use the stream hook again
				// ------------------------------------------------------------------------------
				// res, err := ast.HookStream(c, ctx, messages, msg, contents)
				// if err == nil && res != nil {

				// 	if res.Next != nil {
				// 		err = res.Next.Execute(c, ctx, contents)
				// 		if err != nil {
				// 			chatMessage.New().Error(err.Error()).Done().Write(c.Writer)
				// 		}

				// 		done <- true
				// 		return 0 // break
				// 	}

				// 	if res.Silent {
				// 		return 1 // continue
				// 	}
				// }
				// ------------------------------------------------------------------------------

				// Write the message to the stream
				msgType := msg.Type
				if msgType == "tool_calls_native" {
					msgType = "tool"
				}

				output := chatMessage.New().Map(map[string]interface{}{
					"text":  delta,
					"type":  msgType,
					"new":   isRecall,
					"done":  msg.IsDone && msg.Type != "tool",
					"delta": true,
				})
				isRecall = false

				output.Retry = ctx.Retry   // Retry mode
				output.Silent = ctx.Silent // Silent mode
				if isFirst {
					output.Assistant(ast.ID, ast.Name, ast.Avatar)
					isFirst = false
				}
				output.Callback(cb).Write(c.Writer)
			}

			// Complete the stream
			if msg.IsDone {

				// Send the last message to the client
				if delta != "" && msg.Type != "tool" {
					chatMessage.New().
						Map(map[string]interface{}{
							"assistant_id":     ast.ID,
							"assistant_name":   ast.Name,
							"assistant_avatar": ast.Avatar,
							"text":             delta,
							"type":             "text",
							"delta":            true,
							"done":             true,
							"retry":            ctx.Retry,
							"silent":           ctx.Silent,
						}).
						Callback(cb).
						Write(c.Writer)
				}

				// Remove the last empty data
				contents.RemoveLastEmpty()
				res, hookErr := ast.HookDone(c, ctx, messages, contents)

				// Some error occurred in the hook, return the error
				if hookErr != nil {
					chatMessage.New().Error(hookErr.Error()).Done().Callback(cb).Write(c.Writer)

					done <- true
					return 0 // break
				}
				ast.saveChatHistory(ctx, messages, contents)

				if len(res.Output) > 0 {
					if contents.Data[contents.Current].Type == "tool" {
						contents.Data[contents.Current].Props["result"] = string(res.Output[len(res.Output)-1].Bytes)
						return 0
					}
				}
				// If the hook is successful, execute the next action
				if res != nil && res.Next != nil {
					err := res.Next.Execute(c, ctx, contents, cb)
					if err != nil {
						chatMessage.New().Error(err.Error()).Done().Callback(cb).Write(c.Writer)
					}
					done <- true
					return 0 // break
				}

				// The default output
				output := chatMessage.New().Done()
				if res != nil && res.Output != nil {
					output = chatMessage.New().Map(map[string]interface{}{"text": res.Output, "done": true})
					output.Retry = ctx.Retry
					output.Silent = ctx.Silent
				}

				// has result
				if res != nil && res.Result != nil && cb != nil {
					output.Result = res.Result // Add the result to the output  message
				}

				output.Callback(cb).Write(c.Writer)
				done <- true
				return 0 // break
			}

			return 1 // continue
		}
	})

	// Handle error
	if err != nil {
		return err
	}

	// raw error
	if errorRaw != "" {
		msg, err := chatMessage.NewStringError(errorRaw)
		if err != nil {
			return fmt.Errorf("error: %s", err.Error())
		}
		msg.Retry = ctx.Retry
		msg.Silent = ctx.Silent
		msg.Done().Callback(cb).Write(c.Writer)
	}

	return nil
}

// saveChatHistory saves the chat history if storage is available
func (ast *Assistant) saveChatHistory(ctx chatctx.Context, messages []chatMessage.Message, contents *chatMessage.Contents) {
	if len(contents.Data) > 0 && ctx.Sid != "" && len(messages) > 0 {
		userMessage := messages[len(messages)-1]
		data := []map[string]interface{}{
			{
				"role":    "user",
				"content": userMessage.Content(),
				"name":    ctx.Sid,
			},
			{
				"role":             "assistant",
				"content":          contents.JSON(),
				"name":             ast.ID,
				"assistant_id":     ast.ID,
				"assistant_name":   ast.Name,
				"assistant_avatar": ast.Avatar,
			},
		}

		// if the user message is hidden, just save the assistant message
		if userMessage.Hidden {
			data = []map[string]interface{}{data[1]}
		}
		// if v, ok := contents.Data[contents.Current].Props["function"]; ok && v != "" {
		// 	data = []map[string]interface{}{data[0]}
		// }

		err := storage.SaveHistory(ctx.Sid, data, ctx.ChatID, ctx.Map())
		if err != nil {
			log.Error("save neo history with error:%s", err.Error())
		}
	}
}

func (ast *Assistant) withOptions(options map[string]interface{}) map[string]interface{} {
	if options == nil {
		options = map[string]interface{}{}
	}

	// Add Custom Options
	if ast.Options != nil {
		for key, value := range ast.Options {
			options[key] = value
		}
	}

	// Add tool_calls
	if ast.Tools != nil && ast.Tools.Tools != nil && len(ast.Tools.Tools) > 0 {
		if settings, has := connectorSettings[ast.Connector]; has && settings.Tools {
			options["tools"] = ast.Tools.Tools
			if options["tool_choice"] == nil {
				options["tool_choice"] = "auto"
			}
		}
	}

	return options
}

func (ast *Assistant) withPrompts(messages_history []chatMessage.Message) []chatMessage.Message {
	messages := make([]chatMessage.Message, 0)
	if ast.Prompts != nil {
		for _, prompt := range ast.Prompts {
			name := strings.ReplaceAll(ast.ID, ".", "_") // OpenAI only supports underscore in the name
			if prompt.Name != "" {
				name = prompt.Name
			}
			messages = append(messages, *chatMessage.New().Map(map[string]interface{}{"role": prompt.Role, "content": prompt.Content, "name": name}))
		}
	}

	// Add tool_calls
	if ast.Tools != nil && ast.Tools.Tools != nil && len(ast.Tools.Tools) > 0 {
		settings, has := connectorSettings[ast.Connector]
		if !has || !settings.Tools {
			raw, _ := jsoniter.MarshalToString(ast.Tools.Tools)

			examples := []string{}
			for _, tool := range ast.Tools.Tools {
				example := tool.Example()
				examples = append(examples, example)
			}

			examplesStr := ""
			if len(examples) > 0 {
				examplesStr = "Examples:\n" + strings.Join(examples, "\n\n")
			}

			prompts := []map[string]interface{}{
				{
					"role": "system",
					"name": "duplicate function call check",
					"content": "If the user reply the previous function call with result with same id,\n" +
						"don't make the duplicate function call for same function id\n",
				},
				{
					"role":    "system",
					"name":    "TOOL_CALLS_SCHEMA",
					"content": raw,
				},
				{
					"role": "system",
					"name": "TOOL_CALLS_SCHEMA",
					"content": "## Tool Calls Schema Definition\n" +
						"Each tool call is defined with:\n" +
						"  - type: always 'function'\n" +
						"  - function:\n" +
						"    - id: the unique id of each call\n" +
						"    - name: function name\n" +
						"    - description: function description\n" +
						"    - parameters: function parameters with type and validation rules\n",
				},
				{
					"role": "system",
					"name": "TOOL_CALLS",
					"content": "## Tool Response Format\n" +
						"1. Only use tool calls when a function matches your task exactly\n" +
						"2. Each tool call must be wrapped in <tool> and </tool> tags\n" +
						"3. Tool call must be a valid JSON with:\n" +
						"   {\"id\": \"unique id of call\",\"function\": \"function_name\", \"arguments\": {parameters}}\n" +
						"4. Return the function's result as your response\n" +
						"5. One tool call per response\n" +
						"6. Arguments must match parameter types, rules and description\n" +
						"7. Create unique IDs for each tool call if the call is new\n\n" +
						examplesStr,
				},
				{
					"role": "system",
					"name": "TOOL_CALLS",
					"content": "## Tool Usage Guidelines\n" +
						"1. Use functions defined in TOOL_CALLS_SCHEMA only when they match your needs\n" +
						"2. If no matching function exists, respond normally as a helpful assistant\n" +
						"3. When using tools, arguments must match the schema definition exactly\n" +
						"4. All parameter values must strictly adhere to the validation rules specified in properties\n" +
						"5. Never skip or ignore any validation requirements defined in the schema",
				},
			}

			// Add tool_calls developer prompts
			if ast.Tools.Prompts != nil && len(ast.Tools.Prompts) > 0 {
				for _, prompt := range ast.Tools.Prompts {
					messages = append(messages, *chatMessage.New().Map(map[string]interface{}{
						"role":    prompt.Role,
						"content": prompt.Content,
						"name":    prompt.Name,
					}))
				}
			}

			// Add the prompts
			for _, prompt := range prompts {
				messages = append(messages, *chatMessage.New().Map(prompt))
			}

		}
	}
	messages = append(messages, messages_history...)
	return messages
}

func (ast *Assistant) withHistory(ctx chatctx.Context, input interface{}) ([]chatMessage.Message, error) {

	var userMessage *chatMessage.Message
	var inputMessages []*chatMessage.Message
	switch v := input.(type) {
	case string:
		userMessage = chatMessage.New().Map(map[string]interface{}{"role": "user", "content": v})

	case map[string]interface{}:
		userMessage = chatMessage.New().Map(v)

	case []interface{}:
		raw, err := jsoniter.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshal input error: %s", err.Error())
		}
		err = jsoniter.Unmarshal(raw, &inputMessages)
		if err != nil {
			return nil, fmt.Errorf("unmarshal input error: %s", err.Error())
		}

	case chatMessage.Message:
		userMessage = &v
	case *chatMessage.Message:
		userMessage = v
	default:
		return nil, fmt.Errorf("unknown input type: %T", input)
	}

	messages := []chatMessage.Message{}
	if storage != nil {
		history, err := storage.GetHistory(ctx.Sid, ctx.ChatID)
		if err != nil {
			return nil, err
		}

		// Add history messages
		for _, h := range history {
			msgs, err := chatMessage.NewHistory(h)
			if err != nil {
				return nil, err
			}
			messages = append(messages, msgs...)
		}
	}

	// Add system prompts
	messages = ast.withPrompts(messages)

	// Add user message
	if userMessage != nil {
		messages = append(messages, *userMessage)
	}

	// Add input messages
	if len(inputMessages) > 0 {
		for _, msg := range inputMessages {
			if msg == nil || msg.Role == "" {
				continue
			}
			messages = append(messages, *msg)
		}
	}
	return messages, nil
}

// Chat implements the chat functionality
func (ast *Assistant) Chat(ctx context.Context, messages []chatMessage.Message, option map[string]interface{}, cb func(data []byte) int) error {
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

// formatMessages processes messages to ensure they meet the required standards:
// 1. Filters out duplicate messages with identical content, role, and name
// 2. Moves system messages to the beginning while preserving the order of other messages
// 3. Ensures the first non-system message is a user message (removes leading assistant messages)
// 4. Ensures the last message is a user message (removes trailing assistant messages)
// 5. Merges consecutive assistant messages from the same assistant
func formatMessages(messages []map[string]interface{}) []map[string]interface{} {
	// Filter out duplicate messages with identical content, role, and name
	filteredMessages := []map[string]interface{}{}
	seen := make(map[string]bool)

	for _, msg := range messages {
		// Create a unique key for each message based on role, content, and name
		role := msg["role"].(string)
		content := fmt.Sprintf("%v", msg["content"]) // Convert to string regardless of type

		// Get name if it exists
		name := ""
		if nameVal, exists := msg["name"]; exists {
			name = fmt.Sprintf("%v", nameVal)
		}

		// Create a unique key for this message
		key := fmt.Sprintf("%s:%s:%s", role, content, name)

		// If we haven't seen this message before, add it to filtered messages
		if !seen[key] {
			filteredMessages = append(filteredMessages, msg)
			seen[key] = true
		}
	}

	// Separate system messages while preserving the order of other messages
	systemMessages := []map[string]interface{}{}
	otherMessages := []map[string]interface{}{}

	for _, msg := range filteredMessages {
		if msg["role"].(string) == "system" {
			systemMessages = append(systemMessages, msg)
		} else {
			otherMessages = append(otherMessages, msg)
		}
	}

	// Ensure the first non-system message is a user message
	// If there are no user messages or the first message is not a user message, remove leading assistant messages
	validOtherMessages := []map[string]interface{}{}
	foundUserMessage := false

	for _, msg := range otherMessages {
		if msg["role"].(string) == "user" {
			foundUserMessage = true
			validOtherMessages = append(validOtherMessages, msg)
		} else if foundUserMessage {
			// Only keep assistant messages that come after a user message
			validOtherMessages = append(validOtherMessages, msg)
		}
		// Skip assistant messages that come before any user message
	}

	// If no valid messages remain, return just the system messages
	if len(validOtherMessages) == 0 {
		return systemMessages
	}

	// Ensure the last message is a user message
	// Remove any trailing assistant messages
	lastUserIndex := -1
	for i := len(validOtherMessages) - 1; i >= 0; i-- {
		if validOtherMessages[i]["role"].(string) == "user" {
			lastUserIndex = i
			break
		}
	}

	// If we found a user message, trim any assistant messages after it
	if lastUserIndex >= 0 && lastUserIndex < len(validOtherMessages)-1 {
		validOtherMessages = validOtherMessages[:lastUserIndex+1]
	}

	// If there are no user messages left after filtering, return just the system messages
	if len(validOtherMessages) == 0 {
		return systemMessages
	}

	// Combine system messages first, followed by other valid messages in their original order
	orderedMessages := append(systemMessages, validOtherMessages...)

	// Merge consecutive assistant messages
	mergedMessages := []map[string]interface{}{}
	var lastMessage map[string]interface{}

	for _, msg := range orderedMessages {
		// If this is the first message, just add it
		if lastMessage == nil {
			mergedMessages = append(mergedMessages, msg)
			lastMessage = msg
			continue
		}

		// If both current and last messages are from assistant, check if they can be merged
		if msg["role"].(string) == "assistant" && lastMessage["role"].(string) == "assistant" {
			// Get name information
			nameVal, hasName := msg["name"]

			// Prepare name prefix for the content
			namePrefix := ""
			if hasName {
				namePrefix = fmt.Sprintf("[%v]: ", nameVal)
			}

			// Merge the content, including name information if available
			lastContent := fmt.Sprintf("%v", lastMessage["content"])
			content := fmt.Sprintf("%v", msg["content"])

			// Add the name prefix to the content
			if namePrefix != "" {
				content = namePrefix + content
			}

			// Merge the messages
			lastMessage["content"] = lastContent + "\n" + content
			continue
		}

		// If we can't merge, add as a new message
		mergedMessages = append(mergedMessages, msg)
		lastMessage = msg
	}

	return mergedMessages
}

func (ast *Assistant) requestMessages(ctx context.Context, messages []chatMessage.Message) ([]map[string]interface{}, error) {
	newMessages := []map[string]interface{}{}
	length := len(messages)

	for index, message := range messages {
		// Ignore the tool, think, error
		if (message.Type == "tool" && len(message.ToolCalls) == 0) || message.Type == "think" || message.Type == "error" {
			continue
		}

		role := message.Role
		if role == "" {
			continue
			// return nil, fmt.Errorf("role must be string")
		}

		content := message.String()
		// if content == "" {
		// 	return nil, fmt.Errorf("content must be string")
		// }

		newMessage := map[string]interface{}{
			"role": role,
		}
		if role == "tool" {
			newMessage["tool_calls"] = message.ToolCalls
			newMessage["tool_call_id"] = message.ToolCallId
		} else {
			newMessage["content"] = content
		}

		// Keep the name for user messages
		if name := message.Name; name != "" {
			if role != "system" {
				newMessage["name"] = stringHash(name)
			} else {
				newMessage["name"] = name
			}
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
				contents, err := ast.withAttachments(ctx, &message)
				if err != nil {
					return nil, fmt.Errorf("with attachments error: %s", err.Error())
				}

				// if current assistant is vision capable, add the contents directly
				if ast.vision {
					newMessage["content"] = contents
					continue
				}

				// If current assistant is not vision capable, add the description of the image
				if contents != nil {
					for _, content := range contents {
						newMessages = append(newMessages, content)
					}
				}
			}
		}

		newMessages = append(newMessages, newMessage)
	}
	newMessages = MergeMessages(newMessages)

	// Process messages to standardize format, filter duplicates, and merge consecutive assistant messages
	processedMessages := formatMessages(newMessages)

	// For debug environment, print the request messages
	if os.Getenv("YAO_AGENT_PRINT_REQUEST_MESSAGES") == "true" {
		for _, message := range processedMessages {
			raw, _ := jsoniter.MarshalToString(message)
			log.Trace("[Request Message] %s", raw)
		}
	}

	return processedMessages, nil
}

// MergeMessages merges adjacent messages with the same role and moves system messages to the front
func MergeMessages(messages []map[string]interface{}) []map[string]interface{} {
	if len(messages) == 0 {
		return messages
	}

	// Separate system messages and non-system messages
	systemMsgs := []map[string]interface{}{}
	otherMsgs := []map[string]interface{}{}

	for _, msg := range messages {
		if msg["role"] == "system" {
			systemMsgs = append(systemMsgs, msg)
		} else {
			otherMsgs = append(otherMsgs, msg)
		}
	}

	// Merge adjacent messages in the non-system messages
	merged := []map[string]interface{}{}
	if len(otherMsgs) > 0 {
		merged = append(merged, otherMsgs[0])
		for i := 1; i < len(otherMsgs); i++ {
			last := merged[len(merged)-1]
			current := otherMsgs[i]

			if last["role"] == current["role"] {
				// Merge content of messages with same role
				last["content"] = fmt.Sprintf("%v\n%v", last["content"], current["content"])
			} else {
				merged = append(merged, current)
			}
		}
	}

	// Combine system messages and other messages
	result := make([]map[string]interface{}, 0, len(systemMsgs)+len(merged))
	result = append(result, systemMsgs...)
	result = append(result, merged...)

	return result
}

func (ast *Assistant) withAttachments(ctx context.Context, msg *chatMessage.Message) ([]map[string]interface{}, error) {
	contents := []map[string]interface{}{{"type": "text", "text": msg.Text}}
	if !ast.vision {
		contents = []map[string]interface{}{{"role": "user", "content": msg.Text}}
	}

	images := []string{}
	for _, attachment := range msg.Attachments {
		if strings.HasPrefix(attachment.ContentType, "image/") {
			if ast.vision {
				images = append(images, attachment.URL)
				continue
			}

			// If the current assistant is not vision capable, add the description of the image
			raw, err := jsoniter.MarshalToString(attachment)
			if err != nil {
				return nil, fmt.Errorf("marshal attachment error: %s", err.Error())
			}
			contents = append(contents, map[string]interface{}{
				"role":    "system",
				"content": raw,
			})
		}
	}

	if len(images) == 0 {
		return contents, nil
	}

	// If the current assistant is vision capable, add the image to the contents directly
	if ast.vision {
		for _, url := range images {

			// If the image is already a URL, add it directly
			if strings.HasPrefix(url, "http") {
				contents = append(contents, map[string]interface{}{
					"type": "image_url",
					"image_url": map[string]string{
						"url": url,
					},
				})
				continue
			}

			// Read base64
			bytes64, err := ast.ReadBase64(ctx, url)
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

	// If the current assistant is not vision capable, add the description of the image

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
