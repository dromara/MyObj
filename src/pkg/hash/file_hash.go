package hash

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"myobj/src/pkg/logger"

	"github.com/shirou/gopsutil/mem"
	"github.com/zeebo/blake3"
)

// 默认配置常量
const (
	DefaultChunkSize     = 256 * 1024             // 256KB 块大小
	MinFileSize          = 500 * 1024 * 1024      // 500MB 内存映射最小阈值
	MaxFileSize          = 2 * 1024 * 1024 * 1024 // 2GB 内存映射最大阈值
	Blake3MaxMemoryRatio = 0.4                    // Blake3最大内存使用比例
)

// FastBlake3Hasher 高性能Blake3哈希计算器
type FastBlake3Hasher struct {
	useMemoryMap bool       // 是否使用内存映射
	chunkSize    int        // 流式处理的块大小
	minSize      int64      // 内存映射的最小文件阈值
	maxSize      int64      // 内存映射的最大文件阈值
	osMemSize    uint64     // 操作系统内存大小
	verbose      bool       // 是否输出详细信息
	bufferPool   *sync.Pool // 缓冲区对象池
}

// MultipleFileResult 批量文件哈希计算结果
type MultipleFileResult struct {
	FileName string        // 文件名
	FileHash string        // 哈希值
	Duration time.Duration // 计算耗时
	Error    error         // 错误信息
}

// NewFastBlake3Hasher 创建新的Blake3哈希计算器
func NewFastBlake3Hasher() *FastBlake3Hasher {
	// 获取系统内存信息
	vmStat, err := mem.VirtualMemory()
	var totalMemory uint64
	if err != nil {
		if logger.LOG != nil {
			logger.LOG.Warn("获取系统内存失败，使用默认值", "error", err)
		}
		totalMemory = 8 * 1024 * 1024 * 1024 // 默认8GB
	} else {
		totalMemory = vmStat.Total
	}

	return &FastBlake3Hasher{
		useMemoryMap: true,
		chunkSize:    DefaultChunkSize,
		minSize:      MinFileSize,
		maxSize:      MaxFileSize,
		osMemSize:    totalMemory,
		verbose:      false,
		bufferPool: &sync.Pool{
			New: func() interface{} {
				buf := make([]byte, DefaultChunkSize)
				return &buf
			},
		},
	}
}

// WithMemoryMap 设置是否使用内存映射
func (h *FastBlake3Hasher) WithMemoryMap(enable bool) *FastBlake3Hasher {
	h.useMemoryMap = enable
	return h
}

// WithChunkSize 设置块大小（KB）
func (h *FastBlake3Hasher) WithChunkSize(sizeKB int) *FastBlake3Hasher {
	h.chunkSize = sizeKB * 1024
	// 更新缓冲池
	h.bufferPool = &sync.Pool{
		New: func() interface{} {
			buf := make([]byte, h.chunkSize)
			return &buf
		},
	}
	return h
}

// WithVerbose 设置是否输出详细信息
func (h *FastBlake3Hasher) WithVerbose(enable bool) *FastBlake3Hasher {
	h.verbose = enable
	return h
}

// ComputeFileHash 计算文件的Blake3哈希
func (h *FastBlake3Hasher) ComputeFileHash(filePath string) (string, time.Duration, error) {
	startTime := time.Now()

	// 检查文件大小决定使用哪种策略
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("获取文件信息失败: %w", err)
	}
	fileSize := fileInfo.Size()

	if h.verbose {
		if logger.LOG != nil {
			logger.LOG.Debug("文件哈希计算",
				"文件", filePath,
				"大小MB", float64(fileSize)/(1024*1024))
		}
	}

	// 计算可用内存阈值
	memoryThreshold := int64(float64(h.osMemSize) * Blake3MaxMemoryRatio)

	var hashResult string

	// 决定使用内存映射还是流式处理
	if h.useMemoryMap &&
		fileSize >= h.minSize &&
		fileSize <= h.maxSize &&
		fileSize <= memoryThreshold {

		if h.verbose {
			if logger.LOG != nil {
				logger.LOG.Debug("使用内存映射模式计算哈希")
			}
		}
		hashResult, err = h.hashWithMmap(filePath)
	} else {
		if h.verbose {
			if logger.LOG != nil {
				logger.LOG.Debug("使用流式处理模式计算哈希")
			}
		}
		hashResult, err = h.hashStreaming(filePath, fileSize)
	}

	if err != nil {
		return "", 0, err
	}

	duration := time.Since(startTime)
	if h.verbose {
		if logger.LOG != nil {
			logger.LOG.Debug("哈希计算完成",
				"耗时", duration,
				"哈希", hashResult)
		}
	}

	return hashResult, duration, nil
}

