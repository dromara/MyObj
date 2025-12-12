package download

import (
	"context"
	"fmt"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/pkg/upload"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// TorrentDownloadResult 种子下载结果
type TorrentDownloadResult struct {
	SuccessFiles []string     // 成功上传的文件ID列表
	FailedFiles  []FailedFile // 失败的文件详情
	TotalFiles   int          // 总文件数
}

// FailedFile 失败的文件信息
type FailedFile struct {
	FileName string // 文件名
	FilePath string // 种子内的相对路径
	Error    string // 失败原因
}

// TorrentDownloadOptions 下载配置选项
type TorrentDownloadOptions struct {
	MaxConcurrentPeers int  // 最大并发peer连接数，0表示使用默认值
	DownloadRateMbps   int  // 下载速率限制(Mbps)，0表示不限速
	UploadRateMbps     int  // 上传速率限制(Mbps)，0表示不限速
	EnableEncryption   bool // 是否加密存储文件
}

// DownloadTorrent 下载磁力链或种子文件
// 参数:
//   - magnetOrTorrentPath: 磁力链接(magnet:?xt=urn:btih:...)或种子文件路径(.torrent)
//   - userID: 用户ID
//   - tempDir: 临时存放文件的目录
//   - virtualPath: 要关联的逻辑路径(如: /home/我的文件)
//   - repoFactory: 数据库仓储工厂
//   - opts: 下载配置选项(可选，传nil使用默认配置)
//
// 返回:
//   - result: 下载结果，包含成功和失败的文件列表
//   - err: 错误信息
func DownloadTorrent(
	magnetOrTorrentPath string,
	userID string,
	tempDir string,
	virtualPath string,
	repoFactory *impl.RepositoryFactory,
	opts *TorrentDownloadOptions,
) (*TorrentDownloadResult, error) {
	ctx := context.Background()

	// 使用默认配置
	if opts == nil {
		opts = &TorrentDownloadOptions{
			MaxConcurrentPeers: 100,
			EnableEncryption:   false,
		}
	}

	// 创建唯一的临时子目录，避免冲突
	sessionID := uuid.New().String()[:8]
	sessionTempDir := filepath.Join(tempDir, fmt.Sprintf("torrent_%s", sessionID))
	if err := os.MkdirAll(sessionTempDir, 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}

	// 确保临时目录在结束时清理
	defer func() {
		if err := os.RemoveAll(sessionTempDir); err != nil {
			logger.LOG.Warn("清理临时目录失败", "path", sessionTempDir, "error", err)
		}
	}()

	// 配置torrent客户端
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = sessionTempDir
	cfg.NoUpload = false // 允许上传以提高下载速度
	cfg.Seed = false     // 下载完成后不做种

	// 配置并发连接数
	if opts.MaxConcurrentPeers > 0 {
		cfg.EstablishedConnsPerTorrent = opts.MaxConcurrentPeers
	}

	// 配置速率限制（使用 golang.org/x/time/rate 包）
	if opts.DownloadRateMbps > 0 {
		// rate.Limiter 的单位是 bytes/second
		limit := rate.Limit(int64(opts.DownloadRateMbps) * 1024 * 1024 / 8)
		cfg.DownloadRateLimiter = rate.NewLimiter(limit, int(limit))
	}
	if opts.UploadRateMbps > 0 {
		limit := rate.Limit(int64(opts.UploadRateMbps) * 1024 * 1024 / 8)
		cfg.UploadRateLimiter = rate.NewLimiter(limit, int(limit))
	}

	// 创建torrent客户端
	client, err := torrent.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建torrent客户端失败: %w", err)
	}
	defer client.Close()

	// 智能判断输入类型并添加torrent
	var t *torrent.Torrent
	if strings.HasPrefix(magnetOrTorrentPath, "magnet:") {
		// 磁力链接
		t, err = client.AddMagnet(magnetOrTorrentPath)
		if err != nil {
			return nil, fmt.Errorf("添加磁力链接失败: %w", err)
		}
		logger.LOG.Info("添加磁力链接成功", "magnet", magnetOrTorrentPath)
	} else {
		// 种子文件
		t, err = client.AddTorrentFromFile(magnetOrTorrentPath)
		if err != nil {
			return nil, fmt.Errorf("添加种子文件失败: %w", err)
		}
		logger.LOG.Info("添加种子文件成功", "torrent", magnetOrTorrentPath)
	}

	// 等待获取种子元数据
	logger.LOG.Info("等待获取种子元数据...")
	<-t.GotInfo()
	info := t.Info()
	logger.LOG.Info("种子元数据获取成功", "name", info.Name, "files", len(info.Files), "totalSize", info.TotalLength())

	// 提取种子名称（去除.torrent后缀）
	torrentName := info.Name
	if strings.HasSuffix(strings.ToLower(torrentName), ".torrent") {
		torrentName = torrentName[:len(torrentName)-8]
	}

	// 下载所有文件
	t.DownloadAll()

	// 等待下载完成，带进度监控
	logger.LOG.Info("开始下载文件...")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := t.Stats()
			progress := float64(stats.BytesRead.Int64()) / float64(info.TotalLength()) * 100
			logger.LOG.Info("下载进度",
				"progress", fmt.Sprintf("%.2f%%", progress),
				"downloaded", stats.BytesRead.Int64(),
				"total", info.TotalLength(),
				"peers", stats.ConnectedSeeders+stats.ActivePeers,
			)
		default:
			if t.BytesCompleted() == info.TotalLength() {
				logger.LOG.Info("下载完成", "totalSize", info.TotalLength())
				goto DownloadComplete
			}
			time.Sleep(500 * time.Millisecond)
		}
	}

