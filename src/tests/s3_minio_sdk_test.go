package tests

import (
	"bytes"
	"context"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// TestS3MinIOSDKCompatibility 测试MinIO SDK兼容性
func TestS3MinIOSDKCompatibility(t *testing.T) {
	// 初始化MinIO客户端
	// 注意：需要先启动服务器并创建API Key
	client, err := minio.New("localhost:8080", &minio.Options{
		Creds:  credentials.NewStaticV4("your-access-key-id", "your-secret-key", ""),
		Secure: false,
		Region: "us-east-1",
	})
	if err != nil {
		t.Fatalf("Failed to create MinIO client: %v", err)
	}

	ctx := context.Background()

	// 测试1: ListBuckets
	t.Run("ListBuckets", func(t *testing.T) {
		buckets, err := client.ListBuckets(ctx)
		if err != nil {
			t.Fatalf("ListBuckets failed: %v", err)
		}
		t.Logf("Found %d buckets", len(buckets))
		for _, b := range buckets {
			t.Logf("Bucket: %s, Created: %v", b.Name, b.CreationDate)
		}
	})

	// 测试2: CreateBucket
	bucketName := "test-bucket"
	t.Run("CreateBucket", func(t *testing.T) {
		err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
			Region: "us-east-1",
		})
		if err != nil {
			// 检查bucket是否已存在
			exists, errBucketExists := client.BucketExists(ctx, bucketName)
			if errBucketExists == nil && exists {
				t.Log("Bucket already exists, skipping creation")
			} else {
				t.Fatalf("MakeBucket failed: %v", err)
			}
		} else {
			t.Log("Bucket created successfully")
		}
	})

	// 测试3: BucketExists
	t.Run("BucketExists", func(t *testing.T) {
		exists, err := client.BucketExists(ctx, bucketName)
		if err != nil {
			t.Fatalf("BucketExists failed: %v", err)
		}
		if !exists {
			t.Fatal("Bucket should exist but doesn't")
		}
		t.Log("Bucket exists confirmed")
	})

	// 测试4: PutObject
	objectName := "test-file.txt"
	t.Run("PutObject", func(t *testing.T) {
		content := []byte("Hello, MyObj S3!")
		_, err := client.PutObject(ctx, bucketName, objectName,
			bytes.NewReader(content), int64(len(content)),
			minio.PutObjectOptions{
				ContentType: "text/plain",
				UserMetadata: map[string]string{
					"x-amz-meta-author": "test-user",
				},
			})
		if err != nil {
			t.Fatalf("PutObject failed: %v", err)
		}
		t.Log("Object uploaded successfully")
	})

	// 测试5: StatObject
	t.Run("StatObject", func(t *testing.T) {
		info, err := client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
		if err != nil {
			t.Fatalf("StatObject failed: %v", err)
		}
		t.Logf("Object info: Size=%d, ContentType=%s, ETag=%s",
			info.Size, info.ContentType, info.ETag)
	})

	// 测试6: GetObject
	t.Run("GetObject", func(t *testing.T) {
		obj, err := client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
		if err != nil {
			t.Fatalf("GetObject failed: %v", err)
		}
		defer obj.Close()

		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(obj)
		if err != nil {
			t.Fatalf("Read object failed: %v", err)
		}

		content := buf.String()
		expected := "Hello, MyObj S3!"
		if content != expected {
			t.Fatalf("Content mismatch: got %s, expected %s", content, expected)
		}
		t.Log("Object downloaded and verified successfully")
	})

	// 测试7: ListObjects
	t.Run("ListObjects", func(t *testing.T) {
		objectCh := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
			Prefix:    "",
			Recursive: true,
		})

		count := 0
		for object := range objectCh {
			if object.Err != nil {
				t.Fatalf("ListObjects failed: %v", object.Err)
			}
			t.Logf("Object: %s, Size: %d, LastModified: %v",
				object.Key, object.Size, object.LastModified)
			count++
		}
		t.Logf("Found %d objects", count)
	})

	// 测试8: RemoveObject
	t.Run("RemoveObject", func(t *testing.T) {
		err := client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
		if err != nil {
			t.Fatalf("RemoveObject failed: %v", err)
		}
		t.Log("Object removed successfully")
	})

	// 测试9: RemoveBucket
	t.Run("RemoveBucket", func(t *testing.T) {
		err := client.RemoveBucket(ctx, bucketName)
		if err != nil {
			t.Fatalf("RemoveBucket failed: %v", err)
		}
		t.Log("Bucket removed successfully")
	})
}
