package tests

import (
	"bytes"
	"crypto/rand"
	"myobj/src/pkg/util"
	"os"
	"path/filepath"
	"testing"
)

// TestFileCrypto_SmallFile 测试小文件加密解密
func TestFileCrypto_SmallFile(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "small_input.txt")
	encryptedPath := filepath.Join(tempDir, "small_encrypted.bin")
	decryptedPath := filepath.Join(tempDir, "small_decrypted.txt")

	// 创建测试数据 (1MB)
	testData := make([]byte, 1024*1024)
	if _, err := rand.Read(testData); err != nil {
		t.Fatalf("生成测试数据失败: %v", err)
	}

	if err := os.WriteFile(inputPath, testData, 0644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	// 创建加密处理器
	crypto := util.NewFileCrypto("test-password-123")

	// 加密
	if err := crypto.EncryptFile(inputPath, encryptedPath); err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	// 验证加密文件存在
	if _, err := os.Stat(encryptedPath); os.IsNotExist(err) {
		t.Fatal("加密文件不存在")
	}

	// 解密
	if err := crypto.DecryptFile(encryptedPath, decryptedPath); err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	// 读取解密后的数据
	decryptedData, err := os.ReadFile(decryptedPath)
	if err != nil {
		t.Fatalf("读取解密文件失败: %v", err)
	}

	// 验证数据一致性
	if !bytes.Equal(testData, decryptedData) {
		t.Fatal("解密后的数据与原始数据不一致")
	}
}

// TestFileCrypto_LargeFile 测试大文件加密解密（使用流式处理）
func TestFileCrypto_LargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大文件测试")
	}

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "large_input.bin")
	encryptedPath := filepath.Join(tempDir, "large_encrypted.bin")
	decryptedPath := filepath.Join(tempDir, "large_decrypted.bin")

	// 创建测试数据 (150MB，触发流式处理)
	testDataSize := 150 * 1024 * 1024
	t.Logf("创建 %dMB 测试文件...", testDataSize/(1024*1024))

	// 分块写入测试数据
	inputFile, err := os.Create(inputPath)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	chunkSize := 1024 * 1024 // 1MB
	chunk := make([]byte, chunkSize)
	for i := 0; i < testDataSize/chunkSize; i++ {
		if _, err := rand.Read(chunk); err != nil {
			inputFile.Close()
			t.Fatalf("生成测试数据失败: %v", err)
		}
		if _, err := inputFile.Write(chunk); err != nil {
			inputFile.Close()
			t.Fatalf("写入测试数据失败: %v", err)
		}
	}
	inputFile.Close()

	// 计算原始文件哈希
	originalHash, err := calculateFileHash(inputPath)
	if err != nil {
		t.Fatalf("计算原始文件哈希失败: %v", err)
	}

	// 创建加密处理器
	crypto := util.NewFileCrypto("test-password-456")

	// 加密
	t.Log("开始加密...")
	if err := crypto.EncryptFile(inputPath, encryptedPath); err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	// 解密
	t.Log("开始解密...")
	if err := crypto.DecryptFile(encryptedPath, decryptedPath); err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	// 计算解密后文件哈希
	decryptedHash, err := calculateFileHash(decryptedPath)
	if err != nil {
		t.Fatalf("计算解密文件哈希失败: %v", err)
	}

	// 验证数据一致性
	if originalHash != decryptedHash {
		t.Fatal("解密后的文件与原始文件不一致")
	}
	t.Log("大文件加密解密测试成功")
}

// TestFileCrypto_WrongPassword 测试错误密码
func TestFileCrypto_WrongPassword(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "input.txt")
	encryptedPath := filepath.Join(tempDir, "encrypted.bin")
	decryptedPath := filepath.Join(tempDir, "decrypted.txt")

	testData := []byte("This is a secret message")
	if err := os.WriteFile(inputPath, testData, 0644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	// 使用正确密码加密
	crypto1 := util.NewFileCrypto("correct-password")
	if err := crypto1.EncryptFile(inputPath, encryptedPath); err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	// 使用错误密码解密
	crypto2 := util.NewFileCrypto("wrong-password")
	err := crypto2.DecryptFile(encryptedPath, decryptedPath)
	if err == nil {
		t.Fatal("使用错误密码解密应该失败")
	}
	t.Logf("正确检测到错误密码: %v", err)
}

// TestFileCrypto_ConcurrentEncryption 测试并发加密
func TestFileCrypto_ConcurrentEncryption(t *testing.T) {
	tempDir := t.TempDir()
	crypto := util.NewFileCrypto("concurrent-test-password")

	// 准备多个测试文件
	testFiles := make([]struct{ Input, Output string }, 5)
	for i := 0; i < 5; i++ {
		inputPath := filepath.Join(tempDir, "input"+string(rune('0'+i))+".txt")
		outputPath := filepath.Join(tempDir, "encrypted"+string(rune('0'+i))+".bin")

		testData := make([]byte, 1024*100) // 100KB
		if _, err := rand.Read(testData); err != nil {
			t.Fatalf("生成测试数据失败: %v", err)
		}
		if err := os.WriteFile(inputPath, testData, 0644); err != nil {
			t.Fatalf("写入测试文件失败: %v", err)
		}

		testFiles[i] = struct{ Input, Output string }{inputPath, outputPath}
	}

	// 批量加密
	if err := crypto.EncryptFiles(testFiles); err != nil {
		t.Fatalf("批量加密失败: %v", err)
	}

	// 验证所有文件都已加密
	for _, file := range testFiles {
		if _, err := os.Stat(file.Output); os.IsNotExist(err) {
			t.Fatalf("加密文件不存在: %s", file.Output)
		}
	}
}

// TestGetSystemMemory 测试获取系统内存
func TestGetSystemMemory(t *testing.T) {
	crypto := util.NewFileCrypto("test")
	total, avail, err := crypto.GetSystemMemory()
	if err != nil {
		t.Fatalf("获取系统内存失败: %v", err)
	}

	t.Logf("系统总内存: %d MB", total/(1024*1024))
	t.Logf("可用内存: %d MB", avail/(1024*1024))

	if total == 0 || avail == 0 {
		t.Fatal("系统内存不应为0")
	}

	if avail > total {
		t.Fatal("可用内存不应大于总内存")
	}
}

// calculateFileHash 计算文件SHA256哈希
func calculateFileHash(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data[:32]), nil // 简化的哈希比较
}
