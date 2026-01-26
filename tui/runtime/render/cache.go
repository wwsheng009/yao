package render

import (
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/runtime"
)

// Cache manages cached rendering data to improve performance
// It provides thread-safe caching for style conversions, measurements, and rendered text
type Cache struct {
	mu sync.RWMutex

	// Style cache: lipgloss.Style signature -> runtime.CellStyle
	styles map[string]*cachedStyle

	// Measurement cache: text+style signature -> dimensions
	measurements map[string]*cachedMeasurement

	// ANSI parse cache: styled text -> segments
	ansiSegments map[string][]StyledSegment

	// Rendered text cache: text+style -> rendered output
	renderedText map[string]string

	// Statistics
	hits   int
	misses int

	// Configuration
	maxSize     int
	ttl         time.Duration
	enableStats bool
}

// cachedStyle represents a cached CellStyle with expiration
type cachedStyle struct {
	value     runtime.CellStyle
	timestamp time.Time
	accesses  int
}

// cachedMeasurement represents cached text dimensions
type cachedMeasurement struct {
	width  int
	height int
	timestamp time.Time
	accesses  int
}

// CacheConfig configures the cache behavior
type CacheConfig struct {
	MaxSize     int           // Maximum number of entries per cache (0 = unlimited)
	TTL         time.Duration // Time-to-live for cache entries (0 = no expiration)
	EnableStats bool          // Whether to track cache statistics
}

// DefaultCacheConfig returns the default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MaxSize:     1000,
		TTL:         5 * time.Minute,
		EnableStats: false,
	}
}

// NewCache creates a new render cache
func NewCache(config CacheConfig) *Cache {
	if config.MaxSize == 0 {
		config.MaxSize = 1000 // Default to prevent unbounded growth
	}
	if config.TTL == 0 {
		config.TTL = 5 * time.Minute // Default TTL
	}

	return &Cache{
		styles:       make(map[string]*cachedStyle),
		measurements: make(map[string]*cachedMeasurement),
		ansiSegments: make(map[string][]StyledSegment),
		renderedText: make(map[string]string),
		maxSize:      config.MaxSize,
		ttl:          config.TTL,
		enableStats:  config.EnableStats,
	}
}

// globalCache is the default shared cache instance
var globalCache *Cache
var globalCacheOnce sync.Once

// isCacheDisabled checks if cache is disabled via environment variable
// Set TUI_DISABLE_CACHE=1 environment variable to disable
func isCacheDisabled() bool {
	return os.Getenv("TUI_DISABLE_CACHE") == "1"
}

// GetGlobalCache returns the global cache instance (lazy initialization)
func GetGlobalCache() *Cache {
	globalCacheOnce.Do(func() {
		globalCache = NewCache(DefaultCacheConfig())
	})
	return globalCache
}

// SetGlobalCache sets a custom global cache
func SetGlobalCache(cache *Cache) {
	globalCache = cache
}

// GetStyle retrieves a cached CellStyle for the given lipgloss.Style
func (c *Cache) GetStyle(style lipgloss.Style) (runtime.CellStyle, bool) {
	if isCacheDisabled() {
		return runtime.CellStyle{}, false // Cache disabled, always miss
	}
	key := styleSignature(style)

	c.mu.RLock()
	entry, exists := c.styles[key]
	c.mu.RUnlock()

	if !exists {
		if c.enableStats {
			c.mu.Lock()
			c.misses++
			c.mu.Unlock()
		}
		return runtime.CellStyle{}, false
	}

	// Check TTL
	if c.ttl > 0 && time.Since(entry.timestamp) > c.ttl {
		c.mu.Lock()
		delete(c.styles, key)
		c.mu.Unlock()
		if c.enableStats {
			c.mu.Lock()
			c.misses++
			c.mu.Unlock()
		}
		return runtime.CellStyle{}, false
	}

	if c.enableStats {
		c.mu.Lock()
		c.hits++
		entry.accesses++
		c.mu.Unlock()
	}

	return entry.value, true
}

