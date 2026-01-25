# AI 自动开发 TODO List (V3)

> **目标**: 让 AI 能够自动完成 TUI 框架的开发任务，包括代码编写、测试、文档管理
> **使用方式**: AI 逐个执行任务，每个任务完成后标记为 completed
> **工作目录**: `E:\projects\yao\wwsheng009\yao`

---

## 使用说明

### 给 AI 的提示词模板

```
你是一个专业的 Go 语言 TUI 框架开发专家。你的任务是根据 TODO_LIST 完成开发工作。

在开始任何任务前，你必须：
1. 阅读"文档引用"中列出的所有文档
2. 理解架构不变量 (ARCHITECTURE_INVARIANTS.md)
3. 确认任务边界，不超出范围
4. 编写符合规范的代码

完成任务后：
1. 运行测试确保通过
2. 更新相关文档
3. 将任务标记为 completed

当前环境：
- Go 版本: 1.21+
- 工作目录: E:\projects\yao\wwsheng009\yao
- Demo 目录: E:\projects\yao\wwsheng009\yao\tui\demo
```

---

## Phase 0: Foundation (基础建设)

### ✅ 任务 0.1: 创建目录结构

**状态**: `pending`
**优先级**: `P0`
**预估时间**: `30 分钟`

**文档引用**:
- [ARCHITECTURE.md](./ARCHITECTURE.md) - 第 145-307 行 (目录结构)

**AI 提示词**:
```
任务：创建 TUI Framework 目录结构

根据 ARCHITECTURE.md 第 145-307 行定义的目录结构，创建以下目录：

tui/framework/
├── docs/                        # 已存在
├── component/                   # 组件定义
├── runtime/                     # Runtime 内核 (纯 Go，无外部依赖)
│   ├── layout/                  # 布局引擎
│   ├── paint/                   # 绘制系统
│   ├── focus/                   # 焦点系统
│   ├── input/                   # 输入处理
│   ├── action/                  # Action 系统
│   ├── animation/               # 动画系统
│   └── state/                   # 状态管理
├── platform/                    # 平台抽象
│   └── impl/                    # 平台实现
│       ├── default/             # 默认实现
│       └── windows/             # Windows 实现
├── screen/                      # Framework 屏幕管理
├── display/                     # 显示组件
├── input/                       # 输入组件
├── layout/                      # 布局组件
├── interactive/                 # 交互组件
├── overlay/                     # 覆盖层组件
├── form/                        # 表单组件
├── style/                       # 样式系统
├── validation/                  # 验证系统
├── async/                       # 异步任务系统
├── stream/                      # 流式数据系统
├── result/                      # Result 类型
├── paint/                       # Painter 抽象
├── v8/                          # V8 集成
├── event/                       # Framework 事件
└── app.go                       # 应用入口

注意事项：
1. 使用 mkdir -p 创建目录
2. 每个目录创建一个 README.md 说明文件用途
3. 确保 runtime/ 目录是纯 Go 代码，不依赖任何外部包
```

**验收标准**:
- [ ] 所有目录创建完成
- [ ] 每个目录有 README.md
- [ ] 目录结构与 ARCHITECTURE.md 一致

---

### ✅ 任务 0.2: 定义核心接口

**状态**: `pending`
**优先级**: `P0`
**预估时间**: `1 小时`

**文档引用**:
- [ARCHITECTURE.md](./ARCHITECTURE.md) - 第 309-383 行 (核心接口定义)
- [BOUNDARIES.md](./BOUNDARIES.md) - 接口边界
- [ARCHITECTURE_INVARIANTS.md](./ARCHITECTURE_INVARIANTS.md) - 不变量 1

**AI 提示词**:
```
任务：定义核心接口

根据 ARCHITECTURE.md 第 309-383 行，创建以下核心接口文件：

1. component/node.go
   - Node 接口 (ID(), Type())

2. component/mountable.go
   - Mountable 接口 (Mount(), Unmount(), IsMounted())

3. component/measurable.go
   - Measurable 接口 (Measure(), GetSize())

4. component/paintable.go
   - Paintable 接口 (Paint(ctx PaintContext, buf *runtime.CellBuffer))

5. component/actionable.go
   - ActionTarget 接口 (HandleAction(a *runtime.Action) bool)

6. component/focusable.go
   - Focusable 接口 (FocusID(), OnFocus(), OnBlur())

7. component/scrollable.go
   - Scrollable 接口 (ScrollTo(), ScrollBy(), GetScrollPosition())

8. component/validatable.go
   - Validatable 接口 (Validate(), IsValid())

9. component/base.go
   - BaseComponent 组合接口

10. component/container.go
    - Container 接口

注意事项：
1. 必须遵循 Capability Interfaces 设计原则
2. 每个接口都是独立的，小而专注
3. 不使用胖接口
4. 添加详细的 godoc 注释
5. 严格遵守 ARCHITECTURE_INVARIANTS.md 不变量 1：Runtime 永远不知道组件是什么

代码示例参考 ARCHITECTURE.md 第 313-383 行。
```

