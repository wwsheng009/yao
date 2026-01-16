# Yao TUI 引擎

基于 Bubble Tea 框架的终端用户界面引擎，为 Yao 提供声明式 DSL 驱动的 TUI 能力。

## 📋 概述

TUI 引擎允许开发者通过 JSON/YAML 配置文件（`.tui.yao`）定义终端界面，支持：

- 🎨 声明式 UI 布局（Lip Gloss 样式）
- 🔄 响应式状态管理
- ⚡ 异步 Process 调用
- 🤖 AI 流式渲染集成
- 📜 JavaScript/TypeScript 脚本支持（V8Go）
- 🧩 可扩展组件系统

## 🏗️ 架构

```
tui/
├── types.go          # 核心类型定义
├── loader.go         # DSL 加载器
├── driver.go         # Bubble Tea 驱动
├── model.go          # Model 实现
├── action.go         # Action 执行器
├── render.go         # 布局渲染器
├── script.go         # V8 脚本加载器
├── jsapi.go          # JavaScript API
├── process.go        # Process 处理器导出
├── theme.go          # 样式主题
├── metrics.go        # 性能指标
├── components/       # 标准组件库
│   ├── header.go
│   ├── table.go
│   ├── form.go
│   ├── chat.go
│   └── viewport.go
└── mock/            # 测试工具
    ├── program.go
    └── process.go
```

## 🚀 快速开始

### 1. 创建 TUI 配置

```json
// tuis/hello.tui.yao
{
  "name": "Hello TUI",
  "data": {
    "title": "Hello Yao TUI!"
  },
  "layout": {
    "direction": "vertical",
    "children": [
      {
        "type": "header",
        "props": {
          "title": "{{title}}"
        }
      }
    ]
  }
}
```

### 2. 运行 TUI

```bash
yao tui hello
```

## 📚 文档

- [架构设计](ARCHITECTURE.md) - 详细架构说明
- [实施计划](TODO.md) - 开发任务清单
- [API 参考](docs/API.md) - JavaScript API 文档
- [组件开发](docs/COMPONENTS.md) - 组件开发指南

## 🧪 开发

### 运行测试

```bash
# 单元测试
make test-tui

# 性能测试
make bench-tui

# 代码检查
make lint-tui

# 覆盖率
make cover-tui
```

### 测试环境

TUI 测试使用独立的应用目录进行隔离测试：


```bash
# 测试应用目录
YAO_TEST_APPLICATION=../.vscode/yao-docs/YaoApps/tui_app
```

测试目录结构：
```
.vscode/yao-docs/YaoApps/tui_app/
├── app.yao              # 应用配置
└── tuis/                # TUI 配置文件
    ├── hello.tui.yao    # 示例 TUI
    └── admin/
        └── dashboard.tui.yao  # 嵌套示例
```

### golang调试

如果是golang测试，需要在测试代码中调用prepare函数来准备环境，测试用的tui文件保存到指定的目录下

```go
// TestReloadNotFound tests reloading a non-existent TUI
func TestReloadNotFound(t *testing.T) {
	prepare(t)
	defer test.Clean()

	err := Reload("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TUI file not found")
}
```

```bash
export YAO_TEST_APPLICATION=".vscode/yao-docs/YaoApps/tui_app"
go test -v ./tui -run TestReloadNotFound 2>&1
```

### 调试

```bash
# 启用调试模式
export YAO_TUI_DEBUG=true
# 设置应用根目录
export YAO_ROOT=../.vscode/yao-docs/YaoApps/tui_app
yao tui hello --debug
```

## 📊 状态

- **版本**: v0.1.0 (开发中)
- **Go 版本**: >= 1.21
- **测试覆盖率**: 目标 85%+
- **状态**: 🚧 实施阶段

## 🤝 贡献

请参考主项目的 [贡献指南](../CONTRIBUTING.md)。

## 📄 许可

与 Yao 主项目保持一致。
