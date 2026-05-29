package transaction

import (
	"context"
	"math/big"
	"strings"
	"testing"
)

/*
@file service_test.go

@desc
Unit tests for transaction service.
Tests transaction building, gas estimation, and cost calculation.

@test_coverage
- BuildBorrowTransaction: payload construction
- BuildRepayTransaction: repayment payload
- BuildLiquidateTransaction: liquidation payload
- EstimateGasCost: gas estimation
- BuildMulticall: batch transaction building
- GetTransactionQuote: quote generation
*/

// Test BuildBorrowTransaction
func TestBuildBorrowTransaction(t *testing.T) {
	ctx := context.Background()
	service := NewTransactionService()

	tx, err := service.BuildBorrowTransaction(
		ctx,
		"0xmarket123",
		big.NewInt(1000000),
		1,
		500,
	)

	if err != nil {
		t.Fatalf("BuildBorrowTransaction failed: %v", err)
	}

	if tx["to"] != "0xmarket123" {
		t.Errorf("Expected to='0xmarket123', got '%s'", tx["to"])
	}

	if tx["value"] != "0" {
		t.Errorf("Expected value='0', got '%s'", tx["value"])
	}

	data, ok := tx["data"].(string)
	if !ok || !strings.HasPrefix(data, "0x") {
		t.Errorf("Expected valid hex data, got %v", tx["data"])
	}
}

// Test BuildBorrowTransaction - Invalid Amount
func TestBuildBorrowTransactionInvalidAmount(t *testing.T) {
	ctx := context.Background()
	service := NewTransactionService()

	_, err := service.BuildBorrowTransaction(
		ctx,
		"0xmarket123",
		big.NewInt(0),
		1,
		500,
	)

	if err == nil {
		t.Fatal("Expected error for zero amount")
	}
}

// Test BuildRepayTransaction
func TestBuildRepayTransaction(t *testing.T) {
	ctx := context.Background()
	service := NewTransactionService()

	tx, err := service.BuildRepayTransaction(
		ctx,
		"0xmarket123",
		big.NewInt(500000),
	)

	if err != nil {
		t.Fatalf("BuildRepayTransaction failed: %v", err)
	}

	if tx["to"] != "0xmarket123" {
		t.Errorf("Expected to='0xmarket123', got '%s'", tx["to"])
	}

	if tx["value"] != "500000" {
		t.Errorf("Expected value='500000', got '%s'", tx["value"])
	}
}

// Test BuildLiquidateTransaction
func TestBuildLiquidateTransaction(t *testing.T) {
	ctx := context.Background()
	service := NewTransactionService()

	tx, err := service.BuildLiquidateTransaction(
		ctx,
		"0xmarket123",
		1,
		big.NewInt(750000),
	)

	if err != nil {
		t.Fatalf("BuildLiquidateTransaction failed: %v", err)
	}

	if tx["value"] != "750000" {
		t.Errorf("Expected value='750000', got '%s'", tx["value"])
	}
}

// Test EstimateGasCost
func TestEstimateGasCost(t *testing.T) {
	ctx := context.Background()
	service := NewTransactionService()

	estimate, err := service.EstimateGasCost(
		ctx,
		"0xmarket123",
		"0x1234abcd",
		big.NewInt(1000000),
	)

	if err != nil {
		t.Fatalf("EstimateGasCost failed: %v", err)
	}

	if estimate["gas"] == nil {
		t.Error("Expected gas estimate")
	}

	if estimate["gas_price"] == nil {
		t.Error("Expected gas_price estimate")
	}

	if estimate["total_cost"] == nil {
		t.Error("Expected total_cost estimate")
	}
}

// Test BuildMulticall
func TestBuildMulticall(t *testing.T) {
	ctx := context.Background()
	service := NewTransactionService()

	targets := []string{"0xabc", "0xdef"}
	datas := []string{"0x1234", "0x5678"}
	values := []*big.Int{big.NewInt(0), big.NewInt(100000)}

	tx, err := service.BuildMulticall(
		ctx,
		"0xmulticall",
		targets,
		datas,
		values,
	)

	if err != nil {
		t.Fatalf("BuildMulticall failed: %v", err)
	}

	if tx["calls"] != 2 {
		t.Errorf("Expected calls=2, got %v", tx["calls"])
	}

	if tx["value"] != "100000" {
		t.Errorf("Expected value='100000', got '%s'", tx["value"])
	}
}

// Test BuildMulticall - Invalid Input
func TestBuildMulticallInvalidInput(t *testing.T) {
	ctx := context.Background()
	service := NewTransactionService()

	// Mismatched lengths
	targets := []string{"0xabc", "0xdef"}
	datas := []string{"0x1234"} // Too few
	values := []*big.Int{big.NewInt(0), big.NewInt(100000)}

	_, err := service.BuildMulticall(
		ctx,
		"0xmulticall",
		targets,
		datas,
		values,
	)

	if err == nil {
		t.Fatal("Expected error for mismatched lengths")
	}
}

// Test BuildMulticall - Size Limit
func TestBuildMulticallSizeLimit(t *testing.T) {
	ctx := context.Background()
	service := NewTransactionService()

	// Create too many calls
	targets := make([]string, 15)
	datas := make([]string, 15)
	values := make([]*big.Int, 15)
	for i := 0; i < 15; i++ {
		targets[i] = "0xabc"
		datas[i] = "0x1234"
		values[i] = big.NewInt(0)
	}

	_, err := service.BuildMulticall(
		ctx,
		"0xmulticall",
		targets,
		datas,
		values,
	)

	if err == nil {
		t.Fatal("Expected error for batch size exceed")
	}
}

