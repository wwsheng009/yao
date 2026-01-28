package focus

import (
	"math"

	"github.com/yaoapp/yao/tui/runtime"
)

// NavigationDirection represents the direction for focus navigation
type NavigationDirection int

const (
	// DirectionUp navigates upward (decreasing Y)
	DirectionUp NavigationDirection = iota
	// DirectionDown navigates downward (increasing Y)
	DirectionDown
	// DirectionLeft navigates leftward (decreasing X)
	DirectionLeft
	// DirectionRight navigates rightward (increasing X)
	DirectionRight
)

// ComponentBounds represents the geometric bounds of a component
type ComponentBounds struct {
	ID        string
	X, Y      int // Top-left corner
	Width     int
	Height    int
	CenterX   int // Center point X
	CenterY   int // Center point Y
}

// NewComponentBounds creates a ComponentBounds from a LayoutNode
func NewComponentBounds(node *runtime.LayoutNode) *ComponentBounds {
	if node == nil {
		return nil
	}
	return &ComponentBounds{
		ID:      node.ID,
		X:       node.X,
		Y:       node.Y,
		Width:   node.MeasuredWidth,
		Height:  node.MeasuredHeight,
		CenterX: node.X + node.MeasuredWidth/2,
		CenterY: node.Y + node.MeasuredHeight/2,
	}
}

// GetBoundsForDirection returns the search bounds for a given direction
// For horizontal navigation, we consider the vertical range of the current component
// For vertical navigation, we consider the horizontal range of the current component
func (cb *ComponentBounds) GetBoundsForDirection(direction NavigationDirection) (min1, max1, min2, max2 int) {
	switch direction {
	case DirectionUp, DirectionDown:
		// For vertical movement, consider horizontal range
		return cb.X, cb.X + cb.Width, cb.Y, cb.Y + cb.Height
	case DirectionLeft, DirectionRight:
		// For horizontal movement, consider vertical range
		return cb.Y, cb.Y + cb.Height, cb.X, cb.X + cb.Width
	default:
		return 0, 0, 0, 0
	}
}

// GeometricNavigator provides geometric-aware navigation between components
type GeometricNavigator struct {
	rootNode *runtime.LayoutNode
	// Cached bounds for all focusable components
	componentBounds map[string]*ComponentBounds
}

// NewGeometricNavigator creates a new geometric navigator
func NewGeometricNavigator(root *runtime.LayoutNode) *GeometricNavigator {
	return &GeometricNavigator{
		rootNode:        root,
		componentBounds: make(map[string]*ComponentBounds),
	}
}

// RefreshBounds rebuilds the cache of component bounds
func (gn *GeometricNavigator) RefreshBounds(focusableIDs []string) {
	gn.componentBounds = make(map[string]*ComponentBounds)
	for _, id := range focusableIDs {
		node := gn.findNodeByID(gn.rootNode, id)
		if node != nil {
			bounds := NewComponentBounds(node)
			if bounds != nil && bounds.Width > 0 && bounds.Height > 0 {
				gn.componentBounds[id] = bounds
			}
		}
	}
}

// FindNextInDirection finds the next focusable component in the given direction
// Returns the component ID, or empty string if none found
func (gn *GeometricNavigator) FindNextInDirection(currentID string, direction NavigationDirection, focusableIDs []string) string {
	if len(focusableIDs) == 0 {
		return ""
	}

	currentBounds, hasCurrent := gn.componentBounds[currentID]

	// If current component not in cache or no current focus, return first focusable
	if !hasCurrent || currentID == "" {
		// Find the component that is most top-left
		var bestID string
		var bestX, bestY int = math.MaxInt32, math.MaxInt32

		for _, id := range focusableIDs {
			bounds, ok := gn.componentBounds[id]
			if !ok {
				continue
			}
			if bounds.Y < bestY || (bounds.Y == bestY && bounds.X < bestX) {
				bestX, bestY = bounds.X, bounds.Y
				bestID = id
			}
		}
		return bestID
	}

	// Refresh bounds if cache is empty
	if len(gn.componentBounds) == 0 {
		gn.RefreshBounds(focusableIDs)
	}

	var bestID string

	switch direction {
	case DirectionUp:
		bestID, _ = gn.findBestUp(currentBounds, focusableIDs)
	case DirectionDown:
		bestID, _ = gn.findBestDown(currentBounds, focusableIDs)
	case DirectionLeft:
		bestID, _ = gn.findBestLeft(currentBounds, focusableIDs)
	case DirectionRight:
		bestID, _ = gn.findBestRight(currentBounds, focusableIDs)
	}

	return bestID
}

// findBestUp finds the best component above the current one
func (gn *GeometricNavigator) findBestUp(current *ComponentBounds, focusableIDs []string) (string, float64) {
	var bestID string
	var bestScore float64 = -1

	for _, id := range focusableIDs {
		candidate, ok := gn.componentBounds[id]
		if !ok || candidate.ID == current.ID {
			continue
		}

		// Must be above current (bottom edge <= current top edge, or at least higher)
		if candidate.Y+candidate.Height <= current.Y || candidate.CenterY < current.CenterY {
			score := gn.scoreVerticalNavigation(current, candidate, true)
			if score > bestScore {
				bestScore = score
				bestID = candidate.ID
			}
		}
	}

	return bestID, bestScore
}

