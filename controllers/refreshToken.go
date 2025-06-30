package controllers

import (
	"net/http"
	"pathshala/config"
	"pathshala/models"
	"pathshala/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	// Expect format: Bearer <token>
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
		return
	}

	refreshToken := tokenParts[1]

	claims, err := utils.ValidateToken(refreshToken, true)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	email := claims["email"].(string)

	// Check if refresh token is whitelisted
	storedToken, err := config.RedisClient.Get(config.Ctx, "refresh_token:"+email).Result()
	if err != nil || storedToken != refreshToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired or invalidated"})
		return
	}

	var user models.User
	config.DB.Where("email = ?", email).First(&user)
	if user.ID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	accessToken, newRefreshToken, _ := utils.GenerateTokens(user.Email, user.ID)

	_ = config.RedisClient.Set(config.Ctx, "access_token:"+accessToken, "valid", 15*time.Minute).Err()
	_ = config.RedisClient.Set(config.Ctx, "refresh_token:"+email, newRefreshToken, 7*24*time.Hour).Err()

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	})
}