**验收标准**:
- [ ] 所有接口文件创建完成
- [ ] 接口定义与文档一致
- [ ] 有完整的 godoc 注释
- [ ] Runtime 包不依赖 Framework

---

### ✅ 任务 0.3: 实现 Platform 抽象

**状态**: `pending`
**优先级**: `P0`
**预估时间**: `2 小时`

**文档引用**:
- [BOUNDARIES.md](./BOUNDARIES.md) - 第 82-266 行 (Platform 层接口)
- [ARCHITECTURE_INVARIANTS.md](./ARCHITECTURE_INVARIANTS.md) - 不变量 6

**AI 提示词**:
```
任务：实现 Platform 抽象层

根据 BOUNDARIES.md 第 82-266 行，实现 Platform 层接口：

1. platform/screen.go
   - Screen 接口定义
   - DefaultScreen 实现 (Unix)

2. platform/cursor.go
   - Cursor 接口定义
   - CursorStyle 类型

3. platform/input.go
   - InputReader 接口定义
   - RawInput 类型定义
   - SpecialKey、KeyModifier、MouseButton 等常量

4. platform/signal.go
   - SignalHandler 接口定义
   - DefaultSignalHandler 实现

5. platform/impl/default/screen.go
   - Unix 屏幕实现

6. platform/impl/default/cursor.go
   - Unix 光标实现

7. platform/impl/default/input.go
   - Unix 输入实现，使用 ANSI 转义码

注意事项：
1. Platform 层只提供"能力抽象"，不包含"语义"
2. 不理解 Focus、Event、Component、Layout
3. 使用 os、syscall 等标准库
4. 添加完整的错误处理
5. 支持 Linux/macOS 平台

严格遵守 ARCHITECTURE_INVARIANTS.md：
- Platform 绝不依赖 Framework 或 Runtime
- Platform 只提供原始输入/输出能力
```

**验收标准**:
- [ ] Screen 接口实现并能初始化/清屏
- [ ] Cursor 接口实现并能移动光标
- [ ] InputReader 能读取键盘输入
- [ ] SignalHandler 能处理信号
- [ ] 平台抽象测试通过

---

### ✅ 任务 0.4: 实现 Action 系统

**状态**: `pending`
**优先级**: `P0`
**预估时间**: `2 小时`

**文档引用**:
- [ACTION_SYSTEM.md](./ACTION_SYSTEM.md) - 完整 Action 系统设计
- [ARCHITECTURE_INVARIANTS.md](./ARCHITECTURE_INVARIANTS.md) - 不变量 2, 6

**AI 提示词**:
```
任务：实现 Action 系统

根据 ACTION_SYSTEM.md，实现 Runtime Action 系统：

1. runtime/action/action.go
   - ActionType 类型及常量（导航、编辑、选择、光标、表单、鼠标、窗口、视图、系统、AI）
   - Action 结构体
   - NewAction、WithPayload、WithTarget、WithSource、Clone 方法

2. runtime/action/dispatcher.go
   - Dispatcher 结构体
   - Dispatch() 方法
   - Register/Unregister 目标
   - Subscribe/Unsubscribe 全局处理器
   - SetFocusManager
   - 日志记录和统计

3. runtime/action/log.go
   - Logger 结构体
   - LogEntry 结构体
   - 日志记录和回放功能

注意事项：
1. Action 是语义事件，不是原始按键
2. 所有状态变化必须能追溯到 Action
3. 支持全局处理器和目标处理器
4. 记录所有 Action 以便回放
5. 严格遵守 Input ≠ Action 不变量

测试要求：
- 测试 Action 创建和克隆
- 测试 Dispatcher 分发
- 测试全局处理器
- 测试 Action 日志
```

**验收标准**:
- [ ] Action 类型定义完整
- [ ] Dispatcher 能正确分发 Action
- [ ] 支持 Action 日志记录
- [ ] 单元测试覆盖率 > 80%

---

### ✅ 任务 0.5: 实现 KeyMap 输入转换

**状态**: `pending`
**优先级**: `P0`
**预估时间**: `2 小时`

**文档引用**:
- [ACTION_SYSTEM.md](./ACTION_SYSTEM.md) - 第 402-659 行 (KeyMap 定义)

