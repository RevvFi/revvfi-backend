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
	"github.com/Revvfi/revvfi-backend/internal/core/auth"
	"github.com/Revvfi/revvfi-backend/internal/core/borrower"
	"github.com/Revvfi/revvfi-backend/internal/core/liquidation"
	"github.com/Revvfi/revvfi-backend/internal/core/market"
	"github.com/Revvfi/revvfi-backend/internal/core/offer"
	"github.com/Revvfi/revvfi-backend/internal/core/position"
	"github.com/Revvfi/revvfi-backend/internal/core/transaction"
	"github.com/Revvfi/revvfi-backend/internal/core/withdrawal"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := postgres.Open(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

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
	*/
	authService := auth.NewAuthService(authRepo, jwtMgr)

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
	@construction HTTPHandlers
	@desc Initialize all HTTP request handlers with service dependencies.
	Includes auth, market, admin, offer, position, borrower, liquidation, withdrawal, and transaction handlers.
	*/
	router := gin.New()
	routes.Register(router, cfg, authService, routes.Handlers{
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
		Health: apihandlers.NewHealthHandler(db),
	})

	server := &http.Server{
		Addr:    cfg.Server.Host + ":" + cfg.Server.Port,
		Handler: router,
	}

	go func() {
		log.Printf("revvfi api listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("api server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("api graceful shutdown failed: %v", err)
	}
}
