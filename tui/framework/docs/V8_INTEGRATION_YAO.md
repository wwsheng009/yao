# TUI V3 V8 集成方案（基于 Yao 现有基础设施）

> **版本**: V3
> **策略**: 复用 > 重写
> **基础**: `gou/runtime/v8/bridge` + `yao/tui/jsapi.go`
> **复用度**: 90% (直接复用现有 Yao 基础设施)

## 概述

本方案基于 Yao 框架现有的 v8go 基础设施，扩展实现 TUI V3 框架的 V8 集成功能。

### 为什么需要 V8 集成？

**纯 Go 代码的问题**：
```go
// ❌ 每次修改都要重新编译
type MyComponent struct {
    BaseComponent
}

func (c *MyComponent) Paint(...) {
    // 硬编码的 UI
}

// 问题：
// - 需要重新编译
// - 无法热重载
// - 扩展困难
```

**V8 脚本的优势**：
```javascript
// ✅ 脚本驱动的 UI
export default {
  name: 'MyComponent',
  props: ['title', 'items'],
  render(ctx, buf) {
    buf.write(0, 0, this.title);
    // ...
  }
}

// 优势：
// - 热重载
// - 灵活扩展
// - 无需重新编译
```

### 设计目标

1. **组件定义**: 支持 JS/TS 定义组件
2. **事件处理**: 支持脚本处理事件
3. **数据绑定**: 支持响应式数据
4. **生命周期**: 支持组件生命周期钩子
5. **类型安全**: 支持 TypeScript 类型
6. **性能**: 高效的 Go-JS 互操作

## 现有基础设施分析

### 可复用的组件

| 组件 | 位置 | 功能 |
|------|------|------|
| `bridge.JsValue()` | `gou/runtime/v8/bridge/bridge.go` | Go → JS 类型转换 |
| `bridge.GoValue()` | `gou/runtime/v8/bridge/bridge.go` | JS → Go 类型转换 |
| `bridge.FunctionT.Call()` | `gou/runtime/v8/bridge/function.go` | Go 调用 JS 函数 |
| `bridge.PromiseT.Result()` | `gou/runtime/v8/bridge/promise.go` | Promise 结果获取 |
| `bridge.JsException()` | `gou/runtime/v8/bridge/bridge.go` | 创建 JS 错误对象 |
| `tui.SetState()` | `yao/tui/jsapi.go` | 设置状态 |
| `tui.PublishEvent()` | `yao/tui/jsapi.go` | 发布事件 |
| `tui.SubscribeToEvent()` | `yao/tui/jsapi.go` | 订阅事件 |
| `tui.SetFocus()` | `yao/tui/jsapi.go` | 设置焦点 |

### 需要新增的功能

| 功能 | 说明 |
|------|------|
| 组件生命周期 | `onMount`, `onUpdate`, `onUnmount` |
| 渲染方法 | `render(ctx, buf)` |
| 组件注册表 | `ComponentRegistry` |
| 脚本热重载 | 文件监控 + 重新加载 |

## 架构设计

```
┌─────────────────────────────────────────────────────────────────┐
│                        TUI V3 Framework                         │
│                                                                  │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │              ComponentRegistry (NEW)                       │  │
│  │  - Register(scriptPath)                                    │  │
│  │  - Create(name)                                            │  │
│  │  - Reload(name)                                             │  │
│  └───────────────────────────────────────────────────────────┘  │
│                           │                                     │
│                           ▼                                     │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │            ScriptComponent (EXTENDED)                      │  │
│  │  - onMount/onUpdate/onUnmount (NEW)                      │  │
│  │  - render(ctx, buf) (NEW)                                  │  │
│  │  - onAction(action) (NEW)                                  │  │
│  └───────────────────────────────────────────────────────────┘  │
│                           │                                     │
│                           ▼                                     │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │              TUI API Bridge (EXTEND)                        │  │
│  │  - 基于 yao/tui/jsapi.go                                  │  │
│  │  - 复用 bridge.JsValue/GoValue                            │  │
│  │  - 新增组件生命周期支持                                    │  │
│  └───────────────────────────────────────────────────────────┘  │
│                           │                                     │
│                           ▼                                     │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │              gou/runtime/v8/bridge (REUSE)                  │  │
│  │  - JsValue() / GoValue()                                  │  │
│  │  - FunctionT.Call()                                        │  │
│  │  - PromiseT.Result()                                       │  │
│  │  - JsException()                                           │  │
│  └───────────────────────────────────────────────────────────┘  │
│                           │                                     │
│                           ▼                                     │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                    v8go Isolate                            │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## 核心实现

### 1. 扩展 TUI API Bridge

```go
// 位于: yao/tui/v3/component_bridge.go

