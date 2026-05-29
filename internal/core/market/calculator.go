package market

import (
"math"
"math/big"

"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@file calculator.go

@desc
Market calculation engine for APR, utilization, and collateral ratios.
Performs all mathematical operations for market metrics.

@responsibilities
- Calculate utilization rates
- Calculate average APRs
- Calculate collateral ratios
- Calculate interest accrual
- Handle precision math (bps conversions)
*/

/*
@struct Calculator

@desc
Performs market financial calculations.
*/
type Calculator struct{}

/*
@function NewCalculator

@desc
Creates new calculator instance.

@returns
- *Calculator
*/
func NewCalculator() *Calculator {
	return &Calculator{}
}

/*
@function CalculateUtilization

@desc
Calculates market utilization rate as percentage.
Utilization = Total Debt / Total Liquidity * 100

@params
- totalDebt: total outstanding debt
- totalLiquidity: total available liquidity

@returns
- float64: utilization percentage (0-100)
*/
func (c *Calculator) CalculateUtilization(totalDebt, totalLiquidity *big.Int) float64 {
	if totalLiquidity == nil || totalLiquidity.Sign() == 0 {
		return 0
	}

	// Convert to float with precision
	debtFloat := new(big.Float).SetInt(totalDebt)
	liquidityFloat := new(big.Float).SetInt(totalLiquidity)

	ratio := new(big.Float).Quo(debtFloat, liquidityFloat)
	ratio.Mul(ratio, big.NewFloat(100))

	result, _ := ratio.Float64()
	return math.Min(result, 100) // Cap at 100%
}

/*
@function CalculateAverageAPR

@desc
Calculates weighted average APR across active offers.

@params
- offers: active offers

@returns
- int32: weighted average APR in bps
*/
func (c *Calculator) CalculateAverageAPR(offers []models.Offer) int32 {
	if len(offers) == 0 {
		return 0
	}

	totalWeighted := big.NewInt(0)
	totalAmount := big.NewInt(0)

	for _, offer := range offers {
		weight := new(big.Int).Mul(offer.RemainingAmount, big.NewInt(int64(offer.APR)))
		totalWeighted.Add(totalWeighted, weight)
		totalAmount.Add(totalAmount, offer.RemainingAmount)
	}

	if totalAmount.Sign() == 0 {
		return 0
	}

	avgAPR := new(big.Int).Div(totalWeighted, totalAmount)
	return int32(avgAPR.Int64())
}

/*
@function CalculateCollateralRatio

@desc
Calculates current collateral ratio.
Ratio = Collateral Value / Total Debt

@params
- collateralValue: current collateral value in borrow asset
- totalDebt: total outstanding debt

@returns
- float64: collateral ratio (1.0 = 100%)
*/
func (c *Calculator) CalculateCollateralRatio(
collateralValue, totalDebt *big.Int,
) float64 {
	if totalDebt == nil || totalDebt.Sign() == 0 {
		return math.Inf(1) // Infinite ratio if no debt
	}

	collateralFloat := new(big.Float).SetInt(collateralValue)
	debtFloat := new(big.Float).SetInt(totalDebt)

	ratio := new(big.Float).Quo(collateralFloat, debtFloat)
	result, _ := ratio.Float64()
	return result
}

/*
@function CalculateInterestAccrual

@desc
Calculates accrued interest for a position.
Interest = Principal × APR (bps) × Time Elapsed (seconds) / (365 days × 10000 bps)

@params
- principal: original principal amount
- apr: annual percentage rate in bps
- secondsElapsed: seconds since accrual started

@returns
- *big.Int: accrued interest amount
*/
func (c *Calculator) CalculateInterestAccrual(
principal *big.Int,
apr int32,
secondsElapsed int64,
) *big.Int {
	// Interest = Principal × APR × SecondsElapsed / (365 days × 10000 bps)
	const (
secondsPerYear = 365 * 24 * 60 * 60 // 31,536,000
bpsScale       = 10000
)

	// Convert to big.Int for precise calculation
	principalBig := new(big.Int).Set(principal)
	aprBig := big.NewInt(int64(apr))
	secondsBig := big.NewInt(secondsElapsed)

	// numerator = principal × apr × secondsElapsed
	numerator := new(big.Int).Mul(principalBig, aprBig)
	numerator.Mul(numerator, secondsBig)

	// denominator = 365 days × 10000 bps
	denominator := big.NewInt(int64(secondsPerYear * bpsScale))

	// interest = numerator / denominator
	interest := new(big.Int).Div(numerator, denominator)

	return interest
}

/*
@function IsLiquidatable

@desc
Determines if position is liquidatable based on collateral ratio.

@params
- collateralRatio: current collateral ratio
- liquidationThreshold: threshold in bps (9500 = 95% = 0.95 ratio)

@returns
- bool: true if liquidatable
*/
func (c *Calculator) IsLiquidatable(
collateralRatio float64,
liquidationThreshold int32,
) bool {
	thresholdRatio := float64(liquidationThreshold) / 10000
	return collateralRatio < thresholdRatio
}
