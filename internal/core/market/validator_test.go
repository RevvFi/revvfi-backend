package market

import (
"testing"
)

/*
@file validator_test.go

@desc
Unit tests for market validator.
Tests parameter validation and business rule enforcement.

@test_coverage
- ValidateMarketCreation: valid/invalid parameters
- ValidateAPR: APR range validation
- ValidateAmount: amount validation
- isValidAddress: address format validation
*/

func TestValidateMarketCreation(t *testing.T) {
	tests := []struct {
		name                 string
		borrower             string
		borrowAsset          string
		collateralAsset      string
		minCollateralRatio   int32
		liquidationThreshold int32
		expectError          bool
	}{
		{
			name:                 "valid market",
			borrower:             "0x0000000000000000000000000000000000000001",
			borrowAsset:          "0x0000000000000000000000000000000000000002",
			collateralAsset:      "0x0000000000000000000000000000000000000003",
			minCollateralRatio:   15000, // 150%
			liquidationThreshold: 12500, // 125%
			expectError:          false,
		},
		{
			name:                 "invalid borrower",
			borrower:             "invalid",
			borrowAsset:          "0x0000000000000000000000000000000000000002",
			collateralAsset:      "0x0000000000000000000000000000000000000003",
			minCollateralRatio:   15000,
			liquidationThreshold: 12500,
			expectError:          true,
		},
		{
			name:                 "same assets",
			borrower:             "0x0000000000000000000000000000000000000001",
			borrowAsset:          "0x0000000000000000000000000000000000000002",
			collateralAsset:      "0x0000000000000000000000000000000000000002",
			minCollateralRatio:   15000,
			liquidationThreshold: 12500,
			expectError:          true,
		},
		{
			name:                 "collateral ratio too low",
			borrower:             "0x0000000000000000000000000000000000000001",
			borrowAsset:          "0x0000000000000000000000000000000000000002",
			collateralAsset:      "0x0000000000000000000000000000000000000003",
			minCollateralRatio:   5000, // < 100%
			liquidationThreshold: 4000,
			expectError:          true,
		},
		{
			name:                 "liquidation >= collateral ratio",
			borrower:             "0x0000000000000000000000000000000000000001",
			borrowAsset:          "0x0000000000000000000000000000000000000002",
			collateralAsset:      "0x0000000000000000000000000000000000000003",
			minCollateralRatio:   15000,
			liquidationThreshold: 15000, // Should be < minCollateralRatio
			expectError:          true,
		},
	}

	validator := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
err := validator.ValidateMarketCreation(
tt.borrower,
tt.borrowAsset,
tt.collateralAsset,
tt.minCollateralRatio,
tt.liquidationThreshold,
)

if tt.expectError && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateAPR(t *testing.T) {
	tests := []struct {
		name        string
		apr         int32
		expectError bool
	}{
		{
			name:        "valid APR 500 bps (5%)",
			apr:         500,
			expectError: false,
		},
		{
			name:        "max APR 5000 bps (50%)",
			apr:         5000,
			expectError: false,
		},
		{
			name:        "zero APR",
			apr:         0,
			expectError: true,
		},
		{
			name:        "negative APR",
			apr:         -100,
			expectError: true,
		},
		{
			name:        "APR exceeds max",
			apr:         6000,
			expectError: true,
		},
	}

	validator := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
err := validator.ValidateAPR(tt.apr)

if tt.expectError && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name        string
		amount      string
		expectError bool
	}{
		{
			name:        "valid amount",
			amount:      "1000000000000000000",
			expectError: false,
		},
		{
			name:        "empty amount",
			amount:      "",
			expectError: true,
		},
		{
			name:        "zero amount",
			amount:      "0",
			expectError: true,
		},
	}

	validator := NewValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
err := validator.ValidateAmount(tt.amount)

if tt.expectError && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestIsValidAddress(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		expected bool
	}{
		{
			name:     "valid address",
			addr:     "0x0000000000000000000000000000000000000001",
			expected: true,
		},
		{
			name:     "missing 0x prefix",
			addr:     "0000000000000000000000000000000000000001",
			expected: false,
		},
		{
			name:     "too short",
			addr:     "0x1234",
			expected: false,
		},
		{
			name:     "empty",
			addr:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
result := isValidAddress(tt.addr)
if result != tt.expected {
t.Errorf("expected %v, got %v", tt.expected, result)
}
})
	}
}
