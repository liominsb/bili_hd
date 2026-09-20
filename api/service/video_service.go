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
	"sync"
	"sync/atomic"

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
	IsLiked(ctx context.Context, userid uint, videoID uint) (bool, error)
	GetLikeCount(ctx context.Context, videoID uint) (int, error)
	SyncViewCounts(ctx context.Context)
	FlushViewDelta()
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

type viewCounter struct {
	total   atomic.Int64 // 本进程累计，只增不清，给页面显示用
	pending atomic.Int64 // 待刷入 Redis，flush 后清零
}

var viewDelta sync.Map

func addViewDelta(videoID uint) int64 {
	v, _ := viewDelta.LoadOrStore(videoID, &viewCounter{})
	t := v.(*viewCounter)
	t.pending.Add(1)
	return t.total.Add(1)
}

func (s *videoServiceImpl) FlushViewDelta() {
	viewDelta.Range(func(k, v any) bool {
		videoID := k.(uint)
		t := v.(*viewCounter)
		delta := t.pending.Swap(0)
		if delta > 0 {
			s.redisClient.IncrBy(context.Background(),
				"video:views:"+strconv.FormatUint(uint64(videoID), 10), delta)
		}
		return true
	})
}

// FindVideoByIDWithAuthor 按视频ID查找，并带上作者信息
func (s *videoServiceImpl) FindVideoByIDWithAuthor(ctx context.Context, videoID uint) (*models.VideoInfoWithAuthor, error) {
	cacheKey := fmt.Sprintf("VIDEO:%d:author", videoID)
	//展示滞后被缓存放大：详情走 GetCacheOrQuery，整个 VideoInfo 连 view_count 一起缓存 10~70 分钟。
	//所以页面上看到的播放数最多滞后一小时，10 秒同步的及时性全被这层缓存吃了。
	result, err := utils.GetCacheOrQuery(ctx, s.redisClient, cacheKey, func() (*models.VideoInfoWithAuthor, error) {
		return s.videoRepo.FindVideoByIDWithAuthor(ctx, videoID)
	})
	if err != nil || result == nil {
		return nil, err
	}
	//delta, err := s.redisClient.Incr(ctx, "video:views:"+strconv.FormatUint(uint64(result.ID), 10)).Result()
	//if err == nil {
	//	result.ViewCount += int(delta)
	//}

	result.ViewCount += int(addViewDelta(videoID))
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

func (s *videoServiceImpl) IsLiked(ctx context.Context, userid uint, videoID uint) (bool, error) {
	return s.videoRepo.GetUserANDVideoLike(ctx, userid, videoID)
}

func (s *videoServiceImpl) GetLikeCount(ctx context.Context, videoID uint) (int, error) {
	return s.videoRepo.GetVideoLikeCount(ctx, videoID)
}

// SyncViewCounts redis->sql
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
