package local

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/yaoapp/gou/application"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/sui/core"
)

// Build the template
func (tmpl *Template) Build(option *core.BuildOption) error {
	var err error

	root, err := tmpl.local.DSL.PublicRoot()
	if err != nil {
		log.Error("SyncAssets: Get the public root error: %s. use %s", err.Error(), tmpl.local.DSL.Public.Root)
		root = tmpl.local.DSL.Public.Root
	}

	if option.AssetRoot == "" {
		option.AssetRoot = filepath.Join(root, "assets")
	}

	// Sync the assets
	if err = tmpl.SyncAssets(option); err != nil {
		return err
	}

	// Build all pages
	pages, err := tmpl.Pages()
	if err != nil {
		return err
	}

	for _, page := range pages {
		perr := page.Load()
		if err != nil {
			err = multierror.Append(perr)
			continue
		}

		perr = page.Build(option)
		if perr != nil {
			err = multierror.Append(perr)
		}
	}

	return err
}

// SyncAssets sync the assets
func (tmpl *Template) SyncAssets(option *core.BuildOption) error {

	// get source abs path
	sourceRoot := filepath.Join(tmpl.local.fs.Root(), tmpl.Root, "__assets")
	if exist, _ := os.Stat(sourceRoot); exist == nil {
		return nil
	}

	//get target abs path
	root, err := tmpl.local.DSL.PublicRoot()
	if err != nil {
		log.Error("SyncAssets: Get the public root error: %s. use %s", err.Error(), tmpl.local.DSL.Public.Root)
		root = tmpl.local.DSL.Public.Root
	}
	targetRoot := filepath.Join(application.App.Root(), "public", root, "assets")

	if exist, _ := os.Stat(targetRoot); exist == nil {
		os.MkdirAll(targetRoot, os.ModePerm)
	}
	os.RemoveAll(targetRoot)

	return copyDirectory(sourceRoot, targetRoot)
}

// Build is the struct for the public
func (page *Page) Build(option *core.BuildOption) error {

	if option.AssetRoot == "" {
		option.AssetRoot = filepath.Join(page.tmpl.local.DSL.Public.Root, "assets")
	}

	html, err := page.Page.Compile(option)
	if err != nil {
		return err
	}

	// Save the html
	return page.writeHTML([]byte(html))
}

func (page *Page) publicFile() string {
	root, err := page.tmpl.local.DSL.PublicRoot()
	if err != nil {
		log.Error("publicFile: Get the public root error: %s. use %s", err.Error(), page.tmpl.local.DSL.Public.Root)
		root = page.tmpl.local.DSL.Public.Root
	}
	return filepath.Join("/", "public", root, page.Route)
}

// writeHTMLTo write the html to file
func (page *Page) writeHTML(html []byte) error {
	htmlFile := fmt.Sprintf("%s.sui", page.publicFile())
	htmlFileAbs := filepath.Join(application.App.Root(), htmlFile)
	dir := filepath.Dir(htmlFileAbs)
	if exist, _ := os.Stat(dir); exist == nil {
		os.MkdirAll(dir, os.ModePerm)
	}
	err := os.WriteFile(htmlFileAbs, html, 0644)
	if err != nil {
		return err
	}

	core.RemoveCache(htmlFile)
	log.Trace("The page %s is removed", htmlFile)
	return nil
}
