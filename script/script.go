package script

import (
	"fmt"
	"strings"

	"github.com/yaoapp/gou/application"
	v8 "github.com/yaoapp/gou/runtime/v8"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/share"
)

// Load load all scripts and services
func Load(cfg config.Config) error {
	exists, err := application.App.Exists("scripts")
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	v8.CLearModules()
	messages := []string{}
	exts := []string{"*.js", "*.ts"}
	err = application.App.Walk("scripts", func(root, file string, isdir bool) error {
		if isdir {
			return nil
		}
		_, err := v8.Load(file, share.ID(root, file))
		if err != nil {
			messages = append(messages, err.Error())
			return nil
		}
		return nil
	}, exts...)

	if len(messages) > 0 {
		return fmt.Errorf("%s", strings.Join(messages, ";\n"))
	}

	// Load assistants - Move to the neo assistant package
	// err = application.App.Walk("assistants", func(root, file string, isdir bool) error {
	// 	if isdir {
	// 		return nil
	// 	}

	// 	// Keep the src.index only
	// 	if !strings.HasSuffix(file, "src/index.ts") {
	// 		return nil
	// 	}

	// 	id := fmt.Sprintf("assistants.%s", share.ID(root, file))
	// 	id = strings.TrimSuffix(id, ".src.index")
	// 	_, err := v8.Load(file, id)
	// 	return err
	// }, exts...)

	// if err != nil {
	// 	return err
	// }

	messages = []string{}
	err = application.App.Walk("services", func(root, file string, isdir bool) error {
		if isdir {
			return nil
		}
		id := fmt.Sprintf("__yao_service.%s", share.ID(root, file))
		_, err := v8.Load(file, id)
		if err != nil {
			messages = append(messages, err.Error())
			return nil
		}
		return nil
	}, exts...)

	if len(messages) > 0 {
		return fmt.Errorf("%s", strings.Join(messages, ";\n"))
	}

	return nil
}
