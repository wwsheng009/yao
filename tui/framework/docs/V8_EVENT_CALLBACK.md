# V8 Event Callback Bridge Design

> **补充文档**: V8 事件回调桥接实现
> **核心问题**: Go 端收到事件 → 调用 JavaScript 处理器 → 返回结果

## v8go 能力分析

### ✅ v8go 支持的功能

```go
// 1. Go → JavaScript: 调用 JS 函数
val, err := jsFn.Call(global, arg1, arg2)

// 2. JavaScript → Go: JS 调用 Go 函数
template := v8go.NewFunctionTemplate(isolate, callback)
global.Set("goFunction", template.GetFunction())

// 3. 数据互转
// Go 值 → JS 值
jsValue, _ := v8go.NewValue(ctx).Create(value)

// JS 值 → Go 值
goValue := jsValue.String()
goValue := jsValue.AsNumber()
```

### ❌ v8go 不支持的功能

- **直接事件监听**: JavaScript 无法直接"监听"Go 事件
- **异步回调到 Go**: Promise 回调无法直接回到 Go（需要手动泵送）

## 解决方案：桥接模式

### 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                        Go Runtime                           │
│                                                              │
│  用户输入 (stdin)                                            │
│       │                                                      │
│       ▼                                                      │
│  InputProcessor ───► KeyEvent                                 │
│       │                                                      │
│       ▼                                                      │
│  KeyMap ──────────► Action                                   │
│       │                                                      │
│       ▼                                                      │
│  ScriptComponent.HandleAction(action)                        │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────────────┐    │
│  │         V8EventBridge                               │    │
│  │  - 将 Action 转换为 JS 对象                          │    │
│  │  - 调用 JS 中的 onAction 方法                         │    │
│  │  - 处理返回值                                       │    │
│  └─────────────────────────────────────────────────────┘    │
│       │                                                      │
│       ▼                                                      │
│  v8go.Isolate.Context                                       │
│       │                                                      │
│       ▼                                                      │
│  JavaScript: onAction(action) { ... }                       │
│       │                                                      │
│       ▼                                                      │
│  返回结果 (bool: 是否已处理)                                │
└─────────────────────────────────────────────────────────────┘
```

## 核心实现

### 1. V8EventBridge 事件桥接

```go
// 位于: tui/framework/v8/event_bridge.go

package v8

import (
    "sync"
    "rogchap.com/v8go"
)

// EventBridge V8 事件桥接器
type EventBridge struct {
    isolate *v8go.Isolate
    context *v8go.Context

    // JS 模板函数（用于 JS → Go 回调）
    templates map[string]*v8go.FunctionTemplate

    // 同步调用保护
    mu sync.Mutex
}

// NewEventBridge 创建事件桥接器
func NewEventBridge(isolate *v8go.Isolate, context *v8go.Context) *EventBridge {
    bridge := &EventBridge{
        isolate:   isolate,
        context:   context,
        templates: make(map[string]*v8go.FunctionTemplate),
    }

    // 注册内置桥接函数
    bridge.registerBuiltinFunctions()

    return bridge
}

// registerBuiltinFunctions 注册内置函数
func (b *EventBridge) registerBuiltinFunctions() {
    global := b.context.Global()

    // setState - 从 JS 更新状态
    b.RegisterFunction("setState", func(info *v8go.FunctionCallbackInfo) *v8go.Value {
        // 实现见下文
        return nil
    })

    // markDirty - 标记需要重绘
    b.RegisterFunction("markDirty", func(info *v8go.FunctionCallbackInfo) *v8go.Value {
        // 实现见下文
        return nil
    })

    // emitEvent - 从 JS 发射事件到 Go
    b.RegisterFunction("emitEvent", func(info *v8go.FunctionCallbackInfo) *v8go.Value {
        // 实现见下文
        return nil
    })

    // log - 日志输出
    b.RegisterFunction("log", func(info *v8go.FunctionCallbackInfo) *v8go.Value {
        args := info.Args()
        for _, arg := range args {
            str, _ := arg.ToString()
            log.Printf("[JS] %s", str)
        }
        return nil
    })
}

