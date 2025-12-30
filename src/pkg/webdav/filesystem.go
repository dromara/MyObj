package webdav

import (
	"context"
	"fmt"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"
	"myobj/src/pkg/upload"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/webdav"
)

// MyObjFileSystem WebDAV 文件系统实现
type MyObjFileSystem struct {
	user            *models.UserInfo
	fileRepo        repository.FileInfoRepository
	userFilesRepo   repository.UserFilesRepository
	virtualPathRepo repository.VirtualPathRepository
	diskRepo        repository.DiskRepository
	factory         *impl.RepositoryFactory
}

// NewMyObjFileSystem 创建文件系统实例
func NewMyObjFileSystem(user *models.UserInfo, factory *impl.RepositoryFactory) webdav.FileSystem {
	return &MyObjFileSystem{
		user:            user,
		fileRepo:        factory.FileInfo(),
		userFilesRepo:   factory.UserFiles(),
		virtualPathRepo: factory.VirtualPath(),
		diskRepo:        factory.Disk(),
		factory:         factory,
	}
}

// Mkdir 创建目录
func (fs *MyObjFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	logger.LOG.Info("WebDAV Mkdir", "user_id", fs.user.ID, "path", name)

	// 清理路径
	name = fs.cleanPath(name)
	if name == "/" || name == "" {
		return os.ErrExist
	}

	// 检查父目录是否存在
	parentPath := path.Dir(name)
	if parentPath != "/" && parentPath != "" {
		_, err := fs.virtualPathRepo.GetByPath(ctx, fs.user.ID, parentPath)
		if err != nil {
			return fmt.Errorf("父目录不存在")
		}
	}

	// 检查目录是否已存在
	existing, _ := fs.virtualPathRepo.GetByPath(ctx, fs.user.ID, name)
	if existing != nil {
		return os.ErrExist
	}

	// 创建虚拟目录
	vpath := &models.VirtualPath{
		UserID: fs.user.ID,
		Path:   name,
		IsDir:  true,
	}

	if err := fs.virtualPathRepo.Create(ctx, vpath); err != nil {
		logger.LOG.Error("WebDAV 创建目录失败", "path", name, "error", err)
		return err
	}

	return nil
}

