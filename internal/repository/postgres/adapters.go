package postgres

import (
	"context"
	"math/big"

	"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@struct MarketOfferRepository

@desc
Adapter exposing the offer methods required by MarketService.

@responsibilities
- Bridge local service interface method signatures
- Reuse PostgreSQL offer repository queries
*/
type MarketOfferRepository struct {
	offers *OfferRepository
}

/*
@function NewMarketOfferRepository

@desc
Creates a market-service offer repository adapter.

@params
- offers: PostgreSQL offer repository

@returns
- *MarketOfferRepository
*/
func NewMarketOfferRepository(offers *OfferRepository) *MarketOfferRepository {
	return &MarketOfferRepository{offers: offers}
}

/*
@method GetActiveByMarket

@desc
Returns active offers for market metric calculations.
*/
func (r *MarketOfferRepository) GetActiveByMarket(ctx context.Context, marketAddr string) ([]models.Offer, error) {
	return r.offers.GetActiveByMarket(ctx, marketAddr, 1000, 0)
}

/*
@method CountActive

@desc
Counts active offers for market metric calculations.
*/
func (r *MarketOfferRepository) CountActive(ctx context.Context, marketAddr string) (int32, error) {
	return r.offers.CountActive(ctx, marketAddr)
}

/*
@struct MarketPositionRepository

@desc
Adapter exposing position aggregate methods required by MarketService.

@responsibilities
- Count active positions by market
- Sum active principal by market
*/
type MarketPositionRepository struct {
	positions *PositionRepository
}

/*
@function NewMarketPositionRepository

@desc
Creates a market-service position repository adapter.

@params
- positions: PostgreSQL position repository

@returns
- *MarketPositionRepository
*/
func NewMarketPositionRepository(positions *PositionRepository) *MarketPositionRepository {
	return &MarketPositionRepository{positions: positions}
}

/*
@method CountActiveByMarket

@desc
Counts active positions for a market.
*/
func (r *MarketPositionRepository) CountActiveByMarket(ctx context.Context, marketAddr string) (int32, error) {
	return r.positions.CountActiveByMarket(ctx, marketAddr)
}

/*
@method GetTotalPrincipalByMarket

@desc
Sums active principal for a market.
*/
func (r *MarketPositionRepository) GetTotalPrincipalByMarket(ctx context.Context, marketAddr string) (*big.Int, error) {
	return r.positions.GetTotalPrincipalByMarket(ctx, marketAddr)
}
