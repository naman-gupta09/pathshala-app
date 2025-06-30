package controllers

import (
	"context"
	"net/http"
	"pathshala/config"
	"pathshala/models"
	"pathshala/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	config.DB.Where("email = ?", input.Email).First(&user)
	if user.ID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !utils.CheckPasswordHash(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	accessToken, refreshToken, _ := utils.GenerateTokens(user.Email, user.ID)

	ctx := context.Background()
	err := config.RedisClient.Set(ctx, "access_token:"+accessToken, "valid", 15*time.Minute).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store access token"})
		return
	}

	err = config.RedisClient.Set(ctx, "refresh_token:"+user.Email, refreshToken, 7*24*time.Hour).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
		return
	}

	accessToken := tokenParts[1]

	//Check if token exists in Redis (whitelisted)
	redisKey := "access_token:" + accessToken
	exists, err := config.RedisClient.Exists(config.Ctx, redisKey).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking token status"})
		return
	}
	if exists == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token already logged out or invalid"})
		return
	}

	//Delete the access token from Redis
	if err := config.RedisClient.Del(config.Ctx, redisKey).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete access token"})
		return
	}

	//Extract email from token claims to delete refresh token
	claims, err := utils.ValidateToken(accessToken, false)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	email := claims["email"].(string)

	//Delete refresh token using email key
	refreshKey := "refresh_token:" + email
	if err := config.RedisClient.Del(config.Ctx, refreshKey).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
