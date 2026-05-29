package withdrawal

import (
	"math/big"
)

/*
@file epoch.go

@desc
Epoch calculator for withdrawal processing.
Handles fulfillment ratio calculations and epoch metrics.

@responsibilities
- Calculate fulfillment ratios
- Determine withdrawal scheduling
- Compute epoch metrics
*/

/*
@struct EpochCalculator

@desc
Calculates epoch-related metrics.
*/
type EpochCalculator struct{}

/*
@function NewEpochCalculator

@desc
Creates new epoch calculator.

@returns
- *EpochCalculator
*/
func NewEpochCalculator() *EpochCalculator {
	return &EpochCalculator{}
}

/*
@function CalculateFulfillmentRatio

@desc
Calculates fulfillment ratio for withdrawal requests.

@params
- totalRequested: total requested withdrawal amount
- availableLiquidity: available liquidity for fulfillment

@returns
- float64: fulfillment ratio (0-1)
*/
func (e *EpochCalculator) CalculateFulfillmentRatio(
	totalRequested *big.Int,
	availableLiquidity *big.Int,
) float64 {
	if totalRequested.Cmp(big.NewInt(0)) == 0 {
		return 1.0
	}

	if availableLiquidity.Cmp(totalRequested) >= 0 {
		return 1.0
	}

	// Convert to float for ratio calculation
	requestedFloat := new(big.Rat).SetInt(totalRequested)
	liquidityFloat := new(big.Rat).SetInt(availableLiquidity)

	ratio := new(big.Rat).Quo(liquidityFloat, requestedFloat)
	result, _ := ratio.Float64()

	if result > 1.0 {
		return 1.0
	}
	if result < 0.0 {
		return 0.0
	}

	return result
}

/*
@function CalculateEpochDuration

@desc
Calculates duration of withdrawal epoch in seconds.

@params
- epochNumber: epoch number

@returns
- int64: epoch duration in seconds
*/
func (e *EpochCalculator) CalculateEpochDuration(epochNumber int64) int64 {
	// Standard epoch duration: 7 days
	return 7 * 24 * 3600
}

/*
@function CalculateEpochStartTime

@desc
Calculates start time for epoch.

@params
- epochNumber: epoch number
- genesisTime: genesis timestamp

@returns
- int64: epoch start timestamp
*/
func (e *EpochCalculator) CalculateEpochStartTime(epochNumber int64, genesisTime int64) int64 {
	return genesisTime + (epochNumber * e.CalculateEpochDuration(epochNumber))
}

/*
@function CalculateEpochEndTime

@desc
Calculates end time for epoch.

@params
- epochNumber: epoch number
- genesisTime: genesis timestamp

@returns
- int64: epoch end timestamp
*/
func (e *EpochCalculator) CalculateEpochEndTime(epochNumber int64, genesisTime int64) int64 {
	return e.CalculateEpochStartTime(epochNumber, genesisTime) + e.CalculateEpochDuration(epochNumber)
}

/*
@function IsEpochActive

@desc
Checks if epoch is currently active.

@params
- epochNumber: epoch number
- currentTime: current timestamp
- genesisTime: genesis timestamp

@returns
- bool: true if epoch is active
*/
func (e *EpochCalculator) IsEpochActive(epochNumber int64, currentTime int64, genesisTime int64) bool {
	startTime := e.CalculateEpochStartTime(epochNumber, genesisTime)
	endTime := e.CalculateEpochEndTime(epochNumber, genesisTime)

	return currentTime >= startTime && currentTime < endTime
}

/*
@function GetCurrentEpoch

@desc
Calculates current epoch number.

@params
- currentTime: current timestamp
- genesisTime: genesis timestamp

@returns
- int64: current epoch number
*/
func (e *EpochCalculator) GetCurrentEpoch(currentTime int64, genesisTime int64) int64 {
	if currentTime < genesisTime {
		return 0
	}

	epochDuration := e.CalculateEpochDuration(0)
	return (currentTime - genesisTime) / epochDuration
}

/*
@function CalculatePercentageFulfilled

@desc
Calculates percentage of requests fulfilled.

@params
- totalFulfilled: total fulfilled amount
- totalRequested: total requested amount

@returns
- float64: percentage (0-100)
*/
func (e *EpochCalculator) CalculatePercentageFulfilled(
	totalFulfilled *big.Int,
	totalRequested *big.Int,
) float64 {
	if totalRequested.Cmp(big.NewInt(0)) == 0 {
		return 0
	}

	ratio := e.CalculateFulfillmentRatio(totalRequested, totalFulfilled)
	return ratio * 100
}
