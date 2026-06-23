package response

/*
@file admin_reputation.go

@desc
Response DTOs for borrower reputation management admin endpoints.

@responsibilities
- Structure reputation query and listing responses
*/

/*
@struct BorrowerReputation

@desc
Detailed reputation information for a specific borrower.

@fields
- Address: borrower wallet address
- ReputationScore: current score (0-1000)
- RiskLabel: human-readable risk tier (AAA/AA/A/B/C/D)
- SuccessRate: percentage of loans repaid successfully (0.0-1.0)
- TotalLoans: total number of loans taken
- ActiveLoans: currently active loan count
- DefaultedLoans: number of defaulted loans
- TotalBorrowed: cumulative borrowed amount (wei string)
- LastActivity: Unix timestamp of last protocol interaction
*/
type BorrowerReputation struct {
	Address         string  `json:"address"`
	ReputationScore int32   `json:"reputation_score"`
	RiskLabel       string  `json:"risk_label"`
	SuccessRate     float64 `json:"success_rate"`
	TotalLoans      int32   `json:"total_loans"`
	ActiveLoans     int32   `json:"active_loans"`
	DefaultedLoans  int32   `json:"defaulted_loans"`
	TotalBorrowed   string  `json:"total_borrowed"`
	LastActivity    int64   `json:"last_activity,omitempty"`
}

/*
@struct DefaultedBorrowersResponse

@desc
Paginated response for borrowers with at least one defaulted loan.

@fields
- Borrowers: list of borrower info objects
- Pagination: pagination metadata
- TotalDefaulted: total number of defaulted borrowers
*/
type DefaultedBorrowersResponse struct {
	Borrowers      []BorrowerInfo  `json:"borrowers"`
	Pagination     *PaginationInfo `json:"pagination"`
	TotalDefaulted int64           `json:"total_defaulted"`
}