**AI 提示词**:
```
任务：实现 KeyMap 输入转换系统

根据 ACTION_SYSTEM.md 第 402-659 行，实现：

1. runtime/input/raw.go
   - RawInputType、SpecialKey、KeyModifier 类型
   - RawInput 结构体

2. runtime/input/keymap.go
   - KeyMap 结构体
   - Map() 方法将 RawInput 转换为 Action
   - Bind/Unbind 自定义按键
   - BindContext/ PushContext/ PopContext 上下文映射
   - makeKey() 构建映射键

3. runtime/input/processor.go
   - InputProcessor 结构体
   - Parse() 方法解析 ANSI 转义序列
   - Process() 方法处理输入流

默认按键映射（参考文档）：
- Tab → ActionNavigateNext
- Shift+Tab → ActionNavigatePrev
- Up/Down/Left/Right → 方向导航
- Enter → ActionSubmit
- Escape → ActionCancel
- Backspace/Delete → ActionDeleteChar
- Ctrl+C/Q → ActionQuit
- 等等...

注意事项：
1. Platform 只产生 RawInput
2. Runtime 负责转换 RawInput → Action
3. Component 只处理 Action
4. 支持上下文相关按键映射
5. 支持组合键（Ctrl、Alt、Shift）

测试要求：
- 测试所有默认按键映射
- 测试组合键
- 测试上下文映射
- 测试自定义绑定
```

**验收标准**:
- [ ] RawInput → Action 转换正确
- [ ] 所有默认映射工作正常
- [ ] 支持上下文映射
- [ ] 单元测试通过

---

### ✅ 任务 0.6: 实现状态管理

**状态**: `pending`
**优先级**: `P0`
**预估时间**: `3 小时`

**文档引用**:
- [STATE_MANAGEMENT.md](./STATE_MANAGEMENT.md) - 完整状态管理设计
- [ARCHITECTURE_INVARIANTS.md](./ARCHITECTURE_INVARIANTS.md) - 不变量 5, 10

**AI 提示词**:
```
任务：实现状态管理系统

根据 STATE_MANAGEMENT.md，实现 Runtime 状态管理：

1. runtime/state/snapshot.go
   - Snapshot 结构体（FocusPath、Components、Modals、Dirty）
   - ComponentState 结构体（ID、Type、Props、State、Rect、Visible、Disabled）
   - Rect、ModalState、DirtyRegion 类型
   - NewSnapshot、Clone、Equal、GetComponent、SetComponent 方法

2. runtime/state/tracker.go
   - Tracker 结构体（current、past、future、listeners）
   - Current() 获取当前状态
   - BeforeAction() 记录执行前状态
   - AfterAction() 记录执行后状态
   - Undo/Redo/C Undo/CanRedo
   - Subscribe/Unsubscribe 状态变化监听

3. runtime/state/diff.go
   - Diff 结构体
   - ComputeDiff() 计算两个快照的差异

4. runtime/state/serialize.go
   - Serialize/Deserialize JSON 序列化
   - SaveToFile/LoadFromFile

注意事项：
1. 所有状态必须可枚举、可快照、可追溯
2. 不允许隐式全局状态
3. 支持 Undo/Redo
4. 支持状态变化监听
5. 支持状态序列化

严格遵守 ARCHITECTURE_INVARIANTS.md：
- 不变量 5：没有隐式全局状态
- 不变量 10：状态必须是显式的
```

**验收标准**:
- [ ] 状态可以完整快照
- [ ] 支持 Undo/Redo
- [ ] 支持状态监听
- [ ] 支持 JSON 序列化
- [ ] 单元测试覆盖率 > 80%

---

### ✅ 任务 0.7: 实现布局引擎

**状态**: `pending`
**优先级**: `P0`
**预估时间**: `4 小时`

**文档引用**:
- [ARCHITECTURE.md](./ARCHITECTURE.md) - 第 276-312 行 (Layout Engine)
- [RENDERING.md](./RENDERING.md) - 渲染管线

**AI 提示词**:
```
任务：实现 Flexbox 布局引擎

根据 ARCHITECTURE.md 第 276-312 行，实现 Runtime 布局引擎：

1. runtime/layout/engine.go
   - Engine 结构体
   - Layout() 计算布局
   - Measure() 测量尺寸

2. runtime/layout/flex.go
   - FlexDirection 类型 (Row, Column)
   - AlignItems, JustifyContent 类型
   - Flex 布局算法

3. runtime/layout/constraint.go
   - Constraints 结构体 (minWidth, maxWidth, minHeight, maxHeight)
   - Apply() 应用约束

4. runtime/layout/measure.go
   - Size 结构体
   - Measurable 接口

布局规则：
- 支持主轴（main axis）和交叉轴（cross axis）
- 支持对齐方式
- 支持约束传递
- 支持百分比和固定尺寸
- 支持 flex-grow

注意事项：
1. Runtime 是纯内核，不依赖 Framework
2. 只处理抽象 Node，不处理具体 Component
3. 支持嵌套布局
4. 性能优化：只在 dirty 时重新计算

测试要求：
- 测试水平布局
- 测试垂直布局
- 测试嵌套布局
- 测试约束传递
- 测试百分比和固定尺寸
```

