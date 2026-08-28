package models

import (
	"time"

	"gorm.io/gorm"
)

type VideoInfo struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	AuthorId  uint           `json:"author_id"`
	Title     string         `json:"title"`
	Pic       string         `json:"pic"`
	VideoUrl  string         `json:"video_url"`
}

type VideoInput struct {
	Title    string `json:"title"`
	Pic      string `json:"pic"`
	VideoUrl string `json:"video_url"`
}