package v3

import (
    "fmt"
    "sync"

    "github.com/yaoapp/gou/runtime/v8/bridge"
    "github.com/yaoapp/yao/tui/core"
    "rogchap.com/v8go"
)

// ComponentBridge 组件桥接器（基于现有 jsapi 扩展）
type ComponentBridge struct {
    // 复用现有的 Model
    model *Model

    // 组件注册表
    components map[string]*ScriptComponent

    // 热重载
    watcher *fs.Watcher

    mu sync.RWMutex
}

// NewComponentBridge 创建组件桥接器
func NewComponentBridge(model *Model) *ComponentBridge {
    return &ComponentBridge{
        model:      model,
        components: make(map[string]*ScriptComponent),
    }
}

// RegisterComponent 注册组件
func (b *ComponentBridge) RegisterComponent(name, scriptPath string) error {
    b.mu.Lock()
    defer b.mu.Unlock()

    // 读取脚本
    code, err := os.ReadFile(scriptPath)
    if err != nil {
        return err
    }

    // 创建 ScriptComponent
    comp := &ScriptComponent{
        Name:       name,
        ScriptPath:  scriptPath,
        ScriptCode:  string(code),
        bridge:      b,
        model:       b.model,
        state:       make(map[string]interface{}),
        props:       make(map[string]interface{}),
    }

    b.components[name] = comp
    return nil
}

// CreateComponent 创建组件实例
func (b *ComponentBridge) CreateComponent(name string) (*ScriptComponentInstance, error) {
    b.mu.RLock()
    template, ok := b.components[name]
    b.mu.RUnlock()

    if !ok {
        return nil, fmt.Errorf("component %s not found", name)
    }

    return template.CreateInstance()
}

// GetGlobalObject 获取全局 JS 对象（扩展版）
func (b *ComponentBridge) GetGlobalObject(v8ctx *v8go.Context) (*v8go.Value, error) {
    // 复用现有的 injectModelToContext
    // 但添加组件相关的全局函数

    global := v8ctx.Global()

    // 添加组件注册函数
    registerFn := v8go.NewFunctionTemplate(v8ctx.Isolate(), func(info *v8go.FunctionCallbackInfo) *v8go.Value {
        // 实现见下文
        return v8go.Undefined(v8ctx.Isolate())
    })
    global.Set("RegisterComponent", registerFn)

    // 添加渲染上下文函数
    renderCtxFn := v8go.NewFunctionTemplate(v8ctx.Isolate(), func(info *v8go.FunctionCallbackInfo) *v8go.Value {
        return b.createRenderContext(info.Context())
    })
    global.Set("GetRenderContext", renderCtxFn)

    return global, nil
}

// createRenderContext 创建渲染上下文
func (b *ComponentBridge) createRenderContext(v8ctx *v8go.Context) *v8go.Value {
    ctxObj := v8go.NewObjectTemplate(v8ctx.Isolate())

    // 注入绘制方法（基于 Paint 抽象）
    ctxObj.Set("writeCell", b.createWriteCellMethod(v8ctx))
    ctxObj.Set("writeText", b.createWriteTextMethod(v8ctx))
    ctxObj.Set("fillRect", b.createFillRectMethod(v8ctx))
    ctxObj.Set("drawBox", b.createDrawBoxMethod(v8ctx))

    instance, _ := ctxObj.NewInstance(v8ctx)
    return instance.Value
}

// createWriteCellMethod 创建 writeCell 方法
func (b *ComponentBridge) createWriteCellMethod(v8ctx *v8go.Context) *v8go.FunctionTemplate {
    return v8go.NewFunctionTemplate(v8ctx.Isolate(), func(info *v8go.FunctionCallbackInfo) *v8go.Value {
        args := info.Args()
        if len(args) < 3 {
            return bridge.JsException(v8ctx, "writeCell requires x, y, char arguments")
        }

        x, _ := args[0].AsInteger()
        y, _ := args[1].AsInteger()
        char, _ := args[2].ToString()

        // 通过消息发送绘制命令
        b.model.Program.Send(core.PaintMsg{
            Type:  "writeCell",
            X:     int(x),
            Y:     int(y),
            Char:  char,
            Style: b.getCurrentStyle(),
        })

        return v8go.Undefined(v8ctx.Isolate())
    })
}

