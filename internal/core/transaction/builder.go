package transaction

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

/*
@file builder.go

@desc
Transaction data builder for RevvFi smart contracts.
Constructs function call payloads for blockchain interactions.

@responsibilities
- Build borrow transaction data
- Build repayment transaction data
- Build liquidation transaction data
- Build multi-call batch data
*/

/*
@struct TransactionBuilder

@desc
Builds transaction payloads for smart contract calls.
*/
type TransactionBuilder struct{}

/*
@function NewTransactionBuilder

@desc
Creates new transaction builder.

@returns
- *TransactionBuilder
*/
func NewTransactionBuilder() *TransactionBuilder {
	return &TransactionBuilder{}
}

/*
@function BuildBorrowData

@desc
Builds borrow function call data.

@params
- market: market address
- amount: borrow amount
- tokenID: collateral token ID
- minAPR: minimum APR

@returns
- string: encoded function call data
*/
func (b *TransactionBuilder) BuildBorrowData(
	market string,
	amount *big.Int,
	tokenID int64,
	minAPR int32,
) string {
	// borrow(uint256 amount, uint256 tokenID, uint32 minAPR)
	// Function selector: 0x1234abcd (placeholder)
	selector := "1234abcd"

	// Encode parameters (simplified - real implementation would use proper ABI encoding)
	// amount (32 bytes)
	amountHex := fmt.Sprintf("%064x", amount)
	// tokenID (32 bytes)
	tokenIDHex := fmt.Sprintf("%064x", tokenID)
	// minAPR (32 bytes)
	minAPRHex := fmt.Sprintf("%064x", minAPR)

	data := "0x" + selector + amountHex + tokenIDHex + minAPRHex
	return data
}

/*
@function BuildRepayData

@desc
Builds repay function call data.

@params
- market: market address
- amount: repayment amount

@returns
- string: encoded function call data
*/
func (b *TransactionBuilder) BuildRepayData(
	market string,
	amount *big.Int,
) string {
	// repay(uint256 amount)
	// Function selector: 0x5678ef01 (placeholder)
	selector := "5678ef01"

	// Encode amount
	amountHex := fmt.Sprintf("%064x", amount)

	data := "0x" + selector + amountHex
	return data
}

/*
@function BuildLiquidateData

@desc
Builds liquidation bid function call data.

@params
- market: market address
- auctionID: auction ID
- bidAmount: bid amount

@returns
- string: encoded function call data
*/
func (b *TransactionBuilder) BuildLiquidateData(
	market string,
	auctionID int64,
	bidAmount *big.Int,
) string {
	// bid(uint256 auctionID, uint256 amount)
	// Function selector: 0x90b6c5f9 (placeholder)
	selector := "90b6c5f9"

	// Encode parameters
	auctionIDHex := fmt.Sprintf("%064x", auctionID)
	bidAmountHex := fmt.Sprintf("%064x", bidAmount)

	data := "0x" + selector + auctionIDHex + bidAmountHex
	return data
}

/*
@function BuildMulticallData

@desc
Builds multicall aggregate function data.

@params
- targets: list of target addresses
- datas: list of function call datas
- values: list of wei values

@returns
- string: encoded multicall data
*/
func (b *TransactionBuilder) BuildMulticallData(
	targets []string,
	datas []string,
	values []*big.Int,
) string {
	// aggregate(address[] targets, bytes[] datas, uint256[] values)
	// Function selector: 0xffff00ff (placeholder)
	selector := "ffff00ff"

	// Build packed data (simplified)
	var packed string

	// Count
	countHex := fmt.Sprintf("%064x", len(targets))
	packed += countHex

	// Targets
	for _, target := range targets {
		targetAddr := strings.TrimPrefix(target, "0x")
		targetHex := fmt.Sprintf("%064s", targetAddr) // pad to 64 chars
		packed += targetHex
	}

	// Data lengths and data
	for _, data := range datas {
		dataClean := strings.TrimPrefix(data, "0x")
		dataLen := len(dataClean) / 2 // convert hex chars to bytes
		lenHex := fmt.Sprintf("%064x", dataLen)
		packed += lenHex + dataClean
	}

	// Values
	for _, value := range values {
		if value == nil {
			value = big.NewInt(0)
		}
		valueHex := fmt.Sprintf("%064x", value)
		packed += valueHex
	}

	data := "0x" + selector + packed
	return data
}

/*
@function EncodeUint256

@desc
Encodes uint256 to 32-byte hex string.

@params
- value: value to encode

@returns
- string: hex-encoded value
*/
func (b *TransactionBuilder) EncodeUint256(value *big.Int) string {
	return fmt.Sprintf("%064x", value)
}

/*
@function EncodeAddress

@desc
Encodes address to 32-byte hex string.

@params
- address: address to encode

@returns
- string: hex-encoded address
*/
func (b *TransactionBuilder) EncodeAddress(address string) string {
	addr := strings.TrimPrefix(address, "0x")
	// Pad address to 64 chars (32 bytes), left-aligned
	return fmt.Sprintf("%064s", addr)
}

/*
@function EncodeBytes

@desc
Encodes arbitrary bytes.

@params
- data: data to encode

@returns
- string: hex-encoded data
*/
func (b *TransactionBuilder) EncodeBytes(data []byte) string {
	return "0x" + hex.EncodeToString(data)
}

/*
@function DecodeUint256

@desc
Decodes hex string to uint256.

@params
- hexStr: hex string to decode

@returns
- *big.Int: decoded value
- error: if decoding fails
*/
func (b *TransactionBuilder) DecodeUint256(hexStr string) (*big.Int, error) {
	hexStr = strings.TrimPrefix(hexStr, "0x")
	value := new(big.Int)
	_, err := fmt.Sscanf(hexStr, "%x", value)
	return value, err
}
