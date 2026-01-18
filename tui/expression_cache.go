package tui

import (
	"sync"
	"time"

	"github.com/expr-lang/expr/vm"
)

const (
	defaultExprCacheTTL = 5 * time.Minute
)

type cachedExpression struct {
	program *vm.Program
	ttl     time.Time
}

type ExpressionCache struct {
	cache map[string]*cachedExpression
	mu    sync.RWMutex
	ttl   time.Duration
}

func NewExpressionCache() *ExpressionCache {
	return &ExpressionCache{
		cache: make(map[string]*cachedExpression),
		ttl:   defaultExprCacheTTL,
	}
}

func (c *ExpressionCache) GetOrCompile(
	expression string,
	compiler func(string) (*vm.Program, error),
) (*vm.Program, error) {
	c.mu.RLock()
	cached, exists := c.cache[expression]
	c.mu.RUnlock()

	if exists {
		if cached.ttl.IsZero() || time.Now().Before(cached.ttl) {
			return cached.program, nil
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if cached, exists := c.cache[expression]; exists {
		if cached.ttl.IsZero() || time.Now().Before(cached.ttl) {
			return cached.program, nil
		}
	}

	program, err := compiler(expression)
	if err != nil {
		return nil, err
	}

	c.cache[expression] = &cachedExpression{
		program: program,
		ttl:     time.Now().Add(c.ttl),
	}

	return program, nil
}

func (c *ExpressionCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, key)
}

func (c *ExpressionCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*cachedExpression)
}
