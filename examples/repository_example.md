package repository

/*
Repository 使用示例

基本用法：

1. 初始化数据库连接
	db := database.GetDB()

2. 创建仓储工厂
	factory := repository.NewRepositoryFactory(db)

3. 使用仓储进行CRUD操作

用户操作示例：
	ctx := context.Background()

	// 创建用户
	user := &models.UserInfo{
		ID:        "user001",
		Name:      "张三",
		UserName:  "zhangsan",
		Password:  "hashed_password",
		Email:     "zhangsan@example.com",
		Phone:     "13800138000",
		GroupID:   1,
		CreatedAt: custom_type.JsonTime{Time: time.Now()},
		Space:     10737418240,
		FreeSpace: 10737418240,
	}
	err := factory.User().Create(ctx, user)

	// 查询用户
	user, err := factory.User().GetByID(ctx, "user001")
	user, err := factory.User().GetByUserName(ctx, "zhangsan")

	// 更新用户
	user.Name = "李四"
	err := factory.User().Update(ctx, user)

	// 删除用户
	err := factory.User().Delete(ctx, "user001")

	// 分页查询
	users, err := factory.User().List(ctx, 0, 10) // offset=0, limit=10

	// 统计数量
	count, err := factory.User().Count(ctx)

文件操作示例：
	// 创建文件
	file := &models.FileInfo{
		ID:          "file001",
		Name:        "document.pdf",
		RandomName:  "random123456",
		Size:        1024000,
		Mime:        "application/pdf",
		VirtualPath: "/documents/doc.pdf",
		Path:        "/data/files/random123456",
		FileHash:    "sha256hash",
		IsEnc:       false,
		IsChunk:     false,
		CreatedAt:   custom_type.JsonTime{Time: time.Now()},
		UpdatedAt:   custom_type.JsonTime{Time: time.Now()},
	}
	err := factory.FileInfo().Create(ctx, file)

	// 根据哈希查询（用于秒传）
	existingFile, err := factory.FileInfo().GetByHash(ctx, "sha256hash")

	// 批量创建
	files := []*models.FileInfo{file1, file2, file3}
	err := factory.FileInfo().BatchCreate(ctx, files)

分片文件操作示例：
	// 创建分片
	chunks := []*models.FileChunk{
		{
			ID:         "chunk001",
			FileID:     "file001",
			ChunkPath:  "/data/chunks/chunk001",
			ChunkSize:  1048576,
			ChunkHash:  "chunk_hash_001",
			ChunkIndex: 0,
		},
		{
			ID:         "chunk002",
			FileID:     "file001",
			ChunkPath:  "/data/chunks/chunk002",
			ChunkSize:  1048576,
			ChunkHash:  "chunk_hash_002",
			ChunkIndex: 1,
		},
	}
	err := factory.FileChunk().BatchCreate(ctx, chunks)

	// 获取文件所有分片
	chunks, err := factory.FileChunk().GetByFileID(ctx, "file001")

	// 删除文件所有分片
	err := factory.FileChunk().DeleteByFileID(ctx, "file001")

分享链接操作示例：
	// 创建分享
	share := &models.Share{
		UserID:        "user001",
		FileID:        "file001",
		Token:         "unique_share_token",
		ExpiresAt:     custom_type.JsonTime{Time: time.Now().Add(7 * 24 * time.Hour)},
		PasswordHash:  "hashed_password",
		DownloadCount: 0,
		CreatedAt:     custom_type.JsonTime{Time: time.Now()},
	}
	err := factory.Share().Create(ctx, share)

	// 根据token查询分享
	share, err := factory.Share().GetByToken(ctx, "unique_share_token")

	// 增加下载次数
	err := factory.Share().IncrementDownloadCount(ctx, share.ID)

	// 查询用户的所有分享
	shares, err := factory.Share().List(ctx, "user001", 0, 10)

权限管理示例：
	// 创建权限
	power := &models.Power{
		Name:        "上传文件",
		Description: "允许用户上传文件",
		CreatedAt:   custom_type.JsonTime{Time: time.Now()},
	}
	err := factory.Power().Create(ctx, power)

	// 为组分配权限
	groupPowers := []*models.GroupPower{
		{GroupID: 1, PowerID: 1},
		{GroupID: 1, PowerID: 2},
		{GroupID: 1, PowerID: 3},
	}
	err := factory.GroupPower().BatchCreate(ctx, groupPowers)

	// 获取组的所有权限
	powers, err := factory.GroupPower().GetByGroupID(ctx, 1)

	// 删除组的所有权限
	err := factory.GroupPower().DeleteByGroupID(ctx, 1)

虚拟路径操作示例：
	// 创建虚拟路径
	vpath := &models.VirtualPath{
		UserID:      "user001",
		Path:        "/documents",
		IsFile:      false,
		IsDir:       true,
		ParentLevel: "/",
		CreatedTime: custom_type.JsonTime{Time: time.Now()},
		UpdateTime:  custom_type.JsonTime{Time: time.Now()},
	}
	err := factory.VirtualPath().Create(ctx, vpath)

	// 根据路径查询
	vpath, err := factory.VirtualPath().GetByPath(ctx, "user001", "/documents")

	// 查询用户的所有虚拟路径
	vpaths, err := factory.VirtualPath().ListByUserID(ctx, "user001", 0, 100)

事务使用示例：
	err := db.Transaction(func(tx *gorm.DB) error {
		txFactory := repository.NewRepositoryFactory(tx)

		// 在事务中执行操作
		user := &models.UserInfo{...}
		if err := txFactory.User().Create(ctx, user); err != nil {
			return err
		}

		vpath := &models.VirtualPath{...}
		if err := txFactory.VirtualPath().Create(ctx, vpath); err != nil {
			return err
		}

		return nil
	})

注意事项：
1. 所有操作都使用 context.Context，支持超时和取消
2. 批量操作使用 BatchCreate 方法，默认每批100条
3. 更新操作使用 Save 方法，会更新所有字段
4. 删除操作是硬删除，如需软删除请修改模型添加 gorm.DeletedAt
5. 查询不到数据时返回 gorm.ErrRecordNotFound 错误
6. 使用事务时需要创建新的 RepositoryFactory 实例传入 tx
*/
