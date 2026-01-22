package runtime

// Measure phase implementation.
// See measure.go for the measure functionality.
// This file contains exported entry points for measure operations.

// PerformMeasure executes the measure phase on a tree of nodes.
// This is the main entry point for the measure phase.
func PerformMeasure(root *LayoutNode, c BoxConstraints) {
	if root == nil {
		return
	}
	measure(root, c)
}
