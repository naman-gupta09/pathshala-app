package middlewares

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RoleMiddleware(requiredRole ...string) gin.HandlerFunc {
	return func(c *gin.Context) {

		fmt.Println(c.Get("role"))
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Check if role is allowed
		for _, allowed := range requiredRole {
			if role == allowed {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		c.Abort()

	}
}
