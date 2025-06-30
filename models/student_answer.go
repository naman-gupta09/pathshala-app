package models

import (
	"gorm.io/gorm"
)

type StudentAnswer struct {
	gorm.Model
	TestID     uint   `json:"test_id"`
	StudentID  uint   `json:"student_id"`
	QuestionID uint   `json:"question_id"`
	Selected   string `json:"selected"` // Selected option (e.g., "A", "B", "C", "D")

}
