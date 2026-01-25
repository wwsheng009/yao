# TUI 框架 AI 辅助开发指南

> **目标**: 让 AI 能够自主完成 TUI 框架的开发任务，包括代码编写、测试、文档管理等。
> **工作目录**: `E:\projects\yao\wwsheng009\yao`
> **Demo 目录**: `E:\projects\yao\wwsheng009\yao\tui\demo` (或 `/e/projects/yao/yao-projects/yao-docs/docs/YaoApps/tui_app`)

## 目录

1. [环境配置](#环境配置)
2. [开发工作流](#开发工作流)
3. [AI 编程规范](#ai-编程规范)
4. [测试规范](#测试规范)
5. [任务管理](#任务管理)
6. [文档管理](#文档管理)
7. [常用命令](#常用命令)
8. [故障排查](#故障排查)

---

## 环境配置

### 1. 必需的环境变量

在开发调试过程中，需要设置以下环境变量：

```bash
# Windows PowerShell
$env:YAO_ROOT="E:\projects\yao\wwsheng009\yao"
$env:YAO_TEST_APPLICATION="E:\projects\yao\wwsheng009\yao\tui\demo"
$env:YAO_ENV="development"
$env:YAO_LOG_CONSOLE="true"  # 启用控制台输出
$env:YAO_LOG_LEVEL="DEBUG"   # 或 TRACE

# Linux/Mac
export YAO_ROOT="E:/projects/yao/wwsheng009/yao"
export YAO_TEST_APPLICATION="E:/projects/yao/wwsheng009/yao/tui/demo"
export YAO_ENV="development"
export YAO_LOG_CONSOLE="true"
export YAO_LOG_LEVEL="DEBUG"
```

### 2. .env 文件配置

在项目根目录创建 `.env` 文件：

```env
# Yao 框架配置
YAO_ROOT=E:\projects\yao\wwsheng009\yao
YAO_TEST_APPLICATION=E:\projects\yao\wwsheng009\yao\tui\demo
YAO_ENV=development

# 日志配置 (开发调试)
YAO_LOG_CONSOLE=true
YAO_LOG_LEVEL=TRACE
YAO_LOG_MODE=TEXT

# 数据库配置
YAO_DB_DRIVER=sqlite3
YAO_DB_PRIMARY=./yao.db

# 服务器配置
YAO_HOST=127.0.0.1
YAO_PORT=5099
```

### 3. Go 模块配置

```bash
# 确保在项目根目录
cd E:\projects\yao\wwsheng009\yao

# 检查 go.mod
go mod tidy
go mod verify
```

---

## 开发工作流

### 完整开发循环

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        AI 开发工作流                                      │
└─────────────────────────────────────────────────────────────────────────┘

    1. 需求分析
       ├── 阅读设计文档 (tui/framework/docs/)
       ├── 理解架构不变量 (ARCHITECTURE_INVARIANTS.md)
       └── 确认任务边界

    2. 代码编写
       ├── 遵循 Capability Interfaces 设计
       ├── 使用 Go 最佳实践
       └── 添加注释

    3. 单元测试
       ├── 编写测试用例
       ├── 运行测试: make unit-test-tui
       └── 确保覆盖率 > 80%

    4. 集成测试
       ├── Demo 应用测试
       ├── 端到端测试
       └── 性能测试

    5. 文档更新
       ├── 更新设计文档
       ├── 更新 API 文档
       └── 添加示例代码

    6. 代码审查
       ├── 运行: go vet ./...
       ├── 运行: gofmt -s -w ...
       └── 运行: make lint
```

### AI 任务执行步骤

当接收到开发任务时，AI 应按以下步骤执行：

#### 步骤 1: 理解任务

```go
// 1. 阅读相关设计文档
// 2. 查看现有代码实现
// 3. 确认接口定义
// 4. 理解依赖关系
```

#### 步骤 2: 设计方案

```go
// 1. 确定需要实现的能力接口 (Capability Interfaces)
// 2. 设计数据结构
// 3. 规划测试用例
// 4. 列出依赖文件
```

#### 步骤 3: 编写代码

```go
// 1. 创建/修改源文件
// 2. 实现接口方法
// 3. 添加错误处理
// 4. 添加日志输出
```

#### 步骤 4: 编写测试

```go
// 1. 创建测试文件
// 2. 编写表驱动测试
// 3. 添加边界条件测试
// 4. 运行测试验证
```

#### 步骤 5: 更新文档

```go
// 1. 更新相关设计文档
// 2. 添加使用示例
// 3. 更新 ARCHITECTURE.md (如需要)
```

---

## AI 编程规范

### 1. 目录结构

```
tui/framework/
├── component/          # 组件定义
│   ├── node.go        # Node 基础接口
│   ├── mountable.go   # Mountable 能力
│   ├── measurable.go  # Measurable 能力
│   └── ...
├── runtime/           # 运行时内核 (纯 Go，无外部依赖)
│   ├── layout/       # 布局引擎
│   ├── paint/        # 绘制系统
│   ├── focus/        # 焦点系统
│   └── ...
├── platform/         # 平台抽象
│   └── impl/         # 平台实现
├── display/          # 显示组件 (Table, List, Text...)
├── input/            # 输入组件 (TextInput, TextArea...)
├── layout/           # 布局组件 (Flex, Box, Grid...)
├── interactive/      # 交互组件 (Button, Checkbox...)
├── style/            # 样式系统
└── docs/             # 设计文档
```

### 2. 文件命名

```go
// 文件名使用 snake_case
// 文件名应描述主要类型或功能

// ✅ 好的命名
component/node.go
component/measurable.go
runtime/layout/flex.go

// ❌ 不好的命名
component/component.go
runtime/stuff.go
```

### 3. 代码组织

```go
// 1. 包声明
package component

// 2. 导入 (标准库 → 第三方 → 本地)
import (
    "io"                    // 标准库
    "os"

    "github.com/stretchr/testify/assert"  // 第三方
    "rogchap.com/v8go"

    "github.com/yaoapp/yao/tui/framework/util"  // 本地
)

// 3. 常量
const (
    DefaultWidth = 80
    DefaultHeight = 24
)

// 4. 接口定义
type Node interface {
    ID() string
    Type() string
}

// 5. 类型定义
type BaseComponent struct {
    id string
    props map[string]interface{}
}

// 6. 构造函数
func NewBaseComponent(id string) *BaseComponent {
    return &BaseComponent{
        id: id,
        props: make(map[string]interface{}),
    }
}

// 7. 接口实现
func (c *BaseComponent) ID() string {
    return c.id
}

// 8. 辅助方法
func (c *BaseComponent) setProp(key string, value interface{}) {
    c.props[key] = value
}
```

### 4. 接口设计原则

```go
// ✅ 好的设计: 小而专注的能力接口
type Measurable interface {
    Node
    Measure(constraints Constraints) Size
    GetSize() Size
}

type Paintable interface {
    Node
    Paint(ctx PaintContext, buf *CellBuffer)
}

type Focusable interface {
    Node
    FocusID() string
    OnFocus()
    OnBlur()
}

// ❌ 不好的设计: 胖接口
type Component interface {
    ID() string
    Type() string
    Measure(...) Size
    Paint(...)
    FocusID() string
    OnFocus()
    OnBlur()
    ScrollTo(...)
    // ... 20+ 方法
}
```

### 5. 错误处理

```go
// 1. 显式处理错误
result, err := someFunction()
if err != nil {
    return fmt.Errorf("context: %w", err)
}

// 2. 使用结构化错误
type ValidationError struct {
    Field   string
    Message string
    Value   interface{}
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// 3. 错误日志
if err != nil {
    log.With(log.F{
        "component": c.ID(),
        "action": "paint",
        "error": err.Error(),
    }).Error("Failed to paint component")
}
```

### 6. 并发安全

```go
// 使用具名互斥锁
type ComponentRegistry struct {
    components map[string]*Component
    mu         sync.RWMutex
}

// 读操作使用读锁
func (r *ComponentRegistry) Get(id string) (*Component, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    comp, ok := r.components[id]
    return comp, ok
}

// 写操作使用写锁
func (r *ComponentRegistry) Add(comp *Component) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.components[comp.ID()] = comp
}
```

---

## 测试规范

### 1. 测试文件命名

```go
// 测试文件与源文件同目录
// 命名: 源文件名_test.go

// 示例:
component/node.go         → component/node_test.go
runtime/layout/flex.go    → runtime/layout/flex_test.go
display/table.go          → display/table_test.go
```

### 2. 测试结构

```go
package component

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// 表驱动测试
func TestBaseComponent_ID(t *testing.T) {
    tests := []struct {
        name     string
        id       string
        expected string
    }{
        {"simple id", "test-comp", "test-comp"},
        {"empty id", "", ""},
        {"special chars", "comp-123", "comp-123"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            comp := NewBaseComponent(tt.id)
            assert.Equal(t, tt.expected, comp.ID())
        })
    }
}

// Setup/Teardown
func TestWithSetup(t *testing.T) {
    // Setup
    registry := NewTestRegistry()
    defer registry.Close()

    // Test
    comp := registry.NewComponent("test")
    assert.NotNil(t, comp)
}

// 子测试
func TestComponent_Multiple(t *testing.T) {
    comp := NewBaseComponent("test")

    t.Run("ID", func(t *testing.T) {
        assert.Equal(t, "test", comp.ID())
    })

    t.Run("Type", func(t *testing.T) {
        assert.Equal(t, "base", comp.Type())
    })
}
```

### 3. 测试命令

```bash
# 运行所有 TUI 测试
make unit-test-tui

# 运行特定包测试
go test ./tui/framework/component -v

# 运行特定测试
go test ./tui/framework/component -v -run TestBaseComponent_ID

# 测试并生成覆盖率
go test ./tui/framework/component -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# 并行测试 (加快速度)
go test ./tui/... -parallel 8

# 跳过慢速测试
go test ./tui/... -skip=Benchmark

# 运行基准测试
go test ./tui/... -bench=. -benchmem
```

### 4. 测试最佳实践

```go
// 1. 使用 assert 库
assert.Equal(t, expected, actual)
assert.NotNil(t, value)
assert.True(t, condition)
assert.NoError(t, err)

// 2. 使用 require 处理致命错误
require.NotNil(t, setup) // 如果失败，停止测试

// 3. Mock 接口
type MockPainter struct {
    calls []string
}

func (m *MockPainter) SetCell(x, y int, char rune) {
    m.calls = append(m.calls, fmt.Sprintf("SetCell(%d,%d,%c)", x, y, char))
}

// 4. 测试并发
func TestConcurrentAccess(t *testing.T) {
    registry := NewComponentRegistry()
    done := make(chan bool)

    for i := 0; i < 100; i++ {
        go func(id string) {
            registry.Add(NewComponent(id))
            done <- true
        }(fmt.Sprintf("comp-%d", i))
    }

    for i := 0; i < 100; i++ {
        <-done
    }

    assert.Equal(t, 100, registry.Count())
}
```

---

## 任务管理

### 1. 任务分解原则

```go
// 大任务分解为小任务 (每个任务 < 2 小时)

// ❌ 太大的任务
"实现 Table 组件"

// ✅ 分解后的小任务
1. "实现 Table 组件的 Measurable 接口"
2. "实现 Table 组件的 Paintable 接口"
3. "实现 Table 组件的 ActionTarget 接口"
4. "实现 Table 组件的 Focusable 接口"
5. "编写 Table 组件的单元测试"
6. "更新 Table 组件文档"
```

### 2. 任务状态

```go
type TaskStatus string

const (
    TaskPending    TaskStatus = "pending"     // 待开始
    TaskInProgress TaskStatus = "in_progress" // 进行中
    TaskCompleted  TaskStatus = "completed"   // 已完成
    TaskBlocked    TaskStatus = "blocked"    // 被阻塞
    TaskFailed     TaskStatus = "failed"     // 失败
)

type Task struct {
    ID          string
    Title       string
    Description string
    Status      TaskStatus
    DependsOn   []string  // 依赖的任务 ID
    Priority    int       // 优先级 1-10
    Estimated   int       // 预估时间 (分钟)
}
```

### 3. 典型任务模板

#### 任务: 实现新组件

```
标题: 实现 [ComponentName] 组件

描述:
- 实现 Capability Interfaces: Measurable, Paintable
- 支持配置属性: [列出属性]
- 支持事件: [列出事件]

步骤:
1. 阅读 ARCHITECTURE.md 和 COMPONENTS.md
2. 创建 tui/framework/display/[component].go
3. 实现接口
4. 创建测试文件 [component]_test.go
5. 运行测试确保通过
6. 更新 COMPONENTS.md

验收标准:
- [ ] 所有接口实现
- [ ] 测试覆盖率 > 80%
- [ ] 文档已更新
- [ ] Demo 示例运行正常
```

#### 任务: 修复 Bug

```
标题: 修复 [BugDescription]

描述:
- 问题描述: [详细描述]
- 复现步骤: [列出步骤]
- 期望行为: [描述]

步骤:
1. 编写失败的测试用例
2. 定位问题代码
3. 修复问题
4. 确保测试通过
5. 检查是否有类似问题

验收标准:
- [ ] Bug 已修复
- [ ] 测试用例通过
- [ ] 无回归问题
```

---

## 文档管理

### 1. 文档结构

```
tui/framework/docs/
├── ARCHITECTURE.md              # 架构总览 (入口文档)
├── ARCHITECTURE_INVARIANTS.md   # 架构不变量 (必须遵守)
├── BOUNDARIES.md                # 层级边界
├── COMPONENTS.md                # 组件系统
├── ACTION_SYSTEM.md             # Action 系统
├── FOCUS_SYSTEM.md              # Focus 系统
├── EVENT_SYSTEM.md              # 事件系统
├── RENDERING.md                 # 渲染系统
├── STATE_MANAGEMENT.md          # 状态管理
├── ANIMATION_SYSTEM.md          # 动画系统
├── AI_INTEGRATION.md            # AI 集成
├── THEME_SYSTEM.md              # 主题系统
├── VIRTUAL_SCROLL.md            # 虚拟滚动
├── ERROR_HANDLING.md            # 错误处理
├── PAINTER_ABSTRACTION.md       # Painter 抽象
├── ASYNC_TASK.md                # 异步任务
├── TABLE_SUBFOCUS.md            # Table 子焦点
├── STREAM_DATA.md               # 流式数据
├── V8_INTEGRATION_YAO.md        # V8 集成
├── V8_EVENT_CALLBACK.md         # V8 事件回调
├── PROCESS_INTEGRATION_YAO.md   # Yao Process 集成
└── FORM_VALIDATION.md           # Form 验证
```

### 2. 文档编写规范

```markdown
# 文档标题

> **元信息**: 版本、优先级、状态

## 概述
简要描述文档内容

## 核心概念
解释关键概念和术语

## 架构/设计
使用图表说明架构

## API/接口
代码示例和接口定义

## 使用示例
完整的使用示例

## 最佳实践
推荐的做法和禁忌

## 相关文档
引用相关文档链接
```

### 3. 代码注释规范

```go
// Package comment (写在包声明前)
// Package component 提供了 TUI 框架的核心组件接口。
//
// 组件基于 Capability Interfaces 设计，每个组件只实现它需要的能力。
package component

// 类型注释
// BaseComponent 提供了组件的基础实现。
//
// 它实现了 Node 接口，并提供了属性管理和状态管理功能。
type BaseComponent struct {
    // id 是组件的唯一标识符
    id string

    // props 存储组件属性
    props map[string]interface{}
}

// 函数注释 (简洁描述功能和参数)
// NewBaseComponent 创建一个新的基础组件。
//
// 参数:
//   id - 组件的唯一标识符
//
// 返回:
//   初始化后的 BaseComponent 实例
func NewBaseComponent(id string) *BaseComponent {
    return &BaseComponent{
        id:    id,
        props: make(map[string]interface{}),
    }
}

// 重要逻辑的行内注释
// 检查组件是否可聚焦
if c.focusID != "" {
    // 已有焦点，无需重复设置
    return nil
}
```

---

## 常用命令

### 编译

```bash
# 方式 1: 使用 build-noui 脚本 (推荐用于开发调试)
# Windows PowerShell
.\build-noui.ps1

# Linux/Mac
./build-noui.sh

# 方式 2: 使用 Makefile
make no-ui              # 开发编译 (无 UI)
make release            # 生产编译
make linux-release      # Linux 编译
make artifacts-linux    # 交叉编译

# 方式 3: 直接编译
go build -o yao.exe cmd/main.go           # 主程序
go build -o yao-tui.exe cmd/tui/tui.go    # TUI 单独命令

# 编译输出位置
# Windows: %GOPATH%\bin\yao.exe
# Linux/Mac: $GOPATH/bin/yao
```

### TUI 命令

```bash
# TUI 有独立的命令入口 (cmd/tui/)
yao tui help          # 查看帮助
yao tui list          # 列出所有 TUI 界面
yao tui check         # 检查 TUI 配置
yao tui validate      # 验证 TUI 配置
yao tui inspect       # 检查 TUI 组件
yao tui dump          # 导出 TUI 配置
```

### 运行 Demo 应用

```bash
# Demo 应用位置
# Windows: E:\projects\yao\wwsheng009\yao\tui\demo
# Linux/Mac: /e/projects/yao/yao-projects/yao-docs/docs/YaoApps/tui_app

# 设置环境变量
export YAO_ROOT="E:/projects/yao/wwsheng009/yao"
export YAO_TEST_APPLICATION="E:/projects/yao/wwsheng009/yao/tui/demo"
export YAO_ENV="development"
export YAO_LOG_CONSOLE="true"

# 方式 1: 启动 Yao 服务器 (访问 TUI)
cd E:\projects\yao\wwsheng009\yao\tui\demo
yao start

# 方式 2: 运行 TUI 测试
cd E:\projects\yao\wwsheng009\yao
go test ./tui -v -run Demo

# 方式 3: 运行 Demo 测试文件
cd E:\projects\yao\wwsheng009\yao\tui\demo
go test -v -run Demo
```

### 测试

```bash
# TUI 全部测试
make unit-test-tui

# TUI 组件测试
make unit-test-tui-components

# TUI 核心测试
make unit-test-tui-core

# 单个包测试
go test ./tui/framework/component -v

# 单个测试
go test ./tui/framework/component -v -run TestNode_ID

# 覆盖率
go test ./tui/framework/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### 代码质量

```bash
# 格式化
make fmt
# 或
gofmt -s -w .

# 静态检查
make vet
# 或
go vet ./...

# Lint
make lint

# 拼写检查
make misspell-check
```

### 运行

```bash
# 启动服务器
yao start

# 运行脚本
yao run scripts.user.Hello

# 运行 TUI Demo
cd E:\projects\yao\wwsheng009\yao\tui\demo
yao start
```

### Git

```bash
# 查看状态
git status

# 查看差异
git diff

# 提交
git add .
git commit -m "feat(component): add new component"

# 推送
git push
```

---

## 故障排查

### 1. 测试失败

```bash
# 问题: 测试超时
# 解决: 增加超时时间
go test -timeout 30m

# 问题: 并发测试失败
# 解决: 串行运行测试
go test -parallel 1

# 问题: 环境变量未设置
# 解决: 检查 .env 文件
cat .env
```

### 2. 编译错误

```bash
# 问题: 找不到包
# 解决: 更新依赖
go mod tidy

# 问题: 循环依赖
# 解决: 检查 import 顺序，使用依赖注入
```

### 3. 日志问题

```bash
# 启用控制台输出
export YAO_LOG_CONSOLE=true

# 设置日志级别
export YAO_LOG_LEVEL=TRACE  # TRACE, DEBUG, INFO, WARN, ERROR

# 检查日志文件
tail -f application.log
```

---

## AI 开发清单

### 开始新任务前检查

- [ ] 阅读相关设计文档
- [ ] 确认任务边界和依赖
- [ ] 设置环境变量
- [ ] 拉取最新代码

### 编写代码时检查

- [ ] 遵循代码组织规范
- [ ] 实现必要的接口
- [ ] 添加错误处理
- [ ] 添加日志输出
- [ ] 添加代码注释

### 提交代码前检查

- [ ] 运行单元测试
- [ ] 运行 go vet
- [ ] 运行 gofmt
- [ ] 更新文档
- [ ] 编写提交信息

### 提交信息格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

类型 (type):
- `feat`: 新功能
- `fix`: 修复 Bug
- `docs`: 文档更新
- `style`: 代码格式 (不影响功能)
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建/工具相关

示例:
```
feat(component): add Focusable interface

- Add FocusID() method
- Add OnFocus() and OnBlur() hooks
- Update documentation

Closes #123
```

---

## 附录

### Yao 框架能力

TUI 基于 Yao 框架，可以使用以下功能：

1. **Process 系统**: 调用业务逻辑
2. **V8 集成**: 使用 JavaScript 编写组件
3. **数据模型**: 使用 Yao 的数据模型
4. **流系统**: 处理流式数据
5. **验证系统**: 表单验证

### 相关资源

- [Yao 文档](https://yaoapps.com/docs)
- [TUI 架构文档](./ARCHITECTURE.md)
- [组件系统](./COMPONENTS.md)
- [开发示例](../examples/)
