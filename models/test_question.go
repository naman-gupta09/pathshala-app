package models

type TestQuestion struct {
	ID         uint `gorm:"primaryKey"`
	TestID     uint `gorm:"not null"`
	QuestionID uint `gorm:"not null"`
}

// // will remove it later
// type Test struct {
// 	ID        uint `gorm:"primaryKey"`
// 	TeacherID uint
// }

// type Category struct {
// 	ID   uint   `gorm:"primaryKey"`
// 	name string `gorm:"not null"`
// }
