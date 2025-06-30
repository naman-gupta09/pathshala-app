package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Survey struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;"  json:"id"`
	Name         string    `gorm:"not null"  json:"name"`
	Description  string    `json:"description"`
	Category     string    `json:"category"`                    //example golang react
	SurveyType   string    `json:"survey_type"`                 //multiple choice
	ShowResults  bool      `json:"show_results"`                //example true false or toggle
	NumQuestions int       `json:"num_questions"`               //total number of questions
	CreatedBy    uuid.UUID `json:"type:uuid" json:"created_by"` //user (teacher/admin) who created it it store the user id who created it
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Hook: generate UUID automatically before create
func (s *Survey) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuid.New()
	return
}
