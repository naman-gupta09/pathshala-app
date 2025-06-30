package controllers

import (
	"net/http"
	"pathshala/config"
	"strconv"
	"strings"
	"time"

	"pathshala/models"

	"fmt"

	"pathshala/utils"

	"github.com/gin-gonic/gin"
)

// Response structure for enriched result output
type EnrichedResult struct {
	ID          uint   `json:"id"`
	TestID      uint   `json:"test_id"`
	UserID      uint   `json:"user_id"`
	StudentName string `json:"student_name"`
	CollegeName string `json:"college_name"`
	Branch      string `json:"branch"`
	Score       int    `json:"score"`
	Correct     int    `json:"correct"`
	Incorrect   int    `json:"incorrect"`
	Ignored     int    `json:"ignored"`
	TimeTaken   string `json:"time_taken"`
}

// Get Results with optional search filters

func GetResults(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDFloat, ok := userIDInterface.(float64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}
	userID := uint(userIDFloat)

	// test_id required
	testIDStr := c.Query("test_id")
	if testIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing test_id"})
		return
	}
	testID, _ := strconv.Atoi(testIDStr)

	// Verify teacher owns test
	var test models.Test
	if err := config.DB.First(&test, testID).Error; err != nil || test.UserID != userID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Prepare query with test ID
	query := config.DB.Model(&models.Result{}).Where("test_id = ?", testID)

	// Apply search filter
	searchColumn := c.Query("column") // e.g., score, correct
	searchValue := c.Query("search")
	query = utils.ApplySearchFilter(query, searchColumn, searchValue)

	// Execute the query
	var results []models.Result
	if err := query.Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch results"})
		return
	}

	// Enrich the result
	var enrichedResults []EnrichedResult
	for _, result := range results {
		var user models.User
		var student models.Student

		config.DB.Preload("College").First(&user, result.UserID)
		config.DB.Where("user_id = ?", result.UserID).First(&student)

		enrichedResults = append(enrichedResults, EnrichedResult{
			ID:          result.ID,
			TestID:      result.TestID,
			UserID:      result.UserID,
			StudentName: user.Name,
			CollegeName: user.College.Name,
			Branch:      student.Branch,
			Score:       result.Score,
			Correct:     result.Correct,
			Incorrect:   result.Incorrect,
			Ignored:     result.Ignored,
			TimeTaken:   result.TimeTaken,
		})
	}

	// Pagination
	page := 1
	limit := 20
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	offset := (page - 1) * limit
	total := len(enrichedResults)

	// Manual Pagination
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	paginated := enrichedResults[start:end]

	// Final JSON response
	c.JSON(http.StatusOK, gin.H{
		"results":     paginated,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + limit - 1) / limit,
	})
}

func SubmitResults(c *gin.Context) {
	type SubmitResultInput struct {
		TestID uint `json:"test_id" binding:"required"`
		UserID uint `json:"user_id" binding:"required"`
	}

	var req SubmitResultInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user
	var user models.User
	if err := config.DB.Preload("College").First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get student by user_id
	var student models.Student
	if err := config.DB.Where("user_id = ?", req.UserID).First(&student).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	// Get all answers for this test + student
	var answers []models.StudentAnswer
	if err := config.DB.Where("student_id = ? AND test_id = ?", student.ID, req.TestID).Find(&answers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch student answers"})
		return
	}

	correct, incorrect, ignored := 0, 0, 0

	for _, ans := range answers {
		var question models.Question
		if err := config.DB.Preload("Options").First(&question, ans.QuestionID).Error; err != nil {
			continue
		}

		selectedTrimmed := strings.TrimSpace(ans.Selected)
		if selectedTrimmed == "" || question.QuestionType == "Descriptive" {
			ignored++
			continue
		}

		optionIndex, err := strconv.Atoi(selectedTrimmed)
		if err != nil || optionIndex < 1 || optionIndex > len(question.Options) {
			ignored++
			continue
		}

		selectedOption := question.Options[optionIndex-1]
		if question.CorrectOptionID != nil && selectedOption.ID == *question.CorrectOptionID {
			correct++
		} else {
			incorrect++
		}
	}

	// Get test start time
	var studentTest models.StudentTest
	if err := config.DB.Where("student_id = ? AND test_id = ?", student.UserID, req.TestID).First(&studentTest).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Test assignment not found"})
		return
	}

	duration := time.Since(studentTest.StartTime)

	// Save result to DB before responding
	newResult := models.Result{
		TestID:    req.TestID,
		UserID:    req.UserID,
		Score:     correct,
		Correct:   correct,
		Incorrect: incorrect,
		Ignored:   ignored,
		TimeTaken: duration.String(),
	}

	if err := config.DB.Create(&newResult).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save result"})
		return
	}

	// Final response
	c.JSON(http.StatusOK, gin.H{
		"student_name": user.Name,
		"college_name": user.College.Name,
		"branch":       student.Branch,
		"score":        correct,
		"correct":      correct,
		"incorrect":    incorrect,
		"ignored":      ignored,
		"time_taken":   duration.String(),
	})
}

// Get Result of a Specific Student for a Specific Test
func GetStudentTestResult(c *gin.Context) {
	studentName := c.Param("student_name")
	testID := c.Param("test_id")

	var result models.Result
	if err := config.DB.Where("student_name = ? AND test_id = ?", studentName, testID).First(&result).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Result not found for the given student and test"})
		return
	}

	c.JSON(http.StatusOK, result)
}
