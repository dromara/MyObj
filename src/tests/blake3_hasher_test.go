package tests

import (
	"crypto/rand"
	hashpkg "myobj/src/pkg/hash"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestComputeString 测试字符串哈希
func TestComputeString(t *testing.T) {
	testCases := []string{
		"",
		"hello",
		"Hello, World!",
		"测试中文",
	}

	for _, input := range testCases {
		hash := hashpkg.ComputeString(input)

		// 验证哈希长度
		if len(hash) != 64 {
			t.Errorf("字符串 '%s' 的哈希长度不正确，期望: 64, 实际: %d", input, len(hash))
		}

		// 验证一致性
		hash2 := hashpkg.ComputeString(input)
		if hash != hash2 {
			t.Errorf("相同字符串产生了不同的哈希")
		}

		t.Logf("字符串 '%s' => %s", input, hash[:16]+"...")
	}

	t.Log("字符串哈希测试通过")
}

// TestComputeBytes 测试字节数组哈希
func TestComputeBytes(t *testing.T) {
	data := []byte("test data for blake3")
	hashValue := hashpkg.ComputeBytes(data)

	if len(hashValue) != 64 { // Blake3产生256位哈希，hex编码后64字符
		t.Errorf("哈希长度不正确，期望: 64, 实际: %d", len(hashValue))
	}

	// 同样的数据应该产生同样的哈希
	hash2 := hashpkg.ComputeBytes(data)
	if hashValue != hash2 {
		t.Error("相同数据产生了不同的哈希")
	}

	t.Logf("哈希值: %s", hashValue)
}

// TestComputeFileHash_SmallFile 测试小文件哈希计算
func TestComputeFileHash_SmallFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "small_test.txt")

	// 创建1MB测试文件
	testData := make([]byte, 1024*1024)
	if _, err := rand.Read(testData); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatal(err)
	}

	hasher := hashpkg.NewFastBlake3Hasher()
	hashValue, duration, err := hasher.ComputeFileHash(testFile)
	if err != nil {
		t.Fatalf("计算文件哈希失败: %v", err)
	}

	if len(hashValue) != 64 {
		t.Errorf("哈希长度不正确: %d", len(hashValue))
	}

	t.Logf("小文件哈希计算成功，耗时: %v, 哈希: %s", duration, hashValue[:16]+"...")
}

// TestComputeFileHash_LargeFile 测试大文件哈希计算（流式处理）
func TestComputeFileHash_LargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大文件测试")
	}

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "large_test.bin")

	// 创建600MB测试文件（触发流式处理）
	t.Log("创建600MB测试文件...")
	file, err := os.Create(testFile)
	if err != nil {
		t.Fatal(err)
	}

	chunk := make([]byte, 1024*1024) // 1MB
	for i := 0; i < 600; i++ {
		if _, err := rand.Read(chunk); err != nil {
			file.Close()
			t.Fatal(err)
		}
		if _, err := file.Write(chunk); err != nil {
			file.Close()
			t.Fatal(err)
		}
	}
	file.Close()

	// 记录初始内存
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// 计算哈希
	hasher := hashpkg.NewFastBlake3Hasher().WithVerbose(true)
	hashValue, duration, err := hasher.ComputeFileHash(testFile)
	if err != nil {
		t.Fatalf("计算大文件哈希失败: %v", err)
	}

	// 记录最终内存
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	memIncrease := (m2.Alloc - m1.Alloc) / (1024 * 1024)

	t.Logf("大文件哈希计算成功")
	t.Logf("文件大小: 600MB")
	t.Logf("计算耗时: %v", duration)
	t.Logf("哈希值: %s", hashValue[:16]+"...")
	t.Logf("内存增长: %dMB", memIncrease)

	// 验证内存占用合理（应该远小于文件大小）
	if memIncrease > 100 {
		t.Logf("警告: 内存增长较大 %dMB", memIncrease)
	}
}

// TestHashConsistency 测试哈希一致性
func TestHashConsistency(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "consistency_test.txt")

	// 创建测试文件
	testData := []byte("consistency test data for blake3 hashing")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatal(err)
	}

	hasher := hashpkg.NewFastBlake3Hasher()

	// 多次计算应该得到相同结果
	hash1, _, err := hasher.ComputeFileHash(testFile)
	if err != nil {
		t.Fatal(err)
	}

	hash2, _, err := hasher.ComputeFileHash(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if hash1 != hash2 {
		t.Error("多次计算得到不同的哈希值")
	}

	// 使用字节方式计算应该得到相同结果
	expectedHash := hashpkg.ComputeBytes(testData)
	if hash1 != expectedHash {
		t.Errorf("文件哈希与字节哈希不一致\n文件: %s\n字节: %s", hash1, expectedHash)
	}

	t.Log("哈希一致性测试通过")
}

// TestVerifyFileHash 测试哈希验证
func TestVerifyFileHash(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "verify_test.txt")

	testData := []byte("test data for verification")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatal(err)
	}

	hasher := hashpkg.NewFastBlake3Hasher()

	// 计算正确的哈希
	correctHash, _, err := hasher.ComputeFileHash(testFile)
	if err != nil {
		t.Fatal(err)
	}

	// 验证正确的哈希
	valid, err := hasher.VerifyFileHash(testFile, correctHash)
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("正确的哈希验证失败")
	}

	// 验证错误的哈希
	wrongHash := "0000000000000000000000000000000000000000000000000000000000000000"
	valid, err = hasher.VerifyFileHash(testFile, wrongHash)
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Error("错误的哈希不应该验证通过")
	}

	t.Log("哈希验证测试通过")
}

