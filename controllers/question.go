package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pathshala/config"
	"pathshala/models"
	"pathshala/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BaseQuestionRequest struct {
	QuestionText string `form:"question_text" binding:"required"`
	Difficulty   string `form:"difficulty" binding:"required,oneof=easy medium hard"`
	CategoryID   uint   `form:"category_id" binding:"required"`
	Image1Time   *int   `form:"image1_display_time"`
	Image2Time   *int   `form:"image2_display_time"`
	Comment      string `form:"comment"`
	CommentTime  *int   `form:"comment_display_time"`
}

type MCQRequest struct {
	BaseQuestionRequest
	CorrectOptionID uint   `form:"correct_option_id" binding:"required,oneof=1 2 3 4 5"`
	QuestionType    string `form:"question_type" binding:"required,eq=MCQ"`
}

type TrueFalseRequest struct {
	BaseQuestionRequest
	CorrectOptionID uint   `form:"correct_option_id" binding:"required,oneof=1 2"`
	QuestionType    string `form:"question_type" binding:"required,eq=TRUE_FALSE"`
}

type DescriptiveRequest struct {
	BaseQuestionRequest
	QuestionType      string `form:"question_type" binding:"required,eq=DESCRIPTIVE"`
	DescriptiveAnswer string `form:"descriptive_answer" binding:"required"`
}

// 1. Add Questions
func AddQuestion(c *gin.Context) {
	questionType := strings.ToUpper(c.PostForm("question_type"))

	switch questionType {
	case "MCQ":
		handleMCQQuestion(c)
	case "TRUE_FALSE":
		handleTrueFalseQuestion(c)
	case "DESCRIPTIVE":
		handleDescriptiveQuestion(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question_type. Must be MCQ, TRUE_FALSE, or DESCRIPTIVE"})
	}
}

// MCQ
func handleMCQQuestion(c *gin.Context) (*models.Question, error) {
	var req MCQRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return nil, errors.New("something went wrong")
	}

	image1Path, err := handleOptionalImage(c, "image1", req.Image1Time)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, errors.New("something went wrong")
	}

	image2Path, err := handleOptionalImage(c, "image2", req.Image2Time)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, errors.New("something went wrong")
	}

	if req.Comment != "" && req.CommentTime == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment_display_time required if comment is provided"})
		return nil, errors.New("something went wrong")
	}

	if req.Comment == "" && req.CommentTime != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment required if comment_display_time is provided"})
		return nil, errors.New("something went wrong")
	}

	question := models.Question{
		QuestionText:       req.QuestionText,
		QuestionType:       "MCQ",
		Difficulty:         req.Difficulty,
		CategoryID:         req.CategoryID,
		CorrectOptionID:    &req.CorrectOptionID,
		Image1:             image1Path,
		Image1DisplayTime:  req.Image1Time,
		Image2:             image2Path,
		Image2DisplayTime:  req.Image2Time,
		Comment:            req.Comment,
		CommentDisplayTime: req.CommentTime,
	}

	if err := config.DB.Create(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save question"})
		return nil, errors.New("something went wrong")
	}

	for i := 1; i <= 5; i++ {
		optionText := c.PostForm(fmt.Sprintf("option_%d", i))
		if optionText != "" {
			isCorrect := uint(i) == req.CorrectOptionID
			option := models.QuestionOption{
				QuestionID: question.ID,
				OptionID:   uint(i),
				OptionText: optionText,
				IsCorrect:  isCorrect,
			}
			config.DB.Create(&option)
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "MCQ question added successfully"})

	return &question, nil
}

// True/False
func handleTrueFalseQuestion(c *gin.Context) (*models.Question, error) {
	var req TrueFalseRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return nil, errors.New("something went wrong")
	}

	image1Path, err := handleOptionalImage(c, "image1", req.Image1Time)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, errors.New("something went wrong")
	}

	image2Path, err := handleOptionalImage(c, "image2", req.Image2Time)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, errors.New("something went wrong")
	}

	if req.Comment != "" && req.CommentTime == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment_display_time required if comment is provided"})
		return nil, errors.New("something went wrong")
	}

	if req.Comment == "" && req.CommentTime != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment required if comment_display_time is provided"})
		return nil, errors.New("something went wrong")
	}

	question := models.Question{
		QuestionText:       req.QuestionText,
		QuestionType:       "TRUE_FALSE",
		Difficulty:         req.Difficulty,
		CategoryID:         req.CategoryID,
		CorrectOptionID:    &req.CorrectOptionID,
		Image1:             image1Path,
		Image1DisplayTime:  req.Image1Time,
		Image2:             image2Path,
		Image2DisplayTime:  req.Image2Time,
		Comment:            req.Comment,
		CommentDisplayTime: req.CommentTime,
	}

	if err := config.DB.Create(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save question"})
		return nil, errors.New("something went wrong")
	}

	options := []string{"True", "False"}
	for i, text := range options {
		isCorrect := uint(i+1) == req.CorrectOptionID
		option := models.QuestionOption{
			QuestionID: question.ID,
			OptionID:   uint(i + 1),
			OptionText: text,
			IsCorrect:  isCorrect,
		}
		config.DB.Create(&option)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "True/False question added successfully"})

	return &question, nil
}