// OpenFile 打开文件
func (fs *MyObjFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	logger.LOG.Info("WebDAV OpenFile", "user_id", fs.user.ID, "path", name, "flag", flag, "isCreate", flag&os.O_CREATE != 0)

	name = fs.cleanPath(name)
	// 如果是系统文件（desktop.ini 等），返回不存在
	if name == "" {
		return nil, os.ErrNotExist
	}

	// 如果是根目录
	if name == "/" {
		return &davDir{
			fs:   fs,
			path: "", // 根目录使用空字符串
			name: "/",
		}, nil
	}

	// 如果是创建模式，直接尝试创建文件（避免锁冲突）
	if flag&os.O_CREATE != 0 {
		// 先检查文件是否已存在
		_, err := fs.getUserFileByPath(ctx, name)
		if err != nil {
			// 文件不存在，创建新文件
			logger.LOG.Info("WebDAV 创建新文件", "path", name)
			return fs.createFile(ctx, name, flag, perm)
		}
	}

	// 尝试作为目录打开
	vpath, err := fs.virtualPathRepo.GetByPath(ctx, fs.user.ID, name)
	if err != nil {
		logger.LOG.Info("WebDAV OpenFile - 第一次查询失败", "path", name, "error", err)
		// 尝试带 / 前缀的版本
		vpath, err = fs.virtualPathRepo.GetByPath(ctx, fs.user.ID, "/"+name)
		if err != nil {
			logger.LOG.Info("WebDAV OpenFile - 第二次查询失败", "path", "/"+name, "error", err)
			logger.LOG.Info("WebDAV OpenFile - 检查第三次查询条件", "name", name, "contains_slash", strings.Contains(name, "/"))
			// 如果路径包含 /，尝试只用最后一部分（文件夹名）查询
			if strings.Contains(name, "/") {
				folderName := path.Base(name)
				parentPath := path.Dir(name)
				if parentPath == "." {
					parentPath = ""
				}
				logger.LOG.Info("WebDAV OpenFile - 尝试第三次查询", "folderName", folderName, "parentPath", parentPath)
				// 先查询父目录 ID
				var parentID int
				if parentPath == "" || parentPath == "/" {
					// 根目录
					rootPath, err := fs.virtualPathRepo.GetRootPath(ctx, fs.user.ID)
					if err == nil {
						parentID = rootPath.ID
					}
				} else {
					// 查询父目录
					parentVPath, err := fs.virtualPathRepo.GetByPath(ctx, fs.user.ID, parentPath)
					if err != nil {
						parentVPath, err = fs.virtualPathRepo.GetByPath(ctx, fs.user.ID, "/"+parentPath)
					}
					if err == nil {
						parentID = parentVPath.ID
					}
				}
				// 如果找到父目录，查询子目录
				if parentID > 0 {
					allDirs, _ := fs.virtualPathRepo.GetPathByUser(ctx, fs.user.ID)
					parentIDStr := fmt.Sprintf("%d", parentID)
					for _, dir := range allDirs {
						if dir.ParentLevel == parentIDStr && dir.IsDir {
							// 比较文件夹名
							dirName := strings.TrimPrefix(dir.Path, "/")
							if strings.Contains(dirName, "/") {
								dirName = path.Base(dir.Path)
							}
							if dirName == folderName {
								vpath = dir
								err = nil
								logger.LOG.Info("WebDAV OpenFile - 第三次查询成功", "folderName", folderName, "vpath.ID", vpath.ID)
								break
							}
						}
					}
				}
			}
		}
	}
	if err == nil && vpath.IsDir {
		logger.LOG.Info("WebDAV OpenFile - 找到目录", "path", name, "vpath.Path", vpath.Path, "vpath.ID", vpath.ID)
		return &davDir{
			fs:   fs,
			path: name, // 使用标准化后的路径（不带前缀 /）
			name: path.Base(name),
		}, nil
	}

	// 尝试作为文件打开
	userFiles, err := fs.getUserFileByPath(ctx, name)
	if err == nil {
		// 获取文件信息
		fileInfo, err := fs.fileRepo.GetByID(ctx, userFiles.FileID)
		if err != nil {
			return nil, err
		}

		// 打开物理文件
		f, err := os.OpenFile(fileInfo.Path, flag, perm)
		if err != nil {
			logger.LOG.Error("WebDAV 打开文件失败", "path", name, "physical_path", fileInfo.Path, "error", err)
			return nil, err
		}

		return &davFile{
			file:      f,
			name:      path.Base(name),
			fileInfo:  fileInfo,
			userFiles: userFiles,
		}, nil
	}

	// 文件/目录不存在，如果是创建模式
	if flag&os.O_CREATE != 0 {
		return fs.createFile(ctx, name, flag, perm)
	}

	return nil, os.ErrNotExist
}

// RemoveAll 删除文件或目录
func (fs *MyObjFileSystem) RemoveAll(ctx context.Context, name string) error {
	logger.LOG.Info("WebDAV RemoveAll", "user_id", fs.user.ID, "path", name)

	name = fs.cleanPath(name)
	if name == "/" || name == "" {
		return os.ErrPermission
	}

	// 尝试删除文件
	userFiles, err := fs.getUserFileByPath(ctx, name)
	if err == nil {
		// 移到回收站
		recycled := &models.Recycled{
			UserID: fs.user.ID,
			FileID: userFiles.FileID,
		}
		recycledRepo := fs.factory.Recycled()
		if err := recycledRepo.Create(ctx, recycled); err != nil {
			logger.LOG.Error("WebDAV 移入回收站失败", "error", err)
			return err
		}

		// 删除 user_files 记录
		if err := fs.userFilesRepo.Delete(ctx, fs.user.ID, userFiles.FileID); err != nil {
			logger.LOG.Error("WebDAV 删除文件记录失败", "error", err)
			return err
		}

		return nil
	}

	// 尝试删除目录
	vpath, err := fs.virtualPathRepo.GetByPath(ctx, fs.user.ID, name)
	if err == nil {
		return fs.virtualPathRepo.Delete(ctx, vpath.ID)
	}

	return os.ErrNotExist
}

