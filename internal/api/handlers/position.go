package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/request"
	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	"github.com/Revvfi/revvfi-backend/internal/core/position"
)

/*
@struct PositionHandler

@desc
HTTP handler for position endpoints.

@responsibilities
- Bind position queries and path parameters
- Call position service methods
- Return position and portfolio responses
*/
type PositionHandler struct {
	service *position.PositionService
}

/*
@function NewPositionHandler

@desc
Creates a position HTTP handler.

@params
- service: position core service

@returns
- *PositionHandler
*/
func NewPositionHandler(service *position.PositionService) *PositionHandler {
	return &PositionHandler{service: service}
}

/*
@method Get

@desc
Handles position detail requests.

@params
- c: Gin context
*/
func (h *PositionHandler) Get(c *gin.Context) {
	tokenID, err := pathInt64(c, "tokenID")
	if err != nil {
		writeError(c, err)
		return
	}
	found, err := h.service.GetPosition(c.Request.Context(), tokenID)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, positionResponse(found))
}

/*
@method ListMine

@desc
Handles authenticated lender position listing requests.

@params
- c: Gin context
*/
func (h *PositionHandler) ListMine(c *gin.Context) {
	var req request.PositionListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		writeError(c, err)
		return
	}
	limit, offset := pagination(req.Page, req.PageSize)
	positions, err := h.service.GetLenderPositions(c.Request.Context(), c.GetString("wallet"), limit, offset)
	if err != nil {
		writeError(c, err)
		return
	}
	items := make([]response.PositionResponse, 0, len(positions))
	for i := range positions {
		items = append(items, positionResponse(&positions[i]))
	}
	ok(c, http.StatusOK, gin.H{"positions": items, "count": len(items)})
}

/*
@method Portfolio

@desc
Handles authenticated portfolio summary requests.

@params
- c: Gin context
*/
func (h *PositionHandler) Portfolio(c *gin.Context) {
	summary, err := h.service.GetPortfolioSummary(c.Request.Context(), c.GetString("wallet"))
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, summary)
}

/*
@method Claim

@desc
Handles claim simulation requests for settled positions.

@params
- c: Gin context
*/
func (h *PositionHandler) Claim(c *gin.Context) {
	var req request.ClaimPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}
	if err := h.service.SettlePosition(c.Request.Context(), req.TokenID); err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, gin.H{"message": "position marked claimable"})
}
