package borrower

import (
"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@file reputation.go

@desc
Reputation score calculation and risk assessment logic.
Implements the RevvFi reputation algorithm.

@responsibilities
- Calculate reputation scores (0-1000)
- Determine risk labels (AAA-D)
- Calculate risk levels (low/medium/high/critical)
- Calculate health factors
*/

/*
@struct RepCalculator

@desc
Calculates borrower reputation metrics.
*/
type RepCalculator struct{}

/*
@function NewRepCalculator

@desc
Creates new reputation calculator.

@returns
- *RepCalculator
*/
func NewRepCalculator() *RepCalculator {
	return &RepCalculator{}
}

/*
@function CalculateReputation

@desc
Calculates reputation score (0-1000).
Formula: (SuccessfulLoans / TotalLoans × 1000) - (Defaults × 50)
Clamped to [0, 1000].

@params
- totalLoans: total loans
- successfulLoans: successful repayments
- defaultedLoans: defaults

@returns
- int32: reputation score
*/
func (r *RepCalculator) CalculateReputation(
totalLoans, successfulLoans, defaultedLoans int32,
) int32 {
	if totalLoans == 0 {
		return 500 // Default score for new borrowers
	}

	successRate := (int64(successfulLoans) * 1000) / int64(totalLoans)
	penalty := int64(defaultedLoans) * 50

	score := int32(successRate - penalty)

	// Clamp to [0, 1000]
	if score < 0 {
		score = 0
	}
	if score > 1000 {
		score = 1000
	}

	return score
}

/*
@function CalculateRiskLabel

@desc
Determines risk label from reputation score.
- AAA: >= 900
- AA: 800-899
- A: 700-799
- B: 500-699
- C: 300-499
- D: < 300

@params
- score: reputation score

@returns
- string: risk label
*/
func (r *RepCalculator) CalculateRiskLabel(score int32) string {
	switch {
	case score >= 900:
		return "AAA"
	case score >= 800:
		return "AA"
	case score >= 700:
		return "A"
	case score >= 500:
		return "B"
	case score >= 300:
		return "C"
	default:
		return "D"
	}
}

/*
@function GetRiskLevel

@desc
Maps risk label to risk level.

@params
- score: reputation score

@returns
- string: risk level (low|medium|high|critical)
*/
func (r *RepCalculator) GetRiskLevel(score int32) string {
	switch {
	case score >= 800:
		return "low"
	case score >= 600:
		return "medium"
	case score >= 400:
		return "high"
	default:
		return "critical"
	}
}

/*
@function CalculateHealthFactor

@desc
Calculates borrower health factor.
Higher = healthier.

@params
- borrower: borrower profile

@returns
- float64: health factor (0-1 scale)
*/
func (r *RepCalculator) CalculateHealthFactor(borrower *models.Borrower) float64 {
	// Health = (ReputationScore / 1000) × (SuccessRate / 100)
	reputationHealth := float64(borrower.ReputationScore) / 1000.0
	successHealth := borrower.SuccessRate / 100.0

	return reputationHealth * (0.7 + 0.3*successHealth)
}
