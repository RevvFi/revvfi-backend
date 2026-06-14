package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

/*@
file: borrower.go
package: types

purpose:
Events from:
- ReputationRegistry
- ArchController
*/
/*@
 * BorrowerRegisteredEvent
 * @event BorrowerRegistered(address) - NO INDEXED PARAMETERS
 * @contract ReputationRegistry
 * @data: borrower address only
 */
type BorrowerRegisteredEvent struct {
	Borrower common.Address
}

type ReputationBorrowerRemovedEvent struct {
	Borrower common.Address
}

// =====================================================
// Arch Controller
// =====================================================

/*@
 * ArchBorrowerAddedEvent
 * @event BorrowerAdded(address) - NO INDEXED PARAMETERS
 * @contract RevvFiArchController
 * @data: borrower address only
 */
type ArchBorrowerAddedEvent struct {
	Borrower common.Address
}

/*@
 * ArchBorrowerRemovedEvent
 * @event BorrowerRemoved(address) - NO INDEXED PARAMETERS
 * @contract RevvFiArchController
 * @data: borrower address only
 */
type ArchBorrowerRemovedEvent struct {
	Borrower common.Address
}

/*@
 * ArchOwnerUpdatedEvent
 * @desc Emitted when ArchController ownership is transferred
 * @contract RevvFiArchController
 * @fields
 *   - OldOwner: Previous owner address
 *   - NewOwner: New owner address
 * @access onlyOwner
 */
type ArchOwnerUpdatedEvent struct {
    OldOwner common.Address
    NewOwner common.Address
}

/*@
 * ArchTimelockUpdatedEvent
 * @desc Emitted when timelock delay is updated
 * @contract RevvFiArchController
 * @fields
 *   - OldDelay: Previous delay in seconds
 *   - NewDelay: New delay in seconds
 * @default 172800 seconds (2 days)
 */
type ArchTimelockUpdatedEvent struct {
    OldDelay *big.Int
    NewDelay *big.Int
}
// =====================================================
// REPUTATION REGISTRY EVENTS
// =====================================================

/*@
 * ReputationScoreUpdatedEvent
 * @desc Emitted when a borrower's reputation score changes
 * @contract ReputationRegistry
 * @fields
 *   - Borrower: Address of the borrower whose score changed
 *   - OldScore: Previous reputation score (0-1000)
 *   - NewScore: New reputation score (0-1000)
 * @formula: NewScore = (SuccessfulLoans * 1000 / TotalLoans) - (Defaults * 50)
 * @range: 0-1000, clamped
 * @trigger: After successful repayment or default
 */
type ReputationScoreUpdatedEvent struct {
    Borrower common.Address
    OldScore *big.Int
    NewScore *big.Int
}