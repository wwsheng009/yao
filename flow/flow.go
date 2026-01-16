package flow

import (
	"fmt"
	"strings"

	"github.com/yaoapp/gou/application"
	"github.com/yaoapp/gou/flow"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/share"
)

// Load 加载业务逻辑编排
func Load(cfg config.Config) error {

	// Ignore if the flows directory does not exist
	exists, err := application.App.Exists("flows")
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	messages := []string{}
	exts := []string{"*.flow.yao", "*.flow.json", "*.flow.jsonc"}
	err = application.App.Walk("flows", func(root, file string, isdir bool) error {
		if isdir {
			return nil
		}
		_, err := flow.Load(file, share.ID(root, file))
		if err != nil {
			messages = append(messages, err.Error())
		}
		return nil
	}, exts...)

	if len(messages) > 0 {
		return fmt.Errorf("%s", strings.Join(messages, ";\n"))
	}

	return err
}