// Rename 重命名/移动文件或目录
func (fs *MyObjFileSystem) Rename(ctx context.Context, oldName, newName string) error {
	logger.LOG.Info("WebDAV Rename", "user_id", fs.user.ID, "old", oldName, "new", newName)

	oldName = fs.cleanPath(oldName)
	newName = fs.cleanPath(newName)

	// 尝试重命名文件
	userFiles, err := fs.getUserFileByPath(ctx, oldName)
	if err == nil {
		userFiles.FileName = path.Base(newName)
		// 获取新目录的 virtual_path ID
		newDir := path.Dir(newName)
		if newDir == "." {
			newDir = ""
		}
		if newDir == "" || newDir == "/" {
			userFiles.VirtualPath = "0"
		} else {
			// 同时尝试带 / 和不带 / 的版本
			vpath, err := fs.virtualPathRepo.GetByPath(ctx, fs.user.ID, newDir)
			if err != nil {
				vpath, err = fs.virtualPathRepo.GetByPath(ctx, fs.user.ID, "/"+newDir)
			}
			if err != nil {
				return fmt.Errorf("目标目录不存在")
			}
			userFiles.VirtualPath = fmt.Sprintf("%d", vpath.ID)
		}
		return fs.userFilesRepo.Update(ctx, userFiles)
	}

	// 尝试重命名目录
	vpath, err := fs.virtualPathRepo.GetByPath(ctx, fs.user.ID, oldName)
	if err == nil {
		vpath.Path = newName
		return fs.virtualPathRepo.Update(ctx, vpath)
	}

	return os.ErrNotExist
}

// Stat 获取文件/目录信息
func (fs *MyObjFileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	name = fs.cleanPath(name)

	// 根目录
	if name == "/" || name == "" {
		return &davFileInfo{
			name:    "/",
			size:    0,
			isDir:   true,
			modTime: time.Now(),
		}, nil
	}

	// 尝试查找目录
	vpath, err := fs.virtualPathRepo.GetByPath(ctx, fs.user.ID, name)
	if err != nil {
		// 尝试带 / 前缀的版本
		vpath, err = fs.virtualPathRepo.GetByPath(ctx, fs.user.ID, "/"+name)
		if err != nil && strings.Contains(name, "/") {
			// 如果路径包含 /，尝试只用最后一部分（文件夹名）查询
			folderName := path.Base(name)
			parentPath := path.Dir(name)
			if parentPath == "." {
				parentPath = ""
			}
			// 先查询父目录 ID
			var parentID int
			if parentPath == "" || parentPath == "/" {
				// 根目录
				rootPath, err := fs.virtualPathRepo.GetRootPath(ctx, fs.user.ID)
				if err == nil {
					parentID = rootPath.ID
				}
			} else {
				// 查询父目录
				parentVPath, err := fs.virtualPathRepo.GetByPath(ctx, fs.user.ID, parentPath)
				if err != nil {
					parentVPath, err = fs.virtualPathRepo.GetByPath(ctx, fs.user.ID, "/"+parentPath)
				}
				if err == nil {
					parentID = parentVPath.ID
				}
			}
			// 如果找到父目录，查询子目录
			if parentID > 0 {
				allDirs, _ := fs.virtualPathRepo.GetPathByUser(ctx, fs.user.ID)
				parentIDStr := fmt.Sprintf("%d", parentID)
				for _, dir := range allDirs {
					if dir.ParentLevel == parentIDStr && dir.IsDir {
						// 比较文件夹名
						dirName := strings.TrimPrefix(dir.Path, "/")
						if strings.Contains(dirName, "/") {
							dirName = path.Base(dir.Path)
						}
						if dirName == folderName {
							vpath = dir
							err = nil
							break
						}
					}
				}
			}
		}
	}
	if err == nil && vpath.IsDir {
		return &davFileInfo{
			name:    path.Base(name),
			size:    0,
			isDir:   true,
			modTime: time.Now(),
		}, nil
	}

	// 尝试查找文件
	userFiles, err := fs.getUserFileByPath(ctx, name)
	if err == nil {
		// 获取文件信息以获取大小
		fileInfo, err := fs.fileRepo.GetByID(ctx, userFiles.FileID)
		if err == nil {
			return &davFileInfo{
				name:    path.Base(name),
				size:    int64(fileInfo.Size),
				isDir:   false,
				modTime: time.Time(userFiles.CreatedAt),
			}, nil
		}
		return &davFileInfo{
			name:    path.Base(name),
			size:    0,
			isDir:   false,
			modTime: time.Time(userFiles.CreatedAt),
		}, nil
	}

	return nil, os.ErrNotExist
}

