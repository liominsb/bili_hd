package models

import "time"

type Follow struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	FollowerID  uint      `gorm:"uniqueIndex:idx_follow_pair" json:"follower_id"`                      // 粉丝
	FollowingID uint      `gorm:"uniqueIndex:idx_follow_pair;index:idx_following" json:"following_id"` // UP主
}

// UserBrief 关注/粉丝列表用的精简用户信息（联表查询结果，非数据库表）
type UserBrief struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Image    string `json:"image"`
	Bio      string `json:"bio"`
}
