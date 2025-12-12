package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type sysConfigRepository struct {
	db *gorm.DB
}

// NewSysConfigRepository 创建系统配置仓储实例
func NewSysConfigRepository(db *gorm.DB) repository.SysConfigRepository {
	return &sysConfigRepository{db: db}
}

// Create 创建系统配置
func (r *sysConfigRepository) Create(ctx context.Context, config *models.SysConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// GetByID 根据ID获取系统配置
func (r *sysConfigRepository) GetByID(ctx context.Context, id int) (*models.SysConfig, error) {
	var config models.SysConfig
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetByKey 根据配置键获取系统配置
func (r *sysConfigRepository) GetByKey(ctx context.Context, key string) (*models.SysConfig, error) {
	var config models.SysConfig
	err := r.db.WithContext(ctx).Where("`key` = ?", key).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Update 更新系统配置
func (r *sysConfigRepository) Update(ctx context.Context, config *models.SysConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// Delete 删除系统配置
func (r *sysConfigRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.SysConfig{}).Error
}

// List 获取系统配置列表
func (r *sysConfigRepository) List(ctx context.Context, offset, limit int) ([]*models.SysConfig, error) {
	var configs []*models.SysConfig
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&configs).Error
	return configs, err
}

// Count 统计系统配置数量
func (r *sysConfigRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.SysConfig{}).Count(&count).Error
	return count, err
}

// BatchUpdate 批量更新配置
func (r *sysConfigRepository) BatchUpdate(ctx context.Context, configs []*models.SysConfig) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, config := range configs {
			if err := tx.Save(config).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetAllAsMap 获取所有配置并以 key-value 格式返回
func (r *sysConfigRepository) GetAllAsMap(ctx context.Context) (map[string]string, error) {
	var configs []*models.SysConfig
	err := r.db.WithContext(ctx).Find(&configs).Error
	if err != nil {
		return nil, err
	}

	configMap := make(map[string]string, len(configs))
	for _, config := range configs {
		configMap[config.Key] = config.Value
	}
	return configMap, nil
}
