package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	apihandlers "github.com/Revvfi/revvfi-backend/internal/api/handlers"
	"github.com/Revvfi/revvfi-backend/internal/api/routes"
	"github.com/Revvfi/revvfi-backend/internal/config"
	admincore "github.com/Revvfi/revvfi-backend/internal/core/admin"
	"github.com/Revvfi/revvfi-backend/internal/core/auth"
	"github.com/Revvfi/revvfi-backend/internal/core/borrower"
	"github.com/Revvfi/revvfi-backend/internal/core/liquidation"
	"github.com/Revvfi/revvfi-backend/internal/core/market"
	"github.com/Revvfi/revvfi-backend/internal/core/offer"
	"github.com/Revvfi/revvfi-backend/internal/core/position"
	"github.com/Revvfi/revvfi-backend/internal/core/transaction"
	"github.com/Revvfi/revvfi-backend/internal/core/withdrawal"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
	"github.com/Revvfi/revvfi-backend/internal/logger"
	"github.com/Revvfi/revvfi-backend/middleware"

)

/*
@function main

@desc
Starts the RevvFi HTTP API server.

@responsibilities
- Load runtime configuration
- Open PostgreSQL connection
- Construct repositories, services, handlers, and routes
- Run HTTP server with graceful shutdown
*/
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Initialize structured logger FIRST
	env := logger.EnvDevelopment
	if cfg.Environment == "production" {
		env = logger.EnvProduction
	} else if cfg.Environment == "staging" {
		env = logger.EnvStaging
	}
	logger.Init(logger.ServiceBackend, env)
	logger.Info("RevvFi API server starting",
		"environment", env,
		"version", "1.0.0",
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := postgres.Open(ctx, cfg.Database)
	if err != nil {
		logger.Error("Failed to open database",
			logger.WithError(err),
		)
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()
	logger.Info("Database connection established")

	jwtMgr, err := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.TTL, cfg.JWT.Issuer)
	if err != nil {
		log.Fatalf("create jwt manager: %v", err)
	}

	authRepo := postgres.NewAuthRepository(db)
	marketRepo := postgres.NewMarketRepository(db)
	offerRepo := postgres.NewOfferRepository(db)
	positionRepo := postgres.NewPositionRepository(db)
	borrowerRepo := postgres.NewBorrowerRepository(db)
	auctionRepo := postgres.NewAuctionRepository(db)
	withdrawalRepo := postgres.NewWithdrawalRepository(db)

	/*
	@construction AuthService
	@desc Initialize authentication service with JWT manager for SIWE token validation.
	@fix Pass AuthConfig as third parameter
	*/
	authConfig := &auth.AuthConfig{
		Domain:     cfg.Auth.Domain,
		URI:        cfg.Auth.URI,
		Statement:  cfg.Auth.Statement,
		Version:    cfg.Auth.Version,
		ChainID:    cfg.Auth.ChainID,
		NonceTTL:   cfg.Auth.NonceTTL,
		SessionTTL: cfg.Auth.SessionTTL,
	}
	
	authService := auth.NewAuthService(authRepo, jwtMgr, authConfig)


	/*
	@construction MarketService
	@desc Initialize market service with market and offer repositories for market operations.
	*/
	marketService := market.NewMarketService(marketRepo, postgres.NewMarketOfferRepository(offerRepo), postgres.NewMarketPositionRepository(positionRepo))

	/*
	@construction OfferService
	@desc Initialize offer service for lending offer management.
	*/
	offerService := offer.NewOfferService(offerRepo)

	/*
	@construction PositionService
	@desc Initialize position service for lender position tracking.
	*/
	positionService := position.NewPositionService(positionRepo)

	/*
	@construction BorrowerService
	@desc Initialize borrower service for borrower risk assessment and reputation.
	*/
	borrowerService := borrower.NewBorrowerService(borrowerRepo)

	/*
	@construction BorrowerRequestService
	@desc Initialize the off-chain borrower access request queue service.
	*/
	borrowerRequestRepo := postgres.NewBorrowerRequestRepository(db)
	borrowerRequestService := borrower.NewBorrowerRequestService(borrowerRequestRepo, borrowerRepo)

	/*
	@construction LiquidationService
	@desc Initialize liquidation service for auction management and Dutch auctions.
	*/
	liquidationService := liquidation.NewLiquidationService(auctionRepo, marketRepo, borrowerRepo)

	/*
	@construction WithdrawalService
	@desc Initialize withdrawal service for epoch-based withdrawal processing.
	*/
	withdrawalService := withdrawal.NewWithdrawalService(withdrawalRepo, positionRepo)

	/*
	@construction TransactionService
	@desc Initialize transaction service for building and estimating protocol transactions.
	*/
	transactionService := transaction.NewTransactionService()

	/*
	@construction AdminServices
	@desc Initialize admin services and their shared repository.
	*/
	adminRepo := postgres.NewAdminRepository(db)

	adminAuthService := admincore.NewAdminAuthService(cfg.Blockchain, jwtMgr)
	adminBorrowerService := admincore.NewAdminBorrowerService(adminRepo, cfg.Blockchain)
	adminProtocolService := admincore.NewAdminProtocolService(cfg.Blockchain)
	adminMarketRiskService := admincore.NewAdminMarketRiskService(adminRepo, cfg.Blockchain)
	adminLiquidatorService := admincore.NewAdminLiquidatorService(adminRepo, cfg.Blockchain)
	adminReputationService := admincore.NewAdminReputationService(adminRepo, cfg.Blockchain)
	adminAuditService := admincore.NewAdminAuditService(adminRepo)
	adminStatsService := admincore.NewAdminStatsService(adminRepo)
	adminEmergencyService := admincore.NewAdminEmergencyService(adminRepo, cfg.Blockchain)
	adminSystemService := admincore.NewAdminSystemService(adminRepo, cfg.Blockchain)

	/*
	@construction HTTPHandlers
	@desc Initialize all HTTP request handlers with service dependencies.
	Includes auth, market, admin, offer, position, borrower, liquidation, withdrawal, and transaction handlers.
	*/
	router := gin.New()

	// Add structured logging middleware
	router.Use(middleware.CorrelationID())
	router.Use(middleware.RequestLogger())
	router.Use(gin.Recovery()) // Panic recovery

	logger.Info("HTTP middleware configured",
		"middleware", []string{"CorrelationID", "RequestLogger", "Recovery"},
	)

	routes.Register(router, cfg, authService, adminAuthService, routes.Handlers{
		/*
		@handler Auth
		@desc Handles SIWE authentication, nonce generation, login, and logout operations.
		*/
		Auth: apihandlers.NewAuthHandler(authService),

		/*
		@handler Admin
		@desc Handles administrative market operations including creation, status updates, and metrics.
		*/
		Admin: apihandlers.NewAdminHandler(marketService, market.NewValidator()),

		/*
		@handler AdminAuth
		@desc Handles admin role verification and impersonation token generation.
		*/
		AdminAuth: apihandlers.NewAdminAuthHandler(adminAuthService),

		/*
		@handler AdminBorrowers
		@desc Handles admin borrower management and verification operations.
		*/
		AdminBorrowers: apihandlers.NewAdminBorrowerHandler(adminBorrowerService),

		/*
		@handler AdminProtocol
		@desc Handles admin protocol configuration, fees, and upgrade management.
		*/
		AdminProtocol: apihandlers.NewAdminProtocolHandler(adminProtocolService),

		/*
		@handler AdminMarketRisk
		@desc Handles admin market risk parameter management.
		*/
		AdminMarketRisk: apihandlers.NewAdminMarketRiskHandler(adminMarketRiskService),

		/*
		@handler AdminLiquidator
		@desc Handles admin liquidator configuration and active auction management.
		*/
		AdminLiquidator: apihandlers.NewAdminLiquidatorHandler(adminLiquidatorService),

		/*
		@handler AdminReputation
		@desc Handles admin borrower reputation management.
		*/
		AdminReputation: apihandlers.NewAdminReputationHandler(adminReputationService),

		/*
		@handler AdminAudit
		@desc Handles admin audit log querying and compliance monitoring.
		*/
		AdminAudit: apihandlers.NewAdminAuditHandler(adminAuditService),

		/*
		@handler AdminStats
		@desc Handles admin dashboard statistics and analytics.
		*/
		AdminStats: apihandlers.NewAdminStatsHandler(adminStatsService),

		/*
		@handler AdminEmergency
		@desc Handles emergency protocol controls (pause/unpause/drain fees).
		*/
		AdminEmergency: apihandlers.NewAdminEmergencyHandler(adminEmergencyService),

		/*
		@handler AdminSystem
		@desc Handles admin system configuration management.
		*/
		AdminSystem: apihandlers.NewAdminSystemHandler(adminSystemService),

		/*
		@handler Market
		@desc Handles public market queries, creation, and metric retrieval.
		*/
		Market: apihandlers.NewMarketHandler(marketService),

		/*
		@handler Offer
		@desc Handles lending offer creation, cancellation, and quote calculations.
		*/
		Offer: apihandlers.NewOfferHandler(offerService),

		/*
		@handler Position
		@desc Handles lender position queries, portfolio summaries, and position claims.
		*/
		Position: apihandlers.NewPositionHandler(positionService),

		/*
		@handler Borrower
		@desc Handles borrower profile lookup, risk assessment, and registration.
		*/
		Borrower: apihandlers.NewBorrowerHandler(borrowerService),

		/*
		@handler BorrowerRequest
		@desc Handles the off-chain borrower access request queue (submit, check own status, admin list/reject).
		*/
		BorrowerRequest: apihandlers.NewBorrowerRequestHandler(borrowerRequestService),

		/*
		@handler Liquidation
		@desc Handles liquidation auction queries, bidding, and auction status checks.
		*/
		Liquidation: apihandlers.NewLiquidationHandler(liquidationService),

		/*
		@handler Withdrawal
		@desc Handles withdrawal requests, cancellations, and epoch processing.
		*/
		Withdrawal: apihandlers.NewWithdrawalHandler(withdrawalService),

		/*
		@handler Transaction
		@desc Handles building protocol transactions and estimating gas/fees.
		*/
		Transaction: apihandlers.NewTransactionHandler(transactionService),

		/*
		@handler Health
		@desc Handles system health checks and readiness probes.
		*/
		Health: apihandlers.NewHealthHandler(db, cfg.Blockchain.RPCURL),
	})

	server := &http.Server{
		Addr:    cfg.Server.Host + ":" + cfg.Server.Port,
		Handler: router,
	}

	go func() {
		logger.Info("HTTP server starting",
			"address", server.Addr,
			"environment", env,
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed",
				logger.WithError(err),
			)
			log.Fatalf("api server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutdown signal received, gracefully shutting down...")

	ctx, cancel = context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown failed",
			logger.WithError(err),
		)
	} else {
		logger.Info("Server shutdown complete")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("api graceful shutdown failed: %v", err)
	}
}
