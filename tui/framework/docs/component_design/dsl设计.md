
### 54. 完整的 DSL 定义参考 (The Complete Picture)

为了让这套系统可用，你需要定义一套清晰的 JSON/YAML 规范。

**`user_dashboard.tui.json`**:

```json
{
  "name": "User Dashboard",
  "layout": "Grid",
  "theme": "dark",
  "props": {
    "rows": [3, 0, 1], // Header, Content(Flex), Footer
    "columns": [20, 0] // Sidebar, Main
  },
  "state": {
    "currentUser": {},
    "menuItems": []
  },
  "hooks": {
    "onMount": {
      "process": "scripts.dashboard.InitData",
      "target": "state" // 结果合并到 state
    }
  },
  "children": [
    // [0,0] Header (Span 2 cols)
    {
      "type": "Box",
      "gridArea": { "r": 0, "c": 0, "w": 2, "h": 1 },
      "style": { "border": "bottom" },
      "children": [
        { "type": "Text", "value": "Yao Admin TUI", "style": "h1" }
      ]
    },
    
    // [1,0] Sidebar
    {
      "type": "List",
      "id": "sidebar_menu",
      "gridArea": { "r": 1, "c": 0 },
      "props": {
        "items": "{{menuItems}}", 
        "key": "id",
        "label": "name"
      },
      "events": {
        "onSelect": {
          "type": "Script.Run",
          "script": "scripts.dashboard.OnMenuSelect"
        }
      }
    },
    
    // [1,1] Main Content Area
    {
      "type": "Container",
      "id": "main_view",
      "gridArea": { "r": 1, "c": 1 },
      "children": [
        // 初始为空，由 Router 填充
      ]
    },
    
    // [2,0] Footer
    {
      "type": "Text",
      "gridArea": { "r": 2, "c": 0, "w": 2 },
      "value": "Ready. Press F1 for Help.",
      "style": { "fg": "gray" }
    }
  ]
}

```