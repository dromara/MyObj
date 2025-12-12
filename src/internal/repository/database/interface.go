package database

import (
	"fmt"
	"myobj/src/config"
	"myobj/src/pkg/logger"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// databasePool 全局数据库连接池实例
var databasePool *gorm.DB

// SQL 数据库接口定义
// 所有数据库实现(MySQL/SQLite)都需要实现此接口
type SQL interface {
	// GetDB 获取数据库连接实例
	GetDB() *gorm.DB
	// Ping 测试数据库连接是否可用
	Ping() error
	// InitDatabase 初始化数据库连接
	InitDatabase()
}

// InitDataBase 初始化数据库连接
// 根据配置文件中的数据库类型(mysql/sqlite)选择对应的数据库驱动进行初始化
func InitDataBase() {
	dbType := config.CONFIG.Database.Type
	logger.LOG.Info("[数据库] 开始初始化数据库连接", "type", dbType)

	switch dbType {
	case "mysql":
		initMySQL()
	case "sqlite":
		initSQLite()
	default:
		logger.LOG.Error("[数据库] 不支持的数据库类型", "type", dbType)
		panic(fmt.Sprintf("不支持的数据库类型: %s", dbType))
	}

	logger.LOG.Info("[数据库] 数据库连接池初始化成功 ✓")
}

// initMySQL 初始化MySQL数据库连接
func initMySQL() {
	logger.LOG.Info("[数据库] 正在连接MySQL数据库...",
		"host", config.CONFIG.Database.Host,
		"port", config.CONFIG.Database.Port,
		"database", config.CONFIG.Database.DBName)

	mysql := new(Mysql)
	mysql.InitDatabase()

	if err := mysql.Ping(); err != nil {
		logger.LOG.Error("[数据库] MySQL连接测试失败", "error", err)
		panic(fmt.Sprintf("MySQL数据库连接失败: %v", err))
	}

	databasePool = mysql.GetDB()
	logger.LOG.Info("[数据库] MySQL连接成功")
}

// initSQLite 初始化SQLite数据库连接
func initSQLite() {
	logger.LOG.Info("[数据库] 正在连接SQLite数据库...", "path", config.CONFIG.Database.Host)

	sqlite := new(SQLite)
	sqlite.InitDatabase()

	if err := sqlite.Ping(); err != nil {
		logger.LOG.Error("[数据库] SQLite连接测试失败", "error", err)
		panic(fmt.Sprintf("SQLite数据库连接失败: %v", err))
	}

	databasePool = sqlite.GetDB()
	logger.LOG.Info("[数据库] SQLite连接成功")
}

// GetDB 获取全局数据库连接池实例
// 返回已初始化的GORM数据库连接对象
func GetDB() *gorm.DB {
	return databasePool
}

// logLevel 将日志级别字符串转换为GORM日志级别
// 根据应用配置的日志级别返回对应的GORM日志级别
func logLevel(level string) gormlogger.LogLevel {
	switch level {
	case "debug":
		return gormlogger.Info // debug模式下显示SQL详细信息
	case "error":
		return gormlogger.Error
	case "warn":
		return gormlogger.Warn
	default:
		return gormlogger.Info
	}
}
