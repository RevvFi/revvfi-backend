package admin

import (
	"fmt"
	"math/big"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	"github.com/Revvfi/revvfi-backend/internal/config"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*
@file market_risk_service.go

@desc
Service for admin market risk parameter management.
Handles reading risk params from the database and preparing on-chain update transactions.

@responsibilities
- Return current risk parameters for a market
- Prepare ABI-encoded calldata for min collateral ratio updates
- Prepare ABI-encoded calldata for liquidation threshold updates
- Prepare ABI-encoded calldata for oracle address updates
- Write audit log entries after each preparation
*/

/*
@struct AdminMarketRiskService

@desc
Implements admin market risk parameter management.

@fields
- repo: *postgres.AdminRepository for DB access
- cfg: blockchain configuration for contract addresses
*/
type AdminMarketRiskService struct {
	repo *postgres.AdminRepository
	cfg  config.BlockchainConfig
}

/*
@function NewAdminMarketRiskService

@desc
Creates a new AdminMarketRiskService.

@params
- repo: admin PostgreSQL repository
- cfg: blockchain configuration

@returns
- *AdminMarketRiskService
*/
func NewAdminMarketRiskService(repo *postgres.AdminRepository, cfg config.BlockchainConfig) *AdminMarketRiskService {
	return &AdminMarketRiskService{repo: repo, cfg: cfg}
}

/*
@method GetMarketRiskParams

@desc
Returns current risk parameters for the specified market.

@params
- ctx: handler context (passed as interface{} per service interface convention)
- address: market contract address

@returns
- *response.MarketRiskParams: current risk configuration
- error: ErrMarketNotFound if market does not exist
*/
func (s *AdminMarketRiskService) GetMarketRiskParams(ctx interface{}, address string) (*response.MarketRiskParams, error) {
	c := extractCtx(ctx)
	if address == "" {
		return nil, appErr.ErrInvalidAddress
	}
	row, err := s.repo.GetMarketRiskParams(c, address)
	if err != nil {
		return nil, fmt.Errorf("get market risk params: %w", err)
	}
	if row == nil {
		return nil, appErr.ErrMarketNotFound
	}
	return &response.MarketRiskParams{
		MarketAddress:           row.Address,
		MinCollateralRatioBps:   row.MinCollateralRatioBps,
		LiquidationThresholdBps: row.LiquidationThresholdBps,
		CollateralOracle:        row.CollateralOracle,
		IsActive:                row.IsActive,
		IsLiquidating:           row.IsLiquidating,
	}, nil
}

/*
@method PrepareSetMinCR

@desc
Prepares ABI-encoded calldata to update a market's minimum collateral ratio.
Updates the DB record and returns encoded transaction for on-chain submission.

@params
- ctx: handler context
- address: market contract address
- minCRBps: new minimum collateral ratio in basis points (≥ 10000 = 100%)
- reason: admin justification for the change

@returns
- *response.TransactionPreparation: encoded calldata targeting the market contract
- error
*/
func (s *AdminMarketRiskService) PrepareSetMinCR(ctx interface{}, address string, minCRBps int32, reason string) (*response.TransactionPreparation, error) {
	c := extractCtx(ctx)
	if address == "" {
		return nil, appErr.ErrInvalidAddress
	}
	if err := s.repo.UpdateMarketMinCR(c, address, minCRBps); err != nil {
		return nil, fmt.Errorf("update market min CR: %w", err)
	}
	_ = s.repo.InsertAuditLog(c, "", "update_min_cr", "market", address, "", reason)

	sel := functionSelector("setMinCollateralRatio(uint256)")
	calldata := buildCalldata(sel, padUint256(big.NewInt(int64(minCRBps))))
	return &response.TransactionPreparation{
		Target:      address,
		TxData:      calldata,
		Value:       "0",
		Description: fmt.Sprintf("Set min collateral ratio to %d bps on market %s", minCRBps, address),
	}, nil
}

/*
@method PrepareSetLiqThreshold

@desc
Prepares ABI-encoded calldata to update a market's liquidation threshold.

@params
- ctx: handler context
- address: market contract address
- thresholdBps: new liquidation threshold in basis points (≥ 5000 = 50%)
- reason: admin justification

@returns
- *response.TransactionPreparation: encoded calldata
- error
*/
func (s *AdminMarketRiskService) PrepareSetLiqThreshold(ctx interface{}, address string, thresholdBps int32, reason string) (*response.TransactionPreparation, error) {
	c := extractCtx(ctx)
	if address == "" {
		return nil, appErr.ErrInvalidAddress
	}
	if err := s.repo.UpdateMarketLiqThreshold(c, address, thresholdBps); err != nil {
		return nil, fmt.Errorf("update liq threshold: %w", err)
	}
	_ = s.repo.InsertAuditLog(c, "", "update_liq_threshold", "market", address, "", reason)

	sel := functionSelector("setLiquidationThreshold(uint256)")
	calldata := buildCalldata(sel, padUint256(big.NewInt(int64(thresholdBps))))
	return &response.TransactionPreparation{
		Target:      address,
		TxData:      calldata,
		Value:       "0",
		Description: fmt.Sprintf("Set liquidation threshold to %d bps on market %s", thresholdBps, address),
	}, nil
}

/*
@method PrepareSetOracle

@desc
Prepares ABI-encoded calldata to update a market's collateral oracle.

@params
- ctx: handler context
- address: market contract address
- oracleAddress: new Chainlink oracle contract address
- reason: admin justification

@returns
- *response.TransactionPreparation: encoded calldata
- error
*/
func (s *AdminMarketRiskService) PrepareSetOracle(ctx interface{}, address, oracleAddress, reason string) (*response.TransactionPreparation, error) {
	c := extractCtx(ctx)
	if address == "" || oracleAddress == "" {
		return nil, appErr.ErrInvalidAddress
	}
	if err := s.repo.UpdateMarketOracle(c, address, oracleAddress); err != nil {
		return nil, fmt.Errorf("update market oracle: %w", err)
	}
	_ = s.repo.InsertAuditLog(c, "", "update_oracle", "market", address, "", reason)

	sel := functionSelector("setCollateralOracle(address)")
	calldata := buildCalldata(sel, padAddress(oracleAddress))
	return &response.TransactionPreparation{
		Target:      address,
		TxData:      calldata,
		Value:       "0",
		Description: fmt.Sprintf("Set collateral oracle to %s on market %s", oracleAddress, address),
	}, nil
}