// RegisterFunction 注册 Go 函数供 JS 调用
func (b *EventBridge) RegisterFunction(name string, fn v8go.FunctionCallback) {
    tmpl := v8go.NewFunctionTemplate(b.isolate, fn)
    b.templates[name] = tmpl

    fnVal, _ := tmpl.GetFunction()
    global := b.context.Global()
    global.Set(name, fnVal)
}

// CallAction 调用 JS 中的 onAction 方法
func (b *EventBridge) CallAction(instance *v8go.Object, action *Action) (bool, error) {
    b.mu.Lock()
    defer b.mu.Unlock()

    // 获取 onAction 方法
    onActionVal, err := instance.Get("onAction")
    if err != nil {
        return false, nil // 没有 onAction 方法，未处理
    }

    onActionFn, err := onActionVal.AsFunction()
    if err != nil {
        return false, nil // 不是函数，未处理
    }

    // 转换 Action 为 JS 对象
    actionObj, err := b.actionToJS(action)
    if err != nil {
        return false, err
    }

    // 调用 JS 函数
    result, err := onActionFn(instance, actionObj)
    if err != nil {
        return false, fmt.Errorf("onAction error: %w", err)
    }

    // 解析返回值
    if result.IsUndefined() || result.IsNull() {
        return false, nil
    }

    handled, err := result.AsBoolean()
    if err != nil {
        return false, nil
    }

    return handled, nil
}

// CallRender 调用 JS 中的 render 方法
func (b *EventBridge) CallRender(instance *v8go.Object, ctx *RenderContext, buf *Buffer) error {
    b.mu.Lock()
    defer b.mu.Unlock()

    // 获取 render 方法
    renderVal, err := instance.Get("render")
    if err != nil {
        return err
    }

    renderFn, err := renderVal.AsFunction()
    if err != nil {
        return err
    }

    // 创建渲染上下文对象
    ctxObj := b.createRenderContext(ctx, buf)

    // 调用 render
    _, err = renderFn(instance, ctxObj)
    return err
}

// CallLifecycle 调用生命周期钩子
func (b *EventBridge) CallLifecycle(instance *v8go.Object, hook string) error {
    b.mu.Lock()
    defer b.mu.Unlock()

    hookVal, err := instance.Get(hook)
    if err != nil {
        return nil // 没有这个钩子，跳过
    }

    hookFn, err := hookVal.AsFunction()
    if err != nil {
        return nil // 不是函数，跳过
    }

    _, err = hookFn(instance)
    return err
}

// actionToJS 转换 Action 为 JS 对象
func (b *EventBridge) actionToJS(action *Action) (*v8go.Value, error) {
    obj := b.context.CreateObject(nil)

    // type
    typeVal, _ := v8go.NewValue(b.context).Create(string(action.Type))
    obj.Set("type", typeVal)

    // target
    if action.Target != "" {
        targetVal, _ := v8go.NewValue(b.context).Create(action.Target)
        obj.Set("target", targetVal)
    }

    // payload
    if action.Payload != nil {
        payloadVal, err := b.goToJS(action.Payload)
        if err == nil {
            obj.Set("payload", payloadVal)
        }
    }

    // timestamp
    tsVal, _ := v8go.NewValue(b.context).Create(action.Timestamp.Unix())
    obj.Set("timestamp", tsVal)

    return obj, nil
}

// goToJS Go 值转 JS 值
func (b *EventBridge) goToJS(value interface{}) (*v8go.Value, error) {
    switch v := value.(type) {
    case string:
        return v8go.NewValue(b.context).Create(v)
    case int, int32, int64:
        val, _ := v.(int64)
        return v8go.NewValue(b.context).Create(val)
    case float32, float64:
        val, _ := v.(float64)
        return v8go.NewValue(b.context).Create(val)
    case bool:
        return v8go.NewValue(b.context).Create(v)
    case map[string]interface{}:
        obj := b.context.CreateObject(nil)
        for key, val := range v {
            jsVal, err := b.goToJS(val)
            if err != nil {
                continue
            }
            obj.Set(key, jsVal)
        }
        return obj, nil
    case []interface{}:
        arr := b.context.CreateArray(nil)
        for i, val := range v {
            jsVal, err := b.goToJS(val)
            if err != nil {
                continue
            }
            arr.Set(uint32(i), jsVal)
        }
        return arr, nil
    default:
        return b.context.Null(), nil
    }
}