// createWriteTextMethod 创建 writeText 方法
func (b *ComponentBridge) createWriteTextMethod(v8ctx *v8go.Context) *v8go.FunctionTemplate {
    return v8go.NewFunctionTemplate(v8ctx.Isolate(), func(info *v8go.FunctionCallbackInfo) *v8go.Value {
        args := info.Args()
        if len(args) < 3 {
            return bridge.JsException(v8ctx, "writeText requires x, y, text arguments")
        }

        x, _ := args[0].AsInteger()
        y, _ := args[1].AsInteger()
        text, _ := args[2].ToString()

        b.model.Program.Send(core.PaintMsg{
            Type:  "writeText",
            X:     int(x),
            Y:     int(y),
            Text:  text,
            Style: b.getCurrentStyle(),
        })

        return v8go.Undefined(v8ctx.Isolate())
    })
}

// getCurrentStyle 获取当前样式
func (b *ComponentBridge) getCurrentStyle() core.Style {
    // 从模型获取当前样式
    return core.Style{
        Foreground: core.ColorDefault,
        Background: core.ColorDefault,
    }
}
```

### 2. ScriptComponent 实现

```go
// 位于: yao/tui/v3/script_component.go

package v3

import (
    "sync"

    "github.com/yaoapp/gou/runtime/v8/bridge"
    "github.com/yaoapp/yao/tui/core"
    "rogchap.com/v8go"
)

// ScriptComponent 脚本组件
type ScriptComponent struct {
    // 基本信息
    Name       string
    ScriptPath  string
    ScriptCode  string

    // 桥接器
    bridge  *ComponentBridge
    model   *Model

    // 实例模板
    instanceTemplate *v8go.ObjectTemplate

    // 状态
    state   map[string]interface{}
    props   map[string]interface{}
    computed map[string]func() interface{}

    mu sync.RWMutex
}

// ScriptComponentInstance 脚本组件实例
type ScriptComponentInstance struct {
    // 组件定义
    definition *ScriptComponent

    // V8 上下文
    ctx    *v8go.Context
    instance *v8go.Object

    // 状态
    state   map[string]interface{}
    props   map[string]interface{}

    // 生命周期状态
    mounted bool
}

// LoadScript 加载脚本
func (c *ScriptComponent) LoadScript(v8ctx *v8go.Context) error {
    // 执行脚本，获取导出的组件定义
    result, err := v8ctx.RunCode(c.ScriptCode, c.ScriptPath)
    if err != nil {
        return fmt.Errorf("script error: %w", err)
    }

    // 解析组件定义
    def, err := result.AsObject()
    if err != nil {
        return fmt.Errorf("component definition must be an object: %w", err)
    }

    // 解析各个字段
    if name, err := def.Get("name"); err == nil {
        c.Name, _ = name.ToString()
    }

    // 创建实例模板
    c.instanceTemplate = v8go.NewObjectTemplate(v8ctx.Isolate())

    // 注入内置方法
    c.instanceTemplate.Set("setState", c.createSetStateMethod(v8ctx))
    c.instanceTemplate.Set("getState", c.createGetStateMethod(v8ctx))
    c.instanceTemplate.Set("markDirty", c.createMarkDirtyMethod(v8ctx))

    // 复制用户定义的字段
    keys, _ := def.GetOwnPropertyNames()
    for i := 0; i < keys.Length(); i++ {
        key, _ := keys.Get(uint32(i))
        keyStr, _ := key.ToString()

        // 跳过内置方法
        if c.isBuiltinMethod(keyStr) {
            continue
        }

        value, _ := def.Get(keyStr)
        c.instanceTemplate.Set(keyStr, value)
    }

    return nil
}

// isBuiltinMethod 检查是否是内置方法
func (c *ScriptComponent) isBuiltinMethod(name string) bool {
    builtins := []string{
        "setState", "getState", "markDirty",
        "onMount", "onUpdate", "onUnmount",
        "render", "onAction",
    }
    for _, b := range builtins {
        if b == name {
            return true
        }
    }
    return false
}

