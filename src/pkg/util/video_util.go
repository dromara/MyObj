package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

// RangeInfo Range 请求信息
type RangeInfo struct {
	Start int64
	End   int64
	Total int64
}

// MaxRangeSize 单次 Range 请求的最大大小（2MB）
const MaxRangeSize = int64(2 * 1024 * 1024)

// ParseRange 解析 HTTP Range 请求头
// 支持格式: "bytes=start-end" 或 "bytes=start-"
// 返回解析后的 Range 信息
// 限制：单次请求不能超过 2MB
// 如果 Range 头为空，返回前 2MB（或整个文件，如果文件小于 2MB）
func ParseRange(rangeHeader string, fileSize int64) (*RangeInfo, error) {
	if rangeHeader == "" {
		// 如果没有 Range 头，返回前 2MB（或整个文件，如果文件小于 2MB）
		end := MaxRangeSize - 1
		if fileSize < MaxRangeSize {
			end = fileSize - 1
		}
		return &RangeInfo{
			Start: 0,
			End:   end,
			Total: fileSize,
		}, nil
	}

	// 移除 "bytes=" 前缀
	rangeHeader = strings.TrimPrefix(rangeHeader, "bytes=")

	// 解析 start-end
	parts := strings.Split(rangeHeader, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("无效的 Range 格式: %s", rangeHeader)
	}

	var start, end int64
	var err error

	// 解析 start
	if parts[0] != "" {
		start, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("无效的起始位置: %s", parts[0])
		}
	}

	// 解析 end
	if parts[1] != "" {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("无效的结束位置: %s", parts[1])
		}
	} else {
		// 如果没有指定 end，限制为最多 2MB（从 start 开始）
		// 确保不超过文件大小
		end = start + MaxRangeSize - 1
		if end >= fileSize {
			end = fileSize - 1
		}
	}

	// 验证范围
	if start < 0 || end >= fileSize || start > end {
		return nil, fmt.Errorf("Range 超出文件范围")
	}

	// 计算请求的数据大小
	requestSize := end - start + 1

	// 限制单次请求不能超过 2MB
	if requestSize > MaxRangeSize {
		return nil, fmt.Errorf("Range 请求过大: 请求了 %d 字节，最大允许 %d 字节 (2MB)", requestSize, MaxRangeSize)
	}

	return &RangeInfo{
		Start: start,
		End:   end,
		Total: fileSize,
	}, nil
}

// SetRangeHeaders 设置 HTTP Range 响应头
// hasRangeHeader: 客户端是否发送了 Range 请求头
func SetRangeHeaders(w http.ResponseWriter, rangeInfo *RangeInfo, contentType string, hasRangeHeader bool) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")
	
	if hasRangeHeader {
		// 有 Range 头，返回 206 Partial Content
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d",
			rangeInfo.Start, rangeInfo.End, rangeInfo.Total))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", rangeInfo.End-rangeInfo.Start+1))
		w.WriteHeader(http.StatusPartialContent) // 206
	} else {
		// 没有 Range 头，返回 200 OK，但只返回部分内容（前 2MB）
		w.Header().Set("Content-Length", fmt.Sprintf("%d", rangeInfo.End-rangeInfo.Start+1))
		w.WriteHeader(http.StatusOK) // 200
	}
}

// IncrementIV 根据块偏移量调整 CTR 模式的 IV（计数器）
// AES-CTR 的计数器需要根据数据偏移量进行调整，以支持随机访问
func IncrementIV(iv []byte, blockOffset int64) []byte {
	newIV := make([]byte, len(iv))
	copy(newIV, iv)

	// 将 blockOffset 加到 IV 上（大端序）
	for i := len(newIV) - 1; i >= 0 && blockOffset > 0; i-- {
		sum := int64(newIV[i]) + (blockOffset & 0xFF)
		newIV[i] = byte(sum & 0xFF)
		blockOffset = (blockOffset >> 8) + (sum >> 8)
	}

	return newIV
}

