package render

import (
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/yao/tui/runtime"
)

// TestNewCache tests cache creation
func TestNewCache(t *testing.T) {
	config := CacheConfig{
		MaxSize:     100,
		TTL:         1 * time.Minute,
		EnableStats: true,
	}

	cache := NewCache(config)

	assert.NotNil(t, cache)
	assert.Equal(t, 100, cache.maxSize)
	assert.Equal(t, 1*time.Minute, cache.ttl)
	assert.True(t, cache.enableStats)
}

// TestCacheGetSetStyle tests style caching
func TestCacheGetSetStyle(t *testing.T) {
	cache := NewCache(CacheConfig{
		MaxSize:     10,
		TTL:         1 * time.Minute,
		EnableStats: false,
	})

	style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("red"))
	expectedStyle := LipglossToCellStyle(style)

	// First get should miss
	result, ok := cache.GetStyle(style)
	assert.False(t, ok, "First get should miss")
	assert.Equal(t, runtime.CellStyle{}, result, "Miss should return empty style")

	// Set the style
	cache.SetStyle(style, expectedStyle)

	// Second get should hit
	result, ok = cache.GetStyle(style)
	assert.True(t, ok, "Second get should hit")
	assert.Equal(t, expectedStyle, result, "Hit should return cached style")
}

// TestCacheGetSetMeasurement tests measurement caching
func TestCacheGetSetMeasurement(t *testing.T) {
	cache := NewCache(CacheConfig{
		MaxSize:     10,
		TTL:         1 * time.Minute,
		EnableStats: false,
	})

	style := lipgloss.NewStyle().Bold(true)
	text := "Hello, World!"

	// First get should miss
	width, height, ok := cache.GetMeasurement(text, style)
	assert.False(t, ok, "First get should miss")
	assert.Equal(t, 0, width)
	assert.Equal(t, 0, height)

	// Set the measurement
	cache.SetMeasurement(text, style, 13, 1)

	// Second get should hit
	width, height, ok = cache.GetMeasurement(text, style)
	assert.True(t, ok, "Second get should hit")
	assert.Equal(t, 13, width)
	assert.Equal(t, 1, height)
}

// TestCacheGetSetANSISegments tests ANSI segment caching
func TestCacheGetSetANSISegments(t *testing.T) {
	cache := NewCache(CacheConfig{
		MaxSize:     10,
		TTL:         1 * time.Minute,
		EnableStats: false,
	})

	line := "\x1b[1mBold Text\x1b[0m"
	segments := []StyledSegment{
		{Text: "Bold Text", Style: runtime.CellStyle{Bold: true}},
	}

	// First get should miss
	result, ok := cache.GetANSISegments(line)
	assert.False(t, ok, "First get should miss")
	assert.Nil(t, result)

	// Set the segments
	cache.SetANSISegments(line, segments)

	// Second get should hit
	result, ok = cache.GetANSISegments(line)
	assert.True(t, ok, "Second get should hit")
	assert.Equal(t, segments, result)
}

// TestCacheGetSetRenderedText tests rendered text caching
func TestCacheGetSetRenderedText(t *testing.T) {
	cache := NewCache(CacheConfig{
		MaxSize:     10,
		TTL:         1 * time.Minute,
		EnableStats: false,
	})

	style := lipgloss.NewStyle().Bold(true)
	text := "Hello"
	rendered := "\x1b[1mHello\x1b[0m"

	// First get should miss
	result, ok := cache.GetRenderedText(text, style)
	assert.False(t, ok, "First get should miss")
	assert.Equal(t, "", result)

	// Set the rendered text
	cache.SetRenderedText(text, style, rendered)

	// Second get should hit
	result, ok = cache.GetRenderedText(text, style)
	assert.True(t, ok, "Second get should hit")
	assert.Equal(t, rendered, result)
}

// TestCacheStats tests cache statistics
func TestCacheStats(t *testing.T) {
	cache := NewCache(CacheConfig{
		MaxSize:     10,
		TTL:         1 * time.Minute,
		EnableStats: true,
	})

	style := lipgloss.NewStyle().Bold(true)
	text := "Hello"

	// Generate some hits and misses
	cache.GetStyle(style)                     // miss
	cache.SetStyle(style, runtime.CellStyle{}) // set
	cache.GetStyle(style)                     // hit
	cache.GetMeasurement(text, style)         // miss
	cache.SetMeasurement(text, style, 5, 1)   // set
	cache.GetMeasurement(text, style)         // hit

	stats := cache.Stats()

	assert.Equal(t, 1, stats.StyleCount, "Should have 1 style")
	assert.Equal(t, 1, stats.MeasurementCount, "Should have 1 measurement")
	assert.Equal(t, 2, stats.Hits, "Should have 2 hits")
	assert.Equal(t, 2, stats.Misses, "Should have 2 misses")
	assert.Equal(t, 0.5, stats.HitRate, "Hit rate should be 50%")
}