DownloadComplete:
	// 创建虚拟目录结构（包含种子名称的子目录）
	torrentVirtualPath := filepath.Join(virtualPath, torrentName)
	if err := ensureVirtualPath(ctx, userID, torrentVirtualPath, repoFactory); err != nil {
		return nil, fmt.Errorf("创建虚拟目录失败: %w", err)
	}

	// 处理文件上传
	result := &TorrentDownloadResult{
		SuccessFiles: make([]string, 0),
		FailedFiles:  make([]FailedFile, 0),
		TotalFiles:   len(info.Files),
	}

	// 检测文件名冲突并生成唯一文件名
	fileNameMap := make(map[string]int) // 文件名 -> 出现次数
	uniqueFileNames := make(map[int]string)

	for i, file := range info.Files {
		baseName := filepath.Base(file.Path[len(file.Path)-1])
		fileNameMap[baseName]++

		// 如果有重名，添加序号
		if fileNameMap[baseName] > 1 {
			ext := filepath.Ext(baseName)
			nameWithoutExt := strings.TrimSuffix(baseName, ext)
			uniqueName := fmt.Sprintf("%s_%d%s", nameWithoutExt, fileNameMap[baseName]-1, ext)
			uniqueFileNames[i] = uniqueName
		} else {
			uniqueFileNames[i] = baseName
		}
	}

	// 并发上传文件，带重试机制
	var wg sync.WaitGroup
	var mu sync.Mutex
	maxRetries := 3

	for i, file := range info.Files {
		wg.Add(1)
		go func(idx int, torrentFile metainfo.FileInfo) {
			defer wg.Done()

			// 构建文件在种子内的相对路径（保持原始目录结构）
			relativeDir := ""
			if len(torrentFile.Path) > 1 {
				relativeDir = filepath.Join(torrentFile.Path[:len(torrentFile.Path)-1]...)
			}

			// 文件的虚拟路径
			fileVirtualPath := filepath.Join(torrentVirtualPath, relativeDir)

			// 确保文件的虚拟目录存在
			if relativeDir != "" {
				if err := ensureVirtualPath(ctx, userID, fileVirtualPath, repoFactory); err != nil {
					mu.Lock()
					result.FailedFiles = append(result.FailedFiles, FailedFile{
						FileName: uniqueFileNames[idx],
						FilePath: filepath.Join(relativeDir, uniqueFileNames[idx]),
						Error:    fmt.Sprintf("创建虚拟目录失败: %v", err),
					})
					mu.Unlock()
					return
				}
			}

			// 下载后的文件实际路径
			downloadedPath := filepath.Join(sessionTempDir, filepath.Join(torrentFile.Path...))

			// 重试机制
			var uploadErr error
			for attempt := 0; attempt < maxRetries; attempt++ {
				if attempt > 0 {
					logger.LOG.Warn("重试上传文件",
						"file", uniqueFileNames[idx],
						"attempt", attempt+1,
						"maxRetries", maxRetries,
					)
					time.Sleep(time.Duration(attempt) * 2 * time.Second) // 指数退避
				}

				// 检查文件是否存在
				fileInfo, err := os.Stat(downloadedPath)
				if err != nil {
					uploadErr = fmt.Errorf("文件不存在: %w", err)
					continue
				}

				// 准备上传数据
				uploadData := &upload.FileUploadData{
					TempFilePath: downloadedPath,
					FileName:     uniqueFileNames[idx],
					FileSize:     fileInfo.Size(),
					VirtualPath:  fileVirtualPath,
					UserID:       userID,
					IsEnc:        opts.EnableEncryption,
					IsChunk:      false, // BT下载的文件已经是完整的，不是分片上传
				}

				// 调用上传处理
				fileID, err := upload.ProcessUploadedFile(uploadData, repoFactory)
				if err != nil {
					uploadErr = err
					continue
				}

				// 上传成功
				mu.Lock()
				result.SuccessFiles = append(result.SuccessFiles, fileID)
				mu.Unlock()
				logger.LOG.Info("文件上传成功",
					"file", uniqueFileNames[idx],
					"fileID", fileID,
					"size", fileInfo.Size(),
				)
				return
			}

			// 所有重试都失败
			mu.Lock()
			result.FailedFiles = append(result.FailedFiles, FailedFile{
				FileName: uniqueFileNames[idx],
				FilePath: filepath.Join(relativeDir, uniqueFileNames[idx]),
				Error:    uploadErr.Error(),
			})
			mu.Unlock()
			logger.LOG.Error("文件上传失败（已重试3次）",
				"file", uniqueFileNames[idx],
				"error", uploadErr,
			)
		}(i, file)
	}

	// 等待所有上传完成
	wg.Wait()

	logger.LOG.Info("种子下载处理完成",
		"total", result.TotalFiles,
		"success", len(result.SuccessFiles),
		"failed", len(result.FailedFiles),
	)

	return result, nil
}