// cleanPath 清理路径
func (fs *MyObjFileSystem) cleanPath(p string) string {
	p = path.Clean("/" + p)
	if p == "/" {
		return "/"
	}
	// 移除前缀斜杠
	p = strings.TrimPrefix(p, "/")
	// 过滤 Windows 系统文件
	if strings.ToLower(p) == "desktop.ini" || strings.HasSuffix(strings.ToLower(p), "/desktop.ini") {
		return "" // 返回空表示忽略
	}
	return p
}

// getUserFileByPath 根据虚拟路径获取用户文件
func (fs *MyObjFileSystem) getUserFileByPath(ctx context.Context, fullPath string) (*models.UserFiles, error) {
	dir := path.Dir(fullPath)
	name := path.Base(fullPath)

	if dir == "." {
		dir = ""
	}

	// 根据目录路径查找 virtual_path ID
	var pathID string
	if dir == "" || dir == "/" {
		// 根目录，查询根目录 ID
		rootPath, err := fs.virtualPathRepo.GetRootPath(ctx, fs.user.ID)
		if err != nil {
			return nil, os.ErrNotExist
		}
		pathID = fmt.Sprintf("%d", rootPath.ID)
	} else {
		// 查询该目录对应的 virtual_path ID
		// 同时尝试带 / 和不带 / 的版本
		vpath, err := fs.virtualPathRepo.GetByPath(ctx, fs.user.ID, dir)
		if err != nil {
			// 尝试带 / 前缀的版本
			vpath, err = fs.virtualPathRepo.GetByPath(ctx, fs.user.ID, "/"+dir)
		}
		if err != nil {
			return nil, os.ErrNotExist
		}
		pathID = fmt.Sprintf("%d", vpath.ID)
	}

	// 查询该目录下的所有文件
	files, err := fs.userFilesRepo.ListByVirtualPath(ctx, fs.user.ID, pathID, 0, 1000)
	if err != nil {
		return nil, err
	}

	// 查找匹配的文件
	for _, f := range files {
		if f.FileName == name {
			return f, nil
		}
	}

	return nil, os.ErrNotExist
}

