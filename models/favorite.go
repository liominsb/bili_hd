package models

import "time"

type Favorite struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UserID    uint      `gorm:"uniqueIndex:idx_favorite_pair;index:idx_user_created,priority:1" json:"user_id"`
	VideoID   uint      `gorm:"uniqueIndex:idx_favorite_pair;index:idx_video" json:"video_id"`
}

// FavoriteItem 收藏列表的一行：收藏 + 视频 + 作者（联表结果，不是数据库表）
// DeletedAt 是 videos 表的软删时间戳：null=视频有效；非 null=UP主已删，前端渲染灰卡片
type FavoriteItem struct {
	VideoID    uint       `json:"video_id"`
	CreatedAt  time.Time  `json:"created_at"` // 收藏时间，不是视频发布时间
	Title      string     `json:"title"`
	Pic        string     `json:"pic"`
	Duration   int        `json:"duration"`
	ViewCount  int        `json:"view_count"`
	AuthorID   uint       `json:"author_id"`
	AuthorName string     `json:"author_name"`
	DeletedAt  *time.Time `json:"deleted_at"` // 指针：NULL→nil→json null，前端 deleted_at !== null 即失效
}
