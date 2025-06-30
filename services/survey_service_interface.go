package services

import (
	"pathshala/models"

	"github.com/google/uuid"
)

type SurveyServiceInterface interface {
	CreateSurvey(survey *models.Survey) (
		models.Survey, error)
	UpdateSurvey(id uuid.UUID, updatedData map[string]interface{}) (*models.Survey, error)
	DeleteSurvey(id uuid.UUID) error
	GetSurveyByID(id uuid.UUID) (*models.Survey, error)
	GetPaginatedSurveys(page int, pageSize int) ([]models.Survey, int64, error)
	SearchSurveys(query string, page int, pageSize int) ([]models.Survey, int64, error)
}
