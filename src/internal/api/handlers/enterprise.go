package handlers

import (
	"fmt"
	"myobj/src/core/domain/request"
	"myobj/src/core/service"
	"myobj/src/internal/api/middleware"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
)

type EnterpriseHandler struct {
	service     *service.EnterpriseService
	spaceHandler *EnterpriseSpaceHandler
	cache        cache.Cache
}

func NewEnterpriseHandler(service *service.EnterpriseService, spaceHandler *EnterpriseSpaceHandler, cacheLocal cache.Cache) *EnterpriseHandler {
	return &EnterpriseHandler{
		service:      service,
		spaceHandler: spaceHandler,
		cache:        cacheLocal,
	}
}

func (h *EnterpriseHandler) Router(c *gin.RouterGroup) {
	factory := h.service.GetRepository()
	verify := middleware.NewAuthMiddleware(h.cache,
		factory.ApiKey(),
		factory.User(),
		factory.GroupPower(),
		factory.Power())

	enterpriseVerify := middleware.NewEnterpriseMiddleware(
		factory.EnterpriseMember(),
		factory.EnterpriseRole(),
		factory.EnterpriseRolePower(),
		factory.Power())

	enterprise := c.Group("/enterprise")
	enterprise.Use(verify.Verify())
	{
		// 个人操作（不需要企业上下文）
		enterprise.POST("/create", h.CreateEnterprise)
		enterprise.GET("/list", h.GetMyEnterprises)
		enterprise.GET("/info", h.GetEnterpriseInfo)
		enterprise.POST("/switch", h.SwitchEnterprise)
		enterprise.POST("/member/join", h.JoinEnterprise)
		enterprise.POST("/member/leave", h.LeaveEnterprise)
		enterprise.POST("/member/accept", h.AcceptInvite)
		enterprise.POST("/dissolve", h.DissolveEnterprise)

		// 需要企业上下文的操作（成员级别）
		withEnterprise := enterprise.Group("")
		withEnterprise.Use(enterpriseVerify.Verify())
		{
			withEnterprise.GET("/member/list", h.GetMemberList)
			withEnterprise.GET("/role/list", h.GetRoleList)
			withEnterprise.GET("/member/pending", h.GetPendingInvites)

			// 共享空间操作（需要企业上下文 + 具体权限）
			if h.spaceHandler != nil {
				space := withEnterprise.Group("/space")
				{
					space.POST("/mkdir", enterpriseVerify.PowerVerify("enterprise:space:upload"), h.spaceHandler.CreateDir)
					space.GET("/list", enterpriseVerify.PowerVerify("enterprise:space:download"), h.spaceHandler.ListFiles)
					space.POST("/upload/precheck", enterpriseVerify.PowerVerify("enterprise:space:upload"), h.spaceHandler.UploadPrecheck)
					space.POST("/upload", enterpriseVerify.PowerVerify("enterprise:space:upload"), h.spaceHandler.UploadFile)
					space.POST("/delete", enterpriseVerify.PowerVerify("enterprise:space:delete"), h.spaceHandler.DeleteFile)
					space.POST("/delete-dir", enterpriseVerify.PowerVerify("enterprise:space:delete"), h.spaceHandler.DeleteDir)
					space.POST("/rename", enterpriseVerify.PowerVerify("enterprise:space:upload"), h.spaceHandler.RenameFile)
					space.POST("/rename-dir", enterpriseVerify.PowerVerify("enterprise:space:upload"), h.spaceHandler.RenameDir)
					space.GET("/download", enterpriseVerify.PowerVerify("enterprise:space:download"), h.spaceHandler.DownloadFile)
					space.GET("/usage", h.spaceHandler.GetSpaceUsage)
				}
			}

			// 审计日志（需要 enterprise:audit:view 权限）
			auditGroup := withEnterprise.Group("/audit")
			auditGroup.Use(enterpriseVerify.PowerVerify("enterprise:audit:view"))
			{
				auditGroup.GET("/list", h.GetEnterpriseAuditLogs)
				auditGroup.GET("/export", h.ExportEnterpriseAuditLogs)
			}

			// 需要企业管理员权限的操作
			admin := withEnterprise.Group("")
			admin.Use(enterpriseVerify.AdminVerify())
			{
				admin.PUT("/update", h.UpdateEnterprise)
				admin.POST("/member/invite", enterpriseVerify.PowerVerify("enterprise:member:invite"), h.InviteMember)
				admin.GET("/member/invite-code", h.GetInviteCode)
				admin.POST("/member/refresh-code", h.RefreshInviteCode)
				admin.PUT("/member/role", enterpriseVerify.PowerVerify("enterprise:role:manage"), h.UpdateMemberRole)
				admin.POST("/member/remove", enterpriseVerify.PowerVerify("enterprise:member:remove"), h.RemoveMember)
				admin.POST("/role/create", enterpriseVerify.PowerVerify("enterprise:role:manage"), h.CreateRole)
				admin.PUT("/role/update", enterpriseVerify.PowerVerify("enterprise:role:manage"), h.UpdateRole)
				admin.DELETE("/role/delete", enterpriseVerify.PowerVerify("enterprise:role:manage"), h.DeleteRole)
				admin.GET("/powers", h.GetAllPowers)
				admin.POST("/transfer", h.TransferOwnership)
				admin.POST("/toggle-state", h.ToggleEnterpriseState)
				admin.POST("/space/set-quota", h.SetEnterpriseQuota)
			}
		}
	}

	logger.LOG.Info("[路由] 企业路由注册完成✔️")
}

