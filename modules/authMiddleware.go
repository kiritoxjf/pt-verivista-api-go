package modules

import (
	"github.com/gin-gonic/gin"
	"log"
)

// AuthMiddleware 身份验证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Request.Cookie("VeriVista-token")
		log.Fatal(token)
		//if(validateCredential(token))
	}
}

// validateCredential token验证
func validateCredential(token string) bool {
	return false
}
