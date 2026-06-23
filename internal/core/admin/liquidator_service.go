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
@file liquidator_service.go

@desc
Service for admin liquidator configuration and active auction management.

@responsibilities
- Return current Dutch auction parameters from the database
- Prepare ABI-encoded calldata to update liquidator contract parameters
- Return active auction list for monitoring
- Prepare ABI-encoded calldata to stop/cancel an active auction
*/

/*
@struct AdminLiquidatorService

@desc
Implements admin liquidator management.

@fields
- repo: admin repository for liquidator_config and auction data
- cfg: blockchain configuration for the liquidator contract address
*/
type AdminLiquidatorService struct {
	repo *postgres.AdminRepository
	cfg  config.BlockchainConfig
}

/*
@function NewAdminLiquidatorService

@desc
Creates a new AdminLiquidatorService.

@params
- repo: admin PostgreSQL repository
- cfg: blockchain configuration

@returns
- *AdminLiquidatorService
*/
func NewAdminLiquidatorService(repo *postgres.AdminRepository, cfg config.BlockchainConfig) *AdminLiquidatorService {
	return &AdminLiquidatorService{repo: repo, cfg: cfg}
}

/*
@method GetLiquidatorConfig

@desc
Returns the current liquidator contract configuration parameters.

@params
- ctx: handler context

@returns
- *response.LiquidatorConfig: current Dutch auction parameters
- error
*/
func (s *AdminLiquidatorService) GetLiquidatorConfig(ctx interface{}) (*response.LiquidatorConfig, error) {
	c := extractCtx(ctx)
	row, err := s.repo.GetLiquidatorConfig(c)
	if err != nil {
		return nil, fmt.Errorf("get liquidator config: %w", err)
	}
	if row == nil {
		return nil, appErr.ErrInternal
	}
	updatedBy := ""
	if row.UpdatedBy.Valid {
		updatedBy = row.UpdatedBy.String
	}
	return &response.LiquidatorConfig{
		AuctionDurationSeconds:  row.AuctionDurationSeconds,
		PriceDropRateBps:        row.PriceDropRateBps,
		MinBidIncrementBps:      row.MinBidIncrementBps,
		LiquidationIncentiveBps: row.LiquidationIncentiveBps,
		MaxSlippageBps:          row.MaxSlippageBps,
		UpdatedAt:               row.UpdatedAt.Unix(),
		UpdatedBy:               updatedBy,
	}, nil
}

/*
@method PrepareSetLiquidatorParams

@desc
Updates liquidator config in the DB and returns ABI-encoded calldata to update
the on-chain liquidator contract.

@params
- ctx: handler context
- auctionDurationSecs: new auction duration in seconds
- priceDropRateBps: hourly price drop rate in basis points
- minBidIncrementBps: minimum bid increment in basis points
- liquidationIncentiveBps: liquidator bonus in basis points
- maxSlippageBps: maximum slippage in basis points
- updatedBy: admin wallet address
- reason: justification for the change

@returns
- *response.TransactionPreparation: encoded calldata targeting the liquidator contract
- error
*/
func (s *AdminLiquidatorService) PrepareSetLiquidatorParams(
	ctx interface{},
	auctionDurationSecs, priceDropRateBps, minBidIncrementBps, liquidationIncentiveBps, maxSlippageBps int,
	updatedBy, reason string,
) (*response.TransactionPreparation, error) {
	c := extractCtx(ctx)

	row := &postgres.LiquidatorConfigRow{
		AuctionDurationSeconds:  auctionDurationSecs,
		PriceDropRateBps:        priceDropRateBps,
		MinBidIncrementBps:      minBidIncrementBps,
		LiquidationIncentiveBps: liquidationIncentiveBps,
		MaxSlippageBps:          maxSlippageBps,
	}
	if err := s.repo.UpdateLiquidatorConfig(c, row, updatedBy); err != nil {
		return nil, fmt.Errorf("update liquidator config: %w", err)
	}
	_ = s.repo.InsertAuditLog(c, updatedBy, "update_liquidator_params", "liquidator", s.cfg.LiquidatorAddress, "", reason)

	sel := functionSelector("setLiquidatorParams(uint256,uint256,uint256,uint256,uint256)")
	calldata := buildCalldata(sel,
		padUint256(big.NewInt(int64(auctionDurationSecs))),
		padUint256(big.NewInt(int64(priceDropRateBps))),
		padUint256(big.NewInt(int64(minBidIncrementBps))),
		padUint256(big.NewInt(int64(liquidationIncentiveBps))),
		padUint256(big.NewInt(int64(maxSlippageBps))),
	)
	return &response.TransactionPreparation{
		Target:      s.cfg.LiquidatorAddress,
		TxData:      calldata,
		Value:       "0",
		Description: fmt.Sprintf("Update liquidator params: duration=%ds drop=%dbps", auctionDurationSecs, priceDropRateBps),
	}, nil
}

/*
@method GetActiveAuctions

@desc
Returns all currently active liquidation auctions.

@params
- ctx: handler context

@returns
- *response.ActiveAuctionsResponse: list of active auctions with total count
- error
*/
func (s *AdminLiquidatorService) GetActiveAuctions(ctx interface{}) (*response.ActiveAuctionsResponse, error) {
	c := extractCtx(ctx)
	rows, err := s.repo.GetActiveAuctions(c)
	if err != nil {
		return nil, fmt.Errorf("get active auctions: %w", err)
	}

	auctions := make([]response.ActiveAuction, 0, len(rows))
	for _, row := range rows {
		auctions = append(auctions, response.ActiveAuction{
			AuctionID:        row.AuctionID,
			MarketAddress:    row.MarketAddress,
			BorrowerAddress:  row.BorrowerAddress,
			CollateralAmount: row.CollateralAmount,
			DebtAmount:       row.DebtAmount,
			CurrentPrice:     row.CurrentPrice,
			StartTime:        row.StartTime.Unix(),
			EndTime:          row.EndTime.Unix(),
			Status:           row.Status,
		})
	}
	return &response.ActiveAuctionsResponse{
		Auctions: auctions,
		Total:    len(auctions),
	}, nil
}

/*
@method PrepareStopAuction

@desc
Prepares ABI-encoded calldata to cancel an active auction.

@params
- ctx: handler context
- auctionID: numeric auction identifier
- adminAddress: admin wallet address performing the action
- reason: justification for stopping the auction

@returns
- *response.TransactionPreparation: encoded calldata targeting the liquidator contract
- error
*/
func (s *AdminLiquidatorService) PrepareStopAuction(ctx interface{}, auctionID int64, adminAddress, reason string) (*response.TransactionPreparation, error) {
	c := extractCtx(ctx)
	_ = s.repo.InsertAuditLog(c, adminAddress, "stop_auction", "liquidator", s.cfg.LiquidatorAddress, "", reason)

	sel := functionSelector("cancelAuction(uint256)")
	calldata := buildCalldata(sel, padUint256FromInt(auctionID))
	return &response.TransactionPreparation{
		Target:      s.cfg.LiquidatorAddress,
		TxData:      calldata,
		Value:       "0",
		Description: fmt.Sprintf("Cancel auction %d — %s", auctionID, reason),
	}, nil
}
