//go:build theme_demo
// +build theme_demo

package main

import (
	"fmt"

	"github.com/yaoapp/yao/tui/framework"
	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/display"
	"github.com/yaoapp/yao/tui/framework/input"
	"github.com/yaoapp/yao/tui/framework/layout"
)

// =============================================================================
// Theme Switch Demo - 主题切换演示程序（交互式）
// =============================================================================

func main() {
	app := framework.NewApp()

	// 初始化主题
	if err := app.InitTheme("dark"); err != nil {
		fmt.Printf("初始化主题失败: %v\n", err)
		return
	}

	// 创建主界面
	root := createThemeDemoUI(app)
	app.SetRoot(root)

	// 注册快捷键
	app.OnKey('q', app.Quit)

	// 注册主题切换快捷键
	registerThemeShortcuts(app)

	fmt.Println("==============================================")
	fmt.Println("   TUI Framework - 主题切换演示")
	fmt.Println("==============================================")
	fmt.Println()
	fmt.Println("快捷键:")
	fmt.Println("  1-6    - 切换主题")
	fmt.Println("  t      - 切换到下一个主题")
	fmt.Println("  q      - 退出")
	fmt.Println()
	fmt.Println("可用主题:")
	fmt.Println("  1. light       - 亮色主题")
	fmt.Println("  2. dark        - 暗色主题")
	fmt.Println("  3. dracula     - Dracula 主题")
	fmt.Println("  4. nord        - Nord 主题")
	fmt.Println("  5. monokai     - Monokai 主题")
	fmt.Println("  6. tokyo-night - Tokyo Night 主题")
	fmt.Println()

	// 运行应用
	if err := app.Run(); err != nil {
		fmt.Printf("运行出错: %v\n", err)
	}
}

// createThemeDemoUI 创建主题演示界面
func createThemeDemoUI(app *framework.App) component.Node {
	container := layout.NewFlex(layout.Column).WithGap(0)

	// 标题 - 使用主题样式
	title := display.NewText("┌──────────────────────────────────────────┐")
	title.SetStyleID("text.secondary")

	title2 := display.NewText("│          TUI Framework - 主题切换演示        │")
	title2.SetStyleID("text.primary")

	title3 := display.NewText("└──────────────────────────────────────────┘")
	title3.SetStyleID("text.secondary")

	// 当前主题显示（将动态更新）
	themeInfo := display.NewText(fmt.Sprintf("  当前主题: %s", app.GetTheme()))
	themeInfo.SetStyleID("text.primary")

	// 保存 themeInfo 引用，以便后续更新
	app.SetUserData("themeInfo", themeInfo)

	// 说明
	help1 := display.NewText("  快捷键: 1-6 切换主题 | t 循环 | q 退出")
	help1.SetStyleID("text.secondary")

	// 分隔线
	separator := display.NewText("  ──────────────────────────────────────────")
	separator.SetStyleID("text.secondary")

	// 示例输入框
	usernameLabel := display.NewText("  用户名:")
	usernameLabel.SetStyleID("text.secondary")

	usernameInput := input.NewTextInput()
	usernameInput.SetPlaceholder("请输入用户名...")
	usernameInput.SetValue("demo_user")

	// 主题列表
	themeList := display.NewText("  可用主题: 1.light 2.dark 3.dracula 4.nord 5.monokai 6.tokyo-night")
	themeList.SetStyleID("text.secondary")

	// 将所有组件添加到容器
	container.WithChildren(
		title,
		title2,
		title3,
		display.NewText(""),
		themeInfo,
		display.NewText(""),
		help1,
		display.NewText(""),
		separator,
		display.NewText(""),
		usernameLabel,
		usernameInput,
		display.NewText(""),
		themeList,
	)

	return container
}

// placeholderLabel 占位符标签
var placeholderLabel = display.NewText("  占位符:")

// registerThemeShortcuts 注册主题切换快捷键
func registerThemeShortcuts(app *framework.App) {
	themes := []string{"light", "dark", "dracula", "nord", "monokai", "tokyo-night"}

	// 主题切换处理器 - 更新主题信息和重新渲染
	changeTheme := func(themeName string) {
		app.SetTheme(themeName)

		// 更新主题信息文本
		if themeInfo, ok := app.GetUserData("themeInfo").(*display.Text); ok {
			themeInfo.SetContent(fmt.Sprintf("  当前主题: %s", themeName))
			themeInfo.MarkDirty()
		}
	}

	// 数字键 1-6 切换到对应主题
	for i, themeName := range themes {
		index := i
		name := themeName
		app.OnKey(rune('1'+index), func() {
			changeTheme(name)
		})
	}

	// t 键循环切换主题
	currentThemeIndex := 1 // 从 dark 开始
	app.OnKey('t', func() {
		currentThemeIndex = (currentThemeIndex + 1) % len(themes)
		changeTheme(themes[currentThemeIndex])
	})
}
