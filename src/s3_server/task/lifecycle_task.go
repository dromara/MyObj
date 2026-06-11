package task

import (
	"context"
	"encoding/json"
	"fmt"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/s3_server/service"
	"myobj/src/s3_server/types"
	"time"
)

// lifecyclePageSize 生命周期任务分页大小
const lifecyclePageSize = 1000

// LifecycleTask 生命周期管理定时任务
type LifecycleTask struct {
	factory       *impl.RepositoryFactory
	objectService *service.S3ObjectService
}

// NewLifecycleTask 创建生命周期管理定时任务
func NewLifecycleTask(factory *impl.RepositoryFactory, objectService *service.S3ObjectService) *LifecycleTask {
	return &LifecycleTask{
		factory:       factory,
		objectService: objectService,
	}
}

// ExecuteLifecycleRules 执行所有Bucket的生命周期规则
func (t *LifecycleTask) ExecuteLifecycleRules() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	logger.LOG.Info("开始执行生命周期管理任务")

	// 1. 获取所有Lifecycle配置
	lifecycles, err := t.factory.S3BucketLifecycle().ListAll(ctx)
	if err != nil {
		logger.LOG.Error("获取生命周期配置失败", "error", err)
		return fmt.Errorf("获取生命周期配置失败: %w", err)
	}

	if len(lifecycles) == 0 {
		logger.LOG.Debug("没有配置生命周期规则")
		return nil
	}

	logger.LOG.Info("找到生命周期配置", "count", len(lifecycles))

	// 2. 逐个处理每个Bucket的生命周期规则
	totalProcessed := 0
	totalDeleted := 0
	totalTransitions := 0

	for _, lifecycle := range lifecycles {
		processed, deleted, transitions, err := t.processBucketLifecycle(ctx, lifecycle)
		if err != nil {
			logger.LOG.Error("处理Bucket生命周期规则失败",
				"bucket", lifecycle.BucketName,
				"user_id", lifecycle.UserID,
				"error", err)
			continue
		}

		totalProcessed += processed
		totalDeleted += deleted
		totalTransitions += transitions
	}

	logger.LOG.Info("生命周期管理任务完成",
		"buckets_processed", len(lifecycles),
		"objects_processed", totalProcessed,
		"objects_deleted", totalDeleted,
		"objects_transitioned", totalTransitions)

	return nil
}

// processBucketLifecycle 处理单个Bucket的生命周期规则
func (t *LifecycleTask) processBucketLifecycle(ctx context.Context, lifecycle *models.S3BucketLifecycle) (processed, deleted, transitions int, err error) {
	// 1. 解析Lifecycle配置
	var config types.LifecycleConfiguration
	if err := json.Unmarshal([]byte(lifecycle.LifecycleJSON), &config); err != nil {
		return 0, 0, 0, fmt.Errorf("解析生命周期配置失败: %w", err)
	}

	// 2. 处理每个规则
	for _, rule := range config.Rules {
		// 只处理Enabled的规则
		if rule.Status != "Enabled" {
			continue
		}

		// 处理过期删除
		if rule.Expiration != nil {
			count, err := t.processExpiration(ctx, lifecycle.BucketName, lifecycle.UserID, rule)
			if err != nil {
				logger.LOG.Error("处理过期删除失败",
					"bucket", lifecycle.BucketName,
					"rule_id", rule.ID,
					"error", err)
				continue
			}
			deleted += count
			processed += count
		}

		// 处理非当前版本过期
		if rule.NoncurrentVersionExpiration != nil {
			count, err := t.processNoncurrentVersionExpiration(ctx, lifecycle.BucketName, lifecycle.UserID, rule)
			if err != nil {
				logger.LOG.Error("处理非当前版本过期失败",
					"bucket", lifecycle.BucketName,
					"rule_id", rule.ID,
					"error", err)
				continue
			}
			deleted += count
			processed += count
		}

		// 处理存储类别转换
		for _, transition := range rule.Transitions {
			count, err := t.processTransition(ctx, lifecycle.BucketName, lifecycle.UserID, rule, transition)
			if err != nil {
				logger.LOG.Error("处理存储类别转换失败",
					"bucket", lifecycle.BucketName,
					"rule_id", rule.ID,
					"error", err)
				continue
			}
			transitions += count
			processed += count
		}

		// 处理非当前版本转换
		for _, transition := range rule.NoncurrentVersionTransitions {
			count, err := t.processNoncurrentVersionTransition(ctx, lifecycle.BucketName, lifecycle.UserID, rule, transition)
			if err != nil {
				logger.LOG.Error("处理非当前版本转换失败",
					"bucket", lifecycle.BucketName,
					"rule_id", rule.ID,
					"error", err)
				continue
			}
			transitions += count
			processed += count
		}

		// 处理未完成的分片上传
		if rule.AbortIncompleteMultipartUpload != nil {
			count, err := t.processAbortIncompleteMultipartUpload(ctx, lifecycle.BucketName, lifecycle.UserID, rule)
			if err != nil {
				logger.LOG.Error("处理未完成的分片上传失败",
					"bucket", lifecycle.BucketName,
					"rule_id", rule.ID,
					"error", err)
				continue
			}
			processed += count
		}
	}

	return processed, deleted, transitions, nil
}