// createSetStateMethod 创建 setState 方法
func (c *ScriptComponent) createSetStateMethod(v8ctx *v8go.Context) *v8go.FunctionTemplate {
    return v8go.NewFunctionTemplate(v8ctx.Isolate(), func(info *v8go.FunctionCallbackInfo) *v8go.Value {
        args := info.Args()

        if len(args) < 1 {
            return bridge.JsException(v8ctx, "setState requires an object argument")
        }

        newState, err := bridge.GoValue(args[0], v8ctx)
        if err != nil {
            return bridge.JsException(v8ctx, "Invalid state object: "+err.Error())
        }

        stateMap, ok := newState.(map[string]interface{})
        if !ok {
            return bridge.JsException(v8ctx, "State must be an object")
        }

        // 更新状态（通过消息发送）
        // ... 发送 StateUpdateMsg

        return v8go.Undefined(v8ctx.Isolate())
    })
}

// createGetStateMethod 创建 getState 方法
func (c *ScriptComponent) createGetStateMethod(v8ctx *v8go.Context) *v8go.FunctionTemplate {
    return v8go.NewFunctionTemplate(v8ctx.Isolate(), func(info *v8go.FunctionCallbackInfo) *v8go.Value {
        // 从 this 获取组件实例
        thisObj, err := info.This().AsObject()
        if err != nil {
            return bridge.JsException(v8ctx, "Invalid this context")
        }

        // 获取内部字段存储的组件实例
        instanceID := thisObj.GetInternalField(0)
        if instanceID.IsNullOrUndefined() {
            return bridge.JsException(v8ctx, "Component instance not found")
        }

        // ... 获取状态并返回

        return v8go.Undefined(v8ctx.Isolate())
    })
}

// createMarkDirtyMethod 创建 markDirty 方法
func (c *ScriptComponent) createMarkDirtyMethod(v8ctx *v8go.Context) *v8go.FunctionTemplate {
    return v8go.NewFunctionTemplate(v8ctx.Isolate(), func(info *v8go.FunctionCallbackInfo) *v8go.Value {
        // 发送重绘消息
        // ... 发送 PaintMsg

        return v8go.Undefined(v8ctx.Isolate())
    })
}

// CreateInstance 创建组件实例
func (c *ScriptComponent) CreateInstance() (*ScriptComponentInstance, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    // 获取 V8 上下文
    v8ctx := c.model.GetV8Context()

    // 创建实例
    instance, err := c.instanceTemplate.NewInstance(v8ctx)
    if err != nil {
        return nil, err
    }

    instanceObj, err := instance.Value.AsObject()
    if err != nil {
        return nil, err
    }

    // 存储组件定义引用
    goValueID := bridge.RegisterGoObject(c)
    instanceObj.SetInternalFieldCount(1)
    instanceObj.SetInternalField(0, goValueID)

    return &ScriptComponentInstance{
        definition: c,
        ctx:        v8ctx,
        instance:   instanceObj,
        state:      make(map[string]interface{}),
        props:      make(map[string]interface{}),
    }, nil
}
```

### 3. 扩展现有 jsapi.go

```go
// 位于: yao/tui/jsapi_v3.go (扩展现有功能)

package tui

import (
    "github.com/yaoapp/gou/runtime/v8/bridge"
    "github.com/yaoapp/yao/tui/core"
    "rogchap.com/v8go"
)

// NewTUIObjectV3 创建 V3 版本的 TUI 对象
func NewTUIObjectV3(v8ctx *v8go.Context, model *Model) (*v8go.Value, error) {
    jsObjTbl := v8go.NewObjectTemplate(v8ctx.Isolate())

    // === 复用现有的方法 ===
    jsObjTbl.Set("GetState", tuiGetStateMethod(v8ctx.Isolate(), model))
    jsObjTbl.Set("SetState", tuiSetStateMethod(v8ctx.Isolate(), model))
    jsObjTbl.Set("UpdateState", tuiUpdateStateMethod(v8ctx.Isolate(), model))
    jsObjTbl.Set("ExecuteAction", tuiExecuteActionMethod(v8ctx.Isolate(), model))

    // === 新增 V3 方法 ===
    jsObjTbl.Set("RegisterComponent", tuiRegisterComponentMethod(v8ctx.Isolate(), model))
    jsObjTbl.Set("CreateComponent", tuiCreateComponentMethod(v8ctx.Isolate(), model))
    jsObjTbl.Set("RenderComponent", tuiRenderComponentMethod(v8ctx.Isolate(), model))

    instance, err := jsObjTbl.NewInstance(v8ctx)
    if err != nil {
        return nil, err
    }

    return instance.Value, nil
}

