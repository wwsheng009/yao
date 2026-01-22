package main

import (
	"fmt"
	"log"

	"github.com/yaoapp/yao/tui"
)

// 示例：展示双向尺寸协商机制的自适应布局
func main() {
	fmt.Println("启动自适应布局示例...")
	
	// 创建TUI应用配置
	config := &tui.Config{
		Name: "Adaptive Layout Demo",
		Layout: tui.Layout{
			Type:      "flex",
			Direction: "column",
			Children: []tui.Layout{
				{
					Type: "text",
					Props: map[string]interface{}{
						"content":         "双向尺寸协商机制演示",
						"verticalAlign":   "center",
						"horizontalAlign": "center",
					},
					Style: map[string]interface{}{
						"height":        3,
						"backgroundColor": "#2E8B57",
						"color":         "#FFFFFF",
					},
				},
				{
					Type:      "flex",
					Direction: "row",
					Children: []tui.Layout{
						{
							Type: "table",
							Props: map[string]interface{}{
								"data": []map[string]interface{}{
									{"id": 1, "name": "张三", "email": "zhangsan@example.com"},
									{"id": 2, "name": "李四", "email": "lisi@example.com"},
									{"id": 3, "name": "王五", "email": "wangwu@example.com"},
								},
								"columns": []map[string]interface{}{
									{"key": "id", "title": "ID", "width": 10},
									{"key": "name", "title": "姓名", "width": 20},
									{"key": "email", "title": "邮箱", "width": 30},
								},
							},
							Style: map[string]interface{}{
								"width":  "flex",
								"height": "flex",
								"grow":   1,
							},
						},
						{
							Type: "viewport",
							Props: map[string]interface{}{
								"content": `这是 viewport 组件的内容区域。
它会根据可用空间自动调整大小。
支持中文内容显示。
支持ANSI颜色代码显示。`,
							},
							Style: map[string]interface{}{
								"width":  40,
								"height": "flex",
								"shrink": 0,
							},
						},
					},
				},
			},
		},
	}

	// 启动TUI应用
	app, err := tui.New(config)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}