// processExpiration 处理过期删除
func (t *LifecycleTask) processExpiration(ctx context.Context, bucketName, userID string, rule types.LifecycleRule) (int, error) {
	// 计算过期时间
	var expirationTime time.Time
	if rule.Expiration.Days > 0 {
		expirationTime = time.Now().AddDate(0, 0, -rule.Expiration.Days)
	} else if rule.Expiration.Date != "" {
		var err error
		expirationTime, err = time.Parse("2006-01-02T15:04:05Z", rule.Expiration.Date)
		if err != nil {
			return 0, fmt.Errorf("解析过期日期失败: %w", err)
		}
	} else {
		return 0, fmt.Errorf("Expiration必须指定Days或Date")
	}

	// 获取Bucket下的所有对象（根据Prefix或Filter过滤）
	prefix := rule.Prefix
	if rule.Filter != nil && rule.Filter.Prefix != "" {
		prefix = rule.Filter.Prefix
	}

	// 分页列出对象（只获取最新版本，用于过期删除）
	deletedCount := 0
	marker := ""

	for {
		objects, err := t.factory.S3ObjectMetadata().ListByBucket(ctx, bucketName, userID, prefix, lifecyclePageSize, marker)
		if err != nil {
			return deletedCount, fmt.Errorf("列出对象失败: %w", err)
		}

		for _, obj := range objects {
			// 检查对象是否过期
			if time.Time(obj.CreatedAt).Before(expirationTime) {
				// 检查标签过滤（如果规则有Filter且包含Tag）
				if rule.Filter != nil && rule.Filter.Tag != nil {
					if !t.matchTag(obj, rule.Filter.Tag) {
						continue
					}
				}

				// 删除对象
				err := t.objectService.DeleteObject(ctx, bucketName, obj.ObjectKey, userID, "")
				if err != nil {
					logger.LOG.Error("删除过期对象失败",
						"bucket", bucketName,
						"key", obj.ObjectKey,
						"error", err)
					continue
				}

				deletedCount++
				logger.LOG.Info("删除过期对象",
					"bucket", bucketName,
					"key", obj.ObjectKey,
					"created_at", obj.CreatedAt)
			}
		}

		// 如果返回数量少于 lifecyclePageSize，说明已经没有更多对象了
		if len(objects) < lifecyclePageSize {
			break
		}

		// 使用最后一个对象的 ObjectKey 作为下一页的 marker
		marker = objects[len(objects)-1].ObjectKey
	}

	return deletedCount, nil
}

