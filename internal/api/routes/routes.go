package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Revvfi/revvfi-backend/internal/api/handlers"
	"github.com/Revvfi/revvfi-backend/internal/api/middleware"
	"github.com/Revvfi/revvfi-backend/internal/config"
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
	Auth        *handlers.AuthHandler
	Admin       *handlers.AdminHandler
	AdminAuth   *handlers.AdminAuthHandler
	AdminBorrowers *handlers.AdminBorrowerHandler
	Market      *handlers.MarketHandler
	Offer       *handlers.OfferHandler
	Position    *handlers.PositionHandler
	Borrower    *handlers.BorrowerHandler
	Liquidation *handlers.LiquidationHandler
	Withdrawal  *handlers.WithdrawalHandler
	Transaction *handlers.TransactionHandler
	Health      *handlers.HealthHandler
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
- h: handler collection
*/
func Register(router *gin.Engine, cfg *config.Config, authService *auth.AuthService, h Handlers) {
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
			@desc Allow authenticated users to create public markets.
			*/
			protected.POST("/markets", h.Market.Create)

			/*
			@routes Admin
			@desc
			Provide authenticated administrative market management and health monitoring.
			Admin endpoints allow privileged users to create markets, view market metrics,
			update market operational status, and check system health.

			@endpoints
			- POST /admin/markets: Create new isolated lending market with custom parameters
			- GET /admin/markets: List all active markets with pagination and filtering
			- GET /admin/markets/:address: Retrieve detailed market configuration and state
			- GET /admin/markets/:address/metrics: Get market analytics and performance metrics
			- PATCH /admin/markets/:address/status: Update market operational status (active/paused/closed)
			- GET /admin/health: System health check with TVL, active positions, and uptime metrics
			*/
			protected.POST("/admin/markets", h.Admin.CreateMarket)
			protected.GET("/admin/markets", h.Admin.ListMarkets)
			protected.GET("/admin/markets/:address", h.Admin.GetMarket)
			protected.GET("/admin/markets/:address/metrics", h.Admin.GetMarketMetrics)
			protected.PATCH("/admin/markets/:address/status", h.Admin.UpdateMarketStatus)
			protected.GET("/admin/health", h.Admin.HealthCheck)

			/*
			@routes AdminAuthentication
			@desc
			Admin authentication and authorization management endpoints.
			Handles admin role verification, permission checking, and admin impersonation for testing.

			@endpoints
			- GET /admin/check/:address: Verify if address is admin and retrieve permissions
			- POST /admin/auth/impersonate: Generate temporary token for admin context switching (dev only)
			- GET /admin/admins: List all admin addresses with roles and permissions
			*/
			protected.GET("/admin/check/:address", h.AdminAuth.CheckAdmin)
			protected.POST("/admin/auth/impersonate", h.AdminAuth.Impersonate)
			protected.GET("/admin/admins", h.AdminAuth.ListAdmins)

			/*
			@routes AdminBorrowerManagement
			@desc
			Admin endpoints for borrower management and verification.
			Allows admins to list borrowers, verify/block borrowers, and manage borrower status.

			@endpoints
			- GET /admin/borrowers: List all borrowers with pagination and filtering
			- GET /admin/borrowers/:address: Retrieve detailed borrower information and metrics
			- GET /admin/borrowers/pending: List all borrowers awaiting verification
			- POST /admin/borrowers/:address/prepare: Prepare transaction to add borrower
			- DELETE /admin/borrowers/:address/prepare: Prepare transaction to remove borrower
			*/
			protected.GET("/admin/borrowers", h.AdminBorrowers.ListBorrowers)
			protected.GET("/admin/borrowers/:address", h.AdminBorrowers.GetBorrower)
			protected.GET("/admin/borrowers/pending", h.AdminBorrowers.GetPendingBorrowers)
			protected.POST("/admin/borrowers/:address/prepare", h.AdminBorrowers.PrepareBorrowerAddition)
			protected.DELETE("/admin/borrowers/:address/prepare", h.AdminBorrowers.PrepareBorrowerRemoval)

			/*
			@routes AuthenticatedOffers
			@desc Let authenticated lenders create and cancel offers.
			*/
			protected.POST("/offers", h.Offer.Create)
			protected.DELETE("/offers/:offerID", h.Offer.Cancel)

			/*
			@routes AuthenticatedBorrowers
			@desc Register borrower wallets and read the authenticated wallet's positions.
			*/
			protected.POST("/borrowers/register", h.Borrower.Register)
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