// SetStyle caches a CellStyle for the given lipgloss.Style
func (c *Cache) SetStyle(style lipgloss.Style, cellStyle runtime.CellStyle) {
	key := styleSignature(style)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict if at capacity
	if len(c.styles) >= c.maxSize {
		c.evictStyleLRU()
	}

	c.styles[key] = &cachedStyle{
		value:     cellStyle,
		timestamp: time.Now(),
		accesses:  1,
	}
}

// GetMeasurement retrieves cached text dimensions
func (c *Cache) GetMeasurement(text string, style lipgloss.Style) (width, height int, ok bool) {
	if isCacheDisabled() {
		return 0, 0, false // Cache disabled, always miss
	}
	key := textStyleSignature(text, style)

	c.mu.RLock()
	entry, exists := c.measurements[key]
	c.mu.RUnlock()

	if !exists {
		if c.enableStats {
			c.mu.Lock()
			c.misses++
			c.mu.Unlock()
		}
		return 0, 0, false
	}

	// Check TTL
	if c.ttl > 0 && time.Since(entry.timestamp) > c.ttl {
		c.mu.Lock()
		delete(c.measurements, key)
		c.mu.Unlock()
		if c.enableStats {
			c.mu.Lock()
			c.misses++
			c.mu.Unlock()
		}
		return 0, 0, false
	}

	if c.enableStats {
		c.mu.Lock()
		c.hits++
		entry.accesses++
		c.mu.Unlock()
	}

	return entry.width, entry.height, true
}

// SetMeasurement caches text dimensions
func (c *Cache) SetMeasurement(text string, style lipgloss.Style, width, height int) {
	key := textStyleSignature(text, style)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict if at capacity
	if len(c.measurements) >= c.maxSize {
		c.evictMeasurementLRU()
	}

	c.measurements[key] = &cachedMeasurement{
		width:     width,
		height:    height,
		timestamp: time.Now(),
		accesses:  1,
	}
}

// GetANSISegments retrieves cached ANSI segments
func (c *Cache) GetANSISegments(line string) ([]StyledSegment, bool) {
	if isCacheDisabled() {
		return nil, false // Cache disabled, always miss
	}
	c.mu.RLock()
	entry, exists := c.ansiSegments[line]
	c.mu.RUnlock()

	if !exists {
		if c.enableStats {
			c.mu.Lock()
			c.misses++
			c.mu.Unlock()
		}
		return nil, false
	}

	if c.enableStats {
		c.mu.Lock()
		c.hits++
		c.mu.Unlock()
	}

	return entry, true
}

// SetANSISegments caches ANSI segments
func (c *Cache) SetANSISegments(line string, segments []StyledSegment) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict if at capacity
	if len(c.ansiSegments) >= c.maxSize {
		c.evictANSISegmentsLRU()
	}

	c.ansiSegments[line] = segments
}

// GetRenderedText retrieves cached rendered text
func (c *Cache) GetRenderedText(text string, style lipgloss.Style) (string, bool) {
	if isCacheDisabled() {
		return "", false // Cache disabled, always miss
	}
	key := textStyleSignature(text, style)

	c.mu.RLock()
	entry, exists := c.renderedText[key]
	c.mu.RUnlock()

	if !exists {
		if c.enableStats {
			c.mu.Lock()
			c.misses++
			c.mu.Unlock()
		}
		return "", false
	}

	if c.enableStats {
		c.mu.Lock()
		c.hits++
		c.mu.Unlock()
	}

	return entry, true
}

// SetRenderedText caches rendered text
func (c *Cache) SetRenderedText(text string, style lipgloss.Style, rendered string) {
	key := textStyleSignature(text, style)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict if at capacity
	if len(c.renderedText) >= c.maxSize {
		c.evictRenderedTextLRU()
	}

	c.renderedText[key] = rendered
}

// Stats returns cache statistics
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalRequests := c.hits + c.misses
	hitRate := 0.0
	if totalRequests > 0 {
		hitRate = float64(c.hits) / float64(totalRequests)
	}

	return CacheStats{
		StyleCount:       len(c.styles),
		MeasurementCount: len(c.measurements),
		ANSISegmentCount: len(c.ansiSegments),
		RenderedTextCount: len(c.renderedText),
		Hits:             c.hits,
		Misses:           c.misses,
		HitRate:          hitRate,
	}
}

