package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

/*@
file: auction.go
package: types

purpose:
Liquidation auction events.
*/

type AuctionCreatedEvent struct {
	AuctionID        *big.Int
	Market           common.Address
	Borrower         common.Address
	CollateralAmount *big.Int
	DebtAmount       *big.Int
	StartTime        *big.Int
	EndTime          *big.Int
}

type BidPlacedEvent struct {
	AuctionID *big.Int
	Bidder    common.Address
	BidAmount *big.Int
}

type AuctionSettledEvent struct {
	AuctionID             *big.Int
	Winner                common.Address
	WinningBid            *big.Int
	CollateralTransferred *big.Int
	RecoveryRate          *big.Int
}

type AuctionCancelledEvent struct {
	AuctionID *big.Int
	Reason    string
}