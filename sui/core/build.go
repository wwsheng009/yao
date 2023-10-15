package core

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Build is the struct for the public
func (page *Page) Build(option *BuildOption) (*goquery.Document, []string, error) {

	warnings := []string{}
	html, err := page.BuildHTML(option.AssetRoot)
	if err != nil {
		warnings = append(warnings, err.Error())
	}

	// Add Style & Script & Warning
	doc, err := NewDocument([]byte(html))
	if err != nil {
		warnings = append(warnings, err.Error())
	}

	// Add Style
	style, err := page.BuildStyle(option.AssetRoot)
	if err != nil {
		warnings = append(warnings, err.Error())
	}
	doc.Selection.Find("head").AppendHtml(style)

	// Add Script
	script, err := page.BuildScript(option.AssetRoot)
	if err != nil {
		warnings = append(warnings, err.Error())
	}
	doc.Selection.Find("body").AppendHtml(script)
	return doc, warnings, nil
}

// BuildHTML build the html
func (page *Page) BuildHTML(assetRoot string) (string, error) {

	html := string(page.Document)
	if page.Codes.HTML.Code != "" {
		html = strings.Replace(html, "{{ __page }}", page.Codes.HTML.Code, 1)
	}

	code := strings.ReplaceAll(html, "@assets", assetRoot)
	res, err := page.CompileHTML([]byte(code), false)
	if err != nil {
		return "", err
	}

	return string(res), nil
}

// BuildStyle build the style
func (page *Page) BuildStyle(assetRoot string) (string, error) {
	if page.Codes.CSS.Code == "" {
		return "", nil
	}

	code := strings.ReplaceAll(page.Codes.CSS.Code, "@assets", assetRoot)
	res, err := page.CompileCSS([]byte(code), false)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("<style>\n%s\n</style>\n", res), nil
}

// BuildScript build the script
func (page *Page) BuildScript(assetRoot string) (string, error) {

	if page.Codes.JS.Code == "" && page.Codes.TS.Code == "" {
		return "", nil
	}

	if page.Codes.TS.Code != "" {
		res, err := page.CompileTS([]byte(page.Codes.TS.Code), false)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("<script>\n%s\n</script>\n", res), nil
	}

	code := strings.ReplaceAll(page.Codes.JS.Code, "@assets", assetRoot)
	res, err := page.CompileJS([]byte(code), false)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("<script>\n%s\n</script>\n", res), nil
}
