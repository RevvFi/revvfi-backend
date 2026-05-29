package offer

import (
"math/big"

"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@file calculator.go

@desc
Offer-specific calculations for APR metrics and quote generation.

@responsibilities
- Calculate weighted APR
- Calculate effective borrowing rate
- Estimate interest costs
*/

/*
@struct Calculator

@desc
Performs offer-related calculations.
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
@function CalculateWeightedAPR

@desc
Calculates weighted average APR across multiple offers.

@params
- offers: offers to average

@returns
- int32: weighted average APR in bps
*/
func (c *Calculator) CalculateWeightedAPR(offers []models.Offer) int32 {
	if len(offers) == 0 {
		return 0
	}

	totalWeighted := big.NewInt(0)
	totalLiquidity := big.NewInt(0)

	for _, offer := range offers {
		weight := new(big.Int).Mul(offer.RemainingAmount, big.NewInt(int64(offer.APR)))
		totalWeighted.Add(totalWeighted, weight)
		totalLiquidity.Add(totalLiquidity, offer.RemainingAmount)
	}

	if totalLiquidity.Sign() == 0 {
		return 0
	}

	avgAPR := new(big.Int).Div(totalWeighted, totalLiquidity)
	return int32(avgAPR.Int64())
}

/*
@function CalculateEstimatedInterest

@desc
Estimates total interest cost for borrow amount.

@params
- amount: borrow amount in wei
- apr: annual percentage rate in bps
- secondsDuration: loan duration in seconds

@returns
- *big.Int: estimated interest
*/
func (c *Calculator) CalculateEstimatedInterest(
amount *big.Int,
apr int32,
secondsDuration int64,
) *big.Int {
	const (
secondsPerYear = 365 * 24 * 60 * 60
bpsScale       = 10000
)

	// Interest = Amount × APR × Seconds / (365 days × 10000 bps)
	amountBig := new(big.Int).Set(amount)
	aprBig := big.NewInt(int64(apr))
	secondsBig := big.NewInt(secondsDuration)

	numerator := new(big.Int).Mul(amountBig, aprBig)
	numerator.Mul(numerator, secondsBig)

	denominator := big.NewInt(int64(secondsPerYear * bpsScale))

	interest := new(big.Int).Div(numerator, denominator)
	return interest
}

/*
@function CalculateEffectiveRate

@desc
Calculates effective borrowing rate including all offers.

@params
- offers: offers to calculate rate from
- borrowAmount: total borrow amount

@returns
- float64: effective rate (0-1 scale, so 0.05 = 5%)
*/
func (c *Calculator) CalculateEffectiveRate(
offers []models.Offer,
borrowAmount *big.Int,
) float64 {
	weightedAPR := c.CalculateWeightedAPR(offers)
	// Convert from bps to decimal (5000 bps = 0.5)
	return float64(weightedAPR) / 10000
}
