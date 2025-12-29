package download

import (
	"context"
	"encoding/base64"
	"fmt"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/enum"
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

// 种子下载任务管理器
var (
	torrentDownloadTasks   = make(map[string]context.CancelFunc)
	torrentDownloadTasksMu sync.RWMutex
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
	cfg.Debug = false    // 禁用调试日志

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

	// 首先获取用户的根目录（home），作为第一级子目录的父级
	rootPath, err := repoFactory.VirtualPath().GetRootPath(ctx, userID)
	if err != nil {
		return fmt.Errorf("获取根目录失败: %w", err)
	}
	var parentID = fmt.Sprintf("%d", rootPath.ID) // 使用根目录的ID作为第一级子目录的父级ID

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

// TorrentFileInfo 种子文件信息
type TorrentFileInfo struct {
	Index int    // 文件索引
	Name  string // 文件名
	Size  int64  // 文件大小
	Path  string // 文件路径（种子内的相对路径）
}

// ParseTorrentResult 解析种子结果
type ParseTorrentResult struct {
	Name      string            // 种子名称
	InfoHash  string            // InfoHash
	Files     []TorrentFileInfo // 文件列表
	TotalSize int64             // 总大小
}

// ParseTorrent 解析种子或磁力链，返回文件列表
// 参数:
//   - content: 种子文件内容（Base64编码）或磁力链接（magnet:开头）
//   - timeout: 超时时间（秒），默认120秒
//
// 返回:
//   - result: 解析结果
//   - err: 错误信息
func ParseTorrent(content string, timeout int) (*ParseTorrentResult, error) {
	if timeout <= 0 {
		timeout = 120 // 默认120秒超时
	}

	// 创建临时目录
	sessionID := uuid.New().String()[:8]
	sessionTempDir := filepath.Join(os.TempDir(), fmt.Sprintf("torrent_parse_%s", sessionID))
	if err := os.MkdirAll(sessionTempDir, 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(sessionTempDir)

	// 配置torrent客户端(优化DHT和Tracker配置)
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = sessionTempDir
	cfg.NoUpload = true    // 仅解析,不上传
	cfg.Seed = true        // 不做种
	cfg.NoDHT = false      // 启用DHT
	cfg.DisableIPv6 = true // 禁用IPv6(国内网络环境优化)
	cfg.DisableTCP = false
	cfg.DisableUTP = false
	cfg.ListenPort = 0 // 系统自动分配端口,避免多用户冲突
	cfg.Debug = false  // 禁用调试日志

	// 创建torrent客户端
	client, err := torrent.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建torrent客户端失败: %w", err)
	}
	defer client.Close()

	// 添加公共DHT节点
	// BitTorrent官方DHT节点
	client.AddDhtNodes([]string{
		"87.98.162.88:6881",   // router.bittorrent.com
		"82.221.103.244:6881", // router.utorrent.com
		"87.98.162.88:6969",   // 备用节点
		"91.121.145.85:6881",  // dht.transmissionbt.com
		"67.215.246.10:6881",  // dht.libtorrent.org
		"176.9.47.217:6881",   // 欧洲节点
		"176.9.47.217:6969",   // 欧洲节点备用
	})

	// 判断是磁力链还是种子文件
	var t *torrent.Torrent
	if strings.HasPrefix(content, "magnet:") {
		// 磁力链接
		t, err = client.AddMagnet(content)
		if err != nil {
			return nil, fmt.Errorf("添加磁力链接失败: %w", err)
		}
		logger.LOG.Info("添加磁力链接成功", "magnet", content)

		// 添加公共Tracker服务器(加速解析)
		t.AddTrackers([][]string{
			{
				"udp://tracker.opentrackr.org:1337/announce",
				"udp://9.rarbg.com:2810/announce",
				"udp://opentracker.i2p.rocks:6969/announce",
				"https://opentracker.i2p.rocks:443/announce",
				"udp://tracker.openbittorrent.com:6969/announce",
				"udp://tracker.torrent.eu.org:451/announce",
				"udp://open.stealth.si:80/announce",
				"udp://exodus.desync.com:6969/announce",
				"http://tracker.opentrackr.org:1337/announce",
			},
		})
	} else {
		// Base64编码的种子文件
		torrentData, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, fmt.Errorf("Base64解码失败: %w", err)
		}

		// 保存为临时文件
		torrentPath := filepath.Join(sessionTempDir, "temp.torrent")
		if err := os.WriteFile(torrentPath, torrentData, 0644); err != nil {
			return nil, fmt.Errorf("保存种子文件失败: %w", err)
		}

		// 添加种子文件
		t, err = client.AddTorrentFromFile(torrentPath)
		if err != nil {
			return nil, fmt.Errorf("添加种子文件失败: %w", err)
		}
		logger.LOG.Info("添加种子文件成功")
	}

	// 等待获取种子元数据（带超时）
	logger.LOG.Info("等待获取种子元数据...", "timeout", timeout)
	select {
	case <-t.GotInfo():
		// 成功获取元数据
		logger.LOG.Info("种子元数据获取成功")
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil, fmt.Errorf("获取种子元数据超时（%d秒）", timeout)
	}

	info := t.Info()

	// 提取种子名称
	torrentName := info.Name
	if strings.HasSuffix(strings.ToLower(torrentName), ".torrent") {
		torrentName = torrentName[:len(torrentName)-8]
	}

	// 构建文件列表
	var files []TorrentFileInfo
	var totalSize int64

	// 检查是单文件种子还是多文件种子
	if len(info.Files) == 0 {
		// 单文件种子(info.Files为空,使用info.Length)
		logger.LOG.Info("检测到单文件种子", "name", info.Name, "size", info.Length)
		files = []TorrentFileInfo{
			{
				Index: 0,
				Name:  info.Name,
				Size:  info.Length,
				Path:  info.Name,
			},
		}
		totalSize = info.Length
	} else {
		// 多文件种子
		logger.LOG.Info("检测到多文件种子", "fileCount", len(info.Files))
		files = make([]TorrentFileInfo, 0, len(info.Files))
		for i, file := range info.Files {
			fileName := filepath.Base(file.Path[len(file.Path)-1])
			filePath := filepath.Join(file.Path...)

			files = append(files, TorrentFileInfo{
				Index: i,
				Name:  fileName,
				Size:  file.Length,
				Path:  filePath,
			})
			totalSize += file.Length
		}
	}

	// 获取InfoHash
	infoHash := t.InfoHash().String()

	result := &ParseTorrentResult{
		Name:      torrentName,
		InfoHash:  infoHash,
		Files:     files,
		TotalSize: totalSize,
	}

	logger.LOG.Info("种子解析完成",
		"name", torrentName,
		"infoHash", infoHash,
		"files", len(files),
		"totalSize", totalSize,
	)

	return result, nil
}

// TorrentSingleFileDownloadOptions 单文件下载配置
type TorrentSingleFileDownloadOptions struct {
	MaxConcurrentPeers int    // 最大并发peer连接数
	DownloadRateMbps   int    // 下载速率限制(Mbps)
	UploadRateMbps     int    // 上传速率限制(Mbps)
	EnableEncryption   bool   // 是否加密存储
	VirtualPath        string // 虚拟路径
	TorrentName        string // 种子名称
	InfoHash           string // InfoHash
	FilePassword       string // 文件密码（加密存储时必需）
}

// DownloadTorrentSingleFile 下载种子中的单个文件
// 参数:
//   - ctx: 上下文（用于取消下载）
//   - taskID: 下载任务ID
//   - content: 种子文件内容（Base64编码）或磁力链接
//   - fileIndex: 要下载的文件索引
//   - userID: 用户ID
//   - tempDir: 临时目录
//   - repoFactory: 数据库仓储工厂
//   - opts: 下载配置选项
//
// 返回:
//   - fileID: 上传成功的文件ID
//   - err: 错误信息
func DownloadTorrentSingleFile(
	ctx context.Context,
	taskID string,
	content string,
	fileIndex int,
	userID string,
	tempDir string,
	repoFactory *impl.RepositoryFactory,
	opts *TorrentSingleFileDownloadOptions,
) (string, error) {
	// 创建可取消的context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 注册取消函数
	torrentDownloadTasksMu.Lock()
	torrentDownloadTasks[taskID] = cancel
	torrentDownloadTasksMu.Unlock()

	// 确保清理
	defer func() {
		torrentDownloadTasksMu.Lock()
		delete(torrentDownloadTasks, taskID)
		torrentDownloadTasksMu.Unlock()
	}()

	// 使用默认配置
	if opts == nil {
		opts = &TorrentSingleFileDownloadOptions{
			MaxConcurrentPeers: 100,
			EnableEncryption:   false,
			VirtualPath:        "/离线下载/",
		}
	}

	// 使用taskID作为临时目录名，支持断点续传
	sessionTempDir := filepath.Join(tempDir, fmt.Sprintf("torrent_%s", taskID))
	if err := os.MkdirAll(sessionTempDir, 0755); err != nil {
		return "", fmt.Errorf("创建临时目录失败: %w", err)
	}

	// 注意：不在这里使用defer清理，由删除任务时手动清理

	// 配置torrent客户端（优化DHT和Tracker配置）
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = sessionTempDir
	cfg.NoUpload = false
	cfg.Seed = false
	cfg.NoDHT = false      // 启用DHT
	cfg.DisableIPv6 = true // 禁用IPv6（国内网络环境优化）
	cfg.DisableTCP = false
	cfg.DisableUTP = false
	cfg.ListenPort = 0 // 系统自动分配端口,避免多用户冲突
	cfg.Debug = false  // 禁用调试日志

	// 优化并发下载性能
	cfg.EstablishedConnsPerTorrent = 200 // 增加并发连接数
	cfg.HalfOpenConnsPerTorrent = 50     // 半开连接数
	cfg.TorrentPeersHighWater = 200      // peer池上限
	cfg.TorrentPeersLowWater = 50        // peer池下限

	// 配置并发连接数（如果用户指定）
	if opts.MaxConcurrentPeers > 0 {
		cfg.EstablishedConnsPerTorrent = opts.MaxConcurrentPeers
	}

	// 配置速率限制
	if opts.DownloadRateMbps > 0 {
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
		return "", fmt.Errorf("创建torrent客户端失败: %w", err)
	}
	defer client.Close()

	// 添加公共DHT节点(使用IP地址,国内可用)
	client.AddDhtNodes([]string{
		"87.98.162.88:6881",   // router.bittorrent.com
		"82.221.103.244:6881", // router.utorrent.com
		"87.98.162.88:6969",   // 备用节点
		"91.121.145.85:6881",  // dht.transmissionbt.com
		"67.215.246.10:6881",  // dht.libtorrent.org
		"176.9.47.217:6881",   // 欧洲节点
		"176.9.47.217:6969",   // 欧洲节点备用
	})

	// 判断是磁力链还是种子文件
	var t *torrent.Torrent
	if strings.HasPrefix(content, "magnet:") {
		// 磁力链接
		t, err = client.AddMagnet(content)
		if err != nil {
			return "", fmt.Errorf("添加磁力链接失败: %w", err)
		}
		logger.LOG.Info("添加磁力链接成功", "taskID", taskID)

		// 添加公共Tracker服务器(加速解析)
		t.AddTrackers([][]string{
			{
				"udp://tracker.opentrackr.org:1337/announce",
				"udp://9.rarbg.com:2810/announce",
				"udp://opentracker.i2p.rocks:6969/announce",
				"https://opentracker.i2p.rocks:443/announce",
				"udp://tracker.openbittorrent.com:6969/announce",
				"udp://tracker.torrent.eu.org:451/announce",
				"udp://open.stealth.si:80/announce",
				"udp://exodus.desync.com:6969/announce",
				"http://tracker.opentrackr.org:1337/announce",
			},
		})
	} else {
		// Base64编码的种子文件
		torrentData, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return "", fmt.Errorf("Base64解码失败: %w", err)
		}

		torrentPath := filepath.Join(sessionTempDir, "temp.torrent")
		if err := os.WriteFile(torrentPath, torrentData, 0644); err != nil {
			return "", fmt.Errorf("保存种子文件失败: %w", err)
		}

		t, err = client.AddTorrentFromFile(torrentPath)
		if err != nil {
			return "", fmt.Errorf("添加种子文件失败: %w", err)
		}
		logger.LOG.Info("添加种子文件成功", "taskID", taskID)
	}

	// 等待获取种子元数据
	logger.LOG.Info("等待获取种子元数据...", "taskID", taskID)
	select {
	case <-t.GotInfo():
		logger.LOG.Info("种子元数据获取成功", "taskID", taskID)
	case <-ctx.Done():
		// 暂停时不返回错误，直接退出
		logger.LOG.Info("任务已暂停，等待恢复", "taskID", taskID)
		return "", nil
	case <-time.After(2 * time.Minute):
		return "", fmt.Errorf("获取种子元数据超时")
	}

	info := t.Info()

	// 判断是单文件还是多文件种子
	var fileName string
	var fileSize int64

	if len(info.Files) == 0 {
		// 单文件种子
		if fileIndex != 0 {
			return "", fmt.Errorf("单文件种子的文件索引必须为0, 当前: %d", fileIndex)
		}
		fileName = info.Name
		fileSize = info.Length

		// 下载全部内容
		t.DownloadAll()
		logger.LOG.Info("开始下载单文件种子", "taskID", taskID, "fileName", fileName, "size", fileSize)
	} else {
		// 多文件种子
		if fileIndex < 0 || fileIndex >= len(info.Files) {
			return "", fmt.Errorf("文件索引无效: %d, 总文件数: %d", fileIndex, len(info.Files))
		}

		fileInfo := info.Files[fileIndex]
		fileName = filepath.Base(fileInfo.Path[len(fileInfo.Path)-1])
		fileSize = fileInfo.Length

		// 只下载指定的文件
		for i := range info.Files {
			if i == fileIndex {
				t.Files()[i].Download()
			} else {
				t.Files()[i].SetPriority(torrent.PiecePriorityNone)
			}
		}
		logger.LOG.Info("开始下载多文件种子中的指定文件", "taskID", taskID, "fileName", fileName, "fileIndex", fileIndex)
	}

	// 获取下载任务
	task, err := repoFactory.DownloadTask().GetByID(ctx, taskID)
	if err != nil {
		return "", fmt.Errorf("获取下载任务失败: %w", err)
	}

	// 更新任务信息和状态为"下载中"
	task.FileName = fileName
	task.FileSize = fileSize
	task.State = enum.DownloadTaskStateDownloading.Value() // 设置为下载中
	task.UpdateTime = custom_type.Now()
	if err := repoFactory.DownloadTask().Update(ctx, task); err != nil {
		logger.LOG.Warn("更新任务信息失败", "taskID", taskID, "error", err)
	}

	// 等待下载完成，带进度监控
	logger.LOG.Info("开始下载文件...", "taskID", taskID, "fileName", fileName)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// 获取目标文件
	var targetFile *torrent.File
	var targetTotalSize int64
	if len(info.Files) == 0 {
		// 单文件种子
		targetFile = t.Files()[0]
		targetTotalSize = info.Length
	} else {
		// 多文件种子
		targetFile = t.Files()[fileIndex]
		targetTotalSize = info.Files[fileIndex].Length
	}
	var lastCompleted int64
	lastUpdate := time.Now()

	for {
		select {
		case <-ctx.Done():
			// 暂停时不返回错误，直接退出
			logger.LOG.Info("任务已暂停，等待恢复", "taskID", taskID)
			return "", nil
		case <-ticker.C:
			// 获取当前文件的完成字节数
			completed := targetFile.BytesCompleted()
			totalSize := targetTotalSize

			// 计算进度和速度
			progress := int(float64(completed) / float64(totalSize) * 100)
			now := time.Now()
			elapsed := now.Sub(lastUpdate).Seconds()
			var speed int64
			if elapsed > 0 {
				speed = int64(float64(completed-lastCompleted) / elapsed)
			}

			// 更新数据库
			task.DownloadedSize = completed
			task.Progress = progress
			task.Speed = speed
			task.UpdateTime = custom_type.Now()
			if err := repoFactory.DownloadTask().Update(ctx, task); err != nil {
				logger.LOG.Error("更新下载任务进度失败", "taskID", taskID, "error", err)
			}

			lastCompleted = completed
			lastUpdate = now

			// 记录进度
			stats := t.Stats()
			logger.LOG.Info("下载进度",
				"taskID", taskID,
				"progress", fmt.Sprintf("%d%%", progress),
				"downloaded", completed,
				"total", totalSize,
				"speed", fmt.Sprintf("%.2f MB/s", float64(speed)/1024/1024),
				"peers", stats.ConnectedSeeders+stats.ActivePeers,
			)

			// 检查是否完成
			if completed >= totalSize {
				logger.LOG.Info("文件下载完成", "taskID", taskID, "fileName", fileName, "size", totalSize)
				goto DownloadComplete
			}
		}
	}

DownloadComplete:
	// 构建文件的虚拟路径（根据种子名称创建子目录）
	torrentName := opts.TorrentName
	if torrentName == "" {
		torrentName = info.Name
		if strings.HasSuffix(strings.ToLower(torrentName), ".torrent") {
			torrentName = torrentName[:len(torrentName)-8]
		}
	}

	// 构建完整的虚拟路径（基础路径/种子名称/文件相对路径）
	// 注意：虚拟路径始终使用 / 作为分隔符，不使用 filepath.Join（它会在Windows下转换为\）
	var relativeDir string
	if len(info.Files) > 0 && len(info.Files[fileIndex].Path) > 1 {
		// 多文件种子,提取相对目录（使用/作为分隔符）
		pathParts := info.Files[fileIndex].Path[:len(info.Files[fileIndex].Path)-1]
		relativeDir = strings.Join(pathParts, "/")
	}
	// 单文件种子relativeDir保持为空

	// 使用 / 拼接虚拟路径
	torrentVirtualPath := strings.TrimSuffix(opts.VirtualPath, "/") + "/" + torrentName
	fileVirtualPath := torrentVirtualPath
	if relativeDir != "" {
		fileVirtualPath = torrentVirtualPath + "/" + relativeDir
	}

	// 确保虚拟路径存在
	if err := ensureVirtualPath(ctx, userID, fileVirtualPath, repoFactory); err != nil {
		return "", fmt.Errorf("创建虚拟目录失败: %w", err)
	}

	// 下载后的文件实际路径
	// 注意：t.Files()[index].Path() 返回的是相对于 DataDir 的路径
	var downloadedPath string
	if len(info.Files) == 0 {
		// 单文件种子
		// 构建绝对路径：DataDir + 相对路径
		relativePath := t.Files()[0].Path()
		downloadedPath = filepath.Join(sessionTempDir, relativePath)
	} else {
		// 多文件种子
		// 构建绝对路径：DataDir + 相对路径
		relativePath := t.Files()[fileIndex].Path()
		downloadedPath = filepath.Join(sessionTempDir, relativePath)
	}

	// 转换为绝对路径（防止相对路径问题）
	downloadedPath, err = filepath.Abs(downloadedPath)
	if err != nil {
		return "", fmt.Errorf("转换绝对路径失败: %w", err)
	}

	logger.LOG.Debug("文件下载路径", "taskID", taskID, "path", downloadedPath)

	// 检查文件是否存在
	fileStat, err := os.Stat(downloadedPath)
	if err != nil {
		return "", fmt.Errorf("文件不存在: %w", err)
	}

	// 准备上传数据
	uploadData := &upload.FileUploadData{
		TempFilePath: downloadedPath,
		FileName:     fileName,
		FileSize:     fileStat.Size(),
		VirtualPath:  fileVirtualPath,
		UserID:       userID,
		IsEnc:        opts.EnableEncryption,
		IsChunk:      false,
		FilePassword: opts.FilePassword,
	}

	// 调用上传处理
	fileID, err := upload.ProcessUploadedFile(uploadData, repoFactory)
	if err != nil {
		os.Remove(downloadedPath)
		return "", fmt.Errorf("上传文件失败: %w", err)
	}

	logger.LOG.Info("文件上传成功",
		"taskID", taskID,
		"fileName", fileName,
		"fileID", fileID,
		"size", fileStat.Size(),
	)

	return fileID, nil
}

