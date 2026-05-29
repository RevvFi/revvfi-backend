package response

/*
@file market.go

@desc
Response DTOs for market endpoints.
Returns market data, metrics, and health status.
*/

/*
@struct MarketResponse

@desc
Response containing detailed market information.

@fields
- Address: market contract address
- Borrower: market creator
- BorrowAsset: borrow token details
- CollateralAsset: collateral token details
- TotalPrincipal: total borrowed
- TotalLiquidity: available liquidity
- UtilizationRate: debt / liquidity
- WeightedAPR: average APR
- IsActive: operational status
- IsLiquidating: liquidation status
- CreatedAt: creation timestamp
*/
type MarketResponse struct {
	Address            string         `json:"address"`
	Borrower           string         `json:"borrower"`
	BorrowAsset        TokenInfo      `json:"borrow_asset"`
	CollateralAsset    TokenInfo      `json:"collateral_asset"`
	TotalPrincipal     string         `json:"total_principal"`
	TotalLiquidity     string         `json:"total_liquidity"`
	TotalDebt          string         `json:"total_debt"`
	UtilizationRate    float64        `json:"utilization_rate"`
	WeightedAPR        int32          `json:"weighted_apr"`
	IsActive           bool           `json:"is_active"`
	IsLiquidating      bool           `json:"is_liquidating"`
	IsClosed           bool           `json:"is_closed"`
	CreatedAt          int64          `json:"created_at"`
}

/*
@struct TokenInfo

@desc
Information about an ERC20 token.

@fields
- Address: token contract address
- Symbol: token symbol
- Decimals: decimal precision
- Name: human-readable name
*/
type TokenInfo struct {
	Address  string `json:"address"`
	Symbol   string `json:"symbol"`
	Decimals int16  `json:"decimals"`
	Name     string `json:"name,omitempty"`
}

/*
@struct MarketStatusResponse

@desc
Real-time market health status.

@fields
- Address: market address
- HealthFactor: market solvency metric
- OracleFreshness: time since last oracle update
- IsHealthy: health status
- HighestCollateralRatio: max collateral ratio
- LiquidationInProgress: if liquidation is active
*/
type MarketStatusResponse struct {
	Address                string  `json:"address"`
	HealthFactor           float64 `json:"health_factor"`
	OracleFreshness        int64   `json:"oracle_freshness"`
	IsHealthy              bool    `json:"is_healthy"`
	HighestCollateralRatio float64 `json:"highest_collateral_ratio"`
	LiquidationInProgress  bool    `json:"liquidation_in_progress"`
}

/*
@struct MarketMetricsResponse

@desc
Historical market analytics and trends.

@fields
- Address: market address
- TVLHistory: TVL over time
- APRHistory: APR changes
- UtilizationHistory: utilization trends
- ActivePositions: number of active positions
*/
type MarketMetricsResponse struct {
	Address                string                  `json:"address"`
	TVLHistory             []DataPoint             `json:"tvl_history"`
	APRHistory             []DataPoint             `json:"apr_history"`
	UtilizationHistory     []DataPoint             `json:"utilization_history"`
	ActivePositions        int32                   `json:"active_positions"`
	TotalVolumeTraded      string                  `json:"total_volume_traded"`
}

/*
@struct DataPoint

@desc
Time-series data point.

@fields
- Timestamp: unix timestamp
- Value: numeric value
*/
type DataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}
