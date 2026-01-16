# TUI JavaScript/TypeScript 脚本集成指南

## 概述

TUI 引擎集成了 V8 JavaScript/TypeScript 运行时，允许开发者使用 JS/TS 脚本处理复杂的交互逻辑。此功能提供了强大的状态管理和事件处理能力，使 TUI 应用更加动态和响应式。

## 核心概念

### TUI JavaScript 对象

当脚本执行时，TUI 引擎会将一个上下文对象作为第一个参数传递给脚本函数。该上下文对象包含一个 `tui` 子对象，提供以下方法：

- **上下文访问**: `ctx.tui`
  - `ctx` 是传递给每个脚本函数的第一个参数
  - `tui` 是 `ctx` 对象的一个属性，提供对 TUI 功能的访问

**示例**:
```javascript
// 脚本函数接收 ctx 对象作为第一个参数
function myScript(ctx, additionalParam) {
  // 通过 ctx.tui 访问 TUI 功能
  const count = ctx.tui.GetState('count');
  
  // 更新状态
  ctx.tui.SetState('count', count + 1);
}
```

#### `ctx.tui.id`
- **类型**: `string`
- **描述**: 当前 TUI 配置的名称

#### `ctx.tui.width` / `ctx.tui.height`
- **类型**: `number`
- **描述**: 当前终端窗口尺寸（暂时不可用，将在窗口大小变化事件后更新）

#### `ctx.tui.GetState([key])`
- **参数**:
  - `key` (可选): 字符串，状态键名。支持点符号访问嵌套属性（如 `"user.name"`）
- **返回**: 对应的状态值，如果未提供键则返回整个状态对象
- **描述**: 获取当前状态值

**示例**:
```javascript
// 获取单个状态值
const count = ctx.tui.GetState('count');

// 获取嵌套状态值
const userName = ctx.tui.GetState('user.name');

// 获取所有状态
const allState = ctx.tui.GetState();
```

#### `ctx.tui.SetState(key, value)`
- **参数**:
  - `key`: 字符串，状态键名
  - `value`: 任意类型，要设置的值
- **描述**: 设置单个状态值并触发 UI 更新

**示例**:
```javascript
// 设置简单值
ctx.tui.SetState('count', 5);

// 设置对象
ctx.tui.SetState('user', { name: 'John', age: 30 });

// 设置嵌套值（整个对象会被替换）
ctx.tui.SetState('user.profile', { city: 'New York' });
```

#### `ctx.tui.UpdateState(newStateObject)`
- **参数**:
  - `newStateObject`: 对象，包含多个状态键值对
- **描述**: 批量更新多个状态值并触发 UI 更新

**示例**:
```javascript
// 批量更新状态
ctx.tui.UpdateState({
  count: 10,
  title: 'New Title',
  user: { name: 'Jane', age: 25 }
});
```

#### `ctx.tui.ExecuteAction(actionDefinition)`
- **参数**:
  - `actionDefinition`: 对象，定义要执行的动作
- **描述**: 执行一个动作（可能是 Process 或 Script）

**示例**:
```javascript
// 执行另一个脚本
ctx.tui.ExecuteAction({
  script: 'scripts/tui/helper',
  method: 'doSomething'
});

// 执行 Process
ctx.tui.ExecuteAction({
  process: 'models.user.save',
  args: [userId, userData]
});
```

#### `ctx.tui.Refresh()`
- **描述**: 强制刷新 UI

**示例**:
```javascript
// 手动触发 UI 刷新
ctx.tui.Refresh();
```

#### `ctx.tui.Quit()`
- **描述**: 退出 TUI 应用

**示例**:
```javascript
// 退出应用
ctx.tui.Quit();
```

#### `ctx.tui.Interrupt()`
- **描述**: 中断 TUI 应用

**示例**:
```javascript
// 中断应用
ctx.tui.Interrupt();
```

#### `ctx.tui.Suspend()`
- **描述**: 暂停 TUI 应用

**示例**:
```javascript
// 暂停应用
ctx.tui.Suspend();
```

#### `ctx.tui.ClearScreen()`
- **描述**: 清除屏幕内容

**示例**:
```javascript
// 清屏
ctx.tui.ClearScreen();
```

#### `ctx.tui.EnterAltScreen()`
- **描述**: 进入备用屏幕模式