// PauseTorrentDownload 暂停种子下载任务
func PauseTorrentDownload(taskID string, repoFactory *impl.RepositoryFactory) error {
	ctx := context.Background()

	task, err := repoFactory.DownloadTask().GetByID(ctx, taskID)
	if err != nil {
		logger.LOG.Error("获取下载任务失败", "taskID", taskID, "error", err)
		return fmt.Errorf("获取任务失败: %w", err)
	}

	// 取消下载任务的context
	torrentDownloadTasksMu.RLock()
	cancel, exists := torrentDownloadTasks[taskID]
	torrentDownloadTasksMu.RUnlock()

	if exists && cancel != nil {
		cancel() // 取消context，停止下载
		logger.LOG.Info("已取消种子下载任务", "taskID", taskID)
	}

	task.State = 2 // 2=暂停
	task.UpdateTime = custom_type.Now()

	if err := repoFactory.DownloadTask().Update(ctx, task); err != nil {
		logger.LOG.Error("更新任务状态失败", "taskID", taskID, "error", err)
		return fmt.Errorf("暂停任务失败: %w", err)
	}

	logger.LOG.Info("种子下载任务已暂停", "taskID", taskID)
	return nil
}

// ResumeTorrentDownload 恢复种子下载任务
func ResumeTorrentDownload(taskID string, userID string, tempDir string, repoFactory *impl.RepositoryFactory) error {
	ctx := context.Background()

	task, err := repoFactory.DownloadTask().GetByID(ctx, taskID)
	if err != nil {
		logger.LOG.Error("获取下载任务失败", "taskID", taskID, "error", err)
		return fmt.Errorf("获取任务失败: %w", err)
	}

	if task.State != 2 { // 2=暂停
		return fmt.Errorf("任务状态不允许恢复")
	}

	task.State = 1     // 1=下载中
	task.ErrorMsg = "" // 清除错误信息
	task.UpdateTime = custom_type.Now()

	if err := repoFactory.DownloadTask().Update(ctx, task); err != nil {
		logger.LOG.Error("更新任务状态失败", "taskID", taskID, "error", err)
		return fmt.Errorf("恢复任务失败: %w", err)
	}

	// 重新启动下载（异步）
	go func() {
		opts := &TorrentSingleFileDownloadOptions{
			MaxConcurrentPeers: 200, // 提高并发连接数以加速下载
			EnableEncryption:   task.EnableEncryption,
			VirtualPath:        task.VirtualPath,
			TorrentName:        task.TorrentName,
			InfoHash:           task.InfoHash,
		}
		_, err := DownloadTorrentSingleFile(
			context.Background(),
			taskID,
			task.URL, // URL字段存储种子内容/磁力链
			task.FileIndex,
			userID,
			tempDir,
			repoFactory,
			opts,
		)
		if err != nil {
			logger.LOG.Error("恢复种子下载失败", "taskID", taskID, "error", err)
			// 获取最新任务状态，防止覆盖暂停状态
			task, getErr := repoFactory.DownloadTask().GetByID(context.Background(), taskID)
			if getErr == nil && task != nil && task.State != 2 {
				// 只有当任务不是暂停状态时，才更新为失败
				task.State = 4 // 4=失败
				task.ErrorMsg = err.Error()
				task.UpdateTime = custom_type.Now()
				repoFactory.DownloadTask().Update(context.Background(), task)
			}
		} else if err == nil {
			// 成功完成下载，但需要检查是否被暂停
			task, getErr := repoFactory.DownloadTask().GetByID(context.Background(), taskID)
			if getErr == nil && task != nil && task.State != 2 {
				// 只有当任务不是暂停状态时，才更新为完成
				// 注意：err==nil且fileID为空表示暂停，不应该更新为完成
				logger.LOG.Debug("任务正常退出，不更新状态", "taskID", taskID)
			}
		}
	}()

	logger.LOG.Info("种子下载任务已恢复", "taskID", taskID)
	return nil
}

