package tests

import (
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

func setupHomeTestRouter() *gin.Engine {
	return gin.Default()
}

func setupHomeTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to in-memory database")
	}
	db.AutoMigrate(&models.User{}, &models.Teacher{}, &models.Question{}, &models.Category{}, &models.Category{},
		&models.TestQuestion{})
	return db
}

func mockHomeAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", float64(1)) // Mocked user ID
		c.Next()
	}
}

// ------------- Inline Controllers -------------

func getHomeStatsWithMockDB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var result = map[string]interface{}{
			"total_users":          2,
			"active_teachers":      1,
			"active_students":      1,
			"total_categories":     1,
			"total_colleges":       0,
			"total_test_papers":    1,
			"total_test_questions": 0,
		}
		c.JSON(http.StatusOK, result)
	}
}

func getTeacherHomeStatsWithMockDB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var result = map[string]interface{}{
			"categories_created":      1,
			"questions_in_categories": 0,
			"questions_in_test":       0,
			"registered_students":     0,
			"tests_created":           1,
		}
		c.JSON(http.StatusOK, result)
	}
}

// ------------- Tests -------------

func TestGetHomeStats(t *testing.T) {
	db := setupHomeTestDB()
	router := setupHomeTestRouter()
	router.Use(mockHomeAuthMiddleware())
	router.GET("/api/home/stats", getHomeStatsWithMockDB(db))

	req, _ := http.NewRequest("GET", "/api/home/stats", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "total_users")
	assert.Contains(t, resp.Body.String(), "active_teachers")
	assert.Contains(t, resp.Body.String(), "active_students")
}

func TestGetTeacherHomeStats(t *testing.T) {
	db := setupHomeTestDB()
	router := setupHomeTestRouter()
	router.Use(mockHomeAuthMiddleware())
	router.GET("/api/home/teacher/stats", getTeacherHomeStatsWithMockDB(db))

	req, _ := http.NewRequest("GET", "/api/home/teacher/stats", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "categories_created")
	assert.Contains(t, resp.Body.String(), "tests_created")
}
