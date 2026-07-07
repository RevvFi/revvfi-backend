// internal/indexer/decoder/decoder.go
package decoder

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	indexertypes "github.com/Revvfi/revvfi-backend/internal/indexer/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	goEthtypes "github.com/ethereum/go-ethereum/core/types"
)

/*@
struct: EventDecoder

description:
Central ABI decoder used by the RevvFi indexer.

responsibilities:
- Load protocol ABIs from Foundry artifacts
- Decode raw Ethereum logs
- Convert logs into strongly typed event structs
- Route decoding based on event signature topics

contracts:
- RevvFiFactory
- RevvFiMarket
- RevvFiOfferBook
- RevvFiPositionNFT
- RevvFiLiquidator
- RevvFiLiquidityQueue
- ReputationRegistry

lifecycle:
1. Created during indexer startup
2. Loads all protocol ABIs
3. Receives logs from poller
4. Decodes logs into typed events
5. Returns typed events to handlers
*/
type EventDecoder struct {
    factoryABI     *abi.ABI
    marketABI      *abi.ABI
    offerBookABI   *abi.ABI
    positionNFTABI *abi.ABI
    liquidatorABI  *abi.ABI
    queueABI       *abi.ABI
    reputationABI  *abi.ABI
    eventSignatures map[string]common.Hash
}


/*
@function initEventSignatures

@desc
Initializes the in-memory mapping of protocol event names to their
corresponding Ethereum event signature hashes (Topic0 values).

@responsibilities
- Load all configured event signature hashes from EventTopics
- Convert hexadecimal topic strings into common.Hash values
- Populate decoder lookup table for fast event identification

@why
Ethereum logs are identified by Topic0, which is the keccak256 hash
of the event declaration signature.

Example:
MarketDeployed(address,address,address,address)

↓

keccak256(...)

↓

0xabc123...

The indexer compares incoming log.Topic0 values against this map
to determine which protocol event was emitted and which decoder
should process the log.

@used_by
- EventDecoder.Decode()
- Blockchain event indexer
- Log processing pipeline

@note
If a signature in EventTopics does not match the actual Solidity
event definition, emitted events will not be recognized by the
indexer and will be skipped.
*/
func (d *EventDecoder) initEventSignatures() {
	d.eventSignatures = make(map[string]common.Hash)

	for name, hashStr := range EventTopics {
		d.eventSignatures[name] = common.HexToHash(hashStr)
	}
}

func NewEventDecoder(artifactPath string) (*EventDecoder, error) {
    decoder := &EventDecoder{}
    
    // Load each contract ABI
    if err := decoder.loadABIs(artifactPath); err != nil {
        return nil, err
    }
     decoder.initEventSignatures()

    return decoder, nil
}

func (d *EventDecoder) loadABIs(artifactPath string) error {
    // Load Factory ABI
    factoryABI, err := LoadABI("RevvFiFactory", artifactPath)
    if err != nil {
        return err
    }
	/*@ loading Factory ABI*/
	parsedABI, err := abi.JSON(strings.NewReader(string(factoryABI)))
    if err != nil {
    return err
   }
    d.factoryABI = &parsedABI
    
    // Load Market ABI
    marketABI, err := LoadABI("RevvFiMarket", artifactPath)
    if err != nil {
        return err
    }
    parsedMarketABI, err := abi.JSON(strings.NewReader(string(marketABI)))
    if err != nil {
        return err
    }
    d.marketABI = &parsedMarketABI
    
    // Load OfferBook ABI
    offerABI, err := LoadABI("RevvFiOfferBook", artifactPath)
    if err != nil {
        return err
    }
    parsedOfferABI, err := abi.JSON(strings.NewReader(string(offerABI)))
    if err != nil {
        return err
    }
    d.offerBookABI = &parsedOfferABI

    // Load PositionNFT ABI
    posABI, err := LoadABI("RevvFiPositionNFT", artifactPath)
    if err != nil {
        return err
    }
    parsedPosABI, err := abi.JSON(strings.NewReader(string(posABI)))
    if err != nil {
        return err
    }
    d.positionNFTABI = &parsedPosABI

    // Load Liquidator ABI
    liqABI, err := LoadABI("RevvFiLiquidator", artifactPath)
    if err != nil {
        return err
    }
    parsedLiqABI, err := abi.JSON(strings.NewReader(string(liqABI)))
    if err != nil {
        return err
    }
    d.liquidatorABI = &parsedLiqABI
    
    
    // Load LiquidityQueue ABI
    queueABI, err := LoadABI("RevvFiLiquidityQueue", artifactPath)
    if err != nil {
        return err
    }
    parsedQueueABI, err := abi.JSON(strings.NewReader(string(queueABI)))
    if err != nil {
        return err
    }
    d.queueABI = &parsedQueueABI

    // Load ReputationRegistry ABI
    repABI, err := LoadABI("ReputationRegistry", artifactPath)
    if err != nil {
        return err
    }
    parsedRepABI, err := abi.JSON(strings.NewReader(string(repABI)))
    if err != nil {
        return err
    }
    d.reputationABI = &parsedRepABI
   
    return nil
}

func (d *EventDecoder) Decode(log goEthtypes.Log) (interface{}, error) {
    eventName := GetEventName(log.Topics[0].Hex())
    
    switch eventName {
    // Factory Events
    case "MarketDeployed":
        return d.decodeMarketDeployed(log)
    case "CoreContractsSet":
        return d.decodeCoreContractsSet(log)
    case "ArchControllerUpdateRequested":
        return d.decodeArchControllerUpdateRequested(log)  
    case "ArchControllerUpdated":
        return d.decodeArchControllerUpdated(log)          
    case "FeeUpdated":
        return d.decodeFeeUpdated(log)                    
    case "ImplementationsSet":
        return d.decodeImplementationsSet(log)            
    case "OwnershipTransferred":
        return d.decodeOwnershipTransferred(log)          
    case "DeploymentFeeUpdated":
      return d.decodeDeploymentFeeUpdated(log)
    case "DrawdownExecutedOffer":
       return d.decodeDrawdownExecutedOffer(log)
    case "DrawdownExecuted":
       return d.decodeDrawdownExecuted(log)
    case "BorrowActivityRecorded":
       return d.decodeBorrowActivityRecorded(log)
    
    // Market Events
    case "Borrow":
        return d.decodeBorrow(log)
    case "Repay":
        return d.decodeRepay(log)
    case "LenderRepaid":
        return d.decodeLenderRepaid(log)
    case "CollateralDeposited":
        return d.decodeCollateralDeposited(log)
    case "CollateralWithdrawn":
        return d.decodeCollateralWithdrawn(log)
    case "CollateralLiquidated":
        return d.decodeCollateralLiquidated(log)
    case "MinCollateralRatioUpdated":
        return d.decodeMinCollateralRatioUpdated(log)
    case "LiquidationThresholdUpdated":
        return d.decodeLiquidationThresholdUpdated(log)
    case "MarketPaused":
        return d.decodeMarketPaused(log)
    case "MarketUnpaused":
        return d.decodeMarketUnpaused(log)
    case "LiquidationExecuted":
        return d.decodeLiquidationExecuted(log)
    case "LiquidationStartedMarket":
        return d.decodeLiquidationStartedMarket(log)
    case "LiquidationEndedMarket":
        return d.decodeLiquidationEndedMarket(log)
    case "MarketClosedEvent":
        return d.decodeMarketClosedEvent(log)

    // OfferBook Events
    case "OfferSubmitted":
        return d.decodeOfferSubmitted(log)
    case "OfferCancelled":
        return d.decodeOfferCancelled(log)
    case "OfferFilled":
        return d.decodeOfferFilled(log)
    case "OfferModified":
        return d.decodeOfferModified(log)
    
    // Position NFT Events
    case "PositionMinted":
        return d.decodePositionMinted(log)
    case "PositionSettled":
        return d.decodePositionSettled(log)
    case "PositionClaimed":
        return d.decodePositionClaimed(log)
    case "PositionRedeemed":
        return d.decodePositionRedeemed(log)
    case "Transfer":
        return d.decodeTransfer(log)
    
    // Liquidation Events
    case "AuctionCreated":
        return d.decodeAuctionCreated(log)
    case "BidPlaced":
        return d.decodeBidPlaced(log)
    case "AuctionSettled":
        return d.decodeAuctionSettled(log)
    case "AuctionCancelled":
    return d.decodeAuctionCancelled(log)
    
    // Withdrawal Events
    case "WithdrawalRequested":
        return d.decodeWithdrawalRequested(log)
    case "WithdrawalClaimed":
        return d.decodeWithdrawalClaimed(log)
    case "WithdrawalCancelled":
        return d.decodeWithdrawalCancelled(log)          
    case "EpochProcessed":
        return d.decodeEpochProcessed(log)
    
    // Reputation Events
    case "ReputationScoreUpdated":
        return d.decodeReputationScoreUpdated(log)
    case "BorrowerRegistered":
        return d.decodeBorrowerRegistered(log)
    case "SuccessfulRepaymentRecorded":
        return d.decodeSuccessfulRepaymentRecorded(log)   
    case "DefaultRecorded":
        return d.decodeDefaultRecorded(log)               
    
    case "MarketAdded":
         return d.decodeMarketAdded(log)
    case "Initialized":
         return d.decodeInitialized(log)
    case "ControllerFactoryAdded":
        return d.decodeControllerFactoryAdded(log)
    case "ControllerFactoryRemoved":
        return d.decodeControllerFactoryRemoved(log)
    // ArchController Events
    case "BorrowerAdded":
        return d.decodeBorrowerAdded(log)
    case "BorrowerRemoved":
        return d.decodeBorrowerRemovedArch(log)
    case "ControllerAdded":
        return d.decodeControllerAdded(log)
    case "ControllerRemoved":
        return d.decodeControllerRemoved(log)
    case "MarketRegistered":
        return d.decodeMarketRegistered(log)
    case "MarketUnregistered":
        return d.decodeMarketUnregistered(log)
    case "MarketRemoved":
        return d.decodeMarketRemoved(log)
    case "AssetBlacklisted":
        return d.decodeAssetBlacklisted(log)
    case "AssetPermitted":
        return d.decodeAssetPermitted(log)
    case "OwnerUpdated":
        return d.decodeArchControllerOwnerUpdated(log)
    case "TimelockUpdated":
       return d.decodeArchControllerTimelockUpdated(log)
    case "ContractsSet":
       return d.decodeContractsSet(log)
    default:
        return nil, fmt.Errorf("unknown event: %s", eventName)
    }
}

