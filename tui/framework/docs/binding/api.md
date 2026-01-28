# Binding 模块 API 参考

## 目录

- [binding 包](#binding-包)
- [componentbinding 包](#componentbinding-包)

---

## binding 包

### Context 接口

数据上下文接口，提供数据访问能力。

```go
type Context interface {
    Get(path string) (interface{}, bool)
    Set(path string, value interface{}) bool
    Has(path string) bool
    Parent() Context
    Root() Context
    New(data map[string]interface{}) Context
}
```

#### 方法

| 方法 | 说明 |
|------|------|
| `Get(path)` | 获取路径对应的值，支持嵌套路径如 `user.profile.name` |
| `Set(path, value)` | 设置路径对应的值 |
| `Has(path)` | 检查路径是否存在 |
| `Parent()` | 返回父上下文 |
| `Root()` | 返回根上下文 |
| `New(data)` | 创建继承当前上下文的子上下文 |

---

### Scope 结构体

Context 的链表实现，支持作用域继承。

```go
type Scope struct {
    // 包含非导出字段
}
```

#### 构造函数

```go
func NewRootScope(data map[string]interface{}) *Scope
func NewScope(parent *Scope, data map[string]interface{}) *Scope
```

#### 方法

| 方法 | 说明 |
|------|------|
| `Get(path) (interface{}, bool)` | 获取值，支持嵌套和特殊变量 |
| `Set(path, value) bool` | 设置值 |
| `Has(path) bool` | 检查是否存在 |
| `Parent() Context` | 获取父作用域 |
| `Root() Context` | 获取根作用域 |
| `Data() map[string]interface{}` | 获取数据副本 |
| `Merge(data)` | 合并数据 |
| `Clear()` | 清空数据 |
| `Keys() []string` | 获取所有键 |
| `Len() int` | 获取数据量 |
| `IsRoot() bool` | 是否为根作用域 |
| `Clone() *Scope` | 克隆作用域 |

---

### Prop[T] 泛型属性

支持静态值和数据绑定的泛型属性。

```go
type Prop[T any] struct {
    // 包含非导出字段
}
```

#### 构造函数

```go
func NewStatic[T any](val T) Prop[T]
func NewBinding[T any](path string) Prop[T]
func NewExpression[T any](expr string) Prop[T]
func NewComputed[T any](deps []string, fn func(Context) T) Prop[T]
```

#### 便捷构造函数

```go
func NewStringProp(value string) StringProp
func Parse[T any](value string, convert func(string) T) Prop[T]
```

#### 方法

| 方法 | 说明 |
|------|------|
| `Resolve(ctx Context) T` | 解析属性值 |
| `ResolveString(ctx Context) string` | 解析为字符串 |
| `IsBound() bool` | 是否为数据绑定 |
| `IsExpression() bool` | 是否为表达式 |
| `IsStatic() bool` | 是否为静态值 |
| `GetPath() string` | 获取绑定路径 |
| `GetDependencies() []string` | 获取依赖项 |
| `SetStatic(val T)` | 设置为静态值 |

#### 类型别名

```go
type StringProp = Prop[string]
type IntProp = Prop[int]
type BoolProp = Prop[bool]
```

---

### PropBuilder 属性构建器

用于批量构建属性。

```go
type PropBuilder struct {
    // 包含非导出字段
}
```

#### 方法

```go
func NewPropBuilder() *PropBuilder
func (b *PropBuilder) String(key, value string) *PropBuilder
func (b *PropBuilder) Int(key string, value int) *PropBuilder
func (b *PropBuilder) Bool(key string, value bool) *PropBuilder
func (b *PropBuilder) Binding(key, path string) *PropBuilder
func (b *PropBuilder) Build() map[string]interface{}
```

#### 使用示例

```go
props := binding.NewPropBuilder().
    String("title", "Hello").
    Int("count", 42).
    Bool("enabled", true).
    Build()
```

---

### Expression 表达式

解析和求值表达式。

```go
type Expression struct {
    // 包含非导出字段
}
```

#### 构造函数

```go
func ParseExpression(expr string) *Expression
```

#### 方法

| 方法 | 说明 |
|------|------|
| `Evaluate(ctx Context) interface{}` | 求值表达式 |
| `Dependencies() []string` | 获取依赖的变量 |

#### 工具函数

```go
func EvaluateExpression(expr string, ctx Context) interface{}
func IsValidExpression(expr string) bool
func FormatExpression(expr string) string
```

---

### ReactiveStore 响应式存储

支持订阅通知的状态存储。

```go
type ReactiveStore struct {
    // 包含非导出字段
}
```

#### 构造函数

```go
func NewReactiveStore() *ReactiveStore
```

#### 数据操作

| 方法 | 说明 |
|------|------|
| `Get(path) (interface{}, bool)` | 获取值 |
| `Set(path, value)` | 设置值 |
| `Delete(path)` | 删除值 |
| `Has(path) bool` | 检查是否存在 |
| `GetAll() map[string]interface{}` | 获取所有数据 |
| `SetAll(data)` | 替换所有数据 |
| `Clear()` | 清空所有数据 |
| `Size() int` | 数据量 |
| `Keys() []string` | 所有键 |

#### 订阅通知

```go
type Notifier func(key string, oldValue, newValue interface{})

func (s *ReactiveStore) Subscribe(path string, notifier Notifier) func()
func (s *ReactiveStore) SubscribeGlobal(notifier Notifier) func()
func (s *ReactiveStore) Unsubscribe(path string, notifier Notifier)
func (s *ReactiveStore) UnsubscribeGlobal(notifier Notifier)
func (s *ReactiveStore) Enable()
func (s *ReactiveStore) Disable()
func (s *ReactiveStore) IsEnabled() bool
```

#### 批量操作

```go
func (s *ReactiveStore) BeginBatch()
func (s *ReactiveStore) EndBatch()
```

#### 上下文转换

```go
func (s *ReactiveStore) ToContext() Context
```

---

### StoreComputed 计算属性

基于依赖自动更新的计算值。

```go
type StoreComputed struct {
    // 包含非导出字段
}
```

#### 构造函数

```go
func NewStoreComputed(store *ReactiveStore, deps []string, compute func() interface{}) *StoreComputed
```

#### 方法

| 方法 | 说明 |
|------|------|
| `Value() interface{}` | 获取当前计算值 |
| `Dispose()` | 停止更新 |

---

### 工具函数

#### 作用域相关

```go
func ListContext(parent Context, items []interface{}) []Context
func WithIndex(parent Context, index int, item interface{}) Context
```

#### 表达式相关

```go
func EvaluateExpression(expr string, ctx Context) interface{}
func IsValidExpression(expr string) bool
```

---

## componentbinding 包

组件集成层，连接 binding 包与组件系统。

### PaintContext

扩展的绘制上下文。

```go
type PaintContext struct {
    component.PaintContext
    Data binding.Context
}
```

#### 方法

```go
func NewPaintContext(base component.PaintContext, data binding.Context) PaintContext
func (c PaintContext) WithData(data binding.Context) PaintContext
```

---

### BindableComponent 接口

支持数据绑定的组件接口。

```go
type BindableComponent interface {
    component.Node
    SetBinding(key string, prop interface{})
    GetBinding(key string) (interface{}, bool)
    BindAll(props map[string]interface{})
    ResolveBindings(ctx binding.Context) map[string]interface{}
}
```

---

### BaseBindable

BindableComponent 的基础实现。

```go
type BaseBindable struct {
    *component.BaseComponent
    *component.StateHolder
    // 包含非导出字段
}
```

#### 构造函数

```go
func NewBaseBindable(typ string) *BaseBindable
```

#### 方法

| 方法 | 说明 |
|------|------|
| `SetBinding(key, prop)` | 设置绑定属性 |
| `GetBinding(key) (interface{}, bool)` | 获取绑定属性 |
| `BindAll(props)` | 批量设置绑定 |
| `ResolveBindings(ctx)` | 解析所有绑定 |

---

### 组件辅助函数

```go
// 绘制辅助
func PaintWithBindings(comp BindableComponent, baseCtx component.PaintContext, buf *paint.Buffer, paintFunc func(PaintContext, *paint.Buffer))

// 上下文创建
func CreateBindingContext(comp component.Node) binding.Context

// 属性应用
func ApplyProps(comp component.Node, props map[string]interface{})
func ParseDSLProps(dslProps map[string]interface{}) map[string]interface{}

// 列表上下文
func CreateListContexts(parent binding.Context, items []interface{}, props map[string]interface{}) []binding.Context
```

---

### StoreAdapter

存储与组件的集成适配器。

```go
type StoreToComponentAdapter struct {
    // 包含非导出字段
}

func NewStoreToComponentAdapter(store *binding.ReactiveStore, comp component.Node, bindings map[string]string) *StoreToComponentAdapter
func (a *StoreToComponentAdapter) Sync()
func (a *StoreToComponentAdapter) Watch(callback func()) func()
```

---

### TwoWayBinding

双向绑定。

```go
type TwoWayBinding struct {
    // 包含非导出字段
}

func NewTwoWayBinding(store *binding.ReactiveStore, comp component.Node, prop, key string) *TwoWayBinding
func (b *TwoWayBinding) Dispose()
```

---

### ComponentWithProps

携带属性的组件包装。

```go
type ComponentWithProps struct {
    component component.Node
    props     map[string]interface{}
}

func NewComponentWithProps(comp component.Node, props map[string]interface{}) *ComponentWithProps
```
