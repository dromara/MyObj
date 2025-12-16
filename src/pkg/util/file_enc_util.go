package util

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"myobj/src/pkg/logger"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/mem"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/sync/semaphore"
)

// 加密配置常量
const (
	SaltLength       = 32
	IVLength         = 16 // AES块大小，用于CTR模式
	KeyLength        = 32
	HMACKeyLength    = 32
	HMACLength       = 32
	BufferSize       = 1024 * 1024 // 1MB 缓冲区
	MaxMemoryRatio   = 0.3         // 内存占用比例
	PBKDF2Iterations = 100000
	SmallFileLimit   = 100 * 1024 * 1024  // 100MB
	LargeFileLimit   = 1024 * 1024 * 1024 // 1GB
)

// FileCrypto 文件加密处理器
type FileCrypto struct {
	password      string
	maxConcurrent *semaphore.Weighted
	bufferPool    *sync.Pool // 缓冲区对象池
}

// EncryptionMethod 加密方法枚举
type EncryptionMethod int

const (
	MemoryMapped EncryptionMethod = iota
	Buffered
	Streaming
)

// NewFileCrypto 创建新的文件加密处理器
func NewFileCrypto(password string) *FileCrypto {
	return &FileCrypto{
		password:      password,
		maxConcurrent: semaphore.NewWeighted(int64(runtime.NumCPU())),
		bufferPool: &sync.Pool{
			New: func() interface{} {
				buf := make([]byte, BufferSize)
				return &buf
			},
		},
	}
}

// DeriveEncryptionKey 从明文密码和用户盐派生加密密钥
// 使用PBKDF2确保从相同的密码和盐总是派生出相同的密钥
func DeriveEncryptionKey(password string, userSalt string) string {
	// 将用户ID或其他唯一标识作为盐
	// 这样每个用户的密钥都不同，但对同一用户总是相同
	salt := []byte(userSalt)

	// 使用PBKDF2派生32字节密钥
	// 注意：这里使用较少的迭代次数(10000)，因为我们还会在加密时再次使用PBKDF2
	derivedKey := pbkdf2.Key([]byte(password), salt, 10000, 32, sha256.New)

	// 返回base64编码的密钥，便于存储和使用
	return string(derivedKey)
}

// EncryptFile 加密文件
func (fc *FileCrypto) EncryptFile(inputPath, outputPath string) error {
	if err := fc.maxConcurrent.Acquire(context.Background(), 1); err != nil {
		return fmt.Errorf("获取并发许可失败: %w", err)
	}
	defer fc.maxConcurrent.Release(1)

	// 获取文件信息
	inputInfo, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}
	fileSize := inputInfo.Size()

	startTime := time.Now()

	// 选择加密方法
	method, err := fc.selectEncryptionMethod(fileSize)
	if err != nil {
		return err
	}

	switch method {
	case MemoryMapped:
		err = fc.encryptWithMmap(inputPath, outputPath)
	case Buffered:
		err = fc.encryptWithBuffered(inputPath, outputPath)
	case Streaming:
		err = fc.encryptWithStreaming(inputPath, outputPath)
	}

	if err != nil {
		return err
	}

	duration := time.Since(startTime)
	log.Printf("加密完成: %s, 耗时: %v", outputPath, duration)
	return nil
}

// DecryptFile 解密文件
func (fc *FileCrypto) DecryptFile(inputPath, outputPath string) error {
	if err := fc.maxConcurrent.Acquire(context.Background(), 1); err != nil {
		return fmt.Errorf("获取并发许可失败: %w", err)
	}
	defer fc.maxConcurrent.Release(1)

	// 获取文件信息
	inputInfo, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}
	fileSize := inputInfo.Size()

	startTime := time.Now()

	// 选择解密方法
	method, err := fc.selectDecryptionMethod(fileSize)
	if err != nil {
		return err
	}

	switch method {
	case MemoryMapped:
		err = fc.decryptWithMmap(inputPath, outputPath)
	case Buffered:
		err = fc.decryptWithBuffered(inputPath, outputPath)
	case Streaming:
		err = fc.decryptWithStreaming(inputPath, outputPath)
	}

	if err != nil {
		return err
	}

	duration := time.Since(startTime)
	log.Printf("解密完成: %s, 耗时: %v", outputPath, duration)
	return nil
}

