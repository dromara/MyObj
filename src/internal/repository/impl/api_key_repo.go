package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type apiKeyRepository struct {
	db *gorm.DB
}

// NewApiKeyRepository 创建API密钥仓储实例
func NewApiKeyRepository(db *gorm.DB) repository.ApiKeyRepository {
	return &apiKeyRepository{db: db}
}

func (r *apiKeyRepository) Create(ctx context.Context, apiKey *models.ApiKey) error {
	return r.db.WithContext(ctx).Create(apiKey).Error
}

func (r *apiKeyRepository) GetByID(ctx context.Context, id int) (*models.ApiKey, error) {
	var apiKey models.ApiKey
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *apiKeyRepository) GetByKey(ctx context.Context, key string) (*models.ApiKey, error) {
	var apiKey models.ApiKey
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *apiKeyRepository) Update(ctx context.Context, apiKey *models.ApiKey) error {
	return r.db.WithContext(ctx).Save(apiKey).Error
}

func (r *apiKeyRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.ApiKey{}).Error
}

func (r *apiKeyRepository) List(ctx context.Context, userID string, offset, limit int) ([]*models.ApiKey, error) {
	var apiKeys []*models.ApiKey
	query := r.db.WithContext(ctx).Offset(offset).Limit(limit)
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	err := query.Find(&apiKeys).Error
	return apiKeys, err
}

func (r *apiKeyRepository) Count(ctx context.Context, userID string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.ApiKey{})
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	err := query.Count(&count).Error
	return count, err
}
