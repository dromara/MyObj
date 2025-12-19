package repository

import (
	"context"
	"myobj/src/pkg/models"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(ctx context.Context, user *models.UserInfo) error
	GetByID(ctx context.Context, id string) (*models.UserInfo, error)
	GetByUserName(ctx context.Context, userName string) (*models.UserInfo, error)
	Update(ctx context.Context, user *models.UserInfo) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*models.UserInfo, error)
	Count(ctx context.Context) (int64, error)
}

// FileInfoRepository 文件信息仓储接口
type FileInfoRepository interface {
	Create(ctx context.Context, file *models.FileInfo) error
	GetByID(ctx context.Context, id string) (*models.FileInfo, error)
	GetByHash(ctx context.Context, hash string) (*models.FileInfo, error)
	GetByChunkSignature(ctx context.Context, signature string, fileSize int64) (*models.FileInfo, error)
	Update(ctx context.Context, file *models.FileInfo) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*models.FileInfo, error)
	Count(ctx context.Context) (int64, error)
	BatchCreate(ctx context.Context, files []*models.FileInfo) error
	SearchByName(ctx context.Context, keyword string, offset, limit int) ([]*models.FileInfo, error)
	CountByName(ctx context.Context, keyword string) (int64, error)
	// ListByVirtualPath 查询指定虚拟路径下的文件
	ListByVirtualPath(ctx context.Context, userID, virtualPath string, offset, limit int) ([]*models.FileInfo, error)
	// CountByVirtualPath 统计指定虚拟路径下的文件数量
	CountByVirtualPath(ctx context.Context, userID, virtualPath string) (int64, error)
}

// GroupRepository 组仓储接口
type GroupRepository interface {
	Create(ctx context.Context, group *models.Group) error
	GetByID(ctx context.Context, id int) (*models.Group, error)
	Update(ctx context.Context, group *models.Group) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, offset, limit int) ([]*models.Group, error)
	Count(ctx context.Context) (int64, error)
	GetDefaultGroup(ctx context.Context) (*models.Group, error)
}

// ShareRepository 分享仓储接口
type ShareRepository interface {
	Create(ctx context.Context, share *models.Share) error
	GetByID(ctx context.Context, id int) (*models.Share, error)
	GetByToken(ctx context.Context, token string) (*models.Share, error)
	Update(ctx context.Context, share *models.Share) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, userID string, offset, limit int) ([]*models.Share, error)
	Count(ctx context.Context, userID string) (int64, error)
	IncrementDownloadCount(ctx context.Context, id int) error
}

// DiskRepository 磁盘仓储接口
type DiskRepository interface {
	Create(ctx context.Context, disk *models.Disk) error
	GetByID(ctx context.Context, id string) (*models.Disk, error)
	GetBigDisk(ctx context.Context) (*models.Disk, error)
	GetByPath(ctx context.Context, path string) (*models.Disk, error)
	Update(ctx context.Context, disk *models.Disk) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*models.Disk, error)
	Count(ctx context.Context) (int64, error)
}

// ApiKeyRepository API密钥仓储接口
type ApiKeyRepository interface {
	Create(ctx context.Context, apiKey *models.ApiKey) error
	GetByID(ctx context.Context, id int) (*models.ApiKey, error)
	GetByKey(ctx context.Context, key string) (*models.ApiKey, error)
	Update(ctx context.Context, apiKey *models.ApiKey) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, userID string, offset, limit int) ([]*models.ApiKey, error)
	Count(ctx context.Context, userID string) (int64, error)
}

// FileChunkRepository 文件分片仓储接口
type FileChunkRepository interface {
	Create(ctx context.Context, chunk *models.FileChunk) error
	GetByID(ctx context.Context, id string) (*models.FileChunk, error)
	GetByFileID(ctx context.Context, fileID string) ([]*models.FileChunk, error)
	Update(ctx context.Context, chunk *models.FileChunk) error
	Delete(ctx context.Context, id string) error
	DeleteByFileID(ctx context.Context, fileID string) error
	BatchCreate(ctx context.Context, chunks []*models.FileChunk) error
}

// PowerRepository 权限仓储接口
type PowerRepository interface {
	Create(ctx context.Context, power *models.Power) error
	GetByID(ctx context.Context, id int) (*models.Power, error)
	Update(ctx context.Context, power *models.Power) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, offset, limit int) ([]*models.Power, error)
	Count(ctx context.Context) (int64, error)
	GetByGroupID(ctx context.Context, groupID int) ([]*models.Power, error)
}

// GroupPowerRepository 组权限关联仓储接口
type GroupPowerRepository interface {
	Create(ctx context.Context, groupPower *models.GroupPower) error
	GetByGroupID(ctx context.Context, groupID int) ([]*models.GroupPower, error)
	Delete(ctx context.Context, groupID, powerID int) error
	DeleteByGroupID(ctx context.Context, groupID int) error
	BatchCreate(ctx context.Context, groupPowers []*models.GroupPower) error
}

