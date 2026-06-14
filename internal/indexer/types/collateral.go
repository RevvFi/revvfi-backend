package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

/*@
file: collateral.go
package: types

purpose:
Collateral escrow events.
*/

type CollateralDepositedEvent struct {
	Borrower common.Address
	Amount   *big.Int
	Total    *big.Int
}

type CollateralWithdrawnEvent struct {
	Borrower common.Address
	Amount   *big.Int
	Total    *big.Int
}