// tuiRegisterComponentMethod 注册组件
func tuiRegisterComponentMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
    return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
        v8ctx := info.Context()
        args := info.Args()

        if len(args) < 2 {
            return bridge.JsException(v8ctx, "RegisterComponent requires name and path arguments")
        }

        name := args[0].String()
        path := args[1].String()

        // 通过 ComponentBridge 注册
        if model.ComponentBridge != nil {
            err := model.ComponentBridge.RegisterComponent(name, path)
            if err != nil {
                return bridge.JsException(v8ctx, "Failed to register: "+err.Error())
            }
        }

        return v8go.Undefined(iso)
    })
}

// tuiCreateComponentMethod 创建组件实例
func tuiCreateComponentMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
    return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
        v8ctx := info.Context()
        args := info.Args()

        if len(args) < 1 {
            return bridge.JsException(v8ctx, "CreateComponent requires name argument")
        }

        name := args[0].String()

        // 通过 ComponentBridge 创建
        if model.ComponentBridge != nil {
            instance, err := model.ComponentBridge.CreateComponent(name)
            if err != nil {
                return bridge.JsException(v8ctx, "Failed to create: "+err.Error())
            }

            // 返回组件的 JS 表示
            return instance.ToJSValue(v8ctx)
        }

        return v8go.Undefined(iso)
    })
}

// tuiRenderComponentMethod 渲染组件
func tuiRenderComponentMethod(iso *v8go.Isolate, model *Model) *v8go.FunctionTemplate {
    return v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
        v8ctx := info.Context()
        args := info.Args()

        if len(args) < 1 {
            return bridge.JsException(v8ctx, "RenderComponent requires name argument")
        }

        name := args[0].String()

        // 通过 ComponentBridge 获取组件并渲染
        if model.ComponentBridge != nil {
            instance, err := model.ComponentBridge.CreateComponent(name)
            if err != nil {
                return bridge.JsException(v8ctx, "Failed to get component: "+err.Error())
            }

            // 调用组件的 render 方法
            result, err := instance.CallRender()
            if err != nil {
                return bridge.JsException(v8ctx, "Render failed: "+err.Error())
            }

            // 返回渲染结果
            return bridge.JsValue(v8ctx, result)
        }

        return v8go.Undefined(iso)
    })
}
```

### 4. 组件热重载

```go
// 位于: yao/tui/v3/hot_reload.go

package v3

import (
    "os"
    "path/filepath"
    "time"

    "github.com/fsnotify/fsnotify"
)

// HotReloader 热重载器
type HotReloader struct {
    bridge *ComponentBridge
    watcher *fsnotify.Watcher
    model  *Model
}

// NewHotReloader 创建热重载器
func NewHotReloader(bridge *ComponentBridge, model *Model) (*HotReloader, error) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }

    return &HotReloader{
        bridge:  bridge,
        watcher: watcher,
        model:   model,
    }, nil
}

// Watch 监控组件目录
func (h *HotReloader) Watch(dir string) error {
    err := h.watcher.Add(dir)
    if err != nil {
        return err
    }

    // 启动监控循环
    go h.watchLoop()

    return nil
}

// watchLoop 监控循环
func (h *HotReloader) watchLoop() {
    debounce := make(map[string]time.Time)

    for {
        select {
        case event, ok := <-h.watcher.Events:
            if !ok {
                return
            }

            // 只处理 .js 和 .ts 文件
            if !strings.HasSuffix(event.Name, ".js") && !strings.HasSuffix(event.Name, ".ts") {
                continue
            }

            // 防抖：同一文件 500ms 内只处理一次
            if lastTime, ok := debounce[event.Name]; ok {
                if time.Since(lastTime) < 500*time.Millisecond {
                    continue
                }
            }
            debounce[event.Name] = time.Now()

            // 重新加载组件
            h.reloadComponent(event.Name)

        case <-time.After(1 * time.Second):
            // 清理过期的防抖记录
            now := time.Now()
            for file, lastTime := range debounce {
                if now.Sub(lastTime) > 1*time.Second {
                    delete(debounce, file)
                }
            }
        }
    }
}

