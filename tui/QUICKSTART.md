# TUI 快速开始指南

5 分钟快速入门 Yao TUI 引擎。

---

## 前提条件

- Go >= 1.21
- Yao 项目已安装
- 终端支持 256 色

---

## 步骤 1: 安装依赖

```bash
# 进入 tui 目录
cd tui

# 安装依赖（首次运行）
go mod download

# 验证依赖
go mod verify
```

---

## 步骤 2: 创建第一个 TUI

在项目根目录创建 `tuis/hello.tui.yao`:

```json
{
  "name": "我的第一个 TUI",
  "data": {
    "title": "Hello Yao TUI!",
    "message": "欢迎使用终端界面"
  },
  "layout": {
    "direction": "vertical",
    "children": [
      {
        "type": "header",
        "props": {
          "title": "{{title}}"
        }
      },
      {
        "type": "text",
        "props": {
          "content": "{{message}}"
        }
      }
    ]
  },
  "bindings": {
    "q": {
      "process": "tui.Quit"
    }
  }
}
```

---

## 步骤 3: 运行 TUI

```bash
# 启动 TUI
yao tui hello

# 或使用 make
make run-tui ID=hello
```

你应该看到：

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 Hello Yao TUI!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

欢迎使用终端界面

按 'q' 退出
```

---

## 步骤 4: 添加交互功能

创建 `tuis/counter.tui.yao`:

```json
{
  "name": "计数器",
  "data": {
    "count": 0
  },
  "layout": {
    "direction": "vertical",
    "children": [
      {
        "type": "header",
        "props": {
          "title": "计数器: {{count}}"
        }
      },
      {
        "type": "text",
        "props": {
          "content": "按 + 增加, 按 - 减少, 按 r 重置"
        }
      }
    ]
  },
  "bindings": {
    "+": {
      "script": "scripts/tui/counter",
      "method": "increment"
    },
    "-": {
      "script": "scripts/tui/counter",
      "method": "decrement"
    },
    "r": {
      "script": "scripts/tui/counter",
      "method": "reset"
    }
  }
}
```

创建 `scripts/tui/counter.ts`:

```typescript
function increment(tui: any) {
    const count = tui.GetState("count") || 0;
    tui.SetState("count", count + 1);
}

function decrement(tui: any) {
    const count = tui.GetState("count") || 0;
    tui.SetState("count", count - 1);
}

function reset(tui: any) {
    tui.SetState("count", 0);
}
```

运行：

```bash
yao tui counter
```

---

## 常用命令

```bash
# 运行 TUI
yao tui <id>

# 调试模式
yao tui <id> --debug

# 验证配置
yao tui validate <id>

# 列出所有 TUI
yao tui list

# 查看帮助
yao tui --help
```

---

## 下一步

1. 阅读 [架构文档](ARCHITECTURE.md) 了解设计细节
2. 查看 [TODO](TODO.md) 了解开发进度
3. 参考 [贡献指南](docs/CONTRIBUTING.md) 参与开发
4. 查看 [示例项目](examples/) 学习最佳实践

---

## 获取帮助

- GitHub Issues: https://github.com/yaoapp/yao/issues
- Discord: https://discord.gg/yao
- 文档: https://yaoapps.com/doc

祝你使用愉快！🎉
