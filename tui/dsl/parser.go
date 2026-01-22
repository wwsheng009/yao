// Package dsl provides the DSL (Domain Specific Language) parser for TUI layouts.
//
// This file implements the parser for JSON, JSONC (JSON with comments), and YAML formats.
package dsl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/yaoapp/kun/log"
	"gopkg.in/yaml.v3"
)

// Regular expressions for stripping comments from JSONC
var (
	// Matches // comments (but not inside strings)
	lineCommentRe = regexp.MustCompile(`//.*`)
	// Matches /* */ block comments (but not inside strings)
	blockCommentRe = regexp.MustCompile(`/\*[\s\S]*?\*/`)
)

// Parse parses TUI configuration from data in JSON, JSONC, or YAML format.
// The format is auto-detected based on the content.
func Parse(data []byte) (*Config, error) {
	// Try JSON first
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err == nil {
		log.Trace("DSL: Parsed as JSON")
		return &cfg, nil
	}

	// Try JSONC (JSON with comments)
	stripped := stripJSONCComments(data)
	if stripped != nil {
		if err := json.Unmarshal(stripped, &cfg); err == nil {
			log.Trace("DSL: Parsed as JSONC (JSON with comments)")
			return &cfg, nil
		}
	}

	// Try YAML
	if err := yaml.Unmarshal(data, &cfg); err == nil {
		log.Trace("DSL: Parsed as YAML")
		return &cfg, nil
	}

	return nil, fmt.Errorf("failed to parse as JSON, JSONC, or YAML")
}

// stripJSONCComments removes // and /* */ style comments from JSONC data.
// It preserves strings that contain comment-like patterns.
func stripJSONCComments(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}

	var result bytes.Buffer
	var inString bool
	var escape bool
	var inLineComment bool
	var inBlockComment bool
	var braceDepth int
	var parenDepth int
	var bracketDepth int

	for i := 0; i < len(data); i++ {
		c := data[i]

		if escape {
			result.WriteByte(c)
			escape = false
			continue
		}

		if c == '\\' && inString {
			result.WriteByte(c)
			escape = true
			continue
		}

		if c == '"' && !inLineComment && !inBlockComment {
			inString = !inString
			result.WriteByte(c)
			continue
		}

		if inString {
			result.WriteByte(c)
			continue
		}

		// Check for // line comment
		if !inBlockComment && i+1 < len(data) && c == '/' && data[i+1] == '/' {
			inLineComment = true
			// Replace with space to preserve positions
			result.WriteByte(' ')
			continue
		}

		// Check for /* block comment
		if !inLineComment && i+1 < len(data) && c == '/' && data[i+1] == '*' {
			inBlockComment = true
			// Replace with space to preserve positions
			result.WriteByte(' ')
			continue
		}

		// Check for end of line comment
		if inLineComment && c == '\n' {
			inLineComment = false
			result.WriteByte(c)
			continue
		}

		// Check for end of block comment
		if inBlockComment && i+1 < len(data) && c == '*' && data[i+1] == '/' {
			inBlockComment = false
			result.WriteByte(' ')
			result.WriteByte(' ')
			i++ // skip the '/'
			continue
		}

		if inLineComment || inBlockComment {
			// Replace comment content with space
			if c != '\n' && c != '\r' {
				result.WriteByte(' ')
			} else {
				result.WriteByte(c)
			}
			continue
		}

		// Track nesting levels for better comma handling
		switch c {
		case '{':
			braceDepth++
		case '}':
			braceDepth--
		case '[':
			bracketDepth++
		case ']':
			bracketDepth--
		case '(':
			parenDepth++
		case ')':
			parenDepth--
		}

		result.WriteByte(c)
	}

	// Remove trailing commas (not valid in JSON)
	stripped := removeTrailingCommas(result.Bytes())
	return stripped
}

// removeTrailingCommas removes commas before closing brackets/braces.
// This allows JSONC files to have trailing commas.
func removeTrailingCommas(data []byte) []byte {
	var result bytes.Buffer
	var inString bool
	var escape bool

	for i := 0; i < len(data); i++ {
		c := data[i]

		if escape {
			result.WriteByte(c)
			escape = false
			continue
		}

		if c == '\\' && inString {
			result.WriteByte(c)
			escape = true
			continue
		}

		if c == '"' {
			inString = !inString
			result.WriteByte(c)
			continue
		}

		if inString {
			result.WriteByte(c)
			continue
		}

		// Check for comma followed by whitespace and closing bracket/brace
		if c == ',' && i+1 < len(data) {
			// Skip whitespace
			j := i + 1
			for j < len(data) && isJSONSpace(data[j]) {
				j++
			}
			// If next non-whitespace is closing bracket/brace, skip the comma
			if j < len(data) && (data[j] == ']' || data[j] == '}') {
				// Write the whitespace but not the comma
				for k := i + 1; k < j; k++ {
					result.WriteByte(data[k])
				}
				i = j - 1
				continue
			}
		}

		result.WriteByte(c)
	}

	return result.Bytes()
}

