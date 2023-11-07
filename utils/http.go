package utils

import "github.com/gin-gonic/gin"

// get the origin
func GetOrigin(c *gin.Context) string {
	referer := c.Request.Referer()
	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = referer
	}
	return origin
}
