package liquidation

import (
	"math/big"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@file auction.go

@desc
Dutch auction management for liquidation.

@responsibilities
- Calculate Dutch prices
- Track price decay
- Manage auction state
*/

/*
@struct AuctionManager

@desc
Manages Dutch auction mechanics.
*/
type AuctionManager struct{}

/*
@function NewAuctionManager

@desc
Creates new auction manager.

@returns
- *AuctionManager
*/
func NewAuctionManager() *AuctionManager {
	return &AuctionManager{}
}

/*
@function CalculateDutchPrice

@desc
Calculates current Dutch auction price.
Linearly decays from start to end price over duration.

@params
- auction: auction to calculate for

@returns
- *big.Int: current price in wei
*/
func (m *AuctionManager) CalculateDutchPrice(auction *models.Auction) *big.Int {
	now := time.Now()

	// Auction hasn't started
	if now.Before(auction.StartTime) {
		return new(big.Int).Set(auction.DebtAmount)
	}

	// Auction expired
	if now.After(auction.EndTime) {
		return new(big.Int).Set(auction.CurrentPrice)
	}

	// Calculate elapsed time ratio
	duration := auction.EndTime.Sub(auction.StartTime)
	elapsed := now.Sub(auction.StartTime)

	// Price decay: debt - (debt - current) * (elapsed / duration)
	priceDiff := new(big.Int).Sub(auction.DebtAmount, auction.CurrentPrice)
	
	// Calculate fraction of duration elapsed
	fractionElapsed := float64(elapsed.Seconds()) / float64(duration.Seconds())
	
	// Apply decay
	decayAmount := new(big.Float).Mul(new(big.Float).SetInt(priceDiff), big.NewFloat(fractionElapsed))
	decayInt := new(big.Int)
	decayAmount.Int(decayInt)

	resultPrice := new(big.Int).Sub(auction.DebtAmount, decayInt)

	// Ensure price doesn't go below current price
	if resultPrice.Cmp(auction.CurrentPrice) < 0 {
		resultPrice.Set(auction.CurrentPrice)
	}

	return resultPrice
}

/*
@function IsAuctionExpired

@desc
Checks if auction has expired.

@params
- auction: auction to check

@returns
- bool: true if expired
*/
func (m *AuctionManager) IsAuctionExpired(auction *models.Auction) bool {
	return time.Now().After(auction.EndTime)
}

/*
@function GetRemainingTime

@desc
Gets remaining auction duration.

@params
- auction: auction to check

@returns
- int64: seconds remaining (0 if expired)
*/
func (m *AuctionManager) GetRemainingTime(auction *models.Auction) int64 {
	remaining := auction.EndTime.Sub(time.Now())
	if remaining < 0 {
		return 0
	}
	return int64(remaining.Seconds())
}

/*
@function CalculatePriceDecay

@desc
Calculates price decay per second.

@params
- auction: auction to calculate for

@returns
- *big.Int: price decay per second in wei
*/
func (m *AuctionManager) CalculatePriceDecay(auction *models.Auction) *big.Int {
	if auction.StartTime.After(auction.EndTime) {
		return big.NewInt(0)
	}

	duration := auction.EndTime.Sub(auction.StartTime)
	priceDiff := new(big.Int).Sub(auction.DebtAmount, auction.CurrentPrice)
	
	decay := new(big.Int).Div(priceDiff, big.NewInt(int64(duration.Seconds())))
	return decay
}

/*
@function HasValidBid

@desc
Checks if auction has valid bid.

@params
- auction: auction to check

@returns
- bool: true if auction has bid
*/
func (m *AuctionManager) HasValidBid(auction *models.Auction) bool {
	return auction.HighestBid != nil && auction.HighestBid.Cmp(big.NewInt(0)) > 0
}

/*
@function CalculateBidSurplus

@desc
Calculates amount bid above current price.

@params
- auction: auction to check
- currentPrice: current auction price

@returns
- *big.Int: surplus amount in wei
*/
func (m *AuctionManager) CalculateBidSurplus(
	auction *models.Auction,
	currentPrice *big.Int,
) *big.Int {
	if auction.HighestBid == nil {
		return big.NewInt(0)
	}

	surplus := new(big.Int).Sub(auction.HighestBid, currentPrice)
	if surplus.Cmp(big.NewInt(0)) < 0 {
		return big.NewInt(0)
	}

	return surplus
}
