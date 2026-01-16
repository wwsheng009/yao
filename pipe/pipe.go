package pipe

import (
	"fmt"
	"path"
	"strings"

	"github.com/yaoapp/gou/application"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/share"
)

var pipes = map[string]*Pipe{}

func whitelistAllowsProcess(whitelist Whitelist, process string) bool {
	// Nil or empty whitelist means "no restriction".
	if len(whitelist) == 0 {
		return true
	}

	// Fast path: exact match.
	if _, ok := whitelist[process]; ok {
		return true
	}

	// Wildcards: support common prefix form "utils.*" and standard glob patterns.
	for pattern := range whitelist {
		if pattern == process {
			continue
		}

		// "utils.*" should match "utils.validate_age".
		// Only apply this fast-path when the prefix itself contains no other glob symbols.
		if strings.HasSuffix(pattern, ".*") && !strings.ContainsAny(strings.TrimSuffix(pattern, ".*"), "*?[]") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(process, prefix) {
				return true
			}
			continue
		}

		// Glob patterns (* ? []): e.g. "utils.*" "*.fmt.*".
		if !strings.ContainsAny(pattern, "*?[]") {
			continue
		}

		matched, err := path.Match(pattern, process)
		if err == nil && matched {
			return true
		}
	}

	return false
}

// Load the pipe
func Load(cfg config.Config) error {
	messages := []string{}

	// Ignore if the pipes directory does not exist
	exists, err := application.App.Exists("pipes")
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	exts := []string{"*.pip.yao", "*.pipe.yao"}
	err = application.App.Walk("pipes", func(root, file string, isdir bool) error {
		if isdir {
			return nil
		}

		id := share.ID(root, file)
		pipe, err := NewFile(file, root)
		if err != nil {
			messages = append(messages, err.Error())
			return nil
		}

		Set(id, pipe)
		return nil
	}, exts...)

	if len(messages) > 0 {
		return fmt.Errorf("%s", strings.Join(messages, ";\n"))
	}

	return err
}

// New create Pipe
func New(source []byte) (*Pipe, error) {
	pipe := Pipe{}
	err := application.Parse("<source>.yao", source, &pipe)
	if err != nil {
		return nil, fmt.Errorf("parse pipe: %s", err)
	}

	err = (&pipe).build()
	if err != nil {
		return nil, fmt.Errorf("build pipe: %s: %w", pipe.Name, err)
	}

	return &pipe, nil
}

// NewFile create pipe from file
func NewFile(file string, root string) (*Pipe, error) {
	source, err := application.App.Read(file)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", file, err)
	}

	id := share.ID(root, file)
	pipe := Pipe{ID: id}
	err = application.Parse(file, source, &pipe)
	if err != nil {
		return nil, fmt.Errorf("parse pipe file %s: %w", file, err)
	}

	err = (&pipe).build()
	if err != nil {
		return nil, fmt.Errorf("build pipe from file %s: %w", file, err)
	}

	return &pipe, nil
}

// Set pipe to
func Set(id string, pipe *Pipe) {
	pipes[id] = pipe
}

// Remove the pipe
func Remove(id string) {
	delete(pipes, id)
}

// Get the pipe
func Get(id string) (*Pipe, error) {
	if pipe, has := pipes[id]; has {
		return pipe, nil
	}
	return nil, fmt.Errorf("pipe %s not found", id)
}

// Build the pipe
func (pipe *Pipe) build() error {

	if len(pipe.Nodes) == 0 {
		return fmt.Errorf("pipe: %s (ID: %s) nodes is required", pipe.Name, pipe.ID)
	}

	return pipe._build()
}

// HasNodes check if the pipe has nodes
func (pipe *Pipe) HasNodes() bool {
	return len(pipe.Nodes) > 0
}

func (pipe *Pipe) _build() error {

	pipe.mapping = map[string]*Node{}
	if pipe.Nodes == nil {
		return nil
	}

	for i, node := range pipe.Nodes {
		if node.Name == "" {
			return fmt.Errorf("pipe: %s (ID: %s) nodes[%d] name is required", pipe.Name, pipe.ID, i)
		}

		pipe.Nodes[i].index = i
		pipe.mapping[node.Name] = &pipe.Nodes[i]

		// Set the label of the node
		if node.Label == "" {
			pipe.Nodes[i].Label = strings.ToUpper(node.Name)
		}

		// Set the type of the node
		if node.Process != nil {
			pipe.Nodes[i].Type = "process"

			// Validate the process
			if node.Process.Name == "" {
				return fmt.Errorf("pipe: %s (ID: %s) nodes[%d] (name: %s) process name is required", pipe.Name, pipe.ID, i, node.Name)
			}

			// Security check
			// NOTE: An empty whitelist (e.g. "whitelist": []) is treated as "no restriction".
			if pipe.Whitelist != nil {
				if !whitelistAllowsProcess(pipe.Whitelist, node.Process.Name) {
					return fmt.Errorf("pipe: %s (ID: %s) nodes[%d] (name: %s) process %s is not in the whitelist", pipe.Name, pipe.ID, i, node.Name, node.Process.Name)
				}
			}
			continue

		} else if node.Request != nil {
			pipe.Nodes[i].Type = "request"
			continue

		} else if node.Prompts != nil {
			pipe.Nodes[i].Type = "ai"
			continue

		} else if node.UI != "" {
			pipe.Nodes[i].Type = "user-input"
			if node.UI != "cli" && node.UI != "web" && node.UI != "app" && node.UI != "wxapp" { // Vaildate the UI type
				return fmt.Errorf("pipe: %s (ID: %s) nodes[%d] (name: %s) the type of the UI must be cli, web, app, wxapp", pipe.Name, pipe.ID, i, node.Name)
			}
			continue

		} else if node.Switch != nil {
			pipe.Nodes[i].Type = "switch"
			for key, pip := range node.Switch {
				key = ref(key)
				pip.Whitelist = pipe.Whitelist // Copy the whitelist
				pip.namespace = node.Name
				pip.parent = pipe
				if pip.ID == "" {
					pip.ID = fmt.Sprintf("%s.%s#%s", pipe.ID, node.Name, key)
				}
				if pip.Name == "" {
					pip.Name = fmt.Sprintf("%s(%s#%s)", pipe.Name, node.Name, key)
				}
				pip._build()
			}
			continue
		}

		return fmt.Errorf("pipe: %s (ID: %s) nodes[%d] (name: %s) process, request, case, prompts or ui is required at least one", pipe.Name, pipe.ID, i, node.Name)
	}

	return nil
}
