package response

/*
@file admin_borrowers.go

@desc
Response DTOs for borrower management admin endpoints.
Structures borrower information for API responses.

@structure
- BorrowerInfo: Individual borrower details with status and reputation
- BorrowerListResponse: Paginated list of borrowers
- PendingBorrowerResponse: Borrowers awaiting verification
- BorrowerCountResponse: Summary statistics on borrowers
*/

/*
@struct BorrowerInfo

@desc
Complete information about a single borrower.
Includes verification status, reputation metrics, and platform activity.

@fields
- address: Ethereum address of borrower
- is_verified: Whether borrower is verified/approved
- reputation_score: Numerical reputation score (0-1000)
- total_borrowed: Total amount borrowed (in wei, as string for precision)
- total_repaid: Total amount repaid (in wei)
- active_offers: Number of active loan offers
- default_count: Number of defaults (0 if perfect history)
- added_at: Timestamp when added to platform
- verified_at: Timestamp when verified (null if not verified)
*/
type BorrowerInfo struct {
	Address      string `json:"address"`
	IsVerified   bool   `json:"is_verified"`
	ReputationScore int32   `json:"reputation_score"`
	TotalBorrowed   string `json:"total_borrowed"`
	TotalRepaid     string `json:"total_repaid"`
	ActiveOffers    int32   `json:"active_offers"`
	DefaultCount    int32   `json:"default_count"`
	AddedAt         int64   `json:"added_at"`
	VerifiedAt      int64   `json:"verified_at,omitempty"`
}

/*
@struct BorrowerListResponse

@desc
Paginated list of borrowers with metadata.
Used for listing all borrowers or filtered borrower sets.

@fields
- borrowers: Array of borrower information objects
- pagination: Pagination metadata (page, limit, total)
- total_reputation_average: Average reputation score across all listed borrowers
*/
type BorrowerListResponse struct {
	Borrowers             []BorrowerInfo `json:"borrowers"`
	Pagination            *PaginationInfo `json:"pagination"`
	TotalReputationAverage float64        `json:"total_reputation_average"`
}

/*
@struct PendingBorrowerResponse

@desc
Response for listing borrowers pending verification.
Includes verification request details and timeline.

@fields
- pending_count: Number of borrowers awaiting verification
- borrowers: Array of borrower pending verification
- oldest_pending_at: Timestamp of oldest pending request
*/
type PendingBorrowerResponse struct {
	PendingCount     int32           `json:"pending_count"`
	Borrowers        []BorrowerInfo  `json:"borrowers"`
	OldestPendingAt  int64           `json:"oldest_pending_at,omitempty"`
}

/*
@struct BorrowerCountResponse

@desc
Summary statistics about borrowers on the platform.
Provides high-level metrics for dashboard.

@fields
- total_borrowers: Total number of borrowers
- verified_borrowers: Count of verified/approved borrowers
- pending_borrowers: Count awaiting verification
- blocked_borrowers: Count of blocked borrowers
- active_borrowers: Count with active loans
- average_reputation: Average reputation score across all borrowers
*/
type BorrowerCountResponse struct {
	TotalBorrowers    int64   `json:"total_borrowers"`
	VerifiedBorrowers int64   `json:"verified_borrowers"`
	PendingBorrowers  int64   `json:"pending_borrowers"`
	BlockedBorrowers  int64   `json:"blocked_borrowers"`
	ActiveBorrowers   int64   `json:"active_borrowers"`
	AverageReputation float64 `json:"average_reputation"`
}
