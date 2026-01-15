package routers

import (
	"fmt"
	"myobj/src/config"
	"myobj/src/core/service"
	"myobj/src/internal/api/handlers"
	"myobj/src/internal/api/middleware"
	"myobj/src/internal/repository/database"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/task"
	"myobj/src/s3_server/router"
	s3Service "myobj/src/s3_server/service"
	s3Task "myobj/src/s3_server/task"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "myobj/docs" // 引入 Swagger 文档
)

// Handler 路由处理器接口
// 所有的Handler都应该实现此接口，通过Router方法注册路由
type Handler interface {
	// Router 注册路由到指定的路由组
	Router(c *gin.RouterGroup)
}

// initRouter 初始化路由
// 创建Gin引擎并配置中间件和路由
func initRouter(factory *service.ServerFactory, cache cache.Cache) *gin.Engine {
	logger.LOG.Info("[路由] 开始初始化路由...")
	if config.CONFIG.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
		logger.LOG.Info("[路由] Gin运行模式", "mode", "debug")
	} else {
		gin.SetMode(gin.ReleaseMode)
		logger.LOG.Info("[路由] Gin运行模式", "mode", "release")
	}
	// 创建Gin引擎
	r := gin.New()
	// 注册全局中间件
	logger.LOG.Info("[路由] 正在注册全局中间件...")
	r.Use(middleware.CORS())      // CORS跨域中间件
	r.Use(middleware.GinLogger()) // 自定义日志中间件
	r.Use(gin.Recovery())         // 恢复中间件
	logger.LOG.Info("[路由] 中间件注册完成✔️")

	// 注册路由组
	logger.LOG.Info("[路由] 正在注册API路由...")
	// 尝试加载 HTML 模板（如果存在）
	if _, err := os.Stat("templates"); err == nil {
		r.LoadHTMLGlob("templates/*")
		logger.LOG.Info("[路由] HTML模板已加载")
	}

	// 托管前端静态文件
	r.Static("/assets", "./webview/dist/assets")
	r.StaticFile("/vite.svg", "./webview/dist/vite.svg")
	logger.LOG.Info("[路由] 前端静态资源已注册")

	// Swagger API 文档路由（根据配置决定是否启用）
	if config.CONFIG.Server.Swagger {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		logger.LOG.Info("[路由] Swagger 文档已启用", "url", "http://"+config.CONFIG.Server.Host+fmt.Sprintf(":%d/swagger/index.html", config.CONFIG.Server.Port))
	} else {
		logger.LOG.Info("[路由] Swagger 文档已禁用（可在config.toml中设置 server.swagger=true 启用）")
	}

	api := r.Group("/api")
	{
		// 用户相关路由
		handlers.NewUserHandler(factory.UserService(), cache).Router(api)
		handlers.NewFileHandler(factory.FileService(), cache).Router(api)
		handlers.NewSharesHandler(factory.ShareService(), cache).Router(api)
		handlers.NewDownloadHandler(factory.DownloadService(), cache).Router(api)
		handlers.NewRecycledHandler(factory.RecycledService(), cache).Router(api)
		// 视频播放路由
		handlers.NewVideoHandler(factory.FileService(), cache).Router(api)
		// 管理路由
		handlers.NewAdminHandler(factory.AdminService(), cache).Router(api)
		// TODO: 这里可以注册更多的路由处理器
	}

	// 前端路由处理（SPA支持）
	// 所有非API、非静态资源请求都返回 index.html，由前端路由处理
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// API路径和静态资源路径不处理
		if len(path) >= 4 && path[:4] == "/api" || len(path) >= 7 && path[:7] == "/assets" || path == "/vite.svg" || len(path) >= 8 && path[:8] == "/swagger" {
			c.Next()
			return
		}
		// 其他所有请求返回前端首页
		c.File("./webview/dist/index.html")
	})

	logger.LOG.Info("[路由] 路由初始化完成 ✔️")
	logger.LOG.Info("[路由] 前端页面访问地址", "url", "http://"+config.CONFIG.Server.Host+fmt.Sprintf(":%d", config.CONFIG.Server.Port))

	// 注册S3路由（如果启用）
	if config.CONFIG.S3.Enable {
		logger.LOG.Info("[路由] 正在注册S3 API路由...")
		factory := impl.NewRepositoryFactory(database.GetDB())
		fileService := service.NewFileService(factory, cache)
		router.SetupS3Router(r, factory, fileService)
		logger.LOG.Info("[路由] S3 API路由注册完成 ✔️")

		if config.CONFIG.S3.SharePort {
			logger.LOG.Info("[路由] S3服务共用主端口", "port", config.CONFIG.Server.Port)
		} else {
			logger.LOG.Info("[路由] S3服务独立端口（待实现）", "port", config.CONFIG.S3.Port)
		}
	}

	return r
}

// Execute 执行服务器
// 启动HTTP服务器并开始监听请求
func Execute(cacheLocal cache.Cache) {
	logger.LOG.Info("========== HTTP服务器启动 ==========")

	factory := impl.NewRepositoryFactory(database.GetDB())
	serverFactory := service.NewServiceFactory(factory, cacheLocal)
	// 启动回收站定时清理任务
	recycledTask := task.NewRecycledTask(factory)
	recycledTask.StartScheduledCleanup(30, 24*time.Hour)
	// 启动上传任务定时清理任务（每天清理一次过期任务）
	uploadTask := task.NewUploadTask(factory)
	uploadTask.StartScheduledCleanup(24 * time.Hour)

	// 启动S3生命周期管理定时任务（如果启用S3服务）
	if config.CONFIG.S3.Enable {
		logger.LOG.Info("[定时任务] 正在启动S3生命周期管理任务...")
		fileService := service.NewFileService(factory, cacheLocal)
		s3ObjectService := s3Service.NewS3ObjectService(factory, fileService)
		lifecycleTask := s3Task.NewLifecycleTask(factory, s3ObjectService)
		// 每小时执行一次生命周期规则检查
		lifecycleTask.StartScheduledExecution(1 * time.Hour)
		logger.LOG.Info("[定时任务] S3生命周期管理任务已启动 ✔️")
	}

	// 初始化路由
	router := initRouter(serverFactory, cacheLocal)

	// 构建监听地址
	addr := fmt.Sprintf("%s:%d", config.CONFIG.Server.Host, config.CONFIG.Server.Port)
	logger.LOG.Info("服务器将在以下地址启动", "address", addr)

	// 启动服务器
	logger.LOG.Info("服务器正在启动，按 Ctrl+C 停止...")
	if err := router.Run(addr); err != nil {
		logger.LOG.Error("服务器启动失败", "error", err)
		panic(fmt.Sprintf("HTTP服务器启动失败: %v", err))
	}
}
