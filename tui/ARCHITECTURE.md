# TUI 引擎架构设计

## 1. 总体架构

### 1.1 设计理念

Yao TUI 引擎遵循以下核心原则：

- **声明式优先**: 通过 DSL 配置驱动 UI，减少命令式代码
- **响应式状态**: 单向数据流，State 驱动 View 更新
- **非阻塞执行**: 所有耗时操作异步执行，保持 UI 流畅
- **组件化设计**: 标准组件库 + 可扩展机制
- **AI 原生**: 深度集成 Yao 的 AIGC 能力

### 1.2 技术栈

| 组件 | 技术 | 作用 |
|------|------|------|
| TUI 框架 | Bubble Tea v0.25.0 | 事件循环和生命周期管理 |
| 样式系统 | Lip Gloss v0.9.1 | 终端样式和布局 |
| 组件库 | Bubbles v0.17.1 | 标准交互组件 |
| Markdown | Glamour v0.6.0 | AI 内容渲染 |
| JS 运行时 | V8Go (Yao 集成) | 脚本执行 |
| 测试 | testify v1.8.4 | 单元测试 |

### 1.3 架构分层

```
┌─────────────────────────────────────────────┐
│          CLI Layer (cmd/tui.go)             │
│   命令行入口、参数解析、信号处理            │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│        Process Layer (tui/process.go)       │
│   导出 Yao Process、与其他模块集成          │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│       Core Engine (tui/driver.go)           │
│   Bubble Tea 生命周期、消息循环             │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────┬─────────────┬─────────────┐
│  State Manager  │   Action    │   Renderer  │
│   (model.go)    │ (action.go) │ (render.go) │
│   状态管理      │  行为执行   │  视图渲染   │
└─────────────────┴─────────────┴─────────────┘
         ↓                ↓              ↓
┌─────────────────┬─────────────┬─────────────┐
│  V8 Runtime     │  Component  │   Theme     │
│  (script.go)    │  (components)│ (theme.go) │
│  脚本执行       │  组件库     │  样式系统   │
└─────────────────┴─────────────┴─────────────┘
```

---

## 2. 核心模块设计

### 2.1 DSL 加载器 (loader.go)

**职责**:
- 扫描 `tuis/` 目录
- 解析 `.tui.yao` 配置文件
- 建立 ID 映射和缓存

**关键函数**:
```go
func Load() error
func Get(id string) (*Config, error)
func LoadScript(file string) (*Script, error)
```

**数据流**:
```
tuis/*.tui.yao → JSON Parser → Config Struct → sync.Map Cache
```

### 2.2 状态管理 (model.go)

**职责**:
- 实现 `tea.Model` 接口
- 管理响应式状态
- 处理消息循环

**核心结构**:
```go
type Model struct {
    Config   *Config
    State    map[string]interface{}
    StateMu  sync.RWMutex  // 并发安全
    Width    int
    Height   int
    Ready    bool
    Program  *tea.Program
}
```

**生命周期**:
```
Init() → Update(msg) → View() → [循环]
```

### 2.3 Action 执行器 (action.go)

**职责**:
- 解析 Action 配置
- 执行 Process 或 Script
- 处理异步回调

**执行流程**:
```
Action Config → Parse Args ({{}} 插值)
              ↓
         Execute Process/Script (async)
              ↓
         Return tea.Msg
              ↓
         Update State
```

### 2.4 渲染引擎 (render.go)

**职责**:
- 递归渲染 Layout 树
- 应用 Lip Gloss 样式
- 组件实例化和组合

**渲染流程**:
```
Layout Tree → Traverse Children
            ↓
       Component Factory (type → instance)
            ↓
       Apply State Binding ({{path}})
            ↓
       Lip Gloss Style
            ↓
       Join Vertical/Horizontal
```

### 2.5 V8 集成 (script.go + jsapi.go)

**职责**:
- 加载和编译 TS/JS 脚本
- 注入 TUI 对象到 JS 上下文
- 提供 State 操作 API

**JavaScript API**:
```typescript
interface TUI {
    GetState(key: string): any;
    SetState(key: string, value: any): void;
    UpdateState(updates: object): void;
    ExecuteAction(action: Action): void;
    Release(): void;
}
```

**集成模式**:
```
Go Model → bridge.RegisterGoObject → goValueID
                                          ↓
                            V8 Internal Field (隐藏)
                                          ↓
                            JS Function Call
                                          ↓
                            bridge.GetGoObject → Go Model
```

