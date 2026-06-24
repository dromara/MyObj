package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"myobj/src/core/domain/request"
	"myobj/src/core/domain/response"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cloud"
	"myobj/src/pkg/logger"
)

// CloudDownloadService 云盘下载服务
type CloudDownloadService struct {
	factory      *impl.RepositoryFactory
	providers    map[string]cloud.CloudProvider
	downloadDir  string
	maxConcurrent int
}

// NewCloudDownloadService 创建云盘下载服务
func NewCloudDownloadService(factory *impl.RepositoryFactory, providers map[string]cloud.CloudProvider) *CloudDownloadService {
	return &CloudDownloadService{
		factory:       factory,
		providers:     providers,
		downloadDir:   "./downloads",
		maxConcurrent: 3,
	}
}

// SaveShareFiles 保存分享文件到本地
func (s *CloudDownloadService) SaveShareFiles(ctx context.Context, req *request.SaveShareFilesRequest, userID string) (*response.SaveShareFilesResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// 获取云盘提供者
	provider, ok := s.providers[req.Provider]
	if !ok {
		return nil, fmt.Errorf("不支持的云盘类型: %s", req.Provider)
	}

	// 创建下载目录
	userDir := filepath.Join(s.downloadDir, userID, req.ShareID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return nil, fmt.Errorf("创建下载目录失败: %w", err)
	}

	// 根据保存类型处理
	switch req.SaveType {
	case "single":
		// 保存单个文件
		return s.saveSingleFile(ctx, provider, req, userDir)
	case "multiple":
		// 保存多个文件
		return s.saveMultipleFiles(ctx, provider, req, userDir)
	case "all":
		// 保存全部文件
		return s.saveAllFiles(ctx, provider, req, userDir)
	case "directory":
		// 保存目录
		return s.saveDirectory(ctx, provider, req, userDir)
	default:
		return nil, fmt.Errorf("不支持的保存类型: %s", req.SaveType)
	}
}

// saveSingleFile 保存单个文件
func (s *CloudDownloadService) saveSingleFile(ctx context.Context, provider cloud.CloudProvider, req *request.SaveShareFilesRequest, userDir string) (*response.SaveShareFilesResponse, error) {
	if len(req.FileIDs) == 0 {
		return nil, fmt.Errorf("未指定文件ID")
	}

	fileID := req.FileIDs[0]
	
	// 获取下载链接
	downloadInfo, err := provider.GetDownloadLink(ctx, req.ShareID, fileID)
	if err != nil {
		return nil, fmt.Errorf("获取下载链接失败: %w", err)
	}

	// 获取文件信息
	files, err := provider.ListShareFiles(ctx, req.ShareID, "")
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	var fileName string
	for _, f := range files {
		if f.FileID == fileID {
			fileName = f.Name
			break
		}
	}
	if fileName == "" {
		fileName = fileID
	}

	// 下载文件
	filePath := filepath.Join(userDir, fileName)
	if err := s.downloadFile(ctx, downloadInfo.URL, filePath); err != nil {
		return nil, fmt.Errorf("下载文件失败: %w", err)
	}

	return &response.SaveShareFilesResponse{
		SuccessCount: 1,
		FailedCount:  0,
		SavedFiles: []response.SavedFileInfo{
			{
				FileID:   fileID,
				FileName: fileName,
				FilePath: filePath,
				Size:     downloadInfo.Size,
			},
		},
	}, nil
}

// saveMultipleFiles 保存多个文件
func (s *CloudDownloadService) saveMultipleFiles(ctx context.Context, provider cloud.CloudProvider, req *request.SaveShareFilesRequest, userDir string) (*response.SaveShareFilesResponse, error) {
	if len(req.FileIDs) == 0 {
		return nil, fmt.Errorf("未指定文件ID")
	}

	var (
		savedFiles  []response.SavedFileInfo
		failedFiles []response.FailedFileInfo
		mu          sync.Mutex
		wg          sync.WaitGroup
		sem         = make(chan struct{}, s.maxConcurrent)
	)

	for _, fileID := range req.FileIDs {
		wg.Add(1)
		go func(fid string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// 获取下载链接
			downloadInfo, err := provider.GetDownloadLink(ctx, req.ShareID, fid)
			if err != nil {
				mu.Lock()
				failedFiles = append(failedFiles, response.FailedFileInfo{
					FileID: fid,
					Error:  err.Error(),
				})
				mu.Unlock()
				return
			}

			// 获取文件名
			files, err := provider.ListShareFiles(ctx, req.ShareID, "")
			if err != nil {
				mu.Lock()
				failedFiles = append(failedFiles, response.FailedFileInfo{
					FileID: fid,
					Error:  err.Error(),
				})
				mu.Unlock()
				return
			}

			var fileName string
			for _, f := range files {
				if f.FileID == fid {
					fileName = f.Name
					break
				}
			}
			if fileName == "" {
				fileName = fid
			}

			// 下载文件
			filePath := filepath.Join(userDir, fileName)
			if err := s.downloadFile(ctx, downloadInfo.URL, filePath); err != nil {
				mu.Lock()
				failedFiles = append(failedFiles, response.FailedFileInfo{
					FileID: fid,
					Error:  err.Error(),
				})
				mu.Unlock()
				return
			}

			mu.Lock()
			savedFiles = append(savedFiles, response.SavedFileInfo{
				FileID:   fid,
				FileName: fileName,
				FilePath: filePath,
				Size:     downloadInfo.Size,
			})
			mu.Unlock()
		}(fileID)
	}

	wg.Wait()

	return &response.SaveShareFilesResponse{
		SuccessCount: len(savedFiles),
		FailedCount:  len(failedFiles),
		SavedFiles:   savedFiles,
		FailedFiles:  failedFiles,
	}, nil
}