// ensureVirtualPath 确保虚拟路径存在，不存在则创建（支持层级结构）
func ensureVirtualPath(ctx context.Context, userID, fullPath string, repoFactory *impl.RepositoryFactory) error {
	// 分割路径为各层级
	parts := strings.Split(strings.Trim(fullPath, "/"), "/")
	if len(parts) == 0 {
		return fmt.Errorf("无效的虚拟路径: %s", fullPath)
	}

	var parentID = "" // 根目录的父级ID为空字符串

	// 逐层创建虚拟路径
	for i, part := range parts {
		if part == "" {
			continue
		}

		currentPath := "/" + part

		// 查询当前层级路径是否存在（通过用户ID和路径匹配）
		existingPaths, err := repoFactory.VirtualPath().ListByUserID(ctx, userID, 0, 1000)
		if err != nil {
			return fmt.Errorf("查询虚拟路径失败: %w", err)
		}

		// 查找是否已存在当前路径且父级匹配的记录
		var existingPath *models.VirtualPath
		for _, vp := range existingPaths {
			if vp.Path == currentPath && vp.ParentLevel == parentID {
				existingPath = vp
				break
			}
		}

		if existingPath != nil {
			// 路径已存在，使用该路径的ID作为下一层的父级ID
			parentID = fmt.Sprintf("%d", existingPath.ID)
			continue
		}

		// 路径不存在，创建新记录
		newPath := &models.VirtualPath{
			UserID:      userID,
			Path:        currentPath,
			IsFile:      false,
			IsDir:       true,
			ParentLevel: parentID,
			CreatedTime: custom_type.Now(),
			UpdateTime:  custom_type.Now(),
		}

		if err := repoFactory.VirtualPath().Create(ctx, newPath); err != nil {
			// 可能是并发创建导致的重复，再次查询
			existingPaths, queryErr := repoFactory.VirtualPath().ListByUserID(ctx, userID, 0, 1000)
			if queryErr != nil {
				return fmt.Errorf("创建虚拟路径失败: %w, 查询失败: %w", err, queryErr)
			}
			for _, vp := range existingPaths {
				if vp.Path == currentPath && vp.ParentLevel == parentID {
					existingPath = vp
					break
				}
			}
			if existingPath != nil {
				parentID = fmt.Sprintf("%d", existingPath.ID)
			} else {
				return fmt.Errorf("创建虚拟路径失败且无法查询到已创建的路径")
			}
		} else {
			parentID = fmt.Sprintf("%d", newPath.ID)
		}

		logger.LOG.Debug("创建虚拟路径",
			"userID", userID,
			"path", currentPath,
			"parentID", parentID,
			"level", i+1,
		)
	}

	return nil
}
