package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

/*@
file: market.go
package: types
layer: indexer

purpose:
Decoded events emitted by market contracts.

contracts:
- RevvFiMarket
*/


/*@
event: Repay
description:
Borrower repays debt.
actual_data: 96 bytes (amount + interestPaid + principalPaid)
*/
type RepayEvent struct {
	Market        common.Address
	Borrower      common.Address
	Amount        *big.Int
	InterestPaid  *big.Int
	PrincipalPaid *big.Int
}

/*@
event: MarketPaused
description:
Emergency pause activated.
*/
type MarketPausedEvent struct {
	Market   common.Address
	PausedBy common.Address
	Reason   string
}

/*@
event: MarketUnpaused
description:
Market resumed after pause.
*/
type MarketUnpausedEvent struct {
	Market     common.Address
	UnpausedBy common.Address
}


/*@
 * LiquidationExecutedEvent
 * @desc Emitted when liquidation is executed
 * @contract RevvFiMarket
 * @fields
 *   - Borrower: Address of borrower being liquidated
 *   - DebtAmount: Amount of debt being liquidated
 *   - CollateralAmount: Amount of collateral seized
 * @trigger When collateral ratio falls below liquidation threshold
 */
type LiquidationExecutedEvent struct {
	Borrower        common.Address
	DebtAmount      *big.Int
	CollateralAmount *big.Int
}

/*@
 * MarketClosedEvent
 * @desc Emitted when a market is closed
 * @contract RevvFiMarket
 * @topics
 *   - topics[0]: Event signature hash
 *   - topics[1]: borrower address (indexed)
 * @data
 *   - bytes[0:32]: timestamp (uint256)
 * @trigger When market is closed by admin or automatically
 */
type MarketClosedEvent struct {
	Market    common.Address
	Borrower  common.Address
	Timestamp *big.Int
}


type BorrowEvent struct {
    Market       common.Address
    Borrower     common.Address
    Amount       *big.Int
    WeightedAPR  *big.Int
}

type DrawdownExecutedEvent struct {
    TotalAmount  *big.Int
    WeightedAPR  *big.Int
    PositionIds  []*big.Int
}

// LiquidationStartedMarket is emitted by RevvFiMarket when liquidation begins
type LiquidationStartedMarketEvent struct {
    Market   common.Address
    Borrower common.Address
}

// LiquidationEndedMarket is emitted by RevvFiMarket when liquidation completes
type LiquidationEndedMarketEvent struct {
    Market   common.Address
    Borrower common.Address
}
