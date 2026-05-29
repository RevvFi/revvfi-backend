package transaction

import (
	"context"
	"fmt"
	"math/big"

	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@file service.go

@desc
Transaction service manages blockchain transaction construction and estimation.
Builds complex multi-call transactions for RevvFi operations.

@responsibilities
- Generate transaction quotes
- Build transaction payloads
- Estimate gas costs
- Optimize multi-call batching
*/

/*
@struct TransactionService

@desc
Manages blockchain transaction building and gas estimation.

@dependencies
- TransactionBuilder: transaction payload construction
- GasCalculator: gas cost estimation
*/
type TransactionService struct {
	builder *TransactionBuilder
	gas     *GasCalculator
}

/*
@function NewTransactionService

@desc
Creates new transaction service.

@returns
- *TransactionService
*/
func NewTransactionService() *TransactionService {
	return &TransactionService{
		builder: NewTransactionBuilder(),
		gas:     NewGasCalculator(),
	}
}

/*
@function BuildBorrowTransaction

@desc
Builds borrow transaction payload.

@params
- ctx: request context
- market: market address
- amount: borrow amount in wei
- tokenID: collateral token ID
- minAPR: minimum acceptable APR

@returns
- map[string]interface{}: transaction data
- error: if building fails
*/
func (s *TransactionService) BuildBorrowTransaction(
	ctx context.Context,
	market string,
	amount *big.Int,
	tokenID int64,
	minAPR int32,
) (map[string]interface{}, error) {
	// Validate parameters
	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, appErr.ErrInvalidAmount
	}

	data := s.builder.BuildBorrowData(market, amount, tokenID, minAPR)

	return map[string]interface{}{
		"to":    market,
		"data":  data,
		"value": "0",
	}, nil
}

/*
@function BuildRepayTransaction

@desc
Builds repayment transaction payload.

@params
- ctx: request context
- market: market address
- amount: repayment amount in wei

@returns
- map[string]interface{}: transaction data
- error: if building fails
*/
func (s *TransactionService) BuildRepayTransaction(
	ctx context.Context,
	market string,
	amount *big.Int,
) (map[string]interface{}, error) {
	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, appErr.ErrInvalidAmount
	}

	data := s.builder.BuildRepayData(market, amount)

	return map[string]interface{}{
		"to":    market,
		"data":  data,
		"value": amount.String(),
	}, nil
}

/*
@function BuildLiquidateTransaction

@desc
Builds liquidation transaction payload.

@params
- ctx: request context
- market: market address
- auctionID: auction ID
- bidAmount: bid amount in wei

@returns
- map[string]interface{}: transaction data
- error: if building fails
*/
func (s *TransactionService) BuildLiquidateTransaction(
	ctx context.Context,
	market string,
	auctionID int64,
	bidAmount *big.Int,
) (map[string]interface{}, error) {
	if bidAmount == nil || bidAmount.Cmp(big.NewInt(0)) <= 0 {
		return nil, appErr.ErrInvalidAmount
	}

	data := s.builder.BuildLiquidateData(market, auctionID, bidAmount)

	return map[string]interface{}{
		"to":    market,
		"data":  data,
		"value": bidAmount.String(),
	}, nil
}

/*
@function EstimateGasCost

@desc
Estimates gas cost for transaction.

@params
- ctx: request context
- to: recipient address
- data: transaction data
- value: transaction value

@returns
- map[string]interface{}: gas estimate data
- error: if estimation fails
*/
func (s *TransactionService) EstimateGasCost(
	ctx context.Context,
	to string,
	data string,
	value *big.Int,
) (map[string]interface{}, error) {
	gasEstimate := s.gas.EstimateGas(len(data))
	gasPrice := s.gas.GetGasPrice()

	gasCost := new(big.Int).Mul(big.NewInt(gasEstimate), gasPrice)

	return map[string]interface{}{
		"gas":       gasEstimate,
		"gas_price": gasPrice.String(),
		"total_cost": gasCost.String(),
	}, nil
}

/*
@function BuildMulticall

@desc
Builds multi-call transaction batching multiple operations.

@params
- ctx: request context
- multicallAddress: multicall contract address
- targets: list of target addresses
- datas: list of transaction data
- values: list of values per call

@returns
- map[string]interface{}: multicall transaction data
- error: if building fails
*/
func (s *TransactionService) BuildMulticall(
	ctx context.Context,
	multicallAddress string,
	targets []string,
	datas []string,
	values []*big.Int,
) (map[string]interface{}, error) {
	if len(targets) == 0 || len(targets) != len(datas) {
		return nil, appErr.ErrInvalidInput
	}

	if len(targets) > 10 {
		return nil, fmt.Errorf("multicall batch size exceeds limit: %w", appErr.ErrInvalidInput)
	}

	data := s.builder.BuildMulticallData(targets, datas, values)

	totalValue := big.NewInt(0)
	for _, v := range values {
		if v != nil {
			totalValue.Add(totalValue, v)
		}
	}

	return map[string]interface{}{
		"to":       multicallAddress,
		"data":     data,
		"value":    totalValue.String(),
		"calls":    len(targets),
	}, nil
}

/*
@function GetTransactionQuote

@desc
Gets estimated transaction costs and details.

@params
- ctx: request context
- txType: transaction type (borrow, repay, liquidate, etc)
- params: transaction parameters

@returns
- map[string]interface{}: quote details
- error: if quote generation fails
*/
func (s *TransactionService) GetTransactionQuote(
	ctx context.Context,
	txType string,
	params map[string]interface{},
) (map[string]interface{}, error) {
	var gasEstimate int64

	switch txType {
	case "borrow":
		gasEstimate = 150000
	case "repay":
		gasEstimate = 120000
	case "liquidate":
		gasEstimate = 180000
	case "withdraw":
		gasEstimate = 100000
	default:
		return nil, fmt.Errorf("unknown transaction type: %w", appErr.ErrInvalidInput)
	}

	gasPrice := s.gas.GetGasPrice()
	gasCost := new(big.Int).Mul(big.NewInt(gasEstimate), gasPrice)

	return map[string]interface{}{
		"tx_type":   txType,
		"gas":       gasEstimate,
		"gas_price": gasPrice.String(),
		"total_cost": gasCost.String(),
		"params":    params,
	}, nil
}
