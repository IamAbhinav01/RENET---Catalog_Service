package middleware

import "github.com/gin-gonic/gin"

func TestAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", 1)
		c.Next()
	}
}