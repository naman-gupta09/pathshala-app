package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"pathshala/config"
	"pathshala/models"
	"pathshala/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AddQuestionToTestRequest struct {
	QuestionIDs []uint `json:"question_ids" binding:"required,min=1,dive,required"`
}

// 1. Add existing question to test
func AddExistingQuestionToTest(c *gin.Context) {
	testIDStr := c.Param("test_id")
	testID, err := strconv.Atoi(testIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test ID"})
		return
	}

	// Authorization check
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	userIDFloat, ok := userIDVal.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID type"})
		return
	}
	userID := uint(userIDFloat)
	role := c.GetString("role")

	if err := utils.AuthorizeTestAccess(uint(testID), userID, role); err != nil {
		if errors.Is(err, utils.ErrTestNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Test not found"})
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to modify this test"})
		}
		return
	}

	var req AddQuestionToTestRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	var added []uint
	var alreadyPresent []uint
	var invalidIDs []uint

	for _, questionID := range req.QuestionIDs {
		// Check if question exists
		var question models.Question
		if err := config.DB.First(&question, questionID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				invalidIDs = append(invalidIDs, questionID)
				continue
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify question existence"})
			return
		}
		// Check if already added
		var existing models.TestQuestion
		err := config.DB.
			Where("test_id = ? AND question_id = ?", testID, questionID).
			First(&existing).Error

		if err == nil {
			alreadyPresent = append(alreadyPresent, questionID)
			continue
		} else if err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking existing questions"})
			return
		}

		// Add to test
		entry := models.TestQuestion{
			TestID:     uint(testID),
			QuestionID: questionID,
		}
		if err := config.DB.Create(&entry).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add question to test"})
			return
		}
		added = append(added, questionID)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":              "Processed question list",
		"added":                added,
		"already_present":      alreadyPresent,
		"invalid_question_ids": invalidIDs,
	})
}

// 2. Add New question to test
func AddNewQuestionToTest(c *gin.Context) {
	testIDStr := c.Param("test_id")
	testID, err := strconv.Atoi(testIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test ID"})
		return
	}

	// Authorization check
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	userIDFloat, ok := userIDVal.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID type"})
		return
	}
	userID := uint(userIDFloat)
	role := c.GetString("role")
	if err := utils.AuthorizeTestAccess(uint(testID), userID, role); err != nil {
		if errors.Is(err, utils.ErrTestNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Test not found"})
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to modify this test"})
		}
		return
	}

	questionType := strings.ToUpper(c.PostForm("question_type"))

	switch questionType {
	case "MCQ":
		question, err := handleMCQQuestion(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save question"})
		}
		if err := LinkQuestionToTest(uint(testID), question.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link question with test"})
		}
	case "TRUE_FALSE":
		question, err := handleTrueFalseQuestion(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save question"})
		}
		if err := LinkQuestionToTest(uint(testID), question.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link question with test"})
		}
	case "DESCRIPTIVE":
		question, err := handleDescriptiveQuestion(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save question"})
		}
		if err := LinkQuestionToTest(uint(testID), question.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link question with test"})
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question_type. Must be MCQ, TRUE_FALSE, or DESCRIPTIVE"})
	}
}

func LinkQuestionToTest(testID, questionID uint) error {
	return config.DB.Create(&models.TestQuestion{
		TestID:     testID,
		QuestionID: questionID,
	}).Error
}

// 3. Delete question from a test
func DeleteTestQuestion(c *gin.Context) {
	testIDStr := c.Param("test_id")
	questionIDStr := c.Param("question_id")

	// Convert params to uint
	testID, err := strconv.ParseUint(testIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test_id"})
		return
	}

	// Authorization check
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	userIDFloat, ok := userIDVal.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID type"})
		return
	}
	userID := uint(userIDFloat)
	role := c.GetString("role")
	if err := utils.AuthorizeTestAccess(uint(testID), userID, role); err != nil {
		if errors.Is(err, utils.ErrTestNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Test not found"})
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to modify this test"})
		}
		return
	}

	questionID, err := strconv.ParseUint(questionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question_id"})
		return
	}

	// Try to find the test-question linkage
	var testQuestion models.TestQuestion
	if err := config.DB.Where("test_id = ? AND question_id = ?", testID, questionID).First(&testQuestion).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found in this test"})
		return
	}

	// Delete the linkage, not the question itself
	if err := config.DB.Delete(&testQuestion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete test-question link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Question removed from the test successfully"})
}

