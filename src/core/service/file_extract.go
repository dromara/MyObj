package service

import (
	"context"
	"fmt"
	"myobj/src/core/domain/request"
	"myobj/src/core/domain/response"
	"myobj/src/pkg/download"
	"myobj/src/pkg/extract"
	"myobj/src/pkg/hash"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/pkg/upload"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var extractTasks sync.Map

type ExtractTask struct {
	TaskID             string
	UserID             string
	FileID             string
	ArchiveName        string
	ArchiveType        string
	TotalFiles         int
	TotalSize          int64
	Completed          int
	Failed             int
	Skipped            int
	ConflictResolution string
	Progress           int
	Status             string
	ErrorMsg           string
	CurrentFile        string
	CreatedAt          time.Time
	mu                 sync.Mutex
}

func (f *FileService) CreateExtractTask(req *request.ExtractFileRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	userFile, err := f.factory.UserFiles().GetByUserIDAndUfID(ctx, userID, req.FileID)
	if err != nil {
		return nil, fmt.Errorf("file not found or no permission: %s", req.FileID)
	}
	if userFile.UserID != userID {
		return nil, fmt.Errorf("no permission to access file: %s", req.FileID)
	}

	fileInfo, err := f.factory.FileInfo().GetByID(ctx, userFile.FileID)
	if err != nil {
		return nil, fmt.Errorf("file info not found: %w", err)
	}

	if !extract.IsSupportedArchive(fileInfo.Name) {
		return nil, fmt.Errorf("unsupported archive format: %s, supported: %s",
			filepath.Ext(fileInfo.Name), strings.Join(extract.GetSupportedFormats(), ", "))
	}

	archiveType := extract.DetectArchiveType(fileInfo.Name)

	taskID := uuid.New().String()
	task := &ExtractTask{
		TaskID:             taskID,
		UserID:             userID,
		FileID:             req.FileID,
		ArchiveName:        fileInfo.Name,
		ArchiveType:        archiveTypeStr(archiveType),
		ConflictResolution: req.ConflictResolution,
		Status:             "preparing",
		Progress:           0,
		CreatedAt:          time.Now(),
	}
	extractTasks.Store(taskID, task)

	go f.runExtractTask(ctx, task, userFile, fileInfo, req)

	return models.NewJsonResponse(200, "extract task created", response.ExtractCreateResponse{
		TaskID:      taskID,
		ArchiveName: fileInfo.Name,
		ArchiveType: task.ArchiveType,
		TotalFiles:  0,
		TotalSize:   0,
		Status:      "preparing",
	}), nil
}

func (f *FileService) runExtractTask(ctx context.Context, task *ExtractTask, userFile *models.UserFiles, fileInfo *models.FileInfo, req *request.ExtractFileRequest) {
	defer func() {
		if r := recover(); r != nil {
			task.mu.Lock()
			task.Status = "failed"
			task.ErrorMsg = fmt.Sprintf("extract panic: %v", r)
			task.mu.Unlock()
			logger.LOG.Error("extract task panic", "taskID", task.TaskID, "error", r)
		}
	}()

	disk, err := f.factory.Disk().GetBigDisk(ctx)
	if err != nil {
		taskFail(task, "get disk failed: "+err.Error())
		return
	}

	workDir := filepath.Join(disk.DataPath, "temp", "extract_"+task.TaskID)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		taskFail(task, "create work dir failed: "+err.Error())
		return
	}
	defer func() {
		os.RemoveAll(workDir)
	}()

	task.mu.Lock()
	task.Status = "downloading"
	task.mu.Unlock()

	opts := &download.LocalFileDownloadOptions{}
	if fileInfo.IsEnc {
		if req.FilePassword == "" {
			taskFail(task, "encrypted archive requires password")
			return
		}
		opts.FilePassword = req.FilePassword
	}

	downloadResult, err := download.PrepareLocalFileDownload(
		ctx,
		userFile.FileID,
		task.UserID,
		workDir,
		f.factory,
		opts,
	)
	if err != nil {
		taskFail(task, "download/prepare file failed: "+err.Error())
		return
	}

	archivePath := downloadResult.TempFilePath
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		taskFail(task, "archive file not found")
		return
	}

	// 存储文件使用 .data 后缀，需要恢复原始文件名以便格式检测
	if !extract.IsSupportedArchive(filepath.Base(archivePath)) {
		linkPath := filepath.Join(workDir, fileInfo.Name)
		if err := os.Link(archivePath, linkPath); err != nil {
			taskFail(task, "create hardlink for archive failed: "+err.Error())
			return
		}
		archivePath = linkPath
	}

	task.mu.Lock()
	task.Status = "extracting"
	task.mu.Unlock()

	targetPathID := req.TargetPathID
	if targetPathID == "" {
		targetPathID = "home"
	}

	targetDir := filepath.Join(workDir, "extracted")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		taskFail(task, "create extract dir failed: "+err.Error())
		return
	}

	extractResult, err := extract.ExtractArchive(archivePath, targetDir, &extract.ExtractOptions{})
	if err != nil {
		taskFail(task, "extract failed: "+err.Error())
		return
	}

	task.mu.Lock()
	task.TotalFiles = extractResult.TotalFiles
	task.TotalSize = extractResult.TotalSize
	task.Status = "uploading"
	task.mu.Unlock()

	logger.LOG.Info("extract complete, starting upload",
		"taskID", task.TaskID,
		"totalFiles", extractResult.TotalFiles,
		"totalSize", extractResult.TotalSize,
	)

	for i, entry := range extractResult.Entries {
		if entry.IsDir {
			continue
		}

		task.mu.Lock()
		task.Progress = (i + 1) * 100 / task.TotalFiles
		task.CurrentFile = entry.FileName
		task.mu.Unlock()

		err := f.uploadExtractedFile(ctx, entry.FilePath, entry.FileName, task.UserID, targetPathID, task.ConflictResolution)
		if err != nil {
			if err.Error() == "skipped" {
				task.mu.Lock()
				task.Skipped++
				task.mu.Unlock()
			} else {
				logger.LOG.Warn("upload extracted file failed",
					"taskID", task.TaskID,
					"fileName", entry.FileName,
					"error", err,
				)
				task.mu.Lock()
				task.Failed++
				task.mu.Unlock()
			}
		} else {
			task.mu.Lock()
			task.Completed++
			task.mu.Unlock()
		}
	}

	task.mu.Lock()
	if task.Failed > 0 {
		task.Status = fmt.Sprintf("partial: %d/%d completed, %d failed", task.Completed, task.TotalFiles, task.Failed)
	} else if task.Skipped > 0 {
		task.Status = fmt.Sprintf("completed, %d skipped", task.Skipped)
	} else {
		task.Status = "completed"
	}
	task.Progress = 100
	task.mu.Unlock()

	logger.LOG.Info("extract task finished",
		"taskID", task.TaskID,
		"completed", task.Completed,
		"failed", task.Failed,
		"skipped", task.Skipped,
	)
}

