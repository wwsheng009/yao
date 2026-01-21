package tui

import (
	"os"

	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/tui"
)

var langs = map[string]string{
	"Run a terminal user interface":                                     "运行终端用户界面",
	"Run a terminal user interface defined in .tui.yao files":           "运行在 .tui.yao 文件中定义的终端用户界面",
	"Enable debug mode":                                                 "启用调试模式",
	"Enable verbose output":                                             "启用详细输出",
	"List all loaded TUI configurations":                                "列出所有已加载的 TUI 配置",
	"List all loaded TUI configurations with details":                   "列出所有已加载的 TUI 配置及详细信息",
	"Validate a TUI configuration":                                      "验证 TUI 配置",
	"Validate a specific TUI configuration and show validation details": "验证特定的 TUI 配置并显示验证详情",
	"Inspect a TUI configuration":                                       "检查 TUI 配置",
	"Inspect a TUI configuration and show detailed information":         "检查 TUI 配置并显示详细信息",
	"Check a TUI configuration":                                         "检查 TUI 配置",
	"Initialize a TUI and check for initialization errors":              "初始化 TUI 并检查初始化错误",
	"Dump TUI configuration as JSON":                                    "将 TUI 配置导出为 JSON",
	"Dump the raw TUI configuration JSON for debugging purposes":        "将原始 TUI 配置 JSON 导出用于调试",
	"Show help for TUI command":                                         "显示 TUI 命令的帮助信息",
}

// L 多语言切换
func L(words string) string {
	var lang = os.Getenv("YAO_LANG")
	if lang == "" {
		return words
	}

	if trans, has := langs[words]; has {
		return trans
	}
	return words
}

// Boot 设定配置
func Boot() {
	root := config.Conf.Root
	if root == "" {
		root, _ = os.Getwd()
	}

	config.Conf = config.LoadFrom(root + "/.env")

	if config.Conf.Mode == "production" {
		config.Production()
	} else if config.Conf.Mode == "development" {
		config.Development()
	}
}

// Count returns number of loaded TUI configurations.
func Count() int {
	return tui.Count()
}

// List returns all loaded TUI IDs.
func List() []string {
	return tui.List()
}
