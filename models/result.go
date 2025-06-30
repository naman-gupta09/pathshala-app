package models

import "gorm.io/gorm"

type Result struct {
	gorm.Model
	TestID    uint   `json:"test_id"`
	UserID    uint   `json:"user_id"` // Link to User
	Score     int    `json:"score"`
	Correct   int    `json:"correct"`
	Incorrect int    `json:"incorrect"`
	Ignored   int    `json:"ignored"`
	TimeTaken string `json:"timeTaken"`
}
