package tui

import (
	"testing"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

func BenchmarkExpressionCache(b *testing.B) {
	cache := NewExpressionCache()

	expressions := []string{
		"username",
		"users[0].name",
		"len(users)",
		"count > 10",
		"(a + b) * c",
		"map(users, .email)",
	}

	env := map[string]interface{}{
		"username": "test",
		"users": []map[string]interface{}{
			{"name": "Alice", "email": "alice@example.com"},
			{"name": "Bob", "email": "bob@example.com"},
		},
		"count": 15,
		"a":     5,
		"b":     3,
		"c":     2,
	}

	b.Run("WithCache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, exprStr := range expressions {
				program, err := cache.GetOrCompile(exprStr, func(s string) (*vm.Program, error) {
					return expr.Compile(s, expr.Env(env))
				})
				if err != nil {
					b.Fatal(err)
				}

				_, err = expr.Run(program, env)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("WithoutCache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, exprStr := range expressions {
				program, err := expr.Compile(exprStr, expr.Env(env))
				if err != nil {
					b.Fatal(err)
				}

				_, err = expr.Run(program, env)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}

func BenchmarkExpressionCacheTTL(b *testing.B) {
	cache := NewExpressionCache()

	exprStr := "(a + b) * c"
	env := map[string]interface{}{
		"a": 5,
		"b": 3,
		"c": 2,
	}

	b.Run("FirstCompile", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			newCache := NewExpressionCache()
			_, err := newCache.GetOrCompile(exprStr, func(s string) (*vm.Program, error) {
				return expr.Compile(s, expr.Env(env))
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("CacheHit", func(b *testing.B) {
		_, err := cache.GetOrCompile(exprStr, func(s string) (*vm.Program, error) {
			return expr.Compile(s, expr.Env(env))
		})
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := cache.GetOrCompile(exprStr, func(s string) (*vm.Program, error) {
				return expr.Compile(s, expr.Env(env))
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestExpressionCache(t *testing.T) {
	cache := NewExpressionCache()

	exprStr := "a + b"
	env := map[string]interface{}{
		"a": 10,
		"b": 20,
	}

	program, err := cache.GetOrCompile(exprStr, func(s string) (*vm.Program, error) {
		return expr.Compile(s, expr.Env(env))
	})
	if err != nil {
		t.Fatalf("First compilation failed: %v", err)
	}

	result, err := expr.Run(program, env)
	if err != nil {
		t.Fatalf("First evaluation failed: %v", err)
	}

	if result != 30 {
		t.Fatalf("Expected 30, got %v", result)
	}

	program2, err := cache.GetOrCompile(exprStr, func(s string) (*vm.Program, error) {
		t.Fatal("Should not call compilation function on cache hit")
		return expr.Compile(s, expr.Env(env))
	})
	if err != nil {
		t.Fatalf("Cache lookup failed: %v", err)
	}

	result2, err := expr.Run(program2, env)
	if err != nil {
		t.Fatalf("Cache evaluation failed: %v", err)
	}

	if result2 != 30 {
		t.Fatalf("Expected 30, got %v", result2)
	}
}

func TestExpressionCacheInvalidate(t *testing.T) {
	cache := NewExpressionCache()

	exprStr := "a + b"
	env := map[string]interface{}{
		"a": 10,
		"b": 20,
	}

	_, err := cache.GetOrCompile(exprStr, func(s string) (*vm.Program, error) {
		return expr.Compile(s, expr.Env(env))
	})
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	cache.Invalidate(exprStr)

	callCount := 0
	_, err = cache.GetOrCompile(exprStr, func(s string) (*vm.Program, error) {
		callCount++
		return expr.Compile(s, expr.Env(env))
	})
	if err != nil {
		t.Fatalf("Re-compilation failed: %v", err)
	}

	if callCount != 1 {
		t.Fatalf("Expected 1 compilation call after invalidate, got %d", callCount)
	}
}

func TestExpressionCacheClear(t *testing.T) {
	cache := NewExpressionCache()

	expressions := []string{"a", "b", "c"}
	env := map[string]interface{}{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	for _, exprStr := range expressions {
		_, err := cache.GetOrCompile(exprStr, func(s string) (*vm.Program, error) {
			return expr.Compile(s, expr.Env(env))
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	cache.Clear()

	callCount := 0
	for _, exprStr := range expressions {
		_, err := cache.GetOrCompile(exprStr, func(s string) (*vm.Program, error) {
			callCount++
			return expr.Compile(s, expr.Env(env))
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	if callCount != len(expressions) {
		t.Fatalf("Expected %d compilation calls after clear, got %d", len(expressions), callCount)
	}
}
