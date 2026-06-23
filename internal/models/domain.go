package models

import (
	"database/sql"
	"math/big"
	"time"
)

/*
@struct Market

@desc
Represents an isolated lending pool created by a borrower.
Contains all market configuration and state.

@fields
- ID: database identifier
- Address: smart contract address
- Borrower: market creator address
- BorrowAsset: ERC20 token address to borrow
- CollateralAsset: ERC20 collateral token address
- CollateralOracle: Chainlink oracle address for price feeds
- MinCollateralRatio: minimum collateral ratio in bps (11000 = 110%)
- LiquidationThreshold: liquidation trigger in bps (9500 = 95%)
- TotalPrincipal: total outstanding principal
- TotalAccruedInterest: accumulated interest
- TotalLiquidity: available liquidity in offers
- BorrowIndex: tracks interest accrual (scaled to 1e18)
- WeightedAvgAPR: weighted average APR across offers (in bps)
- UtilizationRate: calculated from totalDebt / totalLiquidity
- IsActive: market operational state
- IsLiquidating: active liquidation state
- IsClosed: market closed state
- CreatedAt: market creation timestamp
- LastInterestAccrual: last time interest was accrued
*/
type Market struct {
	ID                    int64
	Address               string
	Borrower              string
	BorrowAsset           string
	BorrowAssetSymbol     sql.NullString
	BorrowAssetDecimals   int16
	CollateralAsset       string
	CollateralAssetSymbol sql.NullString
	CollateralAssetDecimals int16
	CollateralOracle      string
	MinCollateralRatio    int32  // bps
	LiquidationThreshold  int32  // bps
	TotalPrincipal        *big.Int
	TotalAccruedInterest  *big.Int
	TotalDebt             *big.Int // calculated
	TotalLiquidity        *big.Int
	BorrowIndex           *big.Int // 1e18 scale
	WeightedAvgAPR        int32
	UtilizationRate       float64
	IsActive              bool
	IsLiquidating         bool
	IsClosed              bool
	CurrentAuctionID      sql.NullInt64
	CreatedAt             time.Time
	ClosedAt              sql.NullTime
	LastInterestAccrual   time.Time
	LastHealthCheck       sql.NullTime
}

/*
@struct Position

@desc
Represents a lender's NFT position in a market.
Tracks principal, accrued interest, and settlement state.

@fields
- ID: database identifier
- TokenID: NFT token ID
- Lender: position owner address
- MarketAddress: market this position belongs to
- Principal: original deposit amount
- CurrentPrincipal: principal after repayments
- AccruedInterest: unpaid interest
- ClaimableAmount: ready to claim amount
- APR: position APR in bps
- Seniority: 0=Senior, 1=Junior
- Status: active|settled|liquidated
- IsActive: position active flag
- IsSettled: position settled flag
- MintedAt: creation timestamp
- SettledAt: settlement timestamp
- BlockNumber: block where position was minted
- TxHash: transaction hash for auditing
*/
type Position struct {
	ID              int64
	TokenID         int64
	Lender          string
	MarketAddress   string
	Principal       *big.Int
	CurrentPrincipal *big.Int
	AccruedInterest *big.Int
	ClaimableAmount *big.Int
	APR             int32
	Seniority       int16 // 0=Senior, 1=Junior
	Status          string
	IsActive        bool
	IsSettled       bool
	MintedAt        time.Time
	SettledAt       sql.NullTime
	LastUpdated     time.Time
	BlockNumber     int64
	TxHash          string
	LogIndex        int32
}

/*
@struct Offer

@desc
Represents a lender's liquidity commitment with APR and seniority tier.
Tracks filled amount and expiry.

@fields
- ID: database identifier
- OfferID: blockchain offer ID
- Lender: offer creator address
- MarketAddress: market this offer is for
- Amount: total offer amount
- RemainingAmount: unfilled amount
- APR: annual percentage rate in bps
- Seniority: 0=Senior, 1=Junior
- Status: active|partially_filled|filled|cancelled|expired
- Expiry: offer expiration timestamp
- CreatedAt: offer creation timestamp
- ExpiryTimestamp: when offer expires on chain
*/
type Offer struct {
	ID              int64
	OfferID         int64
	Lender          string
	MarketAddress   string
	Amount          *big.Int
	RemainingAmount *big.Int
	APR             int32
	Seniority       int16 // 0=Senior, 1=Junior
	Status          string
	Expiry          time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CancelledAt     sql.NullTime
	ExpiredAt       sql.NullTime
	BlockNumber     int64
	TxHash          string
}

