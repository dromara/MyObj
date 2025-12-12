package tests

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	hashpkg "myobj/src/pkg/hash"
	"myobj/src/pkg/models"
)

// TestQuickHash_SmallFile 测试小文件快速hash计算
func TestQuickHash_SmallFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "small_file.bin")

	// 创建10MB测试文件
	testData := make([]byte, 10*1024*1024)
	if _, err := rand.Read(testData); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatal(err)
	}

	// 计算快速hash（只计算前3个分片）
	startTime := time.Now()
	result, err := hashpkg.ComputeQuickHash(testFile, nil)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("计算快速hash失败: %v", err)
	}

	t.Logf("快速hash计算完成，耗时: %v", duration)
	t.Logf("文件大小: %d MB", result.FileSize/(1024*1024))
	t.Logf("分片签名: %s", result.ChunkSignature)
	t.Logf("分片数量: %d", len(result.ChunkHashes))

	// 验证结果
	if result.ChunkSignature == "" {
		t.Error("分片签名不应为空")
	}
	if len(result.ChunkHashes) == 0 {
		t.Error("应该至少有一个分片hash")
	}
}

// TestQuickHash_LargeFile 测试大文件快速hash（对比全量hash）
func TestQuickHash_LargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大文件测试")
	}

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "large_file.bin")

	// 创建500MB测试文件
	t.Log("创建500MB测试文件...")
	file, err := os.Create(testFile)
	if err != nil {
		t.Fatal(err)
	}

	chunk := make([]byte, 1024*1024) // 1MB
	for i := 0; i < 500; i++ {
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

	// 测试1: 只计算快速hash（前3个分片）
	t.Log("测试1: 计算快速hash（前3个分片）...")
	startQuick := time.Now()
	quickResult, err := hashpkg.ComputeQuickHash(testFile, nil)
	quickDuration := time.Since(startQuick)
	if err != nil {
		t.Fatalf("计算快速hash失败: %v", err)
	}

	// 测试2: 计算全量hash
	t.Log("测试2: 计算全量hash（完整文件）...")
	hasher := hashpkg.NewFastBlake3Hasher()
	startFull := time.Now()
	fullHash, _, err := hasher.ComputeFileHash(testFile)
	fullDuration := time.Since(startFull)
	if err != nil {
		t.Fatalf("计算全量hash失败: %v", err)
	}

	// 对比结果
	t.Logf("\n========== 性能对比 ==========")
	t.Logf("文件大小: 500MB")
	t.Logf("快速hash耗时: %v", quickDuration)
	t.Logf("全量hash耗时: %v", fullDuration)
	t.Logf("性能提升: %.2fx", float64(fullDuration)/float64(quickDuration))
	t.Logf("分片签名: %s", quickResult.ChunkSignature)
	t.Logf("全量hash: %s", fullHash)

	// 快速hash应该显著快于全量hash
	if quickDuration >= fullDuration {
		t.Error("快速hash应该比全量hash快")
	}
}

// TestQuickHash_WithFullHash 测试同时计算快速hash和全量hash
func TestQuickHash_WithFullHash(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.bin")

	// 创建20MB测试文件
	testData := make([]byte, 20*1024*1024)
	if _, err := rand.Read(testData); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatal(err)
	}

	// 配置：计算分片签名 + 全量hash
	config := &hashpkg.QuickHashConfig{
		SampleChunkCount: 3,
		ChunkSize:        4 * 1024 * 1024,
		ComputeFullHash:  true, // 同时计算全量hash
	}

	result, err := hashpkg.ComputeQuickHash(testFile, config)
	if err != nil {
		t.Fatalf("计算hash失败: %v", err)
	}

	// 验证结果
	if result.ChunkSignature == "" {
		t.Error("分片签名不应为空")
	}
	if result.FullHash == "" {
		t.Error("全量hash不应为空")
	}

	t.Logf("分片签名: %s", result.ChunkSignature)
	t.Logf("全量hash: %s", result.FullHash)
	t.Logf("分片数量: %d", len(result.ChunkHashes))
}

