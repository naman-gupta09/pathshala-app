package services

import "pathshala/models"

type ReportServiceInterface interface {
	GetAllReportTypes() ([]models.ReportType, error)
	GetTestWithStudentScores(testID string) ([]map[string]interface{}, error)
	GetStudentParticipationRanking() ([]map[string]interface{}, error)
}

var _ ReportServiceInterface = &ReportService{}
