package repository

import (
	"context"
	"go_bili/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FollowRepository interface {
	Follow(ctx context.Context, followerID uint, followingID uint) error
	Unfollow(ctx context.Context, followerID uint, followingID uint) error
	IsFollowing(ctx context.Context, followerID uint, followingID uint) (bool, error)
	CountFollowing(ctx context.Context, userID uint) (int64, error)
	CountFollowers(ctx context.Context, userID uint) (int64, error)
	ListFollowing(ctx context.Context, userID uint, list *[]models.UserBrief, offset int, limit int) error
	ListFollowers(ctx context.Context, userID uint, list *[]models.UserBrief, offset int, limit int) error
}

type followRepoImpl struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) FollowRepository {
	return &followRepoImpl{db: db}
}

// Follow 关注：OnConflict 忽略冲突，重复关注幂等返回成功
func (r *followRepoImpl) Follow(ctx context.Context, followerID uint, followingID uint) error {
	follow := models.Follow{FollowerID: followerID, FollowingID: followingID}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&follow).Error
}

// Unfollow 取关：物理删除，关系不存在也返回成功（幂等）
func (r *followRepoImpl) Unfollow(ctx context.Context, followerID uint, followingID uint) error {
	return r.db.WithContext(ctx).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(&models.Follow{}).Error
}

func (r *followRepoImpl) IsFollowing(ctx context.Context, followerID uint, followingID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Follow{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count).Error
	return count > 0, err
}

// CountFollowing 我关注了多少人
func (r *followRepoImpl) CountFollowing(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Follow{}).
		Where("follower_id = ?", userID).Count(&count).Error
	return count, err
}

// CountFollowers 有多少人关注我
func (r *followRepoImpl) CountFollowers(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Follow{}).
		Where("following_id = ?", userID).Count(&count).Error
	return count, err
}

// ListFollowing 我关注的人：follows.following_id 关联 users 拿头像昵称
func (r *followRepoImpl) ListFollowing(ctx context.Context, userID uint, list *[]models.UserBrief, offset int, limit int) error {
	return r.db.WithContext(ctx).
		Table("follows").
		Select("users.id, users.username, users.image, users.bio").
		Joins("JOIN users ON users.id = follows.following_id").
		Where("follows.follower_id = ?", userID).
		Order("follows.created_at DESC").
		Offset(offset).Limit(limit).
		Scan(list).Error
}

// ListFollowers 关注我的人：follows.follower_id 关联 users
func (r *followRepoImpl) ListFollowers(ctx context.Context, userID uint, list *[]models.UserBrief, offset int, limit int) error {
	return r.db.WithContext(ctx).
		Table("follows").
		Select("users.id, users.username, users.image, users.bio").
		Joins("JOIN users ON users.id = follows.follower_id").
		Where("follows.following_id = ?", userID).
		Order("follows.created_at DESC").
		Offset(offset).Limit(limit).
		Scan(list).Error
}
