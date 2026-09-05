package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func TestAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := 1

		if uidStr := c.GetHeader("X-User-ID"); uidStr != "" {
			if parsed, err := strconv.Atoi(uidStr); err == nil && parsed > 0 {
				userID = parsed
			}
		} else if uidStr := c.Query("user_id"); uidStr != "" {
			if parsed, err := strconv.Atoi(uidStr); err == nil && parsed > 0 {
				userID = parsed
			}
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