// reloadComponent 重新加载组件
func (h *HotReloader) reloadComponent(scriptPath string) {
    // 查找对应的组件名称
    compName := h.findComponentByPath(scriptPath)
    if compName == "" {
        return
    }

    // 重新注册组件
    err := h.bridge.RegisterComponent(compName, scriptPath)
    if err != nil {
        // 记录错误但不崩溃
        log.Printf("Failed to reload component %s: %v", compName, err)
        return
    }

    // 发送重载事件
    h.model.Program.Send(core.ComponentReloadMsg{
        Name: compName,
        Path: scriptPath,
    })

    log.Printf("Component %s reloaded", compName)
}

// findComponentByPath 根据路径查找组件名称
func (h *HotReloader) findComponentByPath(path string) string {
    // 遍历所有注册的组件，找到匹配的
    h.bridge.mu.RLock()
    defer h.bridge.mu.RUnlock()

    for name, comp := range h.bridge.components {
        if comp.ScriptPath == path {
            return name
        }
    }

    return ""
}
```

### 5. JavaScript 组件示例

```javascript
// scripts/components/my_list.ui.js

// 注册组件（复用现有 xui 全局对象）
xui.RegisterComponent('MyList', 'scripts/components/my_list.ui.js');

// 组件定义
export default {
    name: 'MyList',

    // 状态（复用现有的 setState）
    state: {
        items: ['Item 1', 'Item 2', 'Item 3'],
        selectedIndex: 0
    },

    // 生命周期钩子
    onMount() {
        console.log('MyList mounted');
        // 可以在这里设置定时器、订阅事件等
    },

    onUpdate(prevProps, prevState) {
        console.log('MyList updated');
        // 处理更新
    },

    onUnmount() {
        console.log('MyList unmounted');
        // 清理资源
    },

    // 渲染方法（使用现有的 GetRenderContext）
    render() {
        const ctx = xui.GetRenderContext();

        // 绘制边框
        ctx.drawBox(0, 0, 40, 10, 'My List');

        // 绘制列表项
        for (let i = 0; i < this.state.items.length; i++) {
            const item = this.state.items[i];
            const y = 1 + i;

            // 选中的项高亮
            if (i === this.state.selectedIndex) {
                ctx.fillText(2, y, '> ' + item);
            } else {
                ctx.fillText(2, y, '  ' + item);
            }
        }

        // 绘制提示
        ctx.fillText(0, 9, '↑↓ Navigate, Enter to Submit');
    },

    // 事件处理（复用现有的 ExecuteAction）
    onAction(action) {
        console.log('Received action:', action.type);

        switch (action.type) {
            case 'navigate_down':
                if (this.state.selectedIndex < this.state.items.length - 1) {
                    this.setState({
                        selectedIndex: this.state.selectedIndex + 1
                    });
                    xui.Refresh(); // 触发重绘
                    return true;
                }
                break;

            case 'navigate_up':
                if (this.state.selectedIndex > 0) {
                    this.setState({
                        selectedIndex: this.state.selectedIndex - 1
                    });
                    xui.Refresh();
                    return true;
                }
                break;

            case 'submit':
                const selectedItem = this.state.items[this.state.selectedIndex];
                xui.PublishEvent('MyList', 'selected', selectedItem);
                return true;
        }

        return false;
    },

    // 工具方法（由桥接器注入）
    setState(newState) {
        // 通过 xui.SetState 调用 Go 方法
        xui.SetState('items', newState);
        xui.SetState('selectedIndex', newState.selectedIndex);
    },

    markDirty() {
        xui.Refresh();
    }
};
```

### 6. 消息扩展

```go
// 位于: yao/tui/core/messages_v3.go

package core

// ComponentReloadMsg 组件重载消息
type ComponentReloadMsg struct {
    Name string
    Path string
}

// PaintMsg 绘制消息
type PaintMsg struct {
    Type  string // "writeCell", "writeText", "fillRect", "drawBox"
    X     int
    Y     int
    Char  string
    Text  string
    Width int
    Height int
    Style Style
}

// RenderRequestMsg 渲染请求消息
type RenderRequestMsg struct {
    ComponentID string
    InstanceID string
}
```

## 集成步骤

### 步骤 1: 扩展 Model

```go
// 位于: yao/tui/model.go (扩展)

