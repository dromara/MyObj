package handlers

import (
	"fmt"
	"myobj/src/core/domain/request"
	"myobj/src/core/service"
	"myobj/src/internal/api/middleware"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	service *service.AuditService
	cache   cache.Cache
}

func NewAuditHandler(service *service.AuditService, cacheLocal cache.Cache) *AuditHandler {
	return &AuditHandler{
		service: service,
		cache:   cacheLocal,
	}
}

func (h *AuditHandler) Router(c *gin.RouterGroup) {
	verify := middleware.NewAuthMiddleware(h.cache,
		h.service.GetRepository().ApiKey(),
		h.service.GetRepository().User(),
		h.service.GetRepository().GroupPower(),
		h.service.GetRepository().Power())

	admin := c.Group("/admin")
	admin.Use(verify.Verify())
	admin.Use(middleware.AdminVerify())
	{
		// 审计日志查询
		admin.GET("/audit/list", h.GetAuditLogList)
		// 审计日志导出
		admin.GET("/audit/export", h.ExportAuditLog)
	}

	logger.LOG.Info("[路由] 审计日志路由注册完成✔️")
}

// GetAuditLogList godoc
// @Summary 查询审计日志
// @Description 分页查询审计日志，支持按用户/操作类型/关键词/时间范围筛选
// @Tags 审计日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int true "页码" minimum(1)
// @Param pageSize query int true "每页数量" minimum(1) maximum(100)
// @Param user_id query string false "用户ID"
// @Param action query string false "操作类型"
// @Param keyword query string false "关键词"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Success 200 {object} models.JsonResponse{data=object} "审计日志列表"
// @Failure 400 {object} models.JsonResponse "参数错误"
// @Failure 500 {object} models.JsonResponse "查询失败"
// @Router /admin/audit/list [get]
func (h *AuditHandler) GetAuditLogList(c *gin.Context) {
	req := new(request.AuditLogListRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}

	result, err := h.service.GetAuditLogList(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "查询审计日志失败", err.Error()))
		return
	}

	c.JSON(200, result)
}

// ExportAuditLog godoc
// @Summary 导出审计日志
// @Description 导出审计日志为CSV文件
// @Tags 审计日志
// @Accept json
// @Produce text/csv
// @Security BearerAuth
// @Param user_id query string false "用户ID"
// @Param action query string false "操作类型"
// @Param keyword query string false "关键词"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Success 200 {file} csv "CSV文件"
// @Failure 500 {object} models.JsonResponse "导出失败"
// @Router /admin/audit/export [get]
func (h *AuditHandler) ExportAuditLog(c *gin.Context) {
	req := new(request.AuditLogExportRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}

	data, err := h.service.ExportAuditLog(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "导出审计日志失败", err.Error()))
		return
	}

	fileName := fmt.Sprintf("audit_log_%s.csv", time.Now().Format("20060102150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}
