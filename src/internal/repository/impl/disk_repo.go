package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type diskRepository struct {
	db *gorm.DB
}

// NewDiskRepository 创建磁盘仓储实例
func NewDiskRepository(db *gorm.DB) repository.DiskRepository {
	return &diskRepository{db: db}
}

func (r *diskRepository) Create(ctx context.Context, disk *models.Disk) error {
	return r.db.WithContext(ctx).Create(disk).Error
}

func (r *diskRepository) GetByID(ctx context.Context, id string) (*models.Disk, error) {
	var disk models.Disk
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&disk).Error
	if err != nil {
		return nil, err
	}
	return &disk, nil
}

func (r *diskRepository) GetBigDisk(ctx context.Context) (*models.Disk, error) {
	var disk models.Disk
	err := r.db.WithContext(ctx).Order("size desc").Limit(1).First(&disk).Error
	if err != nil {
		return nil, err
	}
	return &disk, nil
}

func (r *diskRepository) GetByPath(ctx context.Context, path string) (*models.Disk, error) {
	var disk models.Disk
	err := r.db.WithContext(ctx).Where("disk_path = ?", path).First(&disk).Error
	if err != nil {
		return nil, err
	}
	return &disk, nil
}

func (r *diskRepository) Update(ctx context.Context, disk *models.Disk) error {
	return r.db.WithContext(ctx).Save(disk).Error
}

func (r *diskRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Disk{}).Error
}

func (r *diskRepository) List(ctx context.Context, offset, limit int) ([]*models.Disk, error) {
	var disks []*models.Disk
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&disks).Error
	return disks, err
}

func (r *diskRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Disk{}).Count(&count).Error
	return count, err
}