// createFile 创建新文件
func (fs *MyObjFileSystem) createFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	logger.LOG.Info("WebDAV 创建文件", "user_id", fs.user.ID, "path", name)

	// 1. 获取目标目录的 virtual_path ID
	dir := path.Dir(name)
	if dir == "." {
		dir = ""
	}

	var virtualPathID int
	if dir == "" || dir == "/" {
		// 根目录
		rootPath, err := fs.virtualPathRepo.GetRootPath(ctx, fs.user.ID)
		if err != nil {
			logger.LOG.Error("WebDAV 获取根目录失败", "error", err)
			return nil, os.ErrNotExist
		}
		virtualPathID = rootPath.ID
	} else {
		// 检查 dir 是否为纯数字（virtual_path ID）
		if id, err := strconv.Atoi(dir); err == nil {
			// dir 是数字，直接使用为 virtual_path ID
			logger.LOG.Info("WebDAV 目标目录为 ID", "dir", dir, "virtualPathID", id)
			virtualPathID = id
		} else {
			// dir 是路径名，查询目录
			vpath, err := fs.virtualPathRepo.GetByPath(ctx, fs.user.ID, dir)
			if err != nil {
				// 尝试带 / 前缀的版本
				vpath, err = fs.virtualPathRepo.GetByPath(ctx, fs.user.ID, "/"+dir)
			}
			if err != nil {
				logger.LOG.Error("WebDAV 目标目录不存在", "dir", dir, "error", err)
				return nil, os.ErrNotExist
			}
			virtualPathID = vpath.ID
		}
	}

	// 2. 选择最大剩余空间的磁盘
	bestDisk, err := fs.diskRepo.GetBigDisk(ctx)
	if err != nil {
		logger.LOG.Error("WebDAV 获取磁盘失败", "error", err)
		return nil, fmt.Errorf("无可用磁盘")
	}

	// 3. 创建临时目录：{DiskPath}/temp/{fileName}_{sessionID}/
	fileName := path.Base(name)
	sessionID := uuid.Must(uuid.NewV7()).String()[:8]
	fileNameWithoutExt := fileName
	if idx := strings.LastIndex(fileName, "."); idx != -1 {
		fileNameWithoutExt = fileName[:idx]
	}
	tempBaseDir := filepath.Join(bestDisk.DataPath, "temp", fmt.Sprintf("%s_%s", fileNameWithoutExt, sessionID))
	if err := os.MkdirAll(tempBaseDir, 0755); err != nil {
		logger.LOG.Error("WebDAV 创建临时目录失败", "error", err, "path", tempBaseDir)
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}

	// 4. 创建临时文件
	tempFilePath := filepath.Join(tempBaseDir, "upload.tmp")
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		os.RemoveAll(tempBaseDir) // 清理临时目录
		logger.LOG.Error("WebDAV 创建临时文件失败", "error", err, "path", tempFilePath)
		return nil, fmt.Errorf("创建临时文件失败: %w", err)
	}

	logger.LOG.Info("WebDAV 临时文件已创建", "path", tempFilePath, "diskPath", bestDisk.DataPath)

	// 5. 返回上传文件对象
	return &davUploadFile{
		file:          tempFile,
		name:          fileName,
		tempFilePath:  tempFilePath,
		tempDir:       tempBaseDir,
		virtualPathID: virtualPathID,
		userID:        fs.user.ID,
		fs:            fs,
	}, nil
}

// davDir 目录对象
type davDir struct {
	fs   *MyObjFileSystem
	path string
	name string
	pos  int
}

func (d *davDir) Close() error {
	return nil
}

func (d *davDir) Read(p []byte) (int, error) {
	return 0, os.ErrInvalid
}

func (d *davDir) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

