# Binding 模块

Yao TUI 的数据绑定和响应式状态管理模块。

## 概述

Binding 模块提供了声明式数据绑定能力，使 TUI 组件能够响应数据变化而自动更新。它采用**编译时绑定**模式，在 DSL 解析阶段创建属性绑定，在渲染阶段解析当前值。

## 核心特性

- **编译时绑定** - DSL 解析时创建 Prop[T]，避免运行时字符串解析
- **作用域链** - 支持数据继承和嵌套访问
- **响应式存储** - 细粒度依赖追踪和变更通知
- **表达式求值** - 支持简单的数学和逻辑表达式
- **零外部依赖** - 核心模块只使用 Go 标准库

## 模块结构

```
tui/framework/
├── binding/                    # 核心数据绑定（零依赖）
│   ├── context.go              # 数据上下文接口
│   ├── prop.go                 # 泛型属性
│   ├── scope.go                # 作用域链
│   ├── expression.go           # 表达式解析器
│   ├── store.go                # 响应式存储
│   └── binding_test.go         # 核心测试
│
└── component/
    └── binding/                # 组件集成层
        ├── integration.go       # 组件集成
        └── integration_test.go  # 集成测试
```

## 快速开始

### 1. 静态属性

```go
import "github.com/yaoapp/yao/tui/framework/binding"

// 创建静态属性
titleProp := binding.NewStatic("Hello World")
countProp := binding.NewStatic(42)

// 解析属性
ctx := binding.NewRootScope(nil)
title := titleProp.Resolve(ctx)  // "Hello World"
```

### 2. 数据绑定

```go
// 创建数据绑定
nameProp := binding.NewBinding[string]("user.name")

// 设置数据上下文
ctx := binding.NewRootScope(map[string]interface{}{
    "user": map[string]interface{}{
        "name": "Alice",
    },
})

// 解析属性
name := nameProp.Resolve(ctx)  // "Alice"
```

### 3. 自动检测

```go
// 自动检测 {{ }} 语法
prop := binding.NewStringProp("{{ user.name }}")

if prop.IsBound() {
    // 这是数据绑定
    name := prop.Resolve(ctx)
}
```

### 4. 响应式存储

```go
store := binding.NewReactiveStore()

// 订阅变化
cancel := store.Subscribe("user.name", func(key string, old, new interface{}) {
    fmt.Printf("%s changed from %v to %v\n", key, old, new)
})

// 更新数据（触发通知）
store.Set("user.name", "Bob")

// 取消订阅
cancel()
```

## 设计原则

1. **零依赖** - 核心绑定模块不依赖任何框架代码
2. **类型安全** - 使用 Go 泛型确保类型安全
3. **高性能** - 编译时创建绑定，运行时仅解析
4. **可组合** - 所有组件都是可组合的

## 下一步

- [架构设计](architecture.md) - 详细架构说明
- [使用指南](guide.md) - 完整使用教程
- [API 参考](api.md) - 完整 API 文档
