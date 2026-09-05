package service

import (
	"context"
	"errors"
	"go_bili/api/repository"
	"go_bili/models"

	"gorm.io/gorm"
)

type FollowService interface {
	Follow(ctx context.Context, followerID uint, followingID uint) error
	Unfollow(ctx context.Context, followerID uint, followingID uint) error
	IsFollowing(ctx context.Context, followerID uint, followingID uint) (bool, error)
	CountFollowing(ctx context.Context, userID uint) (int64, error)
	CountFollowers(ctx context.Context, userID uint) (int64, error)
	ListFollowing(ctx context.Context, userID uint, offset int, limit int) (*[]models.UserBrief, error)
	ListFollowers(ctx context.Context, userID uint, offset int, limit int) (*[]models.UserBrief, error)
}

type followServiceImpl struct {
	followRepo repository.FollowRepository
	authRepo   repository.AuthRepository
}

func NewFollowService(followRepo repository.FollowRepository, authRepo repository.AuthRepository) FollowService {
	return &followServiceImpl{followRepo: followRepo, authRepo: authRepo}
}

// Follow 关注：校验不能关注自己 + 目标用户存在
func (s *followServiceImpl) Follow(ctx context.Context, followerID uint, followingID uint) error {
	if followerID == followingID {
		return errors.New("不能关注自己")
	}
	if err := s.checkUserExists(ctx, followingID); err != nil {
		return err
	}
	return s.followRepo.Follow(ctx, followerID, followingID)
}

func (s *followServiceImpl) Unfollow(ctx context.Context, followerID uint, followingID uint) error {
	return s.followRepo.Unfollow(ctx, followerID, followingID)
}

func (s *followServiceImpl) IsFollowing(ctx context.Context, followerID uint, followingID uint) (bool, error) {
	return s.followRepo.IsFollowing(ctx, followerID, followingID)
}

func (s *followServiceImpl) CountFollowing(ctx context.Context, userID uint) (int64, error) {
	return s.followRepo.CountFollowing(ctx, userID)
}

func (s *followServiceImpl) CountFollowers(ctx context.Context, userID uint) (int64, error) {
	return s.followRepo.CountFollowers(ctx, userID)
}

func (s *followServiceImpl) ListFollowing(ctx context.Context, userID uint, offset int, limit int) (*[]models.UserBrief, error) {
	list := &[]models.UserBrief{}
	if err := s.followRepo.ListFollowing(ctx, userID, list, offset, limit); err != nil {
		return nil, err
	}
	return list, nil
}

func (s *followServiceImpl) ListFollowers(ctx context.Context, userID uint, offset int, limit int) (*[]models.UserBrief, error) {
	list := &[]models.UserBrief{}
	if err := s.followRepo.ListFollowers(ctx, userID, list, offset, limit); err != nil {
		return nil, err
	}
	return list, nil
}

// checkUserExists 校验目标用户是否存在，复用 authRepo 的查询方法
func (s *followServiceImpl) checkUserExists(ctx context.Context, userID uint) error {
	var user models.User
	err := s.authRepo.GetUserByID(ctx, &user, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("目标用户不存在")
	}
	return err
}
