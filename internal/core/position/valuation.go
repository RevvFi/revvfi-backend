package position

import (
"math/big"

"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@file valuation.go

@desc
Handles position valuation and interest calculations.

@responsibilities
- Calculate accrued interest
- Determine position value
- Compute portfolio APR
*/

const (
secondsPerYear = 31536000
bpsScale       = 10000
)

/*
@struct Valuator

@desc
Calculates position valuations and interest.
*/
type Valuator struct{}

/*
@function NewValuator

@desc
Creates new valuator instance.

@returns
- *Valuator
*/
func NewValuator() *Valuator {
	return &Valuator{}
}

/*
@function CalculateAccruedInterest

@desc
Calculates accrued interest for a position.

@params
- position: position to calculate for
- secondsElapsed: seconds since position creation
- compoundingPeriod: compounding period in seconds (0 for simple interest)

@returns
- *big.Int: accrued interest in wei
*/
func (v *Valuator) CalculateAccruedInterest(
position *models.Position,
secondsElapsed int64,
compoundingPeriod int64,
) *big.Int {
	if secondsElapsed <= 0 || position.APR <= 0 {
		return big.NewInt(0)
	}

	principal := position.Principal
	apr := big.NewInt(int64(position.APR))

	// Interest = Principal × APR × (Seconds / SecondsPerYear) / 10000
	interest := new(big.Int).Mul(principal, apr)
	interest.Mul(interest, big.NewInt(secondsElapsed))
	interest.Div(interest, big.NewInt(secondsPerYear))
	interest.Div(interest, big.NewInt(bpsScale))

	return interest
}

/*
@function CalculatePositionValue

@desc
Calculates total position value (principal + interest).

@params
- position: position with accrued interest

@returns
- *big.Int: total position value in wei
*/
func (v *Valuator) CalculatePositionValue(position *models.Position) *big.Int {
	value := new(big.Int).Add(position.Principal, position.AccruedInterest)
	return value
}

/*
@function CalculatePortfolioAPR

@desc
Calculates weighted average APR across positions.

@params
- positions: lender positions

@returns
- float64: weighted average APR in basis points
*/
func (v *Valuator) CalculatePortfolioAPR(positions []models.Position) float64 {
	if len(positions) == 0 {
		return 0
	}

	totalPrincipal := big.NewInt(0)
	weightedAPR := big.NewInt(0)

	for _, pos := range positions {
		if !pos.IsActive {
			continue
		}

		totalPrincipal.Add(totalPrincipal, pos.Principal)
		contribution := new(big.Int).Mul(pos.Principal, big.NewInt(int64(pos.APR)))
		weightedAPR.Add(weightedAPR, contribution)
	}

	if totalPrincipal.Cmp(big.NewInt(0)) == 0 {
		return 0
	}

	result := new(big.Rat).SetInt(weightedAPR)
	divisor := new(big.Rat).SetInt(totalPrincipal)
	result.Quo(result, divisor)

	floatResult, _ := result.Float64()
	return floatResult
}

/*
@function CalculatePortfolioValue

@desc
Calculates total portfolio value.

@params
- positions: lender positions

@returns
- *big.Int: total portfolio value in wei
*/
func (v *Valuator) CalculatePortfolioValue(positions []models.Position) *big.Int {
	totalValue := big.NewInt(0)

	for _, pos := range positions {
		if pos.IsActive {
			value := v.CalculatePositionValue(&pos)
			totalValue.Add(totalValue, value)
		}
	}

	return totalValue
}

/*
@function CalculateEarnings

@desc
Calculates total earnings across positions.

@params
- positions: lender positions

@returns
- *big.Int: total accrued interest in wei
*/
func (v *Valuator) CalculateEarnings(positions []models.Position) *big.Int {
	totalEarnings := big.NewInt(0)

	for _, pos := range positions {
		totalEarnings.Add(totalEarnings, pos.AccruedInterest)
	}

	return totalEarnings
}

/*
@function CalculateSeniorityValue

@desc
Calculates portfolio value by seniority.

@params
- positions: lender positions

@returns
- map[string]*big.Int: value by seniority ("senior"/"junior")
*/
func (v *Valuator) CalculateSeniorityValue(positions []models.Position) map[string]*big.Int {
	result := map[string]*big.Int{
		"senior": big.NewInt(0),
		"junior": big.NewInt(0),
	}

	for _, pos := range positions {
		if !pos.IsActive {
			continue
		}

		value := v.CalculatePositionValue(&pos)
		if pos.Seniority == 0 {
			result["senior"].Add(result["senior"], value)
		} else {
			result["junior"].Add(result["junior"], value)
		}
	}

	return result
}

/*
@function CalculateMaturityDate

@desc
Estimates position maturity based on offer details.

@params
- createdAt: position creation timestamp
- offerDuration: offer duration in seconds

@returns
- int64: estimated maturity timestamp
*/
func (v *Valuator) CalculateMaturityDate(createdAt int64, offerDuration int64) int64 {
	return createdAt + offerDuration
}
