package repository

import (
	"context"
	"go_bili/models"

	"gorm.io/gorm"
)

type OAuthRepository interface {
	GetBinding(ctx context.Context, binding *models.OAuthBinding, provider string, providerUID string) error
	CreateBinding(ctx context.Context, binding *models.OAuthBinding) error
}

type oauthRepoImpl struct {
	db *gorm.DB
}

func NewOAuthRepository(db *gorm.DB) OAuthRepository {
	return &oauthRepoImpl{db: db}
}

// GetBinding 按「平台 + 平台数字id」查绑定关系，查不到返回 gorm.ErrRecordNotFound
func (r *oauthRepoImpl) GetBinding(ctx context.Context, binding *models.OAuthBinding, provider string, providerUID string) error {
	return r.db.WithContext(ctx).
		Where("provider = ? AND provider_uid = ?", provider, providerUID).
		First(binding).Error
}

func (r *oauthRepoImpl) CreateBinding(ctx context.Context, binding *models.OAuthBinding) error {
	return r.db.WithContext(ctx).Create(binding).Error
}
