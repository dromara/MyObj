package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"myobj/src/core/domain/request"
	"myobj/src/core/domain/response"
	"myobj/src/core/service"
	"myobj/src/internal/api/middleware"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/pkg/util"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VideoHandler struct {
	fileService *service.FileService
	cache       cache.Cache
}

func NewVideoHandler(fileService *service.FileService, cacheLocal cache.Cache) *VideoHandler {
	return &VideoHandler{
		fileService: fileService,
		cache:       cacheLocal,
	}
}

func (v *VideoHandler) Router(c *gin.RouterGroup) {
	verify := middleware.NewAuthMiddleware(v.cache,
		v.fileService.GetRepository().ApiKey(),
		v.fileService.GetRepository().User(),
		v.fileService.GetRepository().GroupPower(),
		v.fileService.GetRepository().Power())

	videoGroup := c.Group("/video")
	{
		// 创建视频播放预检（需要登录）
		videoGroup.POST("/play/precheck", verify.Verify(), middleware.PowerVerify("file:preview"), v.CreateVideoPlay)
		// 视频流播放
		videoGroup.GET("/stream", verify.Verify(), middleware.PowerVerify("file:preview"), v.VideoPlay)
	}

	logger.LOG.Info("[路由] 视频播放路由注册完成✔️")
}

// PlayTokenInfo 播放 Token 信息（存储在缓存中）
type PlayTokenInfo struct {
	FileID      string    `json:"file_id"`
	UserID      string    `json:"user_id"`
	PasswordKey string    `json:"password_key"` // 解密密钥（如果是加密文件）
	FileSize    int64     `json:"file_size"`    // 文件大小
	EncPath     string    `json:"enc_path"`     // 加密文件路径
	IsEnc       bool      `json:"is_enc"`       // 是否加密
	MimeType    string    `json:"mime_type"`    // MIME 类型
	CreatedAt   time.Time `json:"created_at"`   // 创建时间
}

