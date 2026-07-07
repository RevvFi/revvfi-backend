package response

/*
@file borrower_request.go

@desc
Response DTOs for the off-chain borrower access request queue.
*/

/*
@struct BorrowerRequestInfo

@desc
A single borrower access request.

@fields
- id: request ID
- wallet_address: the requesting wallet
- status: pending, approved, rejected
- note: optional admin note (set on rejection)
- requested_at: unix timestamp the request was submitted
- decided_at: unix timestamp of admin/on-chain decision (0 if still pending)
- decided_by: admin wallet that rejected, or "on-chain" if auto-approved from a BorrowerAdded event
*/
type BorrowerRequestInfo struct {
	ID            int64  `json:"id"`
	WalletAddress string `json:"wallet_address"`
	Status        string `json:"status"`
	Note          string `json:"note,omitempty"`
	RequestedAt   int64  `json:"requested_at"`
	DecidedAt     int64  `json:"decided_at,omitempty"`
	DecidedBy     string `json:"decided_by,omitempty"`
}

/*
@struct BorrowerRequestListResponse
*/
type BorrowerRequestListResponse struct {
	Count    int                   `json:"count"`
	Requests []BorrowerRequestInfo `json:"requests"`
}
