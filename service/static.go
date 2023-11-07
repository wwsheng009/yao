package service

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/yaoapp/yao/data"
	"github.com/yaoapp/yao/service/fs"
	"github.com/yaoapp/yao/share"
)

// AppFileServer static file server
var AppFileServer http.Handler

// spaFileServers spa static file server
var spaFileServers map[string]http.Handler = map[string]http.Handler{}

// SpaRoots SPA static file server
var SpaRoots map[string]int = map[string]int{}

// XGenFileServerV1 XGen v1.0
var XGenFileServerV1 http.Handler = http.FileServer(data.XgenV1())

// AdminRoot cache
var AdminRoot = ""

// AdminRootLen cache
var AdminRootLen = 0

// SetupStatic setup static file server
func SetupStatic(allows ...string) error {

	// SetAdmin Root
	adminRoot()

	if isPWA() {
		AppFileServer = addCorsHeader(http.FileServer(fs.DirPWA("public")), allows...)
		return nil
	}

	for _, root := range spaApps() {
		spaFileServers[root] = addCorsHeader(http.FileServer(fs.DirPWA(filepath.Join("public", root))), allows...)
		SpaRoots[root] = len(root)
	}

	AppFileServer = addCorsHeader(http.FileServer(fs.Dir("public")), allows...)

	return nil
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

// rewrite path
func isPWA() bool {
	if share.App.Static == nil {
		return false
	}
	return share.App.Static.PWA
}

// rewrite path
func spaApps() []string {
	if share.App.Static == nil {
		return []string{}
	}
	return share.App.Static.Apps
}

// SetupAdmin setup admin static root
func adminRoot() (string, int) {
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