---

## 3. 组件系统

### 3.1 标准组件

| 组件 | 文件 | 功能 |
|------|------|------|
| Header | components/header.go | 标题栏 |
| Text | components/text.go | 文本显示 |
| Table | components/table.go | 数据表格 |
| Form | components/form.go | 表单输入 |
| Input | components/input.go | 单行输入 |
| Viewport | components/viewport.go | 滚动视图 |
| Chat | components/chat.go | AI 聊天（流式） |

### 3.2 组件接口

```go
type Component interface {
    Init(props map[string]interface{}) tea.Cmd
    Update(msg tea.Msg) (Component, tea.Cmd)
    View() string
    SetProps(props map[string]interface{})
    ID() string
}
```

### 3.3 组件注册

```go
var componentRegistry = map[string]ComponentFactory{
    "header": NewHeader,
    "table":  NewTable,
    "chat":   NewChat,
}

func RegisterComponent(name string, factory ComponentFactory) {
    componentRegistry[name] = factory
}
```

---

## 4. 数据流

### 4.1 单向数据流

```
User Input → KeyMsg
                ↓
         Match Bindings
                ↓
         Execute Action
                ↓
         Process/Script Execution
                ↓
         ProcessResultMsg
                ↓
         Update State
                ↓
         Trigger Re-render
                ↓
         View() → Terminal Output
```

### 4.2 消息类型

```go
// Bubble Tea 内置消息
tea.KeyMsg          // 键盘输入
tea.MouseMsg        // 鼠标事件
tea.WindowSizeMsg   // 窗口大小

// TUI 自定义消息
ProcessResultMsg    // Process 执行结果
StateUpdateMsg      // 单个状态更新
StateBatchUpdateMsg // 批量状态更新
StreamChunkMsg      // AI 流式数据块
StreamDoneMsg       // 流式完成
ErrorMsg            // 错误消息
```

---

## 5. 性能优化

### 5.1 脚本预编译

```go
// 应用启动时预编译所有脚本
func PrecompileScripts() error {
    return application.App.Walk("scripts/tui", ...)
}
```

### 5.2 Context 池化

```go
type ContextPool struct {
    pool sync.Pool
}

// 复用 V8 Context，减少创建开销
```

### 5.3 渲染优化

- **静态内容缓存**: Header/Footer 等静态部分只渲染一次
- **增量更新**: 仅重绘变化的组件
- **虚拟滚动**: 大数据集只渲染可见区域

### 5.4 性能目标

| 操作 | 目标延迟 |
|------|----------|
| ModelUpdate | < 100ns |
| RenderLayout (3组件) | < 10µs |
| StateRead | < 50ns |
| StateWrite | < 100ns |
| ScriptExecution | < 1ms |

---

## 6. 安全设计

### 6.1 并发安全

```go
// State 访问必须加锁
model.StateMu.RLock()
value := model.State[key]
model.StateMu.RUnlock()
```

### 6.2 脚本沙箱

- V8 Isolate 隔离
- 内存限制: 50MB/脚本
- 执行超时: 5秒
- 禁止文件系统直接访问

### 6.3 输入验证

```go
// 所有来自 Input 组件的数据需要 sanitize
func SanitizeInput(input string) string {
    // 防止注入攻击
}
```

---

## 7. 扩展性

### 7.1 自定义组件

开发者可以注册自定义组件：

```go
// 在 init() 中注册
func init() {
    tui.RegisterComponent("my-component", NewMyComponent)
}
```

### 7.2 插件机制

```go
type Plugin interface {
    Name() string
    Init(tui *TUI) error
    Hooks() map[string]HookFunc
}
```

### 7.3 主题系统

```go
type Theme struct {
    Primary   lipgloss.Color
    Secondary lipgloss.Color
    Warning   lipgloss.Color
    Error     lipgloss.Color
}

// 支持自定义主题
tui.SetTheme(myTheme)
```

---

## 8. 监控和诊断

### 8.1 性能指标

```go
type Metrics struct {
    TotalRenders      int64
    TotalActions      int64
    TotalErrors       int64
    AvgRenderTime     time.Duration
    ActiveInstances   int64
}
```

### 8.2 健康检查

```go
func HealthCheck() HealthStatus {
    // 检查活跃实例数、错误率等
}
```