func (f *FileService) uploadExtractedFile(ctx context.Context, filePath, fileName, userID, pathID, conflictResolution string) error {
	fileStat, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("stat file failed: %w", err)
	}

	// 解析目标路径为实际的数字路径ID
	resolvedPathID, resolveErr := f.resolveVirtualPathID(ctx, userID, pathID)
	if resolveErr != nil {
		return fmt.Errorf("resolve virtual path failed: %w", resolveErr)
	}

	// 检查目标路径下是否已有同名文件
	existingFile := f.findExistingFile(ctx, userID, resolvedPathID, fileName)

	if existingFile != nil {
		switch conflictResolution {
		case "overwrite":
			// 计算新文件 hash
			hasher := hash.NewFastBlake3Hasher()
			newHash, _, hashErr := hasher.ComputeFileHash(filePath)
			if hashErr != nil {
				logger.LOG.Warn("compute hash failed for overwrite check", "fileName", fileName, "error", hashErr)
				// hash 计算失败，直接覆盖（删除旧文件后上传新文件）
				if delErr := f.factory.UserFiles().Delete(ctx, userID, existingFile.UfID); delErr != nil {
					return fmt.Errorf("delete existing file failed: %w", delErr)
				}
				break
			}
			// 查询旧文件的 FileInfo 获取 hash
			oldFileInfo, infoErr := f.factory.FileInfo().GetByID(ctx, existingFile.FileID)
			if infoErr != nil {
				logger.LOG.Warn("get old file info failed for overwrite", "fileName", fileName, "error", infoErr)
				if delErr := f.factory.UserFiles().Delete(ctx, userID, existingFile.UfID); delErr != nil {
					return fmt.Errorf("delete existing file failed: %w", delErr)
				}
				break
			}
			// hash 相同，内容一致，跳过上传
			if oldFileInfo.FileHash == newHash {
				logger.LOG.Info("overwrite: same hash, skip upload", "fileName", fileName, "hash", newHash)
				return fmt.Errorf("skipped")
			}
			// hash 不同，软删除旧文件后上传新文件
			if delErr := f.factory.UserFiles().Delete(ctx, userID, existingFile.UfID); delErr != nil {
				return fmt.Errorf("delete existing file failed: %w", delErr)
			}
		case "keep_both":
			// 自动重命名: a.txt -> a (1).txt -> a (2).txt ...
			fileName = f.generateUniqueName(ctx, userID, resolvedPathID, fileName)
		case "cancel":
			// 跳过该文件
			return fmt.Errorf("skipped")
		default:
			// 无策略或未知策略，按默认行为：直接上传（会创建重复记录）
		}
	}

	uploadData := &upload.FileUploadData{
		TempFilePath: filePath,
		FileName:     fileName,
		FileSize:     fileStat.Size(),
		IsEnc:        false,
		IsChunk:      false,
		VirtualPath:  resolvedPathID, // 使用解析后的数字路径ID，确保与冲突检测一致
		UserID:       userID,
		SkipCleanup:  true, // 解压目录由 runExtractTask 统一清理
	}

	_, err = upload.ProcessUploadedFile(uploadData, f.factory)
	if err != nil {
		return fmt.Errorf("process uploaded file failed: %w", err)
	}

	return nil
}

