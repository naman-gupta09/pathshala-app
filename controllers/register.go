// controllers/auth.go
package controllers

import (
	"net/http"
	"pathshala/config"
	"pathshala/models"
	"pathshala/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	requesterRole, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input struct {
		Name            string  `json:"name" binding:"required"`
		Email           string  `json:"email" binding:"required,email"`
		Password        string  `json:"password" binding:"required"`
		ConfirmPassword string  `json:"confirm_password" binding:"required"`
		SecondaryEmail  *string `json:"secondary_email,omitempty"`
		User_role       string  `json:"user_role" binding:"required,oneof=admin teacher student"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Confirm password match
	if input.Password != input.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
		return
	}

	// Validate password format
	if err := utils.ValidatePassword(input.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if requesterRole == "teacher" && input.User_role != "student" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Teachers can only create students"})
		return
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{Name: input.Name, Email: input.Email, Password: hashedPassword, SecondaryEmail: input.SecondaryEmail, Role: input.User_role}
	result := config.DB.Create(&user)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"register_successfully": user})

}
