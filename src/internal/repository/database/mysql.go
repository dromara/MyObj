package database

import (
	"fmt"
	"myobj/src/config"
	"myobj/src/pkg/logger"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Mysql struct {
	database *gorm.DB
}

func (sql *Mysql) InitDatabase() {
	dbConfig := config.CONFIG.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: &GormSlogAdapter{
			level: logLevel(config.CONFIG.Log.Level),
		},
	})
	if err != nil {
		logger.LOG.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	sqlDB, err := db.DB()
	if err != nil {
		logger.LOG.Error("Failed to get database instance", err)
	}
	// 设置连接池参数
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpen)                                     // 最大连接数
	sqlDB.SetMaxIdleConns(dbConfig.MaxIdle)                                     // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(time.Duration(dbConfig.MaxLife) * time.Hour)       // 连接最大存活时间
	sqlDB.SetConnMaxIdleTime(time.Duration(dbConfig.MaxIdleLife) * time.Minute) // 空闲连接最大存活时间
	logger.LOG.Info("数据库连接成功📡")
	sql.database = db
}

func (sql *Mysql) GetDB() *gorm.DB {
	return sql.database
}

func (sql *Mysql) Ping() error {
	sqlDB, err := sql.database.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