**验收标准**:
- [ ] Flexbox 布局正确工作
- [ ] 支持嵌套布局
- [ ] 约束系统工作正常
- [ ] 性能测试通过
- [ ] 单元测试覆盖率 > 80%

---

### ✅ 任务 0.8: 实现绘制系统

**状态**: `pending`
**优先级**: `P0`
**预估时间**: `3 小时`

**文档引用**:
- [ARCHITECTURE.md](./ARCHITECTURE.md) - 第 314-357 行 (Paint Engine)
- [RENDERING.md](./RENDERING.md) - 渲染系统

**AI 提示词**:
```
任务：实现绘制系统

根据 ARCHITECTURE.md 第 314-357 行，实现 Runtime 绘制系统：

1. runtime/paint/cell.go
   - Cell 结构体（Char、Style、FG、BG、Bold等）
   - CellStyle 结构体

2. runtime/paint/buffer.go
   - CellBuffer 结构体
   - SetCell、GetCell、Fill、Clear 方法
   - Width、Height 方法

3. runtime/paint/tree.go
   - RenderNode 结构体（ID、Bounds、Z、Paint、Children）
   - RenderTree 结构

4. runtime/paint/dirty.go
   - DirtyTracker 结构体
   - Mark、MarkRegion、MarkAll 方法
   - GetRegions、Clear 方法

5. runtime/paint/diff.go
   - DiffEngine 结构体
   - Compute 计算差异
   - DiffChange 结构体

注意事项：
1. Paint 不返回 string，而是写入 CellBuffer
2. 使用 RenderTree 作为中间态
3. 支持 Dirty Region 跟踪
4. 支持 Z-index 层级
5. 支持增量渲染

严格遵守 ARCHITECTURE_INVARIANTS.md：
- 不变量 3：Render 永远是幂等的
- 不变量 4：Component 不允许直接操作 Terminal
```

**验收标准**:
- [ ] CellBuffer 工作正常
- [ ] RenderTree 结构正确
- [ ] Dirty Region 跟踪工作
- [ ] Diff 引擎正确计算差异
- [ ] 单元测试覆盖率 > 80%

---

### ✅ 任务 0.9: 实现焦点系统

**状态**: `pending`
**优先级**: `P0`
**预估时间**: `2 小时`

**文档引用**:
- [ARCHITECTURE.md](./ARCHITECTURE.md) - 第 359-390 行 (Focus Manager)
- [FOCUS_SYSTEM.md](./FOCUS_SYSTEM.md) - 焦点系统详细设计

**AI 提示词**:
```
任务：实现焦点系统

根据 ARCHITECTURE.md 第 359-390 行和 FOCUS_SYSTEM.md，实现：

1. runtime/focus/path.go
   - FocusPath 类型（[]string）
   - String、Equals、Clone、Append、Parent 方法

2. runtime/focus/scope.go
   - ScopeType 类型
   - FocusScope 结构体（ID、Type、Focusables、parent）
   - PushScope、PopScope

3. runtime/focus/manager.go
   - Manager 结构体（path、scopes、registry）
   - Current、CurrentID 方法
   - SetFocus、GetFocus 方法
   - Navigate 方法（上下左右、下一个、上一个）
   - PushScope、PopScope 方法

4. runtime/focus/modal.go
   - ModalManager 结构体
   - PushModal、PopModal 方法

焦点规则：
- FocusPath 表示焦点路径 ["root", "main", "form", "username"]
- FocusScope 用于 Modal、Dialog 等场景
- Navigate 支持方向导航和 Tab 顺序
- Modal 锁定焦点在其作用域内

注意事项：
1. 使用 FocusPath 而不是 bool
2. 支持 FocusScope 实现 Modal
3. 支持键盘导航
4. 焦点变化触发 Action

测试要求：
- 测试 FocusPath 操作
- 测试 Scope 压栈/出栈
- 测试导航功能
- 测试 Modal 焦点锁定
```

**验收标准**:
- [ ] FocusPath 工作正常
- [ ] Scope 管理正确
- [ ] 导航功能工作
- [ ] Modal 焦点锁定正常
- [ ] 单元测试覆盖率 > 80%

---

### ✅ 任务 0.10: 实现 Runtime 核心

**状态**: `pending`
**优先级**: `P0`
**预估时间**: `2 小时`

**文档引用**:
- [ARCHITECTURE.md](./ARCHITECTURE.md) - 第 419-456 行 (Runtime 核心)

