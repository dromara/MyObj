package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"
	"time"

	"gorm.io/gorm"
)

type uploadTaskRepository struct {
	db *gorm.DB
}

// NewUploadTaskRepository 创建上传任务仓储实例
func NewUploadTaskRepository(db *gorm.DB) repository.UploadTaskRepository {
	return &uploadTaskRepository{db: db}
}

// Create 创建上传任务记录
func (r *uploadTaskRepository) Create(ctx context.Context, task *models.UploadTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// GetByID 根据ID获取上传任务
func (r *uploadTaskRepository) GetByID(ctx context.Context, id string) (*models.UploadTask, error) {
	var task models.UploadTask
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByUserID 根据用户ID获取所有上传任务
func (r *uploadTaskRepository) GetByUserID(ctx context.Context, userID string) ([]*models.UploadTask, error) {
	var tasks []*models.UploadTask
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("create_time DESC").Find(&tasks).Error
	return tasks, err
}

// GetUncompletedByUserID 根据用户ID获取未完成的上传任务
func (r *uploadTaskRepository) GetUncompletedByUserID(ctx context.Context, userID string) ([]*models.UploadTask, error) {
	var tasks []*models.UploadTask
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status IN (?) AND expire_time > ?", userID, []string{"pending", "uploading"}, now).
		Order("create_time DESC").Find(&tasks).Error
	return tasks, err
}

// Update 更新上传任务
func (r *uploadTaskRepository) Update(ctx context.Context, task *models.UploadTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// Delete 删除上传任务
func (r *uploadTaskRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.UploadTask{}).Error
}

// DeleteExpired 删除过期的上传任务（所有用户）
func (r *uploadTaskRepository) DeleteExpired(ctx context.Context) (int64, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Where("expire_time < ? AND status IN (?)", now, []string{"pending", "uploading", "aborted"}).
		Delete(&models.UploadTask{})
	return result.RowsAffected, result.Error
}

// DeleteExpiredByUserID 删除指定用户的过期上传任务
func (r *uploadTaskRepository) DeleteExpiredByUserID(ctx context.Context, userID string) (int64, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND expire_time < ? AND status IN (?)", userID, now, []string{"pending", "uploading", "aborted"}).
		Delete(&models.UploadTask{})
	return result.RowsAffected, result.Error
}