// TestComputeMultipleFiles 测试批量计算文件哈希
func TestComputeMultipleFiles(t *testing.T) {
	tempDir := t.TempDir()

	// 创建多个测试文件
	fileCount := 5
	filePaths := make([]string, fileCount)
	for i := 0; i < fileCount; i++ {
		filePath := filepath.Join(tempDir, "test_"+string(rune('0'+i))+".txt")
		testData := make([]byte, 100*1024) // 100KB
		if _, err := rand.Read(testData); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filePath, testData, 0644); err != nil {
			t.Fatal(err)
		}
		filePaths[i] = filePath
	}

	hasher := hashpkg.NewFastBlake3Hasher()
	results := hasher.ComputeMultipleFiles(filePaths)

	if len(results) != fileCount {
		t.Errorf("结果数量不匹配，期望: %d, 实际: %d", fileCount, len(results))
	}

	// 验证所有文件都成功计算
	for i, result := range results {
		if result.Error != nil {
			t.Errorf("文件 %d 计算失败: %v", i, result.Error)
		}
		if len(result.FileHash) != 64 {
			t.Errorf("文件 %d 哈希长度不正确: %d", i, len(result.FileHash))
		}
		t.Logf("文件 %d: 耗时=%v, 哈希=%s...", i, result.Duration, result.FileHash[:16])
	}

	t.Log("批量文件哈希计算测试通过")
}

// TestComputeMultipleFilesConcurrent 测试并发批量计算
func TestComputeMultipleFilesConcurrent(t *testing.T) {
	tempDir := t.TempDir()

	// 创建多个测试文件
	fileCount := 10
	filePaths := make([]string, fileCount)
	for i := 0; i < fileCount; i++ {
		filePath := filepath.Join(tempDir, "concurrent_test_"+string(rune('0'+i))+".bin")
		testData := make([]byte, 500*1024) // 500KB
		if _, err := rand.Read(testData); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filePath, testData, 0644); err != nil {
			t.Fatal(err)
		}
		filePaths[i] = filePath
	}

	hasher := hashpkg.NewFastBlake3Hasher()

	// 并发计算
	results := hasher.ComputeMultipleFilesConcurrent(filePaths, 4)

	if len(results) != fileCount {
		t.Errorf("结果数量不匹配，期望: %d, 实际: %d", fileCount, len(results))
	}

	// 验证所有文件都成功计算
	successCount := 0
	for i, result := range results {
		if result.Error == nil {
			successCount++
			if len(result.FileHash) != 64 {
				t.Errorf("文件 %d 哈希长度不正确: %d", i, len(result.FileHash))
			}
		} else {
			t.Errorf("文件 %d 计算失败: %v", i, result.Error)
		}
	}

	t.Logf("并发批量计算测试通过，成功: %d/%d", successCount, fileCount)
}

// TestMemoryMapMode 测试内存映射模式
func TestMemoryMapMode(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过内存映射测试")
	}

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "mmap_test.bin")

	// 创建600MB文件（在阈值范围内）
	t.Log("创建测试文件...")
	file, err := os.Create(testFile)
	if err != nil {
		t.Fatal(err)
	}

	chunk := make([]byte, 1024*1024)
	for i := 0; i < 600; i++ {
		if _, err := rand.Read(chunk); err != nil {
			file.Close()
			t.Fatal(err)
		}
		if _, err := file.Write(chunk); err != nil {
			file.Close()
			t.Fatal(err)
		}
	}
	file.Close()

	// 使用内存映射模式
	hasher1 := hashpkg.NewFastBlake3Hasher().WithMemoryMap(true)
	hash1, duration1, err := hasher1.ComputeFileHash(testFile)
	if err != nil {
		t.Fatalf("内存映射模式计算失败: %v", err)
	}

	// 使用流式模式
	hasher2 := hashpkg.NewFastBlake3Hasher().WithMemoryMap(false)
	hash2, duration2, err := hasher2.ComputeFileHash(testFile)
	if err != nil {
		t.Fatalf("流式模式计算失败: %v", err)
	}

	// 两种模式应该产生相同的哈希
	if hash1 != hash2 {
		t.Error("不同模式产生了不同的哈希值")
	}

	t.Logf("内存映射模式耗时: %v", duration1)
	t.Logf("流式模式耗时: %v", duration2)
	t.Logf("哈希值: %s", hash1[:16]+"...")
}

// BenchmarkBlake3SmallFile 小文件性能基准测试
func BenchmarkBlake3SmallFile(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "bench_small.bin")

	// 创建1MB测试文件
	testData := make([]byte, 1024*1024)
	rand.Read(testData)
	os.WriteFile(testFile, testData, 0644)

	hasher := hashpkg.NewFastBlake3Hasher()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		hasher.ComputeFileHash(testFile)
	}
}

// BenchmarkBlake3MediumFile 中等文件性能基准测试
func BenchmarkBlake3MediumFile(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "bench_medium.bin")

	// 创建50MB测试文件
	file, _ := os.Create(testFile)
	chunk := make([]byte, 1024*1024)
	for i := 0; i < 50; i++ {
		rand.Read(chunk)
		file.Write(chunk)
	}
	file.Close()

	hasher := hashpkg.NewFastBlake3Hasher()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		hasher.ComputeFileHash(testFile)
	}
}
