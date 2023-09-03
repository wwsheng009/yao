package studio

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yaoapp/kun/exception"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/xun"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/helper"
)

// hdRecovered custom recovered
func hdRecovered(c *gin.Context, recovered interface{}) {

	var code = http.StatusInternalServerError

	if err, ok := recovered.(string); ok {
		c.JSON(code, xun.R{
			"code":    code,
			"message": fmt.Sprintf("%s", err),
		})
	} else if err, ok := recovered.(exception.Exception); ok {
		code = err.Code
		c.JSON(code, xun.R{
			"code":    code,
			"message": err.Message,
		})
	} else if err, ok := recovered.(*exception.Exception); ok {
		code = err.Code
		c.JSON(code, xun.R{
			"code":    code,
			"message": err.Message,
		})
	} else {
		c.JSON(code, xun.R{
			"code":    code,
			"message": fmt.Sprintf("%v", recovered),
		})
	}

	c.AbortWithStatus(code)
}

// CORS Cross-origin
func hdCORS(c *gin.Context) {
	//当前端配置withCredentials=true时, 后端配置Access-Control-Allow-Origin不能为*, 必须是相应地址
	referer := c.Request.Referer()
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

// studio API Auth
func hdAuth(c *gin.Context) {

	tokenString := c.Request.Header.Get("Authorization")

	// Get token from query
	if c.Query("studio") != "" {
		tokenString = c.Query("studio")
	}

	if strings.HasPrefix(tokenString, "Bearer") {
		tokenString = strings.TrimSpace(strings.TrimPrefix(tokenString, "Bearer "))
		if tokenString == "" {
			c.JSON(403, gin.H{"code": 403, "message": "No permission"})
			c.Abort()
			return
		}

		claims := helper.JwtValidate(tokenString, []byte(config.Conf.Studio.Secret))
		c.Set("__sid", claims.SID)
		c.Next()
		return

	} else if strings.HasPrefix(tokenString, "Signature ") { // For Yao Studio
		signature := strings.TrimSpace(strings.TrimPrefix(tokenString, "Signature "))
		nonce := c.Request.Header.Get("Studio-Nonce")
		ts := c.Request.Header.Get("Studio-Timestamp")
		query := c.Request.URL.Query()
		log.Trace("[Studio] %s, %s %s %v", signature, nonce, ts, query)
		c.Next()
		return
	}

	c.JSON(403, gin.H{"code": 403, "message": "No permission"})
	c.Abort()
	return
}