// =====================================================
// FACTORY EVENT DECODERS
// =====================================================

/*@
 * @method decodeMarketDeployed
 * @contract RevvFiFactory
 * @event MarketDeployed(address indexed market, address indexed borrower, address borrowAsset, address collateralAsset, address oracle, uint256 minCollateralRatio)
 *
 * @topics_structure (3 topics total):
 *   topics[0]: Event signature hash (keccak256 of event signature)
 *   topics[1]: market address (indexed) - 32 bytes, address right-aligned
 *   topics[2]: borrower address (indexed) - 32 bytes, address right-aligned
 *
 * @data_structure (128 bytes total):
 *   bytes[0:32]:   borrowAsset address - 32 bytes with 12-byte left padding, actual address at [12:32]
 *   bytes[32:64]:  collateralAsset address - 32 bytes with 12-byte left padding, actual address at [44:64]
 *   bytes[64:96]:  oracle address - 32 bytes with 12-byte left padding, actual address at [76:96]
 *   bytes[96:128]: minCollateralRatio (uint256) - 32 bytes
 *
 * @why_this_works:
 *   - Ethereum ABI encoding stores addresses as 32-byte values with 12 bytes of zero padding on the left
 *   - To extract an address from data, we skip the first 12 bytes and take the remaining 20 bytes
 *   - Example: bytes[0:32] contains an address, but actual address is at bytes[12:32]
 *   - For sequential addresses in data: first is [12:32], second is [44:64], third is [76:96]
 *
 * @fix_applied: Data length reduced from 160 to 128 bytes based on actual blockchain log data
 *               LiquidationThreshold parameter was removed from the event definition
 *
 * @status: ✓ WORKING (verified by user)
 */
func (d *EventDecoder) decodeMarketDeployed(log goEthtypes.Log) (*indexertypes.MarketDeployedEvent, error) {
    event := &indexertypes.MarketDeployedEvent{}

    /*@ Validate topic count - expecting 3 topics (signature + 2 indexed params) */
    if len(log.Topics) < 3 {
        return nil, fmt.Errorf("MarketDeployed: expected at least 3 topics, got %d", len(log.Topics))
    }

    /*@ Extract indexed parameters from topics
     * Topics store addresses as full 32-byte hex strings, can use HexToAddress directly */
    event.MarketAddress = common.HexToAddress(log.Topics[1].Hex())
    event.Borrower = common.HexToAddress(log.Topics[2].Hex())

    /*@ Validate data length - need 128 bytes for 3 addresses + 1 uint256 */
    if len(log.Data) < 128 {
        return nil, fmt.Errorf("MarketDeployed: insufficient data length")
    }

    /*@ Extract non-indexed parameters from data
     * Each address in data has 12 bytes of left padding, so we skip to byte 12 for actual address
     * Positions: [12:32] for first, [44:64] for second, [76:96] for third address */
    event.BorrowAsset = common.HexToAddress("0x" + hex.EncodeToString(log.Data[12:32]))
    event.CollateralAsset = common.HexToAddress("0x" + hex.EncodeToString(log.Data[44:64]))
    event.Oracle = common.HexToAddress("0x" + hex.EncodeToString(log.Data[76:96]))

    /*@ Extract uint256 value - no padding, full 32 bytes */
    event.MinCollateralRatio = new(big.Int).SetBytes(log.Data[96:128])

    return event, nil
}

/*@
 * @method decodeCoreContractsSet
 * @contract RevvFiFactory
 * @event CoreContractsSet(address archController, address positionNFT, address liquidator, address reputationRegistry)
 *
 * @topics_structure (1 topic total):
 *   topics[0]: Event signature hash (keccak256 of event signature)
 *
 * @data_structure (128 bytes total):
 *   bytes[0:32]:   archController address - 32 bytes with 12-byte left padding, actual address at [12:32]
 *   bytes[32:64]:  positionNFT address - 32 bytes with 12-byte left padding, actual address at [44:64]
 *   bytes[64:96]:  liquidator address - 32 bytes with 12-byte left padding, actual address at [76:96]
 *   bytes[96:128]: reputationRegistry address - 32 bytes with 12-byte left padding, actual address at [108:128]
 *
 * @why_no_indexed_params:
 *   This event has no indexed parameters, so only topic[0] (signature) is present
 *   All 4 contract addresses are stored in the data field
 *
 * @extraction_pattern:
 *   First address:  skip 12 bytes padding → [12:32]
 *   Second address: skip 32+12 bytes → [44:64]
 *   Third address:  skip 64+12 bytes → [76:96]
 *   Fourth address: skip 96+12 bytes → [108:128]
 */
func (d *EventDecoder) decodeCoreContractsSet(log goEthtypes.Log) (*indexertypes.CoreContractsSetEvent, error) {
    event := &indexertypes.CoreContractsSetEvent{}

    /*@ Validate data length - need 128 bytes for 4 addresses (4 × 32 bytes) */
    if len(log.Data) < 128 {
        return nil, fmt.Errorf("CoreContractsSet: insufficient data length")
    }

    /*@ Extract all 4 address parameters from data, skipping 12 bytes of padding for each */
    event.ArchController = common.HexToAddress("0x" + hex.EncodeToString(log.Data[12:32]))
    event.PositionNFT = common.HexToAddress("0x" + hex.EncodeToString(log.Data[44:64]))
    event.Liquidator = common.HexToAddress("0x" + hex.EncodeToString(log.Data[76:96]))
    event.ReputationRegistry = common.HexToAddress("0x" + hex.EncodeToString(log.Data[108:128]))

    return event, nil
}

/*@
 * @method decodeDeploymentFeeUpdated
 * @contract RevvFiFactory
 * @event DeploymentFeeUpdated(uint256 oldFee, uint256 newFee)
 *
 * @topics_structure (1 topic total):
 *   topics[0]: Event signature hash
 *
 * @data_structure (64 bytes total):
 *   bytes[0:32]:  oldFee (uint256) - 32 bytes
 *   bytes[32:64]: newFee (uint256) - 32 bytes
 *
 * @why_simple_extraction:
 *   uint256 values are stored as full 32-byte values with no padding
 *   Can directly use SetBytes on the 32-byte chunks
 */
func (d *EventDecoder) decodeDeploymentFeeUpdated(log goEthtypes.Log) (*indexertypes.DeploymentFeeUpdatedEvent, error) {
    event := &indexertypes.DeploymentFeeUpdatedEvent{}

    /*@ Validate data length - need 64 bytes for 2 uint256 values */
    if len(log.Data) < 64 {
        return nil, fmt.Errorf("DeploymentFeeUpdated: insufficient data length")
    }

    /*@ Extract uint256 values - no padding needed for integers */
    event.OldFee = new(big.Int).SetBytes(log.Data[0:32])
    event.NewFee = new(big.Int).SetBytes(log.Data[32:64])

    return event, nil
}

// =====================================================
// MARKET EVENT DECODERS
// =====================================================

/*@
 * @method decodeBorrow
 * @contract RevvFiMarket
 * @event Borrow(address indexed borrower, uint256 amount, uint256 weightedApr)
 *
 * @topics_structure (2 topics total):
 *   topics[0]: Event signature hash
 *   topics[1]: borrower (indexed)
 *
 * @data_structure (64 bytes total):
 *   bytes[0:32]: amount (uint256)
 *   bytes[32:64]: weightedApr (uint256)
 */