// findExistingFile 在目标路径下查找同名文件
func (f *FileService) findExistingFile(ctx context.Context, userID, virtualPath, fileName string) *models.UserFiles {
	userFiles, err := f.factory.UserFiles().ListByVirtualPath(ctx, userID, virtualPath, 0, 1000)
	if err != nil {
		return nil
	}
	for _, uf := range userFiles {
		if uf.FileName == fileName {
			return uf
		}
	}
	return nil
}

// generateUniqueName 自动生成不冲突的文件名: a.txt -> a (1).txt -> a (2).txt ...
func (f *FileService) generateUniqueName(ctx context.Context, userID, virtualPath, fileName string) string {
	// 获取目标路径下所有文件名
	userFiles, err := f.factory.UserFiles().ListByVirtualPath(ctx, userID, virtualPath, 0, 1000)
	if err != nil {
		return fileName
	}

	existingNames := make(map[string]bool)
	for _, uf := range userFiles {
		existingNames[uf.FileName] = true
	}

	// 如果没有冲突，直接返回
	if !existingNames[fileName] {
		return fileName
	}

	// 分离文件名和扩展名
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)

	counter := 1
	for {
		newName := fmt.Sprintf("%s (%d)%s", nameWithoutExt, counter, ext)
		if !existingNames[newName] {
			return newName
		}
		counter++
	}
}

func (f *FileService) GetExtractProgress(taskID, userID string) (*models.JsonResponse, error) {
	value, ok := extractTasks.Load(taskID)
	if !ok {
		return nil, fmt.Errorf("extract task not found")
	}

	task := value.(*ExtractTask)
	if task.UserID != userID {
		return nil, fmt.Errorf("no permission to access this task")
	}

	task.mu.Lock()
	defer task.mu.Unlock()

	return models.NewJsonResponse(200, "query success", response.ExtractProgressResponse{
		TaskID:       task.TaskID,
		Status:       task.Status,
		Progress:     task.Progress,
		CurrentFile:  task.CurrentFile,
		CurrentIndex: task.Completed + task.Failed + task.Skipped,
		TotalFiles:   task.TotalFiles,
		Completed:    task.Completed,
		Failed:       task.Failed,
		Skipped:      task.Skipped,
		ErrorMsg:     task.ErrorMsg,
	}), nil
}

func taskFail(task *ExtractTask, errMsg string) {
	task.mu.Lock()
	task.Status = "failed"
	task.ErrorMsg = errMsg
	task.mu.Unlock()
	logger.LOG.Error("extract task failed", "taskID", task.TaskID, "error", errMsg)
}

func archiveTypeStr(t extract.ArchiveType) string {
	switch t {
	case extract.ArchiveTypeZIP:
		return "zip"
	case extract.ArchiveTypeTAR:
		return "tar"
	case extract.ArchiveTypeTARGZ:
		return "tar.gz"
	case extract.ArchiveTypeTARBZ2:
		return "tar.bz2"
	case extract.ArchiveType7Z:
		return "7z"
	case extract.ArchiveTypeRAR:
		return "rar"
	case extract.ArchiveTypeTARXZ:
		return "tar.xz"
	case extract.ArchiveTypeTARZST:
		return "tar.zst"
	default:
		return "unknown"
	}
}

