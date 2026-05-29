package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/request"
	"github.com/Revvfi/revvfi-backend/internal/core/transaction"
)

/*
@struct TransactionHandler

@desc
HTTP handler for transaction build and quote endpoints.

@responsibilities
- Bind transaction request DTOs
- Call transaction service methods
- Return unsigned wallet transaction payloads
*/
type TransactionHandler struct {
	service *transaction.TransactionService
}

/*
@function NewTransactionHandler

@desc
Creates a transaction HTTP handler.

@params
- service: transaction core service

@returns
- *TransactionHandler
*/
func NewTransactionHandler(service *transaction.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

/*
@method BuildBorrow

@desc
Handles borrow transaction build requests.

@params
- c: Gin context
*/
func (h *TransactionHandler) BuildBorrow(c *gin.Context) {
	var req request.BorrowTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}
	amount, err := parseBigInt(req.Amount)
	if err != nil {
		writeError(c, err)
		return
	}
	tx, err := h.service.BuildBorrowTransaction(c.Request.Context(), req.MarketAddress, amount, req.TokenID, req.MaxAPR)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, tx)
}

/*
@method BuildRepay

@desc
Handles repayment transaction build requests.

@params
- c: Gin context
*/
func (h *TransactionHandler) BuildRepay(c *gin.Context) {
	var req request.RepayTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}
	amount, err := parseBigInt(req.Amount)
	if err != nil {
		writeError(c, err)
		return
	}
	tx, err := h.service.BuildRepayTransaction(c.Request.Context(), req.MarketAddress, amount)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, tx)
}

/*
@method BuildLiquidate

@desc
Handles liquidation transaction build requests.

@params
- c: Gin context
*/
func (h *TransactionHandler) BuildLiquidate(c *gin.Context) {
	var req request.LiquidateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}
	amount, err := parseBigInt(req.BidAmount)
	if err != nil {
		writeError(c, err)
		return
	}
	tx, err := h.service.BuildLiquidateTransaction(c.Request.Context(), req.MarketAddress, req.AuctionID, amount)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, tx)
}

/*
@method Quote

@desc
Handles transaction quote requests.

@params
- c: Gin context
*/
func (h *TransactionHandler) Quote(c *gin.Context) {
	var req request.TransactionQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}
	quote, err := h.service.GetTransactionQuote(c.Request.Context(), req.Type, req.Params)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, quote)
}
