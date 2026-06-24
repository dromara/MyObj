package service

import (
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cache"
)

type ServerFactoryInterface interface {
	GetRepository() *impl.RepositoryFactory
}

type ServerFactory struct {
	userService         *UserService
	fileService         *FileService
	shareService        *SharesService
	downloadService     *DownloadService
	recycledService     *RecycledService
	adminService        *AdminService
	cloudService        *CloudService
	cloudAccountService *CloudAccountService
	fileCategoryService *FileCategoryService
	thumbnailService    *ThumbnailService
	cloudTransferService *CloudTransferService
}

func NewServiceFactory(factory *impl.RepositoryFactory, cacheLocal cache.Cache) *ServerFactory {
	return &ServerFactory{
		userService:         NewUserService(factory, cacheLocal),
		fileService:         NewFileService(factory, cacheLocal),
		shareService:        NewSharesService(factory, cacheLocal),
		downloadService:     NewDownloadService(factory),
		recycledService:     NewRecycledService(factory, cacheLocal),
		adminService:        NewAdminService(factory),
		cloudService:        NewCloudService(factory),
		cloudAccountService: NewCloudAccountService(factory),
		fileCategoryService: NewFileCategoryService(factory, cacheLocal),
		thumbnailService:    NewThumbnailService(factory),
		cloudTransferService: NewCloudTransferService(factory),
	}
}

func (f *ServerFactory) UserService() *UserService {
	return f.userService
}

func (f *ServerFactory) FileService() *FileService {
	return f.fileService
}

func (f *ServerFactory) FileCategoryService() *FileCategoryService {
	return f.fileCategoryService
}

func (f *ServerFactory) ShareService() *SharesService {
	return f.shareService
}

func (f *ServerFactory) DownloadService() *DownloadService {
	return f.downloadService
}

func (f *ServerFactory) RecycledService() *RecycledService {
	return f.recycledService
}

func (f *ServerFactory) AdminService() *AdminService {
	return f.adminService
}

func (f *ServerFactory) CloudService() *CloudService {
	return f.cloudService
}

func (f *ServerFactory) CloudAccountService() *CloudAccountService {
	return f.cloudAccountService
}

func (f *ServerFactory) ThumbnailService() *ThumbnailService {
	return f.thumbnailService
}

func (f *ServerFactory) CloudTransferService() *CloudTransferService {
	return f.cloudTransferService
}
