package repository

import (
	"context"
	"go_bili/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FavoriteRepository interface {
	Favorite(ctx context.Context, userID uint, videoID uint) error
	Unfavorite(ctx context.Context, userID uint, videoID uint) error
	IsFavorite(ctx context.Context, userID uint, videoID uint) (bool, error)
	CountByVideo(ctx context.Context, videoID uint) (int64, error)
	ListMyFavorites(ctx context.Context, userID uint, list *[]models.FavoriteItem, offset int, limit int) error
}

type favoriteRepoImpl struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) FavoriteRepository {
	return &favoriteRepoImpl{db: db}
}

// Favorite 收藏：OnConflict 忽略冲突，重复收藏幂等返回成功（沿用 follow 的套路）
func (r *favoriteRepoImpl) Favorite(ctx context.Context, userID uint, videoID uint) error {
	favorite := models.Favorite{UserID: userID, VideoID: videoID}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&favorite).Error
}

// Unfavorite 取消收藏：物理删 favorites 行，不碰 videos 表——视频删没删都能取消
func (r *favoriteRepoImpl) Unfavorite(ctx context.Context, userID uint, videoID uint) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND video_id = ?", userID, videoID).
		Delete(&models.Favorite{}).Error
}

func (r *favoriteRepoImpl) IsFavorite(ctx context.Context, userID uint, videoID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Favorite{}).
		Where("user_id = ? AND video_id = ?", userID, videoID).
		Count(&count).Error
	return count > 0, err
}

// CountByVideo 这个视频被多少人收藏（详情页计数用）
func (r *favoriteRepoImpl) CountByVideo(ctx context.Context, videoID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Favorite{}).
		Where("video_id = ?", videoID).Count(&count).Error
	return count, err
}

// ListMyFavorites 我的收藏列表：JOIN videos + users 拼卡片数据
// 注意：JOIN 里故意不过滤 videos.deleted_at——软删的视频要保留在列表里（B 站行为），
// 但必须把 videos.deleted_at Select 出去，前端才知道它是失效视频该渲染成灰卡片。
// 手写 Joins 时 GORM 不会自动给 JOIN 的表加软删过滤（只对 Model() 的主模型加），
// 所以这里加不加、怎么加，全由我们自己说了算。
func (r *favoriteRepoImpl) ListMyFavorites(ctx context.Context, userID uint, list *[]models.FavoriteItem, offset int, limit int) error {
	return r.db.WithContext(ctx).Raw(`
        SELECT
            favorites.video_id,
            favorites.created_at,
            video_infos.title,
            video_infos.pic,
            video_infos.duration,
            video_infos.view_count,
            video_infos.author_id,
            users.username AS author_name
        FROM favorites
        JOIN video_infos ON video_infos.id = favorites.video_id
        LEFT JOIN users ON users.id = video_infos.author_id
        WHERE favorites.user_id = ?
        ORDER BY favorites.created_at DESC
        LIMIT ? OFFSET ?
    `, userID, limit, offset).Scan(list).Error
}