// processNoncurrentVersionExpiration 处理非当前版本过期
func (t *LifecycleTask) processNoncurrentVersionExpiration(ctx context.Context, bucketName, userID string, rule types.LifecycleRule) (int, error) {
	// 计算过期时间
	expirationTime := time.Now().AddDate(0, 0, -rule.NoncurrentVersionExpiration.NoncurrentDays)

	// 分页列出所有对象（包括非当前版本），用于过期删除
	deletedCount := 0
	marker := ""

	for {
		objects, err := t.factory.S3ObjectMetadata().ListByBucket(ctx, bucketName, userID, rule.Prefix, lifecyclePageSize, marker)
		if err != nil {
			return deletedCount, fmt.Errorf("列出对象失败: %w", err)
		}

		for _, obj := range objects {
			// 只处理非当前版本
			if !obj.IsLatest {
				// 检查是否过期
				if time.Time(obj.CreatedAt).Before(expirationTime) {
					// 删除特定版本
					err := t.objectService.DeleteObject(ctx, bucketName, obj.ObjectKey, userID, obj.VersionID)
					if err != nil {
						logger.LOG.Error("删除过期非当前版本失败",
							"bucket", bucketName,
							"key", obj.ObjectKey,
							"version_id", obj.VersionID,
							"error", err)
						continue
					}

					deletedCount++
				}
			}
		}

		// 如果返回数量少于 lifecyclePageSize，说明已经没有更多对象了
		if len(objects) < lifecyclePageSize {
			break
		}

		// 使用最后一个对象的 ObjectKey 作为下一页的 marker
		marker = objects[len(objects)-1].ObjectKey
	}

	return deletedCount, nil
}

// processTransition 处理存储类别转换
func (t *LifecycleTask) processTransition(ctx context.Context, bucketName, userID string, rule types.LifecycleRule, transition types.LifecycleTransition) (int, error) {
	// 计算转换时间
	var transitionTime time.Time
	if transition.Days > 0 {
		transitionTime = time.Now().AddDate(0, 0, -transition.Days)
	} else if transition.Date != "" {
		var err error
		transitionTime, err = time.Parse("2006-01-02T15:04:05Z", transition.Date)
		if err != nil {
			return 0, fmt.Errorf("解析转换日期失败: %w", err)
		}
	} else {
		return 0, fmt.Errorf("Transition必须指定Days或Date")
	}

	// 获取Bucket下的所有对象（只获取最新版本，用于存储类别转换）
	prefix := rule.Prefix
	if rule.Filter != nil && rule.Filter.Prefix != "" {
		prefix = rule.Filter.Prefix
	}

	// 分页列出对象（只获取最新版本，用于存储类别转换）
	transitionedCount := 0
	marker := ""

	for {
		objects, err := t.factory.S3ObjectMetadata().ListByBucket(ctx, bucketName, userID, prefix, lifecyclePageSize, marker)
		if err != nil {
			return transitionedCount, fmt.Errorf("列出对象失败: %w", err)
		}

		for _, obj := range objects {
			// 检查对象是否满足转换条件
			if time.Time(obj.CreatedAt).Before(transitionTime) {
				// 检查当前存储类别是否已经是目标类别
				if obj.StorageClass == transition.StorageClass {
					continue
				}

				// 检查标签过滤
				if rule.Filter != nil && rule.Filter.Tag != nil {
					if !t.matchTag(obj, rule.Filter.Tag) {
						continue
					}
				}

				// 更新存储类别
				obj.StorageClass = transition.StorageClass
				err := t.factory.S3ObjectMetadata().Update(ctx, obj)
				if err != nil {
					logger.LOG.Error("转换存储类别失败",
						"bucket", bucketName,
						"key", obj.ObjectKey,
						"storage_class", transition.StorageClass,
						"error", err)
					continue
				}

				transitionedCount++
				logger.LOG.Info("转换存储类别",
					"bucket", bucketName,
					"key", obj.ObjectKey,
					"storage_class", transition.StorageClass)
			}
		}

		// 如果返回数量少于 lifecyclePageSize，说明已经没有更多对象了
		if len(objects) < lifecyclePageSize {
			break
		}

		// 使用最后一个对象的 ObjectKey 作为下一页的 marker
		marker = objects[len(objects)-1].ObjectKey
	}

	return transitionedCount, nil
}

