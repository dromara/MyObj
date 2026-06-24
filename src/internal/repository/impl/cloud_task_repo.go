package impl

import (
	"context"
	"myobj/src/pkg/models"
	"time"

	"gorm.io/gorm"
)

// CloudTaskRepository 云盘任务仓储
type CloudTaskRepository struct {
	db *gorm.DB
}

// NewCloudTaskRepository 创建云盘任务仓储
func NewCloudTaskRepository(db *gorm.DB) *CloudTaskRepository {
	return &CloudTaskRepository{db: db}
}

// Create 创建任务
func (r *CloudTaskRepository) Create(ctx context.Context, task *models.CloudTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// GetByID 根据ID获取任务
func (r *CloudTaskRepository) GetByID(ctx context.Context, id int) (*models.CloudTask, error) {
	var task models.CloudTask
	err := r.db.WithContext(ctx).First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Update 更新任务
func (r *CloudTaskRepository) Update(ctx context.Context, task *models.CloudTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// UpdateStatus 更新任务状态
func (r *CloudTaskRepository) UpdateStatus(ctx context.Context, id int, status int, errorMsg string) error {
	updates := map[string]interface{}{
		"status":     status,
		"error_msg":  errorMsg,
		"updated_at": time.Now(),
	}
	if status == models.CloudTaskStatusCompleted || status == models.CloudTaskStatusFailed {
		now := time.Now()
		updates["completed_at"] = &now
	}
	return r.db.WithContext(ctx).Model(&models.CloudTask{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateProgress 更新任务进度
func (r *CloudTaskRepository) UpdateProgress(ctx context.Context, id int, successCount, failedCount int) error {
	return r.db.WithContext(ctx).Model(&models.CloudTask{}).Where("id = ?", id).Updates(map[string]interface{}{
		"success_count": successCount,
		"failed_count":  failedCount,
		"updated_at":    time.Now(),
	}).Error
}

// ListByUserID 获取用户的任务列表
func (r *CloudTaskRepository) ListByUserID(ctx context.Context, userID string, page, pageSize int) ([]*models.CloudTask, int64, error) {
	var tasks []*models.CloudTask
	var total int64

	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	// 获取总数
	if err := query.Model(&models.CloudTask{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// Delete 删除任务
func (r *CloudTaskRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.CloudTask{}, id).Error
}

// CloudTaskFileRepository 云盘任务文件仓储
type CloudTaskFileRepository struct {
	db *gorm.DB
}

// NewCloudTaskFileRepository 创建云盘任务文件仓储
func NewCloudTaskFileRepository(db *gorm.DB) *CloudTaskFileRepository {
	return &CloudTaskFileRepository{db: db}
}

// Create 创建任务文件
func (r *CloudTaskFileRepository) Create(ctx context.Context, file *models.CloudTaskFile) error {
	return r.db.WithContext(ctx).Create(file).Error
}

// BatchCreate 批量创建任务文件
func (r *CloudTaskFileRepository) BatchCreate(ctx context.Context, files []*models.CloudTaskFile) error {
	return r.db.WithContext(ctx).CreateInBatches(files, 100).Error
}

// GetByID 根据ID获取任务文件
func (r *CloudTaskFileRepository) GetByID(ctx context.Context, id int) (*models.CloudTaskFile, error) {
	var file models.CloudTaskFile
	err := r.db.WithContext(ctx).First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// Update 更新任务文件
func (r *CloudTaskFileRepository) Update(ctx context.Context, file *models.CloudTaskFile) error {
	return r.db.WithContext(ctx).Save(file).Error
}

// UpdateStatus 更新任务文件状态
func (r *CloudTaskFileRepository) UpdateStatus(ctx context.Context, id int, status int, errorMsg string) error {
	return r.db.WithContext(ctx).Model(&models.CloudTaskFile{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"error_msg":  errorMsg,
		"updated_at": time.Now(),
	}).Error
}

// UpdateLocalPath 更新本地路径
func (r *CloudTaskFileRepository) UpdateLocalPath(ctx context.Context, id int, localPath, localFileID string) error {
	return r.db.WithContext(ctx).Model(&models.CloudTaskFile{}).Where("id = ?", id).Updates(map[string]interface{}{
		"local_path":    localPath,
		"local_file_id": localFileID,
		"status":        models.CloudFileStatusCompleted,
		"updated_at":    time.Now(),
	}).Error
}

// ListByTaskID 获取任务的文件列表
func (r *CloudTaskFileRepository) ListByTaskID(ctx context.Context, taskID int) ([]*models.CloudTaskFile, error) {
	var files []*models.CloudTaskFile
	err := r.db.WithContext(ctx).Where("task_id = ?", taskID).Order("is_dir DESC, file_name ASC").Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

// ListPendingByTaskID 获取任务的待处理文件列表
func (r *CloudTaskFileRepository) ListPendingByTaskID(ctx context.Context, taskID int) ([]*models.CloudTaskFile, error) {
	var files []*models.CloudTaskFile
	err := r.db.WithContext(ctx).Where("task_id = ? AND status = ?", taskID, models.CloudFileStatusPending).Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

// CountByStatus 统计任务文件状态
func (r *CloudTaskFileRepository) CountByStatus(ctx context.Context, taskID int) (map[int]int64, error) {
	var results []struct {
		Status int
		Count  int64
	}
	err := r.db.WithContext(ctx).Model(&models.CloudTaskFile{}).Where("task_id = ?", taskID).Select("status, count(*) as count").Group("status").Find(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[int]int64)
	for _, r := range results {
		counts[r.Status] = r.Count
	}
	return counts, nil
}
