// internal/repository/postgres/event.go
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@struct EventRepository
@desc PostgreSQL implementation for event tracking and idempotency
@implements repository.EventRepository interface
*/
type EventRepository struct {
    db *DB
}

/*
@function NewEventRepository
@desc Creates a new event repository instance
*/
func NewEventRepository(db *DB) *EventRepository {
    return &EventRepository{db: db}
}

// =====================================================
// IDEMPOTENCY METHODS
// =====================================================

/*
@method IsProcessed
@desc Checks if an event has already been processed
@param txHash Transaction hash of the event
@param logIndex Log index within the transaction
@return true if event already processed, false otherwise
@security Prevents duplicate event processing
*/
func (r *EventRepository) IsProcessed(ctx context.Context, txHash string, logIndex int) (bool, error) {
    var exists bool
    err := r.db.conn.QueryRowContext(ctx, `
        SELECT EXISTS(
            SELECT 1 FROM processed_events 
            WHERE tx_hash = $1 AND log_index = $2
        )
    `, txHash, logIndex).Scan(&exists)
    
    if err != nil {
        return false, fmt.Errorf("failed to check processed event: %w", err)
    }
    return exists, nil
}

/*
@method MarkProcessed
@desc Marks an event as processed in the database
@param event ProcessedEvent to record
@security Uses ON CONFLICT to handle race conditions
*/
func (r *EventRepository) MarkProcessed(ctx context.Context, event *models.ProcessedEvent) error {
    _, err := r.db.conn.ExecContext(ctx, `
        INSERT INTO processed_events (
            tx_hash, log_index, block_number, event_name, 
            contract_address, processed_at
        ) VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (tx_hash, log_index) DO NOTHING
    `, event.TxHash, event.LogIndex, event.BlockNumber, 
       event.EventName, event.ContractAddress, time.Now())
    
    return mapError(err)
}
// =====================================================
// FAILED EVENTS METHODS (NEW)
// =====================================================

/*
@method SaveFailedEvent
@desc Stores a failed event with all details for later retry
@param eventName Name of the event that failed
@param txHash Transaction hash of the failed event
@param blockNumber Block number where event occurred
@param reason Error message explaining the failure
*/
func (r *EventRepository) SaveFailedEvent(ctx context.Context, eventName, txHash string, blockNumber uint64, reason string) error {
    payload := json.RawMessage(`{"error": "` + reason + `"}`)
    _, err := r.db.conn.ExecContext(ctx, `
        INSERT INTO failed_events (
            event_name, tx_hash, block_number, error_message, reason, payload, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, NOW())
    `, eventName, txHash, blockNumber, reason, reason, payload)
    
    return mapError(err)
}

/*
@method SaveFailedEventWithData
@desc Stores a failed event with raw event data for later analysis
@param eventName Name of the event that failed
@param txHash Transaction hash
@param blockNumber Block number
@param eventData Raw event data as JSON
@param reason Error message
*/
func (r *EventRepository) SaveFailedEventWithData(ctx context.Context, eventName, txHash string, blockNumber uint64, eventData interface{}, reason string) error {
    eventDataJSON, err := json.Marshal(eventData)
    if err != nil {
        return err
    }
    
    _, err = r.db.conn.ExecContext(ctx, `
        INSERT INTO failed_events (
            event_name, tx_hash, block_number, error_message, event_data, reason, payload, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $6, NOW())
    `, eventName, txHash, blockNumber, reason, eventDataJSON, reason)
    
    return mapError(err)
}

/*
@method GetFailedEvents
@desc Retrieves failed events that need retry
@param limit Maximum number of events to retrieve
@return Slice of failed event records
*/
func (r *EventRepository) GetFailedEvents(ctx context.Context, limit int) ([]*FailedEventRecord, error) {
    rows, err := r.db.conn.QueryContext(ctx, `
        SELECT id, event_name, tx_hash, block_number, error_message, created_at
        FROM failed_events
        WHERE retry_count < 5
        ORDER BY created_at ASC
        LIMIT $1
    `, limit)
    
    if err != nil {
        return nil, mapError(err)
    }
    defer rows.Close()
    
    var events []*FailedEventRecord
    for rows.Next() {
        var event FailedEventRecord
        err := rows.Scan(&event.ID, &event.EventName, &event.TxHash, &event.BlockNumber, &event.FailureReason, &event.CreatedAt)
        if err != nil {
            return nil, mapError(err)
        }
        events = append(events, &event)
    }
    
    return events, rows.Err()
}

