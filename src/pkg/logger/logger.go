package logger

import (
	"context"
	"log"
	"log/slog"
	"myobj/src/config"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

var LOG *slog.Logger

// MultiHandler 自定义 Handler：同时输出到控制台和文件
type MultiHandler struct {
	consoleHandler slog.Handler
	fileHandler    slog.Handler
}

func NewMultiHandler(consoleHandler, fileHandler slog.Handler) *MultiHandler {
	return &MultiHandler{
		consoleHandler: consoleHandler,
		fileHandler:    fileHandler,
	}
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.consoleHandler.Enabled(ctx, level) || h.fileHandler.Enabled(ctx, level)
}

func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	// 仅对 Info 及以上级别添加源代码位置
	if r.Level != slog.LevelInfo {
		// 获取调用栈信息 (跳过3层：runtime.Callers -> 此方法 -> slog记录点)
		var pcs [1]uintptr
		runtime.Callers(4, pcs[:])
		frames := runtime.CallersFrames(pcs[:])

		frame, _ := frames.Next()
		if frame.PC != 0 {
			// 提取简洁的文件名和函数名
			file := filepath.Base(frame.File)
			function := shortenFuncName(frame.Function)

			// 添加到日志属性
			r.AddAttrs(
				slog.String("source", file),
				slog.Int("line", frame.Line),
				slog.String("func", function),
			)
		}
	}
	err1 := h.consoleHandler.Handle(ctx, r)
	err2 := h.fileHandler.Handle(ctx, r)
	if err1 != nil {
		return err1
	}
	return err2
}

// 简化函数名 (去掉包路径)
func shortenFuncName(f string) string {
	// 去掉包路径前缀
	if idx := strings.LastIndex(f, "/"); idx != -1 {
		f = f[idx+1:]
	}

	// 去掉类型接收器部分
	if idx := strings.Index(f, "."); idx != -1 {
		return f[idx+1:]
	}
	return f
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewMultiHandler(
		h.consoleHandler.WithAttrs(attrs),
		h.fileHandler.WithAttrs(attrs),
	)
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	return NewMultiHandler(
		h.consoleHandler.WithGroup(name),
		h.fileHandler.WithGroup(name),
	)
}

// InitLogger 初始化日志系统
func InitLogger() {
	cfg := config.CONFIG.Log
	// 确保日志目录存在
	if err := os.MkdirAll(cfg.LogPath, 0750); err != nil {
		log.Fatalf("创建日志目录失败: %v", err)
	}

	// 配置日志级别
	logLevel := new(slog.LevelVar)
	switch cfg.Level {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "warn":
		logLevel.Set(slog.LevelWarn)
	case "error":
		logLevel.Set(slog.LevelError)
	default:
		logLevel.Set(slog.LevelInfo)
	}

	// 配置日志轮转规则
	rotationLog, err := rotatelogs.New(
		cfg.LogPath+"app.%Y%m%d.log",                                  // 按日期分片
		rotatelogs.WithRotationTime(24*time.Hour),                     // 每天轮转
		rotatelogs.WithMaxAge(time.Duration(cfg.MaxAge)*24*time.Hour), // 保留天数
		rotatelogs.WithRotationSize(int64(cfg.MaxSize)*1024*1024),     // 按大小分片
		//rotatelogs.WithRotationCount(uint(cfg.MaxBackups)),            // 保留文件数
	)
	if err != nil {
		log.Fatalf("创建日志轮转器失败: %v", err)
	}

	// 创建控制台处理器（文本格式）
	consoleHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// 自定义时间格式
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(time.Now().Format("2006-01-02 15:04:05"))
			}
			return a
		},
	})

	// 创建文件处理器（JSON格式）
	fileHandler := slog.NewJSONHandler(rotationLog, &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// 自定义时间格式
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(time.Now().Format("2006-01-02 15:04:05"))
			}
			return a
		},
	})

	// 创建组合处理器
	multiHandler := NewMultiHandler(consoleHandler, fileHandler)

	// 创建日志器
	logger := slog.New(multiHandler)
	slog.SetDefault(logger)
	LOG = logger
	LOG.Info("日志系统初始化完成🧩🧩🧩🧩🧩🧩")
}
