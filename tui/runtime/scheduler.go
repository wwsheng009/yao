package runtime

import (
	"time"

	"github.com/yaoapp/yao/tui/runtime/priority"
)

// Scheduler manages priority-based rendering with time slicing
type Scheduler struct {
	defaultBudget time.Duration
}

// NewScheduler creates a new scheduler
func NewScheduler() *Scheduler {
	return &Scheduler{
		defaultBudget: 2 * time.Millisecond,
	}
}

// NewSchedulerWithBudget creates a new scheduler with a custom time budget
func NewSchedulerWithBudget(budget time.Duration) *Scheduler {
	return &Scheduler{
		defaultBudget: budget,
	}
}

// NodeRenderer is the interface for layout and rendering operations
type NodeRenderer interface {
	Layout(node *LayoutNode)
	Paint(node *LayoutNode)
}

// FrameResult represents the result of a frame render
type FrameResult struct {
	ProcessedCount  int
	OutOfTime       bool
	HighProcessed   int
	NormalProcessed int
	LowProcessed    int
}

// RenderFrame renders dirty nodes by priority with time budget
func (s *Scheduler) RenderFrame(root *LayoutNode, renderer NodeRenderer) FrameResult {
	result := FrameResult{}

	// Process in priority order: High → Normal → Low
	budgets := []struct {
		level  priority.DirtyLevel
		budget time.Duration
	}{
		{priority.DirtyHigh, s.defaultBudget},
		{priority.DirtyNormal, s.defaultBudget},
		{priority.DirtyLow, s.defaultBudget},
	}

	for _, b := range budgets {
		if s.hasDirtyAtLevel(root, b.level) {
			levelResult := s.processDirty(root, b.level, b.budget, renderer)
			result.ProcessedCount += levelResult.ProcessedCount

			switch b.level {
			case priority.DirtyHigh:
				result.HighProcessed = levelResult.ProcessedCount
			case priority.DirtyNormal:
				result.NormalProcessed = levelResult.ProcessedCount
			case priority.DirtyLow:
				result.LowProcessed = levelResult.ProcessedCount
			}

			// If we ran out of budget, defer remaining to next frame
			if levelResult.OutOfTime {
				result.OutOfTime = true
				break
			}
		}
	}

	return result
}

// processDirty processes all dirty nodes at a given level
func (s *Scheduler) processDirty(root *LayoutNode, level priority.DirtyLevel, budget time.Duration, renderer NodeRenderer) FrameResult {
	start := time.Now()
	result := FrameResult{}

	// Collect dirty nodes at this level
	dirtyNodes := s.collectDirtyByLevel(root, level)

	for _, node := range dirtyNodes {
		// Check time budget
		if time.Since(start) > budget {
			result.OutOfTime = true
			break
		}

		// Layout if needed
		if node.IsLayoutDirty() {
			renderer.Layout(node)
			node.ClearLayoutDirty()
		}

		// Paint if needed
		if node.IsPaintDirty() {
			renderer.Paint(node)
			node.ClearPaintDirty()
		}

		result.ProcessedCount++
	}

	return result
}

// collectDirtyByLevel collects all dirty nodes at a given priority level
func (s *Scheduler) collectDirtyByLevel(root *LayoutNode, level priority.DirtyLevel) []*LayoutNode {
	var result []*LayoutNode

	var traverse func(*LayoutNode)
	traverse = func(node *LayoutNode) {
		if node == nil {
			return
		}

		// Check if node matches priority and is dirty
		if node.GetPriority() == level && (node.IsLayoutDirty() || node.IsPaintDirty()) {
			result = append(result, node)
		}

		// Traverse children
		for _, child := range node.Children {
			traverse(child)
		}
	}

	traverse(root)
	return result
}

// hasDirtyAtLevel checks if there are any dirty nodes at a given level
func (s *Scheduler) hasDirtyAtLevel(root *LayoutNode, level priority.DirtyLevel) bool {
	if root == nil {
		return false
	}

	if root.GetPriority() == level && (root.IsLayoutDirty() || root.IsPaintDirty()) {
		return true
	}

	for _, child := range root.Children {
		if s.hasDirtyAtLevel(child, level) {
			return true
		}
	}

	return false
}

// SetDefaultBudget sets the default time budget per priority level
func (s *Scheduler) SetDefaultBudget(budget time.Duration) {
	s.defaultBudget = budget
}

// GetDefaultBudget returns the current time budget
func (s *Scheduler) GetDefaultBudget() time.Duration {
	return s.defaultBudget
}

// CountDirtyByLevel returns the count of dirty nodes at each level
func (s *Scheduler) CountDirtyByLevel(root *LayoutNode) map[priority.DirtyLevel]int {
	counts := map[priority.DirtyLevel]int{
		priority.DirtyHigh:   0,
		priority.DirtyNormal: 0,
		priority.DirtyLow:    0,
	}

	var traverse func(*LayoutNode)
	traverse = func(node *LayoutNode) {
		if node == nil {
			return
		}

		if node.IsLayoutDirty() || node.IsPaintDirty() {
			level := node.GetPriority()
			counts[level]++
		}

		for _, child := range node.Children {
			traverse(child)
		}
	}

	traverse(root)
	return counts
}
