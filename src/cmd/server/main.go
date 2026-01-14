package main

import (
	"fmt"
	"log"
	"myobj/src/config"
	"myobj/src/internal/api/routers"
	"myobj/src/internal/repository/database"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/webdav"
	"os"
	"os/signal"
	"syscall"
)

// @title MyObj 文件存储系统 API
// @version 1.0
// @description MyObj 是一个功能强大的文件存储系统，支持文件上传、下载、分享、回收站等功能
// @description 支持大文件分片上传、秒传、文件加密等高级特性
// @termsOfService https://gitee.com/MR-wind/my-obj.git

// @contact.name API Support
// @contact.url https://gitee.com/MR-wind/my-obj.git/issues
// @contact.email support@myobj.com

// @license.name Apache-2.0
// @license.url https://opensource.org/licenses/Apache-2.0

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入 "Bearer {token}" 进行身份验证

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API Key 身份验证

func main() {
	// 1. 初始化配置系统
	if err := initConfig(); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}

	// 2. 初始化日志系统
	if err := initLogger(); err != nil {
		log.Fatalf("日志系统初始化失败: %v", err)
	}

	logger.LOG.Info("========== MyObj 服务器启动中 ==========", "host", config.CONFIG.Server.Host, "port", config.CONFIG.Server.Port)
	logger.LOG.Info("数据库类型", "type", config.CONFIG.Database.Type)
	logger.LOG.Info("日志级别", "level", config.CONFIG.Log.Level)

	// 3. 初始化数据库连接
	if err := initDatabase(); err != nil {
		logger.LOG.Error("数据库初始化失败", "error", err)
		os.Exit(1)
	}
	localCache := cache.InitCache()

	// 4. 启动 WebDAV 服务（如果启用）
	if config.CONFIG.WebDAV.Enable {
		logger.LOG.Info("WebDAV 服务已启用，正在启动...")
		factory := impl.NewRepositoryFactory(database.GetDB())
		webdavServer := webdav.NewServer(factory)
		go func() {
			if err := webdavServer.Start(); err != nil {
				logger.LOG.Error("WebDAV 服务器启动失败", "error", err)
			}
		}()
	} else {
		logger.LOG.Info("WebDAV 服务未启用（可在 config.toml 中设置 webdav.enable=true 启用）")
	}

	// 4.5. 启动 S3 服务（如果启用）
	if config.CONFIG.S3.Enable {
		logger.LOG.Info("S3 服务已启用")
		logger.LOG.Info("S3服务配置",
			"region", config.CONFIG.S3.Region,
			"share_port", config.CONFIG.S3.SharePort,
			"port", config.CONFIG.S3.Port,
		)
		// S3服务通过路由集成到主服务器，在 startServer() 中注册
	} else {
		logger.LOG.Info("S3 服务未启用（可在 config.toml 中设置 s3.enable=true 启用）")
	}

	// 5. 注册关闭信号处理
	setupGracefulShutdown(localCache)
	fmt.Printf("apiKey开启情况: %v", config.CONFIG.Auth.ApiKey)
	// 6. 启动HTTP服务器
	if err := startServer(localCache); err != nil {
		logger.LOG.Error("服务器启动失败", "error", err)
		os.Exit(1)
	}
}

// initConfig 初始化配置文件
// 从 config.toml 加载配置到全局 CONFIG 变量
func initConfig() error {
	log.Println("[初始化] 正在加载配置文件...")
	if err := config.InitConfig(); err != nil {
		logger.LOG.Error("配置初始化失败", "error", err)
		return err
	}
	log.Println("[成功] 配置文件加载完成")
	return nil
}

// initLogger 初始化日志系统
// 根据配置创建日志处理器并设置日志级别
func initLogger() error {
	log.Println("[初始化] 正在初始化日志系统...")
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[错误] 日志系统初始化失败: %v\n", r)
			panic(r)
		}
	}()

	logger.InitLogger()
	log.Println("[成功] 日志系统初始化完成")
	return nil
}

// initDatabase 初始化数据库连接
// 根据配置类型(MySQL/SQLite)建立数据库连接并测试连通性
func initDatabase() error {
	logger.LOG.Info("[初始化] 正在连接数据库...", "type", config.CONFIG.Database.Type)
	defer func() {
		if r := recover(); r != nil {
			logger.LOG.Error("[错误] 数据库连接失败，程序即将退出", "panic", r)
			panic(r)
		}
	}()

	database.InitDataBase()
	logger.LOG.Info("[成功] 数据库连接已建立")

	// 迁移S3数据表（如果启用S3服务）
	if config.CONFIG.S3.Enable {
		logger.LOG.Info("[初始化] 正在迁移S3数据表...")
		if err := database.MigrateS3Tables(database.GetDB()); err != nil {
			logger.LOG.Error("S3数据表迁移失败", "error", err)
			// 不阻塞启动，S3功能可能暂时不可用
		} else {
			logger.LOG.Info("[成功] S3数据表迁移完成")
		}
	}

	return nil
}

// startServer 启动HTTP服务器
// 初始化路由并启动Gin服务器监听请求
func startServer(cacheLocal cache.Cache) error {
	logger.LOG.Info("[初始化] 正在启动HTTP服务器...")
	addr := fmt.Sprintf("%s:%d", config.CONFIG.Server.Host, config.CONFIG.Server.Port)
	logger.LOG.Info("服务器监听地址", "address", addr)
	// 启动服务器
	routers.Execute(cacheLocal)
	return nil
}

// setupGracefulShutdown 设置优雅关闭信号处理
// 监听系统中断信号，在收到信号时优雅关闭应用
func setupGracefulShutdown(cacheLocal cache.Cache) {
	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// 在后台协程中监听关闭信号
	go func() {
		sig := <-sigChan
		logger.LOG.Warn("收到关闭信号，系统即将关闭", "signal", sig.String())

		// 关闭数据库连接
		db := database.GetDB()
		if db != nil {
			sqlDB, err := db.DB()
			if err == nil {
				if err := sqlDB.Close(); err != nil {
					logger.LOG.Error("关闭数据库连接失败", "error", err)
				} else {
					logger.LOG.Info("数据库连接已关闭")
				}
			}
		}
		cacheLocal.Clear()
		cacheLocal.Stop()
		logger.LOG.Info("缓存清理完成")
		logger.LOG.Info("========== MyObj 服务器已停止 ==========")
		os.Exit(0)
	}()
}
