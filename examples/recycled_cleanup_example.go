package main

import (
	"fmt"
	"myobj/src/config"
	"myobj/src/internal/repository/database"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/task"
	"time"
)

// 示例：如何使用回收站定时清理功能
func main() {
	// 1. 初始化配置
	if err := config.LoadConfig("../config.toml"); err != nil {
		panic(err)
	}

	// 2. 初始化数据库
	db := database.GetDB()

	// 3. 创建仓储工厂
	repoFactory := impl.NewRepositoryFactory(db)

	// 4. 创建回收站任务实例
	recycledTask := task.NewRecycledTask(repoFactory)

	// ========== 方式1: 手动执行一次清理 ==========
	fmt.Println("方式1: 手动执行一次清理")
	// 清理超过30天的回收站文件
	if err := recycledTask.CleanupExpiredFiles(30); err != nil {
		fmt.Printf("清理失败: %v\n", err)
	} else {
		fmt.Println("清理成功")
	}

	// ========== 方式2: 启动定时清理任务 ==========
	fmt.Println("\n方式2: 启动定时清理任务")
	// 每天凌晨2点执行一次清理（保留30天的文件）
	recycledTask.StartScheduledCleanup(30, 24*time.Hour)

	// 保持程序运行
	select {}
}

/*
清理逻辑说明：

1. 查找超过指定天数的回收站记录
2. 对于每条记录：
   a. 检查文件是否被其他用户持有（通过UserFiles表统计引用数）
   b. 如果引用数 > 1，说明有其他用户持有，仅删除回收站记录
   c. 如果引用数 = 1，说明仅该用户持有，执行完整删除：
      - 删除物理文件（普通文件或加密文件）
      - 删除缩略图
      - 删除分片文件和记录（如果是分片文件）
      - 删除FileInfo记录
      - 删除回收站记录
      - 归还用户空间（仅非无限空间用户，即Space > 0）

3. 所有操作在数据库事务中执行，保证数据一致性

空间归还规则：
- 无限空间用户（Space <= 0）：只删除记录，不归还空间
- 普通用户（Space > 0）：删除记录并归还 FreeSpace += FileSize

文件引用逻辑：
- 通过UserFiles表的file_id字段统计引用数
- 引用数 = 1：仅原用户持有，可以物理删除
- 引用数 > 1：其他用户也持有（秒传场景），仅删除回收站记录

定时任务配置建议：
- 开发环境：每小时执行一次，保留1天（快速测试）
  recycledTask.StartScheduledCleanup(1, time.Hour)

- 生产环境：每天凌晨执行一次，保留30天（推荐）
  recycledTask.StartScheduledCleanup(30, 24*time.Hour)

- 高频场景：每6小时执行一次，保留7天
  recycledTask.StartScheduledCleanup(7, 6*time.Hour)
*/