// UserFilesRepository 用户文件关联仓储接口
type UserFilesRepository interface {
	Create(ctx context.Context, userFile *models.UserFiles) error
	GetByUserIDAndFileID(ctx context.Context, userID, fileID string) (*models.UserFiles, error)
	Update(ctx context.Context, userFile *models.UserFiles) error
	Delete(ctx context.Context, userID, fileID string) error
	ListByUserID(ctx context.Context, userID string, offset, limit int) ([]*models.UserFiles, error)
	Count(ctx context.Context, userID string) (int64, error)
	ListPublicFiles(ctx context.Context, offset, limit int) ([]*models.UserFiles, error)
	CountPublicFiles(ctx context.Context) (int64, error)
	SearchPublicFiles(ctx context.Context, keyword string, offset, limit int) ([]*models.UserFiles, error)
	CountPublicFilesByKeyword(ctx context.Context, keyword string) (int64, error)
	SearchUserFiles(ctx context.Context, userID, keyword string, offset, limit int) ([]*models.UserFiles, error)
	CountUserFilesByKeyword(ctx context.Context, userID, keyword string) (int64, error)
	GetByUserIDAndUfID(ctx context.Context, userID, ufID string) (*models.UserFiles, error)
}

// VirtualPathRepository 虚拟路径仓储接口
type VirtualPathRepository interface {
	Create(ctx context.Context, vpath *models.VirtualPath) error
	GetByID(ctx context.Context, id int) (*models.VirtualPath, error)
	GetByPath(ctx context.Context, userID, path string) (*models.VirtualPath, error)
	Update(ctx context.Context, vpath *models.VirtualPath) error
	Delete(ctx context.Context, id int) error
	ListByUserID(ctx context.Context, userID string, offset, limit int) ([]*models.VirtualPath, error)
	Count(ctx context.Context, userID string) (int64, error)
	// ListSubFolders 查询指定父目录ID下的子目录
	ListSubFoldersByParentID(ctx context.Context, userID string, parentID int, offset, limit int) ([]*models.VirtualPath, error)
	// CountSubFolders 统计指定父目录ID下的子目录数量
	CountSubFoldersByParentID(ctx context.Context, userID string, parentID int) (int64, error)
	// GetRootPath 获取用户根目录
	GetRootPath(ctx context.Context, userID string) (*models.VirtualPath, error)
	// GetPathByUser 获取用户所有路径
	GetPathByUser(ctx context.Context, userID string) ([]*models.VirtualPath, error)
}

// RecycledRepository 回收站仓储接口
type RecycledRepository interface {
	Create(ctx context.Context, recycled *models.Recycled) error
	GetByID(ctx context.Context, id string) (*models.Recycled, error)
	GetByUserIDAndFileID(ctx context.Context, userID, fileID string) (*models.Recycled, error)
	Delete(ctx context.Context, id string) error
	ListByUserID(ctx context.Context, userID string, offset, limit int) ([]*models.Recycled, error)
	Count(ctx context.Context, userID string) (int64, error)
	// GetExpiredRecords 获取超过指定天数的回收站记录
	GetExpiredRecords(ctx context.Context, days int) ([]*models.Recycled, error)
	// CountFileReferences 统计指定文件被多少个用户持有
	CountFileReferences(ctx context.Context, fileID string) (int64, error)
}

// DownloadTaskRepository 下载任务仓储接口
type DownloadTaskRepository interface {
	Create(ctx context.Context, task *models.DownloadTask) error
	GetByID(ctx context.Context, id string) (*models.DownloadTask, error)
	Update(ctx context.Context, task *models.DownloadTask) error
	Delete(ctx context.Context, id string) error
	ListByUserID(ctx context.Context, userID string, offset, limit int) ([]*models.DownloadTask, error)
	Count(ctx context.Context, userID string) (int64, error)
	// ListByState 查询指定状态的任务
	ListByState(ctx context.Context, userID string, state int, offset, limit int) ([]*models.DownloadTask, error)
	// CountByState 统计指定状态的任务数量
	CountByState(ctx context.Context, userID string, state int) (int64, error)
}

// SysConfigRepository 系统配置仓储接口
type SysConfigRepository interface {
	Create(ctx context.Context, config *models.SysConfig) error
	GetByID(ctx context.Context, id int) (*models.SysConfig, error)
	GetByKey(ctx context.Context, key string) (*models.SysConfig, error)
	Update(ctx context.Context, config *models.SysConfig) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, offset, limit int) ([]*models.SysConfig, error)
	Count(ctx context.Context) (int64, error)
	// BatchUpdate 批量更新配置
	BatchUpdate(ctx context.Context, configs []*models.SysConfig) error
	// GetAllAsMap 获取所有配置并以 key-value 格式返回
	GetAllAsMap(ctx context.Context) (map[string]string, error)
}

// UploadChunkRepository 上传分片信息仓储接口
type UploadChunkRepository interface {
	Create(ctx context.Context, chunk *models.UploadChunk) error
	GetByID(ctx context.Context, chunkID int) (*models.UploadChunk, error)
	Update(ctx context.Context, chunk *models.UploadChunk) error
	Delete(ctx context.Context, chunkID int) error
	ListByUserID(ctx context.Context, userID string, offset, limit int) ([]*models.UploadChunk, error)
	Count(ctx context.Context, userID string) (int64, error)
	// GetByUserIDAndFileName 根据用户ID和文件名获取分片信息
	GetByUserIDAndFileName(ctx context.Context, userID, fileName string) ([]models.UploadChunk, error)
	// DeleteByUserID 删除用户的所有上传分片记录
	DeleteByUserID(ctx context.Context, userID string) error
	// ListByPathID 根据路径ID获取分片列表
	ListByPathID(ctx context.Context, pathID string, offset, limit int) ([]*models.UploadChunk, error)
}
