package service

import (
	"context"
	"go_bili/api/repository"
	"go_bili/models"
)

type HistoryService interface {
	Report(ctx context.Context, userID uint, videoID uint, progress int) error
	Delete(ctx context.Context, userID uint, videoID uint) error
	Clear(ctx context.Context, userID uint) error
	List(ctx context.Context, userID uint, offset int, limit int) (*[]models.PlayHistoryItem, error)
}

type historyServiceImpl struct {
	historyRepo repository.HistoryRepository
}

func NewHistoryService(historyRepo repository.HistoryRepository) HistoryService {
	return &historyServiceImpl{historyRepo: historyRepo}
}

// Report 上报播放进度：裁剪掉负数后 upsert
func (s *historyServiceImpl) Report(ctx context.Context, userID uint, videoID uint, progress int) error {
	if progress < 0 {
		progress = 0
	}
	return s.historyRepo.Upsert(ctx, &models.PlayHistory{
		UserID:   userID,
		VideoID:  videoID,
		Progress: progress,
	})
}

func (s *historyServiceImpl) Delete(ctx context.Context, userID uint, videoID uint) error {
	return s.historyRepo.Delete(ctx, userID, videoID)
}

func (s *historyServiceImpl) Clear(ctx context.Context, userID uint) error {
	return s.historyRepo.Clear(ctx, userID)
}

func (s *historyServiceImpl) List(ctx context.Context, userID uint, offset int, limit int) (*[]models.PlayHistoryItem, error) {
	list := &[]models.PlayHistoryItem{}
	if err := s.historyRepo.List(ctx, list, userID, offset, limit); err != nil {
		return nil, err
	}
	return list, nil
}
