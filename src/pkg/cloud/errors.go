package cloud

import "errors"

var (
	// ErrUnsupportedProvider 不支持的云盘提供者
	ErrUnsupportedProvider = errors.New("unsupported cloud provider")

	// ErrInvalidShareURL 无效的分享链接
	ErrInvalidShareURL = errors.New("invalid share URL")

	// ErrShareNotFound 分享不存在或已失效
	ErrShareNotFound = errors.New("share not found or expired")

	// ErrSharePasswordRequired 需要提取码
	ErrSharePasswordRequired = errors.New("share password required")

	// ErrSharePasswordWrong 提取码错误
	ErrSharePasswordWrong = errors.New("share password is wrong")

	// ErrFileNotFound 文件不存在
	ErrFileNotFound = errors.New("file not found")

	// ErrRateLimitExceeded 请求频率限制
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrAPIError API调用错误
	ErrAPIError = errors.New("API error")
)