func (d *EventDecoder) decodeBorrow(log goEthtypes.Log) (*indexertypes.BorrowEvent, error) {
    event := &indexertypes.BorrowEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("Borrow: expected at least 2 topics, got %d", len(log.Topics))
    }
    
    event.Borrower = common.HexToAddress(log.Topics[1].Hex())
    
    if len(log.Data) < 64 {
        return nil, fmt.Errorf("Borrow: insufficient data length, need 64 bytes, got %d", len(log.Data))
    }
    
    event.Amount = new(big.Int).SetBytes(log.Data[0:32])
    event.WeightedAPR = new(big.Int).SetBytes(log.Data[32:64])
    event.Market = log.Address

    return event, nil
}

/*@
 * @method decodeRepay
 * @contract RevvFiMarket
 * @event Repay(address indexed borrower, uint256 amount, uint256 interestPaid, uint256 principalPaid)
 *
 * @topics_structure (2 topics total):
 *   topics[0]: Event signature hash
 *   topics[1]: borrower address (indexed)
 *
 * @data_structure (96 bytes total):
 *   bytes[0:32]:   amount (uint256)
 *   bytes[32:64]:  interestPaid (uint256)
 *   bytes[64:96]:  principalPaid (uint256)
 *
 * @fix_applied: Data length changed from 128 to 96 bytes
 *               Removed TotalDebt field that was not in actual blockchain data
 */
func (d *EventDecoder) decodeRepay(log goEthtypes.Log) (*indexertypes.RepayEvent, error) {
    event := &indexertypes.RepayEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("Repay: expected at least 2 topics")
    }

    event.Borrower = common.HexToAddress(log.Topics[1].Hex())

    if len(log.Data) < 96 {
        return nil, fmt.Errorf("Repay: insufficient data length, expected 96 bytes, got %d", len(log.Data))
    }

    event.Amount = new(big.Int).SetBytes(log.Data[0:32])
    event.InterestPaid = new(big.Int).SetBytes(log.Data[32:64])
    event.PrincipalPaid = new(big.Int).SetBytes(log.Data[64:96])
    event.Market = log.Address

    return event, nil
}

/*
@method decodeMarketPaused
@desc Decodes MarketPaused event
*/
func (d *EventDecoder) decodeMarketPaused(log goEthtypes.Log) (*indexertypes.MarketPausedEvent, error) {
    event := &indexertypes.MarketPausedEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("MarketPaused: expected at least 2 topics")
    }
    
    event.Market = common.HexToAddress(log.Topics[1].Hex())
    
    if len(log.Topics) > 2 {
        event.PausedBy = common.HexToAddress(log.Topics[2].Hex())
    }
    
    if len(log.Data) > 0 {
        // Reason string (if present)
        event.Reason = string(log.Data)
    }
    
    return event, nil
}

/*
@method decodeMarketUnpaused
@desc Decodes MarketUnpaused event
*/
func (d *EventDecoder) decodeMarketUnpaused(log goEthtypes.Log) (*indexertypes.MarketUnpausedEvent, error) {
    event := &indexertypes.MarketUnpausedEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("MarketUnpaused: expected at least 2 topics")
    }
    
    event.Market = common.HexToAddress(log.Topics[1].Hex())
    
    if len(log.Topics) > 2 {
        event.UnpausedBy = common.HexToAddress(log.Topics[2].Hex())
    }
    
    return event, nil
}

// =====================================================
// OFFERBOOK EVENT DECODERS
// =====================================================

/*
@method decodeOfferSubmitted
@desc Decodes OfferSubmitted event from RevvFiOfferBook
*/
func (d *EventDecoder) decodeOfferSubmitted(log goEthtypes.Log) (*indexertypes.OfferSubmittedEvent, error) {
    event := &indexertypes.OfferSubmittedEvent{}
      /*@fix : OfferSubmitted has 3 topics (signature + offerId + lender)
     Market is NOT indexed - it's in the data*/
    if len(log.Topics) < 3 {
        return nil, fmt.Errorf("OfferSubmitted: expected 4 topics, got %d", len(log.Topics))
    }
    
    // Indexed parameters
    event.OfferID = new(big.Int).SetBytes(log.Topics[1].Bytes())
    event.Lender = common.HexToAddress(log.Topics[2].Hex())
    //event.Market = common.HexToAddress(log.Topics[3].Hex())
    
    if len(log.Data) < 128 {
        return nil, fmt.Errorf("OfferSubmitted: insufficient data length")
    }
    
    // data layout: [amount(32)][apr(32)][seniority(32, right-aligned uint8)][expiry(32)]
    event.Amount = new(big.Int).SetBytes(log.Data[0:32])
    event.APR = new(big.Int).SetBytes(log.Data[32:64])
    event.Seniority = log.Data[95] // uint8 is right-aligned in 32-byte ABI slot
    event.Expiry = new(big.Int).SetBytes(log.Data[96:128])
    // Market is not in OfferSubmitted event; use the emitting contract address as a proxy
    event.Market = log.Address

    return event, nil
}

/*
@method decodeOfferCancelled
@desc Decodes OfferCancelled event
*/
func (d *EventDecoder) decodeOfferCancelled(log goEthtypes.Log) (*indexertypes.OfferCancelledEvent, error) {
    event := &indexertypes.OfferCancelledEvent{}

    if len(log.Topics) < 3 {
        return nil, fmt.Errorf("OfferCancelled: insufficient topics")
    }
    
    event.OfferID = new(big.Int).SetBytes(log.Topics[1].Bytes())
    event.Lender = common.HexToAddress(log.Topics[2].Hex())

    if len(log.Topics) > 3 {
        event.Market = common.HexToAddress(log.Topics[3].Hex())
    }

    if len(log.Data) >= 32 {
        event.RemainingAmount = new(big.Int).SetBytes(log.Data[0:32])
    }

    return event, nil
}

/*@
 * @method decodeOfferFilled
 * @contract RevvFiOfferBook
 * @event OfferFilled(uint256 indexed offerId, address indexed lender, uint256 amountFilled)
 *
 * @topics_structure (3 topics total):
 *   topics[0]: Event signature hash
 *   topics[1]: offerId (indexed)
 *   topics[2]: lender (indexed)
 *
 * @data_structure (32 bytes total):
 *   bytes[0:32]: amountFilled (uint256)
 */
func (d *EventDecoder) decodeOfferFilled(log goEthtypes.Log) (*indexertypes.OfferFilledEvent, error) {
    event := &indexertypes.OfferFilledEvent{}

    if len(log.Topics) < 3 {
        return nil, fmt.Errorf("OfferFilled: expected at least 3 topics, got %d", len(log.Topics))
    }
    
    event.OfferID = new(big.Int).SetBytes(log.Topics[1].Bytes())
    event.Lender = common.HexToAddress(log.Topics[2].Hex())

    if len(log.Data) < 32 {
        return nil, fmt.Errorf("OfferFilled: insufficient data length, need 32 bytes, got %d", len(log.Data))
    }

    event.Amount = new(big.Int).SetBytes(log.Data[0:32])

    return event, nil
}

/*
@method decodeOfferModified
@desc Decodes OfferModified event
*/
func (d *EventDecoder) decodeOfferModified(log goEthtypes.Log) (*indexertypes.OfferModifiedEvent, error) {
    event := &indexertypes.OfferModifiedEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("OfferModified: insufficient topics")
    }

    event.OfferID = new(big.Int).SetBytes(log.Topics[1].Bytes())

    if len(log.Topics) > 2 {
        event.Lender = common.HexToAddress(log.Topics[2].Hex())
    }

    if len(log.Data) < 96 {
        return nil, fmt.Errorf("OfferModified: insufficient data length")
    }
    
    event.NewAmount = new(big.Int).SetBytes(log.Data[0:32])
    event.NewAPR = new(big.Int).SetBytes(log.Data[32:64])
    event.NewExpiry = new(big.Int).SetBytes(log.Data[64:96])
    
    return event, nil
}

// =====================================================
// POSITION NFT EVENT DECODERS
// =====================================================

/*
@method decodePositionMinted
@desc Decodes PositionMinted event from RevvFiPositionNFT
*/
func (d *EventDecoder) decodePositionMinted(log goEthtypes.Log) (*indexertypes.PositionMintedEvent, error) {
    event := &indexertypes.PositionMintedEvent{}

    if len(log.Topics) < 4 {
        return nil, fmt.Errorf("PositionMinted: expected 4 topics, got %d", len(log.Topics))
    }
    
    event.TokenID = new(big.Int).SetBytes(log.Topics[1].Bytes())
    event.Lender = common.HexToAddress(log.Topics[2].Hex())
    event.Market = common.HexToAddress(log.Topics[3].Hex())
    
    if len(log.Data) < 64 {
        return nil, fmt.Errorf("PositionMinted: insufficient data length")
    }
    
    event.Principal = new(big.Int).SetBytes(log.Data[0:32])
    event.APR = new(big.Int).SetBytes(log.Data[32:64])
    event.Seniority = uint8(log.Data[64])
    
    return event, nil
}

