package internal

import (
	"net/http"
	"time"
)

const DefaultRequestTimeout = 30 * time.Second

// DefaultHTTPClient 共享 HTTP 客户端
func DefaultHTTPClient() *http.Client {
	return &http.Client{Timeout: DefaultRequestTimeout}
}
