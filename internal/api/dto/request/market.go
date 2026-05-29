package request

/*
@file market.go

@desc
Request DTOs for market endpoints.
Handles market creation and query parameters.
*/

/*
@struct CreateMarketRequest

@desc
Request body for creating a new lending market.
Requires borrower authentication.

@fields
- BorrowAsset: ERC20 token to borrow
- BorrowAssetDecimals: decimal precision of borrow asset
- CollateralAsset: ERC20 collateral token
- CollateralAssetDecimals: decimal precision of collateral asset
- CollateralOracle: Chainlink oracle for price feeds
- MinCollateralRatio: minimum collateral ratio in bps (11000 = 110%)
- LiquidationThreshold: liquidation threshold in bps (9500 = 95%)
*/
type CreateMarketRequest struct {
	BorrowAsset              string `json:"borrow_asset" binding:"required,eth_addr"`
	BorrowAssetDecimals      int16  `json:"borrow_asset_decimals" binding:"required,min=0,max=18"`
	CollateralAsset          string `json:"collateral_asset" binding:"required,eth_addr"`
	CollateralAssetDecimals  int16  `json:"collateral_asset_decimals" binding:"required,min=0,max=18"`
	CollateralOracle         string `json:"collateral_oracle" binding:"required,eth_addr"`
	MinCollateralRatio       int32  `json:"min_collateral_ratio" binding:"required,min=10000,max=50000"`
	LiquidationThreshold     int32  `json:"liquidation_threshold" binding:"required,min=5000,max=20000"`
}

/*
@struct MarketListQuery

@desc
Query parameters for market listing.
Supports pagination, filtering, and sorting.

@fields
- Page: page number (1-indexed)
- PageSize: results per page
- SortBy: sort column (tvl|apr|utilization)
- Order: asc|desc
- BorrowAsset: filter by borrow asset
- IsActive: filter by active status
*/
type MarketListQuery struct {
	Page        int32  `form:"page" binding:"omitempty,min=1"`
	PageSize    int32  `form:"page_size" binding:"omitempty,min=1,max=100"`
	SortBy      string `form:"sort_by" binding:"omitempty,oneof=tvl apr utilization"`
	Order       string `form:"order" binding:"omitempty,oneof=asc desc"`
	BorrowAsset string `form:"borrow_asset" binding:"omitempty,eth_addr"`
	IsActive    *bool  `form:"is_active" binding:"omitempty"`
}