/*@
 * @method decodePositionSettled
 * @contract RevvFiPositionNFT
 * @event PositionSettled(uint256 indexed tokenId, uint256 claimableAmount)
 *
 * @topics_structure (2 topics total):
 *   topics[0]: Event signature hash
 *   topics[1]: tokenId (indexed)
 *
 * @data_structure (32 bytes total):
 *   bytes[0:32]: claimableAmount (uint256)
 *
 * @fix_applied: Topic count changed from 3 to 2
 *               Removed Lender field that was not in actual blockchain data
 */
func (d *EventDecoder) decodePositionSettled(log goEthtypes.Log) (*indexertypes.PositionSettledEvent, error) {
    event := &indexertypes.PositionSettledEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("PositionSettled: expected at least 2 topics, got %d", len(log.Topics))
    }

    event.TokenID = new(big.Int).SetBytes(log.Topics[1].Bytes())

    if len(log.Data) < 32 {
        return nil, fmt.Errorf("PositionSettled: insufficient data length, expected 32 bytes, got %d", len(log.Data))
    }

    event.ClaimableAmount = new(big.Int).SetBytes(log.Data[0:32])

    return event, nil
}

/*
@method decodeTransfer
@desc Decodes ERC721 Transfer event
*/
func (d *EventDecoder) decodeTransfer(log goEthtypes.Log) (*indexertypes.PositionTransferEvent, error) {
    event := &indexertypes.PositionTransferEvent{}

    if len(log.Topics) < 4 {
        return nil, fmt.Errorf("Transfer: expected 4 topics, got %d", len(log.Topics))
    }
    
    event.From = common.HexToAddress(log.Topics[1].Hex())
    event.To = common.HexToAddress(log.Topics[2].Hex())
    event.TokenID = new(big.Int).SetBytes(log.Topics[3].Bytes())
    
    return event, nil
}

// =====================================================
// LIQUIDATOR EVENT DECODERS
// =====================================================

/*
@method decodeAuctionCreated
@desc Decodes AuctionCreated event from RevvFiLiquidator
*/
func (d *EventDecoder) decodeAuctionCreated(log goEthtypes.Log) (*indexertypes.AuctionCreatedEvent, error) {
    event := &indexertypes.AuctionCreatedEvent{}

    if len(log.Topics) < 4 {
        return nil, fmt.Errorf("AuctionCreated: expected 4 topics, got %d", len(log.Topics))
    }
    
    event.AuctionID = new(big.Int).SetBytes(log.Topics[1].Bytes())
    event.Market = common.HexToAddress(log.Topics[2].Hex())
    event.Borrower = common.HexToAddress(log.Topics[3].Hex())
    
    if len(log.Data) < 192 {
        return nil, fmt.Errorf("AuctionCreated: insufficient data length")
    }
    
    // data layout: [borrowAsset(skip)][collateralAsset(skip)][collateralAmount][debtAmount][startTime][endTime]
    event.CollateralAmount = new(big.Int).SetBytes(log.Data[64:96])
    event.DebtAmount = new(big.Int).SetBytes(log.Data[96:128])
    event.StartTime = new(big.Int).SetBytes(log.Data[128:160])
    event.EndTime = new(big.Int).SetBytes(log.Data[160:192])

    return event, nil
}

/*
@method decodeBidPlaced
@desc Decodes BidPlaced event
*/
func (d *EventDecoder) decodeBidPlaced(log goEthtypes.Log) (*indexertypes.BidPlacedEvent, error) {
    event := &indexertypes.BidPlacedEvent{}

    if len(log.Topics) < 3 {
        return nil, fmt.Errorf("BidPlaced: insufficient topics")
    }
    
    event.AuctionID = new(big.Int).SetBytes(log.Topics[1].Bytes())
    event.Bidder = common.HexToAddress(log.Topics[2].Hex())
    
    if len(log.Data) < 32 {
        return nil, fmt.Errorf("BidPlaced: insufficient data length")
    }
    
    event.BidAmount = new(big.Int).SetBytes(log.Data[0:32])
    
    return event, nil
}

/*
@method decodeAuctionSettled
@desc Decodes AuctionSettled event
*/
func (d *EventDecoder) decodeAuctionSettled(log goEthtypes.Log) (*indexertypes.AuctionSettledEvent, error) {
    event := &indexertypes.AuctionSettledEvent{}

    if len(log.Topics) < 3 {
        return nil, fmt.Errorf("AuctionSettled: insufficient topics")
    }
    
    event.AuctionID = new(big.Int).SetBytes(log.Topics[1].Bytes())
    event.Winner = common.HexToAddress(log.Topics[2].Hex())
    
    // event: AuctionSettled(uint256 indexed auctionId, address indexed winner,
    //        address collateralAsset, uint256 collateralAmount, uint256 bidAmount, uint256 debtRepaid)
    // data layout (128 bytes): [collateralAsset(skip)][collateralAmount][bidAmount][debtRepaid]
    if len(log.Data) < 128 {
        return nil, fmt.Errorf("AuctionSettled: insufficient data length, got %d", len(log.Data))
    }

    // data[0:32] = collateralAsset address — skip
    event.CollateralTransferred = new(big.Int).SetBytes(log.Data[32:64]) // collateralAmount
    event.WinningBid = new(big.Int).SetBytes(log.Data[64:96])            // bidAmount
    event.RecoveryRate = new(big.Int).SetBytes(log.Data[96:128])         // debtRepaid

    return event, nil
}

/*
@method decodeAuctionCancelled
@desc Decodes AuctionCancelled event
*/
func (d *EventDecoder) decodeAuctionCancelled(log goEthtypes.Log) (*indexertypes.AuctionCancelledEvent, error) {
    event := &indexertypes.AuctionCancelledEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("AuctionCancelled: insufficient topics")
    }
    
    event.AuctionID = new(big.Int).SetBytes(log.Topics[1].Bytes())
    
    if len(log.Data) > 0 {
        event.Reason = string(log.Data)
    }
    
    return event, nil
}

// =====================================================
// WITHDRAWAL EVENT DECODERS
// =====================================================

/*
@method decodeWithdrawalRequested
@desc Decodes WithdrawalRequested event from RevvFiLiquidityQueue
*/
func (d *EventDecoder) decodeWithdrawalRequested(log goEthtypes.Log) (*indexertypes.WithdrawalRequestedEvent, error) {
    event := &indexertypes.WithdrawalRequestedEvent{}

    if len(log.Topics) < 3 {
        return nil, fmt.Errorf("WithdrawalRequested: insufficient topics")
    }
    
    event.RequestID = new(big.Int).SetBytes(log.Topics[1].Bytes())
    event.Lender = common.HexToAddress(log.Topics[2].Hex())
    
    if len(log.Topics) > 3 {
        event.PositionID = new(big.Int).SetBytes(log.Topics[3].Bytes())
    }
    
    if len(log.Data) < 96 {
        return nil, fmt.Errorf("WithdrawalRequested: insufficient data length")
    }
    
    event.Amount = new(big.Int).SetBytes(log.Data[0:32])
    event.EpochNumber = new(big.Int).SetBytes(log.Data[64:96])
    
    return event, nil
}

/*
@method decodeWithdrawalClaimed
@desc Decodes WithdrawalClaimed event
*/
func (d *EventDecoder) decodeWithdrawalClaimed(log goEthtypes.Log) (*indexertypes.WithdrawalClaimedEvent, error) {
    event := &indexertypes.WithdrawalClaimedEvent{}

    if len(log.Topics) < 3 {
        return nil, fmt.Errorf("WithdrawalClaimed: insufficient topics")
    }
    
    event.RequestID = new(big.Int).SetBytes(log.Topics[1].Bytes())
    event.Lender = common.HexToAddress(log.Topics[2].Hex())
    
    if len(log.Data) < 32 {
        return nil, fmt.Errorf("WithdrawalClaimed: insufficient data length")
    }
    
    event.Amount = new(big.Int).SetBytes(log.Data[0:32])
    
    return event, nil
}

/*
@method decodeWithdrawalCancelled
@desc Decodes WithdrawalCancelled event
*/
func (d *EventDecoder) decodeWithdrawalCancelled(log goEthtypes.Log) (*indexertypes.WithdrawalCancelledEvent, error) {
    event := &indexertypes.WithdrawalCancelledEvent{}

    if len(log.Topics) < 3 {
        return nil, fmt.Errorf("WithdrawalCancelled: insufficient topics")
    }
    
    event.RequestID = new(big.Int).SetBytes(log.Topics[1].Bytes())
    event.Lender = common.HexToAddress(log.Topics[2].Hex())
    
    return event, nil
}

