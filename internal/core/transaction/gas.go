package transaction

import (
	"math/big"
)

/*
@file gas.go

@desc
Gas cost calculation for transactions.
Estimates gas consumption and pricing for blockchain operations.

@responsibilities
- Estimate gas consumption
- Calculate gas prices
- Optimize gas usage
- Estimate total costs
*/

/*
@struct GasCalculator

@desc
Calculates gas costs for transactions.
*/
type GasCalculator struct {
	baseGasPrice *big.Int
	multiplier   float64
}

/*
@function NewGasCalculator

@desc
Creates new gas calculator.

@returns
- *GasCalculator
*/
func NewGasCalculator() *GasCalculator {
	return &GasCalculator{
		baseGasPrice: big.NewInt(1000000000), // 1 Gwei default
		multiplier:   1.0,
	}
}

/*
@function EstimateGas

@desc
Estimates gas consumption for transaction data.

@params
- dataLength: length of transaction data in bytes

@returns
- int64: estimated gas amount
*/
func (g *GasCalculator) EstimateGas(dataLength int) int64 {
	// Base cost for transaction
	baseCost := int64(21000)

	// Cost per byte of data (4 gas for 0 bytes, 16 for non-zero)
	// Average 8 gas per byte
	dataCost := int64(dataLength) * 8

	return baseCost + dataCost
}

/*
@function GetGasPrice

@desc
Gets current gas price.

@returns
- *big.Int: gas price in wei
*/
func (g *GasCalculator) GetGasPrice() *big.Int {
	return new(big.Int).Set(g.baseGasPrice)
}

/*
@function SetGasPrice

@desc
Sets gas price.

@params
- gasPrice: new gas price in wei
*/
func (g *GasCalculator) SetGasPrice(gasPrice *big.Int) {
	g.baseGasPrice = new(big.Int).Set(gasPrice)
}

/*
@function SetMultiplier

@desc
Sets gas price multiplier for priority.

@params
- multiplier: price multiplier (1.0 = normal, 2.0 = double)
*/
func (g *GasCalculator) SetMultiplier(multiplier float64) {
	g.multiplier = multiplier
}

/*
@function CalculateGasCost

@desc
Calculates total gas cost.

@params
- gasAmount: amount of gas
- gasPrice: price per gas unit

@returns
- *big.Int: total cost in wei
*/
func (g *GasCalculator) CalculateGasCost(gasAmount int64, gasPrice *big.Int) *big.Int {
	cost := new(big.Int).Mul(big.NewInt(gasAmount), gasPrice)
	return cost
}

/*
@function EstimateTransactionCost

@desc
Estimates total transaction cost.

@params
- dataLength: transaction data length
- gasPrice: gas price (optional, uses default if nil)

@returns
- *big.Int: estimated cost in wei
*/
func (g *GasCalculator) EstimateTransactionCost(dataLength int, gasPrice *big.Int) *big.Int {
	gas := g.EstimateGas(dataLength)

	if gasPrice == nil {
		gasPrice = g.GetGasPrice()
	}

	return g.CalculateGasCost(gas, gasPrice)
}

/*
@function CalculateL2GasCost

@desc
Calculates gas cost with L2 fee (for networks like Arbitrum/Optimism).

@params
- txData: transaction data
- baseGasCost: base layer gas cost
- l2FeeScalar: L2 fee scalar

@returns
- *big.Int: total cost including L2 fee
*/
func (g *GasCalculator) CalculateL2GasCost(
	txData string,
	baseGasCost *big.Int,
	l2FeeScalar float64,
) *big.Int {
	// L2 fee = dataLength * 16 * gasPrice * scalar
	dataLength := len(txData) / 2 // approx byte length
	dataGas := big.NewInt(int64(dataLength * 16))

	l2Fee := new(big.Int).Mul(dataGas, g.GetGasPrice())
	l2FeeInt := new(big.Int).Mul(l2Fee, big.NewInt(int64(l2FeeScalar*1000)))
	l2FeeInt.Div(l2FeeInt, big.NewInt(1000))

	totalCost := new(big.Int).Add(baseGasCost, l2FeeInt)
	return totalCost
}

/*
@function GetPriorityLevel

@desc
Gets gas price for priority level.

@params
- priority: "low" | "standard" | "high" | "urgent"

@returns
- *big.Int: gas price in wei
*/
func (g *GasCalculator) GetPriorityLevel(priority string) *big.Int {
	basePrice := g.GetGasPrice()

	var multiplier float64
	switch priority {
	case "low":
		multiplier = 0.8
	case "standard":
		multiplier = 1.0
	case "high":
		multiplier = 1.5
	case "urgent":
		multiplier = 2.0
	default:
		multiplier = 1.0
	}

	result := new(big.Int).Mul(basePrice, big.NewInt(int64(multiplier*1000)))
	result.Div(result, big.NewInt(1000))
	return result
}

/*
@function OptimizeGasForBatch

@desc
Optimizes gas estimation for batch transactions.

@params
- txCount: number of transactions
- averageDataLength: average data length per tx

@returns
- int64: optimized total gas estimate
*/
func (g *GasCalculator) OptimizeGasForBatch(txCount int, averageDataLength int) int64 {
	if txCount <= 1 {
		return g.EstimateGas(averageDataLength)
	}

	// Base cost shared across batch
	baseCost := int64(21000)

	// Per-transaction overhead (multicall doesn't repeat base cost)
	txOverhead := int64(5000) * int64(txCount)

	// Data cost
	dataCost := int64(averageDataLength*txCount) * 8

	// Multicall overhead
	multicallOverhead := int64(10000)

	totalGas := baseCost + txOverhead + dataCost + multicallOverhead
	return totalGas
}
