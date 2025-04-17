package models

import (
	"time"

	"gorm.io/gorm"
)

// Admin represents system administrators
type Admin struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"type:varchar(50);unique;not null" json:"username"`
	Password  string    `gorm:"type:varchar(100);not null" json:"-"` // Password not exposed in JSON
	Email     string    `gorm:"type:varchar(100);unique" json:"email"`
	Phone     string    `gorm:"type:varchar(20)" json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate is a GORM hook that runs before creating a new record
func (a *Admin) BeforeCreate(tx *gorm.DB) error {
	// TODO: Add password hashing logic here
	return nil
}
