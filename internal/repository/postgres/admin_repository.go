package postgres

import (
	"context"
	"database/sql"
	"time"
)

/*
@file admin_repository.go

@desc
PostgreSQL repository for admin-specific tables and aggregate analytics.
Owns admin_audit_logs, liquidator_config, and system_config tables.
Also provides cross-table aggregate queries for the dashboard stats service.

@responsibilities
- Insert and query audit log entries
- Read and update liquidator Dutch-auction parameters
- List and upsert system configuration key-value pairs
- Run aggregate COUNT/SUM analytics across markets, borrowers, positions, auctions, and repayments
*/

/*
@struct AdminRepository

@desc
PostgreSQL implementation for admin-specific persistence and analytics.

@fields
- db: shared database wrapper
*/
type AdminRepository struct{ db *DB }

/*
@function NewAdminRepository

@desc
Creates a PostgreSQL admin repository.

@params
- db: shared database wrapper

@returns
- *AdminRepository
*/
func NewAdminRepository(db *DB) *AdminRepository { return &AdminRepository{db: db} }

// ---------------------------------------------------------------------------
// Audit Logs
// ---------------------------------------------------------------------------

/*
@struct AuditLogRow

@desc
Raw audit log row from the database.
*/
type AuditLogRow struct {
	ID            int64
	AdminAddress  string
	Action        string
	TargetType    string
	TargetAddress sql.NullString
	TxHash        sql.NullString
	Result        string
	Reason        sql.NullString
	CreatedAt     time.Time
}

/*
@method InsertAuditLog

@desc
Inserts a new admin audit log entry. Append-only; rows are never updated.

@params
- ctx: request context
- adminAddress: wallet address of the acting admin
- action: action performed (e.g., update_min_cr)
- targetType: entity type affected (market, borrower, protocol, system, liquidator)
- targetAddress: optional address of the affected contract/wallet
- txHash: optional on-chain transaction hash
- reason: justification provided by the admin

@returns
- error
*/
func (r *AdminRepository) InsertAuditLog(
	ctx context.Context,
	adminAddress, action, targetType, targetAddress, txHash, reason string,
) error {
	_, err := r.db.conn.ExecContext(ctx,
		`INSERT INTO admin_audit_logs
		(admin_address, action, target_type, target_address, tx_hash, result, reason)
		VALUES ($1, $2, $3, NULLIF($4,''), NULLIF($5,''), 'success', NULLIF($6,''))`,
		adminAddress, action, targetType, targetAddress, txHash, reason)
	return mapError(err)
}

/*
@method ListAuditLogs

@desc
Returns a paginated, filtered list of audit log entries.

@params
- ctx: request context
- page: 1-indexed page number
- limit: entries per page (max 100)
- adminAddr: optional filter by admin address
- action: optional filter by action type
- targetType: optional filter by target type

@returns
- []AuditLogRow: log entries
- int64: total count matching the filter
- error
*/
func (r *AdminRepository) ListAuditLogs(
	ctx context.Context,
	page, limit int32,
	adminAddr, action, targetType string,
) ([]AuditLogRow, int64, error) {
	offset := (page - 1) * limit

	countQ := `SELECT COUNT(*) FROM admin_audit_logs WHERE
		($1 = '' OR admin_address = $1) AND
		($2 = '' OR action = $2) AND
		($3 = '' OR target_type = $3)`
	var total int64
	if err := r.db.conn.QueryRowContext(ctx, countQ, adminAddr, action, targetType).Scan(&total); err != nil {
		return nil, 0, mapError(err)
	}

	rows, err := r.db.conn.QueryContext(ctx,
		`SELECT id, admin_address, action, target_type, target_address, tx_hash, result, reason, created_at
		FROM admin_audit_logs
		WHERE ($1 = '' OR admin_address = $1)
		  AND ($2 = '' OR action = $2)
		  AND ($3 = '' OR target_type = $3)
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5`,
		adminAddr, action, targetType, limit, offset)
	if err != nil {
		return nil, 0, mapError(err)
	}
	defer rows.Close()

	var items []AuditLogRow
	for rows.Next() {
		var row AuditLogRow
		if err := rows.Scan(
			&row.ID, &row.AdminAddress, &row.Action, &row.TargetType,
			&row.TargetAddress, &row.TxHash, &row.Result, &row.Reason, &row.CreatedAt,
		); err != nil {
			return nil, 0, mapError(err)
		}
		items = append(items, row)
	}
	return items, total, mapError(rows.Err())
}

