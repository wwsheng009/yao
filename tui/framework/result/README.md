# Result

Result 类型实现（类似 Rust 的 Result<T, E>）。

## 职责

- 提供类型安全的错误处理
- 减少错误检查样板代码
- 链式操作支持

## 使用示例

```go
func ParseInt(s string) result.Result[int] {
    i, err := strconv.Atoi(s)
    if err != nil {
        return result.Err[int](err)
    }
    return result.Ok(i)
}

// 使用
r := ParseInt("42")
if r.IsOk() {
    val := r.Unwrap()
    fmt.Println(val)
}
```

## 相关文件

- `result.go` - Result 类型定义
- `map.go` - Map 操作
- `and_then.go` - 链式操作
