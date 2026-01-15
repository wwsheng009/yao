package pipe

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/gou/session"
	"github.com/yaoapp/kun/any"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/share"
	"github.com/yaoapp/yao/test"
)

func TestRunCli(t *testing.T) {
	prepare(t)
	defer test.Clean()
	translator, err := Get("cli.translator")
	if err != nil {
		t.Fatal(err)
	}

	sid := session.ID()
	context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx := translator.
		Create().
		With(context).
		WithGlobal(map[string]interface{}{"foo": "bar"}).
		WithSid(sid)
	defer Close(ctx.id)

	output, err := ctx.Exec(map[string]interface{}{"placeholder": "translate\nhello world"})
	if err != nil {
		t.Fatal(err)
	}

	res := any.Of(output).Map().MapStrAny.Dot()
	assert.True(t, res.Has("global"))
	assert.True(t, res.Has("input"))
	assert.True(t, res.Has("output"))
	assert.True(t, res.Has("sid"))
	assert.True(t, res.Has("switch"))
	assert.Equal(t, "bar", res.Get("global.foo"))
	assert.Equal(t, "translate\nhello world", res.Get("input[0].placeholder"))
	assert.Len(t, res.Get("switch"), 2)
}

func TestRunWeb(t *testing.T) {
	prepare(t)
	defer test.Clean()
	translator, err := Get("web.translator")
	if err != nil {
		t.Fatal(err)
	}

	sid := session.ID()
	context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx := translator.
		Create().
		With(context).
		WithGlobal(map[string]interface{}{"foo": "bar"}).
		WithSid(sid)
	defer Close(ctx.id)

	web := ctx.Run("hello web world")
	resume := web.(ResumeContext)
	assert.Equal(t, Input{"hello web world"}, resume.Input)

	ctx, err = Open(resume.ID)
	if err != nil {
		t.Fatal(err)
	}
	output := ctx.Resume(resume.ID, "translate", "hello web world")

	res := any.Of(output).Map().MapStrAny.Dot()
	assert.True(t, res.Has("global"))
	assert.True(t, res.Has("input"))
	assert.True(t, res.Has("output"))
	assert.True(t, res.Has("sid"))
	assert.True(t, res.Has("switch"))
	assert.Equal(t, "bar", res.Get("global.foo"))
	assert.Equal(t, "hello web world", res.Get("input[0]"))
	assert.Len(t, res.Get("switch"), 2)
}

func TestSwitchCaseGotoWithoutNodes(t *testing.T) {
	source := []byte(`{
		"name": "switch-goto-no-nodes",
		"nodes": [
			{
				"name": "user",
				"label": "Enter the command",
				"ui": "cli",
				"autofill": { "value": "{{ $in[0].placeholder }}", "action": "exit" },
				"output": { "cmd": "{{$out[0]}}", "args": "{{$out[1:]}}" }
			},
			{
				"name": "switch",
				"case": {
					"default": { "goto": "help", "input": ["{{ user }}"] }
				}
			},
			{
				"name": "help",
				"label": "Help",
				"ui": "web"
			}
		]
	}`)

	pip, err := New(source)
	if err != nil {
		t.Fatal(err)
	}

	sid := session.ID()
	ctx := pip.Create().
		With(context.Background()).
		WithSid(sid)

	ret, err := ctx.Exec(map[string]interface{}{"placeholder": "unknown"})
	if err != nil {
		defer Close(ctx.id)
		t.Fatal(err)
	}

	resume := ret.(ResumeContext)
	assert.Equal(t, "help", resume.Node.Name)
	assert.Len(t, resume.Input, 1)
	cmd := any.Of(resume.Input[0]).Map().Get("cmd")
	assert.Equal(t, "unknown", cmd)
	Close(resume.ID)
}

func prepare(t *testing.T) {
	test.Prepare(t, config.Conf)
	mirror := os.Getenv("TEST_MOAPI_MIRROR")
	secret := os.Getenv("TEST_MOAPI_SECRET")
	share.App = share.AppInfo{
		Moapi: share.Moapi{Channel: "stable", Mirrors: []string{mirror}, Secret: secret},
	}
	err := Load(config.Conf)
	if err != nil {
		t.Fatal(err)
	}
}