/*
@method decodeEpochProcessed
@desc Decodes EpochProcessed event
*/
func (d *EventDecoder) decodeEpochProcessed(log goEthtypes.Log) (*indexertypes.EpochProcessedEvent, error) {
    event := &indexertypes.EpochProcessedEvent{}
    
    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("EpochProcessed: insufficient topics")
    }
    
    event.EpochNumber = new(big.Int).SetBytes(log.Topics[1].Bytes())
    
    if len(log.Data) < 64 {
        return nil, fmt.Errorf("EpochProcessed: insufficient data length")
    }
    
    event.TotalRequested = new(big.Int).SetBytes(log.Data[0:32])
    event.TotalFulfilled = new(big.Int).SetBytes(log.Data[32:64])
    
    return event, nil
}

// =====================================================
// REPUTATION EVENT DECODERS
// =====================================================

/*
@method decodeReputationScoreUpdated
@desc Decodes ReputationScoreUpdated event
*/
func (d *EventDecoder) decodeReputationScoreUpdated(log goEthtypes.Log) (*indexertypes.ReputationScoreUpdatedEvent, error) {
    event := &indexertypes.ReputationScoreUpdatedEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("ReputationScoreUpdated: insufficient topics")
    }
    
    event.Borrower = common.HexToAddress(log.Topics[1].Hex())
    
    if len(log.Data) < 64 {
        return nil, fmt.Errorf("ReputationScoreUpdated: insufficient data length")
    }
    
    event.OldScore = new(big.Int).SetBytes(log.Data[0:32])
    event.NewScore = new(big.Int).SetBytes(log.Data[32:64])
    
    return event, nil
}

func (d *EventDecoder) decodeBorrowerRegistered(log goEthtypes.Log) (*indexertypes.BorrowerRegisteredEvent, error) {
    event := &indexertypes.BorrowerRegisteredEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("BorrowerRegistered: expected at least 2 topics, got %d", len(log.Topics))
    }

    event.Borrower = common.HexToAddress(log.Topics[1].Hex())
    return event, nil
}

/*
@method decodeBorrowerRemoved
@desc Decodes BorrowerRemoved event from ReputationRegistry
*/
func (d *EventDecoder) decodeBorrowerRemoved(log goEthtypes.Log) (*indexertypes.ReputationBorrowerRemovedEvent, error) {
    event := &indexertypes.ReputationBorrowerRemovedEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("BorrowerRemoved: insufficient topics")
    }
    
    event.Borrower = common.HexToAddress(log.Topics[1].Hex())
    
    return event, nil
}



// =====================================================
// ARCH CONTROLLER EVENT DECODERS (Continued)
// =====================================================

/*@
 * @method decodeBorrowerAdded
 * @contract RevvFiArchController
 * @event BorrowerAdded(address) - NO INDEXED PARAMETERS
 *
 * @topics_structure (1 topic total):
 *   topics[0]: Event signature hash
 *
 * @data_structure (32 bytes total):
 *   bytes[0:32]: borrower address - 32 bytes with 12-byte left padding, actual address at [12:32]
 *
 * @fix_applied: Borrower address is NOT indexed - read from data, not topics
 */
func (d *EventDecoder) decodeBorrowerAdded(log goEthtypes.Log) (*indexertypes.ArchBorrowerAddedEvent, error) {
    event := &indexertypes.ArchBorrowerAddedEvent{}

    // Validate topic count - BorrowerAdded has only 1 topic (signature)
    if len(log.Topics) < 1 {
        return nil, fmt.Errorf("BorrowerAdded: expected at least 1 topic, got %d", len(log.Topics))
    }

    // Validate data length - need 32 bytes for borrower address
    if len(log.Data) < 32 {
        return nil, fmt.Errorf("BorrowerAdded: insufficient data length, expected 32 bytes, got %d", len(log.Data))
    }

    // Extract borrower address from data (skip 12 bytes of padding)
    event.Borrower = common.HexToAddress("0x" + hex.EncodeToString(log.Data[12:32]))

    return event, nil
}

/*@
 * @method decodeBorrowerRemovedArch
 * @contract RevvFiArchController
 * @event BorrowerRemoved(address) - NO INDEXED PARAMETERS
 *
 * @topics_structure (1 topic total):
 *   topics[0]: Event signature hash
 *
 * @data_structure (32 bytes total):
 *   bytes[0:32]: borrower address - 32 bytes with 12-byte left padding, actual address at [12:32]
 *
 * @fix_applied: Borrower address is NOT indexed - read from data, not topics
 */
func (d *EventDecoder) decodeBorrowerRemovedArch(log goEthtypes.Log) (*indexertypes.ArchBorrowerRemovedEvent, error) {
    event := &indexertypes.ArchBorrowerRemovedEvent{}

    // Validate topic count - BorrowerRemoved has only 1 topic (signature)
    if len(log.Topics) < 1 {
        return nil, fmt.Errorf("BorrowerRemoved: expected at least 1 topic, got %d", len(log.Topics))
    }

    // Validate data length - need 32 bytes for borrower address
    if len(log.Data) < 32 {
        return nil, fmt.Errorf("BorrowerRemoved: insufficient data length, expected 32 bytes, got %d", len(log.Data))
    }

    // Extract borrower address from data (skip 12 bytes of padding)
    event.Borrower = common.HexToAddress("0x" + hex.EncodeToString(log.Data[12:32]))

    return event, nil
}

/*
@method decodeArchControllerOwnerUpdated
@desc Decodes OwnerUpdated event from RevvFiArchController (if exists)
@event_signature OwnerUpdated(address indexed oldOwner, address indexed newOwner)
@topics
  - topics[0]: Event signature hash
  - topics[1]: oldOwner address (indexed)
  - topics[2]: newOwner address (indexed)
*/
func (d *EventDecoder) decodeArchControllerOwnerUpdated(log goEthtypes.Log) (*indexertypes.ArchOwnerUpdatedEvent, error) {
    event := &indexertypes.ArchOwnerUpdatedEvent{}

    if len(log.Topics) < 3 {
        return nil, fmt.Errorf("OwnerUpdated: expected at least 3 topics, got %d", len(log.Topics))
    }
    
    event.OldOwner = common.HexToAddress(log.Topics[1].Hex())
    event.NewOwner = common.HexToAddress(log.Topics[2].Hex())
    
    return event, nil
}

/*
@method decodeArchControllerTimelockUpdated
@desc Decodes TimelockUpdated event from RevvFiArchController (if exists)
@event_signature TimelockUpdated(uint256 oldDelay, uint256 newDelay)
@topics
  - topics[0]: Event signature hash
@data
  - bytes[0:32]: oldDelay (uint256)
  - bytes[32:64]: newDelay (uint256)
*/
func (d *EventDecoder) decodeArchControllerTimelockUpdated(log goEthtypes.Log) (*indexertypes.ArchTimelockUpdatedEvent, error) {
    event := &indexertypes.ArchTimelockUpdatedEvent{}

    if len(log.Data) < 64 {
        return nil, fmt.Errorf("TimelockUpdated: insufficient data length, need 64 bytes, got %d", len(log.Data))
    }
    
    event.OldDelay = new(big.Int).SetBytes(log.Data[0:32])
    event.NewDelay = new(big.Int).SetBytes(log.Data[32:64])
    
    return event, nil
}


/*@
 * decodeReputationUpdated
 * @desc Decodes ReputationScoreUpdated event from ReputationRegistry
 * @event_signature ReputationScoreUpdated(address indexed borrower, uint256 oldScore, uint256 newScore)
 * @topics
 *   - topics[0]: Event signature hash
 *   - topics[1]: borrower address (indexed)
 * @data
 *   - bytes[0:32]: oldScore (uint256)
 *   - bytes[32:64]: newScore (uint256)
 * 
 * @formula Reputation Score:
 *   score = (successfulLoans * 1000 / totalLoans) - (defaultedLoans * 50)
 *   clamped between 0 and 1000
 * 
 * @risk_labels:
 *   - 900-1000: AAA (Excellent)
 *   - 800-899: AA (Very Good)
 *   - 700-799: A (Good)
 *   - 500-699: B (Fair)
 *   - 300-499: C (Poor)
 *   - 0-299: D (Default)
 */
