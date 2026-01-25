# Form

表单组件和验证系统。

## 职责

- 表单容器组件
- 表单字段组件
- 表单验证和提交

## 组件列表

- `Form` - 表单容器
- `Input` - 输入框
- `TextArea` - 多行输入
- `Checkbox` - 复选框
- `Radio` - 单选框
- `Select` - 下拉选择

## 相关文件

- `form.go` - 表单容器
- `input.go` - 输入框
- `textarea.go` - 多行输入
- `checkbox.go` - 复选框

---

## 死锁问题与解决方案

### 问题描述

在实现表单提交功能时，遇到了死锁问题。当用户按 Enter 键提交表单时，程序完全无响应。

### 根本原因

**Go 的 sync.RWMutex 不是可重入的**

调用链导致死锁：
```
HandleAction() [持有 f.mu.Lock()]
    ↓
Submit() [尝试获取 f.mu.Lock() - 死锁!]
    ↓
Validate() [尝试获取 f.mu.Lock() - 死锁!]
    ↓
GetValues() [尝试获取 f.mu.RLock() - 死锁!]
```

即使在同一个 goroutine 中，持有写锁时再次尝试获取读锁也会导致死锁。

### 错误代码示例

```go
// ❌ 错误：会导致死锁
func (f *Form) HandleAction(a action.Action) bool {
    f.mu.Lock()
    defer f.mu.Unlock()

    switch a.Type {
    case action.ActionSubmit:
        if err := f.Submit(); err != nil {  // 死锁！
            return true
        }
        return f.onSubmit != nil
    }
}

func (f *Form) Submit() error {
    f.mu.Lock()  // 死锁！HandleAction 已持有此锁
    defer f.mu.Unlock()
    // ...
}
```

### 解决方案

**策略：细粒度锁 + 内部不加锁方法**

#### 1. 公共方法不加锁

```go
// ✅ 正确：Validate() 不加锁，调用者负责线程安全
func (f *Form) Validate() error {
    // 不加锁，因为可能从已持有锁的方法调用
    // 调用者需要确保线程安全

    for name, field := range f.fields {
        if field.Visible && !field.Disabled {
            if err := field.Validate(); err != nil {
                return err
            }
        }
    }
    return nil
}
```

#### 2. HandleAction 细粒度锁

```go
// ✅ 正确：只在必要时加锁
func (f *Form) HandleAction(a action.Action) bool {
    switch a.Type {
    case action.ActionSubmit:
        // 不持有锁，避免死锁
        return f.handleSubmit()

    case action.ActionNavigateDown:
        f.mu.Lock()
        f.navigateField(1)
        f.mu.Unlock()
        return true

    case action.ActionCancel:
        f.mu.Lock()
        onCancel := f.onCancel
        f.mu.Unlock()
        if onCancel != nil {
            onCancel()
        }
        return true
    }
}
```

#### 3. 内部提交方法

```go
// ✅ 正确：内部方法使用细粒度锁
func (f *Form) handleSubmit() bool {
    // 验证 - 不需要锁
    if err := f.Validate(); err != nil {
        return true
    }

    // 标记字段 - 需要锁
    f.mu.Lock()
    for _, field := range f.fields {
        field.Touched = true
    }
    onSubmit := f.onSubmit
    f.mu.Unlock()

    // 调用回调 - 不需要锁
    if onSubmit != nil {
        values := f.GetValues()  // 只在读时获取锁
        if err := onSubmit(values); err != nil {
            return true
        }
    }

    // 更新状态 - 需要锁
    f.mu.Lock()
    f.submitted = true
    f.mu.Unlock()

    return true
}
```

### 锁的使用原则

| 场景 | 是否需要锁 | 说明 |
|------|-----------|------|
| 读取 `f.fields` map | 需要 | map 并发读写需要保护 |
| 读取 `f.onSubmit` | 不需要 | 函数指针读取是原子的 |
| 调用 `f.onSubmit()` | 不需要 | 用户回调不应持有表单锁 |
| 读取 `f.currentField` | 需要 | 多个 goroutine 可能访问 |
| 读取 field.Value | 不需要 | 字段本身有锁保护 |

### 时序图

```
用户按 Enter
    │
    ├─> Form.HandleEvent()
    │       │
    │       └─> Form.HandleAction(ActionSubmit)
    │               │
    │               └─> Form.handleSubmit() [无锁]
    │                       │
    │                       ├─> Form.Validate() [无锁]
    │                       │       │
    │                       │       └─> 遍历字段验证
    │                       │
    │                       ├─> Lock: 标记 Touched
    │                       ├─> Unlock
    │                       │
    │                       ├─> Form.GetValues() [RLock]
    │                       ├─> Unlock
    │                       │
    │                       ├─> onSubmit(values) [无锁]
    │                       │       │
    │                       │       └─> 用户回调
    │                       │
    │                       └─> Lock: submitted = true
    │                           └─> Unlock
```

### 调试技巧

**添加死锁检测日志：**

```go
func (f *Form) HandleAction(a action.Action) bool {
    fmt.Printf("[HandleAction] 开始, type=%d\n", a.Type)
    switch a.Type {
    case action.ActionSubmit:
        fmt.Printf("[HandleAction] 调用 handleSubmit\n")
        result := f.handleSubmit()
        fmt.Printf("[HandleAction] handleSubmit 返回 %v\n", result)
        return result
    }
}
```

**使用 Go 死锁检测器：**

```bash
# 启用死锁检测
export GODEBUG=deadlock=1
go run main.go
```

**使用 pprof 分析：**

```bash
# 获取 goroutine 堆栈
curl http://localhost:6060/debug/pprof/goroutine?debug=2
```

### 相关修改

| 文件 | 修改内容 |
|------|----------|
| `form.go` | 重构 `HandleAction()` 使用细粒度锁 |
| `form.go` | `Validate()` 移除锁 |
| `form.go` | 新增 `handleSubmit()` 内部方法 |
| `form.go` | `Submit()` 使用细粒度锁 |

### 验证步骤

1. 空表单按 Enter → 应显示验证错误
2. 输入有效数据后按 Enter → 应成功提交
3. 多次快速按 Enter → 不应死锁
4. 同时导航和提交 → 不应死锁
