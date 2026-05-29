package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/request"
	"github.com/Revvfi/revvfi-backend/internal/core/market"
)

/*
@struct MarketHandler

@desc
HTTP handler for market endpoints.

@responsibilities
- Bind market request DTOs
- Call market service methods
- Map domain markets to response DTOs
*/
type MarketHandler struct {
	service *market.MarketService
}

/*
@function NewMarketHandler

@desc
Creates a market HTTP handler.

@params
- service: market core service

@returns
- *MarketHandler
*/
func NewMarketHandler(service *market.MarketService) *MarketHandler {
	return &MarketHandler{service: service}
}

/*
@method Create

@desc
Handles market creation requests.

@params
- c: Gin context
*/
func (h *MarketHandler) Create(c *gin.Context) {
	var req request.CreateMarketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}

	wallet := c.GetString("wallet")
	created, err := h.service.CreateMarket(
		c.Request.Context(),
		"",
		wallet,
		req.BorrowAsset,
		req.CollateralAsset,
		req.CollateralOracle,
		req.MinCollateralRatio,
		req.LiquidationThreshold,
	)
	if err != nil {
		writeError(c, err)
		return
	}

	ok(c, http.StatusCreated, marketResponse(created))
}

/*
@method Get

@desc
Handles market detail requests.

@params
- c: Gin context
*/
func (h *MarketHandler) Get(c *gin.Context) {
	found, err := h.service.GetMarket(c.Request.Context(), c.Param("address"))
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, marketResponse(found))
}

/*
@method List

@desc
Handles paginated market listing requests.

@params
- c: Gin context
*/
func (h *MarketHandler) List(c *gin.Context) {
	var req request.MarketListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		writeError(c, err)
		return
	}

	limit, offset := pagination(req.Page, req.PageSize)
	markets, err := h.service.ListMarkets(c.Request.Context(), limit, offset)
	if err != nil {
		writeError(c, err)
		return
	}

	items := make([]interface{}, 0, len(markets))
	for i := range markets {
		items = append(items, marketResponse(&markets[i]))
	}
	ok(c, http.StatusOK, gin.H{"markets": items, "count": len(items)})
}

/*
@method Metrics

@desc
Handles market metric calculation requests.

@params
- c: Gin context
*/
func (h *MarketHandler) Metrics(c *gin.Context) {
	found, err := h.service.GetMarket(c.Request.Context(), c.Param("address"))
	if err != nil {
		writeError(c, err)
		return
	}
	metrics, err := h.service.CalculateMetrics(c.Request.Context(), found)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, metrics)
}
