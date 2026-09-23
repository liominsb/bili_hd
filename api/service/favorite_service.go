package service

import (
	"context"
	"errors"
	"go_bili/api/repository"
	"go_bili/models"

	"gorm.io/gorm"
)

type FavoriteService interface {
	Favorite(ctx context.Context, userID uint, videoID uint) error
	Unfavorite(ctx context.Context, userID uint, videoID uint) error
	IsFavorite(ctx context.Context, userID uint, videoID uint) (bool, error)
	CountByVideo(ctx context.Context, videoID uint) (int64, error)
	ListMyFavorites(ctx context.Context, userID uint, offset int, limit int) (*[]models.FavoriteItem, error)
}

type favoriteServiceImpl struct {
	favoriteRepo repository.FavoriteRepository
	videoService VideoService // 注意：注入的是 videoService 不是 videoRepo，为的是吃它的 Redis 缓存
}

func NewFavoriteService(favoriteRepo repository.FavoriteRepository, videoService VideoService) FavoriteService {
	return &favoriteServiceImpl{favoriteRepo: favoriteRepo, videoService: videoService}
}

// Favorite 收藏：只校验视频存在（没有"不能收藏自己"这种限制）
func (s *favoriteServiceImpl) Favorite(ctx context.Context, userID uint, videoID uint) error {
	if err := s.checkVideoExists(ctx, videoID); err != nil {
		return err
	}
	return s.favoriteRepo.Favorite(ctx, userID, videoID)
}

func (s *favoriteServiceImpl) Unfavorite(ctx context.Context, userID uint, videoID uint) error {
	return s.favoriteRepo.Unfavorite(ctx, userID, videoID)
}

func (s *favoriteServiceImpl) IsFavorite(ctx context.Context, userID uint, videoID uint) (bool, error) {
	return s.favoriteRepo.IsFavorite(ctx, userID, videoID)
}

func (s *favoriteServiceImpl) CountByVideo(ctx context.Context, videoID uint) (int64, error) {
	return s.favoriteRepo.CountByVideo(ctx, videoID)
}

func (s *favoriteServiceImpl) ListMyFavorites(ctx context.Context, userID uint, offset int, limit int) (*[]models.FavoriteItem, error) {
	list := &[]models.FavoriteItem{}
	if err := s.favoriteRepo.ListMyFavorites(ctx, userID, list, offset, limit); err != nil {
		return nil, err
	}
	return list, nil
}

// checkVideoExists 校验视频存在：与视频详情共用 VIDEO:<id> 缓存
func (s *favoriteServiceImpl) checkVideoExists(ctx context.Context, videoID uint) error {
	video, err := s.videoService.FindVideoByID(ctx, videoID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("视频不存在")
	}
	if video == nil && err == nil {
		return errors.New("视频不存在")
	}
	return err
}