// isJSONSpace checks if a byte is JSON whitespace.
func isJSONSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// ParseReader parses TUI configuration from an io.Reader.
// The format is auto-detected based on the content.
func ParseReader(r io.Reader) (*Config, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read: %w", err)
	}
	return Parse(data)
}

// ParseFile parses a TUI configuration file.
// The format is determined by the file extension.
// Supported extensions:
//   - .json, .jsonc - JSON or JSON with comments
//   - .yml, .yaml - YAML format
//   - .tui.yao - JSONC (JSON with comments, recommended format)
func ParseFile(data []byte, filename string) (*Config, error) {
	ext := strings.ToLower(filename)
	if strings.HasSuffix(ext, ".json") {
		return ParseJSON(data)
	}
	if strings.HasSuffix(ext, ".jsonc") {
		return ParseJSONC(data)
	}
	if strings.HasSuffix(ext, ".tui.yao") {
		// .tui.yao files use JSONC format (JSON with comments)
		return ParseJSONC(data)
	}
	if strings.HasSuffix(ext, ".yml") || strings.HasSuffix(ext, ".yaml") {
		return ParseYAML(data)
	}
	// Auto-detect
	return Parse(data)
}

// ParseJSON parses TUI configuration from JSON data.
func ParseJSON(data []byte) (*Config, error) {
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return &cfg, nil
}

// ParseJSONC parses TUI configuration from JSONC data (JSON with comments).
// Supports both // line comments and /* */ block comments.
// Also supports trailing commas.
func ParseJSONC(data []byte) (*Config, error) {
	stripped := stripJSONCComments(data)
	if stripped == nil {
		return nil, fmt.Errorf("failed to strip comments from JSONC")
	}
	var cfg Config
	if err := json.Unmarshal(stripped, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse JSONC: %w", err)
	}
	return &cfg, nil
}

// ParseYAML parses TUI configuration from YAML data.
func ParseYAML(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return &cfg, nil
}

// Validate validates the parsed configuration.
func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("config.name is required")
	}

	if c.Layout == nil {
		return fmt.Errorf("config.layout is required")
	}

	// Validate layout tree
	if err := c.Layout.Validate(); err != nil {
		return fmt.Errorf("layout validation failed: %w", err)
	}

	return nil
}

// Validate validates a node and its children.
func (n *Node) Validate() error {
	// If type is not specified but there are children, treat as layout node
	if n.Type == "" {
		if len(n.Children) > 0 {
			// Implicit layout node
			n.Type = "layout"
		} else {
			// Leaf node must have a type
			return fmt.Errorf("node missing 'type' field")
		}
	}

	// Validate direction for layout nodes
	if isLayoutType(n.Type) {
		dir := normalizeDirection(n.Direction)
		if dir == "" {
			// Default to column for layout nodes
			n.Direction = "column"
		} else {
			n.Direction = dir
		}
	}

	// Validate children
	for i, child := range n.Children {
		if err := child.Validate(); err != nil {
			return fmt.Errorf("child %d: %w", i, err)
		}
	}

	return nil
}

// isLayoutType checks if a type is a layout container.
func isLayoutType(typ string) bool {
	switch typ {
	case "layout", "row", "column", "vertical", "horizontal":
		return true
	default:
		return false
	}
}

// normalizeDirection normalizes direction values.
func normalizeDirection(dir string) string {
	switch strings.ToLower(dir) {
	case "row", "horizontal":
		return "row"
	case "column", "vertical":
		return "column"
	case "":
		return ""
	default:
		return dir
	}
}

// AssignIDs assigns unique IDs to nodes that don't have an ID.
func (n *Node) AssignIDs(prefix string, counter *int) {
	if n.ID == "" {
		*counter++
		n.ID = fmt.Sprintf("%s_%d", prefix, *counter)
	}

	for _, child := range n.Children {
		child.AssignIDs(n.ID, counter)
	}
}

// GetNodeByID finds a node by its ID.
func (n *Node) GetNodeByID(id string) *Node {
	if n.ID == id {
		return n
	}

	for _, child := range n.Children {
		if found := child.GetNodeByID(id); found != nil {
			return found
		}
	}

	return nil
}

// GetAllNodes returns all nodes in the tree as a flat map.
func (n *Node) GetAllNodes() map[string]*Node {
	result := make(map[string]*Node)
	n.collectNodes(result)
	return result
}

// collectNodes recursively collects nodes into the result map.
func (n *Node) collectNodes(result map[string]*Node) {
	result[n.ID] = n
	for _, child := range n.Children {
		child.collectNodes(result)
	}
}

// ToJSON converts the config to JSON.
func (c *Config) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// ToYAML converts the config to YAML.
func (c *Config) ToYAML() ([]byte, error) {
	return yaml.Marshal(c)
}

// FlattenData flattens nested data structures for template binding.
// Converts {"user": {"name": "John"}} to {"user.name": "John"}.
func FlattenData(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	flatten(data, "", result)
	return result
}

// flatten recursively flattens a nested map.
func flatten(data map[string]interface{}, prefix string, result map[string]interface{}) {
	for key, value := range data {
		newKey := key
		if prefix != "" {
			newKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			flatten(v, newKey, result)
		default:
			result[newKey] = value
		}
	}
}
