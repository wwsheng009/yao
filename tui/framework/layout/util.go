package layout

// utf8RuneCount 计算 UTF-8 字符数
func utf8RuneCount(s string) int {
	count := 0
	for range s {
		count++
	}
	return count
}
