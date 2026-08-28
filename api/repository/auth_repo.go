package repository

import (
	"context"
	"go_bili/models"

	"gorm.io/gorm"
)

type AuthRepository interface {
	Register(ctx context.Context, user *models.User) error
	GetUserByUsername(ctx context.Context, user *models.User, username string) error
	GetUserByID(ctx context.Context, user *models.User, userID uint) error
	UpdatePassword(ctx context.Context, userID uint, hashedPassword string) error
	UpdateProfile(ctx context.Context, userID uint, username string, image string, bio string) error
}

type authRepoImpl struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepoImpl{db: db}
}

func (r *authRepoImpl) Register(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *authRepoImpl) GetUserByUsername(ctx context.Context, user *models.User, username string) error {
	return r.db.WithContext(ctx).Where("username = ?", username).First(user).Error
}

// 按ID获取用户
func (r *authRepoImpl) GetUserByID(ctx context.Context, user *models.User, userID uint) error {
	return r.db.WithContext(ctx).Where("id = ?", userID).First(user).Error
}

func (r *authRepoImpl) UpdatePassword(ctx context.Context, userID uint, hashedPassword string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

func (r *authRepoImpl) UpdateProfile(ctx context.Context, userID uint, username string, image string, bio string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"username": username,
		"image":    image,
		"bio":      bio,
	}).Error
}
