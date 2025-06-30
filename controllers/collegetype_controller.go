package controllers

import (
	"net/http"
	"pathshala/models"
	"pathshala/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func CreateCollegeType(c *gin.Context, db *gorm.DB) {
	var input struct {
		Name            string `json:"name" binding:"required"`
		TypeDescription string `json:"type_description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors := utils.FormatValidationError(validationErrors)
			c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collegeType := models.CollegeType{
		Name:            input.Name,
		TypeDescription: input.TypeDescription,
	}

	if err := db.Create(&collegeType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create college type"})
		return
	}

	c.JSON(http.StatusCreated, collegeType)
}

func GetAllCollegeTypes(c *gin.Context, db *gorm.DB) {
	column := c.Query("column")
	value := c.Query("value")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Optional column validation
	validColumns := map[string]bool{
		"name":             true,
		"type_description": true,
	}

	// Validate column (only if column is provided)
	if column != "" && !validColumns[column] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid column for search"})
		return
	}

	var collegeTypes []models.CollegeType
	query := db.Model(&models.CollegeType{})

	if column != "" && value != "" {
		query = query.Where(column+" ILIKE ?", "%"+value+"%")
	}

	if err := query.Offset(offset).Limit(limit).Find(&collegeTypes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve college types"})
		return
	}

	if len(collegeTypes) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No matching results"})
		return
	}

	// c.JSON(http.StatusOK, collegeTypes)
	c.JSON(http.StatusOK, gin.H{
		"data": collegeTypes,
	})

}

// controllers/college_type_controller.go

func GetCollegeTypeByID(c *gin.Context, db *gorm.DB) {
	id := c.Param("id")
	var collegeType models.CollegeType

	if err := db.First(&collegeType, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "College type not found"})
		return
	}

	c.JSON(http.StatusOK, collegeType)
}

func UpdateCollegeType(c *gin.Context, db *gorm.DB) {
	id := c.Param("id")
	var collegeType models.CollegeType

	// Check if college type exists
	if err := db.First(&collegeType, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "College type not found"})
		return
	}

	// Bind input
	var input struct {
		Name            string `json:"name"`
		TypeDescription string `json:"type_description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	collegeType.Name = input.Name
	collegeType.TypeDescription = input.TypeDescription

	if err := db.Save(&collegeType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update college type"})
		return
	}

	c.JSON(http.StatusOK, collegeType)
}

func DeleteCollegeType(c *gin.Context, db *gorm.DB) {
	id := c.Param("id")
	var collegeType models.CollegeType

	// Check if exists
	if err := db.First(&collegeType, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "College type not found"})
		return
	}

	// Delete
	if err := db.Delete(&collegeType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete college type"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "College type deleted successfully"})
}
