package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"myobj/src/core/domain/request"
	"myobj/src/core/domain/response"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"
)

type AuditService struct {
	factory *impl.RepositoryFactory
}

func NewAuditService(factory *impl.RepositoryFactory) *AuditService {
	return &AuditService{factory: factory}
}

func (s *AuditService) GetRepository() *impl.RepositoryFactory {
	return s.factory
}

// GetAuditLogList 分页查询审计日志
func (s *AuditService) GetAuditLogList(req *request.AuditLogListRequest) (*models.JsonResponse, error) {
	query := &repository.AuditLogQuery{
		UserID:    req.UserID,
		Action:    req.Action,
		Keyword:   req.Keyword,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}

	logs, total, err := s.factory.AuditLog().ListByCondition(context.Background(), query)
	if err != nil {
		return models.NewJsonResponse(500, "查询审计日志失败", nil), err
	}

	list := make([]*response.AuditLogResponse, 0, len(logs))
	for _, l := range logs {
		list = append(list, &response.AuditLogResponse{
			ID:         l.ID,
			UserID:     l.UserID,
			UserName:   l.UserName,
			Action:     l.Action,
			TargetType: l.TargetType,
			TargetPath: l.TargetPath,
			TargetName: l.TargetName,
			Detail:     l.Detail,
			IP:         l.IP,
			CreatedAt:  l.CreatedAt,
		})
	}

	return models.NewJsonResponse(200, "ok", map[string]interface{}{
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
		"list":     list,
	}), nil
}

// ExportAuditLog 导出审计日志为CSV
func (s *AuditService) ExportAuditLog(req *request.AuditLogExportRequest) ([]byte, error) {
	query := &repository.AuditLogQuery{
		UserID:    req.UserID,
		Action:    req.Action,
		Keyword:   req.Keyword,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Page:      1,
		PageSize:  10000,
	}

	logs, _, err := s.factory.AuditLog().ListByCondition(context.Background(), query)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	// Write UTF-8 BOM for Excel compatibility
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(&buf)
	// Write header
	writer.Write([]string{"时间", "用户", "操作类型", "目标类型", "目标名称", "目标路径", "详情", "IP"})

	actionMap := map[string]string{
		"upload":      "上传",
		"download":    "下载",
		"rename":      "重命名",
		"move":        "移动",
		"delete":      "删除",
		"open":        "打开",
		"mkdir":       "创建目录",
		"set_public":  "设置公开",
		"extract":     "解压",
		"package":     "打包下载",
		"share":       "分享",
		"restore":     "还原",
		"permanent_delete": "永久删除",
		"empty_recycle":    "清空回收站",
	}

	targetMap := map[string]string{
		"file": "文件",
		"dir":  "目录",
	}

	for _, l := range logs {
		actionName := actionMap[l.Action]
		if actionName == "" {
			actionName = l.Action
		}
		targetName := targetMap[l.TargetType]
		if targetName == "" {
			targetName = l.TargetType
		}
		writer.Write([]string{
			l.CreatedAt.Format("2006-01-02 15:04:05"),
			l.UserName,
			actionName,
			targetName,
			l.TargetName,
			l.TargetPath,
			l.Detail,
			l.IP,
		})
	}

	writer.Flush()
	return buf.Bytes(), nil
}