// Descriptive
func handleDescriptiveQuestion(c *gin.Context) (*models.Question, error) {
	var req DescriptiveRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return nil, errors.New("something went wrong")
	}

	image1Path, err := handleOptionalImage(c, "image1", req.Image1Time)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, errors.New("something went wrong")
	}

	image2Path, err := handleOptionalImage(c, "image2", req.Image2Time)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, errors.New("something went wrong")
	}

	if req.Comment != "" && req.CommentTime == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment_display_time required if comment is provided"})
		return nil, errors.New("something went wrong")
	}

	if req.Comment == "" && req.CommentTime != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment required if comment_display_time is provided"})
		return nil, errors.New("something went wrong")
	}

	question := models.Question{
		QuestionText:       req.QuestionText,
		QuestionType:       "DESCRIPTIVE",
		Difficulty:         req.Difficulty,
		CategoryID:         req.CategoryID,
		Image1:             image1Path,
		Image1DisplayTime:  req.Image1Time,
		Image2:             image2Path,
		Image2DisplayTime:  req.Image2Time,
		Comment:            req.Comment,
		CommentDisplayTime: req.CommentTime,
	}

	if err := config.DB.Create(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save question"})
		return nil, errors.New("something went wrong")
	}

	// Save descriptive answer in question_options
	descriptiveOption := models.QuestionOption{
		QuestionID: question.ID,
		OptionID:   1, // Default value for descriptive
		OptionText: req.DescriptiveAnswer,
		IsCorrect:  true,
	}

	// log.Println("Descriptive Answer:", req.DescriptiveAnswer)
	if err := config.DB.Create(&descriptiveOption).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save descriptive answer"})
		return nil, errors.New("something went wrong")
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Descriptive question added successfully"})

	return &question, nil
}

func handleOptionalImage(c *gin.Context, field string, displayTime *int) (string, error) {
	file, err := c.FormFile(field)
	if err != nil {
		if displayTime != nil {
			return "", fmt.Errorf("%s is required if %s_display_time is provided", field, field)
		}
		return "", nil // not provided
	}

	if displayTime == nil {
		return "", fmt.Errorf("%s_display_time is required if %s is provided", field, field)
	}

	// validate image size
	const maxSizeKB = 100
	if file.Size > maxSizeKB*1024 {
		return "", fmt.Errorf("%s must be less than %dKB", field, maxSizeKB)
	}

	uploadPath := "uploads/questions/"
	savedPath, err := utils.SaveUploadedFile(file, uploadPath)
	if err != nil {
		return "", fmt.Errorf("failed to save %s: %v", field, err)
	}

	return savedPath, nil
}

