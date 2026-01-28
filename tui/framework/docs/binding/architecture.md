# Binding 模块架构设计

## 总体架构

Binding 模块采用分层架构，核心层完全独立，集成层负责与组件系统对接。

```
┌─────────────────────────────────────────────────────────────────┐
│                        Application Layer                        │
│                    (DSL 解析 / 组件配置)                          │
└─────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Integration Layer                           │
│                   framework/component/binding                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ BindableComp │  │ PaintContext │  │ StoreAdapter │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Core Layer                                │
│                      framework/binding                          │
│  ┌──────────┐ ┌─────────┐ ┌──────────┐ ┌──────────┐           │
│  │   Prop   │ │  Scope  │ │Expression│ │  Store   │           │
│  └──────────┘ └─────────┘ └──────────┘ └──────────┘           │
│  ┌──────────┐ ┌─────────┐                                       │
│  │  Context │ │ Lexer   │                                       │
│  └──────────┘ └─────────┘                                       │
└─────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Go Standard Library                       │
└─────────────────────────────────────────────────────────────────┘
```

## 核心组件

### 1. Prop[T] - 泛型属性

**职责**：封装静态值和数据绑定

```go
type Prop[T any] struct {
    kind       PropKind
    staticVal  T
    bindPath   string
    expression *Expression
}
```

**状态转换**：

```
Parse DSL
    │
    ├─ 无 {{ }}  → NewStatic() → PropStatic
    │
    ├─ 有 {{ }} 且为简单路径 → NewBinding() → PropBound
    │
    └─ 有 {{ }} 且为表达式 → NewExpression() → PropExpression
```

### 2. Scope - 作用域链

**职责**：提供数据上下文和继承机制

```go
type Scope struct {
    data   map[string]interface{}
    parent *Scope
}
```

**查找流程**：

```
Get("user.profile.name")
    │
    ├─ 1. 检查当前 Scope.data
    │
    ├─ 2. 检查父 Scope (递归)
    │
    └─ 3. 返回值或 (nil, false)
```

**特殊变量**：
- `$index` - 列表项索引
- `$item` - 列表项数据
- `$parent` - 父作用域
- `$root` - 根作用域

### 3. Expression - 表达式解析

**职责**：解析和求值表达式

**处理流程**：

```
Expression String
        │
        ▼
┌───────────────┐
│  Lexer (词法) │ → tokens
└───────────────┘
        │
        ▼
┌───────────────┐
│  Parser (语法)│ → AST
└───────────────┘
        │
        ▼
┌───────────────┐
│  Evaluator    │ → result
└───────────────┘
```

**支持的运算**：
- 算术：`+`, `-`, `*`, `/`
- 比较：`==`, `!=`, `<`, `>`, `<=`, `>=`
- 逻辑：`&&`, `||`, `!`
- 成员：`.` (点号访问)

### 4. Store - 响应式存储

**职责**：管理全局状态并通知变更

```go
type ReactiveStore struct {
    mu        sync.RWMutex
    data      map[string]interface{}
    observers map[string][]Notifier
}
```

**变更流程**：

```
Set(key, value)
    │
    ├─ 1. 检查值是否变化
    │
    ├─ 2. 更新 data[key]
    │
    └─ 3. 通知 observers[key]
            │
            ▼
        Component.MarkDirty()
```

## 数据流

### 渲染时数据流

```
┌─────────┐     Parse      ┌─────────┐
│   DSL   │ ──────────────→ │  Prop   │
└─────────┘                 └─────────┘
                                 │
                                 ▼
┌─────────┐     Resolve    ┌─────────┐     Get     ┌─────────┐
│ Render  │ ─────────────→ │  Scope  │ ──────────→ │  Data   │
└─────────┘                 └─────────┘             └─────────┘
```

### 状态变更流

```
┌─────────┐     Update     ┌─────────┐    Notify    ┌─────────┐
│  Event  │ ─────────────→ │  Store  │ ──────────→ │Component │
└─────────┘                 └─────────┘             └─────────┘
                                                           │
                                                           ▼
                                                    ┌─────────┐
                                                    │ MarkDirty │
                                                    └─────────┘
```

## 集成层

### BindableComponent

组件实现数据绑定的接口：

```go
type BindableComponent interface {
    component.Node
    SetBinding(key string, prop interface{})
    GetBinding(key string) (interface{}, bool)
    BindAll(props map[string]interface{})
    ResolveBindings(ctx Context) map[string]interface{}
}
```

### PaintContext

扩展的绘制上下文，携带数据上下文：

```go
type PaintContext struct {
    component.PaintContext
    Data binding.Context
}
```

## 设计决策

### 为什么采用 Prop[T] 而非接口？

1. **类型安全** - 编译时检查类型
2. **零成本抽象** - 无运行时反射
3. **简单直接** - 代码清晰易理解

### 为什么独立于 runtime？

1. **可测试性** - 单元测试无需模拟复杂依赖
2. **可复用性** - 可用于其他 TUI 框架
3. **清晰分层** - 职责明确，便于维护

### 为什么支持表达式但不强制？

1. **渐进增强** - 简单场景用简单绑定
2. **性能考虑** - 表达式求值有开销
3. **调试友好** - 出错时更容易定位
