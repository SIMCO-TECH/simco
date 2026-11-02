package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         *string   `json:"name" validate:"required,min=2,max=100"`
	Email        *string   `json:"email" validate:"required,email" gorm:"unique"`
	Password     *string   `json:"password" validate:"required,min=6"`
	Phone        *string   `json:"phone" validate:"required,min=10,max=10"`
	Token        *string   `json:"token"`
	UserType     *string   `json:"user_type" validate:"required,eq=ADMIN|eq=USER"`
	RefreshToken *string   `json:"refresh_token"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	UserID       uint      `json:"user_id"`
}
