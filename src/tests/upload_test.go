package tests

import (
	"myobj/src/config"
	"myobj/src/internal/repository/database"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/upload"
	"testing"
	"time"
)

func TestUpload(t *testing.T) {
	err := config.InitConfig()
	if err != nil {
		panic(err)
		return
	}
	logger.InitLogger()
	database.InitDataBase()
	data := upload.FileUploadData{
		TempFilePath:    "E:\\codes\\myobj\\examples\\proxmox-ve_9.0-1.iso",
		FileName:        "proxmox-ve_9.0-1.iso",
		FileSize:        1641615360,
		ChunkSignature:  "63516131",
		FirstChunkHash:  "sd45f1wc16w",
		SecondChunkHash: "c61q6ec16q1d",
		ThirdChunkHash:  "c16qwe84d16a1d",
		IsEnc:           true,
		IsChunk:         false,
		ChunkCount:      0,
		VirtualPath:     "/home/我的文件",
		UserID:          "cloaihdnlaishcweoc",
	}
	tn := time.Now()
	factory := impl.NewRepositoryFactory(database.GetDB())
	file, err := upload.ProcessUploadedFile(&data, factory)
	if err != nil {
		panic(err)
		return
	}
	t.Log(time.Now().Sub(tn))
	t.Log(file)
}

func TestDownload(t *testing.T) {
	err := config.InitConfig()
	if err != nil {
		panic(err)
		return
	}
	logger.InitLogger()
	database.InitDataBase()
	//temp := "E:\\obj_data\\temp"
	//factory := impl.NewRepositoryFactory(database.GetDB())
	//forDownload, m, err := download.PrepareFileForDownload("019a7dd3-f44c-71ae-b2eb-bf1e1014d3cb", temp, factory)
	//if err != nil {
	//	panic(err)
	//	return
	//}
	//t.Log("可下载文件路径：", forDownload)
	//t.Log("可下载文件信息：", m)
}
