package controllers

import (
	"fmt"
	"net/http"
	"pathshala/config"
	"pathshala/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateMacroCategory adds a new macro category
func CreateMacroCategory(c *gin.Context) {
	var macroCategory models.MacroCategory

	if err := c.ShouldBindJSON(&macroCategory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if macroCategory.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	if err := config.DB.Create(&macroCategory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create macro category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Macro category created successfully", "data": macroCategory})
}

func GetAllMacroCategories(c *gin.Context) {
	column := c.Query("column")
	value := c.Query("value")

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var total int64
	var macroCategories []models.MacroCategory
	query := config.DB.Model(&models.MacroCategory{})

	// Search by column if provided
	if column != "" && value != "" {
		switch column {
		case "name", "description":
			query = query.Where(fmt.Sprintf("%s ILIKE ?", column), "%"+value+"%")
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid column for filtering"})
			return
		}
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count macro categories"})
		return
	}

	// Fetch paginated results
	if err := query.Offset(offset).Limit(limit).Find(&macroCategories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch macro categories"})
		return
	}

	// Respond with paginated data
	c.JSON(http.StatusOK, gin.H{
		"macro_categories": macroCategories,
		"total":            total,
		"page":             page,
		"limit":            limit,
		"total_pages":      (total + int64(limit) - 1) / int64(limit),
	})
}

// UpdateMacroCategory modifies a macro category by ID
func UpdateMacroCategory(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid macro category ID"})
		return
	}

	var macroCategory models.MacroCategory
	if err := config.DB.First(&macroCategory, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Macro category not found"})
		return
	}

	var updatedData models.MacroCategory
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	macroCategory.Name = updatedData.Name
	macroCategory.Description = updatedData.Description
	// macroCategory.UserID = updatedData.UserID

	if err := config.DB.Save(&macroCategory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update macro category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Macro category updated successfully", "data": macroCategory})
}

// DeleteMacroCategory removes a macro category by ID
func DeleteMacroCategory(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid macro category ID"})
		return
	}

	if err := config.DB.Delete(&models.MacroCategory{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete macro category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Macro category deleted successfully"})
}
