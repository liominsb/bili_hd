package service

import (
	"context"
	"go_bili/api/repository"
	"go_bili/models"

	"github.com/redis/go-redis/v9"
)

type CommentService interface {
	AddNewComment(ctx context.Context, comment *models.Comments) error
	DeleteComment(ctx context.Context, id uint) error
	GetCommentsById(ctx context.Context, videoId uint, offset int, limit int) (*[]models.CommentInfo, error)
	UpdateComment(ctx context.Context, id uint, content string) error
	FindCommentByID(ctx context.Context, id uint) (*models.Comments, error)
}

type commentServiceImpl struct {
	commentRepo repository.CommentRepository
	redisClient *redis.Client
}

func NewCommentService(commentRepo repository.CommentRepository, redisClient *redis.Client) CommentService {
	return &commentServiceImpl{commentRepo: commentRepo, redisClient: redisClient}
}

func (s *commentServiceImpl) AddNewComment(ctx context.Context, comment *models.Comments) error {
	return s.commentRepo.AddNewComment(ctx, comment)
}

func (s *commentServiceImpl) DeleteComment(ctx context.Context, id uint) error {
	return s.commentRepo.DeleteComment(ctx, id)
}

func (s *commentServiceImpl) GetCommentsById(ctx context.Context, videoId uint, offset int, limit int) (*[]models.CommentInfo, error) {
	comments := &[]models.CommentInfo{}
	if err := s.commentRepo.GetCommentsById(ctx, videoId, comments, offset, limit); err != nil {
		return nil, err
	}
	return comments, nil
}

func (s *commentServiceImpl) UpdateComment(ctx context.Context, id uint, content string) error {
	return s.commentRepo.UpdateComment(ctx, id, content)
}

func (s *commentServiceImpl) FindCommentByID(ctx context.Context, id uint) (*models.Comments, error) {
	return s.commentRepo.FindCommentByID(ctx, id)
}