//===================================
// OFFER METHODS
// =====================================================

/*
@method SaveOffer
@desc Inserts a new offer from OfferSubmitted event
@param offer Offer model to insert
*/
func (r *EventRepository) SaveOffer(ctx context.Context, offer *models.Offer) error {
    _, err := r.db.conn.ExecContext(ctx, `
        INSERT INTO offers (
            offer_id, lender, market_address, amount, remaining_amount,
            apr, seniority, status, expiry, block_number, tx_hash, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
        ON CONFLICT (offer_id) DO UPDATE SET
            remaining_amount = EXCLUDED.remaining_amount,
            status = EXCLUDED.status,
            updated_at = EXCLUDED.updated_at
    `, offer.OfferID, offer.Lender, offer.MarketAddress, offer.Amount.String(), 
       offer.RemainingAmount.String(), offer.APR, offer.Seniority, offer.Status,
       offer.Expiry, offer.BlockNumber, offer.TxHash, offer.CreatedAt, offer.UpdatedAt)
    
    return mapError(err)
}

/*
@method UpdateOfferStatus
@desc Updates an offer's status (cancelled, filled, expired)
@param offerID Offer ID to update
@param status New status value
*/
func (r *EventRepository) UpdateOfferStatus(ctx context.Context, offerID int64, status string) error {
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE offers 
        SET status = $2, updated_at = NOW()
        WHERE offer_id = $1
    `, offerID, status)
    
    return mapError(err)
}

/*
@method UpdateOfferRemaining
@desc Updates the remaining amount of an offer (when partially filled)
@param offerID Offer ID to update
@param remainingAmount New remaining amount
*/
func (r *EventRepository) UpdateOfferRemaining(ctx context.Context, offerID int64, remainingAmount *big.Int) error {
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE offers 
        SET remaining_amount = $2, 
            status = CASE 
                WHEN $2 = '0' THEN 'filled' 
                WHEN amount != $2 THEN 'partially_filled' 
                ELSE status 
            END,
            updated_at = NOW()
        WHERE offer_id = $1
    `, offerID, remainingAmount.String())
    
    return mapError(err)
}

// =====================================================
// MARKET METHODS
// =====================================================

/*
@method SaveMarket
@desc Inserts a new market from MarketDeployed event
@param market Market model to insert
*/
func (r *EventRepository) SaveMarket(ctx context.Context, market *models.Market) error {
    _, err := r.db.conn.ExecContext(ctx, `
        INSERT INTO markets (
            address, borrower, borrow_asset, borrow_asset_decimals,
            collateral_asset, collateral_asset_decimals, collateral_oracle,
            min_collateral_ratio, liquidation_threshold, is_active, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (address) DO NOTHING
    `, market.Address, market.Borrower, market.BorrowAsset,
       market.BorrowAssetDecimals, market.CollateralAsset, 
       market.CollateralAssetDecimals, market.CollateralOracle,
       market.MinCollateralRatio, market.LiquidationThreshold,
       market.IsActive, market.CreatedAt)
    
    return mapError(err)
}

/*@
 * UpdateMarketStatus
 * @desc Updates market operational status (active/paused/liquidating)
 * @param ctx Context for cancellation
 * @param marketAddress Market contract address
 * @param isActive Whether market is active
 * @param isLiquidating Whether market is in liquidation
 * @returns error if update fails
 *
 * @sql UPDATE markets SET is_active = $2, is_liquidating = $3, last_health_check = NOW()
 */
