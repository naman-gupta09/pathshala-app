package models

type MacroCategory struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	// UserID      uint        `json:"user_id"`
	// User        models.User `gorm:"foreignKey:UserID" json:"-"`
}

type Category struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"type:varchar(100);not null" json:"name"`
	MacroCategoryID *uint          `json:"-"`
	MacroCategory   *MacroCategory `gorm:"foreignKey:MacroCategoryID" json:"macro_category,omitempty"`
	CollegeName     *string        `gorm:"type:varchar(100)" json:"college_name,omitempty"`
	State           *string        `gorm:"type:varchar(100)" json:"state,omitempty"`
	CreatorName     string         `gorm:"type:varchar(100)" json:"creator_name"`
	Description     *string        `gorm:"type:text" json:"description,omitempty"`
	ImagePath       *string        `gorm:"type:text" json:"image_path,omitempty"`
	Status          string         `json:"status" gorm:"default:'draft'"` // values: "draft", "active"
}

type CategoryResponse struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	MacroCategory string `json:"macro_category"`
	CollegeName   string `json:"college_name,omitempty"`
	State         string `json:"state,omitempty"`
	// CreatorName   string `json:"creator_name"`
}

type CreateCategoryInput struct {
	Name          string `json:"name" binding:"required"`
	MacroCategory string `json:"macro_category" binding:"required"`
	Creator       string `json:"creator" binding:"required"`
	CollegeID     uint   `json:"college_id" binding:"required"`
	State         string `json:"state" binding:"required"`
}
