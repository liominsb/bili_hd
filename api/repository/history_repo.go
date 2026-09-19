package repository

import (
	"context"
	"go_bili/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type HistoryRepository interface {
	Upsert(ctx context.Context, history *models.PlayHistory) error
	Delete(ctx context.Context, userID uint, videoID uint) error
	Clear(ctx context.Context, userID uint) error
	List(ctx context.Context, list *[]models.PlayHistoryItem, userID uint, offset int, limit int) error
}

type historyRepoImpl struct {
	db *gorm.DB
}

func NewHistoryRepository(db *gorm.DB) HistoryRepository {
	return &historyRepoImpl{db: db}
}

// Upsert 上报进度：撞 (user_id, video_id) 唯一索引时整行覆盖。
// UpdateAll 会把 updated_at 一并刷成当前值，不需要手动赋值
func (r *historyRepoImpl) Upsert(ctx context.Context, history *models.PlayHistory) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{UpdateAll: true}).
		Create(history).Error
}

// Delete 删单条：按 (user_id, video_id) 删，记录不存在也算成功（幂等）
func (r *historyRepoImpl) Delete(ctx context.Context, userID uint, videoID uint) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND video_id = ?", userID, videoID).
		Delete(&models.PlayHistory{}).Error
}

// Clear 清空某个用户的全部历史
func (r *historyRepoImpl) Clear(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.PlayHistory{}).Error
}

// List 历史列表：play_histories JOIN video_infos 拿标题封面，再 JOIN users 拿作者名
// video_infos 是软删除表，deleted_at 条件要手写；LIMIT ? OFFSET ? 的传参顺序是 limit, offset
func (r *historyRepoImpl) List(ctx context.Context, list *[]models.PlayHistoryItem, userID uint, offset int, limit int) error {
	return r.db.WithContext(ctx).Raw(`
		SELECT ph.video_id, ph.progress, ph.updated_at,
		       v.title, v.pic, v.duration, v.view_count, v.author_id,
		       u.username AS author_name
		FROM play_histories ph
		JOIN video_infos v ON v.id = ph.video_id AND v.deleted_at IS NULL
		LEFT JOIN users u ON u.id = v.author_id
		WHERE ph.user_id = ?
		ORDER BY ph.updated_at DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset).Scan(list).Error
}