// TestChunkSignatureConsistency 测试分片签名一致性
func TestChunkSignatureConsistency(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "consistency.bin")

	// 创建测试文件
	testData := make([]byte, 15*1024*1024) // 15MB
	if _, err := rand.Read(testData); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatal(err)
	}

	// 多次计算应该得到相同的分片签名
	result1, err := hashpkg.ComputeQuickHash(testFile, nil)
	if err != nil {
		t.Fatal(err)
	}

	result2, err := hashpkg.ComputeQuickHash(testFile, nil)
	if err != nil {
		t.Fatal(err)
	}

	if result1.ChunkSignature != result2.ChunkSignature {
		t.Error("多次计算得到不同的分片签名")
	}

	t.Log("分片签名一致性测试通过")
}

// TestComputeChunkSignatureFromHashes 测试从hash列表计算签名
func TestComputeChunkSignatureFromHashes(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.bin")

	// 创建测试文件
	testData := make([]byte, 15*1024*1024)
	if _, err := rand.Read(testData); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatal(err)
	}

	// 方式1: 直接计算
	result, err := hashpkg.ComputeQuickHash(testFile, nil)
	if err != nil {
		t.Fatal(err)
	}

	// 方式2: 从hash列表重建签名
	signature, err := hashpkg.ComputeChunkSignatureFromHashes(result.ChunkHashes)
	if err != nil {
		t.Fatal(err)
	}

	// 两种方式应该得到相同的签名
	if result.ChunkSignature != signature {
		t.Error("从hash列表重建的签名与直接计算的签名不一致")
	}

	t.Log("签名重建测试通过")
}

// TestPrepareFileInfo 测试准备文件信息
func TestPrepareFileInfo(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.bin")

	// 创建测试文件
	testData := make([]byte, 15*1024*1024)
	if _, err := rand.Read(testData); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatal(err)
	}

	// 计算快速hash（包含全量hash）
	config := &hashpkg.QuickHashConfig{
		SampleChunkCount: 3,
		ChunkSize:        4 * 1024 * 1024,
		ComputeFullHash:  true,
	}
	quickResult, err := hashpkg.ComputeQuickHash(testFile, config)
	if err != nil {
		t.Fatal(err)
	}

	// 准备文件信息
	fileInfo := &models.FileInfo{
		ID:   "test-file-001",
		Name: "test.bin",
		Size: int(quickResult.FileSize),
	}

	// 验证
	if fileInfo.ChunkSignature == "" {
		t.Error("分片签名不应为空")
	}
	if fileInfo.FirstChunkHash == "" {
		t.Error("第一个分片hash不应为空")
	}
	if fileInfo.FileHash == "" {
		t.Error("全量hash不应为空")
	}
	if !fileInfo.HasFullHash {
		t.Error("HasFullHash应该为true")
	}

	t.Logf("文件信息准备完成:")
	t.Logf("  分片签名: %s", fileInfo.ChunkSignature)
	t.Logf("  全量hash: %s", fileInfo.FileHash)
	t.Logf("  第一分片: %s", fileInfo.FirstChunkHash)
}

// BenchmarkQuickHash 快速hash性能基准测试
func BenchmarkQuickHash(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "bench.bin")

	// 创建50MB测试文件
	file, _ := os.Create(testFile)
	chunk := make([]byte, 1024*1024)
	for i := 0; i < 50; i++ {
		rand.Read(chunk)
		file.Write(chunk)
	}
	file.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := hashpkg.ComputeQuickHash(testFile, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkQuickHash_vs_FullHash 快速hash vs 全量hash对比
func BenchmarkQuickHash_vs_FullHash(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "bench.bin")

	// 创建100MB测试文件
	file, _ := os.Create(testFile)
	chunk := make([]byte, 1024*1024)
	for i := 0; i < 100; i++ {
		rand.Read(chunk)
		file.Write(chunk)
	}
	file.Close()

	b.Run("QuickHash", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			hashpkg.ComputeQuickHash(testFile, nil)
		}
	})

	b.Run("FullHash", func(b *testing.B) {
		hasher := hashpkg.NewFastBlake3Hasher()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			hasher.ComputeFileHash(testFile)
		}
	})
}