// createRenderContext 创建渲染上下文
func (b *EventBridge) createRenderContext(ctx *RenderContext, buf *Buffer) *v8go.Object {
    ctxObj := b.context.CreateObject(nil)

    // 注入绘制方法
    ctxObj.Set("writeCell", b.createDrawCellFunc(buf))
    ctxObj.Set("writeText", b.createDrawTextFunc(buf))
    ctxObj.Set("fillRect", b.createFillRectFunc(buf))
    ctxObj.Set("drawBox", b.createDrawBoxFunc(buf))

    // 注入上下文属性
    bounds := b.context.CreateObject(nil)
    bounds.Set("x", b.createNumber(float64(ctx.Bounds.X)))
    bounds.Set("y", b.createNumber(float64(ctx.Bounds.Y)))
    bounds.Set("width", b.createNumber(float64(ctx.Bounds.Width)))
    bounds.Set("height", b.createNumber(float64(ctx.Bounds.Height)))
    ctxObj.Set("bounds", bounds)

    return ctxObj
}

// createDrawCellFunc 创建绘制单元格函数
func (b *EventBridge) createDrawCellFunc(buf *Buffer) *v8go.Value {
    return b.context.CreateFunction("writeCell", func(info *v8go.FunctionCallbackInfo) *v8go.Value {
        if len(info.Args()) < 3 {
            return v8go.NullValue(b.context)
        }

        x, _ := info.Args()[0].AsInteger()
        y, _ := info.Args()[1].AsInteger()
        char, _ := info.Args()[2].ToString()

        buf.SetCell(int(x), int(y), []rune(char)[0], Style{})
        return v8go.UndefinedValue(b.context)
    })
}

// createDrawTextFunc 创建绘制文本函数
func (b *EventBridge) createDrawTextFunc(buf *Buffer) *v8go.Value {
    return b.context.CreateFunction("writeText", func(info *v8go.FunctionCallbackInfo) *v8go.Value {
        if len(info.Args()) < 3 {
            return v8go.NullValue(b.context)
        }

        x, _ := info.Args()[0].AsInteger()
        y, _ := info.Args()[1].AsInteger()
        text, _ := info.Args()[2].ToString()

        buf.DrawText(int(x), int(y), text, Style{})
        return v8go.UndefinedValue(b.context)
    })
}

// createNumber 创建数字值
func (b *EventBridge) createNumber(n float64) *v8go.Value {
    val, _ := v8go.NewValue(b.context).Create(n)
    return val
}
```

### 2. ScriptComponent 集成事件桥接

```go
// 位于: tui/framework/v8/script_component.go (更新)

package v8

type ScriptComponent struct {
    // ... 其他字段

    // 事件桥接器
    bridge *EventBridge
}

// HandleAction 处理 Action（通过桥接调用 JS）
func (c *ScriptComponent) HandleAction(action *action.Action) bool {
    if c.instance == nil {
        return false
    }

    // 通过桥接调用 JS 中的 onAction
    handled, err := c.bridge.CallAction(c.instance, action)
    if err != nil {
        log.Errorf("onAction error: %v", err)

        // 将错误附加到 Action
        action.AttachError(err)
        return false
    }

    return handled
}

// Paint 绘制组件（通过桥接调用 JS）
func (c *ScriptComponent) Paint(ctx PaintContext, buf *runtime.CellBuffer) {
    if c.instance == nil {
        return
    }

    // 创建 Go Buffer 包装器
    goBuf := NewGoBufferWrapper(buf)

    // 创建渲染上下文
    renderCtx := &RenderContext{
        Bounds: c.bounds,
        Style:  ctx.Style,
    }

    // 通过桥接调用 JS 中的 render
    if err := c.bridge.CallRender(c.instance, renderCtx, goBuf); err != nil {
        log.Errorf("render error: %v", err)
    }
}
```

### 3. JavaScript 组件示例

```javascript
// scripts/my_component.ui.js

