package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"

	"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@struct MarketRepository

@desc
PostgreSQL implementation for market persistence.

@responsibilities
- Store market state
- Query active markets
- Provide aggregate values required by services
*/
type MarketRepository struct {
	db *DB
}

/*
@function NewMarketRepository

@desc
Creates a PostgreSQL market repository.

@params
- db: PostgreSQL wrapper

@returns
- *MarketRepository
*/
func NewMarketRepository(db *DB) *MarketRepository {
	return &MarketRepository{db: db}
}

/*
@method CreateMarket

@desc
Inserts a market record.
*/
func (r *MarketRepository) CreateMarket(ctx context.Context, market *models.Market) error {
	_, err := r.db.conn.ExecContext(ctx, `insert into markets(address, borrower, borrow_asset, collateral_asset, collateral_oracle, min_collateral_ratio, liquidation_threshold, total_principal, total_accrued_interest, total_debt, total_liquidity, borrow_index, weighted_avg_apr, utilization_rate, is_active, is_liquidating, is_closed, created_at, last_interest_accrual) values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		market.Address, market.Borrower, market.BorrowAsset, market.CollateralAsset, market.CollateralOracle, market.MinCollateralRatio, market.LiquidationThreshold, stringOrZeroDB(market.TotalPrincipal), stringOrZeroDB(market.TotalAccruedInterest), stringOrZeroDB(market.TotalDebt), stringOrZeroDB(market.TotalLiquidity), stringOrZeroDB(market.BorrowIndex), market.WeightedAvgAPR, market.UtilizationRate, market.IsActive, market.IsLiquidating, market.IsClosed, now(), now())
	return mapError(err)
}

/*
@method GetByAddress

@desc
Fetches a market by contract address.
*/
func (r *MarketRepository) GetByAddress(ctx context.Context, address string) (*models.Market, error) {
	row := r.db.conn.QueryRowContext(ctx, `select id,address,borrower,borrow_asset,borrow_asset_symbol,borrow_asset_decimals,collateral_asset,collateral_asset_symbol,collateral_asset_decimals,collateral_oracle,min_collateral_ratio,liquidation_threshold,total_principal,total_accrued_interest,total_debt,total_liquidity,borrow_index,weighted_avg_apr,utilization_rate,is_active,is_liquidating,is_closed,current_auction_id,created_at,closed_at,last_interest_accrual,last_health_check from markets where address=$1`, address)
	market, err := scanMarket(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, mapError(err)
	}
	return market, nil
}

/*
@method ListActive

@desc
Lists active market records with pagination.
*/
func (r *MarketRepository) ListActive(ctx context.Context, limit, offset int32, borrower string) ([]models.Market, error) {
	query := `select id,address,borrower,borrow_asset,borrow_asset_symbol,borrow_asset_decimals,collateral_asset,collateral_asset_symbol,collateral_asset_decimals,collateral_oracle,min_collateral_ratio,liquidation_threshold,total_principal,total_accrued_interest,total_debt,total_liquidity,borrow_index,weighted_avg_apr,utilization_rate,is_active,is_liquidating,is_closed,current_auction_id,created_at,closed_at,last_interest_accrual,last_health_check from markets where is_active=true`
	args := []interface{}{}
	if borrower != "" {
		args = append(args, borrower)
		query += fmt.Sprintf(" and borrower ilike $%d", len(args))
	}
	args = append(args, limit)
	query += fmt.Sprintf(" order by created_at desc limit $%d", len(args))
	args = append(args, offset)
	query += fmt.Sprintf(" offset $%d", len(args))

	rows, err := r.db.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	return scanMarkets(rows)
}

/*
@method GetAllActive

@desc
Lists all active markets for liquidation monitoring.
*/
func (r *MarketRepository) GetAllActive(ctx context.Context) ([]models.Market, error) {
	return r.ListActive(ctx, 1000, 0, "")
}

/*
@method UpdateMarket

@desc
Updates mutable market state.
*/
func (r *MarketRepository) UpdateMarket(ctx context.Context, market *models.Market) error {
	_, err := r.db.conn.ExecContext(ctx, `update markets set total_principal=$2,total_accrued_interest=$3,total_debt=$4,total_liquidity=$5,borrow_index=$6,weighted_avg_apr=$7,utilization_rate=$8,is_active=$9,is_liquidating=$10,is_closed=$11,last_interest_accrual=$12,last_health_check=$13 where address=$1`,
		market.Address, stringOrZeroDB(market.TotalPrincipal), stringOrZeroDB(market.TotalAccruedInterest), stringOrZeroDB(market.TotalDebt), stringOrZeroDB(market.TotalLiquidity), stringOrZeroDB(market.BorrowIndex), market.WeightedAvgAPR, market.UtilizationRate, market.IsActive, market.IsLiquidating, market.IsClosed, market.LastInterestAccrual, market.LastHealthCheck)
	return mapError(err)
}

/*
@method Update

@desc
Compatibility method for shared repository interfaces.
*/
func (r *MarketRepository) Update(ctx context.Context, market *models.Market) error {
	return r.UpdateMarket(ctx, market)
}

/*
@method Create

@desc
Compatibility method for shared repository interfaces.
*/
func (r *MarketRepository) Create(ctx context.Context, market *models.Market) error {
	return r.CreateMarket(ctx, market)
}

/*
@method List

@desc
Compatibility method for shared repository interfaces.
*/
func (r *MarketRepository) List(ctx context.Context, page, pageSize int32, filter map[string]interface{}) ([]models.Market, int64, error) {
	limit := pageSize
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}
	borrower, _ := filter["borrower"].(string)
	items, err := r.ListActive(ctx, limit, offset, borrower)
	return items, int64(len(items)), err
}

/*
@method GetMetrics

@desc
Returns persisted market metrics.
*/
func (r *MarketRepository) GetMetrics(ctx context.Context, address string) (map[string]interface{}, error) {
	market, err := r.GetByAddress(ctx, address)
	if err != nil || market == nil {
		return nil, err
	}
	return map[string]interface{}{"total_debt": stringOrZeroDB(market.TotalDebt), "total_liquidity": stringOrZeroDB(market.TotalLiquidity), "utilization_rate": market.UtilizationRate}, nil
}

/*
@function scanMarket

@desc
Scans a SQL row into a market model.
*/
func scanMarket(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.Market, error) {
	var market models.Market
	var principal, interest, debt, liquidity, borrowIndex string
	err := scanner.Scan(&market.ID, &market.Address, &market.Borrower, &market.BorrowAsset, &market.BorrowAssetSymbol, &market.BorrowAssetDecimals, &market.CollateralAsset, &market.CollateralAssetSymbol, &market.CollateralAssetDecimals, &market.CollateralOracle, &market.MinCollateralRatio, &market.LiquidationThreshold, &principal, &interest, &debt, &liquidity, &borrowIndex, &market.WeightedAvgAPR, &market.UtilizationRate, &market.IsActive, &market.IsLiquidating, &market.IsClosed, &market.CurrentAuctionID, &market.CreatedAt, &market.ClosedAt, &market.LastInterestAccrual, &market.LastHealthCheck)
	if err != nil {
		return nil, err
	}
	market.TotalPrincipal = parseInt(principal)
	market.TotalAccruedInterest = parseInt(interest)
	market.TotalDebt = parseInt(debt)
	market.TotalLiquidity = parseInt(liquidity)
	market.BorrowIndex = parseInt(borrowIndex)
	return &market, nil
}

/*
@function scanMarkets

@desc
Scans SQL rows into market models.
*/
func scanMarkets(rows *sql.Rows) ([]models.Market, error) {
	items := make([]models.Market, 0)
	for rows.Next() {
		item, err := scanMarket(rows)
		if err != nil {
			return nil, mapError(err)
		}
		items = append(items, *item)
	}
	return items, mapError(rows.Err())
}

/*
@function stringOrZeroDB

@desc
Converts big.Int values into decimal strings for persistence.
*/
func stringOrZeroDB(value *big.Int) string {
	if value == nil {
		return "0"
	}
	return value.String()
}
