package controllers

import (
	"net/http"
	"pathshala/config"
	"pathshala/models"
	"sync"

	"github.com/gin-gonic/gin"
	// This is required for *gorm.DB
)

func GetTeacherHomeStats(c *gin.Context) {
	role := c.GetString("role")

	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userIDFloat, ok := userIDInterface.(float64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}
	userID := uint(userIDFloat)

	if role != "teacher" && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var (
		testCount        int64
		questionCount    int64
		categoryCount    int64
		catQuestionCount int64
		studentCount     int64
	)

	// Count total tests created by teacher
	config.DB.Model(&models.Test{}).
		Where("user_id = ?", userID).
		Count(&testCount)

	// Count all test questions (across all tests created by teacher)
	config.DB.
		Table("test_questions").
		Joins("JOIN tests ON tests.id = test_questions.test_id").
		Where("tests.user_id = ?", userID).
		Count(&questionCount)

	// Count categories created by the teacher (based on creator name matching user name)
	var teacher models.User
	if err := config.DB.First(&teacher, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user info"})
		return
	}

	config.DB.Model(&models.Category{}).
		Where("creator_name = ?", teacher.Name).
		Count(&categoryCount)

	// Count questions in categories created by this teacher (assuming test linkage)
	config.DB.
		Table("test_questions").
		Joins("JOIN tests ON tests.id = test_questions.test_id").
		Where("tests.user_id = ?", userID).
		Count(&catQuestionCount)

	// Count students registered by this teacher
	config.DB.Model(&models.Student{}).
		Where("registered_by = ?", userID).
		Count(&studentCount)

	c.JSON(http.StatusOK, gin.H{
		"tests_created":           testCount,
		"questions_in_test":       questionCount,
		"categories_created":      categoryCount,
		"questions_in_categories": catQuestionCount,
		"registered_students":     studentCount,
	})
}

func GetHomeStats(c *gin.Context) {
	db := config.DB

	var (
		totalUsers              int64
		totalColleges           int64
		activeTeachers          int64
		activeStudents          int64
		totalTestQuestions      int64
		totalTestPapers         int64
		categoriesWithQuestions int64
		questionsInCategories   int64
	)

	var wg sync.WaitGroup
	wg.Add(8) // 8 goroutines

	// Users
	go func() {
		defer wg.Done()
		db.Model(&models.User{}).Count(&totalUsers)
	}()

	// Colleges
	go func() {
		defer wg.Done()
		db.Model(&models.College{}).Count(&totalColleges)
	}()

	// Active Teachers
	go func() {
		defer wg.Done()
		db.Model(&models.Teacher{}).Where("status = ?", "active").Count(&activeTeachers)
	}()

	// Active Students
	go func() {
		defer wg.Done()
		db.Model(&models.Student{}).Where("status = ?", "active").Count(&activeStudents)
	}()

	// Total Test Questions
	go func() {
		defer wg.Done()
		db.Model(&models.Question{}).Count(&totalTestQuestions)
	}()

	// Total Test Papers
	go func() {
		defer wg.Done()
		db.Model(&models.Test{}).Count(&totalTestPapers)
	}()

	// Categories with questions (excluding 'New Category')
	go func() {
		defer wg.Done()
		db.Raw(`
		SELECT COUNT(DISTINCT c.id)
		FROM categories c
		INNER JOIN questions q ON c.id = q.category_id
		WHERE c.name IS NOT NULL AND TRIM(c.name) != 'New Category'
	`).Scan(&categoriesWithQuestions)
	}()

	// Questions in categories (excluding 'New Category')
	go func() {
		defer wg.Done()
		db.Raw(`
		SELECT COUNT(q.id)
		FROM questions q
		INNER JOIN categories c ON c.id = q.category_id
		WHERE c.name IS NOT NULL AND TRIM(c.name) != 'New Category'
	`).Scan(&questionsInCategories)
	}()

	// Wait for all goroutines
	wg.Wait()

	c.JSON(http.StatusOK, gin.H{
		"total_users":               totalUsers,
		"total_colleges":            totalColleges,
		"active_teachers":           activeTeachers,
		"active_students":           activeStudents,
		"total_test_questions":      totalTestQuestions,
		"total_test_papers":         totalTestPapers,
		"categories_with_questions": categoriesWithQuestions,
		"questions_in_categories":   questionsInCategories,
	})
}
