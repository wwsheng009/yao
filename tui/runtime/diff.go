package runtime

// DiffResult represents the difference between two frames
type DiffResult struct {
	DirtyRegions []Rect
	HasChanges   bool
}

// ComputeDiff compares two frames and returns the difference
// This identifies regions that have changed between frames to optimize rendering
func ComputeDiff(oldFrame, newFrame Frame) DiffResult {
	result := DiffResult{
		DirtyRegions: make([]Rect, 0),
		HasChanges:   false,
	}

	// If old frame is empty, entire new frame is dirty
	if oldFrame.Buffer == nil || oldFrame.Width == 0 || oldFrame.Height == 0 {
		result.HasChanges = true
		result.DirtyRegions = append(result.DirtyRegions, Rect{
			X:      0,
			Y:      0,
			Width:  newFrame.Width,
			Height: newFrame.Height,
		})
		return result
	}

	// Compare cell by cell
	minWidth := oldFrame.Width
	if newFrame.Width < minWidth {
		minWidth = newFrame.Width
	}
	minHeight := oldFrame.Height
	if newFrame.Height < minHeight {
		minHeight = newFrame.Height
	}

	// Track dirty rectangles
	currentRegion := Rect{}
	inRegion := false

	for y := 0; y < minHeight; y++ {
		for x := 0; x < minWidth; x++ {
			oldCell := oldFrame.Buffer.GetContent(x, y)
			newCell := newFrame.Buffer.GetContent(x, y)

			// Check if cell changed
			changed := oldCell.Char != newCell.Char ||
				oldCell.Style.Bold != newCell.Style.Bold ||
				oldCell.Style.Italic != newCell.Style.Italic ||
				oldCell.Style.Underline != newCell.Style.Underline

			if changed {
				if !inRegion {
					// Start a new dirty region
					currentRegion = Rect{
						X:      x,
						Y:      y,
						Width:  1,
						Height: 1,
					}
					inRegion = true
				} else {
					// Extend current region
					right := x + 1
					if right > currentRegion.X+currentRegion.Width {
						currentRegion.Width = right - currentRegion.X
					}
					bottom := y + 1
					if bottom > currentRegion.Y+currentRegion.Height {
						currentRegion.Height = bottom - currentRegion.Y
					}
				}
			} else if inRegion {
				// End current region and save it
				result.DirtyRegions = append(result.DirtyRegions, currentRegion)
				inRegion = false
			}
		}

		// Check for changes in new frame area (if new frame is larger)
		if newFrame.Width > oldFrame.Width {
			result.HasChanges = true
			result.DirtyRegions = append(result.DirtyRegions, Rect{
				X:      oldFrame.Width,
				Y:      0,
				Width:  newFrame.Width - oldFrame.Width,
				Height: newFrame.Height,
			})
		}

		if newFrame.Height > oldFrame.Height {
			result.HasChanges = true
			result.DirtyRegions = append(result.DirtyRegions, Rect{
				X:      0,
				Y:      oldFrame.Height,
				Width:  newFrame.Width,
				Height: newFrame.Height - oldFrame.Height,
			})
		}
	}

	// Save the last region if we were tracking one
	if inRegion {
		result.DirtyRegions = append(result.DirtyRegions, currentRegion)
	}

	// Merge overlapping dirty regions for efficiency
	result.DirtyRegions = mergeDirtyRegions(result.DirtyRegions)

	result.HasChanges = len(result.DirtyRegions) > 0

	return result
}

// mergeDirtyRegions merges overlapping or adjacent dirty regions
func mergeDirtyRegions(regions []Rect) []Rect {
	if len(regions) == 0 {
		return regions
	}

	// Simple merge: combine overlapping or adjacent regions
	merged := make([]Rect, 0, len(regions))
	merged = append(merged, regions[0])

	for i := 1; i < len(regions); i++ {
		merged = mergeRegion(merged, regions[i])
	}

	return merged
}

// mergeRegion merges a region into a list of regions, merging overlaps
func mergeRegion(regions []Rect, newRegion Rect) []Rect {
	for i, existing := range regions {
		if regionsOverlap(existing, newRegion) {
			// Merge with existing region
			merged := mergeTwoRegions(existing, newRegion)
			regions[i] = merged
			return regions
		}
	}

	// No overlap found, add as new region
	regions = append(regions, newRegion)
	return regions
}

// regionsOverlap checks if two regions overlap or are adjacent
func regionsOverlap(a, b Rect) bool {
	return !(a.X+a.Width < b.X || b.X+b.Width < a.X ||
		a.Y+a.Height < b.Y || b.Y+b.Height < a.Y)
}

// mergeTwoRegions merges two overlapping regions into one
func mergeTwoRegions(a, b Rect) Rect {
	x := a.X
	if b.X < x {
		x = b.X
	}
	y := a.Y
	if b.Y < y {
		y = b.Y
	}
	right := a.X + a.Width
	if b.X+b.Width > right {
		right = b.X + b.Width
	}
	bottom := a.Y + a.Height
	if b.Y+b.Height > bottom {
		bottom = b.Y + b.Height
	}

	return Rect{
		X:      x,
		Y:      y,
		Width:  right - x,
		Height: bottom - y,
	}
}
