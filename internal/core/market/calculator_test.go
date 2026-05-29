package market

import (
	"math"
	"math/big"
	"testing"

	"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@file calculator_test.go

@desc
Unit tests for market calculator.
Tests financial calculations for APR, utilization, and collateral ratios.

@test_coverage
- CalculateUtilization: various ratios
- CalculateAverageAPR: weighted APR
- CalculateCollateralRatio: collateral ratio
- CalculateInterestAccrual: interest calculations
- IsLiquidatable: liquidation checks
*/

func TestCalculateUtilization(t *testing.T) {
	tests := []struct {
		name       string
		totalDebt  *big.Int
		liquidity  *big.Int
		expected   float64
	}{
		{
			name:      "50% utilization",
			totalDebt: big.NewInt(5000000),
			liquidity: big.NewInt(10000000),
			expected:  50.0,
		},
		{
			name:      "0% utilization (no debt)",
			totalDebt: big.NewInt(0),
			liquidity: big.NewInt(10000000),
			expected:  0.0,
		},
		{
			name:      "100% utilization",
			totalDebt: big.NewInt(10000000),
			liquidity: big.NewInt(10000000),
			expected:  100.0,
		},
		{
			name:      "0% utilization (no liquidity)",
			totalDebt: big.NewInt(5000000),
			liquidity: big.NewInt(0),
			expected:  0.0,
		},
	}

	calc := NewCalculator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
result := calc.CalculateUtilization(tt.totalDebt, tt.liquidity)
if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("expected %.2f, got %.2f", tt.expected, result)
			}
		})
	}
}

func TestCalculateAverageAPR(t *testing.T) {
	tests := []struct {
		name      string
		offers    []models.Offer
		expected  int32
	}{
		{
			name:     "single offer",
			offers:   []models.Offer{{APR: 500, RemainingAmount: big.NewInt(1000000)}},
			expected: 500,
		},
		{
			name: "weighted average APR",
			offers: []models.Offer{
				{APR: 500, RemainingAmount: big.NewInt(1000000)},
				{APR: 300, RemainingAmount: big.NewInt(1000000)},
			},
			expected: 400, // (500 + 300) / 2
		},
		{
			name:     "no offers",
			offers:   []models.Offer{},
			expected: 0,
		},
	}

	calc := NewCalculator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
result := calc.CalculateAverageAPR(tt.offers)
if result != tt.expected {
t.Errorf("expected %d, got %d", tt.expected, result)
}
})
	}
}

func TestCalculateCollateralRatio(t *testing.T) {
	tests := []struct {
		name           string
		collateralVal  *big.Int
		totalDebt      *big.Int
		expectedMin    float64
		expectedMax    float64
	}{
		{
			name:          "150% collateral ratio",
			collateralVal: big.NewInt(1500000),
			totalDebt:     big.NewInt(1000000),
			expectedMin:   1.4,
			expectedMax:   1.6,
		},
		{
			name:          "100% collateral ratio",
			collateralVal: big.NewInt(1000000),
			totalDebt:     big.NewInt(1000000),
			expectedMin:   0.9,
			expectedMax:   1.1,
		},
		{
			name:          "infinite ratio (no debt)",
			collateralVal: big.NewInt(1000000),
			totalDebt:     big.NewInt(0),
			expectedMin:   math.Inf(1) - 1,
			expectedMax:   math.Inf(1) + 1,
		},
	}

	calc := NewCalculator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
result := calc.CalculateCollateralRatio(tt.collateralVal, tt.totalDebt)
if result >= tt.expectedMin && result <= tt.expectedMax {
				// ok
			} else {
				t.Errorf("expected between %.2f and %.2f, got %.2f", tt.expectedMin, tt.expectedMax, result)
			}
		})
	}
}

func TestCalculateInterestAccrual(t *testing.T) {
	tests := []struct {
		name           string
		principal      *big.Int
		apr            int32
		secondsElapsed int64
		expectedMin    *big.Int
		expectedMax    *big.Int
	}{
		{
			name:           "1 year at 10% APR",
			principal:      big.NewInt(1000000000000000000), // 1e18
			apr:            1000,                              // 10%
			secondsElapsed: 365 * 24 * 60 * 60,
			expectedMin:    big.NewInt(99000000000000000), // ~0.1e18
			expectedMax:    big.NewInt(101000000000000000),
		},
		{
			name:           "no time elapsed",
			principal:      big.NewInt(1000000000000000000),
			apr:            1000,
			secondsElapsed: 0,
			expectedMin:    big.NewInt(0),
			expectedMax:    big.NewInt(1),
		},
	}

	calc := NewCalculator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
result := calc.CalculateInterestAccrual(tt.principal, tt.apr, tt.secondsElapsed)
if result.Cmp(tt.expectedMin) < 0 || result.Cmp(tt.expectedMax) > 0 {
				t.Errorf("expected between %s and %s, got %s",
tt.expectedMin.String(),
					tt.expectedMax.String(),
					result.String(),
				)
			}
		})
	}
}

func TestIsLiquidatable(t *testing.T) {
	tests := []struct {
		name                 string
		collateralRatio      float64
		liquidationThreshold int32
		expected             bool
	}{
		{
			name:                 "above threshold - not liquidatable",
			collateralRatio:      1.3, // 130%
			liquidationThreshold: 12500, // 125%
			expected:             false,
		},
		{
			name:                 "below threshold - liquidatable",
			collateralRatio:      1.2, // 120%
			liquidationThreshold: 12500, // 125%
			expected:             true,
		},
		{
			name:                 "at threshold - not liquidatable",
			collateralRatio:      1.25, // 125%
			liquidationThreshold: 12500, // 125%
			expected:             false,
		},
	}

	calc := NewCalculator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
result := calc.IsLiquidatable(tt.collateralRatio, tt.liquidationThreshold)
if result != tt.expected {
t.Errorf("expected %v, got %v", tt.expected, result)
}
})
	}
}
