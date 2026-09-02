package models

import (
	"time"

	"gorm.io/gorm"
)

type Comments struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	VideoId   uint           `gorm:"index" json:"video_id"`
	AuthorId  uint           `json:"author_id"`
	ParentId  uint           `gorm:"index;default:0" json:"parent_id"` // 0 = 一级评论
	Content   string         `gorm:"type:text" json:"content"`
	LikeCount uint           `gorm:"default:0" json:"like_count"`
}

type CommentInput struct {
	ParentId uint   `gorm:"index;default:0" json:"parent_id"` // 0 = 一级评论
	Content  string `form:"content" json:"content"`
}

type CommentInfo struct {
	Comments
	AuthorName  string `json:"author_name"`
	AuthorImage string `json:"author_image"`
}
