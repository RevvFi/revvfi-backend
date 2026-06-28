package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	"github.com/Revvfi/revvfi-backend/internal/core/liquidation"
	"github.com/Revvfi/revvfi-backend/internal/models"
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
Handles active auction listing requests.
Returns all active auctions across all markets.

@params
- c: Gin context
*/
func (h *LiquidationHandler) Liquidatable(c *gin.Context) {
	// Check for market filter
	marketAddress := c.Query("market_address")

	var auctions []models.Auction
	var err error

	if marketAddress != "" {
		// Get auctions for specific market
		auctions, err = h.service.GetAuctionsByMarket(c.Request.Context(), marketAddress)
	} else {
		// Get all active auctions
		auctions, err = h.service.GetActiveAuctions(c.Request.Context())
	}

	if err != nil {
		writeError(c, err)
		return
	}

	// Convert to response format with market asset info
	items := make([]response.AuctionResponse, 0, len(auctions))
	for i := range auctions {
		// Fetch market info to get asset details
		market, err := h.service.GetMarketByAddress(c.Request.Context(), auctions[i].MarketAddress)
		if err != nil {
			// If market not found, return auction without asset info
			items = append(items, auctionResponse(&auctions[i], nil))
		} else {
			items = append(items, auctionResponse(&auctions[i], market))
		}
	}

	ok(c, http.StatusOK, gin.H{
		"count":    len(items),
		"auctions": items,
	})
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

	// Fetch market info to include asset details
	market, err := h.service.GetMarketByAddress(c.Request.Context(), auction.MarketAddress)
	if err != nil {
		// Return auction without asset info if market not found
		ok(c, http.StatusOK, auctionResponse(auction, nil))
		return
	}

	ok(c, http.StatusOK, auctionResponse(auction, market))
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
