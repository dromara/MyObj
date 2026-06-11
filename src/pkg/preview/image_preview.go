package preview

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"myobj/src/pkg/logger"
	"os"
	"path/filepath"

	_ "image/png"

	// 图像处理库
	_ "golang.org/x/image/bmp" // BMP格式支持
	"golang.org/x/image/draw"
	_ "golang.org/x/image/tiff" // TIFF格式支持
	_ "golang.org/x/image/webp" // WebP格式支持

	// 标准库图像格式支持
	_ "image/gif"
	_ "image/jpeg"
)

// MaxPixelCount 最大允许的图片像素数（5000万像素）
const MaxPixelCount = 50_000_000

// GenerateImageThumbnail 生成图片缩略图
//
// 参数:
//   - inputPath: 输入图片文件路径
//   - outputPath: 输出缩略图文件路径
//   - maxDimension: 缩略图最大尺寸（保持宽高比）
//
// 支持格式:
//
//	JPEG, PNG, GIF, BMP, TIFF, WebP (通过标准库和golang.org/x/image扩展)
//
// 错误:
//
//	如果文件不存在、格式不支持或处理失败会返回错误
func GenerateImageThumbnail(inputPath, outputPath string, maxDimension uint) error {
	// 验证输入文件是否存在
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("输入文件不存在: %s", inputPath)
	}

	// 打开并解码图片
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("无法打开图片文件: %w", err)
	}
	defer file.Close()

	// 使用 image.DecodeConfig 先获取图片尺寸，避免超大图片直接解码导致内存溢出
	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("读取图片配置失败: %w", err)
	}

	// 检查图片像素数是否超过限制
	pixelCount := int64(config.Width) * int64(config.Height)
	if pixelCount > MaxPixelCount {
		return fmt.Errorf("图片尺寸过大(%dx%d, %d像素)，超过限制(%d像素)，拒绝处理",
			config.Width, config.Height, pixelCount, MaxPixelCount)
	}

	// 重置文件指针到开头，准备解码完整图片
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("重置文件指针失败: %w", err)
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("图片解码失败: %w", err)
	}
	logger.LOG.Info("图片解码成功", "format", format)
	// 获取原图尺寸并计算缩略图尺寸
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	newWidth, newHeight := calculateThumbnailDimensions(width, height, int(maxDimension))
	logger.LOG.Debug("计算缩略图尺寸", "original_width", width,
		"original_height", height,
		"thumbnail_width", newWidth,
		"thumbnail_height", newHeight)

	// 创建目标图像并缩放
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// 使用高质量的CatmullRom缩放算法（类似Lanczos3）
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	// 创建输出文件
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("无法创建输出文件: %w", err)
	}
	defer outFile.Close()

	// 根据文件扩展名选择编码器，直接使用已打开的文件句柄
	if err := encodeImageByExtension(outputPath, outFile, dst); err != nil {
		return fmt.Errorf("图片编码失败: %w", err)
	}
	logger.LOG.Debug("图片缩略图生成成功", "input_path", inputPath,
		"output_path", outputPath)
	return nil
}

// calculateThumbnailDimensions 计算保持宽高比的缩略图尺寸
func calculateThumbnailDimensions(width, height, maxDimension int) (int, int) {
	if width <= maxDimension && height <= maxDimension {
		return width, height
	}

	var newWidth, newHeight int

	if width > height {
		// 宽图
		newWidth = maxDimension
		newHeight = int(float64(height) * float64(maxDimension) / float64(width))
	} else {
		// 高图或方图
		newHeight = maxDimension
		newWidth = int(float64(width) * float64(maxDimension) / float64(height))
	}

	// 确保最小尺寸不为0
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	return newWidth, newHeight
}

// encodeImageByExtension 根据文件扩展名选择图像编码器，使用已打开的文件句柄
func encodeImageByExtension(outputPath string, outFile *os.File, img image.Image) error {
	ext := filepath.Ext(outputPath)

	switch ext {
	case ".jpg", ".jpeg":
		return encodeJPEG(outFile, img)
	case ".png":
		return encodePNG(outFile, img)
	case ".gif":
		return encodeGIF(outFile, img)
	default:
		// 默认使用PNG格式
		return encodePNG(outFile, img)
	}
}

// 各格式编码函数
func encodeJPEG(file *os.File, img image.Image) error {
	// 注意: 实际使用时需要导入 "image/jpeg"
	return jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
}

func encodePNG(file *os.File, img image.Image) error {
	// 注意: 实际使用时需要导入 "image/png"
	return png.Encode(file, img)
}

func encodeGIF(file *os.File, img image.Image) error {
	// 注意: 实际使用时需要导入 "image/gif"
	return gif.Encode(file, img, &gif.Options{})
}
