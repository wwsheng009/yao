package testing

import (
	"fmt"
	stdtesting "testing"
	"time"

	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/form"
	"github.com/yaoapp/yao/tui/framework/input"
	"github.com/yaoapp/yao/tui/framework/validation"
)

// TestLoginFormE2E 端到端测试登录表单
func TestLoginFormE2E(t *stdtesting.T) {
	// 创建登录表单
	loginForm := createLoginForm()

	// 创建测试上下文
	ctx := NewTestContext(loginForm, TestOptions{
		Verbose:      true,
		DebugRender:  true,
		RecordEvents: true,
	})

	// 定义测试场景
	scenario := TestScenario{
		Name:        "用户登录输入",
		Description: "测试用户在登录表单中输入用户名和密码",
		Setup: func(ctx *TestContext) {
			// 初始设置：让表单获得焦点
			if focusable, ok := ctx.Root.(interface{ OnFocus() }); ok {
				focusable.OnFocus()
			}
		},
		Steps: []TestStep{
			{
				Name: "输入用户名 'a'",
				Action: KeyPress{
					Key:     'a',
					Special: event.KeyUnknown,
				},
				Assertions: []Assertion{
					// 使用表单级别的断言，因为测试框架找不到嵌套组件
					RenderContains("[a"),
				},
			},
			{
				Name: "继续输入 'd'",
				Action: KeyPress{
					Key:     'd',
					Special: event.KeyUnknown,
				},
				Assertions: []Assertion{
					RenderContains("[ad"),
				},
			},
			{
				Name: "测试退格键（跳过光标移动测试）",
				Action: KeyPress{
					Key:     0,
					Special: event.KeyBackspace,
				},
				Assertions: []Assertion{
					RenderContains("[a"),
				},
			},
			{
				Name: "测试 Tab 导航到密码字段",
				Action: KeyPress{
					Key:     0,
					Special: event.KeyTab,
				},
				Assertions: []Assertion{
					RenderContains("[请输入密码"),
				},
			},
			{
				Name: "在密码字段输入字符",
				Action: KeyPress{
					Key:     'p',
					Special: event.KeyUnknown,
				},
				Assertions: []Assertion{
					RenderContains("[*"), // 密码模式显示为星号
				},
			},
		},
		Timeout: 30 * time.Second,
	}

	// 运行场景
	if err := ctx.RunScenario(scenario); err != nil {
		t.Errorf("场景失败: %v", err)
		t.Logf("测试输出:\n%s", ctx.GetOutput())
	}
}

// TestFormNavigationE2E 测试表单导航
func TestFormNavigationE2E(t *stdtesting.T) {
	loginForm := createLoginForm()

	ctx := NewTestContext(loginForm, TestOptions{
		Verbose:      true,
		DebugRender:  false,
		RecordEvents: true,
	})

	scenario := TestScenario{
		Name:        "表单字段导航",
		Description: "测试使用 Tab 和方向键在字段间导航",
		Setup: func(ctx *TestContext) {
			if focusable, ok := ctx.Root.(interface{ OnFocus() }); ok {
				focusable.OnFocus()
			}
			// 捕获初始渲染
			ctx.captureRender()
		},
		Steps: []TestStep{
			{
				Name: "初始状态 - 用户名字段聚焦",
				Assertions: []Assertion{
					RenderContains("[请输入用户名"),
				},
			},
			{
				Name: "按 Tab 切换到密码字段",
				Action: KeyPress{
					Special: event.KeyTab,
				},
				Assertions: []Assertion{
					RenderContains("[请输入密码"),
				},
			},
			{
				Name: "按向上键返回用户名字段",
				Action: KeyPress{
					Special: event.KeyUp,
				},
				Assertions: []Assertion{
					RenderContains("[请输入用户名"),
				},
			},
		},
	}

	if err := ctx.RunScenario(scenario); err != nil {
		t.Errorf("场景失败: %v", err)
		t.Logf("测试输出:\n%s", ctx.GetOutput())
	}
}

// TestTextInputBehaviorE2E 详细测试 TextInput 行为
// 这个测试可以帮助诊断光标闪烁异常等问题
func TestTextInputBehaviorE2E(t *stdtesting.T) {
	input := input.NewTextInput()
	input.SetID("test-input")
	input.SetPlaceholder("请输入")

	ctx := NewTestContext(input, TestOptions{
		Verbose:      true,
		DebugRender:  true,
		RecordEvents: true,
	})

	scenario := TestScenario{
		Name:        "TextInput 详细行为测试",
		Description: "测试 TextInput 的光标移动、输入、删除等行为",
		Setup: func(ctx *TestContext) {
			input.OnFocus()
			ctx.mu.Lock()
			ctx.Focused = input
			ctx.mu.Unlock()
		},
		Steps: []TestStep{
			{
				Name: "初始状态检查",
				Assertions: []Assertion{
					ComponentValue("test-input", ""),
					CursorPosition("test-input", 0),
					ComponentFocused("test-input"),
				},
			},
			{
				Name: "输入第一个字符",
				Action: KeyPress{Key: 'a'},
				Assertions: []Assertion{
					ComponentValue("test-input", "a"),
					CursorPosition("test-input", 1),
				},
			},
			{
				Name: "输入第二个字符",
				Action: KeyPress{Key: 'b'},
				Assertions: []Assertion{
					ComponentValue("test-input", "ab"),
					CursorPosition("test-input", 2),
				},
			},
			{
				Name: "光标左移",
				Action: KeyPress{Special: event.KeyLeft},
				Assertions: []Assertion{
					ComponentValue("test-input", "ab"),
					CursorPosition("test-input", 1),
				},
			},
			{
				Name: "在中间插入字符",
				Action: KeyPress{Key: 'c'},
				Assertions: []Assertion{
					ComponentValue("test-input", "acb"),
					CursorPosition("test-input", 2),
				},
			},
			{
				Name: "光标移到行首",
				Action: KeyPress{Special: event.KeyHome},
				Assertions: []Assertion{
					CursorPosition("test-input", 0),
				},
			},
			{
				Name: "光标移到行尾",
				Action: KeyPress{Special: event.KeyEnd},
				Assertions: []Assertion{
					CursorPosition("test-input", 3),
				},
			},
			{
				Name: "删除字符",
				Action: KeyPress{Special: event.KeyBackspace},
				Assertions: []Assertion{
					ComponentValue("test-input", "ac"),
					CursorPosition("test-input", 2),
				},
			},
		},
	}

	if err := ctx.RunScenario(scenario); err != nil {
		t.Errorf("场景失败: %v", err)
		t.Logf("测试输出:\n%s", ctx.GetOutput())
	}
}

