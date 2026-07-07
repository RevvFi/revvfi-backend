package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/request"
	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	"github.com/Revvfi/revvfi-backend/internal/core/offer"
	"github.com/Revvfi/revvfi-backend/internal/logger"
	"github.com/Revvfi/revvfi-backend/internal/models"
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
	ctx := c.Request.Context()
	var req request.CreateOfferRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WarnContext(ctx, "Invalid offer creation request",
			logger.WithError(err),
		)
		writeError(c, err)
		return
	}

	amount, err := parseBigInt(req.Amount)
	if err != nil {
		logger.WarnContext(ctx, "Invalid offer amount",
			logger.WithError(err),
			"amount", req.Amount,
		)
		writeError(c, err)
		return
	}

	wallet := c.GetString("wallet")

	logger.InfoContext(ctx, "Creating offer",
		logger.WithWalletAddress(wallet),
		logger.WithMarketAddress(req.MarketAddress),
		logger.WithAmount(amount.String()),
		logger.WithAPR(int(req.APR)),
		logger.WithSeniority(int(req.Seniority)),
	)

	created, err := h.service.CreateOffer(ctx, time.Now().UnixNano(), wallet, req.MarketAddress, amount, req.APR, req.Seniority, req.ExpiryDays)
	if err != nil {
		logger.ErrorContext(ctx, "Offer creation failed",
			logger.WithError(err),
			logger.WithWalletAddress(wallet),
			logger.WithMarketAddress(req.MarketAddress),
		)
		writeError(c, err)
		return
	}

	logger.InfoContext(ctx, "Offer created successfully",
		logger.WithOfferID(created.OfferID),
		logger.WithWalletAddress(wallet),
		logger.WithMarketAddress(req.MarketAddress),
	)

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

	var offers []models.Offer
	var err error

	switch {
	case req.Lender != "":
		// "My Offers" - every offer (any market, any status) placed by this
		// lender. Was previously unreachable: this query param didn't exist
		// on the DTO at all, and List() always called GetMarketOffers()
		// regardless, so a lender's own offers never showed up anywhere
		// that queried by lender instead of by market.
		page := req.Page
		if page < 1 {
			page = 1
		}
		pageSize := req.PageSize
		if pageSize < 1 {
			pageSize = 20
		}
		offers, err = h.service.GetLenderOffers(c.Request.Context(), req.Lender, page, pageSize)
	case req.MarketAddress != "":
		limit, offset := pagination(req.Page, req.PageSize)
		offers, err = h.service.GetMarketOffers(c.Request.Context(), req.MarketAddress, limit, offset)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "market_address or lender is required",
			},
		})
		return
	}
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