**AI 提示词**:
```
任务：实现 Runtime 核心

根据 ARCHITECTURE.md 第 419-456 行，实现 Runtime 核心：

1. runtime/runtime.go
   - Runtime 结构体（layout、paint、focus、dirty、animation、input、action、state）
   - NewRuntime 创建 Runtime
   - Dispatch 唯一入口点

2. Dispatch 流程：
   1. Focus 处理
   2. 状态更新
   3. 标记脏区域
   4. 动画处理
   5. 布局计算（仅当 dirty）
   6. 绘制（仅 dirty 区域）

注意事项：
1. Dispatch 是唯一的状态变化入口
2. 严格遵守单向数据流
3. 只在 dirty 时重新布局和绘制
4. 支持动画按需 tick

严格遵守 ARCHITECTURE_INVARIANTS.md：
- 不变量 2：所有 UI 行为必须能被 replay
```

**验收标准**:
- [ ] Runtime 核心组装完成
- [ ] Dispatch 流程正确
- [ ] 所有子系统协同工作
- [ ] 集成测试通过

---

## Phase 1: Core Framework (核心框架)

### ✅ 任务 1.1: 实现组件基础类

**状态**: `pending`
**优先级**: `P0`
**预估时间**: `2 小时`

**文档引用**:
- [COMPONENTS.md](./COMPONENTS.md) - Component V3 设计
- [ARCHITECTURE.md](./ARCHITECTURE.md) - 第 428-484 行

**AI 提示词**:
```
任务：实现组件基础类

根据 COMPONENTS.md 和 ARCHITECTURE.md，实现：

1. component/base.go
   - BaseComponent 结构体
   - 实现 Node、Mountable、Measurable、Paintable

2. component/state.go
   - StateHolder 结构体
   - GetState、SetState、GetStateValue、SetStateValue
   - GetProps、SetProps、GetProp、SetProp
   - ExportState、ImportState

3. component/factory.go
   - Factory 结构体
   - CreateFromSpec 从 Spec 创建组件
   - Register 注册组件类型

注意事项：
1. BaseComponent 提供通用实现
2. StateHolder 管理组件状态
3. Factory 支持 DSL/Spec
4. 严格遵守 Capability Interfaces

严格遵守 ARCHITECTURE_INVARIANTS.md：
- 不变量 7：DSL/Spec 是一等公民
```

**验收标准**:
- [ ] BaseComponent 工作正常
- [ ] StateHolder 管理状态
- [ ] Factory 能创建组件
- [ ] 单元测试通过

---

### ✅ 任务 1.2: 实现屏幕管理

**状态**: `pending`
**优先级**: `P0`
**预估时间**: `2 小时`

**文档引用**:
- [ARCHITECTURE.md](./ARCHITECTURE.md) - 第 518-546 行

**AI 提示词**:
```
任务：实现屏幕管理

根据 ARCHITECTURE.md 第 518-546 行，实现：

1. screen/manager.go
   - Manager 结构体（screen、runtime、frontBuf、backBuf）
   - Init、Close 方法
   - Render 渲染缓冲区到屏幕
   - Diff 计算差异并输出

2. screen/painter.go
   - Painter 结构体（Framework → Runtime 适配器）
   - Paint 方法

注意事项：
1. 使用双缓冲（front/back）
2. 只输出差异
3. 适配 Runtime 到 Framework
```

**验收标准**:
- [ ] 屏幕管理器工作
- [ ] 双缓冲正常
- [ ] Diff 渲染工作
- [ ] 集成测试通过

---

## Phase 2: Basic Components (基础组件)

### ✅ 任务 2.1: 实现 Text 组件

**状态**: `pending`
**优先级**: `P1`
**预估时间**: `1 小时`

**文档引用**:
- [IMPLEMENTATION.md](./IMPLEMENTATION.md) - 第 431-470 行

**AI 提示词**:
```
任务：实现 Text 显示组件

创建 display/text.go：

```go
type Text struct {
    BaseComponent
    content string
    align   TextAlign
    wrap    bool
}

// 实现 Paintable
func (t *Text) Paint(ctx PaintContext, buf *CellBuffer)

// 支持：
// - 多行文本
// - 文本对齐（左、中、右）
// - 自动换行
// - 样式应用
```

测试要求：
- 测试文本显示
- 测试对齐
- 测试换行
- 测试样式
```

**验收标准**:
- [ ] Text 正确显示
- [ ] 支持对齐
- [ ] 支持换行
- [ ] 单元测试通过

---

### ✅ 任务 2.2: 实现 Box 容器

**状态**: `pending`
**优先级**: `P1`
**预估时间**: `1.5 小时`

**文档引用**:
- [IMPLEMENTATION.md](./IMPLEMENTATION.md) - 第 471-503 行

