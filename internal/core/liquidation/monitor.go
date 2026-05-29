package liquidation

import (
	"context"
	"math/big"

	"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@file monitor.go

@desc
Health factor monitoring for liquidation detection.

@responsibilities
- Calculate health factors
- Detect liquidatable markets
- Monitor collateral ratios
*/

/*
@struct Monitor

@desc
Monitors market health and liquidation conditions.
*/
type Monitor struct{}

/*
@function NewMonitor

@desc
Creates new monitor instance.

@returns
- *Monitor
*/
func NewMonitor() *Monitor {
	return &Monitor{}
}

/*
@function CalculateHealthFactor

@desc
Calculates market health factor.
Health Factor = CollateralValue / DebtAmount
HF > 1 = healthy, HF < 1 = liquidatable

@params
- market: market to calculate for

@returns
- float64: health factor
*/
func (m *Monitor) CalculateHealthFactor(market *models.Market) float64 {
	if market.TotalDebt == nil || market.TotalDebt.Cmp(big.NewInt(0)) == 0 {
		return 1.0 // No debt = fully healthy
	}

	collateralValue := market.TotalLiquidity
	if collateralValue == nil {
		collateralValue = big.NewInt(0)
	}

	// Convert to float for division
	collateralFloat := new(big.Float).SetInt(collateralValue)
	debtFloat := new(big.Float).SetInt(market.TotalDebt)

	result := new(big.Float).Quo(collateralFloat, debtFloat)
	healthFactor, _ := result.Float64()

	return healthFactor
}

/*
@function IsMarketLiquidatable

@desc
Checks if market can be liquidated.
Market is liquidatable if health factor falls below liquidation threshold.

@params
- market: market to check

@returns
- bool: true if liquidatable
*/
func (m *Monitor) IsMarketLiquidatable(market *models.Market) bool {
	if !market.IsActive || market.IsLiquidating || market.IsClosed {
		return false
	}

	if market.TotalDebt == nil || market.TotalDebt.Cmp(big.NewInt(0)) == 0 {
		return false
	}

	healthFactor := m.CalculateHealthFactor(market)
	
	// Liquidation threshold in bps (e.g., 15000 = 150%)
	thresholdBps := float64(market.LiquidationThreshold) / 10000.0

	return healthFactor < thresholdBps
}

/*
@function CheckLiquidationHealth

@desc
Checks if liquidation health is valid.

@params
- ctx: request context
- market: market to check
- collateralRatio: collateral ratio

@returns
- bool: true if health check passes
- error: if check fails
*/
func (m *Monitor) CheckLiquidationHealth(
	ctx context.Context,
	market *models.Market,
	collateralRatio *big.Float,
) (bool, error) {
	// Market must be active
	if !market.IsActive {
		return false, nil
	}

	// Market must not already be liquidating
	if market.IsLiquidating {
		return false, nil
	}

	// Market must have debt
	if market.TotalDebt == nil || market.TotalDebt.Cmp(big.NewInt(0)) == 0 {
		return false, nil
	}

	// Check if health factor is below liquidation threshold
	healthFactor := m.CalculateHealthFactor(market)
	minRatioBps := float64(market.MinCollateralRatio) / 10000.0

	return healthFactor < minRatioBps, nil
}

/*
@function GetRiskLevel

@desc
Determines market risk level based on health factor.

@params
- market: market to assess

@returns
- string: risk level (critical, high, medium, low)
*/
func (m *Monitor) GetRiskLevel(market *models.Market) string {
	healthFactor := m.CalculateHealthFactor(market)

	if healthFactor < 1.0 {
		return "critical"
	} else if healthFactor < 1.2 {
		return "high"
	} else if healthFactor < 1.5 {
		return "medium"
	}

	return "low"
}

/*
@function GetLiquidationPrice

@desc
Calculates minimum price needed to avoid liquidation.

@params
- market: market to calculate for

@returns
- *big.Int: minimum collateral price in wei
*/
func (m *Monitor) GetLiquidationPrice(market *models.Market) *big.Int {
	if market.TotalDebt == nil || market.TotalDebt.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0)
	}

	// Required collateral = Debt × MinCollateralRatio
	minRatioBps := int64(market.MinCollateralRatio)
	requiredCollateral := new(big.Int).Mul(market.TotalDebt, big.NewInt(minRatioBps))
	requiredCollateral.Div(requiredCollateral, big.NewInt(10000))

	return requiredCollateral
}

/*
@function GetSafetyMargin

@desc
Calculates safety margin percentage.

@params
- market: market to calculate for

@returns
- float64: safety margin as percentage (0-1)
*/
func (m *Monitor) GetSafetyMargin(market *models.Market) float64 {
	healthFactor := m.CalculateHealthFactor(market)
	
	// Safety margin = (HF - 1) / (Threshold - 1)
	thresholdBps := float64(market.LiquidationThreshold) / 10000.0
	
	if thresholdBps <= 1.0 {
		return 0
	}

	margin := (healthFactor - 1.0) / (thresholdBps - 1.0)
	
	if margin < 0 {
		return 0
	} else if margin > 1 {
		return 1
	}

	return margin
}
