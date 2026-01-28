# Binding 模块使用指南

## 目录

1. [基础用法](#基础用法)
2. [数据绑定](#数据绑定)
3. [作用域继承](#作用域继承)
4. [响应式状态](#响应式状态)
5. [表达式](#表达式)
6. [列表渲染](#列表渲染)
7. [组件集成](#组件集成)

## 基础用法

### 创建静态属性

```go
import "github.com/yaoapp/yao/tui/framework/binding"

// 字符串属性
title := binding.NewStatic("Hello World")

// 整数属性
count := binding.NewStatic(42)

// 布尔属性
enabled := binding.NewStatic(true)
```

### 解析属性值

```go
ctx := binding.NewRootScope(nil)
value := title.Resolve(ctx)
```

## 数据绑定

### 创建绑定属性

```go
// 方式一：显式创建
nameProp := binding.NewBinding[string]("user.name")

// 方式二：自动检测（推荐）
titleProp := binding.NewStringProp("{{ page.title }}")
```

### 设置数据上下文

```go
ctx := binding.NewRootScope(map[string]interface{}{
    "user": map[string]interface{}{
        "name": "Alice",
        "age": 30,
    },
    "page": map[string]interface{}{
        "title": "Dashboard",
    },
})
```

### 解析绑定值

```go
name := nameProp.Resolve(ctx)  // "Alice"
title := titleProp.Resolve(ctx) // "Dashboard"
```

## 作用域继承

### 创建子作用域

```go
// 根作用域
root := binding.NewRootScope(map[string]interface{}{
    "app": map[string]interface{}{
        "name": "MyApp",
        "version": "1.0",
    },
})

// 子作用域继承父作用域
userScope := root.New(map[string]interface{}{
    "user": map[string]interface{}{
        "name": "Alice",
    },
})

// 子作用域可访问父作用域数据
appName := userScope.Get("app.name")  // "MyApp"
userName := userScope.Get("user.name") // "Alice"
```

### 特殊变量

```go
// 在列表渲染中使用
ctx := binding.NewRootScope(nil)
ctx.Set("$index", 0)
ctx.Set("$item", map[string]interface{}{"name": "Item 1"})

index := ctx.Get("$index") // 0
item := ctx.Get("$item")   // map[name:Item 1]
```

## 响应式状态

### 创建响应式存储

```go
store := binding.NewReactiveStore()
store.Set("user.name", "Alice")
store.Set("user.age", 30)
```

### 订阅变更

```go
cancel := store.Subscribe("user.name", func(key string, old, new interface{}) {
    fmt.Printf("%s: %v → %v\n", key, old, new)
})

// 更新会触发回调
store.Set("user.name", "Bob")  // 输出: user.name: Alice → Bob

// 取消订阅
cancel()
```

### 批量更新

```go
store.BeginBatch()
store.Set("key1", "value1")
store.Set("key2", "value2")
store.Set("key3", "value3")
store.EndBatch()  // 只触发一次通知
```

### 计算属性

```go
computed := binding.NewStoreComputed(store,
    []string{"price", "quantity"},
    func() interface{} {
        price, _ := store.Get("price")
        quantity, _ := store.Get("quantity")
        return price.(int) * quantity.(int)
    },
)

total := computed.Value()  // 自动计算
computed.Dispose()          // 清理
```

## 表达式

### 基本表达式

```go
// 数学运算
expr1 := binding.ParseExpression("price * quantity")
expr2 := binding.ParseExpression("(a + b) / 2")

// 变量引用
expr3 := binding.ParseExpression("user.firstName + ' ' + user.lastName")
```

### 求值表达式

```go
ctx := binding.NewRootScope(map[string]interface{}{
    "price": 10,
    "quantity": 5,
})

expr := binding.ParseExpression("price * quantity")
result := expr.Evaluate(ctx)  // 50.0
```

### 获取依赖

```go
expr := binding.ParseExpression("price * quantity + tax")
deps := expr.Dependencies()  // ["price", "quantity", "tax"]
```

## 列表渲染

### 创建列表上下文

```go
items := []interface{}{
    map[string]interface{}{"id": 1, "name": "Item 1"},
    map[string]interface{}{"id": 2, "name": "Item 2"},
    map[string]interface{}{"id": 3, "name": "Item 3"},
}

parent := binding.NewRootScope(map[string]interface{}{
    "status": "active",
})

contexts := binding.ListContext(parent, items)

// 每个上下文包含：
// - $index: 当前索引
// - $item: 当前项数据
// - item 的扁平化属性（id, name）
// - 父作用域的所有数据（status）
```

### 遍历渲染

```go
for i, ctx := range contexts {
    index, _ := ctx.Get("$index")  // 0, 1, 2
    name, _ := ctx.Get("name")     // "Item 1", "Item 2", "Item 3"
    status, _ := ctx.Get("status") // "active" (继承)

    // 使用数据渲染行
    renderRow(i, ctx)
}
```

## 组件集成

### 创建可绑定组件

```go
import (
    cb "github.com/yaoapp/yao/tui/framework/component/binding"
)

type MyComponent struct {
    *cb.BaseBindable
}

func NewMyComponent() *MyComponent {
    return &MyComponent{
        BaseBindable: cb.NewBaseBindable("mycomponent"),
    }
}

func (c *MyComponent) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    // 创建绑定上下文
    bindCtx := cb.CreateBindingContext(c)

    // 解析所有绑定
    props := c.ResolveBindings(bindCtx)

    // 使用解析后的属性渲染
    title := props["title"].(string)
    // ...
}
```

### 应用 DSL 属性

```go
// 解析 DSL 配置
dslProps := map[string]interface{}{
    "title": "{{ page.title }}",
    "count": "{{ items.length }}",
    "visible": true,
}

parsed := cb.ParseDSLProps(dslProps)

// 应用到组件
cb.ApplyProps(component, parsed)
```

### Store 与组件集成

```go
store := binding.NewReactiveStore()
store.Set("username", "alice")

// 创建适配器
adapter := cb.NewStoreToComponentAdapter(store, component, map[string]string{
    "value": "username",  // store key → component prop
})

// 同步数据
adapter.Sync()

// 监听变化
cancel := adapter.Watch(func() {
    component.MarkDirty()
})
```

### 双向绑定

```go
// Store ←→ Component 双向同步
binding := cb.NewTwoWayBinding(store, component, "value", "username")

// Store 变化自动更新组件
// 组件变化可手动更新 Store:
//   store.Set("username", newValue)
```

## 完整示例

```go
package main

import (
    "github.com/yaoapp/yao/tui/framework/binding"
    cb "github.com/yaoapp/yao/tui/framework/component/binding"
)

func main() {
    // 1. 创建响应式存储
    store := binding.NewReactiveStore()
    store.Set("user", map[string]interface{}{
        "name": "Alice",
        "email": "alice@example.com",
    })
    store.Set("items", []interface{}{
        map[string]interface{}{"id": 1, "title": "Task 1", "done": true},
        map[string]interface{}{"id": 2, "title": "Task 2", "done": false},
    })

    // 2. 创建根作用域
    rootCtx := store.ToContext()

    // 3. 创建组件
    comp := NewMyComponent()
    comp.SetBinding("title", binding.NewBinding[string]("user.name"))
    comp.SetBinding("count", binding.NewBinding[string]("items.length"))

    // 4. 解析并渲染
    props := comp.ResolveBindings(rootCtx)
    title := props["title"].(string)  // "Alice"

    // 5. 监听变化
    store.Subscribe("user.name", func(key, old, new interface{}) {
        comp.MarkDirty()  // 触发重绘
    })

    // 6. 列表渲染
    items, _ := store.Get("items")
    listCtxs := binding.ListContext(rootCtx, items.([]interface{}))
    for _, ctx := range listCtxs {
        renderListItem(ctx)
    }
}
```

## 最佳实践

### 1. 优先使用静态属性

```go
// 好
title := binding.NewStatic("Dashboard")

// 避免（除非需要动态绑定）
title := binding.NewStringProp("{{ 'Dashboard' }}")
```

### 2. 合理使用作用域

```go
// 为列表创建临时作用域，而非污染全局
for _, item := range items {
    itemCtx := parent.New(map[string]interface{}{
        "$item": item,
    })
    process(itemCtx)
}
```

### 3. 及时取消订阅

```go
// 总是保存取消函数
cancel := store.Subscribe("key", handler)
defer cancel()  // 确保清理
```

### 4. 使用批量更新

```go
// 多个更新时使用批次
store.BeginBatch()
store.Set("key1", val1)
store.Set("key2", val2)
store.Set("key3", val3)
store.EndBatch()  // 一次性通知
```
