package binding

import (
	"strconv"
	"strings"
	"unicode"
)

// Expression represents a parsed expression that can be evaluated.
type Expression struct {
	tokens  []token
	depKeys []string
}

// token represents a single token in the expression.
type token struct {
	kind tokenKind
	text string
}

// tokenKind represents the type of token.
type tokenKind int

const (
	tokenEOF tokenKind = iota
	tokenIdent
	tokenNumber
	tokenString
	tokenOp
	tokenParenOpen
	tokenParenClose
)

// ParseExpression parses an expression string into an Expression.
func ParseExpression(expr string) *Expression {
	lexer := newLexer(expr)
	tokens := lexer.tokenize()

	// Extract dependencies
	deps := extractDependencies(tokens)

	return &Expression{
		tokens:  tokens,
		depKeys: deps,
	}
}

// Evaluate evaluates the expression against the given context.
func (e *Expression) Evaluate(ctx Context) interface{} {
	if len(e.tokens) == 0 {
		return nil
	}

	// Simple evaluator for basic expressions
	// Supports: variables, literals, binary operators
	parser := newParser(e.tokens)
	ast := parser.parse()

	if ast == nil {
		return nil
	}

	return ast.eval(ctx)
}

// Dependencies returns the list of keys this expression depends on.
func (e *Expression) Dependencies() []string {
	if e.depKeys == nil {
		return []string{}
	}
	return e.depKeys
}

// extractDependencies extracts variable dependencies from tokens.
func extractDependencies(tokens []token) []string {
	deps := make(map[string]bool)

	for i, tok := range tokens {
		if tok.kind == tokenIdent {
			// An identifier is a variable dependency if:
			// - It's the first token, OR
			// - The previous token is an operator (this means it's an operand)
			shouldAdd := false
			if i == 0 {
				shouldAdd = true
			} else if i > 0 && tokens[i-1].kind == tokenOp {
				shouldAdd = true
			}

			if shouldAdd {
				// Extract root path (before first dot)
				path := tok.text
				if idx := strings.Index(path, "."); idx > 0 {
					path = path[:idx]
				}
				if path != "" {
					deps[path] = true
				}
			}
		}
	}

	result := make([]string, 0, len(deps))
	for dep := range deps {
		result = append(result, dep)
	}
	return result
}

// lexer is a simple lexical analyzer for expressions.
type lexer struct {
	input string
	pos   int
	ch    byte
}

func newLexer(input string) *lexer {
	l := &lexer{input: input}
	if len(l.input) > 0 {
		l.ch = l.input[0]
	}
	return l
}

func (l *lexer) tokenize() []token {
	var tokens []token

	for l.ch != 0 {
		switch {
		case unicode.IsSpace(rune(l.ch)):
			l.skipWhitespace()
		case l.ch == '(':
			tokens = append(tokens, token{kind: tokenParenOpen, text: "("})
			l.readChar()
		case l.ch == ')':
			tokens = append(tokens, token{kind: tokenParenClose, text: ")"})
			l.readChar()
		case l.ch == '"' || l.ch == '\'':
			tokens = append(tokens, l.readString())
		case l.ch == '+' || l.ch == '-' || l.ch == '*' || l.ch == '/':
			tokens = append(tokens, token{kind: tokenOp, text: string(l.ch)})
			l.readChar()
		case unicode.IsDigit(rune(l.ch)):
			tokens = append(tokens, l.readNumber())
		case isIdentStart(l.ch):
			tokens = append(tokens, l.readIdent())
		default:
			l.readChar()
		}
	}

	return tokens
}

func (l *lexer) readChar() {
	l.pos++
	if l.pos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.pos]
	}
}

func (l *lexer) skipWhitespace() {
	for unicode.IsSpace(rune(l.ch)) {
		l.readChar()
	}
}

func (l *lexer) readIdent() token {
	start := l.pos
	for isIdentPart(l.ch) {
		l.readChar()
	}
	return token{kind: tokenIdent, text: l.input[start:l.pos]}
}

func (l *lexer) readNumber() token {
	start := l.pos
	for unicode.IsDigit(rune(l.ch)) {
		l.readChar()
	}
	// Handle decimal
	if l.ch == '.' {
		l.readChar()
		for unicode.IsDigit(rune(l.ch)) {
			l.readChar()
		}
	}
	return token{kind: tokenNumber, text: l.input[start:l.pos]}
}

func (l *lexer) readString() token {
	quote := l.ch
	l.readChar()
	start := l.pos

	for l.ch != quote && l.ch != 0 {
		l.readChar()
	}

	str := l.input[start:l.pos]
	l.readChar() // Skip closing quote

	return token{kind: tokenString, text: str}
}

func isIdentStart(ch byte) bool {
	return ch == '_' || unicode.IsLetter(rune(ch))
}

func isIdentPart(ch byte) bool {
	return ch == '_' || ch == '.' || unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch))
}