// TestCacheClear tests clearing the cache
func TestCacheClear(t *testing.T) {
	cache := NewCache(CacheConfig{
		MaxSize:     10,
		TTL:         1 * time.Minute,
		EnableStats: true,
	})

	style := lipgloss.NewStyle().Bold(true)
	text := "Hello"

	// Add some entries
	cache.SetStyle(style, runtime.CellStyle{})
	cache.SetMeasurement(text, style, 5, 1)
	cache.SetANSISegments("line", []StyledSegment{})
	cache.SetRenderedText(text, style, "rendered")

	// Clear the cache
	cache.Clear()

	// All should be empty
	assert.Equal(t, 0, len(cache.styles))
	assert.Equal(t, 0, len(cache.measurements))
	assert.Equal(t, 0, len(cache.ansiSegments))
	assert.Equal(t, 0, len(cache.renderedText))
	assert.Equal(t, 0, cache.hits)
	assert.Equal(t, 0, cache.misses)
}

// TestCacheEvictExpired tests TTL-based expiration
func TestCacheEvictExpired(t *testing.T) {
	cache := NewCache(CacheConfig{
		MaxSize:     10,
		TTL:         10 * time.Millisecond,
		EnableStats: false,
	})

	style := lipgloss.NewStyle().Bold(true)
	text := "Hello"

	// Add entries
	cache.SetStyle(style, runtime.CellStyle{})
	cache.SetMeasurement(text, style, 5, 1)

	// Entries should exist
	_, ok := cache.GetStyle(style)
	assert.True(t, ok)

	// Wait for expiration
	time.Sleep(15 * time.Millisecond)

	// Evict expired
	cache.EvictExpired()

	// Entries should be gone
	_, ok = cache.GetStyle(style)
	assert.False(t, ok, "Expired entry should be removed")

	_, _, ok = cache.GetMeasurement(text, style)
	assert.False(t, ok, "Expired measurement should be removed")
}

// TestCacheLRUEviction tests LRU eviction when cache is full
func TestCacheLRUEviction(t *testing.T) {
	cache := NewCache(CacheConfig{
		MaxSize:     3, // Small cache to trigger eviction
		TTL:         1 * time.Minute,
		EnableStats: false,
	})

	// Fill the cache with 5 distinct styles
	styles := []lipgloss.Style{
		lipgloss.NewStyle().Foreground(lipgloss.Color("0")),   // First entry
		lipgloss.NewStyle().Foreground(lipgloss.Color("1")),   // Second entry
		lipgloss.NewStyle().Foreground(lipgloss.Color("2")),   // Third entry
		lipgloss.NewStyle().Foreground(lipgloss.Color("3")),   // Fourth - should evict first
		lipgloss.NewStyle().Foreground(lipgloss.Color("4")),   // Fifth - should evict second
	}

	for _, style := range styles {
		cache.SetStyle(style, runtime.CellStyle{})
	}

	// Cache should only have 3 entries (max size)
	// After 5 inserts, the first 2 should be evicted, leaving styles 2, 3, 4
	assert.Equal(t, 3, len(cache.styles), "Cache should respect max size")
}

// TestGlobalCache tests global cache instance
func TestGlobalCache(t *testing.T) {
	cache1 := GetGlobalCache()
	cache2 := GetGlobalCache()

	assert.Same(t, cache1, cache2, "Should return same instance")
	assert.NotNil(t, cache1)
}

// TestSetGlobalCache tests setting custom global cache
func TestSetGlobalCache(t *testing.T) {
	original := GetGlobalCache()

	customCache := NewCache(CacheConfig{
		MaxSize:     500,
		TTL:         10 * time.Minute,
		EnableStats: true,
	})

	SetGlobalCache(customCache)

	result := GetGlobalCache()
	assert.Same(t, customCache, result, "Should return custom cache")

	// Restore original
	SetGlobalCache(original)
}

