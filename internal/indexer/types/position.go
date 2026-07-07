package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

/*@
file: position.go
package: types

purpose:
Position NFT events.
*/

type PositionMintedEvent struct {
	TokenID   *big.Int
	Lender    common.Address
	Market    common.Address
	Principal *big.Int
	APR       *big.Int
	Seniority uint8
}

/*@
event: PositionSettled
description:
Position settled and claimable.
actual_topics: 2 (signature + tokenId)
actual_data: 32 bytes (claimableAmount)
*/
type PositionSettledEvent struct {
	TokenID         *big.Int
	ClaimableAmount *big.Int
}

type PositionTransferEvent struct {
	From    common.Address
	To      common.Address
	TokenID *big.Int
}

/*@
 * PositionClaimedEvent
 * @desc Emitted when lender claims repaid funds
 * @contract RevvFiMarket / RevvFiPositionNFT
 * @fields
 *   - TokenID: NFT token ID of the position
 *   - Lender: Address of lender claiming funds
 *   - Amount: Amount claimed (principal + interest)
 * @trigger After borrower repayment and when lender calls claimFunds()
 */
type PositionClaimedEvent struct {
	TokenID *big.Int
	Lender  common.Address
	Amount  *big.Int
}

/*@
 * PositionRedeemedEvent
 * @desc Emitted when a position NFT is redeemed
 * @contract RevvFiPositionNFT
 * @topics
 *   - topics[0]: Event signature hash
 *   - topics[1]: tokenId (indexed)
 * @data
 *   - bytes[0:32]: principal (uint256)
 *   - bytes[32:64]: interest (uint256)
 * @trigger When lender redeems their position NFT
 */
type PositionRedeemedEvent struct {
	TokenID   *big.Int
	Principal *big.Int
	Interest  *big.Int
}

/*@
 * LenderRepaidEvent
 * @desc Emitted per-position on a partial repayment (position stays active).
 *       Full settlement instead goes through PositionSettled/PositionRedeemed.
 * @contract RevvFiMarket
 * @topics
 *   - topics[0]: Event signature hash
 *   - topics[1]: lender (indexed)
 *   - topics[2]: positionId (indexed)
 * @data
 *   - bytes[0:32]: newPrincipal (uint256) - absolute remaining principal
 *   - bytes[32:64]: claimableShare (uint256) - incremental amount credited this round
 */
type LenderRepaidEvent struct {
	Lender         common.Address
	PositionID     *big.Int
	NewPrincipal   *big.Int
	ClaimableShare *big.Int
}