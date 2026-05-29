package event

import (
	"context"

	"github.com/Revvfi/revvfi-backend/internal/blockchain/types"
)

/*
@struct Listener

@desc
Event listener scaffold for RevvFi contract logs.

@responsibilities
- Define supported event names
- Provide a polling entrypoint for indexer integration
*/
type Listener struct {
	supported []string
}

/*
@function NewListener

@desc
Creates an event listener with the RevvFi event allowlist.

@returns
- *Listener
*/
func NewListener() *Listener {
	return &Listener{supported: []string{"MarketDeployed", "OfferSubmitted", "OfferCancelled", "PositionMinted", "Borrow", "Repay", "AuctionCreated", "BidPlaced", "AuctionSettled", "WithdrawalRequested", "WithdrawalClaimed", "WithdrawalCancelled"}}
}

/*
@method SupportedEvents

@desc
Returns the list of event names this listener should decode.

@returns
- []string
*/
func (l *Listener) SupportedEvents() []string {
	return append([]string(nil), l.supported...)
}

/*
@method Poll

@desc
Polling hook for future indexer block range processing.

@params
- ctx: lifecycle context
- fromBlock: inclusive start block
- toBlock: inclusive end block

@returns
- []types.Event
- error
*/
func (l *Listener) Poll(ctx context.Context, fromBlock int64, toBlock int64) ([]types.Event, error) {
	return []types.Event{}, nil
}
