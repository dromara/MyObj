package main

import (
	"fmt"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/download"
	"myobj/src/pkg/enum"
	"myobj/src/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// HTTP下载使用示例
func HTTPDownloadExample(db *gorm.DB, userID string) {
	// 1. 创建仓储工厂
	repoFactory := impl.NewRepositoryFactory(db)

	// 2. 创建下载任务记录
	taskID := uuid.Must(uuid.NewV7()).String()
	task := &models.DownloadTask{
		ID:          taskID,
		UserID:      userID,
		Type:        enum.DownloadTaskTypeHttp.Value(),
		URL:         "https://example.com/file.zip",
		VirtualPath: "/离线下载/",
		State:       enum.DownloadTaskStateInit.Value(),
		TargetDir:   "./temp",
		CreateTime:  custom_type.Now(),
		UpdateTime:  custom_type.Now(),
	}

	// 保存任务到数据库
	if err := repoFactory.DownloadTask().Create(nil, task); err != nil {
		fmt.Printf("创建任务失败: %v\n", err)
		return
	}

	// 3. 配置下载选项
	opts := &download.HTTPDownloadOptions{
		EnableEncryption: false,            // 是否加密存储
		VirtualPath:      "/离线下载/",         // 保存的虚拟路径
		MaxRetries:       3,                // 最大重试次数
		ChunkSize:        10 * 1024 * 1024, // 10MB分片
		MaxConcurrent:    4,                // 最多4个并发下载
		Timeout:          300,              // 超时时间（秒）
	}

	// 4. 启动下载（异步执行）
	go func() {
		result, err := download.DownloadHTTP(
			taskID,
			task.URL,
			userID,
			"./temp",
			repoFactory,
			opts,
		)

		if err != nil {
			fmt.Printf("下载失败: %v\n", err)
			return
		}

		fmt.Printf("下载成功！文件ID: %s, 文件名: %s, 大小: %d\n",
			result.FileID, result.FileName, result.FileSize)
	}()

	fmt.Printf("下载任务已创建，任务ID: %s\n", taskID)
}

// 查询下载进度示例
func QueryDownloadProgress(db *gorm.DB, taskID string) {
	repoFactory := impl.NewRepositoryFactory(db)

	task, err := repoFactory.DownloadTask().GetByID(nil, taskID)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	// 显示下载进度
	fmt.Printf("任务状态: %d\n", task.State)
	fmt.Printf("文件名: %s\n", task.FileName)
	fmt.Printf("文件大小: %d\n", task.FileSize)
	fmt.Printf("已下载: %d\n", task.DownloadedSize)
	fmt.Printf("进度: %d%%\n", task.Progress)
	fmt.Printf("速度: %d 字节/秒\n", task.Speed)

	// 状态说明
	switch task.State {
	case enum.DownloadTaskStateInit.Value():
		fmt.Println("状态: 初始化")
	case enum.DownloadTaskStateDownloading.Value():
		fmt.Println("状态: 下载中")
	case enum.DownloadTaskStatePaused.Value():
		fmt.Println("状态: 已暂停")
	case enum.DownloadTaskStateFinished.Value():
		fmt.Println("状态: 已完成")
		fmt.Printf("文件ID: %s\n", task.FileID)
	case enum.DownloadTaskStateFailed.Value():
		fmt.Println("状态: 失败")
		fmt.Printf("错误信息: %s\n", task.ErrorMsg)
	}
}

// 暂停下载示例
func PauseDownloadExample(db *gorm.DB, taskID string) {
	repoFactory := impl.NewRepositoryFactory(db)

	if err := download.PauseDownload(taskID, repoFactory); err != nil {
		fmt.Printf("暂停失败: %v\n", err)
		return
	}

	fmt.Println("下载已暂停")
}

// 恢复下载示例
func ResumeDownloadExample(db *gorm.DB, taskID, userID string) {
	repoFactory := impl.NewRepositoryFactory(db)

	if err := download.ResumeDownload(taskID, userID, "./temp", repoFactory); err != nil {
		fmt.Printf("恢复失败: %v\n", err)
		return
	}

	fmt.Println("下载已恢复")
}

// 取消下载示例
func CancelDownloadExample(db *gorm.DB, taskID string) {
	repoFactory := impl.NewRepositoryFactory(db)

	if err := download.CancelDownload(taskID, repoFactory); err != nil {
		fmt.Printf("取消失败: %v\n", err)
		return
	}

	fmt.Println("下载已取消")
}

// 列出用户所有下载任务
func ListDownloadTasksExample(db *gorm.DB, userID string) {
	repoFactory := impl.NewRepositoryFactory(db)

	// 查询所有任务
	tasks, err := repoFactory.DownloadTask().ListByUserID(nil, userID, 0, 100)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	fmt.Printf("共 %d 个下载任务:\n", len(tasks))
	for _, task := range tasks {
		fmt.Printf("- [%s] %s (进度: %d%%)\n", task.ID, task.FileName, task.Progress)
	}

	// 只查询下载中的任务
	downloadingTasks, err := repoFactory.DownloadTask().ListByState(
		nil,
		userID,
		enum.DownloadTaskStateDownloading.Value(),
		0,
		100,
	)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	fmt.Printf("\n正在下载的任务 (%d 个):\n", len(downloadingTasks))
	for _, task := range downloadingTasks {
		fmt.Printf("- [%s] %s (进度: %d%%, 速度: %d B/s)\n",
			task.ID, task.FileName, task.Progress, task.Speed)
	}
}