func (d *EventDecoder) decodeReputationUpdated(log goEthtypes.Log) (*indexertypes.ReputationScoreUpdatedEvent, error) {
    event := &indexertypes.ReputationScoreUpdatedEvent{}

    // Validate topic count - ReputationScoreUpdated has 2 topics (including signature)
    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("ReputationScoreUpdated: expected at least 2 topics, got %d", len(log.Topics))
    }
    
    // Extract indexed borrower address from topics[1]
    event.Borrower = common.HexToAddress(log.Topics[1].Hex())
    
    // Validate data length - need 64 bytes for oldScore and newScore
    if len(log.Data) < 64 {
        return nil, fmt.Errorf("ReputationScoreUpdated: insufficient data length, need 64 bytes, got %d", len(log.Data))
    }
    
    // Extract non-indexed parameters from data
    // bytes[0:32] = oldScore (uint256)
    // bytes[32:64] = newScore (uint256)
    event.OldScore = new(big.Int).SetBytes(log.Data[0:32])
    event.NewScore = new(big.Int).SetBytes(log.Data[32:64])
    
    return event, nil
}



// =====================================================
// ADDITIONAL FACTORY EVENT DECODERS
// =====================================================

func (d *EventDecoder) decodeArchControllerUpdateRequested(log goEthtypes.Log) (*indexertypes.ArchControllerUpdateRequestedEvent, error) {
    event := &indexertypes.ArchControllerUpdateRequestedEvent{}
    
    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("ArchControllerUpdateRequested: insufficient topics")
    }
    
    event.NewArchController = common.HexToAddress(log.Topics[1].Hex())
    return event, nil
}

func (d *EventDecoder) decodeArchControllerUpdated(log goEthtypes.Log) (*indexertypes.ArchControllerUpdatedEvent, error) {
    event := &indexertypes.ArchControllerUpdatedEvent{}
    
    if len(log.Topics) < 3 {
        return nil, fmt.Errorf("ArchControllerUpdated: insufficient topics")
    }
    
    event.OldArchController = common.HexToAddress(log.Topics[1].Hex())
    event.NewArchController = common.HexToAddress(log.Topics[2].Hex())
    return event, nil
}

func (d *EventDecoder) decodeFeeUpdated(log goEthtypes.Log) (*indexertypes.FeeUpdatedEvent, error) {
    event := &indexertypes.FeeUpdatedEvent{}
    
    if len(log.Data) < 64 {
        return nil, fmt.Errorf("FeeUpdated: insufficient data length")
    }
    
    event.OldFee = new(big.Int).SetBytes(log.Data[0:32])
    event.NewFee = new(big.Int).SetBytes(log.Data[32:64])
    return event, nil
}

func (d *EventDecoder) decodeImplementationsSet(log goEthtypes.Log) (*indexertypes.ImplementationsSetEvent, error) {
    event := &indexertypes.ImplementationsSetEvent{}
    
    if len(log.Data) < 128 {
        return nil, fmt.Errorf("ImplementationsSet: insufficient data length")
    }
    
    event.MarketImpl = common.HexToAddress("0x" + hex.EncodeToString(log.Data[12:32]))
    event.EscrowImpl = common.HexToAddress("0x" + hex.EncodeToString(log.Data[44:64]))
    event.OfferBookImpl = common.HexToAddress("0x" + hex.EncodeToString(log.Data[76:96]))
    event.LiquidityQueueImpl = common.HexToAddress("0x" + hex.EncodeToString(log.Data[108:128]))
    return event, nil
}

func (d *EventDecoder) decodeOwnershipTransferred(log goEthtypes.Log) (*indexertypes.OwnershipTransferredEvent, error) {
    event := &indexertypes.OwnershipTransferredEvent{}
    
    if len(log.Topics) < 3 {
        return nil, fmt.Errorf("OwnershipTransferred: insufficient topics")
    }
    
    event.PreviousOwner = common.HexToAddress(log.Topics[1].Hex())
    event.NewOwner = common.HexToAddress(log.Topics[2].Hex())
    return event, nil
}

func (d *EventDecoder) decodeContractsSet(log goEthtypes.Log) (*indexertypes.ContractsSetEvent, error) {
    event := &indexertypes.ContractsSetEvent{}
    // it is just the signal event so nothing to do for this functtion
    
    return event, nil
}
// =====================================================
// ADDITIONAL MARKET EVENT DECODERS
// =====================================================

func (d *EventDecoder) decodeCollateralDeposited(log goEthtypes.Log) (*indexertypes.CollateralDepositedEvent, error) {
    event := &indexertypes.CollateralDepositedEvent{}
    
    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("CollateralDeposited: insufficient topics")
    }
    
    event.Borrower = common.HexToAddress(log.Topics[1].Hex())
    
    if len(log.Data) < 32 {
        return nil, fmt.Errorf("CollateralDeposited: insufficient data length")
    }
    
    event.Amount = new(big.Int).SetBytes(log.Data[0:32])
    return event, nil
}

func (d *EventDecoder) decodeCollateralWithdrawn(log goEthtypes.Log) (*indexertypes.CollateralWithdrawnEvent, error) {
    event := &indexertypes.CollateralWithdrawnEvent{}
    
    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("CollateralWithdrawn: insufficient topics")
    }
    
    event.Borrower = common.HexToAddress(log.Topics[1].Hex())
    
    if len(log.Data) < 32 {
        return nil, fmt.Errorf("CollateralWithdrawn: insufficient data length")
    }
    
    event.Amount = new(big.Int).SetBytes(log.Data[0:32])
    return event, nil
}

func (d *EventDecoder) decodeLiquidationExecuted(log goEthtypes.Log) (*indexertypes.LiquidationExecutedEvent, error) {
    event := &indexertypes.LiquidationExecutedEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("LiquidationExecuted: insufficient topics")
    }

    event.Borrower = common.HexToAddress(log.Topics[1].Hex())

    if len(log.Data) < 64 {
        return nil, fmt.Errorf("LiquidationExecuted: insufficient data length")
    }

    event.DebtAmount = new(big.Int).SetBytes(log.Data[0:32])
    event.CollateralAmount = new(big.Int).SetBytes(log.Data[32:64])
    return event, nil
}

/*@
 * @method decodeMarketClosedEvent
 * @contract RevvFiMarket
 * @event MarketClosedEvent(address indexed borrower, uint256 timestamp)
 *
 * @topics_structure (2 topics total):
 *   topics[0]: Event signature hash
 *   topics[1]: borrower address (indexed)
 *
 * @data_structure (32 bytes total):
 *   bytes[0:32]: timestamp (uint256)
 */
func (d *EventDecoder) decodeMarketClosedEvent(log goEthtypes.Log) (*indexertypes.MarketClosedEvent, error) {
    event := &indexertypes.MarketClosedEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("MarketClosedEvent: expected at least 2 topics, got %d", len(log.Topics))
    }

    event.Borrower = common.HexToAddress(log.Topics[1].Hex())
    event.Market = log.Address

    if len(log.Data) < 32 {
        return nil, fmt.Errorf("MarketClosedEvent: insufficient data length, expected 32 bytes, got %d", len(log.Data))
    }

    event.Timestamp = new(big.Int).SetBytes(log.Data[0:32])

    return event, nil
}

// =====================================================
// ADDITIONAL POSITION NFT EVENT DECODERS
// =====================================================

func (d *EventDecoder) decodePositionClaimed(log goEthtypes.Log) (*indexertypes.PositionClaimedEvent, error) {
    event := &indexertypes.PositionClaimedEvent{}

    if len(log.Topics) < 3 {
        return nil, fmt.Errorf("PositionClaimed: insufficient topics")
    }

    event.TokenID = new(big.Int).SetBytes(log.Topics[1].Bytes())
    event.Lender = common.HexToAddress(log.Topics[2].Hex())

    if len(log.Data) < 32 {
        return nil, fmt.Errorf("PositionClaimed: insufficient data length")
    }

    event.Amount = new(big.Int).SetBytes(log.Data[0:32])
    return event, nil
}

/*@
 * @method decodePositionRedeemed
 * @contract RevvFiPositionNFT
 * @event PositionRedeemed(uint256 indexed tokenId, uint256 principal, uint256 interest)
 *
 * @topics_structure (2 topics total):
 *   topics[0]: Event signature hash
 *   topics[1]: tokenId (indexed)
 *
 * @data_structure (64 bytes total):
 *   bytes[0:32]:  principal (uint256)
 *   bytes[32:64]: interest (uint256)
 */
func (d *EventDecoder) decodePositionRedeemed(log goEthtypes.Log) (*indexertypes.PositionRedeemedEvent, error) {
    event := &indexertypes.PositionRedeemedEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("PositionRedeemed: expected at least 2 topics, got %d", len(log.Topics))
    }

    event.TokenID = new(big.Int).SetBytes(log.Topics[1].Bytes())

    if len(log.Data) < 64 {
        return nil, fmt.Errorf("PositionRedeemed: insufficient data length, expected 64 bytes, got %d", len(log.Data))
    }

    event.Principal = new(big.Int).SetBytes(log.Data[0:32])
    event.Interest = new(big.Int).SetBytes(log.Data[32:64])

    return event, nil
}

