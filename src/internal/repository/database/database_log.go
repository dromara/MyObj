package database

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm/logger"
)

// GormSlogAdapter GORM 适配器
type GormSlogAdapter struct {
	level logger.LogLevel
}

func (l *GormSlogAdapter) LogMode(level logger.LogLevel) logger.Interface {
	l.level = level
	return l
}

func (l *GormSlogAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Info {
		slog.InfoContext(ctx, msg, "gorm", data)
	}
}

func (l *GormSlogAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Warn {
		slog.WarnContext(ctx, msg, "gorm", data)
	}
}

func (l *GormSlogAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Error {
		slog.ErrorContext(ctx, msg, "gorm", data)
	}
}

func (l *GormSlogAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil {
		slog.ErrorContext(ctx, "gorm trace",
			"err", err,
			"sql", sql,
			"rows", rows,
			"elapsed", elapsed)
	} else {
		slog.DebugContext(ctx, "gorm trace",
			"sql", sql,
			"rows", rows,
			"elapsed", elapsed)
	}
}