// processNoncurrentVersionTransition 处理非当前版本转换
func (t *LifecycleTask) processNoncurrentVersionTransition(ctx context.Context, bucketName, userID string, rule types.LifecycleRule, transition types.NoncurrentVersionTransition) (int, error) {
	// 计算转换时间
	transitionTime := time.Now().AddDate(0, 0, -transition.NoncurrentDays)

	// 获取Bucket下的所有对象版本（包括非当前版本）
	objects, err := t.factory.S3ObjectMetadata().ListVersionsByBucket(ctx, bucketName, userID, rule.Prefix, "", "", lifecyclePageSize)
	if err != nil {
		return 0, fmt.Errorf("列出对象失败: %w", err)
	}

	transitionedCount := 0
	for _, obj := range objects {
		// 只处理非当前版本
		if !obj.IsLatest {
			// 检查是否满足转换条件
			if time.Time(obj.CreatedAt).Before(transitionTime) {
				// 检查当前存储类别是否已经是目标类别
				if obj.StorageClass == transition.StorageClass {
					continue
				}

				// 更新存储类别
				obj.StorageClass = transition.StorageClass
				err := t.factory.S3ObjectMetadata().Update(ctx, obj)
				if err != nil {
					logger.LOG.Error("转换非当前版本存储类别失败",
						"bucket", bucketName,
						"key", obj.ObjectKey,
						"version_id", obj.VersionID,
						"error", err)
					continue
				}

				transitionedCount++
			}
		}
	}

	return transitionedCount, nil
}

// processAbortIncompleteMultipartUpload 处理未完成的分片上传
func (t *LifecycleTask) processAbortIncompleteMultipartUpload(ctx context.Context, bucketName, userID string, rule types.LifecycleRule) (int, error) {
	// 计算过期时间（创建时间早于这个时间的上传将被取消）
	expirationTime := time.Now().AddDate(0, 0, -rule.AbortIncompleteMultipartUpload.DaysAfterInitiation)

	// 获取Bucket下的所有未完成的分片上传（创建时间早于过期时间）
	multipartRepo := t.factory.S3Multipart()

	// 使用ListMultipartUploadsByBucket方法获取过期的上传
	uploads, err := multipartRepo.ListMultipartUploadsByBucket(ctx, bucketName, userID, expirationTime)
	if err != nil {
		logger.LOG.Error("List multipart uploads by bucket failed",
			"bucket", bucketName,
			"user_id", userID,
			"error", err,
		)
		return 0, err
	}

	// 取消每个过期的上传
	abortedCount := 0
	for _, upload := range uploads {
		// 使用objectService的AbortMultipartUpload方法取消上传
		if err := t.objectService.AbortMultipartUpload(ctx, upload.BucketName, upload.ObjectKey, upload.UploadID, upload.UserID); err != nil {
			logger.LOG.Warn("Abort multipart upload failed",
				"bucket", upload.BucketName,
				"key", upload.ObjectKey,
				"upload_id", upload.UploadID,
				"error", err,
			)
			continue
		}
		abortedCount++
		logger.LOG.Info("Aborted incomplete multipart upload",
			"bucket", upload.BucketName,
			"key", upload.ObjectKey,
			"upload_id", upload.UploadID,
			"created_at", upload.CreatedAt,
		)
	}

	logger.LOG.Info("Processed abort incomplete multipart uploads",
		"bucket", bucketName,
		"total", len(uploads),
		"aborted", abortedCount,
		"days", rule.AbortIncompleteMultipartUpload.DaysAfterInitiation,
	)

	return abortedCount, nil
}

// matchTag 检查对象是否匹配标签
func (t *LifecycleTask) matchTag(obj *models.S3ObjectMetadata, tag *types.LifecycleTag) bool {
	if obj.Tags == "" {
		return false
	}

	// 解析对象标签
	var tags map[string]string
	if err := json.Unmarshal([]byte(obj.Tags), &tags); err != nil {
		return false
	}

	value, ok := tags[tag.Key]
	return ok && value == tag.Value
}

// StartScheduledExecution 启动定时执行任务
// interval: 执行间隔（例如每小时1次）
// 返回一个 stop 函数，调用后可停止定时任务
func (t *LifecycleTask) StartScheduledExecution(interval time.Duration) context.CancelFunc {
	logger.LOG.Info("启动生命周期管理定时任务", "interval", interval)

	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()

		// 启动时立即执行一次
		if err := t.ExecuteLifecycleRules(); err != nil {
			logger.LOG.Error("生命周期管理任务执行失败", "error", err)
		}

		// 然后按间隔执行
		for {
			select {
			case <-ctx.Done():
				logger.LOG.Info("生命周期管理定时任务已停止")
				return
			case <-ticker.C:
				if err := t.ExecuteLifecycleRules(); err != nil {
					logger.LOG.Error("生命周期管理任务执行失败", "error", err)
				}
			}
		}
	}()

	return cancel
}
