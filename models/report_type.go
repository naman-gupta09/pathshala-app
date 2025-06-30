package models

type ReportType struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Type string `gorm:"not null;unique" json:"type"`
}