/*
@method GetAuditStats

@desc
Returns aggregate statistics over the entire audit log.

@returns
- total: total audit log entries
- recentActivity: entries in the last 24 hours
- successCount: entries with result='success'
- actionBreakdown: map of action → count
- adminBreakdown: map of admin_address → count
- error
*/
func (r *AdminRepository) GetAuditStats(ctx context.Context) (
	total, recentActivity, successCount int64,
	actionBreakdown, adminBreakdown map[string]int64,
	err error,
) {
	if scanErr := r.db.conn.QueryRowContext(ctx,
		`SELECT COUNT(*),
			SUM(CASE WHEN created_at >= NOW() - INTERVAL '24 hours' THEN 1 ELSE 0 END),
			SUM(CASE WHEN result = 'success' THEN 1 ELSE 0 END)
		FROM admin_audit_logs`).Scan(&total, &recentActivity, &successCount); scanErr != nil {
		err = mapError(scanErr)
		return
	}

	actionBreakdown = make(map[string]int64)
	rows, qErr := r.db.conn.QueryContext(ctx, `SELECT action, COUNT(*) FROM admin_audit_logs GROUP BY action`)
	if qErr == nil {
		defer rows.Close()
		for rows.Next() {
			var k string
			var v int64
			if rows.Scan(&k, &v) == nil {
				actionBreakdown[k] = v
			}
		}
	}

	adminBreakdown = make(map[string]int64)
	rows2, qErr2 := r.db.conn.QueryContext(ctx, `SELECT admin_address, COUNT(*) FROM admin_audit_logs GROUP BY admin_address`)
	if qErr2 == nil {
		defer rows2.Close()
		for rows2.Next() {
			var k string
			var v int64
			if rows2.Scan(&k, &v) == nil {
				adminBreakdown[k] = v
			}
		}
	}
	return
}

// ---------------------------------------------------------------------------
// Liquidator Config
// ---------------------------------------------------------------------------

/*
@struct LiquidatorConfigRow

@desc
Raw liquidator configuration row from the database.
*/
type LiquidatorConfigRow struct {
	AuctionDurationSeconds  int
	PriceDropRateBps        int
	MinBidIncrementBps      int
	LiquidationIncentiveBps int
	MaxSlippageBps          int
	UpdatedAt               time.Time
	UpdatedBy               sql.NullString
}

/*
@method GetLiquidatorConfig

@desc
Reads the single-row liquidator configuration.

@returns
- *LiquidatorConfigRow: current config
- error
*/
func (r *AdminRepository) GetLiquidatorConfig(ctx context.Context) (*LiquidatorConfigRow, error) {
	row := r.db.conn.QueryRowContext(ctx,
		`SELECT auction_duration_seconds, price_drop_rate_bps, min_bid_increment_bps,
			liquidation_incentive_bps, max_slippage_bps, updated_at, updated_by
		FROM liquidator_config ORDER BY id LIMIT 1`)
	var cfg LiquidatorConfigRow
	err := row.Scan(
		&cfg.AuctionDurationSeconds, &cfg.PriceDropRateBps, &cfg.MinBidIncrementBps,
		&cfg.LiquidationIncentiveBps, &cfg.MaxSlippageBps, &cfg.UpdatedAt, &cfg.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, mapError(err)
	}
	return &cfg, nil
}

/*
@method UpdateLiquidatorConfig

@desc
Updates the single-row liquidator configuration.

@params
- ctx: request context
- cfg: new parameter values
- updatedBy: admin wallet address performing the update

@returns
- error
*/
func (r *AdminRepository) UpdateLiquidatorConfig(
	ctx context.Context,
	cfg *LiquidatorConfigRow,
	updatedBy string,
) error {
	_, err := r.db.conn.ExecContext(ctx,
		`UPDATE liquidator_config SET
			auction_duration_seconds=$1, price_drop_rate_bps=$2, min_bid_increment_bps=$3,
			liquidation_incentive_bps=$4, max_slippage_bps=$5, updated_at=$6, updated_by=$7
		WHERE id = (SELECT id FROM liquidator_config ORDER BY id LIMIT 1)`,
		cfg.AuctionDurationSeconds, cfg.PriceDropRateBps, cfg.MinBidIncrementBps,
		cfg.LiquidationIncentiveBps, cfg.MaxSlippageBps, now(), updatedBy)
	return mapError(err)
}

