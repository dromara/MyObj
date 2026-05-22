package service

import (
	"context"
	"fmt"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/models"
	"os"
	"path/filepath"
	"strings"
)

// assertEnterpriseContext 校验请求企业上下文与资源所属企业一致，防止跨企业 IDOR。
func assertEnterpriseContext(contextEnterpriseID, resourceEnterpriseID string) error {
	if contextEnterpriseID == "" || resourceEnterpriseID == "" {
		return fmt.Errorf("企业上下文无效")
	}
	if contextEnterpriseID != resourceEnterpriseID {
		return fmt.Errorf("企业上下文不匹配")
	}
	return nil
}

// verifyActiveMember 校验用户为企业活跃成员。
func (s *EnterpriseSpaceService) verifyActiveMember(ctx context.Context, enterpriseID, userID string) error {
	member, err := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, enterpriseID, userID)
	if err != nil || member == nil || member.Status != 0 {
		return fmt.Errorf("您不是该企业的活跃成员")
	}
	return nil
}

// cleanupPhysicalFileIfUnreferenced 在无其它引用时删除物理文件与 file_info。
func (s *EnterpriseSpaceService) cleanupPhysicalFileIfUnreferenced(ctx context.Context, txFactory *impl.RepositoryFactory, fileID string) error {
	if fileID == "" {
		return nil
	}

	refs, err := txFactory.EnterpriseSharedFile().ListByFileID(ctx, fileID)
	if err != nil {
		return err
	}
	if len(refs) > 0 {
		return nil
	}

	var userFileCount int64
	if err := txFactory.DB().WithContext(ctx).Model(&models.UserFiles{}).
		Where("file_id = ?", fileID).Count(&userFileCount).Error; err != nil {
		return err
	}
	if userFileCount > 0 {
		return nil
	}

	fileInfo, err := txFactory.FileInfo().GetByID(ctx, fileID)
	if err != nil || fileInfo == nil {
		return nil
	}

	s.removeFileInfoFromDisk(fileInfo)
	return txFactory.FileInfo().Delete(ctx, fileID)
}

func (s *EnterpriseSpaceService) removeFileInfoFromDisk(fileInfo *models.FileInfo) {
	if fileInfo == nil {
		return
	}
	if fileInfo.Path != "" {
		_ = os.Remove(fileInfo.Path)
		infoPath := strings.TrimSuffix(fileInfo.Path, ".data") + ".info"
		_ = os.Remove(infoPath)
	}
	if fileInfo.ThumbnailImg != "" {
		_ = os.Remove(fileInfo.ThumbnailImg)
	}
	if fileInfo.EncPath != "" {
		_ = os.Remove(fileInfo.EncPath)
	}
	if fileInfo.Path != "" {
		dir := filepath.Dir(fileInfo.Path)
		if strings.HasSuffix(dir, string(filepath.Separator)+"data") || strings.HasSuffix(dir, "/data") {
			_ = os.Remove(dir)
		}
	}
}
