# TUI (Terminal User Interface) 使用指南

## 概述

TUI (Terminal User Interface) 是一个强大的终端用户界面引擎，允许开发者使用 JSON/YAML 配置文件和 JavaScript/TypeScript 脚本来构建动态的终端应用程序。

## 模板语法

TUI 使用 expr-lang 作为表达式引擎，支持强大的数据绑定和表达式计算。

### 基本数据插值

使用双大括号 `{{ }}` 进行数据绑定：

```json
{
  "props": {
    "title": "欢迎, {{user.name}}!",
    "count": "总数: {{count}}"
  }
}
```

### 内置函数

TUI 提供了多种内置函数：

#### `len(array|object|string)`
获取数组、对象或字符串的长度：

```json
"content": "项目数量: {{len(items)}}"
```

#### `index(object, key)`
从对象中获取指定键的值：

```json
"content": "总计: {{index(stats, 'total')}}"
```

#### 三元运算符
支持条件表达式：

```json
"content": "{{count > 0 ? '有项目' : '无项目'}}"
```

### 表达式示例

```json
{
  "props": {
    "content": "用户: {{user.name}}, 年龄: {{user.age}}, 项目数: {{len(items)}}",
    "status": "{{count > 0 ? '活跃' : '非活跃'}}"
  }
}
```

## JavaScript/TypeScript 脚本集成

TUI 支持使用 JavaScript/TypeScript 脚本来处理复杂逻辑。

### TUI 对象 API

当脚本执行时，TUI 会将一个上下文对象作为第一个参数传递给脚本函数。该上下文对象包含一个 `tui` 子对象，提供以下方法：

**访问方式**: `ctx.tui`
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
当前 TUI 配置的名称。

#### `ctx.tui.GetState([key])`
获取状态值：
- 无参数：返回整个状态对象
- 单个参数：返回指定键的状态值
- 支持嵌套键访问（如 `"user.name"`）

#### `ctx.tui.SetState(key, value)`
设置状态值。

#### `ctx.tui.UpdateState(updates)`
批量更新状态值。

### 脚本示例

```typescript
/**
 * 增加计数器
 * @param {Object} ctx - 上下文对象
 */
function increment(ctx) {
  if (!ctx) {
    console.log("increment called without context");
    return;
  }
  
  const count = ctx.tui.GetState("count") || 0;
  ctx.tui.SetState("count", count + 1);
}

/**
 * 重置计数器
 * @param {Object} ctx - 上下文对象
 */
function reset(ctx) {
  ctx.tui.SetState("count", 0);
}
```

### 配置文件绑定

在 TUI 配置文件中绑定脚本：

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
        "props": {"content": "Press '+' to increment, '-' to decrement, 'r' to reset, 'q' to quit"}
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

## 组件系统

TUI 支持多种组件类型：

### Header 组件
```json
{
  "type": "header",
  "props": {"title": "标题内容"}
}
```

### Text 组件
```json
{
  "type": "text", 
  "props": {"content": "文本内容"}
}
```

## 快速开始

1. 创建 TUI 配置文件（`.tui.yao`）
2. 定义初始数据和布局
3. 可选：创建 JavaScript/TypeScript 脚本文件
4. 使用 `yao tui start <config-name>` 启动应用

## 错误处理

- 确保所有绑定的脚本文件存在
- 验证 JSON 配置语法正确
- 检查表达式语法是否符合 expr-lang 规范