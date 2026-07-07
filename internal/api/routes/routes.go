package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/api/handlers"
	"github.com/Revvfi/revvfi-backend/internal/api/middleware"
	"github.com/Revvfi/revvfi-backend/internal/config"
	admincore "github.com/Revvfi/revvfi-backend/internal/core/admin"
	"github.com/Revvfi/revvfi-backend/internal/core/auth"
)

/*
@struct Handlers

@desc
Collects HTTP handlers required by route registration.

@responsibilities
- Keep route setup independent from service construction
- Make endpoint ownership explicit
*/
type Handlers struct {
	Auth             *handlers.AuthHandler
	Admin            *handlers.AdminHandler
	AdminAuth        *handlers.AdminAuthHandler
	AdminBorrowers   *handlers.AdminBorrowerHandler
	AdminProtocol    *handlers.AdminProtocolHandler
	AdminMarketRisk  *handlers.AdminMarketRiskHandler
	AdminLiquidator  *handlers.AdminLiquidatorHandler
	AdminReputation  *handlers.AdminReputationHandler
	AdminAudit       *handlers.AdminAuditHandler
	AdminStats       *handlers.AdminStatsHandler
	AdminEmergency   *handlers.AdminEmergencyHandler
	AdminSystem      *handlers.AdminSystemHandler
	Market           *handlers.MarketHandler
	Offer            *handlers.OfferHandler
	Position         *handlers.PositionHandler
	Borrower         *handlers.BorrowerHandler
	BorrowerRequest  *handlers.BorrowerRequestHandler
	Liquidation      *handlers.LiquidationHandler
	Withdrawal       *handlers.WithdrawalHandler
	Transaction      *handlers.TransactionHandler
	Health           *handlers.HealthHandler
}


