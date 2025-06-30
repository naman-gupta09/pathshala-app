package controllers

import (
	"net/http"
	"pathshala/config"
	"pathshala/models"
	"strings"
	"time"

	"fmt"

	"pathshala/utils"

	"github.com/gin-gonic/gin"
)

// Get Tests

func GetTests(c *gin.Context) {
	var tests []models.Test

	// Base query with JOIN to access teacher_name
	query := config.DB.Model(&models.Test{}).
		Joins("JOIN users ON users.id = tests.user_id").
		Preload("User")

		// Apply single-column search filter
	searchColumn := c.Query("column") // e.g., "test_name", "min_questions"
	searchValue := c.Query("search")  // e.g., "Algebra", "10"
	query = utils.ApplySearchFilter(query, searchColumn, searchValue)

	// Optional: Special handling for teacher_name search
	if name := c.Query("teacher_name"); name != "" {
		query = query.Where("LOWER(teachers.name) LIKE ?", "%"+strings.ToLower(name)+"%")
	}

	// Dynamic filters
	/*filters := map[string]string{
		"test_name":     "string",
		"user_id":       "int",
		"min_questions": "int",
		"duration":      "float",
	}
	query = utils.Searchfilter(query, c.Request.URL.Query(), filters)

	// Manually handle teacher_name filter
	if name := c.Query("teacher_name"); name != "" {
		query = query.Where("users.name LIKE ?", "%"+name+"%")
	}
	*/

	// Pagination
	page := 1
	limit := 10
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	offset := (page - 1) * limit

	var total int64
	query.Count(&total)

	if err := query.Limit(limit).Offset(offset).Find(&tests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build custom response
	var formatted []gin.H
	for _, t := range tests {
		formatted = append(formatted, gin.H{
			"id":            t.ID,
			"name":          t.TestName,
			"teacher_name":  t.User.Name,
			"min_questions": t.MinQuestions,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tests":      formatted,
		"page":       page,
		"limit":      limit,
		"total":      total,
		"totalPages": (total + int64(limit) - 1) / int64(limit),
	})
}

// Inline custom struct for validation

type CreateTestInput struct {
	TestName     string `json:"test_name" binding:"required,min=3"`
	MinQuestions int    `json:"min_questions" binding:"required,gte=1"`
}

func CreateTest(c *gin.Context) {
	var input CreateTestInput

	// Validate incoming JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get teacher ID from context
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	// fmt.Println(userIDInterface)

	userIDFloat, ok := userIDInterface.(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	userID := uint(userIDFloat)

	// Create test linked to teacher in context
	test := models.Test{
		TestName:     input.TestName,
		MinQuestions: input.MinQuestions,
		UserID:       userID,
	}

	if err := config.DB.Create(&test).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create test"})
		return
	}

	// Preload teacher info
	if err := config.DB.Preload("User").First(&test, test.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch test with teacher info"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":            test.ID,
		"test_name":     test.TestName,
		"min_questions": test.MinQuestions,
		"teacher_name":  test.User.Name, // Nested reference
	})

}

// Send Test to Students based on College & State

func GetStates(c *gin.Context) {
	var states []string
	if err := config.DB.Model(&models.College{}).Distinct().Pluck("state", &states).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch states"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"states": states})
}

func GetCollegesByState(c *gin.Context) {
	state := c.Query("state")
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "State is required"})
		return
	}

	var colleges []models.College
	if err := config.DB.Where("state = ?", state).Find(&colleges).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch colleges"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"colleges": colleges})
}

// SendTest sends a test to students, but only if the requesting teacher owns it
func SendTest(c *gin.Context) {
	var request struct {
		TestID    uint   `json:"test_id" binding:"required"`
		CollegeID uint   `json:"college_id" binding:"required"`
		State     string `json:"state" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//  Get user_id (teacher) from context
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user_id not found in context"})
		return
	}

	// Convert user_id from float64 to uint
	floatID, ok := userIDInterface.(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user_id format"})
		return
	}
	teacherID := uint(floatID)

	// Check if this test belongs to the logged-in teacher
	var test models.Test
	if err := config.DB.First(&test, request.TestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test not found"})
		return
	}
	if test.UserID != teacherID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to send this test"})
		return
	}

	//  Validate college exists in state
	var college models.College
	if err := config.DB.Where("id = ? AND state = ?", request.CollegeID, request.State).First(&college).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "College not found in given state"})
		return
	}

	//  Check number of questions in the test
	var testQuestionsCount int64
	if err := config.DB.Model(&models.TestQuestion{}).Where("test_id = ?", request.TestID).Count(&testQuestionsCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count questions in test"})
		return
	}

	// If the number of questions is less than the minimum required, return an error
	if testQuestionsCount < int64(test.MinQuestions) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Test has fewer questions than the minimum required"})
		return
	}

	//  Fetch students
	var students []models.User
	if err := config.DB.Where("college_id = ? AND role=?", college.ID, "student").Find(&students).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch students"})
		return
	}

	// Assign test
	var studentTests []models.StudentTest
	for _, student := range students {
		studentTests = append(studentTests, models.StudentTest{
			StudentID: student.ID,
			TestID:    request.TestID,
			StartTime: time.Now(),
		})
	}

	if err := config.DB.Create(&studentTests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign test"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Test sent successfully",
		"test_id":    request.TestID,
		"college_id": request.CollegeID,
		"state":      request.State,
		"students":   len(students),
	})
}

// DeleteTest deletes a test, only if owned by the teacher
func DeleteTest(c *gin.Context) {
	testID := c.Param("id")

	// Get teacher ID from context
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user_id not found in context"})
		return
	}

	// Convert user_id from float64 to uint
	floatID, ok := userIDInterface.(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user_id format"})
		return
	}
	teacherID := uint(floatID)

	// Check if test belongs to this teacher
	var test models.Test
	if err := config.DB.First(&test, testID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test not found"})
		return
	}

	if test.UserID != teacherID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to delete this test"})
		return
	}

	// Proceed with deletion
	if err := config.DB.Delete(&models.Test{}, testID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete test"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test deleted successfully"})
}
