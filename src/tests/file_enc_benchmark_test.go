package tests

import (
	"crypto/rand"
	"myobj/src/pkg/util"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// BenchmarkEncryptSmallFile 小文件加密性能测试 (1MB)
func BenchmarkEncryptSmallFile(b *testing.B) {
	tempDir := b.TempDir()
	inputPath := filepath.Join(tempDir, "small.bin")

	// 创建1MB测试文件
	data := make([]byte, 1024*1024)
	rand.Read(data)
	os.WriteFile(inputPath, data, 0644)

	crypto := util.NewFileCrypto("benchmark-password")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(tempDir, "encrypted_small.bin")
		crypto.EncryptFile(inputPath, outputPath)
		os.Remove(outputPath) // 清理
	}
}

// BenchmarkEncryptMediumFile 中等文件加密性能测试 (50MB)
func BenchmarkEncryptMediumFile(b *testing.B) {
	tempDir := b.TempDir()
	inputPath := filepath.Join(tempDir, "medium.bin")

	// 创建50MB测试文件
	inputFile, _ := os.Create(inputPath)
	chunk := make([]byte, 1024*1024)
	for i := 0; i < 50; i++ {
		rand.Read(chunk)
		inputFile.Write(chunk)
	}
	inputFile.Close()

	crypto := util.NewFileCrypto("benchmark-password")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(tempDir, "encrypted_medium.bin")
		crypto.EncryptFile(inputPath, outputPath)
		os.Remove(outputPath)
	}
}

// BenchmarkEncryptLargeFile 大文件加密性能测试 (200MB, 流式处理)
func BenchmarkEncryptLargeFile(b *testing.B) {
	tempDir := b.TempDir()
	inputPath := filepath.Join(tempDir, "large.bin")

	// 创建200MB测试文件
	inputFile, _ := os.Create(inputPath)
	chunk := make([]byte, 1024*1024)
	for i := 0; i < 200; i++ {
		rand.Read(chunk)
		inputFile.Write(chunk)
	}
	inputFile.Close()

	crypto := util.NewFileCrypto("benchmark-password")

	// 记录初始内存
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(tempDir, "encrypted_large.bin")
		crypto.EncryptFile(inputPath, outputPath)
		os.Remove(outputPath)
	}

	// 记录最终内存
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	b.Logf("内存增长: %d MB", (m2.TotalAlloc-m1.TotalAlloc)/(1024*1024))
}

// BenchmarkDecryptLargeFile 大文件解密性能测试
func BenchmarkDecryptLargeFile(b *testing.B) {
	tempDir := b.TempDir()
	inputPath := filepath.Join(tempDir, "large.bin")
	encryptedPath := filepath.Join(tempDir, "encrypted.bin")

	// 创建200MB测试文件
	inputFile, _ := os.Create(inputPath)
	chunk := make([]byte, 1024*1024)
	for i := 0; i < 200; i++ {
		rand.Read(chunk)
		inputFile.Write(chunk)
	}
	inputFile.Close()

	// 先加密
	crypto := util.NewFileCrypto("benchmark-password")
	crypto.EncryptFile(inputPath, encryptedPath)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(tempDir, "decrypted.bin")
		crypto.DecryptFile(encryptedPath, outputPath)
		os.Remove(outputPath)
	}
}

// TestMemoryUsageWithLargeFile 测试大文件处理时的内存使用
func TestMemoryUsageWithLargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过内存测试")
	}

	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "huge.bin")
	encryptedPath := filepath.Join(tempDir, "encrypted.bin")
	decryptedPath := filepath.Join(tempDir, "decrypted.bin")

	// 创建500MB测试文件（超过SmallFileLimit）
	t.Log("创建500MB测试文件...")
	inputFile, err := os.Create(inputPath)
	if err != nil {
		t.Fatal(err)
	}

	chunk := make([]byte, 1024*1024) // 1MB
	for i := 0; i < 500; i++ {
		if _, err := rand.Read(chunk); err != nil {
			inputFile.Close()
			t.Fatal(err)
		}
		if _, err := inputFile.Write(chunk); err != nil {
			inputFile.Close()
			t.Fatal(err)
		}
	}
	inputFile.Close()

	crypto := util.NewFileCrypto("memory-test-password")

	// 强制GC
	runtime.GC()

	// 记录加密前的内存
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	t.Logf("加密前内存: Alloc=%dMB, Sys=%dMB", m1.Alloc/(1024*1024), m1.Sys/(1024*1024))

	// 加密
	t.Log("开始加密500MB文件...")
	if err := crypto.EncryptFile(inputPath, encryptedPath); err != nil {
		t.Fatal(err)
	}

	// 记录加密后的内存
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	t.Logf("加密后内存: Alloc=%dMB, Sys=%dMB", m2.Alloc/(1024*1024), m2.Sys/(1024*1024))

	// 解密
	t.Log("开始解密...")
	if err := crypto.DecryptFile(encryptedPath, decryptedPath); err != nil {
		t.Fatal(err)
	}

	// 记录解密后的内存
	runtime.GC()
	var m3 runtime.MemStats
	runtime.ReadMemStats(&m3)
	t.Logf("解密后内存: Alloc=%dMB, Sys=%dMB", m3.Alloc/(1024*1024), m3.Sys/(1024*1024))

	// 验证内存增长是否合理（不应该超过200MB，因为使用了流式处理）
	memIncrease := (m3.Alloc - m1.Alloc) / (1024 * 1024)
	t.Logf("总内存增长: %dMB", memIncrease)

	if memIncrease > 200 {
		t.Errorf("内存增长过大: %dMB，应该小于200MB", memIncrease)
	}

	t.Log("大文件内存测试通过")
}