// selectEncryptionMethod 根据文件大小选择加密方法
func (fc *FileCrypto) selectEncryptionMethod(fileSize int64) (EncryptionMethod, error) {
	// 小于100MB的文件使用缓冲处理
	if fileSize <= SmallFileLimit {
		return Buffered, nil
	}

	// 获取系统可用内存
	totalMemory, availMemory, err := fc.GetSystemMemory()
	if err != nil {
		logger.LOG.Debug("获取系统内存失败，使用流式处理", "error", err)
		return Streaming, nil
	}

	// 使用可用内存的30%作为阈值
	memoryThreshold := int64(float64(availMemory) * MaxMemoryRatio)
	logger.LOG.Debug("系统内存信息",
		"总计MB", totalMemory/(1024*1024),
		"可用MB", availMemory/(1024*1024),
		"阈值MB", memoryThreshold/(1024*1024),
		"文件大小MB", fileSize/(1024*1024))

	// 文件大小在100MB到1GB之间，且可用内存充足，使用内存映射
	if fileSize <= LargeFileLimit && fileSize <= memoryThreshold {
		return MemoryMapped, nil
	}

	// 其他情况使用流式处理
	return Streaming, nil
}

// selectDecryptionMethod 根据文件大小选择解密方法
func (fc *FileCrypto) selectDecryptionMethod(fileSize int64) (EncryptionMethod, error) {
	// 加密文件包含 salt + iv + hmac + 数据
	if fileSize < int64(SaltLength+IVLength+HMACLength) {
		return 0, fmt.Errorf("文件太小，不是有效的加密文件")
	}

	originalSize := fileSize - int64(SaltLength+IVLength+HMACLength)

	// 小于100MB的文件使用缓冲处理
	if originalSize <= SmallFileLimit {
		return Buffered, nil
	}

	// 获取系统可用内存
	_, availMemory, err := fc.GetSystemMemory()
	if err != nil {
		logger.LOG.Debug("获取系统内存失败，使用流式处理", "error", err)
		return Streaming, nil
	}

	memoryThreshold := int64(float64(availMemory) * MaxMemoryRatio)

	// 文件大小在100MB到1GB之间，且可用内存充足，使用内存映射
	if originalSize <= LargeFileLimit && originalSize <= memoryThreshold {
		return MemoryMapped, nil
	}

	// 其他情况使用流式处理
	return Streaming, nil
}

// GetSystemMemory 获取系统总内存和可用内存（跨平台方案）
// 返回值: (总内存, 可用内存, 错误)
func (fc *FileCrypto) GetSystemMemory() (uint64, uint64, error) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, fmt.Errorf("获取系统内存信息失败: %w", err)
	}

	return vmStat.Total, vmStat.Available, nil
}

// encryptWithMmap 使用内存映射加密（CTR模式+HMAC认证）
func (fc *FileCrypto) encryptWithMmap(inputPath, outputPath string) error {
	// 读取输入文件
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 生成盐和IV
	salt := make([]byte, SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("生成盐失败: %w", err)
	}

	iv := make([]byte, IVLength)
	if _, err := rand.Read(iv); err != nil {
		return fmt.Errorf("生成IV失败: %w", err)
	}

	// 派生加密密钥和HMAC密钥
	encKey := fc.deriveKey(salt)
	hmacKey := fc.deriveHMACKey(salt)

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return fmt.Errorf("创建AES密码器失败: %w", err)
	}

	// 使用CTR模式加密
	stream := cipher.NewCTR(block, iv)
	ciphertext := make([]byte, len(data))
	stream.XORKeyStream(ciphertext, data)

	// 计算HMAC
	hmacHash := hmac.New(sha256.New, hmacKey)
	hmacHash.Write(ciphertext)
	finalHMAC := hmacHash.Sum(nil)

	// 写入输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer func(outputFile *os.File) {
		err := outputFile.Close()
		if err != nil {
			logger.LOG.Error("关闭输出文件失败", "error", err)
		}
	}(outputFile)

	// 写入: 盐 + IV + HMAC + 密文
	if _, err := outputFile.Write(salt); err != nil {
		return fmt.Errorf("写入盐失败: %w", err)
	}
	if _, err := outputFile.Write(iv); err != nil {
		return fmt.Errorf("写入IV失败: %w", err)
	}
	if _, err := outputFile.Write(finalHMAC); err != nil {
		return fmt.Errorf("写入HMAC失败: %w", err)
	}
	if _, err := outputFile.Write(ciphertext); err != nil {
		return fmt.Errorf("写入密文失败: %w", err)
	}

	return nil
}

// encryptWithBuffered 使用缓冲方式加密
func (fc *FileCrypto) encryptWithBuffered(inputPath, outputPath string) error {
	return fc.encryptWithMmap(inputPath, outputPath) // 对于小文件，使用相同实现
}