func (d *davDir) Readdir(count int) ([]os.FileInfo, error) {
	ctx := context.Background()

	var infos []os.FileInfo

	// 读取子目录
	// d.path 已经是标准化后的路径（根目录为空，其他不带 / 前缀）
	virtualPath := d.path

	logger.LOG.Info("WebDAV Readdir", "user_id", d.fs.user.ID, "virtual_path", virtualPath)

	// 首先查询当前路径对应的 virtual_path ID
	var currentPathID int
	if virtualPath == "" {
		// 根目录，查询用户根目录
		rootPath, err := d.fs.virtualPathRepo.GetRootPath(ctx, d.fs.user.ID)
		if err != nil {
			logger.LOG.Error("WebDAV 获取根目录失败", "error", err)
			return infos, nil
		}
		currentPathID = rootPath.ID
	} else {
		// 查询该路径对应的 virtual_path ID
		// 同时尝试带 / 和不带 / 的版本
		vpath, err := d.fs.virtualPathRepo.GetByPath(ctx, d.fs.user.ID, virtualPath)
		if err != nil {
			// 尝试带 / 前缀的版本
			vpath, err = d.fs.virtualPathRepo.GetByPath(ctx, d.fs.user.ID, "/"+virtualPath)
			if err != nil {
				logger.LOG.Warn("WebDAV 路径不存在", "virtual_path", virtualPath)
				return infos, nil
			}
		}
		currentPathID = vpath.ID
	}

	// 查询子文件夹：直接查询 parent_level = currentPathID 的所有目录
	parentIDStr := fmt.Sprintf("%d", currentPathID)
	allDirs, _ := d.fs.virtualPathRepo.GetPathByUser(ctx, d.fs.user.ID)
	for _, dir := range allDirs {
		if !dir.IsDir {
			continue
		}
		// 只显示 parent_level 等于 currentPathID 的目录
		if dir.ParentLevel == parentIDStr {
			// 只返回文件夹名称，不是完整路径
			folderName := strings.TrimPrefix(dir.Path, "/")
			if strings.Contains(folderName, "/") {
				// 如果还有 /，取最后一部分
				folderName = path.Base(dir.Path)
			}
			logger.LOG.Info("WebDAV Readdir - 添加子文件夹", "originalPath", dir.Path, "folderName", folderName)
			infos = append(infos, &davFileInfo{
				name:    folderName,
				isDir:   true,
				modTime: time.Time(dir.CreatedTime),
			})
		}
	}
	logger.LOG.Info("WebDAV Readdir - 子文件夹数量", "count", len(infos))

	// 读取文件
	// 注意：user_files 表的 virtual_path 字段存储的是 virtual_path 表的 ID
	// 使用前面查询到的 currentPathID
	pathID := fmt.Sprintf("%d", currentPathID)
	logger.LOG.Info("WebDAV 查询文件", "current_path_id", pathID)

	files, _ := d.fs.userFilesRepo.ListByVirtualPath(ctx, d.fs.user.ID, pathID, 0, 1000)
	for _, f := range files {
		// 获取文件大小
		fileInfo, err := d.fs.fileRepo.GetByID(ctx, f.FileID)
		size := int64(0)
		if err == nil {
			size = int64(fileInfo.Size)
		}
		logger.LOG.Info("WebDAV Readdir - 添加文件", "name", f.FileName, "size", size, "modTime", time.Time(f.CreatedAt))
		infos = append(infos, &davFileInfo{
			name:    f.FileName,
			size:    size,
			isDir:   false,
			modTime: time.Time(f.CreatedAt),
		})
	}

	logger.LOG.Info("WebDAV Readdir - 返回结果", "count", len(infos))
	return infos, nil
}

func (d *davDir) Stat() (os.FileInfo, error) {
	logger.LOG.Info("WebDAV davDir.Stat", "path", d.path, "name", d.name)
	return &davFileInfo{
		name:    d.name,
		isDir:   true,
		size:    0,
		modTime: time.Now(), // 添加修改时间
	}, nil
}

func (d *davDir) Write(p []byte) (int, error) {
	return 0, os.ErrPermission
}

// StatFS 返回文件系统空间信息（用于显示磁盘容量）
func (d *davDir) StatFS() (total, used, avail int64) {
	// 返回用户的存储空间信息
	total = d.fs.user.Space
	used = d.fs.user.Space - d.fs.user.FreeSpace
	avail = d.fs.user.FreeSpace
	return
}

// davFile 文件对象
type davFile struct {
	file      *os.File
	name      string
	fileInfo  *models.FileInfo
	userFiles *models.UserFiles
}

func (f *davFile) Close() error {
	return f.file.Close()
}

func (f *davFile) Read(p []byte) (int, error) {
	return f.file.Read(p)
}

func (f *davFile) Seek(offset int64, whence int) (int64, error) {
	return f.file.Seek(offset, whence)
}

func (f *davFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}

func (f *davFile) Stat() (os.FileInfo, error) {
	return &davFileInfo{
		name:    f.name,
		size:    int64(f.fileInfo.Size),
		modTime: time.Time(f.userFiles.CreatedAt),
	}, nil
}