### 8.3 调试工具

```bash
# 启用调试日志
export YAO_TUI_DEBUG=true

# 状态快照
tui.DumpState()

# 性能分析
go tool pprof cpu.prof
```

---

## 9. 测试策略

### 9.1 单元测试

- 覆盖率目标: 85%+
- 所有核心模块必须有测试
- 使用 Mock Program 和 Mock Process

### 9.2 集成测试

```go
func TestFullLifecycle(t *testing.T) {
    // 测试从加载到渲染的完整流程
}
```

### 9.3 性能测试

```go
func BenchmarkRenderLayout(b *testing.B) {
    // 基准测试
}
```

### 9.4 并发测试

```bash
go test ./tui/... -race
```

---

## 10. 部署架构

### 10.1 构建流程

```
Go Source → go build (CGO_ENABLED=1)
           ↓
     Embed Resources (go-bindata)
           ↓
     Strip & Compress
           ↓
     Multi-platform Binaries
```

### 10.2 资源打包

```bash
# 将 tuis/ 目录打包到二进制
go-bindata -fs -pkg data -o data/bindata.go tuis/...
```

### 10.3 运行时要求

- Go >= 1.21
- CGO 支持（V8Go 需要）
- 终端支持 256 色或 TrueColor

---

## 11. 版本规划

### v0.1.0 (MVP) - 2 周
- ✅ 核心框架
- ✅ 基础组件 (Header, Text)
- ✅ DSL 加载
- ✅ 简单示例

### v0.2.0 (V8 集成) - 3 周
- ✅ 脚本加载器
- ✅ JavaScript API
- ✅ Action 执行器
- ✅ Counter 示例

### v0.3.0 (组件库) - 3 周
- ✅ Table 组件
- ✅ Form 组件
- ✅ Viewport 组件
- ✅ Chat 组件（AI 流式）

### v1.0.0 (生产就绪) - 2 周
- ✅ 完整测试套件
- ✅ 性能优化
- ✅ 文档完善
- ✅ 生产部署

---

## 12. 参考资料

### 内部参考
- `pipe/pipe.go` - DSL 加载模式
- `sui/core/script.go` - 脚本加载
- `agent/context/jsapi.go` - JavaScript API 模式
- `trace/jsapi/trace.go` - Bridge 注册模式

### 外部参考
- [Bubble Tea 文档](https://github.com/charmbracelet/bubbletea)
- [Lip Gloss 指南](https://github.com/charmbracelet/lipgloss)
- [V8Go API](https://pkg.go.dev/rogchap.com/v8go)

---

## 13. 风险和挑战

### 13.1 技术风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| V8Go 稳定性 | 高 | 添加错误恢复机制 |
| 终端兼容性 | 中 | 自动降级样式 |
| 性能瓶颈 | 中 | 实施缓存和池化 |

### 13.2 开发挑战

- **并发安全**: 严格使用互斥锁
- **内存管理**: V8 隔离的生命周期管理
- **调试困难**: 提供完善的日志和诊断工具

---

## 附录

### A. DSL 配置示例

```json
{
  "name": "完整示例",
  "data": {
    "title": "Dashboard",
    "users": []
  },
  "onLoad": {
    "process": "models.user.Get",
    "args": [{"limit": 10}],
    "onSuccess": "users"
  },
  "layout": {
    "direction": "vertical",
    "children": [
      {
        "type": "header",
        "props": {"title": "{{title}}"}
      },
      {
        "type": "table",
        "bind": "users",
        "props": {
          "columns": [
            {"name": "id", "width": 10},
            {"name": "name", "width": 20}
          ]
        }
      }
    ]
  },
  "bindings": {
    "r": {
      "process": "models.user.Get",
      "onSuccess": "users"
    },
    "q": {"process": "tui.Quit"}
  }
}
```

### B. TypeScript 类型定义

```typescript
declare namespace Yao {
  interface TUI {
    GetState(key: string): any;
    SetState(key: string, value: any): void;
    UpdateState(updates: Record<string, any>): void;
    ExecuteAction(action: Action): void;
  }
  
  interface Action {
    process?: string;
    script?: string;
    method?: string;
    args?: any[];
    onSuccess?: string;
  }
}
```

---

**文档版本**: v1.0  
**最后更新**: 2026-01-16  
**维护者**: Yao TUI Team
