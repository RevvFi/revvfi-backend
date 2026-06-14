package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

/*@
file: withdrawal.go
package: types

purpose:
Liquidity queue events.
*/

type WithdrawalRequestedEvent struct {
	RequestID   *big.Int
	Lender      common.Address
	PositionID  *big.Int
	Market      common.Address
	Amount      *big.Int
	EpochNumber *big.Int
}

type WithdrawalClaimedEvent struct {
	RequestID *big.Int
	Lender    common.Address
	Amount    *big.Int
}

type WithdrawalCancelledEvent struct {
	RequestID *big.Int
	Lender    common.Address
}

type EpochProcessedEvent struct {
	EpochNumber    *big.Int
	TotalRequested *big.Int
	TotalFulfilled *big.Int
}