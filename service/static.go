package service

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/data"
	"github.com/yaoapp/yao/service/fs"
	"github.com/yaoapp/yao/share"
)

// AppFileServer static file server
var AppFileServer http.Handler

// XGenFileServerV1 XGen v1.0
var XGenFileServerV1 http.Handler = http.FileServer(data.XgenV1())

// BuilderFileServer Builder ui
var BuilderFileServer http.Handler = http.FileServer(data.Builder())

// AdminRoot cache
var AdminRoot = ""

// AdminRootLen cache
var AdminRootLen = 0

var rewriteRules = []rewriteRule{}

type rewriteRule struct {
	Pattern     *regexp.Regexp
	Replacement string
}

// SetupStatic setup static file server
func SetupStatic(allows ...string) error {
	setupAdminRoot()
	setupRewrite()
	AppFileServer = addCorsHeader(http.FileServer(fs.Dir("public")), allows...)
	return nil
}

func setupRewrite() {
	if share.App.Static.Rewrite != nil {
		for _, rule := range share.App.Static.Rewrite {

			pattern := ""
			replacement := ""
			for key, value := range rule {
				pattern = key
				replacement = value
				break
			}

			re, err := regexp.Compile(pattern)
			if err != nil {
				log.Error("Invalid rewrite rule: %s", pattern)
				continue
			}

			rewriteRules = append(rewriteRules, rewriteRule{
				Pattern:     re,
				Replacement: replacement,
			})
		}
	}
}

// rewrite path
func setupAdminRoot() (string, int) {
	if AdminRoot != "" {
		return AdminRoot, AdminRootLen
	}

	adminRoot := "/yao/"
	if share.App.AdminRoot != "" {
		root := strings.TrimPrefix(share.App.AdminRoot, "/")
		root = strings.TrimSuffix(root, "/")
		adminRoot = fmt.Sprintf("/%s/", root)
	}
	adminRootLen := len(adminRoot)
	AdminRoot = adminRoot
	AdminRootLen = adminRootLen
	return AdminRoot, AdminRootLen
}

// get the origin
func getOrigin(r *http.Request) string {
	referer := r.Referer()
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = referer
	}
	return origin
}

// IsAllowed check if the referer is in allow list
func IsAllowed(r *http.Request, allowsMap map[string]bool) bool {
	origin := getOrigin(r)
	if origin != "" {
		url, err := url.Parse(origin)
		if err != nil {
			return true
		}

		port := fmt.Sprintf(":%s", url.Port())
		if port == ":" || port == ":80" || port == ":443" {
			port = ""
		}
		host := fmt.Sprintf("%s%s", url.Hostname(), port)
		// fmt.Println(url, host, c.Request.Host)
		// fmt.Println(allowsMap)
		if host == r.Host {
			return true
		}
		if _, has := allowsMap[host]; !has {
			return false
		}
	}
	return true

}

func addCorsHeader(h http.Handler, allows ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 跨域访问
		if len(allows) > 0 {
			allowsMap := map[string]bool{}
			for _, allow := range allows {
				allowsMap[allow] = true
			}
			origin := getOrigin(r)
			if origin != "" {
				if IsAllowed(r, allowsMap) {
					// url parse
					url, _ := url.Parse(origin)
					origin = fmt.Sprintf("%s://%s", url.Scheme, url.Host)
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}
		}
		h.ServeHTTP(w, r)
	}
}