// ---------------------------------------------------------------------------
// System Config
// ---------------------------------------------------------------------------

/*
@struct SystemConfigRow

@desc
Raw system configuration key-value row from the database.
*/
type SystemConfigRow struct {
	Key         string
	Value       string
	Description sql.NullString
	ConfigType  string
	IsSensitive bool
	UpdatedAt   time.Time
	UpdatedBy   sql.NullString
}

/*
@method ListSystemConfigs

@desc
Returns all system configuration entries ordered by key.

@returns
- []SystemConfigRow: all config entries
- error
*/
func (r *AdminRepository) ListSystemConfigs(ctx context.Context) ([]SystemConfigRow, error) {
	rows, err := r.db.conn.QueryContext(ctx,
		`SELECT key, value, description, config_type, is_sensitive, updated_at, updated_by
		FROM system_config ORDER BY key`)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	var items []SystemConfigRow
	for rows.Next() {
		var row SystemConfigRow
		if err := rows.Scan(
			&row.Key, &row.Value, &row.Description, &row.ConfigType,
			&row.IsSensitive, &row.UpdatedAt, &row.UpdatedBy,
		); err != nil {
			return nil, mapError(err)
		}
		items = append(items, row)
	}
	return items, mapError(rows.Err())
}

/*
@method UpsertSystemConfig

@desc
Inserts or updates a system configuration entry by key.

@params
- ctx: request context
- key: config key
- value: new value
- updatedBy: admin wallet address

@returns
- error
*/
func (r *AdminRepository) UpsertSystemConfig(
	ctx context.Context,
	key, value, updatedBy string,
) error {
	_, err := r.db.conn.ExecContext(ctx,
		`UPDATE system_config SET value=$2, updated_at=$3, updated_by=$4 WHERE key=$1`,
		key, value, now(), updatedBy)
	return mapError(err)
}

// ---------------------------------------------------------------------------
// Dashboard Stats (cross-table aggregates)
// ---------------------------------------------------------------------------

/*
@struct OverviewStatsRow

@desc
Aggregate counts and sums for the admin dashboard overview.
*/
type OverviewStatsRow struct {
	TotalMarkets    int64
	ActiveMarkets   int64
	TotalBorrowers  int64
	ActiveBorrowers int64
	ActivePositions int64
	ActiveOffers    int64
	ActiveAuctions  int64
	TotalDebt       string
	TotalLiquidity  string
	TotalPrincipal  string
}

/*
@method GetOverviewStats

@desc
Returns high-level aggregate statistics from all core tables.

@returns
- *OverviewStatsRow: aggregate metrics
- error
*/
func (r *AdminRepository) GetOverviewStats(ctx context.Context) (*OverviewStatsRow, error) {
	var stats OverviewStatsRow
	err := r.db.conn.QueryRowContext(ctx,
		`SELECT
			(SELECT COUNT(*) FROM markets),
			(SELECT COUNT(*) FROM markets WHERE is_active=true AND is_closed=false),
			(SELECT COUNT(*) FROM borrowers),
			(SELECT COUNT(*) FROM borrowers WHERE active_loans > 0),
			(SELECT COUNT(*) FROM positions WHERE is_active=true),
			(SELECT COUNT(*) FROM offers WHERE status='active'),
			(SELECT COUNT(*) FROM auctions WHERE status='active'),
			COALESCE((SELECT SUM(CAST(total_debt AS NUMERIC)) FROM markets WHERE is_active=true)::TEXT, '0'),
			COALESCE((SELECT SUM(CAST(total_liquidity AS NUMERIC)) FROM markets WHERE is_active=true)::TEXT, '0'),
			COALESCE((SELECT SUM(CAST(total_principal AS NUMERIC)) FROM markets WHERE is_active=true)::TEXT, '0')`).
		Scan(
			&stats.TotalMarkets, &stats.ActiveMarkets,
			&stats.TotalBorrowers, &stats.ActiveBorrowers,
			&stats.ActivePositions, &stats.ActiveOffers, &stats.ActiveAuctions,
			&stats.TotalDebt, &stats.TotalLiquidity, &stats.TotalPrincipal,
		)
	if err != nil {
		return nil, mapError(err)
	}
	return &stats, nil
}