**示例**:
```javascript
// 进入备用屏幕
ctx.tui.EnterAltScreen();
```

#### `ctx.tui.ExitAltScreen()`
- **描述**: 退出备用屏幕模式

**示例**:
```javascript
// 退出备用屏幕
ctx.tui.ExitAltScreen();
```

#### `ctx.tui.ShowCursor()`
- **描述**: 显示光标

**示例**:
```javascript
// 显示光标
ctx.tui.ShowCursor();
```

#### `ctx.tui.HideCursor()`
- **描述**: 隐藏光标

**示例**:
```javascript
// 隐藏光标
ctx.tui.HideCursor();
```

#### `ctx.tui.Release()` / `ctx.tui.__release()`
- **描述**: 释放内部 Go 对象资源（通常不需要手动调用）

## 文件结构

TUI 脚本遵循以下目录结构：

```
apps/
├── tui_app/
│   ├── scripts/
│   │   └── tui/          # TUI 脚本目录
│   │       ├── counter.ts
│   │       ├── form.ts
│   │       └── utils.ts
│   └── tuis/             # TUI 配置文件
│       ├── counter.tui.yao
│       └── form.tui.yao
```

## 配置 TUI 与脚本关联

在 `.tui.yao` 配置文件中，可以通过 `bindings` 属性将按键与脚本方法关联：

```json
{
  "name": "Counter Demo",
  "data": {"count": 0},
  "layout": {
    "direction": "vertical",
    "children": [
      {
        "type": "header", 
        "props": {"title": "Counter: {{count}}"}
      },
      {
        "type": "text",
        "props": {"content": "Press '+' to increment, '-' to decrement, 'q' to quit"}
      }
    ]
  },
  "bindings": {
    "+": {"script": "scripts/tui/counter", "method": "increment"},
    "-": {"script": "scripts/tui/counter", "method": "decrement"},
    "r": {"script": "scripts/tui/counter", "method": "reset"}
  }
}
```

## 示例项目

下面是一个完整的示例，展示如何在 TUI 中使用 JS/TS 脚本：

### 1. 创建脚本文件

**`apps/tui_app/scripts/tui/todo.ts`**:

```typescript
/**
 * 添加新的待办事项
 * @param {Object} ctx - 上下文对象
 * @param {string} item - 待办事项文本
 */
function addItem(ctx, item) {
  if (!ctx || !item) {
    console.log("addItem called with invalid parameters");
    return;
  }

  // 获取当前待办事项列表
  const todos = ctx.tui.GetState("todos") || [];
  
  // 添加新项目
  todos.push({
    id: Date.now(),
    text: item,
    completed: false
  });
  
  // 更新状态
  ctx.tui.SetState("todos", todos);
  
  // 清空输入框
  ctx.tui.SetState("input", "");
}

/**
 * 切换待办事项完成状态
 * @param {Object} ctx - 上下文对象
 * @param {number} id - 待办事项 ID
 */
function toggleItem(ctx, id) {
  if (!ctx || id === undefined) {
    console.log("toggleItem called with invalid parameters");
    return;
  }

  const todos = ctx.tui.GetState("todos") || [];
  const updatedTodos = todos.map(todo => 
    todo.id === id ? { ...todo, completed: !todo.completed } : todo
  );
  
  ctx.tui.SetState("todos", updatedTodos);
}

/**
 * 删除待办事项
 * @param {Object} ctx - 上下文对象
 * @param {number} id - 待办事项 ID
 */
function deleteItem(ctx, id) {
  if (!ctx || id === undefined) {
    console.log("deleteItem called with invalid parameters");
    return;
  }

  const todos = ctx.tui.GetState("todos") || [];
  const filteredTodos = todos.filter(todo => todo.id !== id);
  
  ctx.tui.SetState("todos", filteredTodos);
}

/**
 * 清除所有已完成的待办事项
 * @param {Object} ctx - 上下文对象
 */
function clearCompleted(ctx) {
  if (!ctx) {
    console.log("clearCompleted called without TUI context");
    return;
  }

  const todos = ctx.tui.GetState("todos") || [];
  const activeTodos = todos.filter(todo => !todo.completed);
  
  ctx.tui.SetState("todos", activeTodos);
}

/**
 * 设置输入框的值
 * @param {Object} ctx - 上下文对象
 * @param {string} value - 输入值
 */
function setInput(ctx, value) {
  if (!ctx || value === undefined) {
    console.log("setInput called with invalid parameters");
    return;
  }

  ctx.tui.SetState("input", value);
}

// 导出函数以便在 TUI 配置中使用
export { addItem, toggleItem, deleteItem, clearCompleted, setInput };
```

