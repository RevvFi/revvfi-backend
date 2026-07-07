package position

import (
"math/big"
"testing"

"github.com/Revvfi/revvfi-backend/internal/models"
"github.com/stretchr/testify/assert"
)

/*
@file valuation_test.go

@desc
Unit tests for position valuation.
*/

/*
@function TestCalculateAccruedInterest

@desc
Tests accrued interest calculation.
*/
func TestCalculateAccruedInterest(t *testing.T) {
	tests := []struct {
		name            string
		principal       *big.Int
		apr             int32
		secondsElapsed  int64
		expectedMinimum *big.Int
		expectedMaximum *big.Int
	}{
		{
			name:            "one year at 5% (500 bps)",
			principal:       big.NewInt(1000000),
			apr:             500,
			secondsElapsed:  31536000, // 1 year
			expectedMinimum: big.NewInt(50000),
			expectedMaximum: big.NewInt(50000),
		},
		{
			name:            "half year at 10% (1000 bps)",
			principal:       big.NewInt(1000000),
			apr:             1000,
			secondsElapsed:  15768000, // ~0.5 year
			expectedMinimum: big.NewInt(49761),
			expectedMaximum: big.NewInt(50000),
		},
		{
			name:            "zero seconds elapsed",
			principal:       big.NewInt(1000000),
			apr:             500,
			secondsElapsed:  0,
			expectedMinimum: big.NewInt(0),
			expectedMaximum: big.NewInt(0),
		},
		{
			name:            "zero APR",
			principal:       big.NewInt(1000000),
			apr:             0,
			secondsElapsed:  31536000,
			expectedMinimum: big.NewInt(0),
			expectedMaximum: big.NewInt(0),
		},
	}

	valuator := NewValuator()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
pos := &models.Position{
				Principal:        tc.principal,
				CurrentPrincipal: tc.principal,
				APR:              tc.apr,
			}

			interest := valuator.CalculateAccruedInterest(pos, tc.secondsElapsed, 0)

			assert.GreaterOrEqual(t, interest.Cmp(tc.expectedMinimum), 0)
			assert.LessOrEqual(t, interest.Cmp(tc.expectedMaximum), 0)
		})
	}
}

/*
@function TestCalculatePositionValue

@desc
Tests position value calculation.
*/
func TestCalculatePositionValue(t *testing.T) {
	tests := []struct {
		name            string
		principal       *big.Int
		accruedInterest *big.Int
		expected        *big.Int
	}{
		{
			name:            "principal + interest",
			principal:       big.NewInt(1000000),
			accruedInterest: big.NewInt(50000),
			expected:        big.NewInt(1050000),
		},
		{
			name:            "principal only",
			principal:       big.NewInt(1000000),
			accruedInterest: big.NewInt(0),
			expected:        big.NewInt(1000000),
		},
		{
			name:            "large numbers",
			principal:       big.NewInt(10000000000),
			accruedInterest: big.NewInt(500000000),
			expected:        big.NewInt(10500000000),
		},
	}

	valuator := NewValuator()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
pos := &models.Position{
				Principal:       tc.principal,
				AccruedInterest: tc.accruedInterest,
			}

			value := valuator.CalculatePositionValue(pos)

			assert.Equal(t, tc.expected.String(), value.String())
		})
	}
}

/*
@function TestCalculatePortfolioAPR

@desc
Tests portfolio APR calculation.
*/
func TestCalculatePortfolioAPR(t *testing.T) {
	tests := []struct {
		name            string
		positions       []models.Position
		expectedAPRMin  float64
		expectedAPRMax  float64
	}{
		{
			name: "single position",
			positions: []models.Position{
				{
					Principal: big.NewInt(1000000),
					APR:       500,
					IsActive:  true,
				},
			},
			expectedAPRMin: 499,
			expectedAPRMax: 501,
		},
		{
			name: "two positions equal weight",
			positions: []models.Position{
				{
					Principal: big.NewInt(1000000),
					APR:       400,
					IsActive:  true,
				},
				{
					Principal: big.NewInt(1000000),
					APR:       600,
					IsActive:  true,
				},
			},
			expectedAPRMin: 499,
			expectedAPRMax: 501,
		},
		{
			name: "inactive positions ignored",
			positions: []models.Position{
				{
					Principal: big.NewInt(1000000),
					APR:       500,
					IsActive:  true,
				},
				{
					Principal: big.NewInt(1000000),
					APR:       5000,
					IsActive:  false,
				},
			},
			expectedAPRMin: 499,
			expectedAPRMax: 501,
		},
		{
			name:            "empty positions",
			positions:       []models.Position{},
			expectedAPRMin:  0,
			expectedAPRMax:  0,
		},
	}

	valuator := NewValuator()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
apr := valuator.CalculatePortfolioAPR(tc.positions)

assert.GreaterOrEqual(t, apr, tc.expectedAPRMin)
assert.LessOrEqual(t, apr, tc.expectedAPRMax)
})
	}
}