export default {
    name: 'MyComponent',

    // 状态
    state: {
        items: ['Item 1', 'Item 2', 'Item 3'],
        selectedIndex: 0
    },

    // 生命周期钩子
    onMount() {
        console.log('Component mounted!');
        // 可以在这里设置定时器、订阅事件等
    },

    onUpdate(prevProps, prevState) {
        console.log('Component updated');
    },

    onUnmount() {
        console.log('Component unmounted');
        // 清理资源
    },

    // 事件处理
    onAction(action) {
        console.log('Received action:', action.type);

        switch (action.type) {
            case 'navigate_down':
                // 向下导航
                if (this.state.selectedIndex < this.state.items.length - 1) {
                    this.setState({
                        selectedIndex: this.state.selectedIndex + 1
                    });
                    this.markDirty(); // 标记需要重绘
                    return true; // 已处理
                }
                break;

            case 'navigate_up':
                // 向上导航
                if (this.state.selectedIndex > 0) {
                    this.setState({
                        selectedIndex: this.state.selectedIndex - 1
                    });
                    this.markDirty();
                    return true;
                }
                break;

            case 'submit':
                // 提交选择
                const selectedItem = this.state.items[this.state.selectedIndex];
                this.emitEvent('selected', selectedItem);
                return true;
        }

        return false; // 未处理
    },

    // 渲染函数
    render(ctx) {
        const { x, y, width, height } = ctx.bounds;

        // 绘制边框
        ctx.drawBox(x, y, width, height, 'My List');

        // 绘制列表项
        for (let i = 0; i < this.state.items.length; i++) {
            const item = this.state.items[i];
            const itemY = y + 1 + i;

            // 选中的项高亮显示
            if (i === this.state.selectedIndex) {
                ctx.fillText(x + 2, itemY, '> ' + item);
            } else {
                ctx.fillText(x + 2, itemY, '  ' + item);
            }
        }

        // 绘制提示
        ctx.fillText(x, y + height - 1, '↑↓ Navigate, Enter to Submit');
    },

    // 工具方法（由桥接器注入）
    setState(newState) {
        // 由 Go 桥接器实现
        this.state = { ...this.state, ...newState };
    },

    markDirty() {
        // 由 Go 桥接器实现
    },

    emitEvent(name, data) {
        // 由 Go 桥接器实现
        console.log('Emitting event:', name, data);
    }
};
```

## 异步回调处理

### 问题：Promise 回调无法直接回到 Go

```javascript
// ❌ 这样不行：Promise 回调在 V8 线程执行
async function fetchData() {
    const response = await fetch('/api/data');
    this.setState({ data: response }); // 无法更新 Go 状态
}
```

### 解决方案：显式同步

```go
// Go 端提供异步任务管理

// AsyncOperation 异步操作
type AsyncOperation struct {
    Promise *v8go.Value
    Resolve func(interface{})
    Reject  func(error)
}

// RunAsync 运行异步操作
func (b *EventBridge) RunAsync(instance *v8go.Object, fnName string) *AsyncOperation {
    // 获取异步函数
    fnVal, err := instance.Get(fnName)
    if err != nil {
        return nil
    }

    fn, err := fnVal.AsFunction()
    if err != nil {
        return nil
    }

    // 调用函数，返回 Promise
    promise, _ := fn(instance)

    return &AsyncOperation{
        Promise: promise,
    }
}

// AwaitResult 等待 Promise 结果（阻塞）
func (op *AsyncOperation) AwaitResult() (interface{}, error) {
    // v8go 不支持真正的 await，需要手动泵送消息
    // 这里的实现取决于具体使用场景
    return nil, nil
}
```

### 更好的方案：回调模式

```javascript
// ✅ 推荐：使用回调而非 Promise
export default {
    name: 'DataComponent',

    state: {
        data: null,
        loading: false,
        error: null
    },

    // 使用 Go 提供的异步 API
    onAction(action) {
        switch (action.type) {
            case 'load':
                this.setState({ loading: true });
                this.markDirty();

                // 调用 Go 函数加载数据
                goFetchData('/api/data', (error, result) => {
                    this.setState({
                        loading: false,
                        data: result,
                        error: error
                    });
                    this.markDirty();
                });
                return true;
        }
        return false;
    }
};
```

```go
// 注册 goFetchData 函数
bridge.RegisterFunction("goFetchData", func(info *v8go.FunctionCallbackInfo) *v8go.Value {
    if len(info.Args()) < 2 {
        return v8go.NullValue(b.context)
    }

    url, _ := info.Args()[0].ToString()
    callback := info.Args()[1] // 回调函数

    // 异步获取数据
    go func() {
        data, err := fetchDataFromAPI(url) // Go 函数

        // 在主线程回调 JS
        runtime.RunOnMainThread(func() {
            if err != nil {
                errorVal, _ := v8go.NewValue(b.context).Create(err.Error())
                callback.Call(b.context.Global(), errorVal, b.context.Null())
            } else {
                dataVal, _ := b.goToJS(data)
                callback.Call(b.context.Global(), b.context.Null(), dataVal)
            }
        })
    }()

    return v8go.UndefinedValue(b.context)
})
```

## 线程安全

### V8 Isolate 线程限制

v8go 的 Isolate **不是线程安全**的，所有调用必须在创建 Isolate 的线程上执行。

```go
// ThreadSafeV8Bridge 线程安全的 V8 桥接器
type ThreadSafeV8Bridge struct {
    *EventBridge

    // 操作队列
    queue chan func()
    done  chan struct{}
}

