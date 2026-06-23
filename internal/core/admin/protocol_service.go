package admin

import (
	"fmt"
	"math/big"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	"github.com/Revvfi/revvfi-backend/internal/config"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@file protocol_service.go

@desc
Service implementing AdminProtocolServiceInterface.
Returns protocol configuration, fee data, contract status, and timelock information.
Prepares ABI-encoded calldata for governance transactions.

@responsibilities
- Return current protocol configuration from blockchain config
- Prepare calldata for fee changes, recipient changes, fee withdrawals
- Check core contract initialization status
- Prepare calldata for core contract setup
- List and prepare upgrade transactions
- Return timelock status
*/

/*
@struct AdminProtocolService

@desc
Implements admin protocol configuration management.

@fields
- cfg: blockchain configuration with all contract addresses
*/
type AdminProtocolService struct {
	cfg config.BlockchainConfig
}

/*
@function NewAdminProtocolService

@desc
Creates a new AdminProtocolService.

@params
- cfg: blockchain configuration

@returns
- *AdminProtocolService
*/
func NewAdminProtocolService(cfg config.BlockchainConfig) *AdminProtocolService {
	return &AdminProtocolService{cfg: cfg}
}

/*
@method GetProtocolConfig

@desc
Returns the current protocol configuration from the blockchain config.

@params
- ctx: handler context

@returns
- *response.ProtocolConfig: current protocol configuration
- error
*/
func (s *AdminProtocolService) GetProtocolConfig(_ interface{}) (*response.ProtocolConfig, error) {
	return &response.ProtocolConfig{
		DeploymentFeeEth:   "0.1",
		DeploymentFeeWei:   "100000000000000000",
		FeeRecipient:       s.cfg.ArchControllerAddress,
		ArchController:     s.cfg.ArchControllerAddress,
		PositionNFT:        s.cfg.PositionNFTAddress,
		Liquidator:         s.cfg.LiquidatorAddress,
		ReputationRegistry: s.cfg.ReputationRegistryAddress,
		TimelockDelayDays:  2,
		IsCoreContractsSet: s.cfg.LiquidatorAddress != "" && s.cfg.PositionNFTAddress != "",
		LastUpdatedAt:      time.Now().Unix(),
	}, nil
}

/*
@method PrepareFeeChange

@desc
Prepares ABI-encoded calldata to update the protocol deployment fee.

@params
- ctx: handler context
- feeWei: new fee in wei as decimal string

@returns
- *response.TransactionPreparation: encoded calldata
- error
*/
func (s *AdminProtocolService) PrepareFeeChange(_ interface{}, feeWei string) (*response.TransactionPreparation, error) {
	if feeWei == "" {
		return nil, appErr.ErrInvalidAmount
	}
	val, ok := new(big.Int).SetString(feeWei, 10)
	if !ok {
		return nil, appErr.ErrInvalidAmount
	}
	sel := functionSelector("setDeploymentFee(uint256)")
	calldata := buildCalldata(sel, padUint256(val))
	return &response.TransactionPreparation{
		Target:      s.cfg.ArchControllerAddress,
		TxData:      calldata,
		Value:       "0",
		Description: fmt.Sprintf("Set deployment fee to %s wei", feeWei),
	}, nil
}

/*
@method PrepareFeeRecipientChange

@desc
Prepares ABI-encoded calldata to update the fee recipient address.

@params
- ctx: handler context
- recipientAddress: new fee recipient address
- reason: justification for the change

@returns
- *response.TransactionPreparation: encoded calldata
- error
*/
func (s *AdminProtocolService) PrepareFeeRecipientChange(_ interface{}, recipientAddress, reason string) (*response.TransactionPreparation, error) {
	if recipientAddress == "" {
		return nil, appErr.ErrInvalidAddress
	}
	_ = reason
	sel := functionSelector("setFeeRecipient(address)")
	calldata := buildCalldata(sel, padAddress(recipientAddress))
	return &response.TransactionPreparation{
		Target:      s.cfg.ArchControllerAddress,
		TxData:      calldata,
		Value:       "0",
		Description: fmt.Sprintf("Set fee recipient to %s", recipientAddress),
	}, nil
}

/*
@method GetFeesCollected

@desc
Returns accumulated protocol fees. Values are static placeholders until
an on-chain indexer tracks fee collection events.

@params
- ctx: handler context

@returns
- *response.FeeCollectedResponse: fee summary
- error
*/
func (s *AdminProtocolService) GetFeesCollected(_ interface{}) (*response.FeeCollectedResponse, error) {
	return &response.FeeCollectedResponse{
		TotalFeesEth:          "0",
		TotalFeesWei:          "0",
		RecipientAddress:      s.cfg.ArchControllerAddress,
		CollectionPeriodStart: 0,
		CollectionPeriodEnd:   time.Now().Unix(),
	}, nil
}

/*
@method CheckCoreContractsStatus

@desc
Returns the initialization status of core protocol contracts.

@params
- ctx: handler context

@returns
- *response.ContractStatus: contract addresses and initialization state
- error
*/
func (s *AdminProtocolService) CheckCoreContractsStatus(_ interface{}) (*response.ContractStatus, error) {
	isSet := s.cfg.LiquidatorAddress != "" && s.cfg.PositionNFTAddress != "" && s.cfg.ReputationRegistryAddress != ""
	return &response.ContractStatus{
		IsCoreContractsSet:        isSet,
		PositionNFTAddress:        s.cfg.PositionNFTAddress,
		LiquidatorAddress:         s.cfg.LiquidatorAddress,
		ReputationRegistryAddress: s.cfg.ReputationRegistryAddress,
		DeploymentBlock:           0,
		DeploymentTimestamp:       0,
	}, nil
}

/*
@method PrepareCoreContractsSet

@desc
Prepares ABI-encoded calldata to set the core contract addresses in the ArchController.

@params
- ctx: handler context
- positionNFT: position NFT contract address
- liquidator: liquidator contract address
- reputationRegistry: reputation registry contract address

@returns
- *response.TransactionPreparation: encoded calldata
- error
*/
func (s *AdminProtocolService) PrepareCoreContractsSet(_ interface{}, positionNFT, liquidator, reputationRegistry string) (*response.TransactionPreparation, error) {
	if positionNFT == "" || liquidator == "" || reputationRegistry == "" {
		return nil, appErr.ErrInvalidAddress
	}
	sel := functionSelector("setCoreContracts(address,address,address)")
	calldata := buildCalldata(sel, padAddress(positionNFT), padAddress(liquidator), padAddress(reputationRegistry))
	return &response.TransactionPreparation{
		Target:      s.cfg.ArchControllerAddress,
		TxData:      calldata,
		Value:       "0",
		Description: "Set core contracts: positionNFT, liquidator, reputationRegistry",
	}, nil
}

/*
@method ListPendingUpgrades

@desc
Returns pending timelock upgrades. Currently returns an empty list as upgrade
tracking is not yet stored in the database.

@params
- ctx: handler context
- page: page number
- limit: entries per page

@returns
- []response.PendingUpgrade: pending upgrades
- *response.PaginationInfo: pagination metadata
- error
*/
func (s *AdminProtocolService) ListPendingUpgrades(_ interface{}, page, limit int) ([]response.PendingUpgrade, *response.PaginationInfo, error) {
	return []response.PendingUpgrade{}, &response.PaginationInfo{
		Page:       int32(page),
		Limit:      int32(limit),
		Total:      0,
		TotalPages: 0,
	}, nil
}

/*
@method PrepareMarketUpgrade

@desc
Prepares ABI-encoded calldata to queue a market implementation upgrade.

@params
- ctx: handler context
- marketImpl: new market implementation contract address
- reason: justification for the upgrade

@returns
- *response.TransactionPreparation: encoded calldata
- error
*/
func (s *AdminProtocolService) PrepareMarketUpgrade(_ interface{}, marketImpl, reason string) (*response.TransactionPreparation, error) {
	if marketImpl == "" {
		return nil, appErr.ErrInvalidAddress
	}
	_ = reason
	sel := functionSelector("queueMarketUpgrade(address)")
	calldata := buildCalldata(sel, padAddress(marketImpl))
	return &response.TransactionPreparation{
		Target:      s.cfg.FactoryAddress,
		TxData:      calldata,
		Value:       "0",
		Description: fmt.Sprintf("Queue market implementation upgrade to %s (2-day timelock)", marketImpl),
	}, nil
}

/*
@method GetTimelockStatus

@desc
Returns the current timelock configuration.

@params
- ctx: handler context

@returns
- *response.TimelockStatus: timelock delay settings
- error
*/
func (s *AdminProtocolService) GetTimelockStatus(_ interface{}) (*response.TimelockStatus, error) {
	return &response.TimelockStatus{
		DelaySeconds: 172800, // 2 days
		DelayDays:    2,
		MinDelay:     86400,  // 1 day
		MaxDelay:     604800, // 7 days
	}, nil
}

/*
@method WithdrawFees

@desc
Prepares ABI-encoded calldata to withdraw accumulated protocol fees.

@params
- ctx: handler context
- amount: withdrawal amount in wei (decimal string)
- recipient: address to receive the withdrawn fees

@returns
- *response.TransactionPreparation: encoded calldata
- error
*/
func (s *AdminProtocolService) WithdrawFees(_ interface{}, amount, recipient string) (*response.TransactionPreparation, error) {
	if recipient == "" {
		return nil, appErr.ErrInvalidAddress
	}
	val := big.NewInt(0)
	if amount != "" {
		if v, ok := new(big.Int).SetString(amount, 10); ok {
			val = v
		}
	}
	sel := functionSelector("withdrawFees(address,uint256)")
	calldata := buildCalldata(sel, padAddress(recipient), padUint256(val))
	return &response.TransactionPreparation{
		Target:      s.cfg.ArchControllerAddress,
		TxData:      calldata,
		Value:       "0",
		Description: fmt.Sprintf("Withdraw %s wei fees to %s", amount, recipient),
	}, nil
}

/*
@method CancelUpgrade

@desc
Prepares ABI-encoded calldata to cancel a queued upgrade.

@params
- ctx: handler context
- upgradeId: upgrade identifier string
- reason: justification for cancellation

@returns
- *response.TransactionPreparation: encoded calldata
- error
*/
func (s *AdminProtocolService) CancelUpgrade(_ interface{}, upgradeId, reason string) (*response.TransactionPreparation, error) {
	if upgradeId == "" {
		return nil, appErr.ErrInvalidInput
	}
	_ = reason
	sel := functionSelector("cancelUpgrade(bytes32)")
	// Encode upgradeId as bytes32 (right-padded if shorter)
	idBytes := []byte(upgradeId)
	slot := make([]byte, 32)
	copy(slot, idBytes)
	calldata := buildCalldata(sel, slot)
	return &response.TransactionPreparation{
		Target:      s.cfg.FactoryAddress,
		TxData:      calldata,
		Value:       "0",
		Description: fmt.Sprintf("Cancel upgrade %s — %s", upgradeId, reason),
	}, nil
}

/*
@method GetUpgradeQueue

@desc
Returns the current upgrade queue. Returns empty until an upgrade indexer is implemented.

@params
- ctx: handler context

@returns
- []response.UpgradeQueueItem: queued upgrades
- error
*/
func (s *AdminProtocolService) GetUpgradeQueue(_ interface{}) ([]response.UpgradeQueueItem, error) {
	return []response.UpgradeQueueItem{}, nil
}
