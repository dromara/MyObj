package service

import (
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cache"
)

type ServerFactoryInterface interface {
	GetRepository() *impl.RepositoryFactory
}

type ServerFactory struct {
	userService     *UserService
	fileService     *FileService
	shareService    *SharesService
	downloadService *DownloadService
	recycledService *RecycledService
	adminService    *AdminService
}

func NewServiceFactory(factory *impl.RepositoryFactory, cacheLocal cache.Cache) *ServerFactory {
	return &ServerFactory{
		userService:     NewUserService(factory, cacheLocal),
		fileService:     NewFileService(factory, cacheLocal),
		shareService:    NewSharesService(factory, cacheLocal),
		downloadService: NewDownloadService(factory),
		recycledService: NewRecycledService(factory, cacheLocal),
		adminService:    NewAdminService(factory),
	}
}

func (f *ServerFactory) UserService() *UserService {
	return f.userService
}

func (f *ServerFactory) FileService() *FileService {
	return f.fileService
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
