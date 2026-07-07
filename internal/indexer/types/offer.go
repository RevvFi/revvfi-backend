package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

/*@
file: offer.go
package: types
layer: indexer

purpose:
OfferBook contract events.
*/

type OfferSubmittedEvent struct {
	OfferID   *big.Int
	Lender    common.Address
	Market    common.Address 
	Amount    *big.Int
	APR       *big.Int
	Seniority uint8
	Expiry    *big.Int
	Duration  *big.Int
}

type OfferCancelledEvent struct {
	OfferID         *big.Int
	Lender          common.Address
	Market          common.Address
	RemainingAmount *big.Int
}

type OfferExpiredEvent struct {
	OfferID         *big.Int
	Lender          common.Address
	Market          common.Address
	RemainingAmount *big.Int
}

type OfferModifiedEvent struct {
	OfferID   *big.Int
	Lender    common.Address
	Market    common.Address
	NewAmount *big.Int
	NewAPR    *big.Int
	NewExpiry *big.Int
}


type DrawdownExecutedOfferEvent struct {
    Borrower     common.Address
    TotalAmount  *big.Int
    WeightedAPR  *big.Int
}

type OfferFilledEvent struct {
    OfferID *big.Int
    Lender  common.Address
    Market  common.Address
    Amount  *big.Int
}