// saveAllFiles 保存全部文件
func (s *CloudDownloadService) saveAllFiles(ctx context.Context, provider cloud.CloudProvider, req *request.SaveShareFilesRequest, userDir string) (*response.SaveShareFilesResponse, error) {
	// 获取根目录文件列表
	files, err := provider.ListShareFiles(ctx, req.ShareID, "")
	if err != nil {
		return nil, fmt.Errorf("获取文件列表失败: %w", err)
	}

	// 收集所有文件ID
	var fileIDs []string
	for _, file := range files {
		if !file.IsDir {
			fileIDs = append(fileIDs, file.FileID)
		}
	}

	// 使用multiple模式下载
	req.FileIDs = fileIDs
	return s.saveMultipleFiles(ctx, provider, req, userDir)
}

// saveDirectory 保存目录
func (s *CloudDownloadService) saveDirectory(ctx context.Context, provider cloud.CloudProvider, req *request.SaveShareFilesRequest, userDir string) (*response.SaveShareFilesResponse, error) {
	if len(req.FileIDs) == 0 {
		return nil, fmt.Errorf("未指定目录ID")
	}

	dirID := req.FileIDs[0]
	
	// 创建目录
	dirPath := filepath.Join(userDir, req.DirName)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 递归获取目录下的所有文件
	allFiles, err := s.getAllFilesInDir(ctx, provider, req.ShareID, dirID)
	if err != nil {
		return nil, fmt.Errorf("获取目录文件失败: %w", err)
	}

	// 下载所有文件
	var (
		savedFiles  []response.SavedFileInfo
		failedFiles []response.FailedFileInfo
		mu          sync.Mutex
		wg          sync.WaitGroup
		sem         = make(chan struct{}, s.maxConcurrent)
	)

	for _, file := range allFiles {
		wg.Add(1)
		go func(f cloud.ShareFile) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// 获取下载链接
			downloadInfo, err := provider.GetDownloadLink(ctx, req.ShareID, f.FileID)
			if err != nil {
				mu.Lock()
				failedFiles = append(failedFiles, response.FailedFileInfo{
					FileID: f.FileID,
					Error:  err.Error(),
				})
				mu.Unlock()
				return
			}

			// 下载文件
			filePath := filepath.Join(dirPath, f.Name)
			if err := s.downloadFile(ctx, downloadInfo.URL, filePath); err != nil {
				mu.Lock()
				failedFiles = append(failedFiles, response.FailedFileInfo{
					FileID: f.FileID,
					Error:  err.Error(),
				})
				mu.Unlock()
				return
			}

			mu.Lock()
			savedFiles = append(savedFiles, response.SavedFileInfo{
				FileID:   f.FileID,
				FileName: f.Name,
				FilePath: filePath,
				Size:     downloadInfo.Size,
			})
			mu.Unlock()
		}(file)
	}

	wg.Wait()

	return &response.SaveShareFilesResponse{
		SuccessCount: len(savedFiles),
		FailedCount:  len(failedFiles),
		SavedFiles:   savedFiles,
		FailedFiles:  failedFiles,
	}, nil
}

// getAllFilesInDir 递归获取目录下的所有文件
func (s *CloudDownloadService) getAllFilesInDir(ctx context.Context, provider cloud.CloudProvider, shareID, dirID string) ([]cloud.ShareFile, error) {
	var allFiles []cloud.ShareFile

	files, err := provider.ListShareFiles(ctx, shareID, dirID)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir {
			// 递归获取子目录文件
			subFiles, err := s.getAllFilesInDir(ctx, provider, shareID, file.FileID)
			if err != nil {
				logger.LOG.Error("获取子目录文件失败", "error", err, "dir", file.Name)
				continue
			}
			allFiles = append(allFiles, subFiles...)
		} else {
			allFiles = append(allFiles, file)
		}
	}

	return allFiles, nil
}

// downloadFile 下载文件
func (s *CloudDownloadService) downloadFile(ctx context.Context, url, filePath string) error {
	// 创建HTTP请求
	req, err := newHTTPRequestWithContext(ctx, url)
	if err != nil {
		return err
	}

	// 发送请求
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != 200 && resp.StatusCode != 206 {
		return fmt.Errorf("下载失败: status=%d", resp.StatusCode)
	}

	// 创建文件
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 复制内容
	_, err = io.Copy(out, resp.Body)
	return err
}

// httpClient 全局HTTP客户端
var httpClient = &http.Client{
	Timeout: 30 * time.Minute,
}

// newHTTPRequestWithContext 创建带context的HTTP请求
func newHTTPRequestWithContext(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}