// CacheStats holds cache statistics
type CacheStats struct {
	StyleCount        int
	MeasurementCount  int
	ANSISegmentCount  int
	RenderedTextCount int
	Hits              int
	Misses            int
	HitRate           float64
}

// Clear clears all cache entries
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.styles = make(map[string]*cachedStyle)
	c.measurements = make(map[string]*cachedMeasurement)
	c.ansiSegments = make(map[string][]StyledSegment)
	c.renderedText = make(map[string]string)
	c.hits = 0
	c.misses = 0
}

// Evict expired entries from all caches
func (c *Cache) EvictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ttl <= 0 {
		return // No expiration configured
	}

	now := time.Now()

	// Evict expired styles
	for key, entry := range c.styles {
		if now.Sub(entry.timestamp) > c.ttl {
			delete(c.styles, key)
		}
	}

	// Evict expired measurements
	for key, entry := range c.measurements {
		if now.Sub(entry.timestamp) > c.ttl {
			delete(c.measurements, key)
		}
	}
}

// evictStyleLRU evicts the least recently used style entry
func (c *Cache) evictStyleLRU() {
	var oldestKey string
	var oldestTime time.Time
	var oldestAccesses int

	for key, entry := range c.styles {
		if oldestKey == "" || entry.accesses < oldestAccesses ||
			(entry.accesses == oldestAccesses && entry.timestamp.Before(oldestTime)) {
			oldestKey = key
			oldestTime = entry.timestamp
			oldestAccesses = entry.accesses
		}
	}

	if oldestKey != "" {
		delete(c.styles, oldestKey)
	}
}

// evictMeasurementLRU evicts the least recently used measurement entry
func (c *Cache) evictMeasurementLRU() {
	var oldestKey string
	var oldestTime time.Time
	var oldestAccesses int

	for key, entry := range c.measurements {
		if oldestKey == "" || entry.accesses < oldestAccesses ||
			(entry.accesses == oldestAccesses && entry.timestamp.Before(oldestTime)) {
			oldestKey = key
			oldestTime = entry.timestamp
			oldestAccesses = entry.accesses
		}
	}

	if oldestKey != "" {
		delete(c.measurements, oldestKey)
	}
}

// evictANSISegmentsLRU evicts an ANSI segment entry (FIFO order)
func (c *Cache) evictANSISegmentsLRU() {
	// Simple FIFO eviction for ANSI segments
	for key := range c.ansiSegments {
		delete(c.ansiSegments, key)
		return
	}
}

// evictRenderedTextLRU evicts a rendered text entry (FIFO order)
func (c *Cache) evictRenderedTextLRU() {
	// Simple FIFO eviction for rendered text
	for key := range c.renderedText {
		delete(c.renderedText, key)
		return
	}
}

// styleSignature creates a unique signature for a lipgloss.Style
func styleSignature(style lipgloss.Style) string {
	// Create a signature from style properties
	// This is a simplified version - could be more sophisticated
	sig := ""
	if fg := style.GetForeground(); fg != nil {
		if color, ok := fg.(lipgloss.Color); ok {
			sig += "fg:" + string(color) + ";"
		}
	}
	if bg := style.GetBackground(); bg != nil {
		if color, ok := bg.(lipgloss.Color); ok {
			sig += "bg:" + string(color) + ";"
		}
	}
	if style.GetBold() {
		sig += "bold;"
	}
	if style.GetItalic() {
		sig += "italic;"
	}
	if style.GetUnderline() {
		sig += "underline;"
	}
	if style.GetStrikethrough() {
		sig += "strikethrough;"
	}
	if style.GetBlink() {
		sig += "blink;"
	}
	if style.GetReverse() {
		sig += "reverse;"
	}

	return sig
}

// textStyleSignature creates a unique signature for text+style combination
func textStyleSignature(text string, style lipgloss.Style) string {
	return styleSignature(style) + "text:" + text
}