/*@
 * @method decodeLenderRepaid
 * @contract RevvFiMarket
 * @event LenderRepaid(address indexed lender, uint256 indexed positionId, uint256 newPrincipal, uint256 claimableShare)
 *
 * @topics_structure (3 topics total):
 *   topics[0]: Event signature hash
 *   topics[1]: lender (indexed)
 *   topics[2]: positionId (indexed)
 *
 * @data_structure (64 bytes total):
 *   bytes[0:32]:  newPrincipal (uint256)
 *   bytes[32:64]: claimableShare (uint256)
 */
func (d *EventDecoder) decodeLenderRepaid(log goEthtypes.Log) (*indexertypes.LenderRepaidEvent, error) {
    event := &indexertypes.LenderRepaidEvent{}

    if len(log.Topics) < 3 {
        return nil, fmt.Errorf("LenderRepaid: expected at least 3 topics, got %d", len(log.Topics))
    }

    event.Lender = common.HexToAddress(log.Topics[1].Hex())
    event.PositionID = new(big.Int).SetBytes(log.Topics[2].Bytes())

    if len(log.Data) < 64 {
        return nil, fmt.Errorf("LenderRepaid: insufficient data length, expected 64 bytes, got %d", len(log.Data))
    }

    event.NewPrincipal = new(big.Int).SetBytes(log.Data[0:32])
    event.ClaimableShare = new(big.Int).SetBytes(log.Data[32:64])

    return event, nil
}

// =====================================================
// ADDITIONAL REPUTATION EVENT DECODERS
// =====================================================

func (d *EventDecoder) decodeSuccessfulRepaymentRecorded(log goEthtypes.Log) (*indexertypes.SuccessfulRepaymentRecordedEvent, error) {
    event := &indexertypes.SuccessfulRepaymentRecordedEvent{}
    
    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("SuccessfulRepaymentRecorded: insufficient topics")
    }
    
    event.Borrower = common.HexToAddress(log.Topics[1].Hex())
    
    if len(log.Data) < 32 {
        return nil, fmt.Errorf("SuccessfulRepaymentRecorded: insufficient data length")
    }
    
    event.Amount = new(big.Int).SetBytes(log.Data[0:32])
    return event, nil
}

func (d *EventDecoder) decodeDefaultRecorded(log goEthtypes.Log) (*indexertypes.DefaultRecordedEvent, error) {
    event := &indexertypes.DefaultRecordedEvent{}
    
    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("DefaultRecorded: insufficient topics")
    }
    
    event.Borrower = common.HexToAddress(log.Topics[1].Hex())
    
    if len(log.Data) < 32 {
        return nil, fmt.Errorf("DefaultRecorded: insufficient data length")
    }
    
    event.Amount = new(big.Int).SetBytes(log.Data[0:32])
    return event, nil
}

// =====================================================
// ADDITIONAL ARCH CONTROLLER EVENT DECODERS
// =====================================================
/*@
 * @method decodeControllerAdded
 * @contract RevvFiArchController
 * @event ControllerAdded(address indexed controllerFactory, address controller)
 *
 * @topics_structure (3 topics total):
 *   topics[0]: Event signature hash
 *   topics[1]: controllerFactory (indexed)
 *   topics[2]: controller (indexed)
 */
func (d *EventDecoder) decodeControllerAdded(log goEthtypes.Log) (*indexertypes.ControllerAddedEvent, error) {
    event := &indexertypes.ControllerAddedEvent{}
    
    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("ControllerAdded: expected at least 3 topics, got %d", len(log.Topics))
    }
    
    event.ControllerFactory = common.HexToAddress(log.Topics[1].Hex())
    
    return event, nil
}

/*@
 * @method decodeControllerRemoved
 * @contract RevvFiArchController
 * @event ControllerRemoved(address) - NO INDEXED PARAMETERS
 */
func (d *EventDecoder) decodeControllerRemoved(log goEthtypes.Log) (*indexertypes.ControllerRemovedEvent, error) {
    event := &indexertypes.ControllerRemovedEvent{}

    if len(log.Topics) < 1 {
        return nil, fmt.Errorf("ControllerRemoved: expected at least 1 topic, got %d", len(log.Topics))
    }

    if len(log.Data) < 32 {
        return nil, fmt.Errorf("ControllerRemoved: insufficient data length, expected 32 bytes, got %d", len(log.Data))
    }

    event.Controller = common.HexToAddress("0x" + hex.EncodeToString(log.Data[12:32]))
    return event, nil
}

func (d *EventDecoder) decodeMarketRegistered(log goEthtypes.Log) (*indexertypes.MarketRegisteredEvent, error) {
    event := &indexertypes.MarketRegisteredEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("MarketRegistered: expected at least 2 topics, got %d", len(log.Topics))
    }

    event.Market = common.HexToAddress(log.Topics[1].Hex())
    return event, nil
}

func (d *EventDecoder) decodeMarketAdded(log goEthtypes.Log) (*indexertypes.MarketAddedEvent, error) {
    event := &indexertypes.MarketAddedEvent{}
    
    // MarketAdded has 2 topics (signature + market address)
    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("MarketAdded: expected at least 2 topics, got %d", len(log.Topics))
    }
    
    event.Market = common.HexToAddress(log.Topics[1].Hex())
    
    if len(log.Topics) > 2 {
        event.Controller = common.HexToAddress(log.Topics[2].Hex())
    }
    
    return event, nil
}
/*@
 * @method decodeControllerFactoryAdded
 * @contract RevvFiArchController
 * @event ControllerFactoryAdded(address) - NO INDEXED PARAMETERS
 */
func (d *EventDecoder) decodeControllerFactoryAdded(log goEthtypes.Log) (*indexertypes.ControllerFactoryAddedEvent, error) {
    event := &indexertypes.ControllerFactoryAddedEvent{}

    if len(log.Topics) < 1 {
        return nil, fmt.Errorf("ControllerFactoryAdded: expected at least 1 topic, got %d", len(log.Topics))
    }

    if len(log.Data) < 32 {
        return nil, fmt.Errorf("ControllerFactoryAdded: insufficient data length, expected 32 bytes, got %d", len(log.Data))
    }

    event.Factory = common.HexToAddress("0x" + hex.EncodeToString(log.Data[12:32]))
    return event, nil
}
func (d *EventDecoder) decodeInitialized(log goEthtypes.Log) (*indexertypes.InitializedEvent, error) {
    event := &indexertypes.InitializedEvent{}
    
    // Initialized event has 1 topic (signature) and version in data
    if len(log.Data) < 32 {
        return nil, fmt.Errorf("Initialized: insufficient data length")
    }
    
    event.Version = new(big.Int).SetBytes(log.Data[0:32]).Uint64()
    
    return event, nil
}

/*@
 * @method decodeMarketUnregistered
 * @contract RevvFiArchController / RevvFiPositionNFT
 * @event MarketUnregistered(address) - NO INDEXED PARAMETERS
 */
func (d *EventDecoder) decodeMarketUnregistered(log goEthtypes.Log) (*indexertypes.MarketUnregisteredEvent, error) {
    event := &indexertypes.MarketUnregisteredEvent{}

    if len(log.Topics) < 1 {
        return nil, fmt.Errorf("MarketUnregistered: expected at least 1 topic, got %d", len(log.Topics))
    }

    if len(log.Data) < 32 {
        return nil, fmt.Errorf("MarketUnregistered: insufficient data length, expected 32 bytes, got %d", len(log.Data))
    }

    event.Market = common.HexToAddress("0x" + hex.EncodeToString(log.Data[12:32]))
    return event, nil
}

/*@
 * @method decodeControllerFactoryRemoved
 * @contract RevvFiArchController
 * @event ControllerFactoryRemoved(address) - NO INDEXED PARAMETERS
 */
func (d *EventDecoder) decodeControllerFactoryRemoved(log goEthtypes.Log) (*indexertypes.ControllerFactoryRemovedEvent, error) {
    event := &indexertypes.ControllerFactoryRemovedEvent{}

    if len(log.Topics) < 1 {
        return nil, fmt.Errorf("ControllerFactoryRemoved: expected at least 1 topic, got %d", len(log.Topics))
    }

    if len(log.Data) < 32 {
        return nil, fmt.Errorf("ControllerFactoryRemoved: insufficient data length, expected 32 bytes, got %d", len(log.Data))
    }

    event.Factory = common.HexToAddress("0x" + hex.EncodeToString(log.Data[12:32]))
    return event, nil
}

/*@
 * @method decodeMarketRemoved
 * @contract RevvFiArchController
 * @event MarketRemoved(address) - NO INDEXED PARAMETERS
 */
func (d *EventDecoder) decodeMarketRemoved(log goEthtypes.Log) (*indexertypes.MarketRemovedEvent, error) {
    event := &indexertypes.MarketRemovedEvent{}

    if len(log.Topics) < 1 {
        return nil, fmt.Errorf("MarketRemoved: expected at least 1 topic, got %d", len(log.Topics))
    }

    if len(log.Data) < 32 {
        return nil, fmt.Errorf("MarketRemoved: insufficient data length, expected 32 bytes, got %d", len(log.Data))
    }

    event.Market = common.HexToAddress("0x" + hex.EncodeToString(log.Data[12:32]))
    return event, nil
}

