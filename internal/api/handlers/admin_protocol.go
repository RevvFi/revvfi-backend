package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/request"
	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@interface AdminProtocolServiceInterface
@desc Complete interface for admin protocol configuration operations
*/
type AdminProtocolServiceInterface interface {
	GetProtocolConfig(ctx interface{}) (*response.ProtocolConfig, error)
	PrepareFeeChange(ctx interface{}, feeWei string) (*response.TransactionPreparation, error)
	PrepareFeeRecipientChange(ctx interface{}, recipientAddress, reason string) (*response.TransactionPreparation, error)
	GetFeesCollected(ctx interface{}) (*response.FeeCollectedResponse, error)
	CheckCoreContractsStatus(ctx interface{}) (*response.ContractStatus, error)
	PrepareCoreContractsSet(ctx interface{}, positionNFT, liquidator, reputationRegistry string) (*response.TransactionPreparation, error)
	ListPendingUpgrades(ctx interface{}, page, limit int) ([]response.PendingUpgrade, *response.PaginationInfo, error)
	PrepareMarketUpgrade(ctx interface{}, marketImpl, reason string) (*response.TransactionPreparation, error)
	GetTimelockStatus(ctx interface{}) (*response.TimelockStatus, error)
	WithdrawFees(ctx interface{}, amount string, recipient string) (*response.TransactionPreparation, error)
	CancelUpgrade(ctx interface{}, upgradeId string, reason string) (*response.TransactionPreparation, error)
	GetUpgradeQueue(ctx interface{}) ([]response.UpgradeQueueItem, error)
}

type AdminProtocolHandler struct {
	protocolService AdminProtocolServiceInterface
}

func NewAdminProtocolHandler(protocolService AdminProtocolServiceInterface) *AdminProtocolHandler {
	return &AdminProtocolHandler{
		protocolService: protocolService,
	}
}

// GET /api/v1/admin/protocol/config
func (h *AdminProtocolHandler) GetProtocolConfig(c *gin.Context) {
	cfg, err := h.protocolService.GetProtocolConfig(c)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    cfg,
	})
}

// POST /api/v1/admin/protocol/fee/prepare
func (h *AdminProtocolHandler) SetDeploymentFee(c *gin.Context) {
	var req request.SetDeploymentFeeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": fmt.Sprintf("Invalid request: %v", err),
			},
		})
		return
	}

	txPrep, err := h.protocolService.PrepareFeeChange(c, req.FeeWei)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    txPrep,
	})
}

// POST /api/v1/admin/protocol/fee-recipient/prepare
func (h *AdminProtocolHandler) SetFeeRecipient(c *gin.Context) {
	var req request.SetFeeRecipientRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": fmt.Sprintf("Invalid request: %v", err),
			},
		})
		return
	}

	txPrep, err := h.protocolService.PrepareFeeRecipientChange(c, req.RecipientAddress, req.Reason)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    txPrep,
	})
}

// GET /api/v1/admin/protocol/fees-collected
func (h *AdminProtocolHandler) GetFeesCollected(c *gin.Context) {
	fees, err := h.protocolService.GetFeesCollected(c)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    fees,
	})
}

// GET /api/v1/admin/contracts/status
func (h *AdminProtocolHandler) CheckCoreContractsStatus(c *gin.Context) {
	status, err := h.protocolService.CheckCoreContractsStatus(c)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// POST /api/v1/admin/contracts/set-core/prepare
func (h *AdminProtocolHandler) SetCoreContracts(c *gin.Context) {
	var req request.SetCoreContractsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": fmt.Sprintf("Invalid request: %v", err),
			},
		})
		return
	}

	txPrep, err := h.protocolService.PrepareCoreContractsSet(
		c,
		req.PositionNFT,
		req.Liquidator,
		req.ReputationRegistry,
	)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    txPrep,
	})
}

// GET /api/v1/admin/upgrades/pending
func (h *AdminProtocolHandler) ListPendingUpgrades(c *gin.Context) {
	page := 1
	limit := 20

	if p := c.Query("page"); p != "" {
		if _, err := fmt.Sscanf(p, "%d", &page); err != nil || page < 1 {
			page = 1
		}
	}
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil || limit < 1 {
			limit = 20
		} else if limit > 100 {
			limit = 100
		}
	}

	upgrades, pagination, err := h.protocolService.ListPendingUpgrades(c, page, limit)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": err.Error(),
			},
		})
		return
	}

	totalPending := len(upgrades)
	if pagination != nil {
		totalPending = pagination.TotalPages
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"upgrades":       upgrades,
			"pagination":     pagination,
			"total_pending":  totalPending,
		},
	})
}

// POST /api/v1/admin/upgrades/market/prepare
func (h *AdminProtocolHandler) PrepareMarketUpgrade(c *gin.Context) {
	var req request.SetMarketImplRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": fmt.Sprintf("Invalid request: %v", err),
			},
		})
		return
	}

	txPrep, err := h.protocolService.PrepareMarketUpgrade(c, req.MarketImplAddress, req.Reason)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    txPrep,
	})
}

// GET /api/v1/admin/upgrades/timelock-status
func (h *AdminProtocolHandler) GetTimelockStatus(c *gin.Context) {
	status, err := h.protocolService.GetTimelockStatus(c)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// POST /api/v1/admin/protocol/fees/withdraw/prepare
func (h *AdminProtocolHandler) WithdrawFees(c *gin.Context) {
	var req request.WithdrawFeesRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": fmt.Sprintf("Invalid request: %v", err),
			},
		})
		return
	}

	txPrep, err := h.protocolService.WithdrawFees(c, req.Amount, req.Recipient)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    txPrep,
	})
}

// DELETE /api/v1/admin/upgrades/:upgrade_id/prepare
func (h *AdminProtocolHandler) CancelUpgrade(c *gin.Context) {
	upgradeId := c.Param("upgrade_id")
	if upgradeId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "MISSING_UPGRADE_ID",
				"message": "Upgrade ID is required",
			},
		})
		return
	}

	var req request.CancelUpgradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": fmt.Sprintf("Invalid request: %v", err),
			},
		})
		return
	}

	txPrep, err := h.protocolService.CancelUpgrade(c, upgradeId, req.Reason)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    txPrep,
	})
}

// GET /api/v1/admin/upgrades/queue
func (h *AdminProtocolHandler) GetUpgradeQueue(c *gin.Context) {
	queue, err := h.protocolService.GetUpgradeQueue(c)
	if err != nil {
		statusCode := appErr.StatusCode(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.ErrorCode(err),
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    queue,
	})
}