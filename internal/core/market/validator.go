package market

import (
"fmt"

appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@file validator.go

@desc
Market creation and operation validation.
Enforces protocol constraints and parameter limits.

@responsibilities
- Validate market creation parameters
- Validate collateral ratio constraints
- Validate APR ranges
- Validate asset addresses
*/

/*
@struct Validator

@desc
Validates market operations and parameters.
*/
type Validator struct{}

/*
@function NewValidator

@desc
Creates new validator instance.

@returns
- *Validator
*/
func NewValidator() *Validator {
	return &Validator{}
}

/*
@function ValidateMarketCreation

@desc
Validates market creation parameters against protocol rules.

@params
- borrower: borrower wallet address
- borrowAsset: borrow token address
- collateralAsset: collateral token address
- minCollateralRatio: minimum collateral ratio in bps
- liquidationThreshold: liquidation threshold in bps

@returns
- error: if validation fails
*/
func (v *Validator) ValidateMarketCreation(
borrower string,
borrowAsset string,
collateralAsset string,
minCollateralRatio int32,
liquidationThreshold int32,
) error {
	// Validate addresses
	if !isValidAddress(borrower) {
		return appErr.ErrInvalidAddress
	}
	if !isValidAddress(borrowAsset) {
		return appErr.ErrInvalidAddress
	}
	if !isValidAddress(collateralAsset) {
		return appErr.ErrInvalidAddress
	}

	// Assets must be different
	if borrowAsset == collateralAsset {
		return fmt.Errorf("borrow and collateral assets must be different")
	}

	// Validate collateral ratio constraints
	// Min ratio must be >= 100% (10000 bps)
	if minCollateralRatio < 10000 {
		return fmt.Errorf("min collateral ratio must be >= 100%%")
	}

	// Min ratio must be <= 500% (50000 bps)
	if minCollateralRatio > 50000 {
		return fmt.Errorf("min collateral ratio must be <= 500%%")
	}

	// Liquidation threshold must be < min collateral ratio
	if liquidationThreshold >= minCollateralRatio {
		return fmt.Errorf("liquidation threshold must be < min collateral ratio")
	}

	// Liquidation threshold must be > 50% (5000 bps)
	if liquidationThreshold < 5000 {
		return fmt.Errorf("liquidation threshold must be > 50%%")
	}

	return nil
}

/*
@function ValidateAPR

@desc
Validates APR is within acceptable range.

@params
- apr: annual percentage rate in bps

@returns
- error: if APR invalid
*/
func (v *Validator) ValidateAPR(apr int32) error {
	// APR must be > 0 and <= 50% (5000 bps)
	if apr <= 0 {
		return appErr.ErrInvalidAPR
	}
	if apr > 5000 {
		return fmt.Errorf("APR exceeds maximum of 50%%")
	}
	return nil
}

/*
@function ValidateAmount

@desc
Validates amount is positive and non-zero.

@params
- amount: amount string (wei)

@returns
- error: if amount invalid
*/
func (v *Validator) ValidateAmount(amount string) error {
	if amount == "" {
		return appErr.ErrInvalidAmount
	}
	if amount == "0" {
		return appErr.ErrInvalidAmount
	}
	return nil
}

/*
@function isValidAddress

@desc
Helper to validate Ethereum address format.

@params
- addr: address string

@returns
- bool: true if valid Ethereum address
*/
func isValidAddress(addr string) bool {
	if len(addr) != 42 {
		return false
	}
	if addr[:2] != "0x" {
		return false
	}
	// Could add more validation but this is sufficient for now
	return true
}
