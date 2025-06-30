// utils/authorization.go
package utils

import (
	"errors"
	"pathshala/config"
	"pathshala/models"
)

var (
	ErrTestNotFound = errors.New("test not found")
	ErrUnauthorized = errors.New("unauthorized access to test")
)

// AuthorizeTestAccess checks if the user is the creator of the test or an admin
func AuthorizeTestAccess(testID uint, userID uint, role string) error {
	var test models.Test
	if err := config.DB.First(&test, testID).Error; err != nil {
		return ErrTestNotFound
	}

	if role != "admin" && test.UserID != userID {
		return ErrUnauthorized
	}
	return nil
}