// 4. Get Test Questions
func GetTestQuestions(c *gin.Context) {
	testIDStr := c.Param("test_id")
	testID, err := strconv.ParseUint(testIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test_id"})
		return
	}

	// Authorization check
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	userIDFloat, ok := userIDVal.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID type"})
		return
	}
	userID := uint(userIDFloat)
	role := c.GetString("role")
	if err := utils.AuthorizeTestAccess(uint(testID), userID, role); err != nil {
		if errors.Is(err, utils.ErrTestNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Test not found"})
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to modify this test"})
		}
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	column := c.Query("column")
	value := c.Query("value")

	var questions []struct {
		ID           uint   `json:"id"`
		QuestionText string `json:"question_text"`
	}
	var totalCount int64

	// Redis caching for first page without filters
	if column == "" && value == "" && page == 1 {
		cacheKey := fmt.Sprintf("test_questions:test_id:%d:page:%d:limit:%d", testID, page, limit)
		cachedData, err := config.RedisClient.Get(config.Ctx, cacheKey).Result()
		if err == nil {
			log.Println("Cache hit for test questions")
			var cachedResponse map[string]interface{}
			if err := json.Unmarshal([]byte(cachedData), &cachedResponse); err == nil {
				c.JSON(http.StatusOK, cachedResponse)
				return
			} else {
				log.Printf("Failed to unmarshal cached data: %v", err)
			}
		}
	}

	query := config.DB.Table("questions").
		Select("questions.id, questions.question_text").
		Joins("LEFT JOIN test_questions ON questions.id = test_questions.question_id").
		Where("test_questions.test_id = ?", testID)

	countQuery := config.DB.Table("questions").
		Joins("LEFT JOIN test_questions ON questions.id = test_questions.question_id").
		Where("test_questions.test_id = ?", testID)

	switch strings.ToLower(column) {
	case "question":
		likeClause := "%" + strings.ToLower(value) + "%"
		query = query.Where("LOWER(questions.question_text) LIKE ?", likeClause)
		countQuery = countQuery.Where("LOWER(questions.question_text) LIKE ?", likeClause)
	case "":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid column. Only 'question' is supported for test-based search"})
		return
	}

	// Count total filtered results
	countQuery.Count(&totalCount)

	// Fetch paginated filtered results
	if err := query.Offset(offset).Limit(limit).Scan(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search questions"})
		return
	}

	responseData := gin.H{
		"questions":   questions,
		"page":        page,
		"limit":       limit,
		"total_count": totalCount,
		"total_pages": (totalCount + int64(limit) - 1) / int64(limit),
	}

	// Cache only if first page with no filters
	if column == "" && value == "" && page == 1 {
		cacheData, _ := json.Marshal(responseData)
		cacheKey := fmt.Sprintf("test_questions:test_id:%d:page:%d:limit:%d", testID, page, limit)
		err := config.RedisClient.Set(config.Ctx, cacheKey, cacheData, 10*time.Minute).Err()
		if err != nil {
			log.Printf("Failed to cache test questions: %v", err)
		} else {
			log.Println("Cached test questions for initial page without filters")
		}
	}

	c.JSON(http.StatusOK, responseData)
}

// 5. Edit Test Questions
func EditQuestionOfTest(c *gin.Context) {
	testIDStr := c.Param("test_id")
	questionIDStr := c.Param("question_id")

	testID, err := strconv.Atoi(testIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test ID"})
		return
	}

	// Authorization check
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}
	userIDFloat, ok := userIDVal.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID type"})
		return
	}
	userID := uint(userIDFloat)
	role := c.GetString("role")
	if err := utils.AuthorizeTestAccess(uint(testID), userID, role); err != nil {
		if errors.Is(err, utils.ErrTestNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Test not found"})
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to modify this test"})
		}
		return
	}

	questionID, err := strconv.Atoi(questionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}

	// Check if the question belongs to the test
	var testQuestion models.TestQuestion
	if err := config.DB.Where("test_id = ? AND question_id = ?", testID, questionID).First(&testQuestion).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Question not linked to the given test"})
		return
	}

	questionType := strings.ToUpper(c.PostForm("question_type"))

	switch questionType {
	case "MCQ":
		_, err := handleEditMCQ(c, questionIDStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update MCQ question"})
		}
	case "TRUE_FALSE":
		_, err := handleEditTrueFalse(c, questionIDStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update True/False question"})
		}
	case "DESCRIPTIVE":
		_, err := handleEditDescriptive(c, questionIDStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Descriptive question"})
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question_type. Must be MCQ, TRUE_FALSE, or DESCRIPTIVE"})
	}
}
