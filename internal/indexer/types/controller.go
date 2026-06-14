// internal/indexer/types/controller.go
package types

import (
	"github.com/ethereum/go-ethereum/common"
)


/*@
 * ControllerRemovedEvent
 * @event ControllerRemoved(address) - NO INDEXED PARAMETERS
 * @contract RevvFiArchController
 * @data: controller address only
 */
type ControllerRemovedEvent struct {
	Controller common.Address
}

/*@
 * MarketRegisteredEvent
 * @event MarketRegistered(address) - NO INDEXED PARAMETERS
 * @contract RevvFiArchController (PositionNFT also emits this)
 * @data: market address only
 */
type MarketRegisteredEvent struct {
	Market common.Address
}

/*@
 * MarketUnregisteredEvent
 * @event MarketUnregistered(address) - NO INDEXED PARAMETERS
 * @contract RevvFiArchController (PositionNFT also emits this)
 * @data: market address only
 */
type MarketUnregisteredEvent struct {
	Market common.Address
}

/*@
 * ControllerFactoryAddedEvent
 * @event ControllerFactoryAdded(address) - NO INDEXED PARAMETERS
 * @contract RevvFiArchController
 * @data: factory address only
 */
type ControllerFactoryAddedEvent struct {
	Factory common.Address
}

/*@
 * ControllerFactoryRemovedEvent
 * @event ControllerFactoryRemoved(address) - NO INDEXED PARAMETERS
 * @contract RevvFiArchController
 * @data: factory address only
 */
type ControllerFactoryRemovedEvent struct {
	Factory common.Address
}

/*@
 * MarketRemovedEvent
 * @event MarketRemoved(address) - NO INDEXED PARAMETERS
 * @contract RevvFiArchController
 * @data: market address only
 */
type MarketRemovedEvent struct {
	Market common.Address
}

/*@
 * AssetBlacklistedEvent
 * @event AssetBlacklisted(address) - NO INDEXED PARAMETERS
 * @contract RevvFiArchController
 * @data: asset address only
 */
type AssetBlacklistedEvent struct {
	Asset common.Address
}

/*@
 * AssetPermittedEvent
 * @event AssetPermitted(address) - NO INDEXED PARAMETERS
 * @contract RevvFiArchController
 * @data: asset address only
 */
type AssetPermittedEvent struct {
	Asset common.Address
}

/*@
 * MarketAddedEvent
 * @desc Emitted when market is added to ArchController
 * @contract RevvFiArchController
 */
type MarketAddedEvent struct {
	Market     common.Address
	Controller common.Address
}

/*@
 * InitializedEvent
 * @desc Emitted when contract is initialized (from OpenZeppelin)
 */
type InitializedEvent struct {
	Version uint64
}

/*@
 * ContractsSetEvent
 * @desc Emitted when market contracts are set (from RevvFiMarket)
 */
type ContractsSetEvent struct {
	Escrow             common.Address
	OfferBook          common.Address
	PositionNFT        common.Address
	Liquidator         common.Address
	ReputationRegistry common.Address
}

type ControllerAddedEvent struct {
    ControllerFactory common.Address

}