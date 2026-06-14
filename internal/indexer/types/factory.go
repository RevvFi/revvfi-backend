package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

/*@
file: factory.go
package: types
layer: indexer

purpose:
Decoded events emitted by the Factory contract.

contracts:
- RevvFiFactory

notes:
- Used during market creation.
- Creates market records inside PostgreSQL.
*/

/*@
event: MarketDeployed
description:
Emitted whenever a borrower deploys a new isolated market.
*/
type MarketDeployedEvent struct {
	MarketAddress        common.Address
	Borrower             common.Address
	BorrowAsset          common.Address
	CollateralAsset      common.Address
	Oracle               common.Address
	MinCollateralRatio   *big.Int
}

/*@
event: CoreContractsSet
description:
Factory configuration updated.
*/
type CoreContractsSetEvent struct {
	ArchController     common.Address
	PositionNFT        common.Address
	Liquidator         common.Address
	ReputationRegistry common.Address
}

/*@
event: DeploymentFeeUpdated
description:
Deployment fee changed by protocol admin.
*/
type DeploymentFeeUpdatedEvent struct {
	OldFee *big.Int
	NewFee *big.Int
}



/*@
 * ArchControllerUpdateRequestedEvent
 * @desc Emitted when ArchController update is requested (2-day timelock starts)
 * @contract RevvFiFactory
 * @fields
 *   - NewArchController: Proposed new ArchController address
 * @access onlyOwner
 * @timelock 172800 seconds (2 days)
 */
type ArchControllerUpdateRequestedEvent struct {
	NewArchController common.Address
}

/*@
 * ArchControllerUpdatedEvent
 * @desc Emitted after timelock expires and ArchController is updated
 * @contract RevvFiFactory
 * @fields
 *   - OldArchController: Previous ArchController address
 *   - NewArchController: New ArchController address
 * @access onlyOwner (after timelock)
 */
type ArchControllerUpdatedEvent struct {
	OldArchController common.Address
	NewArchController common.Address
}

/*@
 * FeeUpdatedEvent
 * @desc Emitted when deployment fee is updated
 * @contract RevvFiFactory
 * @fields
 *   - OldFee: Previous fee amount in wei
 *   - NewFee: New fee amount in wei
 * @access onlyOwner
 */
type FeeUpdatedEvent struct {
	OldFee *big.Int
	NewFee *big.Int
}

/*@
 * ImplementationsSetEvent
 * @desc Emitted when implementation contracts are set
 * @contract RevvFiFactory
 * @fields
 *   - MarketImpl: Market implementation address
 *   - EscrowImpl: CollateralEscrow implementation address
 *   - OfferBookImpl: OfferBook implementation address
 *   - LiquidityQueueImpl: LiquidityQueue implementation address
 * @access onlyOwner (one-time setup)
 */
type ImplementationsSetEvent struct {
	MarketImpl        common.Address
	EscrowImpl        common.Address
	OfferBookImpl     common.Address
	LiquidityQueueImpl common.Address
}

/*@
 * OwnershipTransferredEvent
 * @desc Emitted when factory ownership is transferred
 * @contract RevvFiFactory (from OpenZeppelin Ownable)
 * @fields
 *   - PreviousOwner: Previous owner address
 *   - NewOwner: New owner address
 * @access onlyOwner
 */
type OwnershipTransferredEvent struct {
	PreviousOwner common.Address
	NewOwner      common.Address
}
