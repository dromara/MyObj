package hash

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/zeebo/blake3"
)

// QuickHashConfig 快速hash配置
type QuickHashConfig struct {
	// 采样分片数量（用于快速预判）
	SampleChunkCount int
	// 每个分片大小（默认4MB）
	ChunkSize int64
	// 是否计算全量hash
	ComputeFullHash bool
}

// QuickHashResult 快速hash计算结果
type QuickHashResult struct {
	// 文件大小
	FileSize int64
	// 分片签名（前N个分片的组合hash，用于快速匹配）
	ChunkSignature string
	// 全量文件hash（如果计算的话）
	FullHash string
	// 采样的分片hash列表
	ChunkHashes []string
}

// DefaultQuickHashConfig 默认配置：采样前3个分片，每片4MB
func DefaultQuickHashConfig() *QuickHashConfig {
	return &QuickHashConfig{
		SampleChunkCount: 3,
		ChunkSize:        4 * 1024 * 1024, // 4MB
		ComputeFullHash:  false,
	}
}

// ComputeQuickHash 快速计算文件hash（仅计算前几个分片）
// 适用场景：客户端上传前快速预检，避免全量hash计算耗时
func ComputeQuickHash(filePath string, config *QuickHashConfig) (*QuickHashResult, error) {
	if config == nil {
		config = DefaultQuickHashConfig()
	}

	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}
	fileSize := fileInfo.Size()

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	result := &QuickHashResult{
		FileSize:    fileSize,
		ChunkHashes: make([]string, 0, config.SampleChunkCount),
	}

	// 计算采样分片的hash
	chunkBuffer := make([]byte, config.ChunkSize)
	signatureHasher := blake3.New() // 用于生成分片签名

	for i := 0; i < config.SampleChunkCount; i++ {
		// 读取一个分片
		n, err := io.ReadFull(file, chunkBuffer)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return nil, fmt.Errorf("读取分片失败: %w", err)
		}
		if n == 0 {
			break // 文件已读完
		}

		// 计算当前分片的hash
		chunkHash := blake3.Sum256(chunkBuffer[:n])
		chunkHashStr := hex.EncodeToString(chunkHash[:])
		result.ChunkHashes = append(result.ChunkHashes, chunkHashStr)

		// 更新分片签名（所有采样分片hash的组合）
		signatureHasher.Write(chunkHash[:])
	}

	// 生成分片签名（用于快速匹配）
	signatureBytes := signatureHasher.Sum(nil)
	result.ChunkSignature = hex.EncodeToString(signatureBytes)

	// 如果需要计算全量hash
	if config.ComputeFullHash {
		// 重置文件指针
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("重置文件指针失败: %w", err)
		}

		// 使用FastBlake3Hasher计算完整hash
		hasher := NewFastBlake3Hasher()
		fullHash, _, err := hasher.ComputeFileHash(filePath)
		if err != nil {
			return nil, fmt.Errorf("计算全量hash失败: %w", err)
		}
		result.FullHash = fullHash
	}

	return result, nil
}

// ComputeChunkSignatureFromHashes 根据分片hash列表计算分片签名
// 用于服务端从已存储的分片hash重建签名
func ComputeChunkSignatureFromHashes(chunkHashes []string) (string, error) {
	if len(chunkHashes) == 0 {
		return "", fmt.Errorf("分片hash列表为空")
	}

	signatureHasher := blake3.New()
	for _, hashStr := range chunkHashes {
		hashBytes, err := hex.DecodeString(hashStr)
		if err != nil {
			return "", fmt.Errorf("解析分片hash失败: %w", err)
		}
		signatureHasher.Write(hashBytes)
	}

	signatureBytes := signatureHasher.Sum(nil)
	return hex.EncodeToString(signatureBytes), nil
}

// VerifyQuickHash 验证文件的快速hash
func VerifyQuickHash(filePath string, expectedSignature string, config *QuickHashConfig) (bool, error) {
	result, err := ComputeQuickHash(filePath, config)
	if err != nil {
		return false, err
	}
	return result.ChunkSignature == expectedSignature, nil
}
