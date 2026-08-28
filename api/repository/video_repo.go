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
	GetVideos(ctx context.Context, videos *[]models.VideoInfo, offset int, limit int) error
	DeleteVideoByID(ctx context.Context, videoID uint) error
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

func (r *videoRepoImpl) GetVideos(ctx context.Context, videos *[]models.VideoInfo, offset int, limit int) error {
	return r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(videos).Error
}

func (r *videoRepoImpl) DeleteVideoByID(ctx context.Context, videoID uint) error {
	return r.db.WithContext(ctx).Where("id = ?", videoID).Delete(&models.VideoInfo{}).Error
}