### 2. 创建 TUI 配置文件

**`apps/tui_app/tuis/todo.tui.yao`**:

```json
{
  "name": "Todo App",
  "data": {
    "todos": [],
    "input": "",
    "filter": "all"
  },
  "layout": {
    "direction": "vertical",
    "children": [
      {
        "type": "header",
        "props": {"title": "Todo App - {{len(todos)}} items"}
      },
      {
        "type": "text",
        "props": {"content": "Enter new todo: {{input}}"}
      },
      {
        "type": "text",
        "props": {"content": "--- Todo List ---"}
      },
      {
        "type": "text",
        "props": {
          "content": "{{range todos}}{{if not .completed}}• {{.text}}{{end}}{{end}}"
        }
      },
      {
        "type": "text",
        "props": {
          "content": "{{range todos}}{{if .completed}}✓ {{.text}}{{end}}{{end}}"
        }
      },
      {
        "type": "text",
        "props": {"content": "Commands: 'a' - add, 'c' - clear completed, 'q' - quit"}
      }
    ]
  },
  "bindings": {
    "a": {"script": "scripts/tui/todo", "method": "addItem"},
    "c": {"script": "scripts/tui/todo", "method": "clearCompleted"},
    "q": {"script": "scripts/tui/todo", "method": "quit"}
  }
}
```

## 最佳实践

### 1. 错误处理
始终检查 `ctx` 对象是否可用：

```javascript
function myFunction(ctx, param) {
  if (!ctx) {
    console.error("Function called without context");
    return;
  }
  
  // 继续执行逻辑
  ctx.tui.SetState('key', param);
}
```

### 2. 状态管理
使用 `UpdateState` 批量更新相关状态，减少 UI 重绘次数：

```javascript
// 好的做法
ctx.tui.UpdateState({
  user: newUser,
  lastUpdated: Date.now(),
  status: 'updated'
});

// 避免频繁单个更新
ctx.tui.SetState('user', newUser);
ctx.tui.SetState('lastUpdated', Date.now());
ctx.tui.SetState('status', 'updated');
```

### 3. 资源清理
虽然通常不需要手动清理，但了解 `tui.Release()` 的存在是有帮助的。

### 4. 性能考虑
- 避免在脚本中执行长时间运行的操作
- 对于异步操作，考虑使用 Process 系统
- 合理使用状态更新，避免不必要的 UI 重绘

## 故障排除

### 常见问题

1. **脚本无法访问 `tui` 对象**
   - 确保脚本是通过 TUI 引擎执行的
   - 检查绑定配置是否正确

2. **状态更新不生效**
   - 确保使用了 `tui.SetState` 或 `tui.UpdateState`
   - 检查是否有 JavaScript 错误阻止了执行

3. **嵌套状态访问失败**
   - 使用点符号（如 `user.profile.name`）
   - 确保中间对象存在

### 调试技巧

```javascript
// 在脚本中添加调试信息
console.log("Current state:", ctx.tui.GetState());

// 检查特定状态值
const value = ctx.tui.GetState('myKey');
console.log("myKey value:", value);
```

## 进阶主题

### 与 Process 系统集成

脚本可以调用 Yao Process 系统来执行数据库操作、API 调用等：

```javascript
function saveUserData(ctx, userId, data) {
  // 通过 TUI 执行 Process
  ctx.tui.ExecuteAction({
    process: 'models.user.save',
    args: [userId, data]
  });
}
```

### 动态 UI 更新

利用状态绑定和表达式实现动态 UI：

```javascript
function updateStatus(ctx, status) {
  ctx.tui.SetState('status', status);
  ctx.tui.SetState('lastUpdate', new Date().toISOString());
  ctx.tui.Refresh(); // 强制刷新 UI
}
```

通过这些功能，TUI 引擎提供了强大而灵活的脚本支持，使得构建复杂的终端用户界面成为可能。