package service

import (
	"context"
	"errors"
	"fmt"
	"go_bili/api/repository"
	"go_bili/models"
	"go_bili/utils"
	"log"
	"strconv"
	"strings"

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
	SyncViewCounts(ctx context.Context)
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
	// 一条命令干两件事：记一次播放 + 拿到还没落库的增量（INCR 的返回值就是自增后的值）
	// 展示校准只治"Redis 里还没搬走的 ≤10 秒"；缓存对象岁数里的已落库量治不了，可接受
	// Incr 失败（Redis 挂）则降级：不记数、不叠加，视频照常返回——
	// 跟 GetCacheOrQuery 的降级哲学一致，不能让 Redis 把详情页拖成 500
	delta, err := s.redisClient.Incr(ctx, "video:views:"+strconv.FormatUint(uint64(result.ID), 10)).Result()
	if err == nil {
		result.ViewCount += int(delta)
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

func (s *videoServiceImpl) SyncViewCounts(ctx context.Context) {
	keys, err := s.redisClient.Keys(ctx, "video:views:*").Result()
	if err != nil {
		log.Println("获取视频浏览量缓存失败:", err)
		return
	}

	for _, key := range keys {
		// "video:views:3" 切三段取最后一段，反解出 videoID
		videoID, err := strconv.ParseUint(strings.Split(key, ":")[2], 10, 64)
		if err != nil {
			log.Println(err)
			continue // key 格式不对就跳过
		}
		valstr, err := s.redisClient.GetDel(ctx, key).Result()
		if errors.Is(err, redis.Nil) {
			continue // key 已被取走/不存在，正常情况
		}
		if err != nil {
			log.Println(err)
			continue
		}
		val, err := strconv.Atoi(valstr)
		if err != nil {
			log.Println(err)
			continue
		}
		err = s.videoRepo.SyncViewCounts(ctx, uint(videoID), val)
		if err != nil {
			log.Println(err)
			s.redisClient.IncrBy(context.Background(), key, int64(val))
			continue
		}
	}
	return
}