/*
@function Register

@desc
Registers RevvFi API routes on the provided Gin engine.

@responsibilities
- Attach global middleware
- Register public API endpoints
- Register authenticated API endpoints
- Keep route paths centralized

@params
- router: Gin engine
- cfg: runtime configuration
- authService: auth service for JWT middleware
- adminAuthService: admin auth service for admin authorization middleware
- h: handler collection
*/
func Register(router *gin.Engine, cfg *config.Config, authService *auth.AuthService, adminAuthService *admincore.AdminAuthService, h Handlers) {
	router.Use(middleware.Recovery())
	router.Use(middleware.Logging())
	router.Use(middleware.CORS(cfg.CORS.AllowedOrigins, cfg.CORS.AllowedMethods, cfg.CORS.AllowedHeaders))
	router.Use(middleware.RateLimit(cfg.RateLimit.Requests, cfg.RateLimit.Window))

	api := router.Group(cfg.Server.BasePath)
	{
		/*
		@routes Health
		@desc Expose liveness and readiness checks for uptime monitoring.
		*/
		api.GET("/health", h.Health.Health)
		api.GET("/ready", h.Health.Ready)

		/*
		@routes Authentication
		@desc Issue SIWE nonces and exchange signed wallet messages for API tokens.
		*/
		api.POST("/auth/nonce", h.Auth.GetNonce)
		api.POST("/auth/login", h.Auth.Login)

		/*
		@routes PublicMarkets
		@desc Let clients discover markets, inspect market details, and read market metrics.
		*/
		api.GET("/markets", h.Market.List)
		api.GET("/markets/:address", h.Market.Get)
		api.GET("/markets/:address/metrics", h.Market.Metrics)

		/*
		@routes PublicOffers
		@desc Let clients browse lending offers, fetch one offer, and request quote previews.
		*/
		api.GET("/offers", h.Offer.List)
		api.GET("/offers/:offerID", h.Offer.Get)
		api.POST("/offers/quote", h.Offer.Quote)

		/*
		@routes PublicBorrowers
		@desc Return borrower profile and risk information by wallet address.
		*/
		api.GET("/borrowers/:address", h.Borrower.Get)
		api.GET("/borrowers/:address/risk", h.Borrower.Risk)
		api.GET("/borrowers/:address/collateral", h.Borrower.Collateral)

		/*
		@routes PublicLiquidations
		@desc Surface liquidatable positions, auction details, and auction prices.
		*/
		api.GET("/liquidations", h.Liquidation.Liquidatable)
		api.GET("/liquidations/auctions/:auctionID", h.Liquidation.GetAuction)
		api.GET("/liquidations/auctions/:auctionID/price", h.Liquidation.Price)

		protected := api.Group("")
		protected.Use(middleware.Auth(authService))
		{
			/*
			@routes AuthenticatedAccount
			@desc Manage the current authenticated wallet session and profile lookup.
			*/
			protected.POST("/auth/logout", h.Auth.Logout)
			protected.GET("/auth/me", h.Auth.GetMe)

			/*
			@routes AuthenticatedMarkets
			@desc REMOVED - Market creation must go through blockchain transactions

			@architecture
			Markets are created by calling RevvFiFactory.deployMarket() on-chain.
			The contract emits MarketCreated event which the indexer processes.
			This API is READ-ONLY for markets.

			Frontend flow:
			1. User signs transaction with wallet (deployMarket)
			2. Smart contract emits MarketCreated event
			3. Indexer catches event and writes to database
			4. GET /markets returns the indexed market

			@see internal/handlers/market_handler.go for event processing
			*/
			// protected.POST("/markets", h.Market.Create) // ❌ REMOVED - Violates blockchain-first architecture


			/*
			@routes AdminProtected
			@desc All admin routes requiring both authentication and admin role verification.
			*/
			admin := protected.Group("")
			admin.Use(middleware.AdminAuth(adminAuthService))
			{
				/*
				@routes Admin
				@desc Administrative market management and health monitoring.

				@architecture
				Admin market creation removed - must use blockchain transactions.
				Admins should use the frontend to sign deployMarket() transactions.
				*/
				// admin.POST("/admin/markets", h.Admin.CreateMarket) // ❌ REMOVED - Violates blockchain-first architecture
				admin.GET("/admin/markets", h.Admin.ListMarkets)
				admin.GET("/admin/markets/:address", h.Admin.GetMarket)
				admin.GET("/admin/markets/:address/metrics", h.Admin.GetMarketMetrics)
				admin.PATCH("/admin/markets/:address/status", h.Admin.UpdateMarketStatus)
				admin.GET("/admin/health", h.Admin.HealthCheck)

				/*
				@routes AdminAuthentication
				@desc Admin authentication and authorization management endpoints.
				*/
				admin.GET("/admin/check/:address", h.AdminAuth.CheckAdmin)
				admin.POST("/admin/auth/impersonate", h.AdminAuth.Impersonate)
				admin.GET("/admin/admins", h.AdminAuth.ListAdmins)

				/*
				@routes AdminBorrowerManagement
				@desc Admin endpoints for borrower management and verification.
				*/
				admin.GET("/admin/borrowers", h.AdminBorrowers.ListBorrowers)
				admin.GET("/admin/borrowers/:address", h.AdminBorrowers.GetBorrower)
				admin.GET("/admin/borrowers/pending", h.AdminBorrowers.GetPendingBorrowers)
				admin.POST("/admin/borrowers/:address/prepare", h.AdminBorrowers.PrepareBorrowerAddition)
				admin.DELETE("/admin/borrowers/:address/prepare", h.AdminBorrowers.PrepareBorrowerRemoval)

				/*
				@routes AdminProtocolConfiguration
				@desc Complete protocol configuration management including fees, core contracts, and upgrades.
				*/
				admin.GET("/protocol/config", h.AdminProtocol.GetProtocolConfig)
				admin.POST("/protocol/fee/prepare", h.AdminProtocol.SetDeploymentFee)
				admin.POST("/protocol/fee-recipient/prepare", h.AdminProtocol.SetFeeRecipient)
				admin.GET("/protocol/fees-collected", h.AdminProtocol.GetFeesCollected)
				admin.POST("/protocol/fees/withdraw/prepare", h.AdminProtocol.WithdrawFees)
				admin.GET("/contracts/status", h.AdminProtocol.CheckCoreContractsStatus)
				admin.POST("/contracts/set-core/prepare", h.AdminProtocol.SetCoreContracts)
				admin.GET("/upgrades/pending", h.AdminProtocol.ListPendingUpgrades)
				admin.POST("/upgrades/market/prepare", h.AdminProtocol.PrepareMarketUpgrade)
				admin.GET("/upgrades/timelock-status", h.AdminProtocol.GetTimelockStatus)
				admin.DELETE("/upgrades/:upgrade_id/prepare", h.AdminProtocol.CancelUpgrade)
				admin.GET("/upgrades/queue", h.AdminProtocol.GetUpgradeQueue)

				/*
				@routes AdminMarketRiskParameters
				@desc Admin market risk parameter management.
				*/
				admin.GET("/admin/markets/:address/risk", h.AdminMarketRisk.GetMarketRiskParams)
				admin.POST("/admin/markets/:address/risk/min-cr/prepare", h.AdminMarketRisk.PrepareSetMinCR)
				admin.POST("/admin/markets/:address/risk/liquidation-threshold/prepare", h.AdminMarketRisk.PrepareSetLiqThreshold)
				admin.POST("/admin/markets/:address/risk/oracle/prepare", h.AdminMarketRisk.PrepareSetOracle)

				/*
				@routes AdminLiquidatorConfiguration
				@desc Admin liquidator configuration and auction management.
				*/
				admin.GET("/admin/liquidator/config", h.AdminLiquidator.GetLiquidatorConfig)
				admin.POST("/admin/liquidator/config/prepare", h.AdminLiquidator.PrepareSetLiquidatorParams)
				admin.GET("/admin/liquidator/auctions", h.AdminLiquidator.GetActiveAuctions)
				admin.POST("/admin/liquidator/auctions/:auctionID/stop/prepare", h.AdminLiquidator.PrepareStopAuction)

				/*
				@routes AdminReputationManagement
				@desc Admin borrower reputation management.
				*/
				admin.GET("/admin/reputation/defaulted", h.AdminReputation.GetDefaultedBorrowers)
				admin.GET("/admin/reputation/:address", h.AdminReputation.GetBorrowerReputation)
				admin.POST("/admin/reputation/:address/prepare", h.AdminReputation.PrepareSetReputation)

				/*
				@routes AdminAuditMonitoring
				@desc Admin audit log querying and compliance monitoring.
				*/
				admin.GET("/admin/audit/logs", h.AdminAudit.GetAuditLogs)
				admin.GET("/admin/audit/stats", h.AdminAudit.GetAuditStats)
				admin.GET("/admin/audit/export", h.AdminAudit.ExportAuditLogs)
				admin.GET("/admin/audit/activity/:adminAddress", h.AdminAudit.GetAdminActivity)
				admin.GET("/admin/audit/actions/:action", h.AdminAudit.GetActionHistory)

				/*
				@routes AdminDashboardStatistics
				@desc Admin dashboard statistics and analytics.
				*/
				admin.GET("/admin/stats/overview", h.AdminStats.GetOverview)
				admin.GET("/admin/stats/borrowers", h.AdminStats.GetBorrowerStats)
				admin.GET("/admin/stats/markets", h.AdminStats.GetMarketStats)
				admin.GET("/admin/stats/revenue", h.AdminStats.GetRevenueStats)
				admin.GET("/admin/stats/liquidations", h.AdminStats.GetLiquidationStats)
				admin.GET("/admin/stats/positions", h.AdminStats.GetPositionStats)

				/*
				@routes AdminEmergencyControls
				@desc Emergency protocol controls (pause/unpause/drain fees).
				*/
				admin.POST("/admin/emergency/pause/prepare", h.AdminEmergency.PrepareEmergencyPause)
				admin.POST("/admin/emergency/unpause/prepare", h.AdminEmergency.PrepareEmergencyUnpause)
				admin.POST("/admin/emergency/drain-fees/prepare", h.AdminEmergency.PrepareDrainFees)

				/*
				@routes AdminSystemConfiguration
				@desc Admin system configuration management.
				*/
				admin.GET("/admin/system/config", h.AdminSystem.GetSystemConfig)
				admin.POST("/admin/system/config/prepare", h.AdminSystem.PrepareUpdateSystemConfig)

				/*
				@routes AdminBorrowerRequests
				@desc Admin review queue for off-chain borrower access requests.

				@architecture
				registerBorrower() is onlyOwner on-chain, so an arbitrary wallet can
				never self-register. Approval still happens on-chain (admin sends
				registerBorrower() from their own connected wallet; the indexer
				auto-resolves the matching request to "approved" when it observes
				the BorrowerAdded event). Rejection has no on-chain effect, so it is
				a plain backend endpoint.
				*/
				admin.GET("/admin/borrower-requests", h.BorrowerRequest.List)
				admin.PATCH("/admin/borrower-requests/:id/reject", h.BorrowerRequest.Reject)
			}

			/*
			@routes AuthenticatedOffers
			@desc REMOVED - Offer creation must go through blockchain transactions

			@architecture
			Offers are created by calling RevvFiOfferBook.submitOffer() on-chain.
			The contract emits OfferSubmitted event which the indexer processes.
			Cancellation also happens on-chain via cancelOffer().
			This API is READ-ONLY for offers.

			@see internal/handlers/offer_handler.go for event processing
			*/
			// protected.POST("/offers", h.Offer.Create) // ❌ REMOVED - Violates blockchain-first architecture
			// protected.DELETE("/offers/:offerID", h.Offer.Cancel) // ❌ REMOVED - Must use blockchain tx

			/*
			@routes AuthenticatedBorrowers
			@desc Borrower access request queue (self-registration is still
			blocked on-chain; see AdminBorrowerRequests below for the actual
			admin approval flow).

			@architecture
			Only admins can register borrowers via ArchController.registerBorrower()
			(onlyOwner on-chain). This route lets any signed-in wallet submit a
			request for the admin to review, instead of a direct (and impossible)
			self-registration.

			@see internal/handlers/arch_controller_handler.go for event processing
			@see app/(app)/admin/page.tsx for admin UI implementation
			*/
			// protected.POST("/borrowers/register", h.Borrower.Register) // ❌ REMOVED - Admin-only on-chain action
			protected.POST("/borrower-requests", h.BorrowerRequest.Create)
			protected.GET("/borrower-requests/me", h.BorrowerRequest.GetMine)
			protected.GET("/positions", h.Position.ListMine)
			protected.GET("/positions/portfolio", h.Position.Portfolio)
			protected.GET("/positions/:tokenID", h.Position.Get)
			protected.POST("/positions/claim", h.Position.Claim)

			/*
			@routes Withdrawals
			@desc List, request, cancel, and inspect epoch-based withdrawal flows.
			*/
			protected.GET("/withdrawals", h.Withdrawal.ListMine)
			protected.POST("/withdrawals", h.Withdrawal.Request)
			protected.POST("/withdrawals/cancel", h.Withdrawal.Cancel)
			protected.GET("/withdrawals/current-epoch", h.Withdrawal.CurrentEpoch)

			/*
			@routes Transactions
			@desc Build unsigned protocol transactions and quote transaction outcomes.
			*/
			protected.POST("/transactions/borrow", h.Transaction.BuildBorrow)
			protected.POST("/transactions/repay", h.Transaction.BuildRepay)
			protected.POST("/transactions/liquidate", h.Transaction.BuildLiquidate)
			protected.POST("/transactions/quote", h.Transaction.Quote)
		}
	}
}