func (r *EventRepository) UpdateMarketStatus(ctx context.Context, marketAddress string, isActive, isLiquidating bool) error {
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE markets 
        SET is_active = $2, 
            is_liquidating = $3,
            last_health_check = NOW()
        WHERE address = $1
    `, marketAddress, isActive, isLiquidating)

    return mapError(err)
}

/*@
 * UpdateMarketBorrowIndex
 * @desc Updates market borrow index for interest accrual tracking
 * @param ctx Context for cancellation
 * @param marketAddress Market contract address
 * @param borrowIndex New borrow index (1e18 scale)
 * @returns error if update fails
 *
 * @formula borrowIndex = borrowIndex * (1 + interestRate * timeElapsed)
 * @scale 1e18 (10^18)
 */
/*
@method UpdateMarketBorrowIndex
@desc Updates market borrow index from InterestAccrued event
@param marketAddress Market contract address
@param borrowIndex New borrow index value (optional, can be nil if not provided)
@note If borrowIndex is nil, just updates last_interest_accrual timestamp
*/
func (r *EventRepository) UpdateMarketBorrowIndex(ctx context.Context, marketAddress string, borrowIndex ...*big.Int) error {
    if len(borrowIndex) > 0 && borrowIndex[0] != nil {
        _, err := r.db.conn.ExecContext(ctx, `
            UPDATE markets
            SET borrow_index = $2,
                last_interest_accrual = NOW()
            WHERE address = $1
        `, marketAddress, borrowIndex[0].String())
        return mapError(err)
    }

    // If no borrow index provided, just update timestamp
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE markets
        SET last_interest_accrual = NOW()
        WHERE address = $1
    `, marketAddress)

    return mapError(err)
}

// =====================================================
// POSITION METHODS
// =====================================================

/*
@method SavePosition
@desc Inserts a new position from PositionMinted event
@param position Position model to insert
*/
func (r *EventRepository) SavePosition(ctx context.Context, position *models.Position) error {
    _, err := r.db.conn.ExecContext(ctx, `
        INSERT INTO positions (
            token_id, lender, market_address, principal, current_principal,
            apr, seniority, status, is_active, minted_at, block_number, tx_hash, log_index
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
        ON CONFLICT (token_id) DO UPDATE SET
            status = EXCLUDED.status,
            is_active = EXCLUDED.is_active,
            claimable_amount = EXCLUDED.claimable_amount
    `, position.TokenID, position.Lender, position.MarketAddress,
       position.Principal.String(), position.CurrentPrincipal.String(),
       position.APR, position.Seniority, position.Status, position.IsActive,
       position.MintedAt, position.BlockNumber, position.TxHash, position.LogIndex)
    
    return mapError(err)
}

/*
@method SettlePosition
@desc Marks a position as settled from PositionSettled event
@param tokenID Position token ID to settle
@param claimableAmount Amount claimable by lender
*/
func (r *EventRepository) SettlePosition(ctx context.Context, tokenID int64, claimableAmount *big.Int) error {
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE positions 
        SET is_active = false, 
            is_settled = true, 
            status = 'settled',
            claimable_amount = $2,
            settled_at = NOW(),
            last_updated = NOW()
        WHERE token_id = $1
    `, tokenID, claimableAmount.String())
    
    return mapError(err)
}
/*@
 * UpdatePositionOwner
 * @desc Updates position owner when NFT is transferred
 * @param ctx Context for cancellation
 * @param tokenID Position token ID
 * @param newOwner New owner address
 * @returns error if update fails
 *
 * @sql UPDATE positions SET lender = $2, last_updated = NOW() WHERE token_id = $1
 */
func (r *EventRepository) UpdatePositionOwner(ctx context.Context, tokenID int64, newOwner string) error {
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE positions 
        SET lender = $2,
            last_updated = NOW()
        WHERE token_id = $1
    `, tokenID, newOwner)

    return mapError(err)
}

// =====================================================
// BORROW METHODS
// =====================================================

