package repository

import (
	"context"
	"go_bili/models"

	"gorm.io/gorm"
)

type CommentRepository interface {
	AddNewComment(ctx context.Context, comment *models.Comments) error
	DeleteComment(ctx context.Context, id uint) error
	GetCommentsById(ctx context.Context, videoId uint, comments *[]models.CommentInfo, offset int, limit int) error
	UpdateComment(ctx context.Context, id uint, content string) error
	FindCommentByID(ctx context.Context, id uint) (*models.Comments, error)
}
type commentRepoImpl struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepoImpl{db: db}
}

func (r *commentRepoImpl) AddNewComment(ctx context.Context, comment *models.Comments) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *commentRepoImpl) DeleteComment(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Comments{}, "id = ?", id).Error
}

func (r *commentRepoImpl) GetCommentsById(ctx context.Context, videoId uint, comments *[]models.CommentInfo, offset int, limit int) error {
	return r.db.WithContext(ctx).
		Table("comments").
		Select("comments.*, users.username AS author_name, users.image AS author_image").
		Joins("LEFT JOIN users ON users.id = comments.author_id").
		Where("comments.video_id = ?", videoId).
		Order("comments.created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(comments).Error
}

func (r *commentRepoImpl) UpdateComment(ctx context.Context, id uint, content string) error {
	return r.db.WithContext(ctx).Model(models.Comments{}).Where("id  = ?", id).Update("content", content).Error
}

func (r *commentRepoImpl) FindCommentByID(ctx context.Context, id uint) (*models.Comments, error) {
	comment := &models.Comments{}
	if err := r.db.WithContext(ctx).First(comment, id).Error; err != nil {
		return nil, err
	}
	return comment, nil
}
