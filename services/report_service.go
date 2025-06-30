package services

import (
	"pathshala/models"

	"gorm.io/gorm"
)

type ReportService struct {
	DB *gorm.DB
}

func NewReportService(db *gorm.DB) *ReportService {
	return &ReportService{DB: db}
}

func (rs *ReportService) GetAllReportTypes() ([]models.ReportType, error) {
	var types []models.ReportType
	err := rs.DB.Find(&types).Error
	return types, err
}

func (rs *ReportService) GetTestWithStudentScores(testID string) ([]map[string]interface{}, error) {
	var data []map[string]interface{}
	query := `
		SELECT s.name AS student_name, r.score
		FROM results r
		JOIN students s ON r.student_id = s.id
		WHERE r.test_id = ?
		ORDER BY r.score DESC`
	err := rs.DB.Raw(query, testID).Scan(&data).Error
	return data, err
}

func (rs *ReportService) GetStudentParticipationRanking() ([]map[string]interface{}, error) {
	var data []map[string]interface{}
	query := `
		SELECT s.name AS student_name, COUNT(st.test_id) AS test_count
		FROM student_tests st
		JOIN students s ON st.student_id = s.id
		GROUP BY s.name
		ORDER BY test_count DESC`
	err := rs.DB.Raw(query).Scan(&data).Error
	return data, err
}
