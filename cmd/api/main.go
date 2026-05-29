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

	authService := auth.NewAuthService(authRepo, jwtMgr)
	marketService := market.NewMarketService(marketRepo, postgres.NewMarketOfferRepository(offerRepo), postgres.NewMarketPositionRepository(positionRepo))
	offerService := offer.NewOfferService(offerRepo)
	positionService := position.NewPositionService(positionRepo)
	borrowerService := borrower.NewBorrowerService(borrowerRepo)
	liquidationService := liquidation.NewLiquidationService(auctionRepo, marketRepo, borrowerRepo)
	withdrawalService := withdrawal.NewWithdrawalService(withdrawalRepo, positionRepo)
	transactionService := transaction.NewTransactionService()

	router := gin.New()
	routes.Register(router, cfg, authService, routes.Handlers{
		Auth:        apihandlers.NewAuthHandler(authService),
		Market:      apihandlers.NewMarketHandler(marketService),
		Offer:       apihandlers.NewOfferHandler(offerService),
		Position:    apihandlers.NewPositionHandler(positionService),
		Borrower:    apihandlers.NewBorrowerHandler(borrowerService),
		Liquidation: apihandlers.NewLiquidationHandler(liquidationService),
		Withdrawal:  apihandlers.NewWithdrawalHandler(withdrawalService),
		Transaction: apihandlers.NewTransactionHandler(transactionService),
		Health:      apihandlers.NewHealthHandler(db),
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
