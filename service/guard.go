package service

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yaoapp/yao/helper"
	"github.com/yaoapp/yao/utils"

	"github.com/yaoapp/yao/widgets/chart"
	"github.com/yaoapp/yao/widgets/dashboard"
	"github.com/yaoapp/yao/widgets/form"
	"github.com/yaoapp/yao/widgets/list"
	"github.com/yaoapp/yao/widgets/table"
)

// Guards middlewares
var Guards = map[string]gin.HandlerFunc{
	"bearer-jwt":       guardBearerJWT,   // Bearer JWT
	"query-jwt":        guardQueryJWT,    // Get JWT Token from query string  "__tk"
	"cross-origin":     guardCrossOrigin, // Cross-Origin Resource Sharing
	"widget-table":     table.Guard,      // Widget Table Guard
	"widget-list":      list.Guard,       // Widget List Guard
	"widget-form":      form.Guard,       // Widget Form Guard
	"widget-chart":     chart.Guard,      // Widget Chart Guard
	"widget-dashboard": dashboard.Guard,  // Widget Dashboard Guard
}

// JWT Bearer JWT
func guardBearerJWT(c *gin.Context) {
	tokenString := c.Request.Header.Get("Authorization")
	tokenString = strings.TrimSpace(strings.TrimPrefix(tokenString, "Bearer "))
	if tokenString == "" {
		c.JSON(403, gin.H{"code": 403, "message": "No permission"})
		c.Abort()
		return
	}

	claims := helper.JwtValidate(tokenString)
	c.Set("__sid", claims.SID)
}

// JWT Bearer JWT
func guardQueryJWT(c *gin.Context) {
	tokenString := c.Query("__tk")
	if tokenString == "" {
		c.JSON(403, gin.H{"code": 403, "message": "No permission"})
		c.Abort()
		return
	}

	claims := helper.JwtValidate(tokenString)
	c.Set("__sid", claims.SID)
}

// CORS Cross Origin
func guardCrossOrigin(c *gin.Context) {
	//当前端配置withCredentials=true时, 后端配置Access-Control-Allow-Origin不能为*, 必须是相应地址
	referer := utils.GetOrigin(c) //c.Request.Referer()
	if referer != "" {
		url, _ := url.Parse(referer)
		referer = fmt.Sprintf("%s://%s", url.Scheme, url.Host)
		c.Writer.Header().Set("Access-Control-Allow-Origin", referer)
	} else {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	}
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, odata-version, odata-maxversion, mime-version")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}
	c.Next()
}
