package service

import (
	"time"

	nethttp "net/http"

	"github.com/gin-gonic/gin"
	"github.com/yaoapp/gou/api"
	"github.com/yaoapp/gou/server/http"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/openapi"
	"github.com/yaoapp/yao/service/fs"
	"github.com/yaoapp/yao/share"
)

// Start the yao service
func Start(cfg config.Config) (*http.Server, error) {

	if cfg.AllowFrom == nil {
		cfg.AllowFrom = []string{}
	}

	err := prepare(cfg.AllowFrom...)
	if err != nil {
		return nil, err
	}

	router := gin.New()
	
	router.Use(Middlewares...)

	var apiRoot string
	if openapi.Server != nil {
		// OpenAPI mode: use OAuth guards and dynamic routing
		apiRoot = openapi.Server.Config.BaseURL
		api.SetGuards(OpenAPIGuards())

		// Developer APIs: use dynamic proxy (supports hot-reload)
		router.Any(apiRoot+"/api/*path", DynamicAPIHandler)

		// Widgets and system APIs: static registration
		api.SetRoutes(router, apiRoot, cfg.AllowFrom...)

		// Build route table for dynamic lookup
		api.BuildRouteTable()

		// Attach OpenAPI built-in features
		openapi.Server.Attach(router)
	} else {
		// Traditional mode: unchanged
		apiRoot = "/api"
		api.SetGuards(Guards)
		api.SetRoutes(router, "/api", cfg.AllowFrom...)
	}
	router.NoRoute(func(c *gin.Context) {
		staticDir := fs.Dir("public") // 获取 Yao 文件系统实例
		files := []string{"/404.html", "/notFound.html"}
		for _, file := range files {
			if f, err := staticDir.Open(file); err == nil {
				defer f.Close()
				
				// 利用 http.ServeContent 直接渲染 http.File
				// 这会自动处理 Content-Type、修改时间以及可能的 Range 请求
				stat, err := f.Stat()
				if err == nil {
					c.Header("Content-Type", "text/html; charset=utf-8")
					nethttp.ServeContent(c.Writer, c.Request, stat.Name(), stat.ModTime(), f)
					return
				}
			}
		}
		// 兜底逻辑：如果 public 下没有 404 相关 HTML，则返回 JSON
		c.JSON(nethttp.StatusNotFound, gin.H{"code": 404, "message": "Resource Not Found"})
	})

	srv := http.New(router, http.Option{
		Host:    cfg.Host,
		Port:    cfg.Port,
		Root:    apiRoot,
		Allows:  cfg.AllowFrom,
		Timeout: 5 * time.Second,
	})

	go func() {
		err = srv.Start()
	}()

	return srv, nil
}

// Restart the yao service
func Restart(srv *http.Server, cfg config.Config) error {
	router := gin.New()
	router.Use(Middlewares...)

	if openapi.Server != nil {
		// OpenAPI mode
		baseURL := openapi.Server.Config.BaseURL
		api.SetGuards(OpenAPIGuards())
		router.Any(baseURL+"/api/*path", DynamicAPIHandler)
		api.SetRoutes(router, baseURL, cfg.AllowFrom...)
		api.BuildRouteTable()
		openapi.Server.Attach(router)
	} else {
		// Traditional mode: unchanged
		api.SetGuards(Guards)
		api.SetRoutes(router, "/api", cfg.AllowFrom...)
	}

	srv.Reset(router)
	return srv.Restart()
}

// Stop the yao service
func Stop(srv *http.Server) error {
	err := srv.Stop()
	if err != nil {
		return err
	}
	<-srv.Event()
	return nil
}

func prepare(allows ...string) error {

	// Session server
	err := share.SessionStart()
	if err != nil {
		return err
	}

	err = SetupStatic(allows...)
	if err != nil {
		return err
	}

	return nil
}
