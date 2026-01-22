package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaoapp/yao/tui/layout"
)

// DemoDir 指向 demo 文件的目录
// 假设测试运行在 E:\projects\yao\wwsheng009\yao\tui 目录下
const DemoDir = "demo/tuis/layouts"

// loadDemoConfig 从文件加载 TUI 配置
func loadDemoConfig(t *testing.T, filename string) *Config {
	// 尝试绝对路径和相对路径
	paths := []string{
		filepath.Join(DemoDir, filename),
		filepath.Join("..", DemoDir, filename),                                      // 如果在子目录运行
		filepath.Join("E:\\projects\\yao\\wwsheng009\\yao\\tui", DemoDir, filename), // 绝对路径回退
	}

	var content []byte
	var err error
	var pathUsed string

	for _, p := range paths {
		content, err = os.ReadFile(p)
		if err == nil {
			pathUsed = p
			break
		}
	}
	require.NoError(t, err, "Failed to read config file: %s", filename)
	t.Logf("Loaded config from: %s", pathUsed)

	var cfg Config
	err = json.Unmarshal(content, &cfg)
	require.NoError(t, err, "Failed to parse JSON config")

	return &cfg
}

// setupModel 初始化模型并触发布局计算
func setupModel(t *testing.T, filename string, width, height int) *Model {
	cfg := loadDemoConfig(t, filename)
	model := NewModel(cfg, nil)
	model.InitializeComponents()

	// 触发 WindowSizeMsg 来进行布局计算
	model.Update(tea.WindowSizeMsg{Width: width, Height: height})

	// 强制执行布局计算，更新节点 Bounds
	if model.LayoutEngine != nil {
		model.LayoutEngine.Layout()
	}

	return model
}

// getRealRoot returns the actual layout root, bypassing any implicit wrapper
func getRealRoot(model *Model) *layout.LayoutNode {
	root := model.LayoutRoot
	// If root has no ID and only one child that looks like a generated container
	if root != nil && root.ID == "" && len(root.Children) == 1 && root.Children[0].ID != "" {
		return root.Children[0]
	}
	return root
}

// findNodeByID 在布局树中查找特定ID的节点 (深度优先)
// 注意：Demo JSON 中可能没有显式 ID，Layout 引擎会自动生成 ID。
// 这里我们主要通过结构遍历或 Type/Props 来识别，或者在测试用例中修改 Config 添加 ID。
// 为简单起见，我们编写一个辅助函数来打印树结构，方便调试，以及通过路径查找。
func printLayoutTree(node *layout.LayoutNode, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	fmt.Printf("%sNode: ID=%s Type=%s Bound=(%d,%d %dx%d)\n",
		indent, node.ID, node.ComponentType,
		node.Bound.X, node.Bound.Y, node.Bound.Width, node.Bound.Height)

	for _, child := range node.Children {
		printLayoutTree(child, depth+1)
	}
}

// TestLayoutDashboard 测试仪表盘布局
// 重点验证：Sidebar 固定宽度，Content 自适应，Header/Footer 高度
func TestLayoutDashboard(t *testing.T) {
	width, height := 100, 30
	model := setupModel(t, "dashboard.tui.yao", width, height)
	root := getRealRoot(model)

	require.NotNil(t, root)
	// printLayoutTree(root, 0)

	// Structure:
	// Root (Vertical)
	//   - Header (Height 3)
	//   - Main Row (Flex)
	//     - Sidebar (Width 20)
	//     - Content (Flex)
	//   - Footer (Height 1)

	children := root.Children
	require.Equal(t, 3, len(children), "Dashboard should have 3 main sections")

	header := children[0]
	mainRow := children[1]
	footer := children[2]

	// 1. Verify Header
	assert.Equal(t, 3, header.Bound.Height, "Header height should be 3")
	assert.Equal(t, width, header.Bound.Width, "Header width should fill window")

	// 2. Verify Footer
	assert.Equal(t, 1, footer.Bound.Height, "Footer height should be 1")
	assert.Equal(t, width, footer.Bound.Width, "Footer width should fill window")
	assert.Equal(t, height-1, footer.Bound.Y, "Footer should be at the bottom")

	// 3. Verify Main Row
	expectedMainHeight := height - 3 - 1 // Total - Header - Footer
	assert.Equal(t, expectedMainHeight, mainRow.Bound.Height, "Main row should fill remaining height")

	// 4. Verify Sidebar vs Content
	require.Equal(t, 2, len(mainRow.Children), "Main row should have Sidebar and Content")
	sidebar := mainRow.Children[0]
	content := mainRow.Children[1]

	assert.Equal(t, 20, sidebar.Bound.Width, "Sidebar width should be fixed at 20")

	// Content width = Total - Sidebar - Padding/Border (if any)
	// Note: Engine calculates Flex based on available space.
	// Check if Content fills the rest
	assert.Greater(t, content.Bound.Width, 50, "Content should take remaining space")
	assert.Equal(t, width-20, content.Bound.Width, "Content width should be Total - Sidebar")
}