**AI 提示词**:
```
任务：实现 Box 容器组件

创建 layout/box.go：

```go
type Box struct {
    BaseContainer
    border  Border
    padding BoxSpacing
    margin  BoxSpacing
}

// 支持：
// - 边框样式
// - 内边距
// - 外边距
// - 嵌套子组件
```

测试要求：
- 测试边框渲染
- 测试空间计算
- 测试嵌套
```

**验收标准**:
- [ ] Box 正确渲染
- [ ] 边框样式正确
- [ ] 空间计算正确
- [ ] 支持嵌套

---

### ✅ 任务 2.3: 实现 TextInput 组件

**状态**: `pending`
**优先级**: `P1`
**预估时间**: `3 小时`

**文档引用**:
- [ACTION_SYSTEM.md](./ACTION_SYSTEM.md) - 第 988-1058 行
- [STATE_MANAGEMENT.md](./STATE_MANAGEMENT.md) - 第 733-802 行

**AI 提示词**:
```
任务：实现 TextInput 输入组件

创建 input/textinput.go：

```go
type TextInput struct {
    BaseComponent
    value       string
    cursor      int
    placeholder string
    password    bool
    selection   Selection
}

// 实现 ActionTarget
func (t *TextInput) HandleAction(a *Action) bool

// 支持的 Action：
// - ActionInputText: 输入文本
// - ActionDeleteChar: 删除字符
// - ActionDeleteWord: 删除单词
// - ActionCursorHome/End: 光标移动
// - ActionSelectAll: 全选
// - ActionSubmit: 提交
```

注意事项：
1. 只处理 Action，不处理 KeyEvent
2. 状态通过 StateHolder 管理
3. 支持选择、光标、占位符
4. 支持密码模式

测试要求：
- 测试字符输入
- 测试光标移动
- 测试删除
- 测试密码模式
- 测试选择
```

**验收标准**:
- [ ] 输入功能正常
- [ ] 光标操作正常
- [ ] 删除功能正常
- [ ] 密码模式工作
- [ ] 单元测试覆盖率 > 80%

---

### ✅ 任务 2.4: 实现 Button 组件

**状态**: `pending`
**优先级**: `P1`
**预估时间**: `1 小时`

**文档引用**:
- [IMPLEMENTATION.md](./IMPLEMENTATION.md) - 第 580-620 行

**AI 提示词**:
```
任务：实现 Button 交互组件

创建 interactive/button.go：

```go
type Button struct {
    InteractiveComponent
    label   string
    onClick func()
}

// 实现 ActionTarget
func (b *Button) HandleAction(a *Action) bool
// ActionSubmit → onClick()
```

测试要求：
- 测试点击
- 测试快捷键
- 测试样式变化
```

**验收标准**:
- [ ] 点击功能正常
- [ ] 样式变化正确
- [ ] 单元测试通过

---

## Phase 3: Advanced Components (高级组件)

### ✅ 任务 3.1: 实现 List 组件

**状态**: `pending`
**优先级**: `P2`
**预估时间**: `3 小时`

**文档引用**:
- [VIRTUAL_SCROLL.md](./VIRTUAL_SCROLL.md) - 虚拟滚动设计
- [ACTION_SYSTEM.md](./ACTION_SYSTEM.md) - 第 1062-1104 行

**AI 提示词**:
```
任务：实现 List 组件（支持虚拟滚动）

创建 display/list.go：

```go
type List struct {
    InteractiveComponent
    items      []interface{}
    cursor     int
    offset     int
    selected   map[int]bool
    height     int  // 可见高度
}

// 支持的 Action：
// - ActionNavigateUp/Down
// - ActionNavigatePageUp/Down
// - ActionNavigateFirst/Last
// - ActionSelectItem
// - ActionSubmit
```

注意事项：
1. 支持虚拟滚动（只渲染可见项）
2. 支持键盘导航
3. 支持多选
4. 支持大数据量

测试要求：
- 测试导航
- 测试虚拟滚动
- 测试多选
- 测试大数据量性能
```

**验收标准**:
- [ ] 列表显示正常
- [ ] 导航流畅
- [ ] 虚拟滚动工作
- [ ] 性能测试通过
- [ ] 单元测试覆盖率 > 80%

---

### ✅ 任务 3.2: 实现 Table 组件

**状态**: `pending`
**优先级**: `P2`
**预估时间**: `4 小时`

**文档引用**:
- [TABLE_SUBFOCUS.md](./TABLE_SUBFOCUS.md) - Table 子焦点设计
- [VIRTUAL_SCROLL.md](./VIRTUAL_SCROLL.md)

**AI 提示词**:
```
任务：实现 Table 组件（支持虚拟滚动、排序、子焦点）

创建 display/table.go：

```go
type Table struct {
    InteractiveComponent
    columns   []Column
    rows      [][]string
    cursor    int
    offset    int
    sortColumn int
    sortAsc    bool
    subFocus  *TableCellFocus  // 单元格级焦点
}

