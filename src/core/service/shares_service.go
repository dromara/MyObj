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
func (s *SharesService) GetRepository() *impl.RepositoryFactory {
	return s.factory
}

// CreateShare 创建分享
func (s *SharesService) CreateShare(req *request.CreateShareRequest, userID string) (*models.JsonResponse, error) {
	uid := fmt.Sprintf("%s-%v", uuid.New().String(), util.TimeUtil{}.GetTimestamp())
	password, err := util.GeneratePassword(req.Password)
	if err != nil {
		logger.LOG.Error("生成密码失败", "error", err)
		return nil, err
	}
	data := &models.Share{
		UserID:        userID,
		FileID:        req.FileID,
		Token:         uid,
		ExpiresAt:     req.Expire,
		PasswordHash:  password,
		DownloadCount: 0,
		CreatedAt:     custom_type.Now(),
	}
	err = s.factory.Share().Create(context.Background(), data)
	if err != nil {
		logger.LOG.Error("创建分享失败", "error", err)
		return nil, err
	}
	return models.NewJsonResponse(200, "ok", fmt.Sprintf("/api/share/download/%s", uid)), nil
}

// GetShare 获取分享
func (s *SharesService) GetShare(token, psw string) *response.SharesDownloadResponse {
	ctx := context.Background()
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
	tmpDir := path.Join(disk.DiskPath, config.CONFIG.File.TempDir, util.TimeUtil{}.GetTime())
	forDownload, _, err := download.PrepareFileForDownload(byToken.FileID, tmpDir, s.factory)
	if err != nil {
		logger.LOG.Error("准备文件下载失败", "error", err)
		sdr.Err = "准备文件下载失败"
		return sdr
	}
	id, err := s.factory.UserFiles().GetByUserIDAndFileID(ctx, byToken.UserID, byToken.FileID)
	if err != nil {
		logger.LOG.Error("获取文件失败", "error", err)
		sdr.Err = "获取文件失败"
		return sdr
	}
	byToken.DownloadCount += 1
	err = s.factory.Share().Update(ctx, byToken)
	if err != nil {
		logger.LOG.Error("更新分享失败", "error", err)
		sdr.Err = "获取分享失败"
		return sdr
	}
	sdr.FileName = id.FileName
	sdr.Path = forDownload
	sdr.Temp = tmpDir
	return sdr
}

// GetShareList 获取用户的分享列表
func (s *SharesService) GetShareList(userID string) (*models.JsonResponse, error) {
	ctx := context.Background()
	shares, err := s.factory.Share().List(ctx, userID, 0, 1000)
	if err != nil {
		logger.LOG.Error("获取分享列表失败", "error", err)
		return nil, fmt.Errorf("获取分享列表失败")
	}

	// 构建带文件名的分享列表
	var shareList []map[string]interface{}
	for _, share := range shares {
		// 获取用户文件信息
		userFile, err := s.factory.UserFiles().GetByUserIDAndFileID(ctx, share.UserID, share.FileID)
		if err != nil {
			logger.LOG.Error("获取用户文件失败", "error", err, "fileID", share.FileID)
			continue
		}

		shareItem := map[string]interface{}{
			"id":             share.ID,
			"user_id":        share.UserID,
			"file_id":        share.FileID,
			"file_name":      userFile.FileName,
			"token":          share.Token,
			"expires_at":     share.ExpiresAt.Format("2006-01-02 15:04:05"),
			"password_hash":  share.PasswordHash,
			"download_count": share.DownloadCount,
			"created_at":     share.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		shareList = append(shareList, shareItem)
	}

	return models.NewJsonResponse(200, "ok", shareList), nil
}

// DeleteShare 删除分享
func (s *SharesService) DeleteShare(shareID int, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()
	// 验证分享是否属于该用户
	share, err := s.factory.Share().GetByID(ctx, shareID)
	if err != nil {
		logger.LOG.Error("获取分享失败", "error", err)
		return nil, fmt.Errorf("分享不存在")
	}
	if share.UserID != userID {
		return nil, fmt.Errorf("无权限删除该分享")
	}
	err = s.factory.Share().Delete(ctx, shareID)
	if err != nil {
		logger.LOG.Error("删除分享失败", "error", err)
		return nil, fmt.Errorf("删除分享失败")
	}
	return models.NewJsonResponse(200, "ok", nil), nil
}

// UpdateSharePassword 修改分享密码
func (s *SharesService) UpdateSharePassword(shareID int, password string, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()
	// 验证分享是否属于该用户
	share, err := s.factory.Share().GetByID(ctx, shareID)
	if err != nil {
		logger.LOG.Error("获取分享失败", "error", err)
		return nil, fmt.Errorf("分享不存在")
	}
	if share.UserID != userID {
		return nil, fmt.Errorf("无权限修改该分享")
	}

	// 生成新密码哈希
	passwordHash, err := util.GeneratePassword(password)
	if err != nil {
		logger.LOG.Error("生成密码失败", "error", err)
		return nil, fmt.Errorf("生成密码失败")
	}

	// 更新密码
	share.PasswordHash = passwordHash
	err = s.factory.Share().Update(ctx, share)
	if err != nil {
		logger.LOG.Error("更新分享密码失败", "error", err)
		return nil, fmt.Errorf("更新密码失败")
	}

	return models.NewJsonResponse(200, "ok", nil), nil
}
