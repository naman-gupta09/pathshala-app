package models

type User struct {
	ID             uint    `json:"id" gorm:"primaryKey"`
	Name           string  `json:"name" binding:"required"`
	Email          string  `gorm:"unique" json:"email" binding:"required,email"`
	Password       string  `json:"-" gorm:"not null"`
	CollegeID      *uint   `json:"college_id"` // Nullable
	College        College `gorm:"foreignKey:CollegeID" json:"college,omitempty"`
	Role           string  `json:"role" binding:"required,oneof=student teacher admin"`
	SecondaryEmail *string `json:"secondary_email,omitempty"`
	Profile_image  string  `json:"profile_image,omitempty"`
}

type Student struct {
	ID     uint   `gorm:"primaryKey"`
	UserID uint   `json:"user_id" binding:"required"`
	Status string `json:"status" binding:"required,oneof=active inactive"`
	User   User   `gorm:"foreignKey:UserID" json:"user"`
	Branch string `json:"branch"`
	Gender string `json:"gender"`
	//Status string `gorm:"default:active"`
}

type Teacher struct {
	ID          uint   `gorm:"primaryKey"`
	UserID      uint   `json:"user_id" binding:"required"`
	State       string `json:"state" binding:"required"`
	TeacherType string `json:"teacher_type" binding:"required"`
	User        User   `gorm:"foreignKey:UserID" json:"user"`
	Super       bool   `json:"super_teacher"` // True if added as super teacher
	Status      string `json:"status" binding:"required,oneof=active inactive"`
}

type StudentResponse struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	College string `json:"college"` // College name only
	Status  string `json:"status"`
}

type TeacherResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	State       string `json:"state"`
	College     string `json:"college"` // Just college name
	TeacherType string `json:"teacher_type"`
	Status      string `json:"status" binding:"required,oneof=active inactive"`
}