// 2. Edit Questions
func EditQuestion(c *gin.Context) {
	questionType := strings.ToUpper(c.PostForm("question_type"))
	questionID := c.Param("id")

	switch questionType {
	case "MCQ":
		handleEditMCQ(c, questionID)
	case "TRUE_FALSE":
		handleEditTrueFalse(c, questionID)
	case "DESCRIPTIVE":
		handleEditDescriptive(c, questionID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question_type. Must be MCQ, TRUE_FALSE, or DESCRIPTIVE"})
	}
}

func handleEditMCQ(c *gin.Context, questionID string) (*models.Question, error) {
	var req MCQRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return nil, errors.New("something went wrong")
	}

	var question models.Question
	if err := config.DB.First(&question, questionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return nil, errors.New("something went wrong")
	}

	image1Path, err := handleOptionalImage(c, "image1", req.Image1Time)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, errors.New("something went wrong")
	}

	image2Path, err := handleOptionalImage(c, "image2", req.Image2Time)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, errors.New("something went wrong")
	}

	if req.Comment != "" && req.CommentTime == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment_display_time required if comment is provided"})
		return nil, errors.New("something went wrong")
	}

	if req.Comment == "" && req.CommentTime != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment required if comment_display_time is provided"})
		return nil, errors.New("something went wrong")
	}

	// Update fields
	question.QuestionText = req.QuestionText
	question.QuestionType = "MCQ"
	question.Difficulty = req.Difficulty
	question.CategoryID = req.CategoryID
	question.CorrectOptionID = &req.CorrectOptionID
	question.Comment = req.Comment
	question.CommentDisplayTime = req.CommentTime

	if image1Path != "" {
		utils.DeleteFileIfExists(question.Image1) // delete old image
		question.Image1 = image1Path
		question.Image1DisplayTime = req.Image1Time
	}
	if image2Path != "" {
		utils.DeleteFileIfExists(question.Image2) // delete old image
		question.Image2 = image2Path
		question.Image2DisplayTime = req.Image2Time
	}

	if err := config.DB.Save(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update question"})
		return nil, errors.New("something went wrong")
	}

	// Delete existing options
	config.DB.Where("question_id = ?", question.ID).Delete(&models.QuestionOption{})

	// Add new options
	for i := 1; i <= 5; i++ {
		optionText := c.PostForm(fmt.Sprintf("option_%d", i))
		if optionText != "" {
			isCorrect := uint(i) == req.CorrectOptionID
			option := models.QuestionOption{
				QuestionID: question.ID,
				OptionID:   uint(i),
				OptionText: optionText,
				IsCorrect:  isCorrect,
			}
			config.DB.Create(&option)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "MCQ question updated successfully"})
	return &question, nil
}

func handleEditTrueFalse(c *gin.Context, questionID string) (*models.Question, error) {
	var req TrueFalseRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return nil, errors.New("something went wrong")
	}

	var question models.Question
	if err := config.DB.First(&question, questionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return nil, errors.New("something went wrong")
	}

	image1Path, err := handleOptionalImage(c, "image1", req.Image1Time)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, errors.New("something went wrong")
	}
	image2Path, err := handleOptionalImage(c, "image2", req.Image2Time)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, errors.New("something went wrong")
	}

	if req.Comment != "" && req.CommentTime == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment_display_time required if comment is provided"})
		return nil, errors.New("something went wrong")
	}

	if req.Comment == "" && req.CommentTime != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment required if comment_display_time is provided"})
		return nil, errors.New("something went wrong")
	}

	question.QuestionText = req.QuestionText
	question.QuestionType = "TRUE_FALSE"
	question.Difficulty = req.Difficulty
	question.CategoryID = req.CategoryID
	question.CorrectOptionID = &req.CorrectOptionID
	question.Comment = req.Comment
	question.CommentDisplayTime = req.CommentTime

	if image1Path != "" {
		utils.DeleteFileIfExists(question.Image1) // delete old image
		question.Image1 = image1Path
		question.Image1DisplayTime = req.Image1Time
	}
	if image2Path != "" {
		utils.DeleteFileIfExists(question.Image2) // delete old image
		question.Image2 = image2Path
		question.Image2DisplayTime = req.Image2Time
	}

	if err := config.DB.Save(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update question"})
		return nil, errors.New("something went wrong")
	}

	// Replace True/False options
	config.DB.Where("question_id = ?", question.ID).Delete(&models.QuestionOption{})
	options := []string{"True", "False"}
	for i, text := range options {
		isCorrect := uint(i+1) == req.CorrectOptionID
		option := models.QuestionOption{
			QuestionID: question.ID,
			OptionID:   uint(i + 1),
			OptionText: text,
			IsCorrect:  isCorrect,
		}
		config.DB.Create(&option)
	}

	c.JSON(http.StatusOK, gin.H{"message": "True/False question updated successfully"})
	return &question, nil
}

func handleEditDescriptive(c *gin.Context, questionID string) (*models.Question, error) {
	var req DescriptiveRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return nil, errors.New("something went wrong")
	}

	var question models.Question
	if err := config.DB.First(&question, questionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return nil, errors.New("something went wrong")
	}

	image1Path, err := handleOptionalImage(c, "image1", req.Image1Time)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, errors.New("something went wrong")
	}
	image2Path, err := handleOptionalImage(c, "image2", req.Image2Time)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, errors.New("something went wrong")
	}

	if req.Comment != "" && req.CommentTime == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment_display_time required if comment is provided"})
		return nil, errors.New("something went wrong")
	}

	if req.Comment == "" && req.CommentTime != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment required if comment_display_time is provided"})
		return nil, errors.New("something went wrong")
	}

	question.QuestionText = req.QuestionText
	question.QuestionType = "DESCRIPTIVE"
	question.Difficulty = req.Difficulty
	question.CategoryID = req.CategoryID
	question.Comment = req.Comment
	question.CommentDisplayTime = req.CommentTime

	if image1Path != "" {
		utils.DeleteFileIfExists(question.Image1) // delete old image
		question.Image1 = image1Path
		question.Image1DisplayTime = req.Image1Time
	}
	if image2Path != "" {
		utils.DeleteFileIfExists(question.Image1) // delete old image
		question.Image2 = image2Path
		question.Image2DisplayTime = req.Image2Time
	}

	if err := config.DB.Save(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update question"})
		return nil, errors.New("something went wrong")
	}

	// Save descriptive answer in question_options
	descriptiveOption := models.QuestionOption{
		QuestionID: question.ID,
		OptionID:   1, // Default value for descriptive
		OptionText: req.DescriptiveAnswer,
		IsCorrect:  true,
	}
	if err := config.DB.Create(&descriptiveOption).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save descriptive answer"})
		return nil, errors.New("something went wrong")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Descriptive question updated successfully"})
	return &question, nil
}