type Model struct {
    // ... 现有字段

    // V3 新增
    ComponentBridge *v3.ComponentBridge
    HotReloader     *v3.HotReloader
}

// InitV3 初始化 V3 功能
func (m *Model) InitV3() error {
    // 创建组件桥接器
    m.ComponentBridge = v3.NewComponentBridge(m)

    // 创建热重载器
    watcher, err := v3.NewHotReloader(m.ComponentBridge, m)
    if err != nil {
        return err
    }

    // 监控 scripts 目录
    if err := watcher.Watch("scripts/components"); err != nil {
        log.Printf("Warning: component hot reload disabled: %v", err)
    } else {
        m.HotReloader = watcher
    }

    return nil
}
```

### 步骤 2: 更新 Init 函数

```go
// 位于: yao/tui/init.go (修改)

func InitV3() error {
    // 现有的初始化代码
    // ...

    // V3 初始化
    if model != nil {
        if err := model.InitV3(); err != nil {
            return err
        }
    }

    return nil
}
```

### 步骤 3: 更新消息处理

```go
// 位于: yao/tui/update.go (扩展)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    // ... 现有消息处理 ...

    // V3 新增消息
    case v3.ComponentReloadMsg:
        // 组件重载
        return m, m.handleComponentReload(msg)

    case core.PaintMsg:
        // 绘制消息（来自 JS render）
        return m, m.handlePaint(msg)

    case core.RenderRequestMsg:
        // 渲染请求
        return m, m.handleRenderRequest(msg)
    }

    return m, nil
}
```

## 复用策略总结

| 功能 | 复用 | 新增 |
|------|------|------|
| 类型转换 | ✅ `bridge.JsValue/GoValue` | - |
| 状态管理 | ✅ `SetState/GetState` | - |
| 事件系统 | ✅ `PublishEvent/SubscribeToEvent` | - |
| 焦点管理 | ✅ `SetFocus/FocusNextInput` | - |
| 组件注册表 | - | ✅ `ComponentRegistry` |
| 组件生命周期 | - | ✅ `onMount/onUpdate/onUnmount` |
| 渲染方法 | - | ✅ `render(ctx, buf)` |
| 热重载 | - | ✅ `HotReloader` |

## 使用示例

### 完整示例

```javascript
// 1. 定义组件（scripts/components/my_app.ui.js）
export default {
    name: 'MyApp',

    state: {
        message: 'Hello from V3!',
        count: 0
    },

    onMount() {
        // 启动定时器
        this.timer = setInterval(() => {
            this.state.count++;
            xui.Refresh();
        }, 1000);
    },

    onUnmount() {
        clearInterval(this.timer);
    },

    render() {
        const ctx = xui.GetRenderContext();

        // 绘制标题
        ctx.fillText(0, 0, this.state.message);

        // 绘制计数
        ctx.fillText(0, 2, `Count: ${this.state.count}`);

        // 绘制边框
        ctx.drawBox(0, 0, 40, 5, 'My App');
    },

    onAction(action) {
        if (action.type === 'quit') {
            xui.Quit();
            return true;
        }
        return false;
    }
};

// 2. 在 Go 中使用
func main() {
    model := tui.NewModel()
    model.InitV3() // 初始化 V3 功能

    // 注册脚本目录
    model.ComponentBridge.RegisterComponent('MyApp', 'scripts/components/my_app.ui.js')

    // 创建组件实例
    comp, _ := model.ComponentBridge.CreateComponent('MyApp')

    // 挂载到界面
    model.Mount(comp)

    // 运行
    program := tea.NewProgram(model)
    tea.NewProgram(program, tea.WithAltScreen()).Run()
}
```

## TypeScript 类型定义

```typescript
// 位于: scripts/types/component.d.ts

// 组件定义类型
export interface ComponentDefinition<P = any, S = any> {
  name: string;
  props?: (keyof P)[];
  state?: S;
  computed?: ComputedValues<P, S>;
  watch?: WatchValues<P, S>;

  onMount?(): void;
  onUpdate?(prevProps: P, prevState: S): void;
  onUnmount?(): void;

  render(ctx: RenderContext, buf: Buffer): void;
  onAction?(action: Action): boolean;
}

