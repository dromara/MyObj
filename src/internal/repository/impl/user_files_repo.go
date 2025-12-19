package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type userFilesRepository struct {
	db *gorm.DB
}

// NewUserFilesRepository 创建用户文件关联仓储实例
func NewUserFilesRepository(db *gorm.DB) repository.UserFilesRepository {
	return &userFilesRepository{db: db}
}

func (r *userFilesRepository) Create(ctx context.Context, userFile *models.UserFiles) error {
	return r.db.WithContext(ctx).Create(userFile).Error
}

func (r *userFilesRepository) GetByUserIDAndFileID(ctx context.Context, userID, fileID string) (*models.UserFiles, error) {
	var userFile models.UserFiles
	err := r.db.WithContext(ctx).Where("user_id = ? AND file_id = ?", userID, fileID).First(&userFile).Error
	if err != nil {
		return nil, err
	}
	return &userFile, nil
}

func (r *userFilesRepository) Update(ctx context.Context, userFile *models.UserFiles) error {
	return r.db.WithContext(ctx).Where("user_id = ? and file_id = ?", userFile.UserID, userFile.FileID).Save(userFile).Error
}

func (r *userFilesRepository) Delete(ctx context.Context, userID, fileID string) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND file_id = ?", userID, fileID).
		Delete(&models.UserFiles{}).Error
}

func (r *userFilesRepository) ListByUserID(ctx context.Context, userID string, offset, limit int) ([]*models.UserFiles, error) {
	var userFiles []*models.UserFiles
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Offset(offset).Limit(limit).Find(&userFiles).Error
	return userFiles, err
}

func (r *userFilesRepository) Count(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.UserFiles{}).
		Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// ListPublicFiles 获取所有公开文件
func (r *userFilesRepository) ListPublicFiles(ctx context.Context, offset, limit int) ([]*models.UserFiles, error) {
	var userFiles []*models.UserFiles
	err := r.db.WithContext(ctx).Where("public = ?", true).
		Offset(offset).Limit(limit).Find(&userFiles).Error
	return userFiles, err
}

// CountPublicFiles 统计公开文件数量
func (r *userFilesRepository) CountPublicFiles(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.UserFiles{}).
		Where("public = ?", true).Count(&count).Error
	return count, err
}

// SearchPublicFiles 搜索公开文件（根据文件名）
func (r *userFilesRepository) SearchPublicFiles(ctx context.Context, keyword string, offset, limit int) ([]*models.UserFiles, error) {
	var userFiles []*models.UserFiles
	err := r.db.WithContext(ctx).
		Joins("JOIN file_info ON user_files.file_id = file_info.id").
		Where("user_files.public = ? AND file_info.name LIKE ?", true, "%"+keyword+"%").
		Offset(offset).Limit(limit).
		Find(&userFiles).Error
	return userFiles, err
}

// CountPublicFilesByKeyword 统计匹配关键词的公开文件数量
func (r *userFilesRepository) CountPublicFilesByKeyword(ctx context.Context, keyword string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.UserFiles{}).
		Joins("JOIN file_info ON user_files.file_id = file_info.id").
		Where("user_files.public = ? AND file_info.name LIKE ?", true, "%"+keyword+"%").
		Count(&count).Error
	return count, err
}

// SearchUserFiles 搜索用户文件（根据文件名）
func (r *userFilesRepository) SearchUserFiles(ctx context.Context, userID, keyword string, offset, limit int) ([]*models.UserFiles, error) {
	var userFiles []*models.UserFiles
	err := r.db.WithContext(ctx).
		Joins("JOIN file_info ON user_files.file_id = file_info.id").
		Where("user_files.user_id = ? AND file_info.name LIKE ?", userID, "%"+keyword+"%").
		Offset(offset).Limit(limit).
		Find(&userFiles).Error
	return userFiles, err
}

// CountUserFilesByKeyword 统计用户匹配关键词的文件数量
func (r *userFilesRepository) CountUserFilesByKeyword(ctx context.Context, userID, keyword string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.UserFiles{}).
		Joins("JOIN file_info ON user_files.file_id = file_info.id").
		Where("user_files.user_id = ? AND file_info.name LIKE ?", userID, "%"+keyword+"%").
		Count(&count).Error
	return count, err
}

// GetByUserIDAndUfID 获取用户文件关联
func (r *userFilesRepository) GetByUserIDAndUfID(ctx context.Context, userID, ufID string) (*models.UserFiles, error) {
	var userFile models.UserFiles
	err := r.db.WithContext(ctx).Where("user_id = ? AND uf_id = ?", userID, ufID).First(&userFile).Error
	if err != nil {
		return nil, err
	}
	return &userFile, nil
}