/*
@struct BorrowerStatsRow

@desc
Borrower aggregate metrics for admin analytics.
*/
type BorrowerStatsRow struct {
	TotalBorrowers     int64
	ActiveBorrowers    int64
	DefaultedCount     int64
	AverageReputation  float64
	AverageSuccessRate float64
	TotalVolume        string
}

/*
@method GetBorrowerStats

@desc
Returns borrower growth and reputation statistics.

@returns
- *BorrowerStatsRow
- error
*/
func (r *AdminRepository) GetBorrowerStats(ctx context.Context) (*BorrowerStatsRow, error) {
	var stats BorrowerStatsRow
	err := r.db.conn.QueryRowContext(ctx,
		`SELECT
			COUNT(*),
			SUM(CASE WHEN active_loans > 0 THEN 1 ELSE 0 END),
			SUM(CASE WHEN defaulted_loans > 0 THEN 1 ELSE 0 END),
			COALESCE(AVG(reputation_score), 0),
			COALESCE(AVG(success_rate), 0),
			COALESCE(SUM(CAST(total_borrowed AS NUMERIC))::TEXT, '0')
		FROM borrowers`).
		Scan(
			&stats.TotalBorrowers, &stats.ActiveBorrowers, &stats.DefaultedCount,
			&stats.AverageReputation, &stats.AverageSuccessRate, &stats.TotalVolume,
		)
	if err != nil {
		return nil, mapError(err)
	}
	return &stats, nil
}

/*
@struct MarketStatsRow

@desc
Market creation and utilization statistics.
*/
type MarketStatsRow struct {
	TotalMarkets       int64
	ActiveMarkets      int64
	ClosedMarkets      int64
	LiquidatingMarkets int64
	AverageUtilization float64
	AverageAPR         float64
}

/*
@method GetMarketStats

@desc
Returns market creation and utilization metrics.

@returns
- *MarketStatsRow
- error
*/
func (r *AdminRepository) GetMarketStats(ctx context.Context) (*MarketStatsRow, error) {
	var stats MarketStatsRow
	err := r.db.conn.QueryRowContext(ctx,
		`SELECT
			COUNT(*),
			SUM(CASE WHEN is_active=true AND is_closed=false THEN 1 ELSE 0 END),
			SUM(CASE WHEN is_closed=true THEN 1 ELSE 0 END),
			SUM(CASE WHEN is_liquidating=true THEN 1 ELSE 0 END),
			COALESCE(AVG(utilization_rate) FILTER (WHERE is_active=true), 0),
			COALESCE(AVG(weighted_avg_apr) FILTER (WHERE is_active=true), 0)
		FROM markets`).
		Scan(
			&stats.TotalMarkets, &stats.ActiveMarkets, &stats.ClosedMarkets,
			&stats.LiquidatingMarkets, &stats.AverageUtilization, &stats.AverageAPR,
		)
	if err != nil {
		return nil, mapError(err)
	}
	return &stats, nil
}

/*
@struct RevenueStatsRow

@desc
Protocol revenue metrics from repayments.
*/
type RevenueStatsRow struct {
	TotalInterestCollected string
	TotalRepaid            string
	TotalRepayments        int64
}

/*
@method GetRevenueStats

@desc
Returns protocol revenue metrics aggregated from repayments.

@returns
- *RevenueStatsRow
- error
*/
func (r *AdminRepository) GetRevenueStats(ctx context.Context) (*RevenueStatsRow, error) {
	var stats RevenueStatsRow
	err := r.db.conn.QueryRowContext(ctx,
		`SELECT
			COALESCE(SUM(CAST(interest_paid AS NUMERIC))::TEXT, '0'),
			COALESCE(SUM(CAST(amount AS NUMERIC))::TEXT, '0'),
			COUNT(*)
		FROM repayments`).
		Scan(&stats.TotalInterestCollected, &stats.TotalRepaid, &stats.TotalRepayments)
	if err != nil {
		return nil, mapError(err)
	}
	return &stats, nil
}