/*
@method UpdateMarketDebt
@desc Updates market debt after Borrow event
@param marketAddress Market address
@param newDebt New total debt amount
*/
func (r *EventRepository) UpdateMarketDebt(ctx context.Context, marketAddress string, newDebt *big.Int) error {
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE markets 
        SET total_debt = $2, 
            total_principal = $2,
            last_health_check = NOW()
        WHERE address = $1
    `, marketAddress, newDebt.String())
    
    return mapError(err)
}

// =====================================================
// REPAYMENT METHODS
// =====================================================

/*
@method SaveRepayment
@desc Records a repayment transaction
@param repayment Repayment model to insert
*/
func (r *EventRepository) SaveRepayment(ctx context.Context, repayment *models.Repayment) error {
    _, err := r.db.conn.ExecContext(ctx, `
        INSERT INTO repayments (
            market_address, borrower_address, amount, interest_paid, 
            principal_paid, repayment_type, block_number, tx_hash, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `, repayment.MarketAddress, repayment.BorrowerAddress, 
       repayment.Amount.String(), repayment.InterestPaid.String(),
       repayment.PrincipalPaid.String(), repayment.RepaymentType,
       repayment.BlockNumber, repayment.TxHash, repayment.CreatedAt)
    
    return mapError(err)
}

// =====================================================
// AUCTION METHODS
// =====================================================

/*
@method SaveAuction
@desc Inserts a new auction from AuctionCreated event
@param auction Auction model to insert
*/
func (r *EventRepository) SaveAuction(ctx context.Context, auction *models.Auction) error {
    _, err := r.db.conn.ExecContext(ctx, `
        INSERT INTO auctions (
            auction_id, market_address, borrower_address, collateral_amount,
            debt_amount, current_price, start_time, end_time, status, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        ON CONFLICT (auction_id) DO NOTHING
    `, auction.AuctionID, auction.MarketAddress, auction.BorrowerAddress,
       auction.CollateralAmount.String(), auction.DebtAmount.String(),
       auction.CurrentPrice.String(), auction.StartTime, auction.EndTime,
       auction.Status, auction.CreatedAt)
    
    return mapError(err)
}

/*
@method UpdateAuctionBid
@desc Updates auction with highest bid
@param auctionID Auction ID
@param highestBid Highest bid amount
@param highestBidder Address of highest bidder
*/
func (r *EventRepository) UpdateAuctionBid(ctx context.Context, auctionID int64, highestBid *big.Int, highestBidder string) error {
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE auctions 
        SET highest_bid = $2, 
            highest_bidder = $3, 
            updated_at = NOW()
        WHERE auction_id = $1
    `, auctionID, highestBid.String(), highestBidder)
    
    return mapError(err)
}

/*
@method SettleAuction
@desc Marks auction as settled with winner info
@param auctionID Auction ID
@param winner Winner address
@param winningBid Winning bid amount
@param recoveryRate Recovery rate percentage
*/
func (r *EventRepository) SettleAuction(ctx context.Context, auctionID int64, winner string, winningBid *big.Int, recoveryRate float64) error {
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE auctions 
        SET status = 'settled',
            winner = $2,
            winning_bid = $3,
            recovery_rate = $4,
            settlement_time = NOW(),
            updated_at = NOW()
        WHERE auction_id = $1
    `, auctionID, winner, winningBid.String(), recoveryRate)
    
    return mapError(err)
}

/*@
 * UpdateAuctionStatus
 * @desc Updates auction status (cancelled, expired, failed)
 * @param ctx Context for cancellation
 * @param auctionID Auction ID to update
 * @param status New status value
 * @returns error if update fails
 *
 * @valid_status_values
 *   - "active": Auction is ongoing
 *   - "settled": Auction completed with successful bid
 *   - "cancelled": Auction cancelled (no bids, retry scheduled)
 *   - "failed": Auction permanently failed after max retries
 *   - "expired": Auction ended without any bids
 *
 * @sql UPDATE auctions SET status = $2, updated_at = NOW() WHERE auction_id = $1
 *
 * @usage
 *   - Called from handleAuctionCancelled when auction has no bids
 *   - Called when auction expires without reaching reserve price
 *   - Called when max retry attempts exceeded
 */
func (r *EventRepository) UpdateAuctionStatus(ctx context.Context, auctionID int64, status string) error {
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE auctions 
        SET status = $2, 
            updated_at = NOW()
        WHERE auction_id = $1
    `, auctionID, status)

    return mapError(err)
}
// =====================================================
// WITHDRAWAL METHODS
// =====================================================

/*
@method SaveWithdrawalRequest
@desc Inserts a withdrawal request from WithdrawalRequested event
@param request WithdrawalRequest model to insert
*/
func (r *EventRepository) SaveWithdrawalRequest(ctx context.Context, request *models.WithdrawalRequest) error {
    _, err := r.db.conn.ExecContext(ctx, `
        INSERT INTO withdrawal_requests (
            request_id, lender, position_id, market_address, requested_amount,
            epoch_number, status, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (request_id) DO NOTHING
    `, request.RequestID, request.Lender, request.PositionID, 
       request.MarketAddress, request.RequestedAmount.String(),
       request.EpochNumber, request.Status, request.CreatedAt)
    
    return mapError(err)
}

