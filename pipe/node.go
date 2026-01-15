package pipe

import (
	"fmt"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/yaoapp/gou/process"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/openai"
	"github.com/yaoapp/yao/pipe/ui/cli"
)

// Case Execute the user input
func (node *Node) Case(ctx *Context, input Input) (any, error) {

	if node.Switch == nil || len(node.Switch) == 0 {
		return nil, node.Errorf(ctx, "switch case not found")
	}

	input, err := ctx.parseNodeInput(node, input)
	if err != nil {
		return nil, err
	}

	// Find the case
	var child *Pipe = node.Switch["default"]
	matched := "default"
	data := ctx.data(node)

	for expr, pip := range node.Switch {

		expr, err := data.replaceString(expr)
		if err != nil {
			return nil, err
		}

		v, err := data.Exec(expr)
		if err != nil {
			return nil, err
		}

		if v == true {
			child = pip
			matched = expr
		}
	}

	if child == nil {
		log.Error("pipe: switch case not found pipe=%s(%s) ctx=%s node=%s", ctx.Name, ctx.Pipe.ID, ctx.id, node.Name)
		return nil, node.Errorf(ctx, "switch case not found")
	}

	log.Debug("pipe: switch matched pipe=%s(%s) ctx=%s node=%s case=%s child=%s(%s) hasNodes=%t goto=%s", ctx.Name, ctx.Pipe.ID, ctx.id, node.Name, matched, child.Name, child.ID, child.HasNodes(), child.Goto)

	// Execute the child pipe
	var res any = nil
	subctx := child.Create().inheritance(ctx)

	// Child pipe may be a "control pipe" (goto/input/output only, no nodes)
	if subctx.current == nil {
		log.Debug("pipe: switch control child pipe=%s(%s) ctx=%s parentCtx=%s", child.Name, child.ID, subctx.id, ctx.id)
		defer Close(subctx.id)

		// Apply child pipe input mapping
		mappedInput, err := subctx.parseInput(input)
		if err != nil {
			return nil, err
		}

		// Apply child pipe goto to parent context (evaluated with child data)
		if child.Goto != "" {
			next, err := subctx.data(nil).replaceString(child.Goto)
			if err != nil {
				return nil, err
			}
			ctx.gotoNext = next
			log.Debug("pipe: switch control goto parent pipe=%s(%s) ctx=%s to=%s", ctx.Name, ctx.Pipe.ID, ctx.id, next)
		}

		// Apply child pipe output mapping (fallback to mapped input)
		if child.Output != nil {
			res, err = subctx.parseOutput()
			if err != nil {
				return nil, err
			}
		} else {
			res = mappedInput
		}

	} else {
		log.Debug("pipe: switch exec child pipe=%s(%s) ctx=%s parentCtx=%s", child.Name, child.ID, subctx.id, ctx.id)
		res, err = subctx.Exec(input...)
		if err != nil {
			return nil, err
		}

		// If the child pipe defines a top-level goto, forward it to parent context
		if child.Goto != "" {
			next, err := subctx.data(nil).replaceString(child.Goto)
			if err != nil {
				return nil, err
			}
			ctx.gotoNext = next
			log.Debug("pipe: switch forward goto parent pipe=%s(%s) ctx=%s to=%s", ctx.Name, ctx.Pipe.ID, ctx.id, next)
		}
	}

	output, err := ctx.parseNodeOutput(node, res)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// YaoProcess Execute the Yao Process
func (node *Node) YaoProcess(ctx *Context, input Input) (any, error) {

	if node.Process == nil {
		return nil, node.Errorf(ctx, "process not set")
	}

	input, err := ctx.parseNodeInput(node, input)
	if err != nil {
		return nil, err
	}

	data := ctx.data(node)
	args, err := data.replaceArray(node.Process.Args)

	log.Debug("pipe: process pipe=%s(%s) ctx=%s node=%s name=%s argc=%d", ctx.Name, ctx.Pipe.ID, ctx.id, node.Name, node.Process.Name, len(args))

	// Execute the process
	process, err := process.Of(node.Process.Name, args...)
	if err != nil {
		return nil, node.Errorf(ctx, "%v", err)
	}

	res, err := process.WithGlobal(ctx.global).WithSID(ctx.sid).Exec()
	if err != nil {
		return nil, node.Errorf(ctx, "%v", err)
	}

	output, err := ctx.parseNodeOutput(node, res)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// AI Execute the AI input
func (node *Node) AI(ctx *Context, input Input) (any, error) {

	if node.Prompts == nil || len(node.Prompts) == 0 {
		log.Error("pipe: ai prompts not found pipe=%s(%s) ctx=%s node=%s", ctx.Name, ctx.Pipe.ID, ctx.id, node.Name)
		return nil, node.Errorf(ctx, "prompts not found")
	}

	input, err := ctx.parseNodeInput(node, input)
	if err != nil {
		return nil, err
	}

	data := ctx.data(node)
	prompts, err := data.replacePrompts(node.Prompts)
	if err != nil {
		return nil, err
	}
	prompts = node.aiMergeHistory(ctx, prompts)
	log.Debug("pipe: ai call pipe=%s(%s) ctx=%s node=%s model=%s prompts=%d", ctx.Name, ctx.Pipe.ID, ctx.id, node.Name, node.Model, len(prompts))

	res, err := node.chatCompletions(ctx, prompts, node.Options)
	if err != nil {
		return nil, err
	}

	output, err := ctx.parseNodeOutput(node, res)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (node *Node) chatCompletions(ctx *Context, prompts []Prompt, options map[string]interface{}) (any, error) {
	// moapi call
	ai, err := openai.NewMoapi(node.Model)
	if err != nil {
		return nil, err
	}

	response := []string{}
	content := []string{}
	_, ex := ai.ChatCompletions(promptsToMap(prompts), node.Options, func(data []byte) int {

		// Progress Hook
		if ctx.Hooks != nil && ctx.Hooks.Progress != "" {
			// best-effort hook, never break streaming
			proc, err := process.Of(ctx.Hooks.Progress, map[string]any{
				"pipe": map[string]any{"id": ctx.Pipe.ID, "name": ctx.Pipe.Name},
				"ctx":  ctx.id,
				"node": map[string]any{"name": node.Name, "type": node.Type},
				"data": string(data),
			})
			if err != nil {
				log.Error("pipe: hook progress create error pipe=%s(%s) ctx=%s node=%s err=%v", ctx.Name, ctx.Pipe.ID, ctx.id, node.Name, err)
			} else {
				_, herr := proc.WithGlobal(ctx.global).WithSID(ctx.sid).Exec()
				if herr != nil {
					log.Error("pipe: hook progress exec error pipe=%s(%s) ctx=%s node=%s err=%v", ctx.Name, ctx.Pipe.ID, ctx.id, node.Name, herr)
				}
			}
		}

		if len(data) > 5 && string(data[:5]) == "data:" {
			var res ChatCompletionChunk
			err := jsoniter.Unmarshal(data[5:], &res)
			if err != nil {
				return 0
			}
			if len(res.Choices) > 0 {
				response = append(response, res.Choices[0].Delta.Content)
			}
		} else {
			content = append(content, string(data))
		}

		return 1
	})

	if ex != nil {
		log.Error("pipe: ai error pipe=%s(%s) ctx=%s node=%s msg=%s", ctx.Name, ctx.Pipe.ID, ctx.id, node.Name, ex.Message)
		return nil, node.Errorf(ctx, "AI error: %s", ex.Message)
	}

	if (len(response) == 0) && (len(content) > 0) {
		msg := strings.Join(content, "")
		log.Error("pipe: ai empty stream pipe=%s(%s) ctx=%s node=%s msg=%s", ctx.Name, ctx.Pipe.ID, ctx.id, node.Name, msg)
		return nil, node.Errorf(ctx, "AI error: %s", msg)
	}

	raw := strings.Join(response, "")

	// try to parse the response
	var res any
	err = jsoniter.UnmarshalFromString(raw, &res)
	if err != nil {
		return raw, nil
	}

	return res, nil
}

func (node *Node) aiMergeHistory(ctx *Context, prompts []Prompt) []Prompt {
	if ctx.history == nil {
		ctx.history = map[*Node][]Prompt{}
	}
	if ctx.history[node] == nil {
		ctx.history = map[*Node][]Prompt{}
	}
	new := []Prompt{}
	saved := map[string]bool{}

	// filter the prompts
	for _, prompt := range ctx.history[node] {
		saved[prompt.finger()] = true
		new = append(new, prompt)
	}

	for _, prompt := range prompts {
		if saved[prompt.finger()] {
			continue
		}
		new = append(new, prompt)
	}

	// update the history
	ctx.history[node] = new
	return new
}

// Render Execute the user input
func (node *Node) Render(ctx *Context, input Input) (any, bool, error) {

	switch node.UI {

	case "cli":
		log.Debug("pipe: ui cli pipe=%s(%s) ctx=%s node=%s", ctx.Name, ctx.Pipe.ID, ctx.id, node.Name)
		output, err := node.renderCli(ctx, input)
		if err != nil {
			log.Error("pipe: ui cli error pipe=%s(%s) ctx=%s node=%s err=%v", ctx.Name, ctx.Pipe.ID, ctx.id, node.Name, err)
			return nil, false, err
		}
		return output, false, nil

	default:
		input, err := ctx.parseNodeInput(node, input)
		if err != nil {
			log.Error("pipe: ui parse input error pipe=%s(%s) ctx=%s node=%s err=%v", ctx.Name, ctx.Pipe.ID, ctx.id, node.Name, err)
			return nil, true, err
		}

		log.Debug("pipe: ui pause pipe=%s(%s) ctx=%s node=%s ui=%s", ctx.Name, ctx.Pipe.ID, ctx.id, node.Name, node.UI)
		return ResumeContext{
			ID:    ctx.id,
			Input: input,
			Node:  node,
			Data:  ctx.data(node),
			Type:  node.Type,
			UI:    node.UI,
		}, true, nil

	}
}

func (node *Node) renderCli(ctx *Context, input Input) (any, error) {
	input, err := ctx.parseNodeInput(node, input)
	if err != nil {
		return nil, err
	}

	// Set option
	data := ctx.data(node)
	label, err := data.replaceString(node.Label)
	if err != nil {
		return nil, err
	}

	option := &cli.Option{Label: label}
	if node.AutoFill != nil {

		value := fmt.Sprintf("%v", node.AutoFill.Value)
		value, err = data.replaceString(value)
		if value != "" {
			if err != nil {
				return nil, err
			}

			if node.AutoFill.Action == "exit" {
				value = fmt.Sprintf("%s\nexit()\n", value)
			}
			option.Reader = strings.NewReader(value)
		}
	}

	lines, err := cli.New(option).Render(input)
	if err != nil {
		return nil, node.Errorf(ctx, "%v", err)
	}

	output, err := ctx.parseNodeOutput(node, lines)
	if err != nil {
		return nil, node.Errorf(ctx, "%v", err)
	}
	return output, nil
}

// Errorf format the error message
func (node *Node) Errorf(ctx *Context, format string, a ...any) error {
	message := fmt.Sprintf(format, a...)
	pid := ctx.Pipe.ID
	return fmt.Errorf("pipe: %s nodes[%d](%s) %s (%s)", pid, node.index, node.Name, message, ctx.id)
}
