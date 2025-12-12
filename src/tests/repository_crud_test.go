package tests

import (
	"context"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/models"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(
		&models.UserInfo{},
		&models.FileInfo{},
		&models.Group{},
		&models.Share{},
		&models.Disk{},
		&models.ApiKey{},
		&models.FileChunk{},
		&models.Power{},
		&models.GroupPower{},
		&models.UserFiles{},
		&models.VirtualPath{},
	)
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return db
}

func TestUserRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := impl.NewUserRepository(db)
	ctx := context.Background()

	user := &models.UserInfo{
		ID:        "user001",
		Name:      "testuser",
		UserName:  "testuser",
		Password:  "password123",
		Email:     "test@example.com",
		Phone:     "13800138000",
		GroupID:   1,
		CreatedAt: custom_type.JsonTime(time.Now()),
		Space:     10737418240,
		FreeSpace: 10737418240,
	}

	// Create
	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Read
	foundUser, err := repo.GetByID(ctx, "user001")
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if foundUser.UserName != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", foundUser.UserName)
	}

	// Update
	foundUser.Name = "updated"
	err = repo.Update(ctx, foundUser)
	if err != nil {
		t.Fatalf("failed to update user: %v", err)
	}

	// Delete
	err = repo.Delete(ctx, "user001")
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}
}

func TestFileInfoRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := impl.NewFileInfoRepository(db)
	ctx := context.Background()

	file := &models.FileInfo{
		ID:          "file001",
		Name:        "test.txt",
		RandomName:  "random123",
		Size:        1024,
		Mime:        "text/plain",
		VirtualPath: "/test.txt",
		Path:        "/data/random123",
		FileHash:    "hash123",
		IsChunk:     false,
		CreatedAt:   custom_type.JsonTime(time.Now()),
		UpdatedAt:   custom_type.JsonTime(time.Now()),
	}

	err := repo.Create(ctx, file)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	foundFile, err := repo.GetByHash(ctx, "hash123")
	if err != nil {
		t.Fatalf("failed to get file by hash: %v", err)
	}
	if foundFile.Name != "test.txt" {
		t.Errorf("expected name 'test.txt', got '%s'", foundFile.Name)
	}
}

func TestGroupRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := impl.NewGroupRepository(db)
	ctx := context.Background()

	group := &models.Group{
		ID:        1,
		Name:      "admin",
		CreatedAt: custom_type.JsonTime(time.Now()),
	}

	err := repo.Create(ctx, group)
	if err != nil {
		t.Fatalf("failed to create group: %v", err)
	}

	foundGroup, err := repo.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get group: %v", err)
	}
	if foundGroup.Name != "admin" {
		t.Errorf("expected name 'admin', got '%s'", foundGroup.Name)
	}
}

func TestShareRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := impl.NewShareRepository(db)
	ctx := context.Background()

	share := &models.Share{
		UserID:        "user001",
		FileID:        "file001",
		Token:         "token123",
		ExpiresAt:     custom_type.JsonTime(time.Now().Add(24 * time.Hour)),
		PasswordHash:  "hash",
		DownloadCount: 0,
		CreatedAt:     custom_type.JsonTime(time.Now()),
	}

	err := repo.Create(ctx, share)
	if err != nil {
		t.Fatalf("failed to create share: %v", err)
	}

	foundShare, err := repo.GetByToken(ctx, "token123")
	if err != nil {
		t.Fatalf("failed to get share by token: %v", err)
	}
	if foundShare.FileID != "file001" {
		t.Errorf("expected file_id 'file001', got '%s'", foundShare.FileID)
	}
}

func TestFileChunkRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := impl.NewFileChunkRepository(db)
	ctx := context.Background()

	chunks := []*models.FileChunk{
		{
			ID:         "chunk001",
			FileID:     "file001",
			ChunkPath:  "/data/chunks/001",
			ChunkSize:  1048576,
			ChunkHash:  "hash001",
			ChunkIndex: 0,
		},
		{
			ID:         "chunk002",
			FileID:     "file001",
			ChunkPath:  "/data/chunks/002",
			ChunkSize:  1048576,
			ChunkHash:  "hash002",
			ChunkIndex: 1,
		},
	}

	err := repo.BatchCreate(ctx, chunks)
	if err != nil {
		t.Fatalf("failed to batch create chunks: %v", err)
	}

	foundChunks, err := repo.GetByFileID(ctx, "file001")
	if err != nil {
		t.Fatalf("failed to get chunks: %v", err)
	}
	if len(foundChunks) != 2 {
		t.Errorf("expected 2 chunks, got %d", len(foundChunks))
	}
}
