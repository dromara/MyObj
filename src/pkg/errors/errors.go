package errors

import (
	"errors"
	"fmt"
)

// 预定义的业务错误（Sentinel Errors）
var (
	// 用户相关
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUserDisabled      = errors.New("用户已被禁用")
	ErrInvalidPassword   = errors.New("密码错误")
	ErrUserAlreadyExists = errors.New("用户已存在")
	ErrSpaceExceeded     = errors.New("用户空间不足")

	// 文件相关
	ErrFileNotFound      = errors.New("文件不存在")
	ErrFileAlreadyExists = errors.New("文件已存在")
	ErrFileTooLarge      = errors.New("文件过大")
	ErrInvalidFileType   = errors.New("无效的文件类型")
	ErrDiskFull          = errors.New("磁盘空间不足")
	ErrUploadFailed      = errors.New("上传失败")
	ErrDownloadFailed    = errors.New("下载失败")

	// 目录相关
	ErrDirNotFound      = errors.New("目录不存在")
	ErrDirAlreadyExists = errors.New("目录已存在")
	ErrDirNotEmpty      = errors.New("目录不为空")

	// 分享相关
	ErrShareNotFound    = errors.New("分享不存在")
	ErrShareExpired     = errors.New("分享已过期")
	ErrSharePasswordErr = errors.New("分享密码错误")

	// 权限相关
	ErrPermissionDenied = errors.New("权限不足")
	ErrUnauthorized     = errors.New("未授权")

	// 系统相关
	ErrInternal     = errors.New("内部错误")
	ErrInvalidParam = errors.New("参数错误")
	ErrRateLimit    = errors.New("请求过于频繁")

	// S3 相关
	ErrBucketNotFound    = errors.New("Bucket 不存在")
	ErrBucketAlreadyExists = errors.New("Bucket 已存在")
	ErrObjectNotFound    = errors.New("对象不存在")
	ErrNoSuchUpload      = errors.New("上传任务不存在")
)

// Wrap 包装错误，添加上下文信息
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf 包装错误，添加格式化的上下文信息
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	message := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", message, err)
}

// Is 检查错误链中是否包含目标错误
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As 将错误转换为指定类型
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// New 创建新错误
func New(text string) error {
	return errors.New(text)
}