// CreateVideoPlay godoc
// @Summary 创建视频播放预检
// @Description 验证权限并生成24小时有效的播放 Token
// @Tags 视频播放
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.VideoPlayPrecheckRequest true "预检请求"
// @Success 200 {object} models.JsonResponse{data=response.VideoPlayTokenResponse} "播放 Token"
// @Failure 400 {object} models.JsonResponse "请求失败"
// @Failure 403 {object} models.JsonResponse "权限不足"
// @Failure 404 {object} models.JsonResponse "文件不存在"
// @Router /video/play/precheck [post]
func (v *VideoHandler) CreateVideoPlay(c *gin.Context) {
	req := new(request.VideoPlayPrecheckRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		logger.LOG.Error("参数绑定失败", "error", err)
		c.JSON(400, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}

	userID := c.GetString("userID")
	if userID == "" {
		logger.LOG.Error("用户ID为空")
		c.JSON(403, models.NewJsonResponse(403, "未授权访问", nil))
		return
	}

	// 1. 查询文件信息  验证用户权限
	ctx := context.Background()
	userFile, err := v.fileService.GetRepository().UserFiles().GetByUfID(ctx, req.FileID)
	if err != nil {
		logger.LOG.Error("文件不存在", "error", err, "fileID", req.FileID)
		c.JSON(404, models.NewJsonResponse(404, "文件不存在", nil))
		return
	}
	// 验证文件是否属于该用户或是公开文件
	if !userFile.IsPublic && userFile.UserID != userID {
		logger.LOG.Error("用户无权访问该文件", "userID", userID, "fileID", req.FileID)
		c.JSON(403, models.NewJsonResponse(403, "无权访问该文件", nil))
		return
	}
	fileInfo, err := v.fileService.GetRepository().FileInfo().GetByID(ctx, userFile.FileID)
	if err != nil {
		logger.LOG.Error("查询文件信息失败", "error", err, "fileID", req.FileID)
		c.JSON(404, models.NewJsonResponse(404, "文件不存在", nil))
		return
	}

	// 2. 生成唯一播放 Token
	playToken := uuid.New().String()

	// 3. 准备 Token 信息
	// 根据是否加密选择正确的文件路径
	// 对于加密文件，优先使用 EncPath；对于普通文件，使用 Path
	// 如果优先路径为空，则尝试使用另一个路径
	filePath := ""
	if fileInfo.IsEnc {
		filePath = fileInfo.EncPath
		if filePath == "" {
			filePath = fileInfo.Path
		}
	} else {
		filePath = fileInfo.Path
		if filePath == "" {
			filePath = fileInfo.EncPath
		}
	}

	// 如果路径仍然为空，记录错误
	if filePath == "" {
		logger.LOG.Error("文件路径为空", "fileID", fileInfo.ID, "isEnc", fileInfo.IsEnc, "path", fileInfo.Path, "encPath", fileInfo.EncPath)
		c.JSON(500, models.NewJsonResponse(500, "文件路径不存在", nil))
		return
	}

	tokenInfo := PlayTokenInfo{
		FileID:    req.FileID,
		UserID:    userID,
		FileSize:  int64(fileInfo.Size),
		EncPath:   filePath, // 使用实际的文件路径（无论是否加密）
		IsEnc:     fileInfo.IsEnc,
		MimeType:  fileInfo.Mime,
		CreatedAt: time.Now(),
	}

	// 4. 如果是加密文件，需要解密密钥
	if fileInfo.IsEnc {
		// 验证用户是否提供了文件密码
		if req.SharePassword == "" {
			logger.LOG.Warn("加密文件缺少解密密码", "fileID", req.FileID)
			c.JSON(400, models.NewJsonResponse(400, "加密文件需要提供密码", nil))
			return
		}

		// 查询用户信息，验证密码是否正确
		user, err := v.fileService.GetRepository().User().GetByID(c.Request.Context(), userID)
		if err != nil {
			logger.LOG.Error("查询用户信息失败", "error", err, "userID", userID)
			c.JSON(500, models.NewJsonResponse(500, "系统错误", nil))
			return
		}

		// 验证用户输入的密码是否与存储的哈希匹配
		if !util.CheckPassword(user.FilePassword, req.SharePassword) {
			logger.LOG.Warn("文件密码错误", "userID", userID, "fileID", req.FileID)
			c.JSON(403, models.NewJsonResponse(403, "密码错误", nil))
			return
		}

		// 使用 PBKDF2 从明文密码和用户ID派生加密密钥
		tokenInfo.PasswordKey = util.DeriveEncryptionKey(req.SharePassword, userID)
		logger.LOG.Debug("派生视频播放解密密钥", "userID", userID, "keyLength", len(tokenInfo.PasswordKey))
	}

	// 5. 存储到缓存（24小时有效）
	tokenInfoJSON, err := json.Marshal(tokenInfo)
	if err != nil {
		logger.LOG.Error("序列化 Token 信息失败", "error", err)
		c.JSON(500, models.NewJsonResponse(500, "生成播放 Token 失败", nil))
		return
	}

	cacheKey := fmt.Sprintf("video_play:%s", playToken)
	if err := v.cache.Set(cacheKey, string(tokenInfoJSON), 24*60*60); err != nil {
		logger.LOG.Error("存储 Token 到缓存失败", "error", err, "cacheKey", cacheKey)
		c.JSON(500, models.NewJsonResponse(500, "生成播放 Token 失败", nil))
		return
	}
	resp := response.VideoPlayTokenResponse{
		PlayToken: playToken,
		FileInfo: response.VideoFileInfo{
			FileID:   fileInfo.ID,
			FileName: userFile.FileName,
			FileSize: int64(fileInfo.Size),
			IsEnc:    fileInfo.IsEnc,
			MimeType: fileInfo.Mime,
		},
	}

	logger.LOG.Debug("创建视频播放 Token 成功", "userID", userID, "fileID", req.FileID, "token", playToken)
	c.JSON(200, models.NewJsonResponse(200, "成功", resp))
}

// VideoPlay godoc
// @Summary 视频流播放
// @Description 基于播放 Token 流式传输视频，支持 Range 请求
// @Tags 视频播放
// @Produce video/mp4
// @Param token query string true "播放 Token"
// @Param Range header string false "Range 请求头" default(bytes=0-)
// @Success 206 "视频流数据"
// @Failure 400 {object} models.JsonResponse "请求失败"
// @Failure 403 {object} models.JsonResponse "Token 无效或已过期"
// @Failure 404 {object} models.JsonResponse "文件不存在"
// @Router /video/stream [get]
func (v *VideoHandler) VideoPlay(c *gin.Context) {
	// 1. 获取 Token
	playToken := c.Query("token")
	if playToken == "" {
		logger.LOG.Error("缺少播放 Token")
		c.JSON(400, models.NewJsonResponse(400, "缺少播放 Token", nil))
		return
	}

	// 2. 从缓存中获取 Token 信息
	cacheKey := fmt.Sprintf("video_play:%s", playToken)
	tokenInfoStr, err := v.cache.Get(cacheKey)
	if err != nil {
		logger.LOG.Error("Token 无效或已过期", "error", err, "token", playToken)
		c.JSON(403, models.NewJsonResponse(403, "Token 无效或已过期", nil))
		return
	}

	var tokenInfo PlayTokenInfo
	if err := json.Unmarshal([]byte(tokenInfoStr.(string)), &tokenInfo); err != nil {
		logger.LOG.Error("解析 Token 信息失败", "error", err)
		c.JSON(500, models.NewJsonResponse(500, "Token 信息损坏", nil))
		return
	}

	// 3. 解析 Range 请求头
	rangeHeader := c.GetHeader("Range")
	hasRangeHeader := rangeHeader != ""
	rangeInfo, err := util.ParseRange(rangeHeader, tokenInfo.FileSize)
	if err != nil {
		logger.LOG.Error("解析 Range 失败", "error", err, "rangeHeader", rangeHeader)
		c.JSON(400, models.NewJsonResponse(400, "无效的 Range 请求", nil))
		return
	}

	// 4. 设置响应头
	util.SetRangeHeaders(c.Writer, rangeInfo, tokenInfo.MimeType, hasRangeHeader)

	// 5. 根据是否加密选择传输方式
	if tokenInfo.IsEnc {
		// 加密文件：流式解密传输
		if err := util.StreamDecryptRange(c.Writer, tokenInfo.EncPath, tokenInfo.PasswordKey, rangeInfo); err != nil {
			logger.LOG.Error("流式解密传输失败", "error", err, "fileID", tokenInfo.FileID)
			// 这里不能再写 JSON 响应，因为已经开始写入视频流了
			return
		}
	} else {
		// 普通文件：直接流式传输
		if err := util.StreamPlainRange(c.Writer, tokenInfo.EncPath, rangeInfo); err != nil {
			logger.LOG.Error("流式传输失败", "error", err, "fileID", tokenInfo.FileID)
			return
		}
	}

	logger.LOG.Debug("视频流传输完成", "fileID", tokenInfo.FileID, "range", fmt.Sprintf("%d-%d", rangeInfo.Start, rangeInfo.End))
}