/*
@struct LiquidationStatsRow

@desc
Liquidation auction statistics and recovery metrics.
*/
type LiquidationStatsRow struct {
	TotalAuctions             int64
	ActiveAuctions            int64
	SettledAuctions           int64
	AverageRecoveryRate       float64
	TotalCollateralLiquidated string
	TotalDebtLiquidated       string
}

/*
@method GetLiquidationStats

@desc
Returns liquidation auction statistics and recovery rates.

@returns
- *LiquidationStatsRow
- error
*/
func (r *AdminRepository) GetLiquidationStats(ctx context.Context) (*LiquidationStatsRow, error) {
	var stats LiquidationStatsRow
	err := r.db.conn.QueryRowContext(ctx,
		`SELECT
			COUNT(*),
			SUM(CASE WHEN status='active' THEN 1 ELSE 0 END),
			SUM(CASE WHEN status='settled' THEN 1 ELSE 0 END),
			COALESCE(AVG(recovery_rate) FILTER (WHERE status='settled'), 0),
			COALESCE(SUM(CAST(collateral_amount AS NUMERIC)) FILTER (WHERE status='settled')::TEXT, '0'),
			COALESCE(SUM(CAST(debt_amount AS NUMERIC)) FILTER (WHERE status='settled')::TEXT, '0')
		FROM auctions`).
		Scan(
			&stats.TotalAuctions, &stats.ActiveAuctions, &stats.SettledAuctions,
			&stats.AverageRecoveryRate, &stats.TotalCollateralLiquidated, &stats.TotalDebtLiquidated,
		)
	if err != nil {
		return nil, mapError(err)
	}
	return &stats, nil
}

/*
@struct PositionStatsRow

@desc
Position distribution metrics across seniority tiers.
*/
type PositionStatsRow struct {
	TotalPositions      int64
	ActivePositions     int64
	SettledPositions    int64
	SeniorPositions     int64
	JuniorPositions     int64
	TotalPrincipalLocked string
	TotalAccruedInterest string
}

/*
@method GetPositionStats

@desc
Returns position distribution and aggregate principal metrics.

@returns
- *PositionStatsRow
- error
*/
func (r *AdminRepository) GetPositionStats(ctx context.Context) (*PositionStatsRow, error) {
	var stats PositionStatsRow
	err := r.db.conn.QueryRowContext(ctx,
		`SELECT
			COUNT(*),
			SUM(CASE WHEN is_active=true THEN 1 ELSE 0 END),
			SUM(CASE WHEN is_settled=true THEN 1 ELSE 0 END),
			SUM(CASE WHEN seniority=0 THEN 1 ELSE 0 END),
			SUM(CASE WHEN seniority=1 THEN 1 ELSE 0 END),
			COALESCE(SUM(CAST(current_principal AS NUMERIC)) FILTER (WHERE is_active=true)::TEXT, '0'),
			COALESCE(SUM(CAST(accrued_interest AS NUMERIC)) FILTER (WHERE is_active=true)::TEXT, '0')
		FROM positions`).
		Scan(
			&stats.TotalPositions, &stats.ActivePositions, &stats.SettledPositions,
			&stats.SeniorPositions, &stats.JuniorPositions,
			&stats.TotalPrincipalLocked, &stats.TotalAccruedInterest,
		)
	if err != nil {
		return nil, mapError(err)
	}
	return &stats, nil
}

/*
@method GetDefaultedBorrowers

@desc
Returns a paginated list of borrowers who have at least one defaulted loan.

@params
- ctx: request context
- page: 1-indexed page number
- limit: entries per page

@returns
- []BorrowerRow: defaulted borrowers
- int64: total count of defaulted borrowers
- error
*/
func (r *AdminRepository) GetDefaultedBorrowers(ctx context.Context, page, limit int32) ([]BorrowerRow, int64, error) {
	offset := (page - 1) * limit

	var total int64
	if err := r.db.conn.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM borrowers WHERE defaulted_loans > 0`).Scan(&total); err != nil {
		return nil, 0, mapError(err)
	}

	rows, err := r.db.conn.QueryContext(ctx,
		`SELECT address, total_borrowed, total_repaid, outstanding_debt, total_loans,
			successful_loans, defaulted_loans, active_loans, reputation_score, risk_label,
			success_rate, registered_at, last_activity, last_reputation_update
		FROM borrowers WHERE defaulted_loans > 0
		ORDER BY defaulted_loans DESC, reputation_score ASC
		LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, 0, mapError(err)
	}
	defer rows.Close()

	var items []BorrowerRow
	for rows.Next() {
		b, err := scanBorrowerRow(rows)
		if err != nil {
			return nil, 0, mapError(err)
		}
		items = append(items, *b)
	}
	return items, total, mapError(rows.Err())
}