// NewThreadSafeV8Bridge 创建线程安全桥接器
func NewThreadSafeV8Bridge(isolate *v8go.Isolate, context *v8go.Context) *ThreadSafeV8Bridge {
    bridge := &ThreadSafeV8Bridge{
        EventBridge: NewEventBridge(isolate, context),
        queue:       make(chan func(), 100),
        done:        make(chan struct{}),
    }

    // 启动事件循环
    go bridge.eventLoop()

    return bridge
}

// eventLoop V8 事件循环（必须在单线程运行）
func (b *ThreadSafeV8Bridge) eventLoop() {
    for {
        select {
        case fn := <-b.queue:
            fn()
        case <-b.done:
            return
        }
    }
}

// RunOnMainThread 在主线程执行
func (b *ThreadSafeV8Bridge) RunOnMainThread(fn func()) {
    b.queue <- fn
}

// CallAction 线程安全版本
func (b *ThreadSafeV8Bridge) CallAction(instance *v8go.Object, action *Action) (bool, error) {
    result := make(chan bool)
    errCh := make(chan error)

    b.queue <- func() {
        handled, err := b.EventBridge.CallAction(instance, action)
        result <- handled
        errCh <- err
    }

    return <-result, <-errCh
}
```

## 完整示例

```go
// 完整的 V8 组件使用示例

func main() {
    // 1. 创建 V8 Isolate
    isolate := v8go.NewIsolate()

    // 2. 创建线程安全桥接器
    bridge := v8.NewThreadSafeV8Bridge(isolate, v8go.NewContext(isolate))

    // 3. 注册 Go 函数
    bridge.RegisterFunction("goLog", func(info *v8go.FunctionCallbackInfo) *v8go.Value {
        msg, _ := info.Args()[0].ToString()
        log.Printf("[JS LOG] %s", msg)
        return v8go.UndefinedValue(info.Context())
    })

    bridge.RegisterFunction("goFetchData", func(info *v8go.FunctionCallbackInfo) *v8go.Value {
        // 实现见上文
        return v8go.UndefinedValue(info.Context())
    })

    // 4. 加载组件
    comp, err := v8.NewScriptComponent(isolate, "scripts/my_component.ui.js")
    if err != nil {
        log.Fatal(err)
    }

    // 5. 挂载组件
    runtime.Mount(comp)

    // 6. 运行
    runtime.Run()
}
```

## 总结

| 特性 | v8go 支持 | 实现方案 |
|------|-----------|----------|
| Go 调用 JS 函数 | ✅ | `Function.Call()` |
| JS 调用 Go 函数 | ✅ | `FunctionTemplate` + 回调 |
| 事件处理 | ✅ | 桥接器 `CallAction()` |
| 状态同步 | ✅ | `setState()` 桥接函数 |
| 异步回调 | ⚠️ | 需要手动泵送，推荐回调模式 |
| 线程安全 | ❌ | 使用任务队列串行化 |

**关键结论**：v8go **完全支持**事件回调，但需要正确设计桥接层。核心是：
1. 所有 JS 调用必须在创建 Isolate 的线程上执行
2. 使用任务队列串行化多线程访问
3. 异步操作推荐使用回调而非 Promise