/*
@method UpdateWithdrawalFulfillment
@desc Updates withdrawal request with fulfilled amount
@param requestID Request ID
@param fulfilledAmount Amount fulfilled
*/
func (r *EventRepository) UpdateWithdrawalFulfillment(ctx context.Context, requestID int64, fulfilledAmount *big.Int) error {
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE withdrawal_requests 
        SET fulfilled_amount = $2,
            remaining_amount = requested_amount::numeric - $2,
            status = CASE 
                WHEN requested_amount::numeric <= $2 THEN 'ready'
                ELSE 'pending'
            END,
            fulfillment_time = CASE 
                WHEN requested_amount::numeric <= $2 THEN NOW()
                ELSE NULL
            END,
            updated_at = NOW()
        WHERE request_id = $1
    `, requestID, fulfilledAmount.String())
    
    return mapError(err)
}

/*
@method MarkWithdrawalClaimed
@desc Marks withdrawal request as claimed
@param requestID Request ID
*/
func (r *EventRepository) MarkWithdrawalClaimed(ctx context.Context, requestID int64) error {
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE withdrawal_requests 
        SET status = 'claimed', updated_at = NOW()
        WHERE request_id = $1
    `, requestID)
    
    return mapError(err)
}
/*@
 * UpdateWithdrawalStatus
 * @desc Updates withdrawal request status
 * @param ctx Context for cancellation
 * @param requestID Withdrawal request ID
 * @param status New status (cancelled, ready, claimed, expired)
 * @returns error if update fails
 *
 * @valid_status_values
 *   - "pending": Awaiting epoch processing
 *   - "ready": Ready to claim
 *   - "claimed": Already claimed by user
 *   - "cancelled": Cancelled by user
 *   - "expired": Expired before claiming
 */
func (r *EventRepository) UpdateWithdrawalStatus(ctx context.Context, requestID int64, status string) error {
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE withdrawal_requests 
        SET status = $2,
            updated_at = NOW(),
            fulfillment_time = CASE 
                WHEN $2 = 'claimed' AND fulfillment_time IS NULL THEN NOW()
                ELSE fulfillment_time
            END
        WHERE request_id = $1
    `, requestID, status)

    return mapError(err)
}

/*@
 * UpdateEpochFulfillment
 * @desc Updates withdrawal epoch with total requested and fulfilled amounts
 * @param ctx Context for cancellation
 * @param epochNumber Epoch number
 * @param totalRequested Total amount requested in epoch
 * @param totalFulfilled Total amount fulfilled in epoch
 * @returns error if update fails
 *
 * @sql UPDATE withdrawal_epochs SET total_requested = $2, total_fulfilled = $3, processed_at = NOW()
 * @note Called when epoch processing completes
 */
func (r *EventRepository) UpdateEpochFulfillment(ctx context.Context, epochNumber int32, totalRequested, totalFulfilled *big.Int) error {
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE withdrawal_epochs 
        SET total_requested = $2,
            total_fulfilled = $3,
            status = 'completed',
            processed_at = NOW()
        WHERE epoch_number = $1
    `, epochNumber, totalRequested.String(), totalFulfilled.String())

    return mapError(err)
}
// =====================================================
// REPUTATION METHODS
// =====================================================

