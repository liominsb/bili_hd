package repository

import (
	"context"
	"go_bili/models"

	"gorm.io/gorm"
)

type VideoRepository interface {
	AddNewVideo(ctx context.Context, video *models.VideoInfo) error
	UpdateVideo(ctx context.Context, video *models.VideoInfo) error
	FindVideoByID(ctx context.Context, video *models.VideoInfo, videoID uint) error
	FindVideoByIDWithAuthor(ctx context.Context, videoID uint) (*models.VideoInfoWithAuthor, error)
	GetVideos(ctx context.Context, videos *[]models.VideoInfo, offset int, limit int) error
	DeleteVideoByID(ctx context.Context, videoID uint) error
	SearchVideoByTitle(ctx context.Context, videos *[]models.VideoInfo, title string, offset int, limit int) error
	AddVideoLike(ctx context.Context, userid uint, videoID uint) error
	DelVideoLike(ctx context.Context, userid uint, videoID uint) error
	GetUserANDVideoLike(ctx context.Context, userid uint, videoID uint) (bool, error)
}
type videoRepoImpl struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) VideoRepository {
	return &videoRepoImpl{db: db}
}

func (r *videoRepoImpl) AddNewVideo(ctx context.Context, video *models.VideoInfo) error {
	return r.db.WithContext(ctx).Create(video).Error
}

func (r *videoRepoImpl) UpdateVideo(ctx context.Context, video *models.VideoInfo) error {
	return r.db.WithContext(ctx).Updates(video).Error
}

func (r *videoRepoImpl) FindVideoByID(ctx context.Context, video *models.VideoInfo, videoID uint) error {
	return r.db.WithContext(ctx).Where("id = ?", videoID).First(video).Error
}

func (r *videoRepoImpl) FindVideoByIDWithAuthor(ctx context.Context, videoID uint) (*models.VideoInfoWithAuthor, error) {
	result := &models.VideoInfoWithAuthor{}
	err := r.db.WithContext(ctx).
		Table("video_infos").
		Select("video_infos.*, users.username AS author_name, users.image AS author_image").
		Joins("LEFT JOIN users ON users.id = video_infos.author_id").
		Where("video_infos.id = ?", videoID).
		First(result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *videoRepoImpl) GetVideos(ctx context.Context, videos *[]models.VideoInfo, offset int, limit int) error {
	return r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(videos).Error
}

func (r *videoRepoImpl) DeleteVideoByID(ctx context.Context, videoID uint) error {
	return r.db.WithContext(ctx).Where("id = ?", videoID).Delete(&models.VideoInfo{}).Error
}

func (r *videoRepoImpl) SearchVideoByTitle(ctx context.Context, videos *[]models.VideoInfo, title string, offset int, limit int) error {
	return r.db.WithContext(ctx).Offset(offset).Limit(limit).Where("title LIKE ?", title+"%").
		Order("id DESC").Find(videos).Error
}

func (r *videoRepoImpl) GetUserANDVideoLike(ctx context.Context, userid uint, videoID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.VideoLike{}).Where("video_id = ? AND user_id = ?", videoID, userid).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *videoRepoImpl) AddVideoLike(ctx context.Context, userid uint, videoID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&models.VideoInfo{}).Where("id = ?", videoID).
			Update("like_count", gorm.Expr("like_count + 1")).Error
		if err != nil {
			return err
		}
		err = tx.Model(&models.VideoLike{}).Create(&models.VideoLike{VideoId: videoID, UserID: userid}).Error
		if err != nil {
			return err
		}
		return nil
	})
}

func (r *videoRepoImpl) DelVideoLike(ctx context.Context, userid uint, videoID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&models.VideoInfo{}).Where("id = ?", videoID).
			Update("like_count", gorm.Expr("like_count - 1")).Error
		if err != nil {
			return err
		}
		err = tx.Model(&models.VideoLike{}).Where("video_id = ? AND user_id = ?", videoID, userid).Delete(&models.VideoLike{}).Error
		if err != nil {
			return err
		}
		return nil
	})
}
