package cloud

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// DownloadOptions 下载选项
type DownloadOptions struct {
	Headers map[string]string
}

// DownloadToFile 下载URL内容到文件，支持进度回调
func DownloadToFile(ctx context.Context, url, filePath string, headers map[string]string, onProgress func(downloaded, total int64)) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	// 设置默认请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// 设置自定义请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("download failed: status=%d", resp.StatusCode)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create file failed: %w", err)
	}
	defer out.Close()

	buf := make([]byte, 32*1024)
	var downloaded int64
	total := resp.ContentLength
	lastUpdate := time.Now()

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)

			if onProgress != nil && time.Since(lastUpdate) > time.Second {
				onProgress(downloaded, total)
				lastUpdate = time.Now()
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return readErr
		}
	}

	if onProgress != nil {
		onProgress(downloaded, total)
	}

	return nil
}
