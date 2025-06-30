package models

import (
	"time"
)

// Question represents a test question
type Question struct {
	ID                 uint      `gorm:"primaryKey"`
	QuestionType       string    `gorm:"type:varchar(20);not null"` // MCQ, TRUE_FALSE, DESCRIPTIVE
	QuestionText       string    `gorm:"type:text;not null"`
	CorrectOptionID    *uint     `gorm:"default:null"`              // Points to a correct option for MCQ & True/False
	Difficulty         string    `gorm:"type:varchar(10);not null"` // EASY, MEDIUM, HARD
	CategoryID         uint      `gorm:"not null"`                  // Foreign key to categories table
	Category           Category  `gorm:"foreignKey:CategoryID;references:ID"`
	Image1             string    `gorm:"type:text;default:null"` // Image path or Base64
	Image1DisplayTime  *int      `gorm:"default:null"`
	Image2             string    `gorm:"type:text;default:null"`
	Image2DisplayTime  *int      `gorm:"default:null"`
	Comment            string    `gorm:"type:text;default:null"`
	CommentDisplayTime *int      `gorm:"default:null"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`

	Options []QuestionOption `gorm:"foreignKey:QuestionID"` // Relation with options
}

// QuestionOption represents an option for MCQ & True/False questions
type QuestionOption struct {
	ID         uint   `gorm:"primaryKey"`
	QuestionID uint   `gorm:"not null"` // Foreign key to Question
	OptionID   uint   `gorm:"not null"` // 1 to 5 for each question
	OptionText string `gorm:"type:text;not null"`
	IsCorrect  bool   `gorm:"not null;default:false"` // TRUE for correct option
}
