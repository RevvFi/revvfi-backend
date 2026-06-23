package admin

import (
	"fmt"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*
@file audit_service.go

@desc
Service for admin audit log retrieval and statistics.

@responsibilities
- Return paginated, filtered audit log entries
- Return aggregate audit statistics
- Export audit logs for compliance
- Return activity for a specific admin
- Return history for a specific action type
*/

/*
@struct AdminAuditService

@desc
Implements admin audit log querying.

@fields
- repo: admin repository for audit log queries
*/
type AdminAuditService struct {
	repo *postgres.AdminRepository
}

/*
@function NewAdminAuditService

@desc
Creates a new AdminAuditService.

@params
- repo: admin PostgreSQL repository

@returns
- *AdminAuditService
*/
func NewAdminAuditService(repo *postgres.AdminRepository) *AdminAuditService {
	return &AdminAuditService{repo: repo}
}

/*
@method GetAuditLogs

@desc
Returns paginated, filtered admin audit log entries.

@params
- ctx: handler context
- page: 1-indexed page number
- limit: entries per page (max 100)
- adminAddr: optional filter — only return entries by this admin
- action: optional filter — only return entries with this action
- targetType: optional filter — only return entries with this target type

@returns
- *response.AuditLogListResponse: paginated log list with total count
- error
*/
func (s *AdminAuditService) GetAuditLogs(
	ctx interface{},
	page, limit int32,
	adminAddr, action, targetType string,
) (*response.AuditLogListResponse, error) {
	c := extractCtx(ctx)
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, total, err := s.repo.ListAuditLogs(c, page, limit, adminAddr, action, targetType)
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}

	logs := make([]response.AuditLogEntry, 0, len(rows))
	for _, row := range rows {
		logs = append(logs, auditRowToEntry(row))
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	return &response.AuditLogListResponse{
		Logs: logs,
		Pagination: &response.PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
		Total: total,
	}, nil
}

/*
@method GetAuditStats

@desc
Returns aggregate statistics over the entire audit log.

@params
- ctx: handler context

@returns
- *response.AuditStats: aggregate audit statistics
- error
*/
func (s *AdminAuditService) GetAuditStats(ctx interface{}) (*response.AuditStats, error) {
	c := extractCtx(ctx)
	total, recentActivity, successCount, actionBreakdown, adminBreakdown, err := s.repo.GetAuditStats(c)
	if err != nil {
		return nil, fmt.Errorf("get audit stats: %w", err)
	}

	successRate := float64(0)
	if total > 0 {
		successRate = float64(successCount) / float64(total) * 100
	}
	return &response.AuditStats{
		TotalActions:    total,
		ActionBreakdown: actionBreakdown,
		AdminBreakdown:  adminBreakdown,
		RecentActivity:  recentActivity,
		SuccessRate:     successRate,
	}, nil
}

/*
@method ExportAuditLogs

@desc
Exports audit log entries within a time range.

@params
- ctx: handler context
- fromUnix: start of range (Unix timestamp, 0 = no lower bound)
- toUnix: end of range (Unix timestamp, 0 = now)
- format: export format ("json" or "csv")

@returns
- *response.AuditExportResponse: export payload with metadata
- error
*/
func (s *AdminAuditService) ExportAuditLogs(ctx interface{}, fromUnix, toUnix int64, format string) (*response.AuditExportResponse, error) {
	c := extractCtx(ctx)
	if toUnix == 0 {
		toUnix = time.Now().Unix()
	}
	if format == "" {
		format = "json"
	}

	// Use ListAuditLogs with large limit; no cursor-based export needed for MVP.
	rows, total, err := s.repo.ListAuditLogs(c, 1, 1000, "", "", "")
	if err != nil {
		return nil, fmt.Errorf("export audit logs: %w", err)
	}

	logs := make([]response.AuditLogEntry, 0, len(rows))
	for _, row := range rows {
		entry := auditRowToEntry(row)
		if (fromUnix == 0 || entry.CreatedAt >= fromUnix) &&
			(toUnix == 0 || entry.CreatedAt <= toUnix) {
			logs = append(logs, entry)
		}
	}
	_ = total
	return &response.AuditExportResponse{
		Format:       format,
		Data:         logs,
		ExportedAt:   time.Now().Unix(),
		TotalRecords: len(logs),
	}, nil
}

/*
@method GetAdminActivity

@desc
Returns recent audit log entries for a specific admin address.

@params
- ctx: handler context
- adminAddr: wallet address of the admin

@returns
- *response.AuditLogListResponse: audit log entries by this admin
- error
*/
func (s *AdminAuditService) GetAdminActivity(ctx interface{}, adminAddr string) (*response.AuditLogListResponse, error) {
	return s.GetAuditLogs(ctx, 1, 100, adminAddr, "", "")
}

/*
@method GetActionHistory

@desc
Returns all audit log entries for a specific action type.

@params
- ctx: handler context
- action: action name to filter (e.g., update_min_cr)

@returns
- *response.AuditLogListResponse: audit log entries for this action
- error
*/
func (s *AdminAuditService) GetActionHistory(ctx interface{}, action string) (*response.AuditLogListResponse, error) {
	return s.GetAuditLogs(ctx, 1, 100, "", action, "")
}

/*
@function auditRowToEntry

@desc
Converts a raw database AuditLogRow to a response AuditLogEntry.
*/
func auditRowToEntry(row postgres.AuditLogRow) response.AuditLogEntry {
	entry := response.AuditLogEntry{
		ID:           row.ID,
		AdminAddress: row.AdminAddress,
		Action:       row.Action,
		TargetType:   row.TargetType,
		Result:       row.Result,
		CreatedAt:    row.CreatedAt.Unix(),
	}
	if row.TargetAddress.Valid {
		entry.TargetAddress = row.TargetAddress.String
	}
	if row.TxHash.Valid {
		entry.TxHash = row.TxHash.String
	}
	if row.Reason.Valid {
		entry.Reason = row.Reason.String
	}
	return entry
}
