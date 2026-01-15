package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"myobj/src/pkg/logger"
	"os"
	"path/filepath"

	"golang.org/x/crypto/pbkdf2"
)

// EncryptionService 加密服务
type EncryptionService struct {
	masterKey []byte // 主密钥（用于加密数据密钥）
}

// NewEncryptionService 创建加密服务
func NewEncryptionService(masterKey string) *EncryptionService {
	// 如果没有提供主密钥，使用默认密钥（生产环境应该从配置读取）
	if masterKey == "" {
		masterKey = "default-master-key-change-in-production"
	}
	// 使用PBKDF2派生32字节主密钥
	key := pbkdf2.Key([]byte(masterKey), []byte("s3-encryption-salt"), 100000, 32, sha256.New)
	return &EncryptionService{
		masterKey: key,
	}
}

// GenerateDataKey 生成数据加密密钥
func (s *EncryptionService) GenerateDataKey() (keyID string, key []byte, err error) {
	// 生成32字节随机密钥（AES-256）
	key = make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", nil, fmt.Errorf("生成数据密钥失败: %w", err)
	}

	// 生成密钥ID（使用密钥的SHA256哈希的前16字节）
	hash := sha256.Sum256(key)
	keyID = base64.URLEncoding.EncodeToString(hash[:16])

	return keyID, key, nil
}

// EncryptDataKey 加密数据密钥（使用主密钥）
func (s *EncryptionService) EncryptDataKey(dataKey []byte) (string, error) {
	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return "", fmt.Errorf("创建AES密码器失败: %w", err)
	}

	// 使用GCM模式加密
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM失败: %w", err)
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("生成nonce失败: %w", err)
	}

	// 加密数据密钥
	ciphertext := gcm.Seal(nonce, nonce, dataKey, nil)

	// 返回base64编码的加密密钥
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptDataKey 解密数据密钥（使用主密钥）
func (s *EncryptionService) DecryptDataKey(encryptedKey string) ([]byte, error) {
	// 解码base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("解码加密密钥失败: %w", err)
	}

	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return nil, fmt.Errorf("创建AES密码器失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("创建GCM失败: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("密文太短")
	}

	// 提取nonce和密文
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("解密数据密钥失败: %w", err)
	}

	return plaintext, nil
}

// EncryptFile 加密文件（使用AES-256-CTR模式）
func (s *EncryptionService) EncryptFile(inputPath, outputPath string, dataKey []byte) (iv []byte, err error) {
	// 打开输入文件
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("打开输入文件失败: %w", err)
	}
	defer inputFile.Close()

	// 创建输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer func() {
		if closeErr := outputFile.Close(); closeErr != nil {
			logger.LOG.Error("关闭输出文件失败", "error", closeErr)
		}
	}()

	// 创建AES密码器
	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return nil, fmt.Errorf("创建AES密码器失败: %w", err)
	}

	// 生成随机IV（16字节，AES块大小）
	iv = make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("生成IV失败: %w", err)
	}

	// 使用CTR模式加密
	stream := cipher.NewCTR(block, iv)

	// 先写入IV
	if _, err := outputFile.Write(iv); err != nil {
		return nil, fmt.Errorf("写入IV失败: %w", err)
	}

	// 流式加密并写入
	buffer := make([]byte, 64*1024) // 64KB缓冲区
	for {
		n, err := inputFile.Read(buffer)
		if n > 0 {
			ciphertext := make([]byte, n)
			stream.XORKeyStream(ciphertext, buffer[:n])
			if _, err := outputFile.Write(ciphertext); err != nil {
				return nil, fmt.Errorf("写入密文失败: %w", err)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("读取文件失败: %w", err)
		}
	}

	return iv, nil
}

