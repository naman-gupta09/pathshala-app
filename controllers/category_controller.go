package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"pathshala/config"
	"pathshala/models"
	"pathshala/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AddCategory(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - User ID not found"})
		return
	}

	// Convert userIDInterface to float64 and then to uint
	userIDFloat, ok := userIDInterface.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}
	userID := uint(userIDFloat) // Convert float64 to uint

	// Parse multipart form
	name := c.PostForm("name")
	description := c.PostForm("description")
	macroCategoryIDStr := c.PostForm("macro_category_id")

	if name == "" || macroCategoryIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name and macro_category_id are required"})
		return
	}

	// Convert macroCategoryIDStr to uint
	macroCategoryID, err := strconv.ParseUint(macroCategoryIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid macro_category_id"})
		return
	}

	// Handle file upload
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".bmp": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Allowed: jpg, jpeg, png, bmp"})
		return
	}

	// Save image using utils
	imagePath, err := utils.SaveUploadedFile(file, "uploads/categories")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	// Fetch user and related college
	var user models.User
	if err := config.DB.Preload("College").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	// Create category
	category := models.Category{
		Name:            name,
		Description:     &description,
		ImagePath:       &imagePath,
		MacroCategoryID: ptrUint(uint(macroCategoryID)), // Use the helper function to convert to *uint
		CollegeName:     &user.College.Name,
		State:           &user.College.State,
		CreatorName:     user.Name,
	}

	if err := config.DB.Create(&category).Error; err != nil {
		utils.DeleteFileIfExists(imagePath) // Clean up uploaded file on failure
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category "})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category created successfully", "category": category})
}

// Helper function to create a pointer from uint
func ptrUint(u uint) *uint {
	return &u
}

func GetAllCategories(c *gin.Context) {
	column := c.Query("column")
	value := c.Query("value")

	var categories []models.Category

	// Start DB query with preload
	query := config.DB.Preload("MacroCategory")

	// Filtering logic
	if column != "" && value != "" {
		if column == "macro_category" {
			query = query.Joins("JOIN macro_categories ON macro_categories.id = categories.macro_category_id").
				Where("macro_categories.name ILIKE ?", "%"+value+"%")
		} else {
			switch column {
			case "name", "state", "college_name", "creator_name":
				query = query.Where(fmt.Sprintf("%s ILIKE ?", column), "%"+value+"%")
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid column for filtering"})
				return
			}
		}
	}

	// Fetch all filtered categories (no DB pagination)
	if err := query.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}

	// Convert to response format
	var response []models.CategoryResponse
	for _, cat := range categories {
		resp := models.CategoryResponse{
			ID:            cat.ID,
			Name:          cat.Name,
			MacroCategory: "",
		}
		if cat.MacroCategory != nil {
			resp.MacroCategory = cat.MacroCategory.Name
		}
		if cat.CollegeName != nil {
			resp.CollegeName = *cat.CollegeName
		}
		if cat.State != nil {
			resp.State = *cat.State
		}
		response = append(response, resp)
	}

	// Paginate final slice (in memory)
	utils.PaginateSlice(c, response)
}

// Get category by ID
func GetCategoryByID(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	if err := config.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"category": category})
}

// Update an existing category
func UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	if err := config.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// Bind JSON fields for update
	name := c.PostForm("name")
	description := c.PostForm("description")
	macroCategoryIDStr := c.PostForm("macro_category_id")
	collegeName := c.PostForm("college_name")
	state := c.PostForm("state")

	if name != "" {
		category.Name = name
	}
	if description != "" {
		category.Description = &description
	}
	if collegeName != "" {
		category.CollegeName = &collegeName
	}
	if state != "" {
		category.State = &state
	}
	if macroCategoryIDStr != "" {
		macroCategoryID, err := strconv.ParseUint(macroCategoryIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid macro_category_id"})
			return
		}
		category.MacroCategoryID = ptrUint(uint(macroCategoryID)) // Use the helper function to convert to *uint
	}

	// Handle optional image update
	file, err := c.FormFile("image")
	if err == nil {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".bmp": true}
		if !allowedExts[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Allowed: jpg, jpeg, png, bmp"})
			return
		}

		// Save new image
		newImagePath, err := utils.SaveUploadedFile(file, "uploads/categories")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save new image"})
			return
		}

		// Delete old image if exists
		if category.ImagePath != nil {
			utils.DeleteFileIfExists(*category.ImagePath)
		}

		category.ImagePath = &newImagePath
	}

	if err := config.DB.Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category updated successfully", "category": category})
}

// Delete category
func DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	var category models.Category
	if err := config.DB.First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find category"})
		return
	}

	if err := config.DB.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
