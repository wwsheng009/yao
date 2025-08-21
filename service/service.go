package service

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/yaoapp/gou/api"
	"github.com/yaoapp/gou/server/http"
	"github.com/yaoapp/xun"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/neo"
	"github.com/yaoapp/yao/openapi"
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
	api.SetGuards(Guards)
	api.SetRoutes(router, "/api", cfg.AllowFrom...)
	srv := http.New(router, http.Option{
		Host:    cfg.Host,
		Port:    cfg.Port,
		Root:    "/api",
		Allows:  cfg.AllowFrom,
		Timeout: 5 * time.Second,
	})

	// Neo API
	if neo.Neo != nil {
		neo.Neo.API(router, "/api/__yao/neo")
	}
	// 增加内存分析，只能在/api请求下，要不然会拦截到前端页面请求
	if cfg.Mode == "development" {
		pprof.Register(router, "/api/__debug/pprof")

		router.GET("/api/__debug/freememory", func(ctx *gin.Context) {
			runtime.GC()
			debug.FreeOSMemory()
			ctx.JSON(200, xun.R{
				"code":    200,
				"message": fmt.Sprintf("FreeOSMemory finised:%s", time.Now().Format("2006-01-02 15:04:05")),
			})
		})
		router.GET("/api/__debug/heapdump", func(ctx *gin.Context) {
			timestamp := time.Now().Format("2006-01-02_15-04-05")
			filename := fmt.Sprintf("golang_heapdump_%s", timestamp)
			// Create a file with the new filename
			f, err := os.Create(filename)
			if err != nil {
				panic(err)
			}
			debug.WriteHeapDump(f.Fd())
			ctx.JSON(200, xun.R{
				"code":    200,
				"message": fmt.Sprintf("HeapDump File Created: %s", filename),
			})
		})

	}

	// OpenAPI Server
	if openapi.Server != nil {
		openapi.Server.Attach(router)
	}

	go func() {
		err = srv.Start()
	}()

	return srv, nil
}

// Restart the yao service
func Restart(srv *http.Server, cfg config.Config) error {
	router := gin.New()
	router.Use(Middlewares...)
	api.SetGuards(Guards)
	api.SetRoutes(router, "/api", cfg.AllowFrom...)
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
