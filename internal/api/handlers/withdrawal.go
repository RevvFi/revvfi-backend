package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/request"
	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	"github.com/Revvfi/revvfi-backend/internal/core/withdrawal"
)

/*
@struct WithdrawalHandler

@desc
HTTP handler for withdrawal endpoints.

@responsibilities
- Queue withdrawal requests
- Cancel pending requests
- Return request and epoch state
*/
type WithdrawalHandler struct {
	service *withdrawal.WithdrawalService
}

/*
@function NewWithdrawalHandler

@desc
Creates a withdrawal HTTP handler.

@params
- service: withdrawal core service

@returns
- *WithdrawalHandler
*/
func NewWithdrawalHandler(service *withdrawal.WithdrawalService) *WithdrawalHandler {
	return &WithdrawalHandler{service: service}
}

/*
@method Request

@desc
Handles authenticated withdrawal queue requests.

@params
- c: Gin context
*/
func (h *WithdrawalHandler) Request(c *gin.Context) {
	var req request.WithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}
	amount, err := parseBigInt(req.Amount)
	if err != nil {
		writeError(c, err)
		return
	}
	created, err := h.service.RequestWithdrawal(c.Request.Context(), c.GetString("wallet"), req.PositionID, amount)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusCreated, withdrawalResponse(created))
}

/*
@method Cancel

@desc
Handles authenticated withdrawal cancellation requests.

@params
- c: Gin context
*/
func (h *WithdrawalHandler) Cancel(c *gin.Context) {
	var req request.WithdrawalCancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}
	if err := h.service.CancelWithdrawalRequest(c.Request.Context(), req.RequestID, c.GetString("wallet")); err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, gin.H{"message": "withdrawal cancelled"})
}

/*
@method ListMine

@desc
Handles authenticated withdrawal request listing.

@params
- c: Gin context
*/
func (h *WithdrawalHandler) ListMine(c *gin.Context) {
	requests, err := h.service.GetWithdrawalRequests(c.Request.Context(), c.GetString("wallet"), 50, 0)
	if err != nil {
		writeError(c, err)
		return
	}
	items := make([]response.WithdrawalRequestResponse, 0, len(requests))
	for i := range requests {
		items = append(items, withdrawalResponse(&requests[i]))
	}
	ok(c, http.StatusOK, gin.H{"withdrawals": items, "count": len(items)})
}

/*
@method CurrentEpoch

@desc
Handles current withdrawal epoch requests.

@params
- c: Gin context
*/
func (h *WithdrawalHandler) CurrentEpoch(c *gin.Context) {
	epoch, err := h.service.GetCurrentEpoch(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, response.WithdrawalEpochResponse{
		EpochNumber:      epoch.EpochNumber,
		StartTime:        epoch.StartTime.Unix(),
		EndTime:          epoch.EndTime.Unix(),
		Status:           epoch.Status,
		TotalRequested:   stringOrZero(epoch.TotalRequested),
		TotalFulfilled:   stringOrZero(epoch.TotalFulfilled),
		TotalUnfulfilled: "0",
	})
}
