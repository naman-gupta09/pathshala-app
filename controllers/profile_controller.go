package controllers

import (
	"net/http"
	"path/filepath"
	"pathshala/config"
	"pathshala/models"
	"pathshala/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"
)

func GetProfile(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := uint(userIDInterface.(float64))

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Respond without password
	c.JSON(http.StatusOK, gin.H{
		"id":              user.ID,
		"name":            user.Name,
		"email":           user.Email,
		"secondary_email": user.SecondaryEmail,
		"profile_image":   user.Profile_image,
	})
}

func UpdateProfile(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := uint(userIDInterface.(float64))

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get form data
	newPassword := c.PostForm("new_password")
	imageFile, _ := c.FormFile("image")

	// Handle password update
	if newPassword != "" {
		hashedPwd, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hashing failed"})
			return
		}
		user.Password = string(hashedPwd)
	}

	// Handle image upload
	if imageFile != nil {
		ext := strings.ToLower(filepath.Ext(imageFile.Filename))
		allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
		if !allowed[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported image format"})
			return
		}

		// Delete old image if exists
		utils.DeleteFileIfExists(user.Profile_image)

		// Save new image
		savedPath, err := utils.SaveUploadedFile(imageFile, "uploads/profiles")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
			return
		}
		user.Profile_image = savedPath
	}

	// Save to DB
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}
