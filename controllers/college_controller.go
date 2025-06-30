package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"pathshala/models"
	"pathshala/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func CreateCollege(c *gin.Context, db *gorm.DB) {
	var input struct {
		Name            string `json:"name" binding:"required"`
		Description     string `json:"description" binding:"required"`
		State           string `json:"state" binding:"required"`
		CollegeTypeName string `json:"college_type" binding:"required"` // name instead of ID
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

	// Get CollegeType by name
	var collegeType models.CollegeType
	if err := db.Where("name = ?", input.CollegeTypeName).First(&collegeType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid college_type"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch college_type"})
		return
	}

	// Create college with temporary ActiveCandidates as 0
	college := models.College{
		Name:             input.Name,
		Description:      input.Description,
		State:            input.State,
		CollegeTypeID:    collegeType.ID,
		ActiveCandidates: 0,
	}

	if err := db.Create(&college).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create college"})
		return
	}

	// 🔄 Update ActiveCandidates based on users linked to this college
	updateActiveCandidates(db, &college)

	// Save the updated candidate count
	if err := db.Model(&college).Update("active_candidates", college.ActiveCandidates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update active candidates"})
		return
	}

	c.JSON(http.StatusCreated, college)
}

func GetColleges(c *gin.Context, db *gorm.DB) {
	// Query parameters
	column := c.Query("column")
	value := c.Query("value")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit number"})
		return
	}
	offset := (page - 1) * limit

	// Allowed filter columns
	validColumns := map[string]bool{
		"name":        true,
		"state":       true,
		"description": true,
		"type":        true,
	}

	// Prepare query
	var colleges []models.College
	query := db.Preload("CollegeType")

	if column != "" && value != "" {
		if !validColumns[column] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid column for filtering"})
			return
		}

		if column == "type" {
			query = query.Joins("JOIN college_types ON colleges.college_type_id = college_types.id").
				Where("college_types.name ILIKE ?", "%"+value+"%")
		} else {
			query = query.Where(fmt.Sprintf("colleges.%s ILIKE ?", column), "%"+value+"%")
		}
	}

	if err := query.Offset(offset).Limit(limit).Find(&colleges).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch colleges"})
		return
	}

	// Build response with updated active_candidates
	var responses []models.CollegeResponse
	for _, col := range colleges {
		var count int64
		db.Model(&models.User{}).Where("college_id = ?", col.ID).Count(&count)

		// Optional: persist the updated count to the DB
		db.Model(&models.College{}).Where("id = ?", col.ID).Update("active_candidates", count)

		responses = append(responses, models.CollegeResponse{
			ID:               col.ID,
			Name:             col.Name,
			Description:      col.Description,
			State:            col.State,
			CollegeType:      col.CollegeType.Name,
			ActiveCandidates: int(count),
		})
	}

	c.JSON(http.StatusOK, responses)
}

func UpdateCollege(c *gin.Context, db *gorm.DB) {
	// Parse college ID from URL
	idStr := c.Param("id")
	collegeID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid college ID"})
		return
	}

	// Get existing college
	var college models.College
	if err := db.First(&college, collegeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "College not found"})
		return
	}

	// Bind request body
	var input struct {
		Name             string `json:"name" binding:"required"`
		Description      string `json:"description"`
		State            string `json:"state"`
		CollegeTypeID    uint   `json:"college_type_id" binding:"required"`
		ActiveCandidates int    `json:"active_candidates"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate college_type_id exists
	var collegeType models.CollegeType
	if err := db.First(&collegeType, input.CollegeTypeID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid college_type_id"})
		return
	}

	// Update the college
	college.Name = input.Name
	college.Description = input.Description
	college.State = input.State
	college.CollegeTypeID = input.CollegeTypeID
	college.ActiveCandidates = input.ActiveCandidates

	if err := db.Save(&college).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update college"})
		return
	}

	c.JSON(http.StatusOK, college)
}

func DeleteCollege(c *gin.Context, db *gorm.DB) {
	id := c.Param("id")

	var count int64
	if err := db.Model(&models.User{}).Where("college_id = ?", id).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete college. It is referenced by existing users."})
		return
	}

	if err := db.Delete(&models.College{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "College deleted successfully"})
}

func updateActiveCandidates(db *gorm.DB, college *models.College) {
	var count int64
	db.Model(&models.User{}).Where("college_id = ?", college.ID).Count(&count)
	college.ActiveCandidates = int(count)
}
