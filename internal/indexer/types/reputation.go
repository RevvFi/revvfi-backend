// internal/indexer/types/reputation.go
package types

import (
	"math/big"
	"github.com/ethereum/go-ethereum/common"
)

/*@
 * SuccessfulRepaymentRecordedEvent
 * @desc Emitted when a successful repayment is recorded
 * @contract ReputationRegistry
 * @fields
 *   - Borrower: Address of borrower who made repayment
 *   - Amount: Amount repaid (in borrow asset decimals)
 * @trigger When borrower makes a repayment
 * @effect Increases successfulLoans count and updates reputation score
 */
type SuccessfulRepaymentRecordedEvent struct {
	Borrower common.Address
	Amount   *big.Int
}

/*@
 * DefaultRecordedEvent
 * @desc Emitted when a borrower defaults on a loan
 * @contract ReputationRegistry
 * @fields
 *   - Borrower: Address of borrower who defaulted
 *   - Amount: Amount defaulted (in borrow asset decimals)
 * @trigger When liquidation occurs and debt is not fully recovered
 * @effect Increases defaultedLoans count and decreases reputation score by 50 points per default
 */
type DefaultRecordedEvent struct {
	Borrower common.Address
	Amount   *big.Int
}

type BorrowActivityRecordedEvent struct {
    Borrower common.Address
    Amount   *big.Int
}