package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"
	"myobj/src/pkg/util"

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
	return r.db.WithContext(ctx).Where("user_id = ? AND uf_id = ?", userID, fileID).
		Delete(&models.UserFiles{}).Error
}

func (r *userFilesRepository) ListByUserID(ctx context.Context, userID string, offset, limit int) ([]*models.UserFiles, error) {
	var userFiles []*models.UserFiles
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").
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
	err := r.db.WithContext(ctx).Where("is_public = ?", true).
		Order("created_at DESC").
		Offset(offset).Limit(limit).Find(&userFiles).Error
	return userFiles, err
}

// CountPublicFiles 统计公开文件数量
func (r *userFilesRepository) CountPublicFiles(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.UserFiles{}).
		Where("is_public = ?", true).Count(&count).Error
	return count, err
}

// SearchPublicFiles 搜索公开文件（根据文件名）
func (r *userFilesRepository) SearchPublicFiles(ctx context.Context, keyword string, offset, limit int) ([]*models.UserFiles, error) {
	var userFiles []*models.UserFiles
	safeKeyword := util.EscapeLikeKeyword(keyword)
	err := r.db.WithContext(ctx).
		Joins("JOIN file_info ON user_files.file_id = file_info.id").
		Where("user_files.is_public = ? AND file_info.name LIKE ?", true, "%"+safeKeyword+"%").
		Offset(offset).Limit(limit).
		Find(&userFiles).Error
	return userFiles, err
}

// CountPublicFilesByKeyword 统计匹配关键词的公开文件数量
func (r *userFilesRepository) CountPublicFilesByKeyword(ctx context.Context, keyword string) (int64, error) {
	var count int64
	safeKeyword := util.EscapeLikeKeyword(keyword)
	err := r.db.WithContext(ctx).Model(&models.UserFiles{}).
		Joins("JOIN file_info ON user_files.file_id = file_info.id").
		Where("user_files.is_public = ? AND file_info.name LIKE ?", true, "%"+safeKeyword+"%").
		Count(&count).Error
	return count, err
}

// SearchUserFiles 搜索用户文件（根据文件名）
func (r *userFilesRepository) SearchUserFiles(ctx context.Context, userID, keyword string, offset, limit int) ([]*models.UserFiles, error) {
	var userFiles []*models.UserFiles
	safeKeyword := util.EscapeLikeKeyword(keyword)
	err := r.db.WithContext(ctx).
		Joins("JOIN file_info ON user_files.file_id = file_info.id").
		Where("user_files.user_id = ? AND user_files.file_name LIKE ?", userID, "%"+safeKeyword+"%").
		Offset(offset).Limit(limit).
		Find(&userFiles).Error
	return userFiles, err
}

// CountUserFilesByKeyword 统计用户匹配关键词的文件数量
func (r *userFilesRepository) CountUserFilesByKeyword(ctx context.Context, userID, keyword string) (int64, error) {
	var count int64
	safeKeyword := util.EscapeLikeKeyword(keyword)
	err := r.db.WithContext(ctx).Model(&models.UserFiles{}).
		Joins("JOIN file_info ON user_files.file_id = file_info.id").
		Where("user_files.user_id = ? AND file_info.name LIKE ?", userID, "%"+safeKeyword+"%").
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

// GetByUfID 通过 uf_id 查询文件（用于公开文件访问，不要求 user_id）
func (r *userFilesRepository) GetByUfID(ctx context.Context, ufID string) (*models.UserFiles, error) {
	var userFile models.UserFiles
	err := r.db.WithContext(ctx).Where("uf_id = ?", ufID).First(&userFile).Error
	if err != nil {
		return nil, err
	}
	return &userFile, nil
}

// GetByFileID 通过 file_id 查询任意一条记录（用于公开文件权限校验）
func (r *userFilesRepository) GetByFileID(ctx context.Context, fileID string) (*models.UserFiles, error) {
	var userFile models.UserFiles
	err := r.db.WithContext(ctx).Where("file_id = ?", fileID).First(&userFile).Error
	if err != nil {
		return nil, err
	}
	return &userFile, nil
}

// ListByVirtualPath 查询指定虚拟路径下的user_files记录（避免file_id重复问题）
// 直接从 user_files 表查询，每个uf_id都是唯一的，避免了秒传场景下同一file_id有多条记录的问题
func (r *userFilesRepository) ListByVirtualPath(ctx context.Context, userID, virtualPath string, offset, limit int) ([]*models.UserFiles, error) {
	var userFiles []*models.UserFiles
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND virtual_path = ?", userID, virtualPath).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&userFiles).Error
	return userFiles, err
}

// GetByUserIDAndFileIDs 批量查询指定用户和多个fileID的关联记录
func (r *userFilesRepository) GetByUserIDAndFileIDs(ctx context.Context, userID string, fileIDs []string) ([]*models.UserFiles, error) {
	if len(fileIDs) == 0 {
		return nil, nil
	}
	var userFiles []*models.UserFiles
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND file_id IN ?", userID, fileIDs).
		Find(&userFiles).Error
	return userFiles, err
}

// DeleteByUserID 删除指定用户的所有文件关联记录
func (r *userFilesRepository) DeleteByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.UserFiles{}).Error
}

// ExistsByNameInPath 检查指定虚拟路径下是否存在同名文件（排除指定的 ufID）
func (r *userFilesRepository) ExistsByNameInPath(ctx context.Context, userID, virtualPath, fileName string, excludeUfID string) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.UserFiles{}).
		Where("user_id = ? AND virtual_path = ? AND file_name = ?", userID, virtualPath, fileName)
	if excludeUfID != "" {
		query = query.Where("uf_id != ?", excludeUfID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// BatchGetByUserIDAndUfIDs 根据用户ID和多个uf_id批量查询用户文件关联
func (r *userFilesRepository) BatchGetByUserIDAndUfIDs(ctx context.Context, userID string, ufIDs []string) (map[string]*models.UserFiles, error) {
	if len(ufIDs) == 0 {
		return nil, nil
	}
	var userFiles []*models.UserFiles
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND uf_id IN ?", userID, ufIDs).
		Find(&userFiles).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]*models.UserFiles, len(userFiles))
	for _, uf := range userFiles {
		result[uf.UfID] = uf
	}
	return result, nil
}