// CancelTorrentDownload 取消种子下载任务
func CancelTorrentDownload(taskID string, repoFactory *impl.RepositoryFactory) error {
	ctx := context.Background()

	task, err := repoFactory.DownloadTask().GetByID(ctx, taskID)
	if err != nil {
		logger.LOG.Error("获取下载任务失败", "taskID", taskID, "error", err)
		return fmt.Errorf("获取任务失败: %w", err)
	}

	if task.State == 3 { // 3=完成
		return fmt.Errorf("任务已完成，无法取消")
	}

	// 取消下载任务的context
	torrentDownloadTasksMu.RLock()
	cancel, exists := torrentDownloadTasks[taskID]
	torrentDownloadTasksMu.RUnlock()

	if exists && cancel != nil {
		cancel() // 取消context，停止下载
		logger.LOG.Info("已取消种子下载任务", "taskID", taskID)
	}

	task.State = 4 // 4=失败
	task.ErrorMsg = "用户取消下载"
	task.UpdateTime = custom_type.Now()

	if err := repoFactory.DownloadTask().Update(ctx, task); err != nil {
		logger.LOG.Error("更新任务状态失败", "taskID", taskID, "error", err)
		return fmt.Errorf("取消任务失败: %w", err)
	}

	logger.LOG.Info("种子下载任务已取消", "taskID", taskID)
	return nil
}
