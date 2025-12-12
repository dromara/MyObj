package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type virtualPathRepository struct {
	db *gorm.DB
}

// NewVirtualPathRepository 创建虚拟路径仓储实例
func NewVirtualPathRepository(db *gorm.DB) repository.VirtualPathRepository {
	return &virtualPathRepository{db: db}
}

func (r *virtualPathRepository) Create(ctx context.Context, vpath *models.VirtualPath) error {
	return r.db.WithContext(ctx).Create(vpath).Error
}

func (r *virtualPathRepository) GetByID(ctx context.Context, id int) (*models.VirtualPath, error) {
	var vpath models.VirtualPath
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&vpath).Error
	if err != nil {
		return nil, err
	}
	return &vpath, nil
}

func (r *virtualPathRepository) GetByPath(ctx context.Context, userID, path string) (*models.VirtualPath, error) {
	var vpath models.VirtualPath
	err := r.db.WithContext(ctx).Where("user_id = ? AND path = ?", userID, path).First(&vpath).Error
	if err != nil {
		return nil, err
	}
	return &vpath, nil
}

func (r *virtualPathRepository) Update(ctx context.Context, vpath *models.VirtualPath) error {
	return r.db.WithContext(ctx).Save(vpath).Error
}

func (r *virtualPathRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.VirtualPath{}).Error
}

func (r *virtualPathRepository) ListByUserID(ctx context.Context, userID string, offset, limit int) ([]*models.VirtualPath, error) {
	var vpaths []*models.VirtualPath
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Offset(offset).Limit(limit).Find(&vpaths).Error
	return vpaths, err
}

func (r *virtualPathRepository) Count(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.VirtualPath{}).
		Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// ListSubFoldersByParentID 查询指定父目录ID下的子目录
func (r *virtualPathRepository) ListSubFoldersByParentID(ctx context.Context, userID string, parentID int, offset, limit int) ([]*models.VirtualPath, error) {
	var vpaths []*models.VirtualPath
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND parent_level = ? AND is_dir = ?", userID, parentID, true).
		Order("path ASC").
		Offset(offset).Limit(limit).
		Find(&vpaths).Error
	return vpaths, err
}

// CountSubFoldersByParentID 统计指定父目录ID下的子目录数量
func (r *virtualPathRepository) CountSubFoldersByParentID(ctx context.Context, userID string, parentID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.VirtualPath{}).
		Where("user_id = ? AND parent_level = ? AND is_dir = ?", userID, parentID, true).
		Count(&count).Error
	return count, err
}

// GetRootPath 获取用户根目录
func (r *virtualPathRepository) GetRootPath(ctx context.Context, userID string) (*models.VirtualPath, error) {
	var vpath models.VirtualPath
	// 根目录的parent_level为NULL或空字符串
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND (parent_level IS NULL OR parent_level = '')", userID).
		First(&vpath).Error
	if err != nil {
		return nil, err
	}
	return &vpath, nil
}

func (r *virtualPathRepository) GetPathByUser(ctx context.Context, userID string) ([]*models.VirtualPath, error) {
	var vpaths []*models.VirtualPath
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&vpaths).Error
	return vpaths, err
}