// encryptWithStreaming 使用流式加密（CTR模式+HMAC认证）
func (fc *FileCrypto) encryptWithStreaming(inputPath, outputPath string) error {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("打开输入文件失败: %w", err)
	}
	defer func(inputFile *os.File) {
		err := inputFile.Close()
		if err != nil {
			logger.LOG.Error("关闭输出文件失败", "error", err)
		}
	}(inputFile)

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer func() {
		if cerr := outputFile.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("关闭输出文件失败: %w", cerr)
		}
	}()

	// 生成盐和IV
	salt := make([]byte, SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("生成盐失败: %w", err)
	}

	iv := make([]byte, IVLength)
	if _, err := rand.Read(iv); err != nil {
		return fmt.Errorf("生成IV失败: %w", err)
	}

	// 写入盐和IV（预留HMAC位置）
	if _, err := outputFile.Write(salt); err != nil {
		return fmt.Errorf("写入盐失败: %w", err)
	}
	if _, err := outputFile.Write(iv); err != nil {
		return fmt.Errorf("写入IV失败: %w", err)
	}
	// 预留HMAC位置
	hmacPlaceholder := make([]byte, HMACLength)
	if _, err := outputFile.Write(hmacPlaceholder); err != nil {
		return fmt.Errorf("写入HMAC占位符失败: %w", err)
	}

	// 派生加密密钥和HMAC密钥
	encKey := fc.deriveKey(salt)
	hmacKey := fc.deriveHMACKey(salt)

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return fmt.Errorf("创建AES密码器失败: %w", err)
	}

	// 使用CTR模式进行流式加密
	stream := cipher.NewCTR(block, iv)

	// 创建HMAC
	hmacHash := hmac.New(sha256.New, hmacKey)

	// 从对象池获取缓冲区
	bufPtr := fc.bufferPool.Get().(*[]byte)
	buffer := *bufPtr
	defer fc.bufferPool.Put(bufPtr)

	cipherBufPtr := fc.bufferPool.Get().(*[]byte)
	cipherBuffer := *cipherBufPtr
	defer fc.bufferPool.Put(cipherBufPtr)

	for {
		n, err := inputFile.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("读取文件失败: %w", err)
		}
		if n == 0 {
			break
		}

		// 加密数据
		stream.XORKeyStream(cipherBuffer[:n], buffer[:n])

		// 更新HMAC
		hmacHash.Write(cipherBuffer[:n])

		if _, err := outputFile.Write(cipherBuffer[:n]); err != nil {
			return fmt.Errorf("写入密文失败: %w", err)
		}
	}

	// 计算最终HMAC
	finalHMAC := hmacHash.Sum(nil)

	// 回写HMAC到文件头部
	if _, err := outputFile.Seek(int64(SaltLength+IVLength), io.SeekStart); err != nil {
		return fmt.Errorf("定位HMAC位置失败: %w", err)
	}
	if _, err := outputFile.Write(finalHMAC); err != nil {
		return fmt.Errorf("写入HMAC失败: %w", err)
	}

	return nil
}

// decryptWithMmap 使用内存映射解密（CTR模式+HMAC验证）
func (fc *FileCrypto) decryptWithMmap(inputPath, outputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	if len(data) < SaltLength+IVLength+HMACLength {
		return fmt.Errorf("文件太小，不可能是有效的加密文件")
	}

	// 提取: 盐 + IV + HMAC + 密文
	salt := data[:SaltLength]
	iv := data[SaltLength : SaltLength+IVLength]
	storedHMAC := data[SaltLength+IVLength : SaltLength+IVLength+HMACLength]
	ciphertext := data[SaltLength+IVLength+HMACLength:]

	// 派生加密密钥和HMAC密钥
	encKey := fc.deriveKey(salt)
	hmacKey := fc.deriveHMACKey(salt)

	// 验证HMAC
	hmacHash := hmac.New(sha256.New, hmacKey)
	hmacHash.Write(ciphertext)
	computedHMAC := hmacHash.Sum(nil)

	if !hmac.Equal(storedHMAC, computedHMAC) {
		return fmt.Errorf("HMAC验证失败，文件可能已被篡改或密码错误")
	}

	// 解密数据
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return fmt.Errorf("创建AES密码器失败: %w", err)
	}

	stream := cipher.NewCTR(block, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	// 写入输出文件
	if err := os.WriteFile(outputPath, plaintext, 0644); err != nil {
		return fmt.Errorf("写入输出文件失败: %w", err)
	}

	return nil
}

// decryptWithBuffered 使用缓冲方式解密
func (fc *FileCrypto) decryptWithBuffered(inputPath, outputPath string) error {
	return fc.decryptWithMmap(inputPath, outputPath) // 对于小文件，使用相同实现
}