// findBestDown finds the best component below the current one
func (gn *GeometricNavigator) findBestDown(current *ComponentBounds, focusableIDs []string) (string, float64) {
	var bestID string
	var bestScore float64 = -1

	for _, id := range focusableIDs {
		candidate, ok := gn.componentBounds[id]
		if !ok || candidate.ID == current.ID {
			continue
		}

		// Must be below current (top edge >= current bottom edge, or at least lower)
		if candidate.Y >= current.Y+current.Height || candidate.CenterY > current.CenterY {
			score := gn.scoreVerticalNavigation(current, candidate, false)
			if score > bestScore {
				bestScore = score
				bestID = candidate.ID
			}
		}
	}

	return bestID, bestScore
}

// findBestLeft finds the best component to the left of the current one
func (gn *GeometricNavigator) findBestLeft(current *ComponentBounds, focusableIDs []string) (string, float64) {
	var bestID string
	var bestScore float64 = -1

	for _, id := range focusableIDs {
		candidate, ok := gn.componentBounds[id]
		if !ok || candidate.ID == current.ID {
			continue
		}

		// Must be to the left of current
		if candidate.X+candidate.Width <= current.X || candidate.CenterX < current.CenterX {
			score := gn.scoreHorizontalNavigation(current, candidate, true)
			if score > bestScore {
				bestScore = score
				bestID = candidate.ID
			}
		}
	}

	return bestID, bestScore
}

// findBestRight finds the best component to the right of the current one
func (gn *GeometricNavigator) findBestRight(current *ComponentBounds, focusableIDs []string) (string, float64) {
	var bestID string
	var bestScore float64 = -1

	for _, id := range focusableIDs {
		candidate, ok := gn.componentBounds[id]
		if !ok || candidate.ID == current.ID {
			continue
		}

		// Must be to the right of current
		if candidate.X >= current.X+current.Width || candidate.CenterX > current.CenterX {
			score := gn.scoreHorizontalNavigation(current, candidate, false)
			if score > bestScore {
				bestScore = score
				bestID = candidate.ID
			}
		}
	}

	return bestID, bestScore
}

// scoreVerticalNavigation calculates a score for vertical navigation
// Higher score means better candidate
// goingUp is true for DirectionUp, false for DirectionDown
func (gn *GeometricNavigator) scoreVerticalNavigation(current, candidate *ComponentBounds, goingUp bool) float64 {
	// Primary factor: vertical distance (less is better)
	verticalDist := math.Abs(float64(candidate.CenterY - current.CenterY))

	// Secondary factor: horizontal overlap (more is better)
	horizontalOverlap := gn.calculateHorizontalOverlap(current, candidate)

	// Calculate score: prioritize vertical proximity, then horizontal overlap
	// Score = (max possible overlap + 1 - actual distance) scaled
	const maxDistance = 1000.0 // Arbitrary cap for distance

	score := (maxDistance - verticalDist) / maxDistance

	// Bonus for horizontal overlap
	if horizontalOverlap > 0 {
		overlapBonus := float64(horizontalOverlap) / float64(max(current.Width, candidate.Width))
		score += overlapBonus * 0.5 // Up to 50% bonus for good alignment
	}

	return score
}

// scoreHorizontalNavigation calculates a score for horizontal navigation
// Higher score means better candidate
// goingLeft is true for DirectionLeft, false for DirectionRight
func (gn *GeometricNavigator) scoreHorizontalNavigation(current, candidate *ComponentBounds, goingLeft bool) float64 {
	// Primary factor: horizontal distance (less is better)
	horizontalDist := math.Abs(float64(candidate.CenterX - current.CenterX))

	// Secondary factor: vertical overlap (more is better)
	verticalOverlap := gn.calculateVerticalOverlap(current, candidate)

	// Calculate score: prioritize horizontal proximity, then vertical overlap
	const maxDistance = 1000.0

	score := (maxDistance - horizontalDist) / maxDistance

	// Bonus for vertical overlap
	if verticalOverlap > 0 {
		overlapBonus := float64(verticalOverlap) / float64(max(current.Height, candidate.Height))
		score += overlapBonus * 0.5
	}

	return score
}

// calculateHorizontalOverlap returns the horizontal overlap between two bounds
func (gn *GeometricNavigator) calculateHorizontalOverlap(a, b *ComponentBounds) int {
	left := max(a.X, b.X)
	right := min(a.X+a.Width, b.X+b.Width)
	if right > left {
		return right - left
	}
	return 0
}

// calculateVerticalOverlap returns the vertical overlap between two bounds
func (gn *GeometricNavigator) calculateVerticalOverlap(a, b *ComponentBounds) int {
	top := max(a.Y, b.Y)
	bottom := min(a.Y+a.Height, b.Y+b.Height)
	if bottom > top {
		return bottom - top
	}
	return 0
}

// findNodeByID recursively finds a node by ID in the tree
func (gn *GeometricNavigator) findNodeByID(node *runtime.LayoutNode, id string) *runtime.LayoutNode {
	if node == nil {
		return nil
	}
	if node.ID == id {
		return node
	}
	for _, child := range node.Children {
		if found := gn.findNodeByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

// GetBounds returns the cached bounds for a component ID
func (gn *GeometricNavigator) GetBounds(id string) *ComponentBounds {
	return gn.componentBounds[id]
}

// FindNearestInDirection is a convenience method that refreshes bounds and finds the next component
func (gn *GeometricNavigator) FindNearestInDirection(currentID string, direction NavigationDirection, focusableIDs []string) string {
	gn.RefreshBounds(focusableIDs)
	return gn.FindNextInDirection(currentID, direction, focusableIDs)
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
