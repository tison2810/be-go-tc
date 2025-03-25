package models

import (
	"time"
)

type User struct {
	FirstName string `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName  string `json:"last_name" gorm:"type:varchar(100);not null"`
	Mail      string `json:"mail" gorm:"type:varchar(100);primaryKey;unique;not null"`
	// Password  string    `json:"password" gorm:"type:varchar(255);not null"`
	Maso      string    `json:"maso" gorm:"type:varchar(50);not null"`
	Role      string    `json:"role" gorm:"type:varchar(50);not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}
type GoogleUser struct {
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}
