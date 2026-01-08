package database

import (
	"myobj/src/config"
	"myobj/src/pkg/logger"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type SQLite struct {
	database *gorm.DB
}

func (sql *SQLite) InitDatabase() {
	host := config.CONFIG.Database.Host
	db, err := gorm.Open(sqlite.Open(host), &gorm.Config{
		Logger: &GormSlogAdapter{
			level: logLevel(config.CONFIG.Log.Level),
		},
	})
	if err != nil {
		logger.LOG.Error("failed to connect database", "err", err)
		panic("failed to connect database")
	}
	sql.database = db
}
func (sql *SQLite) GetDB() *gorm.DB {
	return sql.database
}
func (sql *SQLite) Ping() error {
	sqlDB, err := sql.database.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
