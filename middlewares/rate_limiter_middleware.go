package middlewares

import (
	"net/http"
	"pathshala/config"
	"time"

	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := "rate_limit:" + ip

		// Increment the counter
		count, err := config.RedisClient.Incr(config.Ctx, key).Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Rate limit error"})
			return
		}

		// Set expiration if it's the first time
		if count == 1 {
			config.RedisClient.Expire(config.Ctx, key, window)
		}

		// If over the limit
		if count > int64(limit) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded. Try again later."})
			return
		}

		c.Next()
	}
}
