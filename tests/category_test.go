package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pathshala/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ------------- Helpers -------------

func String(v string) *string {
	return &v
}

func ptrUint(v uint) *uint {
	return &v
}

// ------------- Setup -------------

func setupCategoryTestRouter() *gin.Engine {
	return gin.Default()
}

func setupCategoryTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to in-memory database")
	}
	db.AutoMigrate(&models.Category{})
	return db
}

func mockCategoryAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", float64(1)) // Mocked user ID
		c.Next()
	}
}

// ------------- Inline Controllers -------------

func addCategoryWithMockDB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var category models.Category
		if err := c.ShouldBindJSON(&category); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := db.Create(&category).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Category created successfully"})
	}
}

func getAllCategoriesWithMockDB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var categories []models.Category
		if err := db.Find(&categories).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, categories)
	}
}

func getCategoryByIDWithMockDB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var category models.Category
		if err := db.First(&category, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}
		c.JSON(http.StatusOK, category)
	}
}

func updateCategoryWithMockDB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var updateData models.Category
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var existing models.Category
		if err := db.First(&existing, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}
		db.Model(&existing).Updates(updateData)
		c.JSON(http.StatusOK, gin.H{"message": "Category updated successfully"})
	}
}

func deleteCategoryWithMockDB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := db.Delete(&models.Category{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
	}
}

// ------------- Tests -------------

func TestAddCategory(t *testing.T) {
	db := setupCategoryTestDB()
	router := setupCategoryTestRouter()
	router.Use(mockCategoryAuthMiddleware())
	router.POST("/categories", addCategoryWithMockDB(db))

	category := models.Category{
		Name:            "Test Category",
		MacroCategoryID: ptrUint(1),
		CollegeName:     String("Test College"),
		State:           String("Test State"),
		CreatorName:     "Test Creator",
		Description:     String("This is a description"),
		ImagePath:       String("path/to/image.png"),
	}

	jsonValue, _ := json.Marshal(category)
	req, _ := http.NewRequest("POST", "/categories", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "Category created successfully")
}

func TestGetCategoryByID(t *testing.T) {
	db := setupCategoryTestDB()
	category := models.Category{
		Name:            "Test Category",
		MacroCategoryID: ptrUint(1),
		CreatorName:     "Test Creator",
		Description:     String("Description"),
		ImagePath:       String("image.png"),
	}
	db.Create(&category)

	router := setupCategoryTestRouter()
	router.Use(mockCategoryAuthMiddleware())
	router.GET("/categories/:id", getCategoryByIDWithMockDB(db))

	req, _ := http.NewRequest("GET", "/categories/1", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "Test Category")
}
func TestGetAllCategories(t *testing.T) {
	db := setupCategoryTestDB()
	db.Create(&models.Category{
		Name:            "Test Category",
		MacroCategoryID: ptrUint(1),
		CreatorName:     "Test Creator",
		Description:     String("Description"),
		ImagePath:       String("image.png"),
	})
	router := setupCategoryTestRouter()
	router.Use(mockCategoryAuthMiddleware())
	router.GET("/categories", getAllCategoriesWithMockDB(db))

	req, _ := http.NewRequest("GET", "/categories", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "Test Category")
}
