package main

import (
	"fmt"
	"myobj/src/config"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/download"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 示例：如何使用种子/磁力链下载功能
func main() {
	// 1. 初始化配置
	if err := config.LoadConfig("../config.toml"); err != nil {
		panic(err)
	}

	// 2. 初始化数据库
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// 3. 创建仓储工厂
	repoFactory := impl.NewRepositoryFactory(db)

	// 4. 设置下载选项（可选）
	opts := &download.TorrentDownloadOptions{
		MaxConcurrentPeers: 100,   // 最大并发连接数
		DownloadRateMbps:   0,     // 不限速
		UploadRateMbps:     0,     // 不限速
		EnableEncryption:   false, // 不加密存储
	}

	// 5. 下载磁力链示例
	magnetLink := "magnet:?xt=urn:btih:XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
	result, err := download.DownloadTorrent(
		magnetLink,
		"user_id_123",
		"./temp",
		"/home/我的文件",
		repoFactory,
		opts,
	)

	// 6. 或者下载种子文件示例
	// torrentFile := "./example.torrent"
	// result, err := download.DownloadTorrent(
	//     torrentFile,
	//     "user_id_123",
	//     "./temp",
	//     "/home/我的文件",
	//     repoFactory,
	//     opts,
	// )

	if err != nil {
		panic(err)
	}

	// 7. 处理结果
	fmt.Printf("总文件数: %d\n", result.TotalFiles)
	fmt.Printf("成功文件数: %d\n", len(result.SuccessFiles))
	fmt.Printf("失败文件数: %d\n", len(result.FailedFiles))

	fmt.Println("\n成功上传的文件ID:")
	for _, fileID := range result.SuccessFiles {
		fmt.Printf("  - %s\n", fileID)
	}

	if len(result.FailedFiles) > 0 {
		fmt.Println("\n失败的文件:")
		for _, failed := range result.FailedFiles {
			fmt.Printf("  - %s (%s): %s\n", failed.FileName, failed.FilePath, failed.Error)
		}
	}
}
