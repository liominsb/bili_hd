package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Username  string         `gorm:"size:50;uniqueIndex" json:"username"`
	Password  string         `json:"-"`
	Bio       string         `gorm:"size:500" json:"bio"`
	Image     string         `gorm:"size:500" json:"image"`
}

type UserInput struct {
	Username string `gorm:"size:50;uniqueIndex" json:"username"`
	Password string `json:"-"`
	Bio      string `gorm:"size:500" json:"bio"`
	Image    string `gorm:"size:500" json:"image"`
}
