package utils

import (
	"runtime"
	"strings"
)

func RepalcePath(s string) string {
	if runtime.GOOS != "windows" {
		return s
	}
	return strings.ReplaceAll(s, "\\", "/")
}