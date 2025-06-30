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

// ----------------- Helpers -----------------

func setupTeacherHomeTestRouter() *gin.Engine {
	return gin.Default()
}

func setupTeacherHomeTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to in-memory database")
	}
	db.AutoMigrate(
		&models.Teacher{},
		&models.Student{},
		&models.Test{},
		&models.TestQuestion{},
		&models.Category{},
	)
	return db
}

func mockTeacherAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", float64(1)) // Mocked user ID
		c.Set("role", "teacher")
		c.Next()
	}
}

// ----------------- Inline Controller -----------------

func getTeacherHomeStatsInline(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := uint(c.GetFloat64("user_id"))

		var createdTests int64
		db.Model(&models.Test{}).Where("user_id = ?", userID).Count(&createdTests)

		var createdQuestions int64
		db.Model(&models.TestQuestion{}).
			Joins("JOIN tests ON tests.id = test_questions.test_id").
			Where("tests.user_id = ?", userID).
			Count(&createdQuestions)

		var totalCategories int64
		db.Model(&models.Category{}).Where("creator_name = ?", "Teacher 1").Count(&totalCategories)

		var totalStudents int64
		db.Model(&models.Student{}).Count(&totalStudents)

		c.JSON(http.StatusOK, gin.H{
			"created_tests":     createdTests,
			"created_questions": createdQuestions,
			"total_categories":  totalCategories,
			"total_students":    totalStudents,
		})
	}
}

func TestGetTeacherHomeStat(t *testing.T) {
	db := setupTeacherHomeTestDB()

	// Seed mock data
	db.Create(&models.Teacher{
		UserID:      1,
		State:       "Delhi",
		TeacherType: "full-time",
		Super:       false,
	})

	db.Create(&models.Test{
		TestName:     "Sample Test",
		UserID:       1, // Matches mocked user_id from middleware
		MinQuestions: 5,
	})

	db.Create(&models.TestQuestion{
		TestID:     1,
		QuestionID: 1,
	})

	db.Create(&models.Category{
		Name:        "Sample Category",
		CreatorName: "Teacher 1",
		//MacroCategoryID: 1, // Optional, based on your schema
	})

	db.Create(&models.Student{
		UserID: 2,
		Status: "active",
		Branch: "CSE",
		Gender: "male",
	})

	router := setupTeacherHomeTestRouter()
	router.Use(mockTeacherAuthMiddleware())
	router.GET("/teacher/home", getTeacherHomeStatsInline(db))

	req, _ := http.NewRequest("GET", "/teacher/home", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "created_tests")
	assert.Contains(t, resp.Body.String(), "created_questions")
	assert.Contains(t, resp.Body.String(), "total_categories")
	assert.Contains(t, resp.Body.String(), "total_students")
}
