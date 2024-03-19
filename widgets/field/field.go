package field

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/data"
)

// LoadAndExport load table
func LoadAndExport(cfg config.Config) error {
	transFilename := "model.trans.json"
	// use different transform for not mysql db
	if strings.ToLower(os.Getenv("YAO_DB_DRIVER")) != "mysql" {
		transFilename = "model-not-mysql.trans.json"
	}
	if os.Getenv("YAO_DEV") != "" {
		file := filepath.Join(os.Getenv("YAO_DEV"), "yao", "fields", transFilename)
		source, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		_, err = OpenTransform(source, "model")
		if err != nil {
			return err
		}
	}

	source, err := data.Read(filepath.Join("yao", "fields", transFilename))
	if err != nil {
		return err
	}

	_, err = OpenTransform(source, "model")
	if err != nil {
		return err
	}

	return nil
}

// SelectTransform select a transform via name
func SelectTransform(name string) (*Transform, error) {
	trans, has := Transforms[name]
	if !has {
		return nil, fmt.Errorf("Transform %s does not found", name)
	}
	return trans, nil
}

// ModelTransform select model transform via name
func ModelTransform() (*Transform, error) {
	return SelectTransform("model")
}
