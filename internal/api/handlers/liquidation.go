package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	"github.com/Revvfi/revvfi-backend/internal/core/liquidation"
)

/*
@struct LiquidationHandler

@desc
HTTP handler for liquidation endpoints.

@responsibilities
- Return liquidatable markets
- Return auction details and prices
- Accept liquidation bids
*/
type LiquidationHandler struct {
	service *liquidation.LiquidationService
}

/*
@function NewLiquidationHandler

@desc
Creates a liquidation HTTP handler.

@params
- service: liquidation core service

@returns
- *LiquidationHandler
*/
func NewLiquidationHandler(service *liquidation.LiquidationService) *LiquidationHandler {
	return &LiquidationHandler{service: service}
}

/*
@method Liquidatable

@desc
Handles liquidatable market listing requests.

@params
- c: Gin context
*/
func (h *LiquidationHandler) Liquidatable(c *gin.Context) {
	markets, err := h.service.GetLiquidatableMarkets(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	items := make([]response.LiquidatableMarket, 0, len(markets))
	for i := range markets {
		items = append(items, response.LiquidatableMarket{
			MarketAddress: markets[i].Address,
			Borrower:      markets[i].Borrower,
			DebtAmount:    stringOrZero(markets[i].TotalDebt),
		})
	}
	ok(c, http.StatusOK, response.LiquidatableResponse{Markets: items, Count: int32(len(items)), TotalCollateral: "0", TotalDebt: "0"})
}

/*
@method GetAuction

@desc
Handles auction detail requests.

@params
- c: Gin context
*/
func (h *LiquidationHandler) GetAuction(c *gin.Context) {
	auction, err := h.service.GetAuction(c.Request.Context(), c.Param("auctionID"))
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, auctionResponse(auction))
}

/*
@method Price

@desc
Handles current Dutch auction price requests.

@params
- c: Gin context
*/
func (h *LiquidationHandler) Price(c *gin.Context) {
	price, err := h.service.GetCurrentPrice(c.Request.Context(), c.Param("auctionID"))
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, gin.H{"auction_id": c.Param("auctionID"), "current_price": price.String()})
}
