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
		api.GET("/health", h.Health.Health)
		api.GET("/ready", h.Health.Ready)

		api.POST("/auth/nonce", h.Auth.GetNonce)
		api.POST("/auth/login", h.Auth.Login)

		api.GET("/markets", h.Market.List)
		api.GET("/markets/:address", h.Market.Get)
		api.GET("/markets/:address/metrics", h.Market.Metrics)

		api.GET("/offers", h.Offer.List)
		api.GET("/offers/:offerID", h.Offer.Get)
		api.POST("/offers/quote", h.Offer.Quote)

		api.GET("/borrowers/:address", h.Borrower.Get)
		api.GET("/borrowers/:address/risk", h.Borrower.Risk)

		api.GET("/liquidations", h.Liquidation.Liquidatable)
		api.GET("/liquidations/auctions/:auctionID", h.Liquidation.GetAuction)
		api.GET("/liquidations/auctions/:auctionID/price", h.Liquidation.Price)

		protected := api.Group("")
		protected.Use(middleware.Auth(authService))
		{
			protected.POST("/auth/logout", h.Auth.Logout)
			protected.GET("/auth/me", h.Auth.GetMe)

			protected.POST("/markets", h.Market.Create)
			protected.POST("/offers", h.Offer.Create)
			protected.DELETE("/offers/:offerID", h.Offer.Cancel)

			protected.POST("/borrowers/register", h.Borrower.Register)
			protected.GET("/positions", h.Position.ListMine)
			protected.GET("/positions/portfolio", h.Position.Portfolio)
			protected.GET("/positions/:tokenID", h.Position.Get)
			protected.POST("/positions/claim", h.Position.Claim)

			protected.GET("/withdrawals", h.Withdrawal.ListMine)
			protected.POST("/withdrawals", h.Withdrawal.Request)
			protected.POST("/withdrawals/cancel", h.Withdrawal.Cancel)
			protected.GET("/withdrawals/current-epoch", h.Withdrawal.CurrentEpoch)

			protected.POST("/transactions/borrow", h.Transaction.BuildBorrow)
			protected.POST("/transactions/repay", h.Transaction.BuildRepay)
			protected.POST("/transactions/liquidate", h.Transaction.BuildLiquidate)
			protected.POST("/transactions/quote", h.Transaction.Quote)
		}
	}
}
