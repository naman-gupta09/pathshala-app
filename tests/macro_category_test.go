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

// func String(v string) *string {
// 	return &v
// }

// func ptrUint(v uint) *uint {
// 	return &v
// }

// ------------- Setup -------------

func setupMacroCategoryTestRouter() *gin.Engine {
	return gin.Default()
}

func setupMacroCategoryTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to in-memory database")
	}
	db.AutoMigrate(&models.MacroCategory{}) // Change this line to include your actual MacroCategory model
	return db
}

func mockMacroCategoryAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", float64(1)) // Mocked user ID
		c.Next()
	}
}

// ------------- Inline Controllers -------------

func addMacroCategoryWithMockDB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var macroCategory models.MacroCategory
		if err := c.ShouldBindJSON(&macroCategory); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := db.Create(&macroCategory).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "MacroCategory created successfully"})
	}
}

func getAllMacroCategoriesWithMockDB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var macroCategories []models.MacroCategory
		if err := db.Find(&macroCategories).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, macroCategories)
	}
}

func getMacroCategoryByIDWithMockDB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var macroCategory models.MacroCategory
		if err := db.First(&macroCategory, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "MacroCategory not found"})
			return
		}
		c.JSON(http.StatusOK, macroCategory)
	}
}

func updateMacroCategoryWithMockDB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var updateData models.MacroCategory
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var existing models.MacroCategory
		if err := db.First(&existing, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "MacroCategory not found"})
			return
		}
		db.Model(&existing).Updates(updateData)
		c.JSON(http.StatusOK, gin.H{"message": "MacroCategory updated successfully"})
	}
}

func deleteMacroCategoryWithMockDB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := db.Delete(&models.MacroCategory{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "MacroCategory deleted successfully"})
	}
}

// ------------- Tests -------------

func TestAddMacroCategory(t *testing.T) {
	db := setupMacroCategoryTestDB()
	router := setupMacroCategoryTestRouter()
	router.Use(mockMacroCategoryAuthMiddleware())
	router.POST("/macro_categories", addMacroCategoryWithMockDB(db))

	macroCategory := models.MacroCategory{
		Name: "Test MacroCategory",
		//UserID: 1, // Assuming you have userID field as required
		Description: "This is a test macro category",
	}

	jsonValue, _ := json.Marshal(macroCategory)
	req, _ := http.NewRequest("POST", "/macro_categories", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "MacroCategory created successfully")
}

func TestGetMacroCategoryByID(t *testing.T) {
	db := setupMacroCategoryTestDB()
	macroCategory := models.MacroCategory{
		Name: "Test MacroCategory",
		//UserID: 1,
		Description: "This is a test macro category",
	}
	db.Create(&macroCategory)

	router := setupMacroCategoryTestRouter()
	router.Use(mockMacroCategoryAuthMiddleware())
	router.GET("/macro_categories/:id", getMacroCategoryByIDWithMockDB(db))

	req, _ := http.NewRequest("GET", "/macro_categories/1", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "Test MacroCategory")
}

func TestGetAllMacroCategories(t *testing.T) {
	db := setupMacroCategoryTestDB()
	db.Create(&models.MacroCategory{
		Name: "Test MacroCategory",
		//UserID: 1,
		Description: "This is a test macro category",
	})
	router := setupMacroCategoryTestRouter()
	router.Use(mockMacroCategoryAuthMiddleware())
	router.GET("/macro_categories", getAllMacroCategoriesWithMockDB(db))

	req, _ := http.NewRequest("GET", "/macro_categories", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "Test MacroCategory")
}

func TestUpdateMacroCategory(t *testing.T) {
	db := setupMacroCategoryTestDB()
	macroCategory := models.MacroCategory{
		Name: "Old MacroCategory",
		//UserID: 1,
		Description: "This is a test macro category",
	}
	db.Create(&macroCategory)

	router := setupMacroCategoryTestRouter()
	router.Use(mockMacroCategoryAuthMiddleware())
	router.PUT("/macro_categories/:id", updateMacroCategoryWithMockDB(db))

	updatedCategory := models.MacroCategory{
		Name: "Updated MacroCategory",
	}

	jsonValue, _ := json.Marshal(updatedCategory)
	req, _ := http.NewRequest("PUT", "/macro_categories/1", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "MacroCategory updated successfully")
}

func TestDeleteMacroCategory(t *testing.T) {
	db := setupMacroCategoryTestDB()
	macroCategory := models.MacroCategory{
		Name: "Test MacroCategory",
		//UserID: 1,
		Description: "This is a test macro category",
	}
	db.Create(&macroCategory)

	router := setupMacroCategoryTestRouter()
	router.Use(mockMacroCategoryAuthMiddleware())
	router.DELETE("/macro_categories/:id", deleteMacroCategoryWithMockDB(db))

	req, _ := http.NewRequest("DELETE", "/macro_categories/1", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "MacroCategory deleted successfully")
}
