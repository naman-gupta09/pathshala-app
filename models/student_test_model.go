package models

import "time"

// StudentTest model (for mapping students to tests)
type StudentTest struct {
	ID        uint `gorm:"primaryKey"`
	StudentID uint
	TestID    uint      `json:"test_id" binding:"required"`
	StartTime time.Time `json:"start_time"` // Automatically added when test is assigned
}
