package models

import (
	"time"
)

type User struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name" validate:"required,min=2,max=100" gorm:"not null"`
	Email         string    `json:"email" validate:"required,email" gorm:"uniqueIndex;not null"`
	Password      string    `json:"password" validate:"required,min=6" gorm:"not null"`
	Phone         string    `json:"phone" validate:"required,len=10"`
	Token         string    `json:"token"`
	UserType      string    `json:"user_type" validate:"required,oneof=ADMIN USER"`
	Refresh_token string    `json:"refresh_token"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