// Test GetTransactionQuote - Borrow
func TestGetTransactionQuoteBorrow(t *testing.T) {
	ctx := context.Background()
	service := NewTransactionService()

	params := map[string]interface{}{
		"amount":  "1000000",
		"tokenID": 1,
		"minAPR":  500,
	}

	quote, err := service.GetTransactionQuote(
		ctx,
		"borrow",
		params,
	)

	if err != nil {
		t.Fatalf("GetTransactionQuote failed: %v", err)
	}

	if quote["tx_type"] != "borrow" {
		t.Errorf("Expected tx_type='borrow', got '%s'", quote["tx_type"])
	}

	if quote["gas"] == nil {
		t.Error("Expected gas estimate")
	}
}

// Test GetTransactionQuote - Repay
func TestGetTransactionQuoteRepay(t *testing.T) {
	ctx := context.Background()
	service := NewTransactionService()

	params := map[string]interface{}{
		"amount": "500000",
	}

	quote, err := service.GetTransactionQuote(
		ctx,
		"repay",
		params,
	)

	if err != nil {
		t.Fatalf("GetTransactionQuote failed: %v", err)
	}

	if quote["tx_type"] != "repay" {
		t.Errorf("Expected tx_type='repay', got '%s'", quote["tx_type"])
	}

	gasEstimate, ok := quote["gas"].(int64)
	if !ok || gasEstimate != 120000 {
		t.Errorf("Expected gas=120000 for repay, got %v", quote["gas"])
	}
}

// Test GetTransactionQuote - Invalid Type
func TestGetTransactionQuoteInvalidType(t *testing.T) {
	ctx := context.Background()
	service := NewTransactionService()

	_, err := service.GetTransactionQuote(
		ctx,
		"invalid",
		map[string]interface{}{},
	)

	if err == nil {
		t.Fatal("Expected error for invalid transaction type")
	}
}

// Test TransactionBuilder - BuildBorrowData
func TestTransactionBuilderBorrowData(t *testing.T) {
	builder := NewTransactionBuilder()

	data := builder.BuildBorrowData(
		"0xmarket",
		big.NewInt(1000000),
		1,
		500,
	)

	if !strings.HasPrefix(data, "0x") {
		t.Errorf("Expected hex string, got %s", data)
	}

	if len(data) < 10 {
		t.Errorf("Expected longer data string, got length %d", len(data))
	}
}

// Test GasCalculator - EstimateGas
func TestGasCalculatorEstimateGas(t *testing.T) {
	calc := NewGasCalculator()

	// Base transaction (32 bytes of data)
	gas := calc.EstimateGas(32)

	if gas < 21000 {
		t.Errorf("Expected gas >= 21000, got %d", gas)
	}

	// Verify data cost is included
	baseFee := calc.EstimateGas(0)
	gasWithData := calc.EstimateGas(100)

	if gasWithData <= baseFee {
		t.Errorf("Gas with data should be higher than base")
	}
}

// Test GasCalculator - SetGasPrice
func TestGasCalculatorSetGasPrice(t *testing.T) {
	calc := NewGasCalculator()

	newPrice := big.NewInt(2000000000) // 2 Gwei
	calc.SetGasPrice(newPrice)

	retrieved := calc.GetGasPrice()
	if retrieved.Cmp(newPrice) != 0 {
		t.Errorf("Expected gas price %v, got %v", newPrice, retrieved)
	}
}

// Test GasCalculator - CalculateGasCost
func TestGasCalculatorCalculateGasCost(t *testing.T) {
	calc := NewGasCalculator()

	gasAmount := int64(100000)
	gasPrice := big.NewInt(1000000000) // 1 Gwei

	cost := calc.CalculateGasCost(gasAmount, gasPrice)

	expected := new(big.Int).Mul(big.NewInt(gasAmount), gasPrice)
	if cost.Cmp(expected) != 0 {
		t.Errorf("Expected cost %v, got %v", expected, cost)
	}
}

// Test GasCalculator - GetPriorityLevel
func TestGasCalculatorGetPriorityLevel(t *testing.T) {
	calc := NewGasCalculator()
	basePrice := calc.GetGasPrice()

	cases := []struct {
		priority   string
		minMulti   float64
		maxMulti   float64
	}{
		{"low", 0.7, 0.9},
		{"standard", 0.9, 1.1},
		{"high", 1.4, 1.6},
		{"urgent", 1.9, 2.1},
	}

	for _, c := range cases {
		price := calc.GetPriorityLevel(c.priority)
		if price.Cmp(big.NewInt(0)) <= 0 {
			t.Errorf("Expected positive price for %s", c.priority)
		}

		ratio := new(big.Rat).SetInt(price)
		ratio.Quo(ratio, new(big.Rat).SetInt(basePrice))
		f, _ := ratio.Float64()

		if f < c.minMulti || f > c.maxMulti {
			t.Errorf("Priority %s multiplier %.2f not in range [%.2f, %.2f]", c.priority, f, c.minMulti, c.maxMulti)
		}
	}
}

// Test GasCalculator - OptimizeGasForBatch
func TestGasCalculatorOptimizeGasForBatch(t *testing.T) {
	calc := NewGasCalculator()

	singleGas := calc.OptimizeGasForBatch(1, 100)
	batchGas := calc.OptimizeGasForBatch(5, 100)

	if batchGas <= singleGas {
		t.Errorf("Batch gas (%d) should be > single (%d)", batchGas, singleGas)
	}
}
