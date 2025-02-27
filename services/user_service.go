package services

import (
	"errors"

	"github.com/tison2810/be-go-tc/models"
	"gorm.io/gorm"
)

func FindUserByEmail(db *gorm.DB, email string) (*models.User, error) {
	var user models.User
	result := db.Where("mail = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, result.Error
}

func CreateUser(db *gorm.DB, user models.User) (*models.User, error) {
	result := db.Create(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}
