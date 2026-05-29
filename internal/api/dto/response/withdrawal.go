package response

/*
@file withdrawal.go

@desc
Response DTOs for withdrawal endpoints.
Returns withdrawal requests and epoch data.
*/

/*
@struct WithdrawalRequestResponse

@desc
Response containing withdrawal request details.

@fields
- RequestID: request identifier
- Lender: requesting lender
- PositionID: position being withdrawn from
- MarketAddress: market address
- RequestedAmount: amount requested
- FulfilledAmount: amount already fulfilled
- RemainingAmount: amount still pending
- EpochNumber: withdrawal epoch
- Status: pending|ready|claimed|expired
- CreatedAt: request creation
*/
type WithdrawalRequestResponse struct {
	RequestID       int64  `json:"request_id"`
	Lender          string `json:"lender"`
	PositionID      int64  `json:"position_id"`
	MarketAddress   string `json:"market_address"`
	RequestedAmount string `json:"requested_amount"`
	FulfilledAmount string `json:"fulfilled_amount"`
	RemainingAmount string `json:"remaining_amount"`
	EpochNumber     int32  `json:"epoch_number"`
	Status          string `json:"status"`
	CreatedAt       int64  `json:"created_at"`
}

/*
@struct WithdrawalEpochResponse

@desc
Response containing withdrawal epoch information.

@fields
- EpochNumber: epoch sequence number
- StartTime: epoch start timestamp
- EndTime: epoch end timestamp
- Status: pending|processing|completed|failed
- TotalRequested: total withdrawal requests
- TotalFulfilled: total fulfilled
- TotalUnfulfilled: remaining unfulfilled
- ProcessedAt: processing timestamp
*/
type WithdrawalEpochResponse struct {
	EpochNumber     int32  `json:"epoch_number"`
	StartTime       int64  `json:"start_time"`
	EndTime         int64  `json:"end_time"`
	Status          string `json:"status"`
	TotalRequested  string `json:"total_requested"`
	TotalFulfilled  string `json:"total_fulfilled"`
	TotalUnfulfilled string `json:"total_unfulfilled"`
	ProcessedAt     *int64 `json:"processed_at,omitempty"`
}

/*
@struct CurrentWithdrawalEpochResponse

@desc
Response with current withdrawal epoch and timing.

@fields
- EpochNumber: current epoch number
- TimeRemaining: seconds until epoch end
- ProgressPercentage: epoch completion percentage
- PreviousEpoch: previous epoch details
*/
type CurrentWithdrawalEpochResponse struct {
	EpochNumber       int32                      `json:"epoch_number"`
	TimeRemaining     int64                      `json:"time_remaining"`
	ProgressPercentage float64                   `json:"progress_percentage"`
	PreviousEpoch     *WithdrawalEpochResponse `json:"previous_epoch,omitempty"`
}