/*
@struct BorrowerRow

@desc
Raw borrower data from the database for admin use.
*/
type BorrowerRow struct {
	Address              string
	TotalBorrowed        string
	TotalRepaid          string
	OutstandingDebt      string
	TotalLoans           int32
	SuccessfulLoans      int32
	DefaultedLoans       int32
	ActiveLoans          int32
	ReputationScore      int32
	RiskLabel            string
	SuccessRate          float64
	RegisteredAt         time.Time
	LastActivity         sql.NullTime
	LastReputationUpdate time.Time
}

/*
@function scanBorrowerRow

@desc
Scans a SQL row into a BorrowerRow for admin use.
*/
func scanBorrowerRow(scanner interface {
	Scan(dest ...interface{}) error
}) (*BorrowerRow, error) {
	var b BorrowerRow
	err := scanner.Scan(
		&b.Address, &b.TotalBorrowed, &b.TotalRepaid, &b.OutstandingDebt,
		&b.TotalLoans, &b.SuccessfulLoans, &b.DefaultedLoans, &b.ActiveLoans,
		&b.ReputationScore, &b.RiskLabel, &b.SuccessRate,
		&b.RegisteredAt, &b.LastActivity, &b.LastReputationUpdate,
	)
	return &b, err
}

/*
@struct MarketRow

@desc
Raw market data for admin risk parameter queries.
*/
type MarketRow struct {
	Address                 string
	MinCollateralRatioBps   int32
	LiquidationThresholdBps int32
	CollateralOracle        string
	IsActive                bool
	IsLiquidating           bool
}

/*
@method GetMarketRiskParams

@desc
Returns the risk parameters for a specific market.

@params
- ctx: request context
- address: market contract address

@returns
- *MarketRow: risk parameters (nil if not found)
- error
*/
func (r *AdminRepository) GetMarketRiskParams(ctx context.Context, address string) (*MarketRow, error) {
	row := r.db.conn.QueryRowContext(ctx,
		`SELECT address, min_collateral_ratio, liquidation_threshold, collateral_oracle, is_active, is_liquidating
		FROM markets WHERE address=$1`, address)
	var m MarketRow
	err := row.Scan(
		&m.Address, &m.MinCollateralRatioBps, &m.LiquidationThresholdBps,
		&m.CollateralOracle, &m.IsActive, &m.IsLiquidating,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, mapError(err)
	}
	return &m, nil
}

/*
@method UpdateMarketMinCR

@desc
Updates the minimum collateral ratio for a specific market.

@params
- ctx: request context
- address: market contract address
- minCRBps: new minimum collateral ratio in basis points

@returns
- error
*/
func (r *AdminRepository) UpdateMarketMinCR(ctx context.Context, address string, minCRBps int32) error {
	_, err := r.db.conn.ExecContext(ctx,
		`UPDATE markets SET min_collateral_ratio=$2 WHERE address=$1`, address, minCRBps)
	return mapError(err)
}

/*
@method UpdateMarketLiqThreshold

@desc
Updates the liquidation threshold for a specific market.

@params
- ctx: request context
- address: market contract address
- thresholdBps: new liquidation threshold in basis points

@returns
- error
*/
func (r *AdminRepository) UpdateMarketLiqThreshold(ctx context.Context, address string, thresholdBps int32) error {
	_, err := r.db.conn.ExecContext(ctx,
		`UPDATE markets SET liquidation_threshold=$2 WHERE address=$1`, address, thresholdBps)
	return mapError(err)
}

