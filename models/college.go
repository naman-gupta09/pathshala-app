package models

type College struct {
	ID               uint        `gorm:"primaryKey" json:"id"`
	Name             string      `json:"name" binding:"required"`
	Description      string      `json:"description" binding:"required"`
	State            string      `json:"state" binding:"required"`
	CollegeTypeID    uint        `json:"college_type_id" binding:"required"`
	ActiveCandidates int         `json:"active_candidates" binding:"gte=0"`
	CollegeType      CollegeType `gorm:"foreignKey:CollegeTypeID;references:ID" json:"-"`
}

type CollegeType struct {
	ID              uint   `json:"id" gorm:"primaryKey"`
	Name            string `json:"name" gorm:"unique;not null" binding:"required"`
	TypeDescription string `json:"type_description" binding:"required"`
}

// Custom response model
type CollegeResponse struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	State            string `json:"state"`
	CollegeType      string `json:"college_type"` // Just the name
	ActiveCandidates int    `json:"active_candidates"`
}
