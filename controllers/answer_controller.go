package controllers

import (
	"net/http"
	"pathshala/config"
	"pathshala/models"

	"github.com/gin-gonic/gin"
)

func SubmitAnswers(c *gin.Context) {
	var answers []models.StudentAnswer

	if err := c.ShouldBindJSON(&answers); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	for _, ans := range answers {
		if err := config.DB.Create(&ans).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save answer"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Answers saved successfully"})
}