// TestLipglossToCellStyleCached tests cached style conversion
func TestLipglossToCellStyleCached(t *testing.T) {
	// Clear global cache first
	GetGlobalCache().Clear()

	style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("red"))

	// First call - cache miss
	result1 := LipglossToCellStyleCached(style)
	assert.True(t, result1.Bold)

	// Second call - cache hit
	result2 := LipglossToCellStyleCached(style)
	assert.Equal(t, result1, result2)
}

// TestMeasureLipglossTextCached tests cached text measurement
func TestMeasureLipglossTextCached(t *testing.T) {
	// Clear global cache first
	GetGlobalCache().Clear()

	style := lipgloss.NewStyle()
	text := "Hello, World!"

	// First call - cache miss
	result1 := MeasureLipglossTextCached(text, style)
	assert.Equal(t, 13, result1)

	// Second call - cache hit
	result2 := MeasureLipglossTextCached(text, style)
	assert.Equal(t, result1, result2)
}

// TestMeasureLipglossTextHeightCached tests cached height measurement
func TestMeasureLipglossTextHeightCached(t *testing.T) {
	// Clear global cache first
	GetGlobalCache().Clear()

	style := lipgloss.NewStyle()
	text := "Line 1\nLine 2\nLine 3"

	// First call - cache miss
	result1 := MeasureLipglossTextHeightCached(text, style)
	assert.Equal(t, 3, result1)

	// Second call - cache hit
	result2 := MeasureLipglossTextHeightCached(text, style)
	assert.Equal(t, result1, result2)
}

// TestRenderLipglossToBufferCached tests cached rendering
func TestRenderLipglossToBufferCached(t *testing.T) {
	// Clear global cache first
	GetGlobalCache().Clear()

	buf := runtime.NewCellBuffer(20, 5)
	text := "Hello"
	style := lipgloss.NewStyle().Bold(true)

	// First call - cache miss
	RenderLipglossToBufferCached(buf, text, style, 0, 0, 0)
	cell := buf.GetCell(0, 0)
	assert.Equal(t, 'H', cell.Char)
	assert.True(t, cell.Style.Bold)

	// Clear buffer and render again - should use cache
	buf2 := runtime.NewCellBuffer(20, 5)
	RenderLipglossToBufferCached(buf2, text, style, 0, 0, 0)
	cell2 := buf2.GetCell(0, 0)
	assert.Equal(t, 'H', cell2.Char)
	assert.True(t, cell2.Style.Bold)
}

// TestStyleSignature tests style signature generation
func TestStyleSignature(t *testing.T) {
	style1 := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("red"))
	style2 := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("red"))
	style3 := lipgloss.NewStyle().Bold(false).Foreground(lipgloss.Color("red"))

	sig1 := styleSignature(style1)
	sig2 := styleSignature(style2)
	sig3 := styleSignature(style3)

	assert.Equal(t, sig1, sig2, "Same styles should have same signature")
	assert.NotEqual(t, sig1, sig3, "Different styles should have different signatures")
}

// TestTextStyleSignature tests text+style signature generation
func TestTextStyleSignature(t *testing.T) {
	style := lipgloss.NewStyle().Bold(true)
	text1 := "Hello"
	text2 := "World"

	sig1 := textStyleSignature(text1, style)
	sig2 := textStyleSignature(text2, style)

	assert.NotEqual(t, sig1, sig2, "Different texts should have different signatures")
}

// BenchmarkLipglossToCellStyle benchmarks uncached vs cached style conversion
func BenchmarkLipglossToCellStyle(b *testing.B) {
	style := lipgloss.NewStyle().
		Bold(true).
		Italic(true).
		Underline(true).
		Foreground(lipgloss.Color("#FF5733")).
		Background(lipgloss.Color("#33FF57"))

	b.Run("Uncached", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = LipglossToCellStyle(style)
		}
	})

	b.Run("Cached", func(b *testing.B) {
		GetGlobalCache().Clear()
		for i := 0; i < b.N; i++ {
			_ = LipglossToCellStyleCached(style)
		}
	})
}

// BenchmarkMeasureLipglossText benchmarks uncached vs cached measurement
func BenchmarkMeasureLipglossText(b *testing.B) {
	style := lipgloss.NewStyle().Bold(true)
	text := "Hello, World! This is a test string."

	b.Run("Uncached", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = MeasureLipglossText(text, style)
		}
	})

	b.Run("Cached", func(b *testing.B) {
		GetGlobalCache().Clear()
		for i := 0; i < b.N; i++ {
			_ = MeasureLipglossTextCached(text, style)
		}
	})
}