/*
@method UpdateBorrowerReputation
@desc Updates borrower reputation score from ReputationScoreUpdated event
@param borrowerAddress Borrower address
@param newScore New reputation score
*/
func (r *EventRepository) UpdateBorrowerReputation(ctx context.Context, borrowerAddress string, newScore int32) error {
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE borrowers 
        SET reputation_score = $2,
            risk_label = CASE 
                WHEN $2 >= 900 THEN 'AAA'
                WHEN $2 >= 800 THEN 'AA'
                WHEN $2 >= 700 THEN 'A'
                WHEN $2 >= 500 THEN 'B'
                WHEN $2 >= 300 THEN 'C'
                ELSE 'D'
            END,
            last_reputation_update = NOW()
        WHERE address = $1
    `, borrowerAddress, newScore)
    
    return mapError(err)
}

/*
@method RegisterBorrowerFromEvent
@desc Creates borrower record from BorrowerRegistered event
@param borrowerAddress Borrower address
*/
func (r *EventRepository) RegisterBorrowerFromEvent(ctx context.Context, borrowerAddress string) error {
    _, err := r.db.conn.ExecContext(ctx, `
        INSERT INTO borrowers (address, reputation_score, risk_label, registered_at)
        VALUES ($1, 500, 'B', NOW())
        ON CONFLICT (address) DO NOTHING
    `, borrowerAddress)
    
    return mapError(err)
}
/*@
 * UpdateBorrowerWhitelistStatus
 * @desc Updates borrower whitelist status (ArchController)
 * @param ctx Context for cancellation
 * @param borrowerAddress Borrower address
 * @param isWhitelisted Whether borrower is whitelisted
 * @returns error if update fails
 *
 * @sql UPDATE borrowers SET is_whitelisted = $2, last_activity = NOW() WHERE address = $1
 * @note This may require adding is_whitelisted column to borrowers table
 */
func (r *EventRepository) UpdateBorrowerWhitelistStatus(ctx context.Context, borrowerAddress string, isWhitelisted bool) error {
    // First check if borrower exists
    var exists bool
    err := r.db.conn.QueryRowContext(ctx, `
        SELECT EXISTS(SELECT 1 FROM borrowers WHERE address = $1)
    `, borrowerAddress).Scan(&exists)
    
    if err != nil {
        return mapError(err)
    }
    
    if !exists {
        // Create borrower if doesn't exist
        _, err = r.db.conn.ExecContext(ctx, `
            INSERT INTO borrowers (address, reputation_score, risk_label, is_whitelisted, registered_at)
            VALUES ($1, 500, 'B', $2, NOW())
        `, borrowerAddress, isWhitelisted)
        return mapError(err)
    }
    
    // Update existing borrower
    _, err = r.db.conn.ExecContext(ctx, `
        UPDATE borrowers 
        SET is_whitelisted = $2,
            last_activity = NOW()
        WHERE address = $1
    `, borrowerAddress, isWhitelisted)
    
    return mapError(err)
}

// =====================================================
// CHECKPOINT METHODS
// =====================================================

/*
@method SaveCheckpoint
@desc Saves a chain checkpoint for reorg protection
@param blockNumber Block number
@param blockHash Block hash
*/
func (r *EventRepository) SaveCheckpoint(ctx context.Context, blockNumber uint64, blockHash string) error {
    _, err := r.db.conn.ExecContext(ctx, `
        INSERT INTO chain_checkpoints (block_number, block_hash, parent_hash, processed_at)
        VALUES ($1, $2, '', NOW())
        ON CONFLICT (block_number) DO UPDATE SET
            block_hash = EXCLUDED.block_hash,
            processed_at = EXCLUDED.processed_at
    `, blockNumber, blockHash)
    
    return mapError(err)
}

/*
@method GetLastCheckpoint
@desc Retrieves the last processed checkpoint
@return Last checkpoint or nil if none exists
*/
func (r *EventRepository) GetLastCheckpoint(ctx context.Context) (*models.ChainCheckpoint, error) {
    var checkpoint models.ChainCheckpoint
    err := r.db.conn.QueryRowContext(ctx, `
        SELECT block_number, block_hash, processed_at
        FROM chain_checkpoints
        ORDER BY block_number DESC
        LIMIT 1
    `).Scan(&checkpoint.BlockNumber, &checkpoint.BlockHash, &checkpoint.ProcessedAt)
    
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, mapError(err)
    }
    
    return &checkpoint, nil
}


/*@
 * UpdateOffer
 * @desc Updates offer with new parameters after modification
 * @param ctx Context for cancellation
 * @param offerID Offer ID to update
 * @param newAmount New total amount (may be different from remaining_amount)
 * @param newAPR New APR in basis points
 * @param newExpiry New expiry timestamp
 * @param status New status (usually "active" unless amount became 0)
 * @returns error if update fails
 *
 * @sql UPDATE offers SET amount=$2, remaining_amount=$2, apr=$3, expiry=$4, status=$5, updated_at=NOW()
 * @note When amount changes, remaining_amount should also update to match new amount
 */
func (r *EventRepository) UpdateOffer(
    ctx context.Context,
    offerID int64,
    newAmount *big.Int,
    newAPR *big.Int,
    newExpiry time.Time,
    status string,
) error {
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE offers 
        SET amount = $2,
            remaining_amount = $2,
            apr = $3,
            expiry = $4,
            status = $5,
            updated_at = NOW()
        WHERE offer_id = $1
    `, offerID, newAmount.String(), newAPR.Int64(), newExpiry, status)

    return mapError(err)
}

