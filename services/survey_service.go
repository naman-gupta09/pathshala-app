package services

import (
	"errors"
	"fmt"
	"pathshala/models"
	"strconv"

	"gorm.io/gorm"
)

type SurveyService struct {
	DB *gorm.DB
}

func NewSurveyService(db *gorm.DB) *SurveyService {
	return &SurveyService{DB: db}
}

func (s *SurveyService) CreateSurvey(survey models.Survey) error {
	return s.DB.Create(survey).Error
}

func (s *SurveyService) GetAllSurveys() ([]models.Survey, error) {
	var surveys []models.Survey
	err := s.DB.Order("created_at desc").Find(&surveys).Error
	return surveys, err
}

func (s *SurveyService) GetSurveyByID(id string) (models.Survey, error) {
	var survey models.Survey
	err := s.DB.First(&survey, "id = ?", id).Error
	return survey, err
}

func (s *SurveyService) UpdateSurvey(id string, updatedData map[string]interface{}) error {
	// Check if survey exists
	var survey models.Survey
	if err := s.DB.First(&survey, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("survey with ID %s not found: %w", id, err)
		}
		return fmt.Errorf("error finding survey: %w", err)
	}

	// Perform the update
	result := s.DB.Model(&survey).Updates(updatedData)
	if result.Error != nil {
		return fmt.Errorf("failed to update survey : %w", result.Error)
	}

	// Optional: check if any row was affected
	if result.RowsAffected == 0 {
		return fmt.Errorf("no changes were made to survey with ID %s", id)
	}
	return nil
}

// function for search survey

func (s *SurveyService) SearchSurveys(query string) ([]models.Survey, error) {
	var surveys []models.Survey

	err := s.DB.Where(
		"LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(survey_type) LIKE ?",
		"%"+query+"%",
		"%"+query+"%",
		"%"+query+"%",
	).Find(&surveys).Error

	if err != nil {
		return nil, fmt.Errorf("failed to search surveys: %w", err)
	}

	return surveys, nil
}

func (s *SurveyService) DeleteSurvey(id string) error {
	return s.DB.Delete(&models.Survey{}, "id = ?", id).Error
}

func (s *SurveyService) GetPaginatedSurveys(pageStr, limitStr string) ([]models.Survey, int64, error) {
	var surveys []models.Survey
	var total int64

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Fetch surveys with limit and offset
	if err := s.DB.Limit(limit).Offset(offset).Order("created_at DESC").Find(&surveys).Error; err != nil {
		return nil, 0, err
	}

	// Get total count for pagination
	if err := s.DB.Model(&models.Survey{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return surveys, total, nil
}
