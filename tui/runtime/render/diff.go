package render

import (
	"fmt"
)

// Rect represents a rectangle region
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

// FrameBuffer defines the interface for accessing frame buffer data
// This avoids circular dependency with runtime package
type FrameBuffer interface {
	GetContent(x, y int) Cell
	Width() int
	Height() int
}

// Cell represents a single cell with content and style
type Cell struct {
	Char  rune
	Style CellStyle
}

// CellStyle represents rendering style for a cell
type CellStyle struct {
	Bold      bool
	Underline bool
	Italic    bool
}

// DiffResult represents the difference between two frames
type DiffResult struct {
	DirtyRegions []Rect
	HasChanges   bool
}

// Frame represents a frame with buffer and dimensions
// This avoids importing runtime.Frame
type Frame struct {
	Buffer FrameBuffer
	Width  int
	Height int
}

// ComputeDiff compares two frames and returns the difference
// It identifies regions that have changed between frames
func ComputeDiff(oldFrame, newFrame Frame) DiffResult {
	result := DiffResult{
		DirtyRegions: []Rect{},
		HasChanges:   false,
	}

	// If old frame is nil or empty, entire new frame is dirty
	if oldFrame.Buffer == nil {
		result.DirtyRegions = append(result.DirtyRegions, Rect{
			X:      0,
			Y:      0,
			Width:  newFrame.Width,
			Height: newFrame.Height,
		})
		result.HasChanges = true
		return result
	}

	// Compare dimensions
	if oldFrame.Width != newFrame.Width || oldFrame.Height != newFrame.Height {
		// Dimensions changed, mark entire frame as dirty
		result.DirtyRegions = append(result.DirtyRegions, Rect{
			X:      0,
			Y:      0,
			Width:  newFrame.Width,
			Height: newFrame.Height,
		})
		result.HasChanges = true
		return result
	}

	// Compare cell by cell and collect dirty regions
	oldBuf := oldFrame.Buffer
	newBuf := newFrame.Buffer

	dirtyCells := make(map[string]bool)

	for y := 0; y < newFrame.Height; y++ {
		for x := 0; x < newFrame.Width; x++ {
			oldCell := oldBuf.GetContent(x, y)
			newCell := newBuf.GetContent(x, y)

			// Check if cell changed
			if oldCell.Char != newCell.Char ||
				oldCell.Style.Bold != newCell.Style.Bold ||
				oldCell.Style.Underline != newCell.Style.Underline ||
				oldCell.Style.Italic != newCell.Style.Italic {
				dirtyCells[cellKey(x, y)] = true
			}
		}
	}

	// If no changes, return empty result
	if len(dirtyCells) == 0 {
		return result
	}

	result.HasChanges = true

	// Merge adjacent dirty cells into regions
	regions := mergeDirtyCells(dirtyCells, newFrame.Width, newFrame.Height)
	result.DirtyRegions = regions

	return result
}

// cellKey creates a unique key for a cell position
func cellKey(x, y int) string {
	return fmt.Sprintf("%d,%d", x, y)
}

// mergeDirtyCells merges adjacent dirty cells into rectangular regions
func mergeDirtyCells(dirtyCells map[string]bool, width, height int) []Rect {
	regions := []Rect{}
	visited := make(map[string]bool)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			key := cellKey(x, y)
			if !dirtyCells[key] || visited[key] {
				continue
			}

			// Found an unvisited dirty cell, grow a region from it
			region := growRegion(dirtyCells, visited, x, y, width, height)
			regions = append(regions, Rect{
				X:      region.X,
				Y:      region.Y,
				Width:  region.Width,
				Height: region.Height,
			})
		}
	}

	return regions
}

// growRegion grows a rectangular region from a starting dirty cell
func growRegion(dirtyCells, visited map[string]bool, startX, startY, width, height int) Rect {
	// Start with a 1x1 region
	region := Rect{
		X:      startX,
		Y:      startY,
		Width:  1,
		Height: 1,
	}

	// Mark starting cell as visited
	visited[cellKey(startX, startY)] = true

	// Try to expand region to the right
	for x := startX + region.Width; x < width; x++ {
		canExpand := true
		for y := region.Y; y < region.Y+region.Height; y++ {
			key := cellKey(x, y)
			if !dirtyCells[key] {
				canExpand = false
				break
			}
		}
		if !canExpand {
			break
		}
		region.Width++
		// Mark new cells as visited
		for y := region.Y; y < region.Y+region.Height; y++ {
			visited[cellKey(x, y)] = true
		}
	}

	// Try to expand region downward
	for y := startY + region.Height; y < height; y++ {
		canExpand := true
		for x := region.X; x < region.X+region.Width; x++ {
			key := cellKey(x, y)
			if !dirtyCells[key] {
				canExpand = false
				break
			}
		}
		if !canExpand {
			break
		}
		region.Height++
		// Mark new cells as visited
		for x := region.X; x < region.X+region.Width; x++ {
			visited[cellKey(x, y)] = true
		}
	}

	return region
}

// IsEmptyDiff checks if a diff result is empty (no changes)
func IsEmptyDiff(result DiffResult) bool {
	return !result.HasChanges || len(result.DirtyRegions) == 0
}

// GetTotalDirtyArea calculates the total area of all dirty regions
func GetTotalDirtyArea(result DiffResult) int {
	total := 0
	for _, region := range result.DirtyRegions {
		total += region.Width * region.Height
	}
	return total
}