/*@
 * FailedEventRecord
 * @desc Represents a failed event from the database
 */
type FailedEventRecord struct {
    ID           int64
    EventName    string
    EventData    json.RawMessage
    RawLog       *json.RawMessage
    FailureReason string
    RetryCount   int
    MaxRetries   int
    NextRetryAt  *time.Time
    Processed    bool
    ResolvedAt   *time.Time
    CreatedAt    time.Time
     TxHash        string      
    BlockNumber   uint64      
    
}

/*@
 * GetUnresolvedFailedEvents
 * @desc Retrieves failed events that are ready for retry
 * @param ctx Context for cancellation
 * @param limit Maximum number of events to retrieve
 * @returns Slice of failed event records and error
 */
func (r *EventRepository) GetUnresolvedFailedEvents(ctx context.Context, limit int) ([]*FailedEventRecord, error) {
    rows, err := r.db.conn.QueryContext(ctx, `
        SELECT id, event_name, event_data, raw_log, failure_reason, 
               retry_count, max_retries, next_retry_at, processed, 
               resolved_at, created_at
        FROM failed_events
        WHERE processed = false
          AND (next_retry_at IS NULL OR next_retry_at <= NOW())
        ORDER BY created_at ASC
        LIMIT $1
    `, limit)
    
    if err != nil {
        return nil, mapError(err)
    }
    defer rows.Close()
    
    var events []*FailedEventRecord
    for rows.Next() {
        var event FailedEventRecord
        var nextRetryAt sql.NullTime
        var resolvedAt sql.NullTime
        var rawLog sql.NullString
        
        err := rows.Scan(
            &event.ID, &event.EventName, &event.EventData, &rawLog,
            &event.FailureReason, &event.RetryCount, &event.MaxRetries,
            &nextRetryAt, &event.Processed, &resolvedAt, &event.CreatedAt,
        )
        if err != nil {
            return nil, mapError(err)
        }
        
        if nextRetryAt.Valid {
            event.NextRetryAt = &nextRetryAt.Time
        }
        if resolvedAt.Valid {
            event.ResolvedAt = &resolvedAt.Time
        }
        if rawLog.Valid && rawLog.String != "" {
            var rm json.RawMessage
            if err := json.Unmarshal([]byte(rawLog.String), &rm); err == nil {
                event.RawLog = &rm
            }
        }
        
        events = append(events, &event)
    }
    
    return events, rows.Err()
}

/*@
 * MarkFailedEventResolved
 * @desc Marks a failed event as resolved (success or permanent failure)
 * @param ctx Context for cancellation
 * @param eventID ID of the failed event
 * @param success Whether event was processed successfully
 * @returns error if update fails
 */
func (r *EventRepository) MarkFailedEventResolved(ctx context.Context, eventID int64, success bool) error {
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE failed_events 
        SET processed = true, 
            resolved_at = NOW(),
            failure_reason = CASE WHEN $2 THEN NULL ELSE failure_reason END
        WHERE id = $1
    `, eventID, success)
    
    return mapError(err)
}

/*@
 * IncrementFailedEventRetry
 * @desc Increments retry count and sets next retry time with exponential backoff
 * @param ctx Context for cancellation
 * @param eventID ID of the failed event
 * @param newError New error message to append
 * @returns error if update fails
 */
func (r *EventRepository) IncrementFailedEventRetry(ctx context.Context, eventID int64, newError string) error {
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE failed_events 
        SET retry_count = retry_count + 1,
            failure_reason = failure_reason || '\n' || $2,
            next_retry_at = NOW() + (POW(2, retry_count + 1) || ' seconds')::INTERVAL
        WHERE id = $1
    `, eventID, newError)
    
    return mapError(err)
}