// decryptWithStreaming 使用流式解密（CTR模式+HMAC验证）
func (fc *FileCrypto) decryptWithStreaming(inputPath, outputPath string) error {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("打开输入文件失败: %w", err)
	}
	defer func(inputFile *os.File) {
		err := inputFile.Close()
		if err != nil {
			logger.LOG.Error("关闭输出文件失败", "error", err)
		}
	}(inputFile)

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer func() {
		if cerr := outputFile.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("关闭输出文件失败: %w", cerr)
		}
	}()

	// 读取盐
	salt := make([]byte, SaltLength)
	if _, err := io.ReadFull(inputFile, salt); err != nil {
		return fmt.Errorf("读取盐失败: %w", err)
	}

	// 读取IV
	iv := make([]byte, IVLength)
	if _, err := io.ReadFull(inputFile, iv); err != nil {
		return fmt.Errorf("读取IV失败: %w", err)
	}

	// 读取HMAC
	storedHMAC := make([]byte, HMACLength)
	if _, err := io.ReadFull(inputFile, storedHMAC); err != nil {
		return fmt.Errorf("读取HMAC失败: %w", err)
	}

	// 派生加密密钥和HMAC密钥
	encKey := fc.deriveKey(salt)
	hmacKey := fc.deriveHMACKey(salt)

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return fmt.Errorf("创建AES密码器失败: %w", err)
	}

	// 使用CTR模式进行流式解密
	stream := cipher.NewCTR(block, iv)

	// 创建HMAC用于验证
	hmacHash := hmac.New(sha256.New, hmacKey)

	// 从对象池获取缓冲区
	bufPtr := fc.bufferPool.Get().(*[]byte)
	buffer := *bufPtr
	defer fc.bufferPool.Put(bufPtr)

	plainBufPtr := fc.bufferPool.Get().(*[]byte)
	plainBuffer := *plainBufPtr
	defer fc.bufferPool.Put(plainBufPtr)

	for {
		n, err := inputFile.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("读取文件失败: %w", err)
		}
		if n == 0 {
			break
		}

		// 验证HMAC
		hmacHash.Write(buffer[:n])

		// 解密数据
		stream.XORKeyStream(plainBuffer[:n], buffer[:n])

		if _, err := outputFile.Write(plainBuffer[:n]); err != nil {
			return fmt.Errorf("写入明文失败: %w", err)
		}
	}

	// 验证HMAC
	computedHMAC := hmacHash.Sum(nil)
	if !hmac.Equal(storedHMAC, computedHMAC) {
		// HMAC验证失败，删除输出文件
		err := os.Remove(outputPath)
		if err != nil {
			return err
		}
		return fmt.Errorf("HMAC验证失败，文件可能已被篡改或密码错误")
	}

	return nil
}

// deriveKey 从密码派生加密密钥
func (fc *FileCrypto) deriveKey(salt []byte) []byte {
	return pbkdf2.Key([]byte(fc.password), salt, PBKDF2Iterations, KeyLength, sha256.New)
}

// deriveHMACKey 从密码派生HMAC密钥
func (fc *FileCrypto) deriveHMACKey(salt []byte) []byte {
	// 使用不同的盐派生HMAC密钥
	hmacSalt := make([]byte, len(salt))
	copy(hmacSalt, salt)
	// 修改盐的第一个字节以生成不同的密钥
	hmacSalt[0] ^= 0xFF
	return pbkdf2.Key([]byte(fc.password), hmacSalt, PBKDF2Iterations, HMACKeyLength, sha256.New)
}

// EncryptFiles 批量加密文件
func (fc *FileCrypto) EncryptFiles(files []struct{ Input, Output string }) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(files))

	for _, file := range files {
		wg.Add(1)
		go func(input, output string) {
			defer wg.Done()
			if err := fc.EncryptFile(input, output); err != nil {
				errCh <- fmt.Errorf("加密文件 %s 失败: %w", input, err)
			}
		}(file.Input, file.Output)
	}

	wg.Wait()
	close(errCh)

	// 收集所有错误
	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("批量加密过程中发生 %d 个错误: %v", len(errors), errors)
	}

	return nil
}

// DecryptFiles 批量解密文件
func (fc *FileCrypto) DecryptFiles(files []struct{ Input, Output string }) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(files))

	for _, file := range files {
		wg.Add(1)
		go func(input, output string) {
			defer wg.Done()
			if err := fc.DecryptFile(input, output); err != nil {
				errCh <- fmt.Errorf("解密文件 %s 失败: %w", input, err)
			}
		}(file.Input, file.Output)
	}

	wg.Wait()
	close(errCh)

	// 收集所有错误
	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("批量解密过程中发生 %d 个错误: %v", len(errors), errors)
	}

	return nil
}
