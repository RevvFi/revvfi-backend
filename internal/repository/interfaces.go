package repository

import (
"context"
"time"

"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@file interfaces.go

@desc
Repository interfaces defining data access contracts.
Implements repository pattern for dependency injection and testing.

@interfaces
- AuthRepository: authentication data access
- MarketRepository: market state access
- OfferRepository: offer book access
- PositionRepository: lender position access
- BorrowerRepository: borrower profile access
- AuctionRepository: liquidation auction access
- WithdrawalRepository: withdrawal request access
- EventRepository: blockchain event processing access
*/

/*
@interface AuthRepository

@desc
Data access for authentication and sessions.
Handles nonces, sessions, and token revocation.

@methods
- StoreNonce: Store SIWE nonce with TTL
- ValidateNonce: Check nonce validity
- StoreSession: Create auth session
- RevokeSession: Revoke JWT token
- IsSessionRevoked: Check revocation status
*/
type AuthRepository interface {
	StoreNonce(ctx context.Context, wallet, nonce string, expiresAt time.Time) error
	ValidateNonce(ctx context.Context, wallet, nonce string) error
	StoreSession(ctx context.Context, session *models.AuthSession) error
	RevokeSession(ctx context.Context, token string) error
	IsSessionRevoked(ctx context.Context, token string) (bool, error)
	GetNonce(ctx context.Context, wallet string) (string, error)
	DeleteNonce(ctx context.Context, wallet string) error
	ValidateNonceExists(ctx context.Context, wallet string, nonce string) error
	ConsumeNonceAtomic(ctx context.Context, wallet string, nonce string) error
	ConsumeNonce(ctx context.Context, wallet string, nonce string) error
}

/*
@interface MarketRepository

@desc
Data access for lending markets.
Queries market state, configuration, and metrics.

@methods
- GetByAddress: Fetch market by address
- List: List markets with pagination
- Create: Create new market record
- Update: Update market state
- GetMetrics: Fetch market metrics
*/
type MarketRepository interface {
	GetByAddress(ctx context.Context, address string) (*models.Market, error)
	List(ctx context.Context, page, pageSize int32, filter map[string]interface{}) ([]models.Market, int64, error)
	Create(ctx context.Context, market *models.Market) error
	Update(ctx context.Context, market *models.Market) error
	GetMetrics(ctx context.Context, address string) (map[string]interface{}, error)
}

/*
@interface OfferRepository

@desc
Data access for lender offers.
Manages offer lifecycle and matching queries.

@methods
- GetByID: Fetch offer by ID
- GetByMarket: Get offers for market
- ListActive: List active offers
- Create: Create new offer
- Update: Update offer (fill amount, status)
- Cancel: Cancel offer
- GetBestOffers: Get best offers for matching
*/
type OfferRepository interface {
	GetByID(ctx context.Context, offerID int64) (*models.Offer, error)
	GetByMarket(ctx context.Context, market string, page, pageSize int32) ([]models.Offer, int64, error)
	GetByLender(ctx context.Context, lender string, page, pageSize int32) ([]models.Offer, int64, error)
	ListActive(ctx context.Context, market string, seniority int16) ([]models.Offer, error)
	Create(ctx context.Context, offer *models.Offer) error
	Update(ctx context.Context, offer *models.Offer) error
	Cancel(ctx context.Context, offerID int64) error
	GetBestOffers(ctx context.Context, market string, seniority int16, limit int32) ([]models.Offer, error)
}

/*
@interface PositionRepository

@desc
Data access for lender positions.
Tracks position state, valuations, and portfolio queries.

@methods
- GetByTokenID: Fetch position by NFT token ID
- GetByLender: Get lender portfolio
- ListClaimable: Get claimable positions
- Create: Create new position
- Update: Update position state
- GetPortfolioValue: Calculate portfolio value
*/
type PositionRepository interface {
	GetByTokenID(ctx context.Context, tokenID int64) (*models.Position, error)
	GetByLender(ctx context.Context, lender string, page, pageSize int32) ([]models.Position, int64, error)
	GetByMarket(ctx context.Context, market string, page, pageSize int32) ([]models.Position, int64, error)
	ListClaimable(ctx context.Context, lender string) ([]models.Position, error)
	Create(ctx context.Context, position *models.Position) error
	Update(ctx context.Context, position *models.Position) error
	GetPortfolioSummary(ctx context.Context, lender string) (map[string]interface{}, error)
}

/*
@interface BorrowerRepository

@desc
Data access for borrower profiles and reputation.
Tracks borrower history and metrics.

@methods
- GetByAddress: Fetch borrower profile
- Create: Create borrower account
- UpdateReputation: Update reputation score
- IncrementLoan: Track new loan
- RecordRepayment: Record repayment
- RecordDefault: Record default/liquidation
- GetLeaderboard: Get borrower rankings
*/
type BorrowerRepository interface {
	GetByAddress(ctx context.Context, address string) (*models.Borrower, error)
	Create(ctx context.Context, borrower *models.Borrower) error
	Update(ctx context.Context, borrower *models.Borrower) error
	UpdateReputation(ctx context.Context, address string, score int32) error
	IncrementLoan(ctx context.Context, address string) error
	RecordRepayment(ctx context.Context, repayment *models.Repayment) error
	GetLeaderboard(ctx context.Context, limit int32) ([]models.Borrower, error)
}

/*
@interface AuctionRepository

@desc
Data access for liquidation auctions.
Tracks auction state, bids, and settlements.

@methods
- GetByID: Fetch auction by ID
- GetActive: Get active auctions
- Create: Create new auction
- UpdatePrice: Update current Dutch price
- RecordBid: Record auction bid
- Settle: Settle auction
- GetHistory: Get liquidation history
*/
type AuctionRepository interface {
	GetByID(ctx context.Context, auctionID int64) (*models.Auction, error)
	GetByMarket(ctx context.Context, market string) (*models.Auction, error)
	GetActive(ctx context.Context, limit int32) ([]models.Auction, error)
	Create(ctx context.Context, auction *models.Auction) error
	Update(ctx context.Context, auction *models.Auction) error
	GetHistory(ctx context.Context, market string, limit int32) ([]models.Auction, error)
}

/*
@interface WithdrawalRepository

@desc
Data access for withdrawal requests and epochs.
Manages epoch-based withdrawal processing.

@methods
- GetRequest: Fetch withdrawal request
- ListByLender: Get lender requests
- ListByEpoch: Get epoch requests
- CreateRequest: Create withdrawal request
- UpdateRequest: Update request status
- GetCurrentEpoch: Get active epoch
- CreateEpoch: Create new epoch
- UpdateEpoch: Update epoch progress
*/
type WithdrawalRepository interface {
	GetRequest(ctx context.Context, requestID int64) (*models.WithdrawalRequest, error)
	ListByLender(ctx context.Context, lender string, page, pageSize int32) ([]models.WithdrawalRequest, int64, error)
	ListByEpoch(ctx context.Context, epochNumber int32, page, pageSize int32) ([]models.WithdrawalRequest, int64, error)
	CreateRequest(ctx context.Context, request *models.WithdrawalRequest) error
	UpdateRequest(ctx context.Context, request *models.WithdrawalRequest) error
	GetCurrentEpoch(ctx context.Context) (*models.WithdrawalEpoch, error)
	GetEpochByNumber(ctx context.Context, epochNumber int32) (*models.WithdrawalEpoch, error)
	CreateEpoch(ctx context.Context, epoch *models.WithdrawalEpoch) error
	UpdateEpoch(ctx context.Context, epoch *models.WithdrawalEpoch) error
}

/*
@interface EventRepository

@desc
Data access for event tracking and idempotency.
Prevents duplicate blockchain event processing.

@methods
- IsProcessed: Check if event already processed
- MarkProcessed: Mark event as processed
- GetFailedEvent: Fetch failed event
- StoreFailedEvent: Store failed event for retry
- GetFailedEvents: List failed events needing retry
*/
type EventRepository interface {
	IsProcessed(ctx context.Context, txHash string, logIndex int32) (bool, error)
	MarkProcessed(ctx context.Context, event *models.ProcessedEvent) error
	StoreFailedEvent(ctx context.Context, event map[string]interface{}, reason string) error
	GetFailedEvents(ctx context.Context, limit int32) ([]map[string]interface{}, error)
	CleanupOldEvents(ctx context.Context, olderThan time.Time) error
}
