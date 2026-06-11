package service

import (
	"context"
	"fmt"
	"myobj/src/config"
	"myobj/src/core/domain/request"
	"myobj/src/core/domain/response"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/download"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/pkg/util"
	"path"
	"time"

	"github.com/google/uuid"
)

// SharesService 分享服务
type SharesService struct {
	factory    *impl.RepositoryFactory
	cacheLocal cache.Cache
}

func NewSharesService(factory *impl.RepositoryFactory, cacheLocal cache.Cache) *SharesService {
	return &SharesService{
		factory:    factory,
		cacheLocal: cacheLocal,
	}
}
func (s *SharesService) GetRepository(ctx context.Context) *impl.RepositoryFactory {
	return s.factory
}

// CreateShare 创建分享
func (s *SharesService) CreateShare(ctx context.Context, req *request.CreateShareRequest, userID string) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	uid := fmt.Sprintf("%s-%v", uuid.New().String(), util.TimeUtil{}.GetTimestamp())

	// 如果密码为空，不生成哈希，直接设置为空字符串
	var passwordHash string
	if req.Password != "" {
		password, err := util.GeneratePassword(req.Password)
		if err != nil {
			logger.LOG.Error("生成密码失败", "error", err)
			return nil, fmt.Errorf("生成密码失败: %w", err)
		}
		passwordHash = password
	}

	userFile, err := s.factory.UserFiles().GetByUserIDAndUfID(ctx, userID, req.FileID)
	if err != nil {
		logger.LOG.Error("获取文件失败", "error", err)
		return nil, fmt.Errorf("获取文件失败: %w", err)
	}
	data := &models.Share{
		UserID:        userID,
		FileID:        userFile.FileID,
		Token:         uid,
		ExpiresAt:     req.Expire,
		PasswordHash:  passwordHash, // 如果密码为空，这里就是空字符串
		DownloadCount: 0,
		CreatedAt:     custom_type.Now(),
	}
	err = s.factory.Share().Create(ctx, data)
	if err != nil {
		logger.LOG.Error("创建分享失败", "error", err)
		return nil, fmt.Errorf("创建分享失败: %w", err)
	}
	return models.NewJsonResponse(200, "ok", fmt.Sprintf("/api/share/download/%s", uid)), nil
}

