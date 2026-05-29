package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/request"
	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	"github.com/Revvfi/revvfi-backend/internal/core/offer"
)

/*
@struct OfferHandler

@desc
HTTP handler for offer endpoints.

@responsibilities
- Bind offer DTOs
- Call offer service methods
- Return offer and quote responses
*/
type OfferHandler struct {
	service *offer.OfferService
}

/*
@function NewOfferHandler

@desc
Creates an offer HTTP handler.

@params
- service: offer core service

@returns
- *OfferHandler
*/
func NewOfferHandler(service *offer.OfferService) *OfferHandler {
	return &OfferHandler{service: service}
}

/*
@method Create

@desc
Handles authenticated offer creation requests.

@params
- c: Gin context
*/
func (h *OfferHandler) Create(c *gin.Context) {
	var req request.CreateOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}

	amount, err := parseBigInt(req.Amount)
	if err != nil {
		writeError(c, err)
		return
	}

	created, err := h.service.CreateOffer(c.Request.Context(), time.Now().UnixNano(), c.GetString("wallet"), req.MarketAddress, amount, req.APR, req.Seniority, req.ExpiryDays)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusCreated, offerResponse(created))
}

/*
@method Get

@desc
Handles offer detail requests.

@params
- c: Gin context
*/
func (h *OfferHandler) Get(c *gin.Context) {
	offerID, err := pathInt64(c, "offerID")
	if err != nil {
		writeError(c, err)
		return
	}
	found, err := h.service.GetOffer(c.Request.Context(), offerID)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, offerResponse(found))
}

/*
@method List

@desc
Handles offer listing requests.

@params
- c: Gin context
*/
func (h *OfferHandler) List(c *gin.Context) {
	var req request.OfferListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		writeError(c, err)
		return
	}
	limit, offset := pagination(req.Page, req.PageSize)
	offers, err := h.service.GetMarketOffers(c.Request.Context(), req.MarketAddress, limit, offset)
	if err != nil {
		writeError(c, err)
		return
	}

	items := make([]response.OfferResponse, 0, len(offers))
	for i := range offers {
		items = append(items, offerResponse(&offers[i]))
	}
	ok(c, http.StatusOK, gin.H{"offers": items, "count": len(items)})
}

/*
@method Quote

@desc
Handles offer quote requests.

@params
- c: Gin context
*/
func (h *OfferHandler) Quote(c *gin.Context) {
	var req request.OfferQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}
	amount, err := parseBigInt(req.BorrowAmount)
	if err != nil {
		writeError(c, err)
		return
	}
	quote, err := h.service.CalculateQuote(c.Request.Context(), req.MarketAddress, amount, req.PreferredAPR)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, quote)
}

/*
@method Cancel

@desc
Handles authenticated offer cancellation requests.

@params
- c: Gin context
*/
func (h *OfferHandler) Cancel(c *gin.Context) {
	offerID, err := pathInt64(c, "offerID")
	if err != nil {
		writeError(c, err)
		return
	}
	if err := h.service.CancelOffer(c.Request.Context(), offerID, c.GetString("wallet")); err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, gin.H{"message": "offer cancelled"})
}
