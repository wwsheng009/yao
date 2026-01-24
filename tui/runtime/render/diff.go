package render

import (
	"github.com/yaoapp/yao/tui/runtime"
)

// DiffResult contains the difference between two frames
type DiffResult struct {
	DirtyRegions []runtime.Rect // Regions that have changed
	HasChanges   bool           // True if any changes detected
	ChangedCells int            // Count of changed cells
}

// ComputeDiff compares two frames and returns the difference
// This identifies dirty regions that need to be re-rendered
func ComputeDiff(oldFrame, newFrame runtime.Frame) DiffResult {
	result := DiffResult{
		DirtyRegions: []runtime.Rect{},
		HasChanges:   false,
		ChangedCells: 0,
	}

	// Handle edge cases
	if oldFrame.Buffer == nil && newFrame.Buffer == nil {
		return result
	}

	if newFrame.Buffer == nil {
		// New frame is empty (clear screen)
		result.HasChanges = true
		result.DirtyRegions = append(result.DirtyRegions, runtime.Rect{
			X:      0,
			Y:      0,
			Width:  oldFrame.Width,
			Height: oldFrame.Height,
		})
		return result
	}

	if oldFrame.Buffer == nil {
		// Old frame was empty (entire screen is new)
		result.HasChanges = true
		result.DirtyRegions = append(result.DirtyRegions, runtime.Rect{
			X:      0,
			Y:      0,
			Width:  newFrame.Width,
			Height: newFrame.Height,
		})
		return result
	}

	// Get dimensions
	minWidth := min(oldFrame.Width, newFrame.Width)
	minHeight := min(oldFrame.Height, newFrame.Height)
	maxWidth := max(oldFrame.Width, newFrame.Width)
	maxHeight := max(oldFrame.Height, newFrame.Height)

	// Track dirty cells in a grid for region merging
	// Use max dimensions to capture extra area when size changes
	dirtyGrid := newDirtyGrid(maxWidth, maxHeight)

	// Compare cells within the overlapping area
	for y := 0; y < minHeight; y++ {
		for x := 0; x < minWidth; x++ {
			oldCell := oldFrame.Buffer.GetCell(x, y)
			newCell := newFrame.Buffer.GetCell(x, y)

			if !cellsEqual(oldCell, newCell) {
				dirtyGrid.Mark(x, y)
				result.ChangedCells++
			}
		}
	}

	// Handle size differences
	if newFrame.Width > oldFrame.Width {
		// New width is larger - mark extra area as dirty
		for y := 0; y < min(newFrame.Height, maxHeight); y++ {
			for x := oldFrame.Width; x < newFrame.Width; x++ {
				dirtyGrid.Mark(x, y)
				result.ChangedCells++
			}
		}
	}

	if newFrame.Height > oldFrame.Height {
		// New height is larger - mark extra area as dirty
		for y := oldFrame.Height; y < newFrame.Height; y++ {
			for x := 0; x < newFrame.Width; x++ {
				dirtyGrid.Mark(x, y)
				result.ChangedCells++
			}
		}
	}

	// Extract dirty regions from the grid
	result.DirtyRegions = dirtyGrid.ExtractRegions()
	result.HasChanges = len(result.DirtyRegions) > 0

	return result
}

// cellsEqual compares two cells for equality
func cellsEqual(a, b runtime.Cell) bool {
	if a.Char != b.Char {
		return false
	}

	// Compare styles
	aStyle := a.Style
	bStyle := b.Style

	if aStyle.Bold != bStyle.Bold ||
		aStyle.Italic != bStyle.Italic ||
		aStyle.Underline != bStyle.Underline ||
		aStyle.Strikethrough != bStyle.Strikethrough ||
		aStyle.Blink != bStyle.Blink ||
		aStyle.Reverse != bStyle.Reverse {
		return false
	}

	// Compare colors
	if aStyle.Foreground != bStyle.Foreground ||
		aStyle.Background != bStyle.Background {
		return false
	}

	return true
}

// dirtyGrid tracks dirty cells for region extraction
type dirtyGrid struct {
	width  int
	height int
	cells  [][]bool
}

// newDirtyGrid creates a new dirty grid
func newDirtyGrid(width, height int) *dirtyGrid {
	cells := make([][]bool, height)
	for i := range cells {
		cells[i] = make([]bool, width)
	}

	return &dirtyGrid{
		width:  width,
		height: height,
		cells:  cells,
	}
}

// Mark marks a cell as dirty
func (dg *dirtyGrid) Mark(x, y int) {
	if x >= 0 && x < dg.width && y >= 0 && y < dg.height {
		dg.cells[y][x] = true
	}
}

// ExtractRegions extracts dirty regions from the grid
// Merges adjacent dirty cells into larger regions for efficiency
func (dg *dirtyGrid) ExtractRegions() []runtime.Rect {
	if dg.width == 0 || dg.height == 0 {
		return []runtime.Rect{}
	}

	visited := make([][]bool, dg.height)
	for i := range visited {
		visited[i] = make([]bool, dg.width)
	}

	var regions []runtime.Rect

	for y := 0; y < dg.height; y++ {
		for x := 0; x < dg.width; x++ {
			if dg.cells[y][x] && !visited[y][x] {
				// Found a new unvisited dirty cell
				region := dg.extractRegion(x, y, visited)
				regions = append(regions, region)
			}
		}
	}

	// Merge overlapping or adjacent regions
	return mergeRegions(regions)
}