// hashWithMmap 使用内存映射方式计算哈希（最快）
func (h *FastBlake3Hasher) hashWithMmap(filePath string) (string, error) {
	// 读取整个文件到内存
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	// 使用Blake3计算哈希
	hash := blake3.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// hashStreaming 使用流式处理方式计算哈希（适用于超大文件）
func (h *FastBlake3Hasher) hashStreaming(filePath string, fileSize int64) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 创建Blake3哈希器
	hasher := blake3.New()

	// 从缓冲池获取缓冲区
	bufPtr := h.bufferPool.Get().(*[]byte)
	buffer := *bufPtr
	defer h.bufferPool.Put(bufPtr)

	var totalBytes int64
	progressInterval := int64(10 * 1024 * 1024) // 每10MB输出一次进度

	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("读取文件失败: %w", err)
		}
		if n == 0 {
			break
		}

		// 更新哈希
		hasher.Write(buffer[:n])
		totalBytes += int64(n)

		// 输出进度（仅在详细模式下）
		if h.verbose && totalBytes%progressInterval == 0 {
			if logger.LOG != nil {
				progress := float64(totalBytes) / float64(fileSize) * 100
				logger.LOG.Debug("哈希计算进度",
					"已处理MB", float64(totalBytes)/(1024*1024),
					"进度%", fmt.Sprintf("%.2f", progress))
			}
		}
	}

	// 获取最终哈希值
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes), nil
}

// ComputeMultipleFiles 批量计算多个文件的哈希
func (h *FastBlake3Hasher) ComputeMultipleFiles(filePaths []string) []MultipleFileResult {
	results := make([]MultipleFileResult, 0, len(filePaths))

	for _, filePath := range filePaths {
		hash, duration, err := h.ComputeFileHash(filePath)

		result := MultipleFileResult{
			FileName: filePath,
			FileHash: hash,
			Duration: duration,
			Error:    err,
		}

		if err != nil {
			if logger.LOG != nil {
				logger.LOG.Error("计算文件哈希失败", "文件", filePath, "error", err)
			}
		}

		results = append(results, result)
	}

	return results
}

// ComputeMultipleFilesConcurrent 并发批量计算多个文件的哈希
func (h *FastBlake3Hasher) ComputeMultipleFilesConcurrent(filePaths []string, maxConcurrency int) []MultipleFileResult {
	if maxConcurrency <= 0 {
		maxConcurrency = 4 // 默认4个并发
	}

	results := make([]MultipleFileResult, len(filePaths))
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrency)

	for i, filePath := range filePaths {
		wg.Add(1)
		go func(index int, path string) {
			defer wg.Done()

			// 获取信号量
			sem <- struct{}{}
			defer func() { <-sem }()

			hash, duration, err := h.ComputeFileHash(path)
			results[index] = MultipleFileResult{
				FileName: path,
				FileHash: hash,
				Duration: duration,
				Error:    err,
			}

			if err != nil {
				if logger.LOG != nil {
					logger.LOG.Error("计算文件哈希失败", "文件", path, "error", err)
				}
			}
		}(i, filePath)
	}

	wg.Wait()
	return results
}

// ComputeBytes 计算字节数组的Blake3哈希
func ComputeBytes(data []byte) string {
	hash := blake3.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// ComputeString 计算字符串的Blake3哈希
func ComputeString(s string) string {
	return ComputeBytes([]byte(s))
}

// VerifyFileHash 验证文件哈希是否匹配
func (h *FastBlake3Hasher) VerifyFileHash(filePath string, expectedHash string) (bool, error) {
	actualHash, _, err := h.ComputeFileHash(filePath)
	if err != nil {
		return false, err
	}
	return actualHash == expectedHash, nil
}