/*@
 * @method decodeAssetBlacklisted
 * @contract RevvFiArchController
 * @event AssetBlacklisted(address) - NO INDEXED PARAMETERS
 */
func (d *EventDecoder) decodeAssetBlacklisted(log goEthtypes.Log) (*indexertypes.AssetBlacklistedEvent, error) {
    event := &indexertypes.AssetBlacklistedEvent{}

    if len(log.Topics) < 1 {
        return nil, fmt.Errorf("AssetBlacklisted: expected at least 1 topic, got %d", len(log.Topics))
    }

    if len(log.Data) < 32 {
        return nil, fmt.Errorf("AssetBlacklisted: insufficient data length, expected 32 bytes, got %d", len(log.Data))
    }

    event.Asset = common.HexToAddress("0x" + hex.EncodeToString(log.Data[12:32]))
    return event, nil
}

/*@
 * @method decodeAssetPermitted
 * @contract RevvFiArchController
 * @event AssetPermitted(address) - NO INDEXED PARAMETERS
 */
func (d *EventDecoder) decodeAssetPermitted(log goEthtypes.Log) (*indexertypes.AssetPermittedEvent, error) {
    event := &indexertypes.AssetPermittedEvent{}

    if len(log.Topics) < 1 {
        return nil, fmt.Errorf("AssetPermitted: expected at least 1 topic, got %d", len(log.Topics))
    }

    if len(log.Data) < 32 {
        return nil, fmt.Errorf("AssetPermitted: insufficient data length, expected 32 bytes, got %d", len(log.Data))
    }

    event.Asset = common.HexToAddress("0x" + hex.EncodeToString(log.Data[12:32]))
    return event, nil
}


/*@
 * @method decodeDrawdownExecutedOffer
 * @contract RevvFiOfferBook
 * @event DrawdownExecutedOffer(address indexed borrower, uint256 totalAmount, uint256 weightedApr)
 *
 * @topics_structure (2 topics total):
 *   topics[0]: Event signature hash
 *   topics[1]: borrower (indexed)
 *
 * @data_structure (64 bytes total):
 *   bytes[0:32]: totalAmount (uint256)
 *   bytes[32:64]: weightedApr (uint256)
 */
func (d *EventDecoder) decodeDrawdownExecutedOffer(log goEthtypes.Log) (*indexertypes.DrawdownExecutedOfferEvent, error) {
    event := &indexertypes.DrawdownExecutedOfferEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("DrawdownExecutedOffer: expected at least 2 topics, got %d", len(log.Topics))
    }

    event.Borrower = common.HexToAddress(log.Topics[1].Hex())

    if len(log.Data) < 64 {
        return nil, fmt.Errorf("DrawdownExecutedOffer: insufficient data length, need 64 bytes, got %d", len(log.Data))
    }

    event.TotalAmount = new(big.Int).SetBytes(log.Data[0:32])
    event.WeightedAPR = new(big.Int).SetBytes(log.Data[32:64])

    return event, nil
}

/*@
 * @method decodeDrawdownExecuted
 * @contract RevvFiMarket
 * @event DrawdownExecuted(uint256 totalAmount, uint256 weightedApr, uint256[] positionIds)
 *
 * @topics_structure (1 topic total):
 *   topics[0]: Event signature hash
 *
 * @data_structure: variable length
 */
func (d *EventDecoder) decodeDrawdownExecuted(log goEthtypes.Log) (*indexertypes.DrawdownExecutedEvent, error) {
    event := &indexertypes.DrawdownExecutedEvent{}

    if len(log.Data) < 64 {
        return nil, fmt.Errorf("DrawdownExecuted: insufficient data length, need at least 64 bytes, got %d", len(log.Data))
    }

    event.TotalAmount = new(big.Int).SetBytes(log.Data[0:32])
    event.WeightedAPR = new(big.Int).SetBytes(log.Data[32:64])

    return event, nil
}
/*@
 * @method decodeBorrowActivityRecorded
 * @contract ReputationRegistry
 * @event BorrowActivityRecorded(address indexed borrower, uint256 amount)
 *
 * @topics_structure (2 topics total):
 *   topics[0]: Event signature hash
 *   topics[1]: borrower (indexed)
 *
 * @data_structure (32 bytes total):
 *   bytes[0:32]: amount (uint256)
 */
func (d *EventDecoder) decodeBorrowActivityRecorded(log goEthtypes.Log) (*indexertypes.BorrowActivityRecordedEvent, error) {
    event := &indexertypes.BorrowActivityRecordedEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("BorrowActivityRecorded: expected at least 2 topics, got %d", len(log.Topics))
    }

    event.Borrower = common.HexToAddress(log.Topics[1].Hex())

    if len(log.Data) < 32 {
        return nil, fmt.Errorf("BorrowActivityRecorded: insufficient data length, need 32 bytes, got %d", len(log.Data))
    }

    event.Amount = new(big.Int).SetBytes(log.Data[0:32])

    return event, nil
}

// =====================================================
// COLLATERAL ESCROW ADDITIONAL DECODERS
// =====================================================

// decodeCollateralLiquidated decodes CollateralLiquidated(address indexed borrower, uint256 collateralAmount, uint256 debtAmount)
func (d *EventDecoder) decodeCollateralLiquidated(log goEthtypes.Log) (*indexertypes.CollateralLiquidatedEvent, error) {
    event := &indexertypes.CollateralLiquidatedEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("CollateralLiquidated: expected at least 2 topics, got %d", len(log.Topics))
    }

    event.Borrower = common.HexToAddress(log.Topics[1].Hex())

    if len(log.Data) < 64 {
        return nil, fmt.Errorf("CollateralLiquidated: insufficient data length, need 64 bytes, got %d", len(log.Data))
    }

    event.Amount = new(big.Int).SetBytes(log.Data[0:32])   // collateralAmount
    event.Proceeds = new(big.Int).SetBytes(log.Data[32:64]) // debtAmount

    return event, nil
}

// =====================================================
// MARKET LIQUIDATION STATE DECODERS
// =====================================================

// decodeLiquidationStartedMarket decodes LiquidationStartedMarket(address indexed borrower)
func (d *EventDecoder) decodeLiquidationStartedMarket(log goEthtypes.Log) (*indexertypes.LiquidationStartedMarketEvent, error) {
    event := &indexertypes.LiquidationStartedMarketEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("LiquidationStartedMarket: expected at least 2 topics, got %d", len(log.Topics))
    }

    event.Borrower = common.HexToAddress(log.Topics[1].Hex())
    event.Market = log.Address

    return event, nil
}

// decodeLiquidationEndedMarket decodes LiquidationEndedMarket(address indexed borrower)
func (d *EventDecoder) decodeLiquidationEndedMarket(log goEthtypes.Log) (*indexertypes.LiquidationEndedMarketEvent, error) {
    event := &indexertypes.LiquidationEndedMarketEvent{}

    if len(log.Topics) < 2 {
        return nil, fmt.Errorf("LiquidationEndedMarket: expected at least 2 topics, got %d", len(log.Topics))
    }

    event.Borrower = common.HexToAddress(log.Topics[1].Hex())
    event.Market = log.Address

    return event, nil
}

// decodeMinCollateralRatioUpdated decodes MinCollateralRatioUpdated(uint256 oldRatio, uint256 newRatio)
func (d *EventDecoder) decodeMinCollateralRatioUpdated(log goEthtypes.Log) (*indexertypes.MinCollateralRatioUpdatedEvent, error) {
    event := &indexertypes.MinCollateralRatioUpdatedEvent{}

    if len(log.Data) < 64 {
        return nil, fmt.Errorf("MinCollateralRatioUpdated: insufficient data, got %d bytes", len(log.Data))
    }

    event.OldRatio = new(big.Int).SetBytes(log.Data[0:32])
    event.NewRatio = new(big.Int).SetBytes(log.Data[32:64])

    return event, nil
}

// decodeLiquidationThresholdUpdated decodes LiquidationThresholdUpdated(uint256 oldThreshold, uint256 newThreshold)
func (d *EventDecoder) decodeLiquidationThresholdUpdated(log goEthtypes.Log) (*indexertypes.LiquidationThresholdUpdatedEvent, error) {
    event := &indexertypes.LiquidationThresholdUpdatedEvent{}

    if len(log.Data) < 64 {
        return nil, fmt.Errorf("LiquidationThresholdUpdated: insufficient data, got %d bytes", len(log.Data))
    }

    event.OldThreshold = new(big.Int).SetBytes(log.Data[0:32])
    event.NewThreshold = new(big.Int).SetBytes(log.Data[32:64])

    return event, nil
}