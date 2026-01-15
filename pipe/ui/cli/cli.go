package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Cli the CLI
type Cli struct {
	option *Option
}

// In the input stream
var reader io.Reader = os.Stdin

// Option the CLI option
type Option struct {
	Label  string
	Reader io.Reader
}

// SetReader set the reader
func SetReader(r io.Reader) {
	reader = r
}

// New create a new CLI
func New(option *Option) *Cli {
	if option.Reader == nil {
		option.Reader = reader
	}
	return &Cli{
		option: option,
	}
}

// Render the CLI UI
func (cli *Cli) Render(args []any) ([]string, error) {

	in := cli.option.Reader
	if in == nil {
		in = reader
	}

	scanner := bufio.NewScanner(in)
	// bufio.Scanner 默认 token 限制是 64K，交互场景下容易被长输入击穿。
	// 这里把上限提升到 1MB，避免 Scan() 直接返回错误。
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	var lines []string
	color.Blue("%s", cli.option.Label)
	fmt.Printf("%s", color.WhiteString("> "))
	for scanner.Scan() {
		raw := scanner.Text()
		// Windows 下可能带 \r（CRLF）；这里统一去掉，避免后续解析时出现多余字符。
		line := strings.TrimSuffix(raw, "\r")
		cmd := strings.TrimSpace(line)
		if cmd == "exit()" || cmd == "exit;" || cmd == "exit" {
			break
		}
		// 忽略纯空行
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
		fmt.Printf("%s", color.WhiteString("> "))
	}

	if err := scanner.Err(); err != nil {
		// 很多时候是运行环境不支持 stdin（例如 VSCode 的 internalConsole）
		return nil, fmt.Errorf("读取输入失败：%w（请确认运行在真实终端/集成终端，或为 cli.Option.Reader 显式传入可读的输入流）", err)
	}

	return lines, nil
}