// TestLayoutHolyGrail 测试圣杯布局
// 重点验证：Left/Right Sidebar 宽度，Center Content Grow
func TestLayoutHolyGrail(t *testing.T) {
	width, height := 120, 40
	model := setupModel(t, "holy-grail.tui.yao", width, height)
	root := getRealRoot(model)

	// Structure:
	// Root
	//   - Header (Height 3)
	//   - Middle Row (Flex, Grow=1)
	//     - Left Nav (Width 20)
	//     - Main Content (Flex, Grow=1)
	//     - Right Ad (Width 15)
	//   - Footer (Height 2)

	children := root.Children
	require.Equal(t, 3, len(children))

	middleRow := children[1]
	require.Equal(t, 3, len(middleRow.Children))

	leftNav := middleRow.Children[0]
	mainContent := middleRow.Children[1]
	rightAd := middleRow.Children[2]

	// Verify Fixed Widths
	assert.Equal(t, 20, leftNav.Bound.Width, "Left Nav width")
	assert.Equal(t, 15, rightAd.Bound.Width, "Right Ad width")

	// Verify Flex Growth
	// Available = 120 - 20 - 15 = 85
	assert.Equal(t, 85, mainContent.Bound.Width, "Main Content should fill remaining width")
}

// TestLayoutResponsive 测试响应式布局 (Grow/Shrink)
func TestLayoutResponsive(t *testing.T) {
	// Case 1: Extra Space (Grow Test)
	t.Run("Grow Distribution", func(t *testing.T) {
		// Window: 100 wide
		// Items: Fixed(10) + Gap(1) + Grow(1) + Gap(1) + Grow(2)
		// Available for Grow: 100 - 10 - 2(gaps) - 4(padding) = 84
		// Grow 1 shares: 84 * (1/3) = 28
		// Grow 2 shares: 84 * (2/3) = 56

		model := setupModel(t, "responsive.tui.yao", 100, 30)
		root := getRealRoot(model)

		// Find the "Grow" container (second child of root)
		// Root -> [Text, Layout(Grow), Text, Text, Layout(Shrink), Text]
		growContainer := root.Children[1]

		require.Equal(t, 3, len(growContainer.Children))

		fixedItem := growContainer.Children[0]
		grow1Item := growContainer.Children[1]
		grow2Item := growContainer.Children[2]

		assert.Equal(t, 10, fixedItem.Bound.Width, "Fixed item width")

		// Allow small margin of error due to integer division
		assert.InDelta(t, 28, grow1Item.Bound.Width, 2, "Grow(1) item width")
		assert.InDelta(t, 56, grow2Item.Bound.Width, 2, "Grow(2) item width")

		// Verify 1:2 ratio roughly
		assert.True(t, grow2Item.Bound.Width > grow1Item.Bound.Width, "Grow(2) should be larger than Grow(1)")
	})

	// Case 2: Insufficient Space (Shrink Test)
	t.Run("Shrink Distribution", func(t *testing.T) {
		// Window: 60 wide
		// Padding: 4 (Left 2 + Right 2) -> Inner: 56
		// Container Padding/Border? The layout has border=true, but layout engine currently might not subtract border width automatically unless specified.
		// Items: NoShrink(20) + Gap(1) + Shrink1(40) + Gap(1) + Shrink3(40)
		// Total Needed: 20 + 40 + 40 + 2 = 102
		// Available: 56
		// Overflow: 102 - 56 = 46 to shrink
		// Shrink Weights: Item2(1), Item3(3). Total = 4.
		// Item2 shrinks: 46 * 1/4 = 11.5 -> New Width: 40 - 11 = 29
		// Item3 shrinks: 46 * 3/4 = 34.5 -> New Width: 40 - 34 = 6

		model := setupModel(t, "responsive.tui.yao", 60, 30)
		root := getRealRoot(model)

		// Find the "Shrink" container (5th child of root)
		shrinkContainer := root.Children[5]

		require.Equal(t, 3, len(shrinkContainer.Children))

		noShrinkItem := shrinkContainer.Children[0]
		shrink1Item := shrinkContainer.Children[1]
		shrink3Item := shrinkContainer.Children[2]

		assert.Equal(t, 20, noShrinkItem.Bound.Width, "NoShrink item should preserve width")

		assert.True(t, shrink1Item.Bound.Width < 40, "Shrink(1) should shrink")
		assert.True(t, shrink3Item.Bound.Width < 40, "Shrink(3) should shrink")

		assert.True(t, shrink1Item.Bound.Width > shrink3Item.Bound.Width, "Shrink(1) should shrink LESS than Shrink(3)")

		t.Logf("Shrink widths: NoShrink=%d, Shrink(1)=%d, Shrink(3)=%d",
			noShrinkItem.Bound.Width, shrink1Item.Bound.Width, shrink3Item.Bound.Width)
	})
}

