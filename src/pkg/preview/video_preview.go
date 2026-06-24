package preview

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"myobj/src/pkg/logger"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/image/draw"
)

var (
	ffmpegOnce     sync.Once
	ffmpegAvailable bool
)

// CheckFFmpegAvailable 检查 ffmpeg 是否可用（结果会被缓存）
func CheckFFmpegAvailable() bool {
	ffmpegOnce.Do(func() {
		_, err := exec.LookPath("ffmpeg")
		ffmpegAvailable = err == nil
		if ffmpegAvailable {
			logger.LOG.Info("ffmpeg 已检测到，视频缩略图功能可用")
		} else {
			logger.LOG.Warn("ffmpeg 未找到，视频缩略图功能不可用，将使用默认图标")
		}
	})
	return ffmpegAvailable
}

// GenerateVideoThumbnail 从视频文件生成缩略图
//
// 参数:
//   - videoPath: 视频文件路径
//   - outputPath: 输出缩略图文件路径
//   - timestamp: 截取时间点（如 "00:00:01"），空字符串则自动取视频开头
//   - maxDimension: 缩略图最大尺寸（保持宽高比）
//
// 支持格式:
//
//	所有 ffmpeg 支持的视频格式（mp4, avi, mkv, mov, wmv, flv, webm 等）
//
// 错误:
//
//	如果 ffmpeg 不可用、文件不存在或处理失败会返回错误
func GenerateVideoThumbnail(videoPath, outputPath, timestamp string, maxDimension uint) error {
	if !CheckFFmpegAvailable() {
		return fmt.Errorf("ffmpeg 不可用，无法生成视频缩略图")
	}

	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return fmt.Errorf("视频文件不存在: %s", videoPath)
	}

	if timestamp == "" {
		timestamp = "00:00:01"
	}

	frameData, err := extractVideoFrame(videoPath, timestamp)
	if err != nil {
		logger.LOG.Warn("首次截取视频帧失败，尝试从开头截取", "error", err, "path", videoPath)
		frameData, err = extractVideoFrame(videoPath, "00:00:00")
		if err != nil {
			return fmt.Errorf("截取视频帧失败: %w", err)
		}
	}

	img, _, err := image.Decode(bytes.NewReader(frameData))
	if err != nil {
		return fmt.Errorf("解码视频帧失败: %w", err)
	}

	resized := resizeImage(img, int(maxDimension))

	outDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer outFile.Close()

	if err := jpeg.Encode(outFile, resized, &jpeg.Options{Quality: 85}); err != nil {
		return fmt.Errorf("编码缩略图失败: %w", err)
	}

	logger.LOG.Debug("视频缩略图生成成功", "input", videoPath, "output", outputPath)
	return nil
}

// extractVideoFrame 使用 ffmpeg 从视频中截取指定时间点的帧
func extractVideoFrame(videoPath, timestamp string) ([]byte, error) {
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-ss", timestamp,
		"-vframes", "1",
		"-f", "image2",
		"-c:v", "mjpeg",
		"pipe:1",
	)
	cmd.Stderr = nil

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg 执行失败: %w", err)
	}

	if len(output) == 0 {
		return nil, fmt.Errorf("ffmpeg 输出为空")
	}

	return output, nil
}

// resizeImage 等比缩放图片到指定最大尺寸
func resizeImage(img image.Image, maxDimension int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width <= maxDimension && height <= maxDimension {
		return img
	}

	var newWidth, newHeight int
	if width > height {
		newWidth = maxDimension
		newHeight = int(float64(height) * float64(maxDimension) / float64(width))
	} else {
		newHeight = maxDimension
		newWidth = int(float64(width) * float64(maxDimension) / float64(height))
	}

	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}

// IsVideoByMime 根据 MIME 类型判断是否为视频
func IsVideoByMime(mimeType string) bool {
	return strings.HasPrefix(mimeType, "video/")
}
