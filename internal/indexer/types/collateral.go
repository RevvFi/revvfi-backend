package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

/*@
file: collateral.go
package: types

purpose:
Events emitted by the RevvFiCollateralEscrow contract.
Each market's embedded escrow emits these events when borrowers
deposit or withdraw collateral, and when admin parameters change.
*/

/*@
 * CollateralDepositedEvent
 * @desc Emitted when a borrower deposits collateral into the escrow
 * @contract RevvFiCollateralEscrow
 * @fields
 *   - Borrower: Address of borrower depositing collateral
 *   - Amount: Amount deposited in this transaction
 *   - Total: New total collateral balance for this borrower (authoritative)
 * @trigger Called when borrower calls depositCollateral()
 */
type CollateralDepositedEvent struct {
	Borrower common.Address
	Amount   *big.Int
	Total    *big.Int
}

/*@
 * CollateralWithdrawnEvent
 * @desc Emitted when a borrower withdraws excess collateral from the escrow
 * @contract RevvFiCollateralEscrow
 * @fields
 *   - Borrower: Address of borrower withdrawing collateral
 *   - Amount: Amount withdrawn in this transaction
 *   - Total: New total collateral balance for this borrower (authoritative)
 * @trigger Called when borrower calls withdrawCollateral() and ratio remains healthy
 */
type CollateralWithdrawnEvent struct {
	Borrower common.Address
	Amount   *big.Int
	Total    *big.Int
}

/*@
 * CollateralLiquidatedEvent
 * @desc Emitted when collateral is seized during liquidation
 * @contract RevvFiCollateralEscrow
 * @fields
 *   - Borrower: Address of borrower being liquidated
 *   - Amount: Amount of collateral seized
 *   - Proceeds: Proceeds from collateral sale (in borrow asset)
 * @trigger Called by the Liquidator contract during auction settlement
 */
type CollateralLiquidatedEvent struct {
	Borrower common.Address
	Amount   *big.Int
	Proceeds *big.Int
}

/*@
 * MinCollateralRatioUpdatedEvent
 * @desc Emitted when the minimum collateral ratio for a market is updated
 * @contract RevvFiCollateralEscrow
 * @fields
 *   - OldRatio: Previous minimum collateral ratio in basis points
 *   - NewRatio: New minimum collateral ratio in basis points (11000 = 110%)
 * @note Escrow address must be resolved to market address via ContractsSet mapping
 * @trigger Called by market admin via setMinCollateralRatio()
 */
type MinCollateralRatioUpdatedEvent struct {
	OldRatio *big.Int
	NewRatio *big.Int
}

/*@
 * LiquidationThresholdUpdatedEvent
 * @desc Emitted when the liquidation threshold for a market is updated
 * @contract RevvFiCollateralEscrow
 * @fields
 *   - OldThreshold: Previous liquidation threshold in basis points
 *   - NewThreshold: New liquidation threshold in basis points (9500 = 95%)
 * @note Escrow address must be resolved to market address via ContractsSet mapping
 * @trigger Called by market admin via setLiquidationThreshold()
 */
type LiquidationThresholdUpdatedEvent struct {
	OldThreshold *big.Int
	NewThreshold *big.Int
}