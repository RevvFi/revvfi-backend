package response

/*
@file admin_audit.go

@desc
Response DTOs for admin audit log endpoints.
Structures audit trail data for compliance and monitoring.

@responsibilities
- Define audit log entry response shapes
- Define audit statistics aggregates
- Define export response format
*/

/*
@struct AuditLogEntry

@desc
Single audit log entry recording an admin action.

@fields
- ID: unique log entry identifier
- AdminAddress: wallet address of the admin who performed the action
- Action: action performed (e.g., update_min_cr, stop_auction)
- TargetType: type of target (market, borrower, protocol, system, liquidator)
- TargetAddress: contract or wallet address affected
- TxHash: transaction hash if action produced an on-chain tx
- Result: outcome of the action (success, failure)
- Reason: justification provided by admin
- CreatedAt: Unix timestamp of the action
*/
type AuditLogEntry struct {
	ID            int64  `json:"id"`
	AdminAddress  string `json:"admin_address"`
	Action        string `json:"action"`
	TargetType    string `json:"target_type"`
	TargetAddress string `json:"target_address,omitempty"`
	TxHash        string `json:"tx_hash,omitempty"`
	Result        string `json:"result"`
	Reason        string `json:"reason,omitempty"`
	CreatedAt     int64  `json:"created_at"`
}

/*
@struct AuditLogListResponse

@desc
Paginated list of audit log entries with metadata.

@fields
- Logs: list of audit log entries
- Pagination: pagination metadata
- Total: total count matching the query
*/
type AuditLogListResponse struct {
	Logs       []AuditLogEntry `json:"logs"`
	Pagination *PaginationInfo `json:"pagination"`
	Total      int64           `json:"total"`
}

/*
@struct AuditStats

@desc
Aggregate statistics over the admin audit log.

@fields
- TotalActions: total number of admin actions logged
- ActionBreakdown: count per action type
- AdminBreakdown: count per admin address
- RecentActivity: actions in the last 24 hours
- SuccessRate: percentage of successful actions
*/
type AuditStats struct {
	TotalActions    int64              `json:"total_actions"`
	ActionBreakdown map[string]int64   `json:"action_breakdown"`
	AdminBreakdown  map[string]int64   `json:"admin_breakdown"`
	RecentActivity  int64              `json:"recent_activity_24h"`
	SuccessRate     float64            `json:"success_rate"`
}

/*
@struct AuditExportResponse

@desc
Export payload for audit log data.

@fields
- Format: export format (json, csv)
- Data: exported audit log entries
- ExportedAt: Unix timestamp of the export
- TotalRecords: total records exported
*/
type AuditExportResponse struct {
	Format       string          `json:"format"`
	Data         []AuditLogEntry `json:"data"`
	ExportedAt   int64           `json:"exported_at"`
	TotalRecords int             `json:"total_records"`
}
