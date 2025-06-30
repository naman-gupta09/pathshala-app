package utils

import (
	"log"
	"os"
	"pathshala/config"
	"pathshala/models"

	"golang.org/x/crypto/bcrypt"
)

func SeedAdminUser() {
	var count int64
	if err := config.DB.Model(&models.User{}).Count(&count).Error; err != nil {
		log.Fatalf("Failed to count users: %v", err)
	}

	if count == 0 {
		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(os.Getenv("ADMIN_PASSWORD")), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash admin password: %v", err)
		}

		admin := models.User{
			Name:     os.Getenv("ADMIN_NAME"),
			Email:    os.Getenv("ADMIN_EMAIL"),
			Password: string(hashedPassword),
			Role:     "admin", // if you use roles
		}

		if err := config.DB.Create(&admin).Error; err != nil {
			log.Fatalf("Failed to create default admin user: %v", err)
		}

		log.Println("Default admin user created.")
	} else {
		log.Println("Users already exist. Skipping admin seed.")
	}
}