func (h *EnterpriseHandler) CreateEnterprise(c *gin.Context) {
	req := new(request.CreateEnterpriseRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.CreateEnterprise(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) GetMyEnterprises(c *gin.Context) {
	result, err := h.service.GetMyEnterprises(c.GetString("userID"))
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "查询失败", err.Error()))
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) GetEnterpriseInfo(c *gin.Context) {
	enterpriseID := c.Query("enterprise_id")
	if enterpriseID == "" {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	result, err := h.service.GetEnterpriseInfo(enterpriseID, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) UpdateEnterprise(c *gin.Context) {
	req := new(request.UpdateEnterpriseRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.UpdateEnterprise(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) SwitchEnterprise(c *gin.Context) {
	req := new(request.SwitchEnterpriseRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.SwitchEnterprise(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) InviteMember(c *gin.Context) {
	req := new(request.InviteMemberRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.InviteMember(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) GetInviteCode(c *gin.Context) {
	enterpriseID := c.Query("enterprise_id")
	if enterpriseID == "" {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	result, err := h.service.GetInviteCode(enterpriseID, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) RefreshInviteCode(c *gin.Context) {
	enterpriseID := c.Query("enterprise_id")
	if enterpriseID == "" {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	result, err := h.service.RefreshInviteCode(enterpriseID, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) JoinEnterprise(c *gin.Context) {
	req := new(request.JoinEnterpriseRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.JoinEnterprise(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) GetMemberList(c *gin.Context) {
	req := new(request.EnterpriseListRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.GetMemberList(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) UpdateMemberRole(c *gin.Context) {
	req := new(request.UpdateMemberRoleRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.UpdateMemberRole(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) RemoveMember(c *gin.Context) {
	req := new(request.RemoveMemberRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.RemoveMember(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) LeaveEnterprise(c *gin.Context) {
	req := new(request.LeaveEnterpriseRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.LeaveEnterprise(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) GetPendingInvites(c *gin.Context) {
	result, err := h.service.GetPendingInvites(c.GetString("userID"))
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "查询失败", err.Error()))
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) AcceptInvite(c *gin.Context) {
	inviteID := c.Query("invite_id")
	if inviteID == "" {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	result, err := h.service.AcceptInvite(inviteID, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) GetRoleList(c *gin.Context) {
	enterpriseID := c.Query("enterprise_id")
	if enterpriseID == "" {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	result, err := h.service.GetRoleList(enterpriseID, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) CreateRole(c *gin.Context) {
	req := new(request.CreateEnterpriseRoleRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.CreateRole(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) UpdateRole(c *gin.Context) {
	req := new(request.UpdateEnterpriseRoleRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.UpdateRole(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) DeleteRole(c *gin.Context) {
	req := new(request.DeleteEnterpriseRoleRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.DeleteRole(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) GetAllPowers(c *gin.Context) {
	result, err := h.service.GetAllPowers()
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "查询失败", err.Error()))
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) TransferOwnership(c *gin.Context) {
	req := new(request.TransferOwnershipRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.TransferOwnership(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) DissolveEnterprise(c *gin.Context) {
	req := new(request.DissolveEnterpriseRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.DissolveEnterprise(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) ToggleEnterpriseState(c *gin.Context) {
	req := new(request.ToggleEnterpriseStateRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.ToggleEnterpriseState(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) SetEnterpriseQuota(c *gin.Context) {
	req := new(request.SetEnterpriseQuotaRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.SetEnterpriseQuota(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) GetEnterpriseAuditLogs(c *gin.Context) {
	req := new(request.EnterpriseAuditListRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.GetEnterpriseAuditLogs(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseHandler) ExportEnterpriseAuditLogs(c *gin.Context) {
	enterpriseID := c.Query("enterprise_id")
	if enterpriseID == "" {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", nil))
		return
	}

	csvData, err := h.service.ExportEnterpriseAuditLogs(
		enterpriseID,
		c.Query("action"),
		c.Query("keyword"),
		c.Query("start_time"),
		c.Query("end_time"),
		c.GetString("userID"),
	)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "导出失败", err.Error()))
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	eid := enterpriseID
	if len(eid) > 8 {
		eid = eid[:8]
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=enterprise_audit_%s.csv", eid))
	c.Data(200, "text/csv; charset=utf-8", csvData)
}
