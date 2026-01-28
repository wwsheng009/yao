package paint

import (
	"bytes"
	"time"
)

// RemoteOptimizer optimizes rendering for remote terminals
type RemoteOptimizer struct {
	frameBuffer    *bytes.Buffer
	lastFlush      time.Time
	frameInterval  time.Duration
	deltaEncoding  bool
	lastFrame      []byte
	maxBufferSize  int
}

// NewRemoteOptimizer creates a new remote optimizer
func NewRemoteOptimizer() *RemoteOptimizer {
	return &RemoteOptimizer{
		frameBuffer:   &bytes.Buffer{},
		frameInterval: 16 * time.Millisecond, // ~60 FPS
		deltaEncoding: true,
		maxBufferSize: 65536, // 64KB max buffer
	}
}

// NewRemoteOptimizerWithInterval creates a new remote optimizer with custom interval
func NewRemoteOptimizerWithInterval(interval time.Duration) *RemoteOptimizer {
	return &RemoteOptimizer{
		frameBuffer:   &bytes.Buffer{},
		frameInterval: interval,
		deltaEncoding: true,
		maxBufferSize: 65536,
	}
}

// BufferFrame buffers a frame for later flushing
func (r *RemoteOptimizer) BufferFrame(data []byte) error {
	if r.frameBuffer.Len()+len(data) > r.maxBufferSize {
		// Buffer too full, flush first
		r.Flush()
	}
	_, err := r.frameBuffer.Write(data)
	return err
}

// ShouldFlush returns true if enough time has passed to flush
func (r *RemoteOptimizer) ShouldFlush() bool {
	return time.Since(r.lastFlush) >= r.frameInterval ||
		r.frameBuffer.Len() > 4096 // Flush if buffer has > 4KB
}

// Flush flushes the buffered frame data
func (r *RemoteOptimizer) Flush() []byte {
	if r.frameBuffer.Len() == 0 {
		return nil
	}

	data := r.frameBuffer.Bytes()
	result := make([]byte, len(data))
	copy(result, data)

	r.frameBuffer.Reset()
	r.lastFlush = time.Now()

	return result
}

// FlushWithDelta flushes the buffered frame data with delta encoding
func (r *RemoteOptimizer) FlushWithDelta() []byte {
	if r.frameBuffer.Len() == 0 {
		return nil
	}

	current := r.frameBuffer.Bytes()
	var result []byte

	if r.deltaEncoding && len(r.lastFrame) > 0 {
		result = r.EncodeDelta(r.lastFrame, current)
	} else {
		result = make([]byte, len(current))
		copy(result, current)
	}

	r.lastFrame = make([]byte, len(current))
	copy(r.lastFrame, current)

	r.frameBuffer.Reset()
	r.lastFlush = time.Now()

	return result
}

// EncodeDelta encodes the difference between two frames
func (r *RemoteOptimizer) EncodeDelta(prev, curr []byte) []byte {
	if len(prev) == 0 {
		return curr
	}

	// Simple delta encoding: find common prefix and suffix
	prefixLen := 0
	maxPrefix := min(len(prev), len(curr))
	for prefixLen < maxPrefix && prev[prefixLen] == curr[prefixLen] {
		prefixLen++
	}

	suffixLen := 0
	maxSuffix := min(len(prev)-prefixLen, len(curr)-prefixLen)
	for suffixLen < maxSuffix &&
		prev[len(prev)-1-suffixLen] == curr[len(curr)-1-suffixLen] {
		suffixLen++
	}

	// If most of the frame changed, send full frame
	changedLen := len(curr) - prefixLen - suffixLen
	if changedLen > len(curr)/2 {
		return curr
	}

	// Otherwise, send delta
	// Format: [prefix_len][changed_data][suffix_len]
	var delta bytes.Buffer
	delta.WriteByte(byte(prefixLen))
	delta.Write(curr[prefixLen : len(curr)-suffixLen])
	delta.WriteByte(byte(suffixLen))

	return delta.Bytes()
}

// DecodeDelta decodes a delta frame
func (r *RemoteOptimizer) DecodeDelta(delta, prev []byte) []byte {
	if len(delta) < 2 {
		return delta
	}

	prefixLen := int(delta[0])
	suffixLen := int(delta[len(delta)-1])
	changedData := delta[1 : len(delta)-1]

	// Reconstruct the frame
	result := make([]byte, prefixLen+len(changedData)+suffixLen)

	// Copy prefix from previous frame
	if prefixLen > 0 && prefixLen <= len(prev) {
		copy(result, prev[:prefixLen])
	}

	// Copy changed data
	copy(result[prefixLen:], changedData)

	// Copy suffix from previous frame
	if suffixLen > 0 && len(prev) >= suffixLen {
		copy(result[prefixLen+len(changedData):], prev[len(prev)-suffixLen:])
	}

	return result
}

// SetFrameInterval sets the minimum interval between frames
func (r *RemoteOptimizer) SetFrameInterval(interval time.Duration) {
	r.frameInterval = interval
}

// GetFrameInterval returns the current frame interval
func (r *RemoteOptimizer) GetFrameInterval() time.Duration {
	return r.frameInterval
}

// EnableDeltaEncoding enables or disables delta encoding
func (r *RemoteOptimizer) EnableDeltaEncoding(enabled bool) {
	r.deltaEncoding = enabled
}

// SetMaxBufferSize sets the maximum buffer size
func (r *RemoteOptimizer) SetMaxBufferSize(size int) {
	r.maxBufferSize = size
}

// BufferSize returns the current buffer size
func (r *RemoteOptimizer) BufferSize() int {
	return r.frameBuffer.Len()
}

// Clear clears the buffer
func (r *RemoteOptimizer) Clear() {
	r.frameBuffer.Reset()
}

// Reset resets the optimizer state
func (r *RemoteOptimizer) Reset() {
	r.frameBuffer.Reset()
	r.lastFrame = nil
	r.lastFlush = time.Time{}
}

// Stats returns statistics about the optimizer
func (r *RemoteOptimizer) Stats() RemoteStats {
	return RemoteStats{
		BufferSize:    r.frameBuffer.Len(),
		LastFlushAge:  time.Since(r.lastFlush),
		DeltaEncoding: r.deltaEncoding,
		FrameInterval: r.frameInterval,
	}
}

// RemoteStats contains statistics about the remote optimizer
type RemoteStats struct {
	BufferSize    int
	LastFlushAge  time.Duration
	DeltaEncoding bool
	FrameInterval time.Duration
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