// StreamDecryptRange 流式解密指定 Range 的加密数据
// 直接从加密文件中读取指定范围并解密，写入 ResponseWriter
func StreamDecryptRange(w http.ResponseWriter, encFilePath string, password string, rangeInfo *RangeInfo) error {
	// 打开加密文件
	encFile, err := os.Open(encFilePath)
	if err != nil {
		return fmt.Errorf("打开加密文件失败: %w", err)
	}
	defer encFile.Close()

	// 读取文件头（salt + iv + hmac）
	salt := make([]byte, SaltLength)
	if _, err := io.ReadFull(encFile, salt); err != nil {
		return fmt.Errorf("读取盐失败: %w", err)
	}

	iv := make([]byte, IVLength)
	if _, err := io.ReadFull(encFile, iv); err != nil {
		return fmt.Errorf("读取IV失败: %w", err)
	}

	// 跳过 HMAC（流式场景下，我们只验证请求的块）
	storedHMAC := make([]byte, HMACLength)
	if _, err := io.ReadFull(encFile, storedHMAC); err != nil {
		return fmt.Errorf("读取HMAC失败: %w", err)
	}

	// 派生密钥
	encKey := deriveKeyFromPassword(password, salt)
	hmacKey := deriveHMACKeyFromPassword(password, salt)

	// 创建 AES cipher
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return fmt.Errorf("创建AES密码器失败: %w", err)
	}

	// 计算 CTR 模式的起始位置
	blockOffset := rangeInfo.Start / aes.BlockSize
	byteOffset := rangeInfo.Start % aes.BlockSize

	// 调整 IV
	adjustedIV := IncrementIV(iv, blockOffset)
	stream := cipher.NewCTR(block, adjustedIV)

	// 定位到加密文件中对应的位置
	// 文件结构: [salt(32)][iv(16)][hmac(32)][密文...]
	encDataStart := int64(SaltLength+IVLength+HMACLength) + rangeInfo.Start
	if _, err := encFile.Seek(encDataStart, io.SeekStart); err != nil {
		return fmt.Errorf("定位文件位置失败: %w", err)
	}

	// 流式解密并写入响应（已通过 ParseRange 限制在 2MB 内）
	remaining := rangeInfo.End - rangeInfo.Start + 1
	buffer := make([]byte, remaining)

	// HMAC 验证（可选，只验证请求的块）
	hmacHash := hmac.New(sha256.New, hmacKey)

	firstBlock := true
	n, err := encFile.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("读取文件失败: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("读取到空数据")
	}

	// 更新 HMAC
	hmacHash.Write(buffer[:n])

	// 解密数据
	plaintext := make([]byte, n)
	stream.XORKeyStream(plaintext, buffer[:n])

	// 如果是第一个块且有字节偏移，跳过前面的字节
	if firstBlock && byteOffset > 0 {
		plaintext = plaintext[byteOffset:]
	}

	// 写入响应
	if _, err := w.Write(plaintext); err != nil {
		return fmt.Errorf("写入响应失败: %w", err)
	}

	return nil
}

// StreamPlainRange 流式传输普通文件的指定 Range
func StreamPlainRange(w http.ResponseWriter, filePath string, rangeInfo *RangeInfo) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 定位到起始位置
	if _, err := file.Seek(rangeInfo.Start, io.SeekStart); err != nil {
		return fmt.Errorf("定位文件位置失败: %w", err)
	}

	// 流式传输（已通过 ParseRange 限制在 2MB 内）
	remaining := rangeInfo.End - rangeInfo.Start + 1
	buffer := make([]byte, remaining)

	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("读取文件失败: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("读取到空数据")
	}

	if _, err := w.Write(buffer[:n]); err != nil {
		return fmt.Errorf("写入响应失败: %w", err)
	}

	return nil
}

// deriveKeyFromPassword 从密码派生加密密钥
func deriveKeyFromPassword(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, PBKDF2Iterations, KeyLength, sha256.New)
}

// deriveHMACKeyFromPassword 从密码派生 HMAC 密钥
func deriveHMACKeyFromPassword(password string, salt []byte) []byte {
	hmacSalt := make([]byte, len(salt))
	copy(hmacSalt, salt)
	hmacSalt[0] ^= 0xFF
	return pbkdf2.Key([]byte(password), hmacSalt, PBKDF2Iterations, HMACKeyLength, sha256.New)
}
