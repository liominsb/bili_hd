package service

import (
	"context"
	"fmt"
	"go_bili/api/repository"
	"go_bili/models"
	"go_bili/utils"
	"log"

	"github.com/redis/go-redis/v9"
)

type VideoService interface {
	AddNewVideo(ctx context.Context, video *models.VideoInfo) error
	UpdateVideo(ctx context.Context, video *models.VideoInfo) error
	FindVideoByID(ctx context.Context, videoID uint) (*models.VideoInfo, error)
	FindVideoByIDWithAuthor(ctx context.Context, videoID uint) (*models.VideoInfoWithAuthor, error)
	GetVideos(ctx context.Context, offset int, limit int) ([]models.VideoInfoWithAuthor, error)
	DeleteVideo(ctx context.Context, id uint) error
	SearchVideoByTitle(ctx context.Context, title string, offset int, limit int) ([]models.VideoInfoWithAuthor, error)
	UpdateVideoLike(ctx context.Context, userid uint, videoID uint) (bool, error)
}

type videoServiceImpl struct {
	videoRepo   repository.VideoRepository
	redisClient *redis.Client
}

func NewVideoService(videoRepo repository.VideoRepository, redisClient *redis.Client) VideoService {
	return &videoServiceImpl{videoRepo: videoRepo, redisClient: redisClient}
}

func (s *videoServiceImpl) AddNewVideo(ctx context.Context, video *models.VideoInfo) error {
	if err := s.videoRepo.AddNewVideo(ctx, video); err != nil {
		return err
	}
	return nil
}
func (s *videoServiceImpl) UpdateVideo(ctx context.Context, video *models.VideoInfo) error {
	cacheKey := fmt.Sprintf("VIDEO:%d", video.ID)
	err := s.videoRepo.UpdateVideo(ctx, video)
	if err != nil {
		return err
	}
	err = s.redisClient.Del(ctx, cacheKey).Err()
	if err != nil {
		log.Println("删除视频缓存失败:", err)
	}
	return nil
}

func (s *videoServiceImpl) FindVideoByID(ctx context.Context, videoID uint) (*models.VideoInfo, error) {
	cacheKey := fmt.Sprintf("VIDEO:%d", videoID)
	result, err := utils.GetCacheOrQuery(ctx, s.redisClient, cacheKey, func() (*models.VideoInfo, error) {
		video := &models.VideoInfo{}
		if err := s.videoRepo.FindVideoByID(ctx, video, videoID); err != nil {
			return nil, err
		}
		return video, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *videoServiceImpl) FindVideoByIDWithAuthor(ctx context.Context, videoID uint) (*models.VideoInfoWithAuthor, error) {
	cacheKey := fmt.Sprintf("VIDEO:%d:author", videoID)
	result, err := utils.GetCacheOrQuery(ctx, s.redisClient, cacheKey, func() (*models.VideoInfoWithAuthor, error) {
		return s.videoRepo.FindVideoByIDWithAuthor(ctx, videoID)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *videoServiceImpl) GetVideos(ctx context.Context, offset int, limit int) ([]models.VideoInfoWithAuthor, error) {
	videos := &[]models.VideoInfoWithAuthor{}
	if err := s.videoRepo.GetVideos(ctx, videos, offset, limit); err != nil {
		return nil, err
	}
	return *videos, nil
}

func (s *videoServiceImpl) DeleteVideo(ctx context.Context, id uint) error {
	cacheKey := fmt.Sprintf("VIDEO:%d", id)
	err := s.videoRepo.DeleteVideoByID(ctx, id)
	if err != nil {
		return err
	}
	err = s.redisClient.Del(ctx, cacheKey).Err()
	if err != nil {
		log.Println("删除视频缓存失败:", err)
	}
	return nil
}

func (s *videoServiceImpl) SearchVideoByTitle(ctx context.Context, title string, offset int, limit int) ([]models.VideoInfoWithAuthor, error) {
	var videos []models.VideoInfoWithAuthor
	if title == "" {
		return videos, nil
	}
	err := s.videoRepo.SearchVideoByTitle(ctx, &videos, title, offset, limit)
	if err != nil {
		return nil, err
	}
	return videos, nil
}

func (s *videoServiceImpl) UpdateVideoLike(ctx context.Context, userid uint, videoID uint) (bool, error) {
	ok, err := s.videoRepo.GetUserANDVideoLike(ctx, userid, videoID)
	if err != nil {
		return false, err
	}
	if ok == true {
		return false, s.videoRepo.DelVideoLike(ctx, userid, videoID)
	}
	err = s.videoRepo.AddVideoLike(ctx, userid, videoID)
	if err != nil {
		return false, err
	}
	return true, nil
}