/*
@function TestCalculatePortfolioValue

@desc
Tests total portfolio value calculation.
*/
func TestCalculatePortfolioValue(t *testing.T) {
	tests := []struct {
		name     string
		positions []models.Position
		expected *big.Int
	}{
		{
			name: "single position",
			positions: []models.Position{
				{
					IsActive:     true,
					Principal:    big.NewInt(1000000),
					AccruedInterest: big.NewInt(50000),
				},
			},
			expected: big.NewInt(1050000),
		},
		{
			name: "multiple positions",
			positions: []models.Position{
				{
					IsActive:     true,
					Principal:    big.NewInt(1000000),
					AccruedInterest: big.NewInt(50000),
				},
				{
					IsActive:     true,
					Principal:    big.NewInt(2000000),
					AccruedInterest: big.NewInt(100000),
				},
			},
			expected: big.NewInt(3150000),
		},
		{
			name: "inactive positions excluded",
			positions: []models.Position{
				{
					IsActive:     true,
					Principal:    big.NewInt(1000000),
					AccruedInterest: big.NewInt(50000),
				},
				{
					IsActive:     false,
					Principal:    big.NewInt(5000000),
					AccruedInterest: big.NewInt(250000),
				},
			},
			expected: big.NewInt(1050000),
		},
	}

	valuator := NewValuator()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
value := valuator.CalculatePortfolioValue(tc.positions)

assert.Equal(t, tc.expected.String(), value.String())
		})
	}
}

/*
@function TestCalculateEarnings

@desc
Tests total earnings calculation.
*/
func TestCalculateEarnings(t *testing.T) {
	tests := []struct {
		name     string
		positions []models.Position
		expected *big.Int
	}{
		{
			name: "single position",
			positions: []models.Position{
				{
					AccruedInterest: big.NewInt(50000),
				},
			},
			expected: big.NewInt(50000),
		},
		{
			name: "multiple positions",
			positions: []models.Position{
				{
					AccruedInterest: big.NewInt(50000),
				},
				{
					AccruedInterest: big.NewInt(100000),
				},
			},
			expected: big.NewInt(150000),
		},
		{
			name:     "no positions",
			positions: []models.Position{},
			expected: big.NewInt(0),
		},
	}

	valuator := NewValuator()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
earnings := valuator.CalculateEarnings(tc.positions)

assert.Equal(t, tc.expected.String(), earnings.String())
		})
	}
}

/*
@function TestCalculateSeniorityValue

@desc
Tests seniority-based value calculation.
*/
func TestCalculateSeniorityValue(t *testing.T) {
	valuator := NewValuator()

	positions := []models.Position{
		{
			IsActive:     true,
			Principal:    big.NewInt(1000000),
			AccruedInterest: big.NewInt(50000),
			Seniority:    0, // senior
		},
		{
			IsActive:     true,
			Principal:    big.NewInt(2000000),
			AccruedInterest: big.NewInt(100000),
			Seniority:    0, // senior
		},
		{
			IsActive:     true,
			Principal:    big.NewInt(500000),
			AccruedInterest: big.NewInt(25000),
			Seniority:    1, // junior
		},
	}

	result := valuator.CalculateSeniorityValue(positions)

	assert.Equal(t, "3150000", result["senior"].String())
	assert.Equal(t, "525000", result["junior"].String())
}

/*
@function TestCalculateMaturityDate

@desc
Tests maturity date calculation.
*/
func TestCalculateMaturityDate(t *testing.T) {
	valuator := NewValuator()

	createdAt := int64(1000000)
	duration := int64(31536000) // 1 year

	maturity := valuator.CalculateMaturityDate(createdAt, duration)

	assert.Equal(t, int64(32536000), maturity)
}
