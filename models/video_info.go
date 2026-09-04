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
	Title     string         `gorm:"size:255;index" json:"title"`
	Pic       string         `json:"pic"`
	VideoUrl  string         `json:"video_url"`
	LikeCount int            `gorm:"default:0" json:"like_count"`
	//// 状态字段：0: 转码处理中, 1: 正常/已发布, 2: 转码失败
	//Status int `gorm:"default:0;index" json:"status"`
}

type VideoInput struct {
	Title    string `json:"title"`
	Pic      string `json:"pic"`
	VideoUrl string `json:"video_url"`
}

// VideoInfoWithAuthor 视频详情 + 作者信息（联表结果，非数据库表）
type VideoInfoWithAuthor struct {
	VideoInfo
	AuthorName  string `json:"author_name"`
	AuthorImage string `json:"author_image"`
	AuthorBio   string `json:"author_bio"`
}

type VideoLike struct {
	ID      uint `gorm:"primarykey" json:"id"`
	UserID  uint `gorm:"index" json:"user_id"`
	VideoId uint `gorm:"index" json:"video_id"`
}