func (f *davFile) Write(p []byte) (int, error) {
	return f.file.Write(p)
}

// davFileInfo 文件信息
type davFileInfo struct {
	name    string
	size    int64
	isDir   bool
	modTime time.Time
}

func (fi *davFileInfo) Name() string { return fi.name }
func (fi *davFileInfo) Size() int64  { return fi.size }
func (fi *davFileInfo) Mode() os.FileMode {
	if fi.isDir {
		return os.ModeDir | 0755
	}
	return 0644
}
func (fi *davFileInfo) ModTime() time.Time { return fi.modTime }
func (fi *davFileInfo) IsDir() bool        { return fi.isDir }
func (fi *davFileInfo) Sys() interface{}   { return nil }

// davUploadFile 上传文件对象
type davUploadFile struct {
	file          *os.File
	name          string
	tempFilePath  string
	tempDir       string
	virtualPathID int
	userID        string
	fs            *MyObjFileSystem
}

func (f *davUploadFile) Close() error {
	// 1. 关闭文件句柄
	if err := f.file.Close(); err != nil {
		logger.LOG.Error("WebDAV 关闭临时文件失败", "error", err)
		os.RemoveAll(f.tempDir)
		return err
	}

	// 2. 获取文件大小
	fileInfo, err := os.Stat(f.tempFilePath)
	if err != nil {
		logger.LOG.Error("WebDAV 获取文件信息失败", "error", err)
		os.RemoveAll(f.tempDir)
		return err
	}

	fileSize := fileInfo.Size()
	logger.LOG.Info("WebDAV 文件上传完成", "name", f.name, "size", fileSize, "tempPath", f.tempFilePath)

	// 如果文件大小为 0，说明没有数据写入，可能是 LOCK 导致的，直接清理临时文件
	if fileSize == 0 {
		logger.LOG.Warn("WebDAV 文件大小为 0，可能被 LOCK 阻止，不处理此文件", "name", f.name)
		os.RemoveAll(f.tempDir)
		return nil // 返回 nil 避免报错
	}

	// 3. 获取虚拟路径 ID（upload.ProcessUploadedFile 期望的是 ID 字符串）
	virtualPathIDStr := fmt.Sprintf("%d", f.virtualPathID)
	logger.LOG.Info("WebDAV 上传到虚拟路径", "virtualPathID", f.virtualPathID, "virtualPathIDStr", virtualPathIDStr)

	// 4. 调用上传处理
	uploadData := &upload.FileUploadData{
		TempFilePath: f.tempFilePath,
		FileName:     f.name,
		FileSize:     fileSize,
		VirtualPath:  virtualPathIDStr, // 传递 ID 字符串
		UserID:       f.userID,
		IsEnc:        false, // WebDAV 不支持加密
		IsChunk:      false, // WebDAV 不支持分片
	}

	fileID, err := upload.ProcessUploadedFile(uploadData, f.fs.factory)
	if err != nil {
		logger.LOG.Error("WebDAV 文件处理失败", "error", err, "name", f.name)
		// 清理临时文件
		os.RemoveAll(f.tempDir)
		return fmt.Errorf("文件上传失败: %w", err)
	}

	// 5. 上传成功，ProcessUploadedFile 会自动清理临时文件
	logger.LOG.Info("WebDAV 文件上传成功", "name", f.name, "fileID", fileID)
	return nil
}

func (f *davUploadFile) Read(p []byte) (int, error) {
	return 0, os.ErrInvalid
}

func (f *davUploadFile) Seek(offset int64, whence int) (int64, error) {
	return f.file.Seek(offset, whence)
}

func (f *davUploadFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, os.ErrInvalid
}

func (f *davUploadFile) Stat() (os.FileInfo, error) {
	return &davFileInfo{
		name:    f.name,
		size:    0,
		isDir:   false,
		modTime: time.Now(),
	}, nil
}

func (f *davUploadFile) Write(p []byte) (int, error) {
	return f.file.Write(p)
}