/*@
 * StoreFailedEvent
 * @desc Stores a failed event for later retry
 * @param ctx Context for cancellation
 * @param eventName Name of the event
 * @param eventData Raw event data
 * @param rawLog Original blockchain log
 * @param reason Failure reason
 * @param maxRetries Maximum retry attempts
 * @returns error if insert fails
 */
func (r *EventRepository) StoreFailedEvent(
    ctx context.Context,
    eventName string,
    eventData interface{},
    rawLog interface{},
    reason string,
    maxRetries int,
) error {
    eventDataJSON, err := json.Marshal(eventData)
    if err != nil {
        return err
    }
    
    var rawLogJSON []byte
    if rawLog != nil {
        rawLogJSON, err = json.Marshal(rawLog)
        if err != nil {
            return err
        }
    }
    
    _, err = r.db.conn.ExecContext(ctx, `
        INSERT INTO failed_events (
            event_name, event_data, raw_log, failure_reason, 
            max_retries, created_at
        ) VALUES ($1, $2, $3, $4, $5, NOW())
    `, eventName, eventDataJSON, rawLogJSON, reason, maxRetries)
    
    return mapError(err)
}



/*@
 * RecordLiquidation
 * @desc Records a liquidation event in the database
 * @param ctx Context for cancellation
 * @param borrower Address of liquidated borrower
 * @param debtAmount Amount of debt liquidated
 * @param collateralAmount Amount of collateral seized
 * @param blockNum Block number where liquidation occurred
 */
func (r *EventRepository) RecordLiquidation(ctx context.Context, borrower string, debtAmount, collateralAmount *big.Int, blockNum uint64) error {
    query := `
        INSERT INTO liquidations (borrower_address, debt_amount, collateral_amount, block_number, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `
    
    _, err := r.db.ExecContext(ctx, query, 
        borrower, 
        debtAmount.String(), 
        collateralAmount.String(), 
        blockNum, 
        time.Now(),
    )
    return err
}

/*@
 * CompleteEpoch
 * @desc Completes a withdrawal epoch and updates all associated requests
 * @param ctx Context for cancellation
 * @param epoch Epoch data to save
 * @param blockNum Block number where epoch was processed
 */
func (r *EventRepository) CompleteEpoch(ctx context.Context, epoch *models.WithdrawalEpoch, blockNum uint64) error {
    // Start transaction
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Save epoch record
    query := `
        INSERT INTO withdrawal_epochs (epoch_number, total_requested, total_fulfilled, status, processed_at, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (epoch_number) DO UPDATE SET
            total_fulfilled = EXCLUDED.total_fulfilled,
            status = EXCLUDED.status,
            processed_at = EXCLUDED.processed_at
    `
    
    _, err = tx.ExecContext(ctx, query,
        epoch.EpochNumber,
        epoch.TotalRequested.String(),
        epoch.TotalFulfilled.String(),
        epoch.Status,
        epoch.ProcessedAt,
        epoch.CreatedAt,
    )
    if err != nil {
        return err
    }
    
    // Update all withdrawal requests in this epoch
    updateQuery := `
        UPDATE withdrawal_requests 
        SET status = 'claimed',
            fulfillment_time = $1,
            block_number = $2
        WHERE epoch_number = $3 AND status = 'pending'
    `
    
    _, err = tx.ExecContext(ctx, updateQuery, time.Now(), blockNum, epoch.EpochNumber)
    if err != nil {
        return err
    }
    
    return tx.Commit()
}

func (r *EventRepository) DecrementOfferRemaining(ctx context.Context, offerID int64, filledAmount *big.Int) error {
    _, err := r.db.conn.ExecContext(ctx, `
        UPDATE offers 
        SET remaining_amount = remaining_amount::numeric - $2,
            status = CASE 
                WHEN remaining_amount::numeric - $2 <= 0 THEN 'filled'
                ELSE 'partially_filled'
            END,
            updated_at = NOW()
        WHERE offer_id = $1
    `, offerID, filledAmount.String())
    
    return mapError(err)
}