// parser is a simple recursive descent parser.
type parser struct {
	tokens  []token
	current int
}

func newParser(tokens []token) *parser {
	return &parser{tokens: tokens, current: 0}
}

func (p *parser) parse() astNode {
	return p.parseExpression()
}

func (p *parser) parseExpression() astNode {
	return p.parseBinaryOp(0)
}

func (p *parser) parseBinaryOp(precedence int) astNode {
	left := p.parseUnary()

	for {
		op := p.peek()
		if op.kind == tokenEOF || op.kind == tokenParenClose {
			break
		}
		if op.kind != tokenOp {
			break
		}

		opPrec := operatorPrecedence(op.text)
		if opPrec < precedence {
			break
		}

		p.advance()
		right := p.parseBinaryOp(opPrec + 1)
		left = &binaryOpNode{op: op.text, left: left, right: right}
	}

	return left
}

func (p *parser) parseUnary() astNode {
	tok := p.peek()

	if tok.kind == tokenOp && (tok.text == "-" || tok.text == "+") {
		p.advance()
		expr := p.parseUnary()
		return &unaryOpNode{op: tok.text, expr: expr}
	}

	if tok.kind == tokenParenOpen {
		p.advance()
		expr := p.parseExpression()
		if p.peek().kind == tokenParenClose {
			p.advance()
		}
		return expr
	}

	return p.parsePrimary()
}

func (p *parser) parsePrimary() astNode {
	tok := p.peek()

	switch tok.kind {
	case tokenIdent:
		p.advance()
		return &identNode{name: tok.text}
	case tokenNumber:
		p.advance()
		val, _ := strconv.ParseFloat(tok.text, 64)
		return &numberLiteral{value: val}
	case tokenString:
		p.advance()
		return &stringLiteral{value: tok.text}
	default:
		p.advance()
		return &nilNode{}
	}
}

func (p *parser) peek() token {
	if p.current >= len(p.tokens) {
		return token{kind: tokenEOF}
	}
	return p.tokens[p.current]
}

func (p *parser) advance() {
	p.current++
}

func operatorPrecedence(op string) int {
	switch op {
	case "*", "/":
		return 2
	case "+", "-":
		return 1
	default:
		return 0
	}
}

// astNode represents an abstract syntax tree node.
type astNode interface {
	eval(ctx Context) interface{}
}

type identNode struct {
	name string
}

func (n *identNode) eval(ctx Context) interface{} {
	val, ok := ctx.Get(n.name)
	if ok {
		return val
	}
	return nil
}

type numberLiteral struct {
	value float64
}

func (n *numberLiteral) eval(ctx Context) interface{} {
	return n.value
}

type stringLiteral struct {
	value string
}

func (n *stringLiteral) eval(ctx Context) interface{} {
	return n.value
}

type nilNode struct{}

func (n *nilNode) eval(ctx Context) interface{} {
	return nil
}

type binaryOpNode struct {
	op    string
	left  astNode
	right astNode
}

func (n *binaryOpNode) eval(ctx Context) interface{} {
	leftVal := n.left.eval(ctx)
	rightVal := n.right.eval(ctx)

	// Handle numeric operations
	if leftNum, ok := toFloat64(leftVal); ok {
		if rightNum, ok := toFloat64(rightVal); ok {
			switch n.op {
			case "+":
				return leftNum + rightNum
			case "-":
				return leftNum - rightNum
			case "*":
				return leftNum * rightNum
			case "/":
				if rightNum != 0 {
					return leftNum / rightNum
				}
				return 0.0
			}
		}
	}

	// Handle string concatenation
	if leftStr, ok := leftVal.(string); ok {
		if rightStr, ok := rightVal.(string); ok {
			return leftStr + rightStr
		}
	}

	return nil
}

type unaryOpNode struct {
	op   string
	expr astNode
}

func (n *unaryOpNode) eval(ctx Context) interface{} {
	val := n.expr.eval(ctx)

	if num, ok := toFloat64(val); ok {
		if n.op == "-" {
			return -num
		}
	}
	return val
}

// toFloat64 converts a value to float64.
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float64:
		return val, true
	case float32:
		return float64(val), true
	default:
		return 0, false
	}
}

// EvaluateExpression is a convenience function to evaluate an expression string.
func EvaluateExpression(expr string, ctx Context) interface{} {
	parsed := ParseExpression(expr)
	return parsed.Evaluate(ctx)
}

// IsValidExpression checks if a string is a valid expression.
func IsValidExpression(expr string) bool {
	lexer := newLexer(expr)
	tokens := lexer.tokenize()

	// Basic validation: must have at least one meaningful token
	for _, tok := range tokens {
		if tok.kind == tokenIdent || tok.kind == tokenNumber || tok.kind == tokenString {
			return true
		}
	}
	return false
}

// FormatExpression formats an expression for display.
func FormatExpression(expr string) string {
	return "{{ " + strings.TrimSpace(expr) + " }}"
}
