package models

type Test struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	TestName     string `json:"test_name"`
	UserID       uint   `json:"teacher_id"`
	User         User   `gorm:"foreignKey:UserID" json:"user"`
	MinQuestions int    `json:"min_questions"`
}
