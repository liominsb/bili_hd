package models

import "time"

type PlayHistory struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UpdatedAt time.Time `gorm:"index:idx_user_updated,priority:2" json:"updated_at"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_video,priority:1;index:idx_user_updated,priority:1" json:"user_id"`
	VideoID   uint      `gorm:"uniqueIndex:idx_user_video,priority:2" json:"video_id"`
	Progress  int       `gorm:"default:0" json:"progress"`
}

// HistoryInput 上报进度的入参
type HistoryInput struct {
	Progress int `json:"progress"`
}

// PlayHistoryItem 历史列表的一行：历史 + 视频 + 作者（联表结果，不是数据库表）
type PlayHistoryItem struct {
	VideoID    uint      `json:"video_id"`
	Progress   int       `json:"progress"`
	UpdatedAt  time.Time `json:"updated_at"`
	Title      string    `json:"title"`
	Pic        string    `json:"pic"`
	Duration   int       `json:"duration"`
	ViewCount  int       `json:"view_count"`
	AuthorID   uint      `json:"author_id"`
	AuthorName string    `json:"author_name"`
}