// 渲染上下文
export interface RenderContext {
  bounds: Rect;
  style: Style;
  theme: Theme;
}

// 缓冲区接口
export interface Buffer {
  writeCell(x: number, y: number, char: string, style: Style): void;
  writeText(x: number, y: number, text: string, style: Style): void;
  fillRect(x: number, y: number, width: number, height: number, char: string, style: Style): void;
  drawBox(x: number, y: number, width: number, height: number, title: string, style: Style): void;
  clear(style: Style): void;
}

// 样式
export interface Style {
  foreground?: Color;
  background?: Color;
  bold?: boolean;
  dim?: boolean;
  italic?: boolean;
  underline?: boolean;
  blink?: boolean;
  reverse?: boolean;
}

// Action
export interface Action {
  type: string;
  target?: string;
  payload?: any;
}

// 计算属性
export type ComputedValues<P, S> = {
  [K in keyof P | keyof S]: () => any;
};

// 监听器
export type WatchValues<P, S> = {
  [K in keyof P | keyof S]?: string;
};

// 矩形
export interface Rect {
  x: number;
  y: number;
  width: number;
  height: number;
}

// 颜色
export type Color =
  | "black" | "red" | "green" | "yellow"
  | "blue" | "magenta" | "cyan" | "white"
  | "brightBlack" | "brightRed" | "brightGreen"
  | "brightYellow" | "brightBlue" | "brightMagenta"
  | "brightCyan" | "brightWhite"
  | number; // 256-color or RGB value

// 主题
export interface Theme {
  name: string;
  colors: {
    primary: Color;
    secondary: Color;
    background: Color;
    foreground: Color;
    border: Color;
    error: Color;
    success: Color;
    warning: Color;
  };
  styles: {
    [key: string]: Style;
  };
}
```

## 测试

```go
// 位于: tui/framework/v3/component_test.go

package v3

import (
    "testing"
    "github.com/yaoapp/yao/tui/runtime"
    "github.com/stretchr/testify/assert"
)

func TestScriptComponent(t *testing.T) {
    bridge := NewComponentBridge(nil)

    // 注册测试组件
    err := bridge.RegisterComponent("test", "testdata/test.ui.js")
    assert.NoError(t, err)

    // 创建组件
    comp, err := bridge.CreateComponent("test")
    assert.NoError(t, err)

    // 设置属性
    comp.SetProp("title", "Test")

    // 测试渲染
    buf := runtime.NewCellBuffer(80, 24)
    ctx := PaintContext{
        Style: runtime.DefaultStyle(),
    }

    comp.Paint(ctx, buf)

    // 验证输出
    output := buf.String()
    assert.Contains(t, output, "Test")
}

func TestComponentReload(t *testing.T) {
    bridge := NewComponentBridge(nil)
    reloader, _ := NewHotReloader(bridge, nil)

    // 注册组件
    bridge.RegisterComponent("reload-test", "testdata/reload.ui.js")

    // 修改文件
    os.WriteFile("testdata/reload.ui.js", []byte(`
        export default {
            name: 'ReloadTest',
            render() { ctx.fillText(0, 0, 'Updated'); }
        }
    `), 0644)

    // 触发重载
    reloader.reloadComponent("testdata/reload.ui.js")

    // 验证更新
    comp, _ := bridge.CreateComponent("reload-test")
    // ... 验证组件已更新
}
```

## 总结

1. **完全复用** `gou/runtime/v8/bridge` 的类型转换
2. **扩展现有** `yao/tui/jsapi.go` 的 TUI API
3. **新增** 组件注册表、生命周期、渲染方法、热重载
4. **最小侵入** - 不破坏现有架构

现有基础设施已提供 90% 的功能，只需少量扩展即可满足 TUI V3 需求。

## 相关文档

- [ARCHITECTURE.md](ARCHITECTURE.md) - 架构总览
- [COMPONENTS.md](COMPONENTS.md) - 组件系统
- [ACTION_SYSTEM.md](ACTION_SYSTEM.md) - Action 系统
- [STATE_MANAGEMENT.md](STATE_MANAGEMENT.md) - 状态管理
- [PROCESS_INTEGRATION_YAO.md](PROCESS_INTEGRATION_YAO.md) - Yao Process 集成
- [V8_EVENT_CALLBACK.md](V8_EVENT_CALLBACK.md) - V8 事件回调桥接