// DecryptFile 解密文件（使用AES-256-CTR模式）
func (s *EncryptionService) DecryptFile(inputPath, outputPath string, dataKey []byte, iv []byte) error {
	// 打开输入文件
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("打开输入文件失败: %w", err)
	}
	defer inputFile.Close()

	// 创建输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer func() {
		if closeErr := outputFile.Close(); closeErr != nil {
			logger.LOG.Error("关闭输出文件失败", "error", closeErr)
		}
	}()

	// 读取并验证IV（如果文件中有IV，跳过；否则使用提供的IV）
	fileIV := make([]byte, aes.BlockSize)
	if n, err := inputFile.Read(fileIV); err != nil && err != io.EOF {
		return fmt.Errorf("读取IV失败: %w", err)
	} else if n == aes.BlockSize {
		// 文件中有IV，使用文件中的IV
		iv = fileIV
	} else if iv == nil {
		return fmt.Errorf("IV未提供且文件中没有IV")
	}

	// 创建AES密码器
	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return fmt.Errorf("创建AES密码器失败: %w", err)
	}

	// 使用CTR模式解密
	stream := cipher.NewCTR(block, iv)

	// 流式解密并写入
	buffer := make([]byte, 64*1024) // 64KB缓冲区
	for {
		n, err := inputFile.Read(buffer)
		if n > 0 {
			plaintext := make([]byte, n)
			stream.XORKeyStream(plaintext, buffer[:n])
			if _, err := outputFile.Write(plaintext); err != nil {
				return fmt.Errorf("写入明文失败: %w", err)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取文件失败: %w", err)
		}
	}

	return nil
}

// EncryptStream 加密流（用于上传时实时加密）
func (s *EncryptionService) EncryptStream(reader io.Reader, writer io.Writer, dataKey []byte) (iv []byte, err error) {
	// 创建AES密码器
	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return nil, fmt.Errorf("创建AES密码器失败: %w", err)
	}

	// 生成随机IV
	iv = make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("生成IV失败: %w", err)
	}

	// 使用CTR模式加密
	stream := cipher.NewCTR(block, iv)

	// 先写入IV
	if _, err := writer.Write(iv); err != nil {
		return nil, fmt.Errorf("写入IV失败: %w", err)
	}

	// 流式加密
	buffer := make([]byte, 64*1024) // 64KB缓冲区
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			ciphertext := make([]byte, n)
			stream.XORKeyStream(ciphertext, buffer[:n])
			if _, err := writer.Write(ciphertext); err != nil {
				return nil, fmt.Errorf("写入密文失败: %w", err)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("读取流失败: %w", err)
		}
	}

	return iv, nil
}

// DecryptStream 解密流（用于下载时实时解密）
func (s *EncryptionService) DecryptStream(reader io.Reader, writer io.Writer, dataKey []byte, iv []byte) error {
	// 读取IV（如果流中有IV）
	if iv == nil {
		fileIV := make([]byte, aes.BlockSize)
		if n, err := reader.Read(fileIV); err != nil && err != io.EOF {
			return fmt.Errorf("读取IV失败: %w", err)
		} else if n == aes.BlockSize {
			iv = fileIV
		} else {
			return fmt.Errorf("IV未提供且流中没有IV")
		}
	}

	// 创建AES密码器
	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return fmt.Errorf("创建AES密码器失败: %w", err)
	}

	// 使用CTR模式解密
	stream := cipher.NewCTR(block, iv)

	// 流式解密
	buffer := make([]byte, 64*1024) // 64KB缓冲区
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			plaintext := make([]byte, n)
			stream.XORKeyStream(plaintext, buffer[:n])
			if _, err := writer.Write(plaintext); err != nil {
				return fmt.Errorf("写入明文失败: %w", err)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取流失败: %w", err)
		}
	}

	return nil
}

// GetEncryptedFilePath 获取加密文件的路径
func (s *EncryptionService) GetEncryptedFilePath(originalPath string) string {
	return originalPath + ".encrypted"
}

// GetDecryptedFilePath 获取解密文件的临时路径
func (s *EncryptionService) GetDecryptedFilePath(encryptedPath string) string {
	dir := filepath.Dir(encryptedPath)
	base := filepath.Base(encryptedPath)
	return filepath.Join(dir, ".decrypted_"+base)
}
