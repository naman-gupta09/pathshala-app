package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"pathshala/config"
	"pathshala/models"
	"pathshala/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		claims, err := utils.ValidateToken(tokenString, false)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("email", claims["email"])
		c.Set("user_id", claims["user_id"])

		redisKey := fmt.Sprintf("user:role:%s", claims["email"])
		role, err := config.RedisClient.Get(config.Ctx, redisKey).Result()
		if err != nil {
			var user models.User
			if err := config.DB.Where("email = ?", claims["email"]).First(&user).Error; err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
				c.Abort()
				return
			}
			role = user.Role
			_ = config.RedisClient.Set(config.Ctx, redisKey, role, 15*time.Minute).Err()
		}

		c.Set("role", role)

		// fmt.Println(c.Get("role"))

		// Access token must exist in Redis
		exists, err := config.RedisClient.Exists(context.Background(), "access_token:"+tokenString).Result()
		if err != nil || exists == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token not whitelisted"})
			return
		}

		c.Next()
	}
}
