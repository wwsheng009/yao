package tui

import (
	"regexp"
	"testing"

	"github.com/expr-lang/expr"
	"github.com/stretchr/testify/assert"
)

func TestSimpleExpr(t *testing.T) {
	// Test the regex pattern
	stmtRe := regexp.MustCompile(`\{\{([\s\S]*?)\}\}`)
	text := "Items count: {{len items}}"
	matches := stmtRe.FindAllStringSubmatch(text, -1)
	
	assert.Equal(t, 1, len(matches))
	assert.Equal(t, "len items", matches[0][1])

	// Test expression evaluation
	env := map[string]interface{}{
		"items": []interface{}{"apple", "banana", "cherry"},
	}

	program, err := expr.Compile("len(items)", expr.AllowUndefinedVariables())
	assert.NoError(t, err)

	result, err := expr.Run(program, env)
	assert.NoError(t, err)
	assert.Equal(t, 3, result)

	// Test with custom functions
	opts := []expr.Option{
		expr.Function("len", func(params ...interface{}) (interface{}, error) {
			if len(params) == 0 {
				return 0, nil
			}
			v := params[0]
			switch val := v.(type) {
			case []interface{}:
				return len(val), nil
			case []map[string]interface{}:
				return len(val), nil
			case string:
				return len(val), nil
			case map[string]interface{}:
				return len(val), nil
			default:
				return 0, nil
			}
		}),
		expr.AllowUndefinedVariables(),
	}

	program2, err := expr.Compile("len(items)", opts...)
	assert.NoError(t, err)

	result2, err := expr.Run(program2, env)
	assert.NoError(t, err)
	assert.Equal(t, 3, result2)
}