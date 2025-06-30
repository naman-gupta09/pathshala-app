package models

import "gorm.io/gorm"

var DB *gorm.DB

func SetDB(db *gorm.DB) {
	DB = db
}

func GetDB() *gorm.DB {
	return DB
}