// 3. Delete questions
func DeleteQuestion(c *gin.Context) {
	// Parse question ID from the path
	idParam := c.Param("id")
	questionID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}

	// Check if the question exists
	var question models.Question
	if err := config.DB.First(&question, questionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch question"})
		}
		return
	}

	// Begin transaction
	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Delete options
	if err := tx.Where("question_id = ?", question.ID).Delete(&models.QuestionOption{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete question options"})
		return
	}

	// Delete images from disk before deleting the question record
	utils.DeleteFileIfExists(question.Image1)
	utils.DeleteFileIfExists(question.Image2)

	// Delete question
	if err := tx.Delete(&question).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete question"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Question and its options deleted successfully"})
}

type QuestionResponse struct {
	ID           uint   `json:"id"`
	QuestionText string `json:"question_text"`
	CategoryName string `json:"category_name"`
}

// 4. Get Questions
func GetQuestions(c *gin.Context) {
	column := c.Query("column")
	value := c.Query("value")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var questions []struct {
		ID           uint   `json:"id"`
		QuestionText string `json:"question_text"`
		CategoryName string `json:"category_name"`
	}
	var totalCount int64

	// Check cache for first page without filters
	if column == "" && value == "" && page == 1 {
		cacheKey := fmt.Sprintf("questions:page:%d:limit:%d", page, limit)
		cachedData, err := config.RedisClient.Get(config.Ctx, cacheKey).Result()
		if err == nil {
			log.Println("Cache hit for questions")

			var cachedResponse map[string]interface{}
			if err := json.Unmarshal([]byte(cachedData), &cachedResponse); err == nil {
				c.JSON(http.StatusOK, cachedResponse)
				return
			} else {
				log.Printf("Failed to unmarshal cached data: %v", err)
			}
		}
	}

	// Build query
	query := config.DB.Table("questions").
		Select("questions.id, questions.question_text, categories.name AS category_name").
		Joins("LEFT JOIN categories ON questions.category_id = categories.id")

	countQuery := config.DB.Table("questions").
		Joins("LEFT JOIN categories ON questions.category_id = categories.id")

	// Apply filters
	switch strings.ToLower(column) {
	case "question":
		likeVal := "%" + strings.ToLower(value) + "%"
		query = query.Where("LOWER(questions.question_text) LIKE ?", likeVal)
		countQuery = countQuery.Where("LOWER(questions.question_text) LIKE ?", likeVal)
	case "category_name":
		likeVal := "%" + strings.ToLower(value) + "%"
		query = query.Where("LOWER(categories.name) LIKE ?", likeVal)
		countQuery = countQuery.Where("LOWER(categories.name) LIKE ?", likeVal)
	case "":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid column. Use 'question' or 'category_name'"})
		return
	}

	// Get total count
	countQuery.Count(&totalCount)

	// Get questions
	if err := query.Offset(offset).Limit(limit).Scan(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search questions"})
		return
	}

	// Prepare response
	responseData := gin.H{
		"questions":   questions,
		"page":        page,
		"limit":       limit,
		"total_count": totalCount,
		"total_pages": (totalCount + int64(limit) - 1) / int64(limit),
	}

	// Cache the response if it's page 1 and no filters
	if column == "" && value == "" && page == 1 {
		cacheData, _ := json.Marshal(responseData)
		cacheKey := fmt.Sprintf("questions:page:%d:limit:%d", page, limit)
		err := config.RedisClient.Set(config.Ctx, cacheKey, cacheData, 10*time.Minute).Err()
		if err != nil {
			log.Printf(" Failed to cache data: %v", err)
		} else {
			log.Println("Data cached for first page with no filters")
		}
	}

	c.JSON(http.StatusOK, responseData)
}