// CheckExtractConflict 检测解压冲突
func (f *FileService) CheckExtractConflict(req *request.ExtractCheckRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	userFile, err := f.factory.UserFiles().GetByUserIDAndUfID(ctx, userID, req.FileID)
	if err != nil {
		return nil, fmt.Errorf("file not found or no permission: %s", req.FileID)
	}
	if userFile.UserID != userID {
		return nil, fmt.Errorf("no permission to access file: %s", req.FileID)
	}

	fileInfo, err := f.factory.FileInfo().GetByID(ctx, userFile.FileID)
	if err != nil {
		return nil, fmt.Errorf("file info not found: %w", err)
	}

	if !extract.IsSupportedArchive(fileInfo.Name) {
		return nil, fmt.Errorf("unsupported archive format: %s", filepath.Ext(fileInfo.Name))
	}

	// 解析目标路径为实际的数字路径ID
	resolvedPathID, err := f.resolveVirtualPathID(ctx, userID, req.TargetPathID)
	if err != nil {
		return nil, fmt.Errorf("resolve virtual path failed: %w", err)
	}

	// 获取目标路径下已有文件
	userFiles, err := f.factory.UserFiles().ListByVirtualPath(ctx, userID, resolvedPathID, 0, 1000)
	if err != nil {
		return nil, fmt.Errorf("query target path files failed: %w", err)
	}
	existingNames := make(map[string]bool)
	for _, uf := range userFiles {
		existingNames[uf.FileName] = true
	}

	// 下载压缩包到临时目录，列出文件名
	disk, err := f.factory.Disk().GetBigDisk(ctx)
	if err != nil {
		return nil, fmt.Errorf("get disk failed: %w", err)
	}

	workDir := filepath.Join(disk.DataPath, "temp", "check_"+uuid.New().String())
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("create work dir failed: %w", err)
	}
	defer os.RemoveAll(workDir)

	opts := &download.LocalFileDownloadOptions{}
	if fileInfo.IsEnc {
		if req.FilePassword == "" {
			return nil, fmt.Errorf("encrypted archive requires password")
		}
		opts.FilePassword = req.FilePassword
	}

	downloadResult, err := download.PrepareLocalFileDownload(
		ctx,
		userFile.FileID,
		userID,
		workDir,
		f.factory,
		opts,
	)
	if err != nil {
		return nil, fmt.Errorf("download/prepare file failed: %w", err)
	}

	archivePath := downloadResult.TempFilePath
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("archive file not found")
	}

	// 存储文件使用 .data 后缀，需要恢复原始文件名
	if !extract.IsSupportedArchive(filepath.Base(archivePath)) {
		linkPath := filepath.Join(workDir, fileInfo.Name)
		if err := os.Link(archivePath, linkPath); err != nil {
			return nil, fmt.Errorf("create hardlink for archive failed: %w", err)
		}
		archivePath = linkPath
	}

	// 列出压缩包中的文件名
	entryNames, err := extract.ListArchiveEntries(archivePath)
	if err != nil {
		return nil, fmt.Errorf("list archive entries failed: %w", err)
	}

	// 比对冲突
	var conflictFiles []string
	for _, name := range entryNames {
		if existingNames[name] {
			conflictFiles = append(conflictFiles, name)
		}
	}

	return models.NewJsonResponse(200, "check success", response.ExtractCheckResponse{
		HasConflict:   len(conflictFiles) > 0,
		ConflictFiles: conflictFiles,
		TotalFiles:    len(entryNames),
	}), nil
}

// resolveVirtualPathID 将前端传入的路径标识解析为数据库中的实际数字路径ID
// 与 ProcessUploadedFile/getVirtualPathID 保持一致：
// - "home" -> 查找 /home 子目录的ID（不是根目录）
// - 纯数字 -> 直接返回
// - 其他字符串 -> 通过 getVirtualPathID 逻辑查找
func (f *FileService) resolveVirtualPathID(ctx context.Context, userID, pathID string) (string, error) {
	// 空路径，使用根目录
	if pathID == "" {
		rootPath, err := f.factory.VirtualPath().GetRootPath(ctx, userID)
		if err != nil {
			return "", fmt.Errorf("get root path failed: %w", err)
		}
		return fmt.Sprintf("%d", rootPath.ID), nil
	}

	// 如果已经是纯数字字符串，直接返回
	if matched, _ := regexp.MatchString(`^\d+$`, pathID); matched {
		return pathID, nil
	}

	// 非数字路径字符串（如 "home"），需要像 getVirtualPathID 一样解析
	// 分割路径为各层级
	parts := strings.Split(strings.Trim(pathID, "/"), "/")

	// 首先获取用户的根目录
	rootPath, err := f.factory.VirtualPath().GetRootPath(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("get root path failed: %w", err)
	}

	parentID := fmt.Sprintf("%d", rootPath.ID)
	lastPathID := parentID

	// 逐层查找路径
	for _, part := range parts {
		if part == "" {
			continue
		}

		currentPath := "/" + part
		existingPaths, err := f.factory.VirtualPath().ListByUserID(ctx, userID, 0, 1000)
		if err != nil {
			return "", fmt.Errorf("query virtual paths failed: %w", err)
		}

		var existingPath *models.VirtualPath
		for _, vp := range existingPaths {
			if vp.Path == currentPath && vp.ParentLevel == parentID {
				existingPath = vp
				break
			}
		}

		if existingPath != nil {
			parentID = fmt.Sprintf("%d", existingPath.ID)
			lastPathID = parentID
		} else {
			// 路径不存在，回退到根目录
			return fmt.Sprintf("%d", rootPath.ID), nil
		}
	}

	return lastPathID, nil
}

