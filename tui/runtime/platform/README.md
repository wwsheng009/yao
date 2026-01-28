# Platform

平台抽象层。

## 职责

- 定义平台接口
- 屏幕管理
- 输入处理
- 光标控制
- 信号处理

## 子目录

### impl/
平台实现：

- `default/` - 默认实现（跨平台）
- `windows/` - Windows 特定实现

## 相关文件

- `platform.go` - 平台接口定义
- `screen.go` - 屏幕管理
- `input.go` - 输入处理
- `cursor.go` - 光标控制
- `terminal.go` - 终端控制
- `signal.go` - 信号处理
