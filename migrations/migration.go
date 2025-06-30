package migrations

import (
	"pathshala/config"
	"pathshala/models"
)

// Migrate creates the necessary tables
func MigrateQuestions() {
	config.DB.AutoMigrate(&models.User{}, &models.CollegeType{}, &models.College{})
	config.DB.AutoMigrate(&models.Teacher{}, &models.Student{})
	config.DB.AutoMigrate(&models.MacroCategory{}, &models.Category{})
	config.DB.AutoMigrate(&models.Question{}, &models.QuestionOption{})
	config.DB.AutoMigrate(&models.Test{}, &models.TestQuestion{})
	config.DB.AutoMigrate(&models.StudentAnswer{}, &models.StudentTest{}, &models.Result{})
	config.DB.AutoMigrate(&models.Survey{}, &models.ReportType{})
}
