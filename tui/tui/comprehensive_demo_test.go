package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComprehensiveDemo tests the comprehensive demo that showcases all features.
func TestComprehensiveDemo(t *testing.T) {
	width, height := 100, 30
	model := setupModel(t, "comprehensive.tui.yao", width, height)
	root := getRealRoot(model)

	require.NotNil(t, root)
	require.NotNil(t, model.RuntimeRoot)

	t.Logf("✅ Comprehensive demo loaded successfully")
	t.Logf("Root ID: %s", root.ID)
	t.Logf("Root Type: %s", root.Type)
	t.Logf("Root Children: %d", len(root.Children))

	// Verify main structure
	require.Equal(t, 3, len(root.Children), "Should have header, main row, footer")

	// Header
	header := root.Children[0]
	assert.Equal(t, 3, header.MeasuredHeight, "Header height should be 3")

	// Main Row
	mainRow := root.Children[1]
	assert.Greater(t, mainRow.MeasuredHeight, 20, "Main row should have significant height")

	// Sidebar + Content
	require.Equal(t, 2, len(mainRow.Children), "Main row should have sidebar and content")

	sidebar := mainRow.Children[0]
	content := mainRow.Children[1]

	assert.Equal(t, 25, sidebar.MeasuredWidth, "Sidebar width should be 25")
	assert.Greater(t, content.MeasuredWidth, 50, "Content should take remaining space")

	// Footer
	footer := root.Children[2]
	assert.Equal(t, 2, footer.MeasuredHeight, "Footer height should be 2")

	t.Logf("✅ All basic layout checks passed")
	t.Logf("   - Header: %dx%d", header.MeasuredWidth, header.MeasuredHeight)
	t.Logf("   - Sidebar: %dx%d", sidebar.MeasuredWidth, sidebar.MeasuredHeight)
	t.Logf("   - Content: %dx%d", content.MeasuredWidth, content.MeasuredHeight)
	t.Logf("   - Footer: %dx%d", footer.MeasuredWidth, footer.MeasuredHeight)
}