type Column struct {
    Title    string
    Width    int
    Align    TextAlign
    Sortable bool
}

type TableCellFocus struct {
    Row int
    Col int
}

// 支持的 Action：
// - 导航 Action
// - ActionSort: 排序
// - ActionTableCellEdit: 单元格编辑
// - ActionTableSelectRow: 选择行
```

注意事项：
1. 支持虚拟滚动
2. 支持列排序
3. 支持子焦点（单元格级）
4. 支持行选择
5. 支持固定列

测试要求：
- 测试显示
- 测试排序
- 测试子焦点
- 测试虚拟滚动
- 测试大数据量性能
```

**验收标准**:
- [ ] 表格显示正常
- [ ] 排序功能正常
- [ ] 子焦点工作
- [ ] 虚拟滚动工作
- [ ] 性能测试通过

---

### ✅ 任务 3.3: 实现 Form 组件

**状态**: `pending`
**优先级**: `P2`
**预估时间**: `3 小时`

**文档引用**:
- [FORM_VALIDATION.md](./FORM_VALIDATION.md) - 表单验证设计

**AI 提示词**:
```
任务：实现 Form 表单组件

创建 form/form.go：

```go
type Form struct {
    ContainerComponent
    fields      []Field
    currentField int
    onSubmit    func(data map[string]interface{})
    onCancel    func()
}

type Field interface {
    Component
    Name() string
    Value() interface{}
    Validate() error
    SetFocus() bool
}

// 支持的 Action：
// - ActionNavigateNext/Prev: 字段导航
// - ActionSubmit: 提交表单
// - ActionCancel: 取消
// - ActionValidate: 验证
```

注意事项：
1. 支持字段导航（Tab/Shift+Tab）
2. 支持实时验证
3. 支持错误显示
4. 支持异步验证
5. 支持 Field 接口扩展

测试要求：
- 测试导航
- 测试验证
- 测试提交
- 测试错误处理
```

**验收标准**:
- [ ] 表单导航正常
- [ ] 验证功能正常
- [ ] 提交功能正常
- [ ] 错误显示正确
- [ ] 单元测试覆盖率 > 80%

---

## Phase 4: AI 集成

### ✅ 任务 4.1: 实现 AI Controller

**状态**: `pending`
**优先级**: `P2`
**预估时间**: `3 小时`

**文档引用**:
- [AI_INTEGRATION.md](./AI_INTEGRATION.md) - AI 集成标准

**AI 提示词**:
```
任务：实现 AI Controller

根据 AI_INTEGRATION.md，实现：

1. runtime/ai/controller.go
   - Controller 接口定义
   - RuntimeController 实现

2. runtime/ai/operations.go
   - Operation 接口
   - ClickOperation、InputOperation、WaitOperation

3. 添加 JSON API：
   - GET /ai/inspect
   - POST /ai/dispatch
   - POST /ai/find
   - POST /ai/query

Controller 能力：
- Inspect(): 获取完整 UI 状态
- Find(selector): 查找组件
- Query(query): 查询状态
- Dispatch(action): 分发 Action
- Click(id): 点击组件
- Input(id, text): 输入文本
- Navigate(direction): 导航
- Execute(ops...): 执行操作序列
- WaitUntil(condition, timeout): 等待状态
- Watch(callback): 监控状态

注意事项：
1. AI 与人类用户平级
2. 使用语义化 Action，不是模拟按键
3. 支持组件查询（类似 DOM Selector）
4. 支持状态监控和等待

测试要求：
- 测试 Inspect
- 测试 Find
- 测试 Query
- 测试操作序列
- 测试状态监控
```

**验收标准**:
- [ ] Controller 工作正常
- [ ] 支持 DOM 风格查询
- [ ] 支持语义操作
- [ ] 支持 Watch/Wait
- [ ] JSON API 工作正常

---

### ✅ 任务 4.2: 实现自动化测试框架

**状态**: `pending`
**优先级**: `P2`
**预估时间**: `3 小时`

**文档引用**:
- [AI_INTEGRATION.md](./AI_INTEGRATION.md) - 第 634-691 行

**AI 提示词**:
```
任务：实现自动化测试框架

创建 framework/testing/ 目录：

1. testing/suite.go
   - TestSuite 结构体
   - Run() 运行测试套件

2. testing/recorder.go
   - Recorder 结构体
   - RecordActions() 记录操作
   - Replay() 回放操作

3. testing/assert.go
   - UIAssert 工具
   - ComponentExists、ComponentVisible、ComponentDisabled 等

4. testing/fixtures.go
   - Fixtures 工具
   - LoadApp() 加载测试应用

测试模式：
- 单元测试：测试单个组件
- 集成测试：测试组件交互
- E2E 测试：完整场景测试
- Replay 测试：回放操作验证
```