// extractRegion extracts a single contiguous dirty region using flood fill
func (dg *dirtyGrid) extractRegion(startX, startY int, visited [][]bool) runtime.Rect {
	minX, minY := startX, startY
	maxX, maxY := startX, startY

	// Flood fill using stack
	stack := []struct{ x, y int }{{startX, startY}}
	visited[startY][startX] = true

	for len(stack) > 0 {
		// Pop
		pos := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		x, y := pos.x, pos.y

		// Update bounds
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}

		// Check neighbors (4-directional)
		neighbors := []struct{ x, y int }{
			{x - 1, y}, // left
			{x + 1, y}, // right
			{x, y - 1}, // up
			{x, y + 1}, // down
		}

		for _, n := range neighbors {
			nx, ny := n.x, n.y

			// Check bounds
			if nx >= 0 && nx < dg.width && ny >= 0 && ny < dg.height {
				// Check if dirty and not visited
				if dg.cells[ny][nx] && !visited[ny][nx] {
					visited[ny][nx] = true
					stack = append(stack, struct{ x, y int }{nx, ny})
				}
			}
		}
	}

	return runtime.Rect{
		X:      minX,
		Y:      minY,
		Width:  maxX - minX + 1,
		Height: maxY - minY + 1,
	}
}

// mergeRegions merges overlapping or adjacent regions
func mergeRegions(regions []runtime.Rect) []runtime.Rect {
	if len(regions) <= 1 {
		return regions
	}

	// Simple greedy merge: merge regions that are close to each other
	merged := true
	for merged {
		merged = false
		for i := 0; i < len(regions); i++ {
			for j := i + 1; j < len(regions); j++ {
				if shouldMerge(regions[i], regions[j]) {
					// Merge regions i and j
					regions[i] = mergeRect(regions[i], regions[j])
					// Remove region j
					regions = append(regions[:j], regions[j+1:]...)
					merged = true
					break
				}
			}
			if merged {
				break
			}
		}
	}

	return regions
}

// shouldMerge checks if two regions should be merged
// Merges if regions are adjacent or overlapping (within 1 cell)
func shouldMerge(a, b runtime.Rect) bool {
	// Check for overlap
	overlapX := a.X < b.X+b.Width && a.X+a.Width > b.X
	overlapY := a.Y < b.Y+b.Height && a.Y+a.Height > b.Y

	if overlapX && overlapY {
		return true
	}

	// Check for adjacency (within 1 cell)
	adjacentX := a.X <= b.X+b.Width+1 && a.X+a.Width >= b.X-1
	adjacentY := a.Y <= b.Y+b.Height+1 && a.Y+a.Height >= b.Y-1

	return adjacentX && adjacentY
}

// mergeRect merges two rectangles into a bounding box
func mergeRect(a, b runtime.Rect) runtime.Rect {
	minX := min(a.X, b.X)
	minY := min(a.Y, b.Y)
	maxX := max(a.X+a.Width, b.X+b.Width)
	maxY := max(a.Y+a.Height, b.Y+b.Height)

	return runtime.Rect{
		X:      minX,
		Y:      minY,
		Width:  maxX - minX,
		Height: maxY - minY,
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RenderWithDiff renders only the dirty regions from a diff
// This is useful for incremental rendering optimization
func RenderWithDiff(buf *runtime.CellBuffer, frame runtime.Frame, diff DiffResult) {
	if buf == nil || frame.Buffer == nil {
		return
	}

	bufWidth := buf.Width()
	bufHeight := buf.Height()

	// For each dirty region, copy cells from new frame to buffer
	for _, region := range diff.DirtyRegions {
		for y := region.Y; y < region.Y+region.Height; y++ {
			for x := region.X; x < region.X+region.Width; x++ {
				if x >= 0 && x < frame.Width && y >= 0 && y < frame.Height {
					if x < bufWidth && y < bufHeight {
						cell := frame.Buffer.GetCell(x, y)
						buf.SetContent(x, y, cell.ZIndex, cell.Char, cell.Style, cell.NodeID)
					}
				}
			}
		}
	}
}

// OptimizeFrame optimizes a frame by applying dirty region tracking
// Returns the optimized frame with only changed regions marked
func OptimizeFrame(prevFrame, newFrame runtime.Frame) runtime.Frame {
	// Compute diff between frames
	diff := ComputeDiff(prevFrame, newFrame)

	// Mark dirty regions in the new frame
	newFrame.Dirty = diff.HasChanges

	return newFrame
}

// GetChangedCellsCount returns the number of cells that changed between two frames
func GetChangedCellsCount(oldFrame, newFrame runtime.Frame) int {
	diff := ComputeDiff(oldFrame, newFrame)
	return diff.ChangedCells
}

// ShouldRerender determines if a frame should be re-rendered based on dirty threshold
// Returns true if more than threshold% of cells have changed
func ShouldRerender(prevFrame, newFrame runtime.Frame, threshold float64) bool {
	if prevFrame.Buffer == nil || newFrame.Buffer == nil {
		return true
	}

	totalCells := newFrame.Width * newFrame.Height
	if totalCells == 0 {
		return false
	}

	changedCells := GetChangedCellsCount(prevFrame, newFrame)
	changedRatio := float64(changedCells) / float64(totalCells)

	return changedRatio >= threshold
}
