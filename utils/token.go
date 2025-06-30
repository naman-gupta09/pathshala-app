// utils/token.go
package utils

import (
	"errors"
	"fmt"
	"os"
	"pathshala/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))
var refreshSecret = []byte(os.Getenv("JWT_REFRESH_SECRET"))

// Generate access and refresh tokens
func GenerateTokens(email string, id uint) (string, string, error) {
	// Access Token (15 minutes)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":   email,
		"user_id": id,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	})
	accessTokenString, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}

	// Refresh Token (7 days)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":   email,
		"user_id": id,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	refreshTokenString, err := refreshToken.SignedString(refreshSecret)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// Validate Access Token
func ValidateToken(tokenString string, isRefresh bool) (map[string]interface{}, error) {

	//Check Redis blacklist first
	exists, err := config.RedisClient.Exists(config.Ctx, tokenString).Result()
	fmt.Println(exists)
	if err != nil {
		return nil, errors.New("error checking token blacklist")
	}
	if exists == 1 {
		fmt.Println("I am here")
		return nil, errors.New("token is blacklisted (logged out)")
	}

	secret := jwtSecret
	if isRefresh {
		secret = refreshSecret
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