// TestLayoutAbsolute 测试绝对定位
// 重点验证：子元素的 X, Y 坐标是否相对于父元素偏移
func TestLayoutAbsolute(t *testing.T) {
	model := setupModel(t, "absolute-layout.tui.yao", 80, 24)
	root := getRealRoot(model)

	// Structure:
	// Root
	//   - Relative Container
	//     - Background Text (Flow)
	//     - Modal (Absolute, Top:4, Left:10)
	//     - Toast (Absolute, Bottom:1, Right:2)

	container := root.Children[0]
	// Container position
	cX, cY := container.Bound.X, container.Bound.Y
	cW, cH := container.Bound.Width, container.Bound.Height

	require.Equal(t, 3, len(container.Children))

	modal := container.Children[1]
	toast := container.Children[2]

	// Verify Modal Position (Top:4, Left:10)
	assert.Equal(t, cX+10, modal.Bound.X, "Modal absolute Left")
	assert.Equal(t, cY+4, modal.Bound.Y, "Modal absolute Top")

	// Verify Toast Position (Bottom:1, Right:2)
	// Right: 2 means X = Width - 2 - ElementWidth
	// Bottom: 1 means Y = Height - 1 - ElementHeight
	// Toast Width: 20, Height: 3
	expectedToastX := cX + cW - 2 - 20
	expectedToastY := cY + cH - 1 - 3

	assert.Equal(t, expectedToastX, toast.Bound.X, "Toast absolute Right")
	assert.Equal(t, expectedToastY, toast.Bound.Y, "Toast absolute Bottom")
}

// TestLayoutAlignment 测试对齐方式
// 重点验证：Justify 导致的位置偏移
func TestLayoutAlignment(t *testing.T) {
	width := 60 // Small width to force spacing visible
	model := setupModel(t, "alignment.tui.yao", width, 40)
	root := getRealRoot(model)

	// Root has many children pairs (Text label + Layout box)
	// Box 1: Justify Start (Children[1])
	// Box 2: Justify Center (Children[3])
	// Box 3: Justify Space Between (Children[5])

	// 1. Justify Start (Standard)
	boxStart := root.Children[1]
	// Items should be at 0, ItemWidth, ItemWidth*2...
	assert.Equal(t, boxStart.Bound.X, boxStart.Children[0].Bound.X, "Start: Item 1 should be at left edge")

	// 2. Justify Center
	boxCenter := root.Children[3]
	// Item 1 should NOT be at left edge
	// It should be shifted right by (TotalWidth - ContentWidth) / 2
	assert.Greater(t, boxCenter.Children[0].Bound.X, boxCenter.Bound.X, "Center: Item 1 should be shifted right")

	// 3. Justify Space Between
	boxSpaceBetween := root.Children[5]

	// Item 1 at Left, Item 3 at Right
	assert.Equal(t, boxSpaceBetween.Bound.X, boxSpaceBetween.Children[0].Bound.X, "SpaceBetween: Item 1 at left")

	lastItem := boxSpaceBetween.Children[len(boxSpaceBetween.Children)-1]
	expectedRightX := boxSpaceBetween.Bound.X + boxSpaceBetween.Bound.Width - lastItem.Bound.Width
	// Allow small rounding error
	assert.InDelta(t, expectedRightX, lastItem.Bound.X, 1, "SpaceBetween: Last item at right edge")
}

// TestLayoutForm 测试表单布局
// 重点验证：标签对齐和输入框 Flex
func TestLayoutForm(t *testing.T) {
	model := setupModel(t, "form-layout.tui.yao", 80, 24)
	root := getRealRoot(model)

	// Structure is nested deeply, let's look for the rows
	// Root -> [Title, FormContainer(Vertical), ...]
	// FormContainer -> [Row1, Row2, Row3...]
	// Row -> [Label(Width 15), Gap, Input(Flex)]

	formContainer := root.Children[1]
	row1 := formContainer.Children[0]

	label := row1.Children[0]
	input := row1.Children[2] // Children[1] is gap/spacer layout

	// Verify Label Width
	assert.Equal(t, 15, label.Bound.Width, "Label should have fixed width")

	// Verify Input Flex
	// Row width - Label(15) - Gap(1) = Input width
	expectedInputWidth := row1.Bound.Width - 15 - 1
	assert.Equal(t, expectedInputWidth, input.Bound.Width, "Input should flex fill row")
}