**验收标准**:
- [ ] 测试框架工作
- [ ] 支持 Replay
- [ ] 断言工具完整
- [ ] 示例测试通过

---

## Phase 5: Polish & Testing (优化与测试)

### ✅ 任务 5.1: 性能优化

**状态**: `pending`
**优先级**: `P1`
**预估时间**: `4 小时`

**AI 提示词**:
```
任务：性能优化

优化项：
1. Dirty Region 跟踪优化
2. 布局缓存
3. 渲染批处理
4. 事件防抖
5. 内存优化

性能目标：
- 1000 组件渲染 < 50ms
- 10000 行 Table 滚动流畅
- 内存占用 < 50MB
- CPU 空闲 < 1%

测试工具：
- Benchmark 测试
- pprof 性能分析
```

**验收标准**:
- [ ] 性能基准达标
- [ ] Benchmark 通过
- [ ] 内存泄漏检测通过

---

### ✅ 任务 5.2: 文档完善

**状态**: `pending`
**优先级**: `P1`
**预估时间**: `3 小时`

**AI 提示词**:
```
任务：文档完善

需要完善的文档：
1. API 文档：godoc 注释
2. 示例代码：examples/ 目录
3. 教程：TUTORIAL.md
4. 常见问题：FAQ.md
5. 迁移指南：MIGRATION.md
```

**验收标准**:
- [ ] godoc 完整
- [ ] 至少 5 个示例
- [ ] 教程完整
- [ ] FAQ 覆盖常见问题

---

## 通用规则（所有任务必须遵守）

### 代码规范

```
1. 文件命名：snake_case.go
2. 包命名：小写单词
3. 接口命名： capability + "able" 后缀
4. 常量命名：PascalCase 或 UPPER_SNAKE_CASE
5. 错误处理：显式处理，不忽略
6. 并发安全：使用具名互斥锁
```

### 提交规范

```
<type>(<scope>): <subject>

类型（type）:
- feat: 新功能
- fix: 修复 Bug
- docs: 文档更新
- refactor: 重构
- test: 测试相关
- chore: 构建/工具

示例：
feat(component): add Focusable interface

- Add FocusID() method
- Add OnFocus() and OnBlur() hooks
- Update documentation
```

### 测试要求

```
1. 单元测试：覆盖率 > 80%
2. 表驱动测试：使用 t.Run
3. Mock 接口：测试隔离
4. 并发测试：测试线程安全
5. Benchmark：性能关键代码
```

---

## 任务状态跟踪

```
总任务数：30+
已完成：0
进行中：0
待开始：30+

进度：
Phase 0: █████████░░░░░░░░░░░░░ 20% (0/10)
Phase 1: ░░░░░░░░░░░░░░░░░░░░░ 0% (0/2)
Phase 2: ░░░░░░░░░░░░░░░░░░░░░ 0% (0/4)
Phase 3: ░░░░░░░░░░░░░░░░░░░░░ 0% (0/3)
Phase 4: ░░░░░░░░░░░░░░░░░░░░░ 0% (0/2)
Phase 5: ░░░░░░░░░░░░░░░░░░░░░ 0% (0/2)
```

---

## 快速启动

### 开发环境设置

```bash
# 设置环境变量
export YAO_ROOT="E:/projects/yao/wwsheng009/yao"
export YAO_TEST_APPLICATION="E:/projects/yao/wwsheng009/yao/tui/demo"
export YAO_ENV="development"
export YAO_LOG_CONSOLE="true"
export YAO_LOG_LEVEL="TRACE"

# 拉取最新代码
git pull

# 创建功能分支
git checkout -b feat/tui-framework-v3
```

### 开始开发

```bash
# 选择一个任务
# 例如：任务 0.1 创建目录结构

# 执行任务
mkdir -p tui/framework/component
mkdir -p tui/framework/runtime/layout
# ... 其他目录

# 运行测试
go test ./tui/framework/... -v

# 提交代码
git add .
git commit -m "feat(tui): create directory structure"
```

---

## 相关文档

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 架构总览
- [ARCHITECTURE_INVARIANTS.md](./ARCHITECTURE_INVARIANTS.md) - 架构不变量
- [BOUNDARIES.md](./BOUNDARIES.md) - 层级边界
- [ACTION_SYSTEM.md](./ACTION_SYSTEM.md) - Action 系统
- [STATE_MANAGEMENT.md](./STATE_MANAGEMENT.md) - 状态管理
- [AI_INTEGRATION.md](./AI_INTEGRATION.md) - AI 集成
- [AI_DEV_GUIDE.md](./AI_DEV_GUIDE.md) - AI 开发指南