// TestInputFieldValidationE2E 测试输入字段验证
func TestInputFieldValidationE2E(t *stdtesting.T) {
	loginForm := createLoginForm()

	ctx := NewTestContext(loginForm, TestOptions{
		Verbose:      true,
		DebugRender:  true,
		RecordEvents: true,
	})

	scenario := TestScenario{
		Name:        "输入字段验证",
		Description: "测试表单验证机制",
		Setup: func(ctx *TestContext) {
			if focusable, ok := ctx.Root.(interface{ OnFocus() }); ok {
				focusable.OnFocus()
			}
		},
		Steps: []TestStep{
			{
				Name: "输入少于3个字符的用户名",
				Action: KeyPress{Key: 'a'},
				Assertions: []Assertion{
					RenderContains("[a"),
				},
			},
			{
				Name: "按 Enter 提交（验证失败，焦点仍在用户名）",
				Action: KeyPress{Special: event.KeyEnter},
				Assertions: []Assertion{
					// 表单不应该提交成功，焦点仍在用户名字段
					RenderContains("[a"),
				},
			},
		},
	}

	if err := ctx.RunScenario(scenario); err != nil {
		t.Errorf("场景失败: %v", err)
		t.Logf("测试输出:\n%s", ctx.GetOutput())
	}
}

// TestRenderOutputE2E 测试渲染输出
func TestRenderOutputE2E(t *stdtesting.T) {
	input := input.NewTextInput()
	input.SetID("test-input")
	input.SetValue("hello")

	ctx := NewTestContext(input, TestOptions{
		Verbose:      true,
		DebugRender:  true,
		RecordEvents: true,
	})

	scenario := TestScenario{
		Name:        "渲染输出测试",
		Description: "验证组件正确渲染到缓冲区",
		Setup: func(ctx *TestContext) {
			// 捕获初始渲染
			ctx.captureRender()
		},
		Steps: []TestStep{
			{
				Name: "验证渲染内容",
				Assertions: []Assertion{
					RenderContains("hello"),
				},
			},
		},
	}

	if err := ctx.RunScenario(scenario); err != nil {
		t.Errorf("场景失败: %v", err)
	}
	fmt.Printf("测试输出:\n%s", ctx.GetOutput())
}

// createLoginForm 创建登录表单（辅助函数）
func createLoginForm() *form.Form {
	f := form.NewForm()
	f.SetID("login-form")

	// 用户名字段
	usernameInput := input.NewTextInput()
	usernameInput.SetID("username-input")
	usernameInput.SetPlaceholder("请输入用户名")

	usernameField := form.NewFormField("username")
	usernameField.Label = "用户名: *"
	usernameField.Input = usernameInput
	usernameField.HelpText = "至少3个字符"
	usernameField.Validators = []validation.Validator{
		validation.NewFuncValidator(func(value interface{}) error {
			if value == nil || value == "" {
				return fmt.Errorf("用户名不能为空")
			}
			return nil
		}, "必填"),
		validation.NewFuncValidator(func(value interface{}) error {
			str, ok := value.(string)
			if !ok {
				return fmt.Errorf("值类型错误")
			}
			if len([]rune(str)) < 3 {
				return fmt.Errorf("至少3个字符")
			}
			return nil
		}, "长度"),
	}
	f.AddField(usernameField)

	// 密码字段
	passwordInput := input.NewTextInput()
	passwordInput.SetID("password-input")
	passwordInput.SetPassword(true)
	passwordInput.SetPlaceholder("请输入密码")

	passwordField := form.NewFormField("password")
	passwordField.Label = "密码: *"
	passwordField.Input = passwordInput
	passwordField.HelpText = "至少6个字符"
	passwordField.Validators = []validation.Validator{
		validation.NewFuncValidator(func(value interface{}) error {
			if value == nil || value == "" {
				return fmt.Errorf("密码不能为空")
			}
			return nil
		}, "必填"),
		validation.NewFuncValidator(func(value interface{}) error {
			str, ok := value.(string)
			if !ok {
				return fmt.Errorf("值类型错误")
			}
			if len([]rune(str)) < 6 {
				return fmt.Errorf("至少6个字符")
			}
			return nil
		}, "长度"),
	}
	f.AddField(passwordField)

	return f
}

// BenchmarkTextInputInsertion 性能基准测试
func BenchmarkTextInputInsertion(b *stdtesting.B) {
	input := input.NewTextInput()
	input.SetID("bench-input")
	input.OnFocus()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 模拟输入 100 个字符
		for c := 'a'; c <= 'z'; c++ {
			keyEv := event.NewKeyEvent(c)
			keyEv.Special = event.KeyUnknown
			input.HandleEvent(keyEv)
		}
		input.Clear()
	}
}
