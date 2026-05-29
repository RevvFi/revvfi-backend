package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/core/borrower"
)

/*
@struct BorrowerHandler

@desc
HTTP handler for borrower endpoints.

@responsibilities
- Register borrower profiles
- Return reputation data
- Return risk assessment data
*/
type BorrowerHandler struct {
	service *borrower.BorrowerService
}

/*
@function NewBorrowerHandler

@desc
Creates a borrower HTTP handler.

@params
- service: borrower core service

@returns
- *BorrowerHandler
*/
func NewBorrowerHandler(service *borrower.BorrowerService) *BorrowerHandler {
	return &BorrowerHandler{service: service}
}

/*
@method Register

@desc
Handles authenticated borrower registration requests.

@params
- c: Gin context
*/
func (h *BorrowerHandler) Register(c *gin.Context) {
	created, err := h.service.RegisterBorrower(c.Request.Context(), c.GetString("wallet"))
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusCreated, borrowerResponse(created))
}

/*
@method Get

@desc
Handles borrower reputation requests.

@params
- c: Gin context
*/
func (h *BorrowerHandler) Get(c *gin.Context) {
	found, err := h.service.GetBorrower(c.Request.Context(), c.Param("address"))
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, borrowerResponse(found))
}

/*
@method Risk

@desc
Handles borrower risk assessment requests.

@params
- c: Gin context
*/
func (h *BorrowerHandler) Risk(c *gin.Context) {
	risk, err := h.service.GetRiskAssessment(c.Request.Context(), c.Param("address"))
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, risk)
}
