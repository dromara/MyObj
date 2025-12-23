package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type downloadTaskRepository struct {
	db *gorm.DB
}

// NewDownloadTaskRepository 创建下载任务仓储实例
func NewDownloadTaskRepository(db *gorm.DB) repository.DownloadTaskRepository {
	return &downloadTaskRepository{db: db}
}

func (r *downloadTaskRepository) Create(ctx context.Context, task *models.DownloadTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *downloadTaskRepository) GetByID(ctx context.Context, id string) (*models.DownloadTask, error) {
	var task models.DownloadTask
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *downloadTaskRepository) Update(ctx context.Context, task *models.DownloadTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *downloadTaskRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.DownloadTask{}).Error
}

func (r *downloadTaskRepository) ListByUserID(ctx context.Context, userID string, offset, limit int) ([]*models.DownloadTask, error) {
	var tasks []*models.DownloadTask
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("create_time DESC").
		Offset(offset).
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

func (r *downloadTaskRepository) Count(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.DownloadTask{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *downloadTaskRepository) ListByState(ctx context.Context, userID string, state int, offset, limit int) ([]*models.DownloadTask, error) {
	var tasks []*models.DownloadTask
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND state = ?", userID, state).
		Order("create_time DESC").
		Offset(offset).
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

func (r *downloadTaskRepository) CountByState(ctx context.Context, userID string, state int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.DownloadTask{}).
		Where("user_id = ? AND state = ?", userID, state).
		Count(&count).Error
	return count, err
}

func (r *downloadTaskRepository) ListByType(ctx context.Context, userID string, taskType int, offset, limit int) ([]*models.DownloadTask, error) {
	var tasks []*models.DownloadTask
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, taskType).
		Order("create_time DESC").
		Offset(offset).
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

func (r *downloadTaskRepository) CountByType(ctx context.Context, userID string, taskType int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.DownloadTask{}).
		Where("user_id = ? AND type = ?", userID, taskType).
		Count(&count).Error
	return count, err
}

func (r *downloadTaskRepository) ListByStateAndType(ctx context.Context, userID string, state int, taskType int, offset, limit int) ([]*models.DownloadTask, error) {
	var tasks []*models.DownloadTask
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND state = ? AND type = ?", userID, state, taskType).
		Order("create_time DESC").
		Offset(offset).
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

func (r *downloadTaskRepository) CountByStateAndType(ctx context.Context, userID string, state int, taskType int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.DownloadTask{}).
		Where("user_id = ? AND state = ? AND type = ?", userID, state, taskType).
		Count(&count).Error
	return count, err
}