/*
@method UpdateMarketOracle

@desc
Updates the collateral oracle address for a specific market.

@params
- ctx: request context
- address: market contract address
- oracleAddress: new Chainlink oracle contract address

@returns
- error
*/
func (r *AdminRepository) UpdateMarketOracle(ctx context.Context, address, oracleAddress string) error {
	_, err := r.db.conn.ExecContext(ctx,
		`UPDATE markets SET collateral_oracle=$2 WHERE address=$1`, address, oracleAddress)
	return mapError(err)
}

/*
@method UpdateMarketStatus

@desc
Updates the operational status flags of a market (used for emergency pause).

@params
- ctx: request context
- address: market contract address
- isActive: whether the market should be marked active
- isLiquidating: whether the market is in liquidation

@returns
- error
*/
func (r *AdminRepository) UpdateMarketStatus(ctx context.Context, address string, isActive, isLiquidating bool) error {
	_, err := r.db.conn.ExecContext(ctx,
		`UPDATE markets SET is_active=$2, is_liquidating=$3 WHERE address=$1`,
		address, isActive, isLiquidating)
	return mapError(err)
}

/*
@method SetAllMarketsActive

@desc
Sets is_active on all markets (used for emergency pause/unpause).

@params
- ctx: request context
- isActive: new active state for all markets

@returns
- error
*/
func (r *AdminRepository) SetAllMarketsActive(ctx context.Context, isActive bool) error {
	_, err := r.db.conn.ExecContext(ctx,
		`UPDATE markets SET is_active=$1 WHERE is_closed=false`, isActive)
	return mapError(err)
}

/*
@struct ActiveAuctionRow

@desc
Raw active auction data for the liquidator management handler.
*/
type ActiveAuctionRow struct {
	AuctionID        int64
	MarketAddress    string
	BorrowerAddress  string
	CollateralAmount string
	DebtAmount       string
	CurrentPrice     string
	StartTime        time.Time
	EndTime          time.Time
	Status           string
}

/*
@method GetActiveAuctions

@desc
Returns all auctions with status='active'.

@returns
- []ActiveAuctionRow
- error
*/
func (r *AdminRepository) GetActiveAuctions(ctx context.Context) ([]ActiveAuctionRow, error) {
	rows, err := r.db.conn.QueryContext(ctx,
		`SELECT auction_id, market_address, borrower_address, collateral_amount,
			debt_amount, current_price, start_time, end_time, status
		FROM auctions WHERE status='active' ORDER BY start_time DESC`)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	var items []ActiveAuctionRow
	for rows.Next() {
		var a ActiveAuctionRow
		if err := rows.Scan(
			&a.AuctionID, &a.MarketAddress, &a.BorrowerAddress, &a.CollateralAmount,
			&a.DebtAmount, &a.CurrentPrice, &a.StartTime, &a.EndTime, &a.Status,
		); err != nil {
			return nil, mapError(err)
		}
		items = append(items, a)
	}
	return items, mapError(rows.Err())
}

/*
@method UpdateBorrowerReputation

@desc
Sets a borrower's reputation score and recalculates the risk label.

@params
- ctx: request context
- address: borrower wallet address
- score: new reputation score (0-1000)
- riskLabel: new risk tier label (AAA/AA/A/B/C/D)

@returns
- error
*/
func (r *AdminRepository) UpdateBorrowerReputation(ctx context.Context, address string, score int32, riskLabel string) error {
	_, err := r.db.conn.ExecContext(ctx,
		`UPDATE borrowers SET reputation_score=$2, risk_label=$3, last_reputation_update=$4 WHERE address=$1`,
		address, score, riskLabel, now())
	return mapError(err)
}

/*
@method GetBorrowerByAddress

@desc
Returns a raw borrower row by wallet address.

@params
- ctx: request context
- address: borrower wallet address

@returns
- *BorrowerRow: borrower data (nil if not found)
- error
*/
func (r *AdminRepository) GetBorrowerByAddress(ctx context.Context, address string) (*BorrowerRow, error) {
	row := r.db.conn.QueryRowContext(ctx,
		`SELECT address, total_borrowed, total_repaid, outstanding_debt, total_loans,
			successful_loans, defaulted_loans, active_loans, reputation_score, risk_label,
			success_rate, registered_at, last_activity, last_reputation_update
		FROM borrowers WHERE address=$1`, address)
	b, err := scanBorrowerRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, mapError(err)
	}
	return b, nil
}