// GetShareInfo 获取分享信息（不触发下载）
// password: 分享密码（如果有密码则必需）
func (s *SharesService) GetShareInfo(ctx context.Context, token string, password string) (*response.ShareInfoResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	byToken, err := s.factory.Share().GetByToken(ctx, token)
	if err != nil {
		logger.LOG.Error("获取分享失败", "error", err)
		return nil, fmt.Errorf("获取分享失败: %w", err)
	}
	if byToken == nil {
		return nil, fmt.Errorf("分享不存在")
	}

	// 检查过期时间
	isExpired := byToken.ExpiresAt.Before(custom_type.Now())
	if isExpired {
		return nil, fmt.Errorf("分享已过期")
	}

	// 如果有密码，验证密码
	if byToken.PasswordHash != "" {
		if password == "" {
			// 只返回基本信息，不返回文件详情
			return &response.ShareInfoResponse{
				HasPassword: true,
				IsExpired:   false,
			}, nil
		}
		if !util.CheckPassword(byToken.PasswordHash, password) {
			return nil, fmt.Errorf("密码错误")
		}
	}

	// 密码验证通过或没有密码，获取文件信息
	userFile, err := s.factory.UserFiles().GetByUserIDAndFileID(ctx, byToken.UserID, byToken.FileID)
	if err != nil {
		logger.LOG.Error("获取文件失败", "error", err)
		return nil, fmt.Errorf("获取文件失败: %w", err)
	}

	// 获取文件详细信息
	fileInfo, err := s.factory.FileInfo().GetByID(ctx, byToken.FileID)
	if err != nil {
		logger.LOG.Error("获取文件信息失败", "error", err)
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	return &response.ShareInfoResponse{
		FileID:        byToken.FileID,
		FileName:      userFile.FileName,
		FileSize:      int64(fileInfo.Size),
		MimeType:      fileInfo.Mime,
		HasPassword:   byToken.PasswordHash != "",
		ExpiresAt:     byToken.ExpiresAt.Format("2006-01-02 15:04:05"),
		DownloadCount: byToken.DownloadCount,
		IsExpired:     false,
	}, nil
}

// DownloadShare 下载分享文件
func (s *SharesService) DownloadShare(ctx context.Context, token, psw string) *response.SharesDownloadResponse {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	sdr := &response.SharesDownloadResponse{}
	byToken, err := s.factory.Share().GetByToken(ctx, token)
	if err != nil {
		logger.LOG.Error("获取分享失败", "error", err)
		sdr.Err = "获取分享失败"
		return sdr
	}
	if byToken == nil {
		sdr.Err = "获取分享失败"
		return sdr
	}
	// 验证密码
	if byToken.PasswordHash != "" {
		if !util.CheckPassword(byToken.PasswordHash, psw) {
			sdr.Err = "密码错误"
			return sdr
		}
	}
	// 检查过期时间
	if byToken.ExpiresAt.Before(custom_type.Now()) {
		sdr.Err = "分享已过期"
		return sdr
	}
	disk, err := s.factory.Disk().GetBigDisk(ctx)
	if err != nil {
		logger.LOG.Error("获取磁盘失败", "error", err)
		sdr.Err = "获取磁盘失败"
		return sdr
	}
	// 使用时间戳生成临时目录，避免Windows文件名非法字符
	tmpDir := path.Join(disk.DataPath, config.CONFIG.File.TempDir, fmt.Sprintf("share_%d", util.TimeUtil{}.GetTimestamp()))

	// 准备文件下载（解密+合并）
	result, err := download.PrepareLocalFileDownload(
		ctx,
		byToken.FileID,
		byToken.UserID, // 使用分享者的UserID
		tmpDir,
		s.factory,
		nil, // 分享文件不需要密码（已经验证过分享密码）
	)
	if err != nil {
		logger.LOG.Error("准备文件下载失败", "error", err)
		sdr.Err = "准备文件下载失败"
		return sdr
	}
	forDownload := result.TempFilePath
	id, err := s.factory.UserFiles().GetByUserIDAndFileID(ctx, byToken.UserID, byToken.FileID)
	if err != nil {
		logger.LOG.Error("获取文件失败", "error", err)
		sdr.Err = "获取文件失败"
		return sdr
	}
	err = s.factory.Share().IncrementDownloadCount(ctx, byToken.ID)
	if err != nil {
		logger.LOG.Error("更新下载次数失败", "error", err)
		sdr.Err = "更新下载次数失败"
		return sdr
	}
	sdr.FileName = id.FileName
	sdr.Path = forDownload
	sdr.Temp = tmpDir
	return sdr
}

// GetShareList 获取用户的分享列表
func (s *SharesService) GetShareList(ctx context.Context, userID string) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	shares, err := s.factory.Share().List(ctx, userID, 0, 1000)
	if err != nil {
		logger.LOG.Error("获取分享列表失败", "error", err)
		return nil, fmt.Errorf("获取分享列表失败: %w", err)
	}

	// 收集所有 fileID，批量查询 UserFiles，避免 N+1 查询
	fileIDs := make([]string, 0, len(shares))
	fileIDSet := make(map[string]struct{}, len(shares))
	for _, share := range shares {
		if _, exists := fileIDSet[share.FileID]; !exists {
			fileIDSet[share.FileID] = struct{}{}
			fileIDs = append(fileIDs, share.FileID)
		}
	}

	// 批量查询 UserFiles（同一 userID 下）
	var userFiles []models.UserFiles
	if len(fileIDs) > 0 {
		if err = s.factory.DB().Where("user_id = ? AND file_id IN ?", userID, fileIDs).Find(&userFiles).Error; err != nil {
			logger.LOG.Error("批量获取用户文件失败", "error", err)
		}
	}
	// 构建 file_id -> fileName 映射
	fileNameMap := make(map[string]string, len(userFiles))
	for _, uf := range userFiles {
		fileNameMap[uf.FileID] = uf.FileName
	}

	// 构建带文件名的分享列表
	var shareList []map[string]interface{}
	for _, share := range shares {
		fileName, ok := fileNameMap[share.FileID]
		if !ok {
			logger.LOG.Error("获取用户文件失败", "fileID", share.FileID)
			continue
		}

		shareItem := map[string]interface{}{
			"id":             share.ID,
			"user_id":        share.UserID,
			"file_id":        share.FileID,
			"file_name":      fileName,
			"token":          share.Token,
			"expires_at":     share.ExpiresAt.Format("2006-01-02 15:04:05"),
			"has_password":   share.PasswordHash != "",
			"download_count": share.DownloadCount,
			"created_at":     share.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		shareList = append(shareList, shareItem)
	}

	return models.NewJsonResponse(200, "ok", shareList), nil
}

// DeleteShare 删除分享
func (s *SharesService) DeleteShare(ctx context.Context, shareID int, userID string) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	// 验证分享是否属于该用户
	share, err := s.factory.Share().GetByID(ctx, shareID)
	if err != nil {
		logger.LOG.Error("获取分享失败", "error", err)
		return nil, fmt.Errorf("分享不存在: %w", err)
	}
	if share.UserID != userID {
		return nil, fmt.Errorf("无权限删除该分享")
	}
	err = s.factory.Share().Delete(ctx, shareID)
	if err != nil {
		logger.LOG.Error("删除分享失败", "error", err)
		return nil, fmt.Errorf("删除分享失败: %w", err)
	}
	return models.NewJsonResponse(200, "ok", nil), nil
}

// UpdateSharePassword 修改分享密码
func (s *SharesService) UpdateSharePassword(ctx context.Context, shareID int, password string, userID string) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	// 验证分享是否属于该用户
	share, err := s.factory.Share().GetByID(ctx, shareID)
	if err != nil {
		logger.LOG.Error("获取分享失败", "error", err)
		return nil, fmt.Errorf("分享不存在: %w", err)
	}
	if share.UserID != userID {
		return nil, fmt.Errorf("无权限修改该分享")
	}

	// 如果密码为空，不生成哈希，直接设置为空字符串（表示无密码）
	var passwordHash string
	if password != "" {
		hash, err := util.GeneratePassword(password)
		if err != nil {
			logger.LOG.Error("生成密码失败", "error", err)
			return nil, fmt.Errorf("生成密码失败: %w", err)
		}
		passwordHash = hash
	}

	// 更新密码
	share.PasswordHash = passwordHash
	err = s.factory.Share().Update(ctx, share)
	if err != nil {
		logger.LOG.Error("更新分享密码失败", "error", err)
		return nil, fmt.Errorf("更新密码失败: %w", err)
	}

	return models.NewJsonResponse(200, "ok", nil), nil
}