/*
@struct Auction

@desc
Represents a liquidation Dutch auction for a market.
Tracks collateral, debt, and bidding state.

@fields
- ID: database identifier
- AuctionID: blockchain auction ID
- MarketAddress: market being liquidated
- BorrowerAddress: borrower address
- CollateralAmount: collateral being auctioned
- DebtAmount: total debt at liquidation
- CurrentPrice: current Dutch auction price
- HighestBid: highest bid so far
- HighestBidder: highest bidder address
- StartTime: auction start timestamp
- EndTime: auction end timestamp
- Status: active|settled|cancelled|failed
- Winner: auction winner address (if settled)
*/
type Auction struct {
	ID              int64
	AuctionID       int64
	MarketAddress   string
	BorrowerAddress string
	CollateralAmount *big.Int
	CollateralValue  *big.Int
	DebtAmount      *big.Int
	HighestBid      *big.Int
	HighestBidder   sql.NullString
	CurrentPrice    *big.Int
	StartTime       time.Time
	EndTime         time.Time
	SettlementTime  sql.NullTime
	Status          string
	WinningBid      *big.Int
	Winner          sql.NullString
	RecoveryRate    float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

/*
@struct Borrower

@desc
Represents a borrower profile and reputation tracking.
Stores reputation score, loan statistics, and history.

@fields
- Address: borrower wallet address (primary key)
- TotalBorrowed: cumulative borrowed amount
- TotalRepaid: cumulative repaid amount
- OutstandingDebt: calculated as total_borrowed - total_repaid
- TotalLoans: number of loans
- SuccessfulLoans: repaid loans
- DefaultedLoans: liquidated loans
- ActiveLoans: currently active loans
- ReputationScore: 0-1000 score
- RiskLabel: AAA|AA|A|B|C|D (calculated)
- SuccessRate: percentage of successful loans
- RegisteredAt: account creation timestamp
- LastActivity: last transaction timestamp
*/
type Borrower struct {
	Address             string
	TotalBorrowed       *big.Int
	TotalRepaid         *big.Int
	OutstandingDebt     *big.Int // calculated
	TotalLoans          int32
	SuccessfulLoans     int32
	DefaultedLoans      int32
	ActiveLoans         int32
	ReputationScore     int32
	RiskLabel           string
	SuccessRate         float64
	RegisteredAt        time.Time
	LastActivity        sql.NullTime
	LastReputationUpdate time.Time
}

/*
@struct WithdrawalRequest

@desc
Represents a lender's withdrawal request queued for an epoch.
Tracks fulfillment state across epochs.

@fields
- ID: database identifier
- RequestID: blockchain request ID
- Lender: requester address
- PositionID: position being withdrawn from
- MarketAddress: market address
- RequestedAmount: amount requested
- FulfilledAmount: amount fulfilled
- EpochNumber: withdrawal epoch
- Status: pending|ready|claimed|expired
*/
type WithdrawalRequest struct {
	ID               int64
	RequestID        int64
	Lender           string
	PositionID       int64
	MarketAddress    string
	RequestedAmount  *big.Int
	FulfilledAmount  *big.Int
	RemainingAmount  *big.Int // calculated
	EpochNumber      int32
	Status           string
	FulfillmentTime  sql.NullTime
	FulfillmentTxHash sql.NullString
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

/*
@struct WithdrawalEpoch

@desc
Represents a withdrawal epoch tracking aggregate withdrawal state.
Used for epoch-based withdrawal processing.

@fields
- ID: database identifier
- EpochNumber: epoch sequence number
- StartTime: epoch start
- EndTime: epoch end
- ProcessedAt: processing timestamp
- TotalRequested: total requested in epoch
- TotalFulfilled: total fulfilled in epoch
- Status: pending|processing|completed|failed
*/
type WithdrawalEpoch struct {
	ID            int64
	EpochNumber   int32
	StartTime     time.Time
	EndTime       time.Time
	ProcessedAt   sql.NullTime
	TotalRequested *big.Int
	TotalFulfilled *big.Int
	Status        string
	CreatedAt     time.Time
}

/*
@struct Repayment

@desc
Records repayment transactions and interest distribution.

@fields
- ID: database identifier
- MarketAddress: market for repayment
- BorrowerAddress: repaying borrower
- Amount: total amount repaid
- InterestPaid: interest portion
- PrincipalPaid: principal portion
- AffectedPositionIDs: array of affected position IDs
- TxHash: blockchain transaction hash
- CreatedAt: repayment timestamp
*/
type Repayment struct {
	ID                  int64
	MarketAddress       string
	BorrowerAddress     string
	Amount              *big.Int
	InterestPaid        *big.Int
	PrincipalPaid       *big.Int
	RepaymentType       string // partial|full|liquidation
	AffectedPositionIDs []int64
	BlockNumber         int64
	TxHash              string
	CreatedAt           time.Time
}

/*
@struct AuthSession

@desc
Represents user authentication session with JWT token.
Tracks active sessions for SIWE (Sign-In With Ethereum) authentication.

@fields
- ID: database identifier
- WalletAddress: user's Ethereum wallet address
- Token: JWT token
- ExpiresAt: token expiration time
- RevokedAt: revocation timestamp (null if active)
- CreatedAt: session creation time
*/
type AuthSession struct {
	ID            int64
	WalletAddress string
	Token         string
	ExpiresAt     time.Time
	RevokedAt     sql.NullTime
	CreatedAt     time.Time
}

/*
@struct CollateralBalance

@desc
Tracks a borrower's collateral balance held in the RevvFiCollateralEscrow.
Populated from CollateralDeposited and CollateralWithdrawn events.
The Balance field reflects the authoritative on-chain total from the event's
Total field; cumulative fields enable historical analysis.

@fields
- ID: database identifier
- Borrower: borrower wallet address (unique per escrow)
- Balance: current total collateral (authoritative, from event Total field)
- TotalDeposited: cumulative sum of all deposit amounts
- TotalWithdrawn: cumulative sum of all withdrawal amounts
- LastDepositAt: timestamp of most recent deposit
- LastWithdrawalAt: timestamp of most recent withdrawal
- UpdatedAt: last record update timestamp
*/
type CollateralBalance struct {
	ID               int64
	Borrower         string
	Balance          *big.Int
	TotalDeposited   *big.Int
	TotalWithdrawn   *big.Int
	LastDepositAt    sql.NullTime
	LastWithdrawalAt sql.NullTime
	UpdatedAt        time.Time
}

/*
@struct ControllerFactory

@desc
Represents a controller factory registered in the RevvFiArchController.
Populated from ControllerFactoryAdded/Removed events.
A factory can create new market controller instances.

@fields
- ID: database identifier
- Address: factory contract address
- AddedAt: timestamp when factory was registered
- RemovedAt: timestamp when factory was deregistered (null if active)
- BlockNumber: block number of the registration event
*/
type ControllerFactory struct {
	ID          int64
	Address     string
	AddedAt     time.Time
	RemovedAt   sql.NullTime
	BlockNumber int64
}

/*
@struct Controller

@desc
Represents a controller registered in the RevvFiArchController.
Populated from ControllerAdded/Removed events.
Controllers govern market creation and administrative functions.

@fields
- ID: database identifier
- Address: controller contract address (from ControllerFactory field)
- AddedAt: timestamp when controller was registered
- RemovedAt: timestamp when controller was deregistered (null if active)
- BlockNumber: block number of the registration event
*/
type Controller struct {
	ID          int64
	Address     string
	AddedAt     time.Time
	RemovedAt   sql.NullTime
	BlockNumber int64
}
