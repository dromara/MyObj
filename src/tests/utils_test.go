package tests

import (
	"fmt"
	"log"
	"myobj/src/config"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/preview"
	"myobj/src/pkg/util"
	"runtime"
	"testing"
	"time"
)

// 测试生成图片缩略图
func TestGenerateImageThumbnail(t *testing.T) {
	config.InitConfig()
	logger.InitLogger()
	err := preview.GenerateImageThumbnail("C:\\Users\\29120\\Pictures\\【哲风壁纸】剑客-水墨.png", "C:\\Users\\29120\\Pictures\\1.png", 200)
	if err != nil {
		panic(err)

	}
}

func TestDiskUtil(t *testing.T) {
	config.InitConfig()
	logger.InitLogger()
	// 示例使用

	// 1. 获取所有磁盘信息
	fmt.Println("=== 所有磁盘信息 ===")
	disks, err := util.GetDiskInfo()
	if err != nil {
		log.Fatalf("获取磁盘信息失败: %v", err)
	}

	for _, disk := range disks {
		fmt.Printf("挂载点: %s\n", disk.Mount)
		fmt.Printf("  总大小: %s\n", util.FormatBytes(disk.Total))
		fmt.Printf("  已使用: %s\n", util.FormatBytes(disk.Used))
		fmt.Printf("  剩余: %s\n", util.FormatBytes(disk.Free))
		fmt.Printf("  可用: %s\n", util.FormatBytes(disk.Avail))
		fmt.Println()
	}

	// 2. 获取指定挂载点信息
	fmt.Println("=== 指定挂载点信息 ===")
	var testMount string
	if runtime.GOOS == "windows" {
		testMount = "C:"
	} else {
		testMount = "/"
	}

	specificDisk, err := util.DiskInfoPath(testMount)
	if err != nil {
		log.Printf("获取指定挂载点信息失败: %v", err)
	} else {
		fmt.Printf("%s 的信息:\n", testMount)
		fmt.Printf("  总大小: %s\n", util.FormatBytes(specificDisk.Total))
		fmt.Printf("  已使用: %s\n", util.FormatBytes(specificDisk.Used))
		fmt.Printf("  剩余: %s\n", util.FormatBytes(specificDisk.Free))
	}

	// 3. 获取当前程序所在磁盘信息
	fmt.Println("\n=== 当前程序所在磁盘信息 ===")
	currentDisk, err := util.GetCurrentDirectoryDiskSpace()
	if err != nil {
		log.Printf("获取当前目录磁盘信息失败: %v", err)
	} else {
		fmt.Printf("当前程序所在磁盘 (%s):\n", currentDisk.Mount)
		fmt.Printf("  总大小: %s\n", util.FormatBytes(currentDisk.Total))
		fmt.Printf("  已使用: %s\n", util.FormatBytes(currentDisk.Used))
		fmt.Printf("  剩余: %s\n", util.FormatBytes(currentDisk.Free))
	}
}

func TestFileEncrypt(t *testing.T) {
	config.InitConfig()
	logger.InitLogger()
	tn := time.Now()
	err := util.NewFileCrypto("123456").EncryptFile("C:\\Users\\29120\\Pictures\\天空小姐姐 黑色唯美裙子 厚涂画风 4k动漫壁纸_彼岸图网.jpg", "C:\\Users\\29120\\Pictures\\1.jpg.enc")
	if err != nil {
		panic(err)
		return
	}
	fmt.Println(time.Since(tn))
	err = util.NewFileCrypto("123456").DecryptFile("C:\\Users\\29120\\Pictures\\1.jpg.enc", "C:\\Users\\29120\\Pictures\\天空1.jpg")
	if err != nil {
		panic(err)
		return
	}
	fmt.Println(time.Since(tn))
	memory, u, err := util.NewFileCrypto("123456").GetSystemMemory()
	if err != nil {
		panic(err)
		return
	}
	fmt.Println(memory, u)
}
