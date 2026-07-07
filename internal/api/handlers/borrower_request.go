package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/request"
	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	"github.com/Revvfi/revvfi-backend/internal/core/borrower"
	"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@file borrower_request.go

@desc
HTTP handlers for the off-chain borrower access request queue.

@handler_set
- Create: POST /api/v1/borrower-requests - authenticated wallet requests borrower access
- GetMine: GET /api/v1/borrower-requests/me - authenticated wallet checks its own request status
- List: GET /api/v1/admin/borrower-requests - admin lists requests (optionally filtered by status)
- Reject: PATCH /api/v1/admin/borrower-requests/:id/reject - admin rejects a pending request
*/
type BorrowerRequestHandler struct {
	service *borrower.BorrowerRequestService
}

func NewBorrowerRequestHandler(service *borrower.BorrowerRequestService) *BorrowerRequestHandler {
	return &BorrowerRequestHandler{service: service}
}

func borrowerRequestResponse(r *models.BorrowerRequest) response.BorrowerRequestInfo {
	info := response.BorrowerRequestInfo{
		ID:            r.ID,
		WalletAddress: r.WalletAddress,
		Status:        r.Status,
		RequestedAt:   r.RequestedAt.Unix(),
	}
	if r.Note.Valid {
		info.Note = r.Note.String
	}
	if r.DecidedAt.Valid {
		info.DecidedAt = r.DecidedAt.Time.Unix()
	}
	if r.DecidedBy.Valid {
		info.DecidedBy = r.DecidedBy.String
	}
	return info
}

/*
@method Create

@desc
Submits a borrower access request for the authenticated wallet. The wallet
is taken from the JWT session (set by the Auth middleware), never from the
request body, so a signed-in wallet can only ever request access for itself.
*/
func (h *BorrowerRequestHandler) Create(c *gin.Context) {
	created, err := h.service.Create(c.Request.Context(), c.GetString("wallet"))
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusCreated, borrowerRequestResponse(created))
}

/*
@method GetMine

@desc
Returns the authenticated wallet's latest borrower request, or null if it
has never submitted one.
*/
func (h *BorrowerRequestHandler) GetMine(c *gin.Context) {
	found, err := h.service.GetMine(c.Request.Context(), c.GetString("wallet"))
	if err != nil {
		writeError(c, err)
		return
	}
	if found == nil {
		ok(c, http.StatusOK, nil)
		return
	}
	ok(c, http.StatusOK, borrowerRequestResponse(found))
}

/*
@method List

@desc
Admin-only listing of borrower requests, optionally filtered by
?status=pending|approved|rejected.
*/
func (h *BorrowerRequestHandler) List(c *gin.Context) {
	status := c.Query("status")
	requests, err := h.service.List(c.Request.Context(), status)
	if err != nil {
		writeError(c, err)
		return
	}
	infos := make([]response.BorrowerRequestInfo, 0, len(requests))
	for _, r := range requests {
		infos = append(infos, borrowerRequestResponse(r))
	}
	ok(c, http.StatusOK, response.BorrowerRequestListResponse{Count: len(infos), Requests: infos})
}

/*
@method Reject

@desc
Admin-only rejection of a pending borrower request. Pure off-chain
decision — no on-chain transaction involved.
*/
func (h *BorrowerRequestHandler) Reject(c *gin.Context) {
	id, err := pathInt64(c, "id")
	if err != nil {
		writeError(c, err)
		return
	}
	var body request.RejectBorrowerRequest
	if err := c.ShouldBindJSON(&body); err != nil && err.Error() != "EOF" {
		writeError(c, err)
		return
	}
	if err := h.service.Reject(c.Request.Context(), id, c.GetString("wallet"), body.Note); err != nil {
		writeError(c, err)
		return
	}
	ok(c, http.StatusOK, gin.H{"message": "borrower request rejected"})
}
