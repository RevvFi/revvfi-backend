package handlers

import (
	"math/big"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/api/dto/response"
	"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@function marketResponse

@desc
Maps a market domain model to API response DTO.

@responsibilities
- Convert big.Int fields to strings
- Convert time fields to unix timestamps
- Preserve public market metadata

@params
- market: domain market model

@returns
- response.MarketResponse
*/
func marketResponse(market *models.Market) response.MarketResponse {
	totalDebt := stringOrZero(market.TotalDebt)
	return response.MarketResponse{
		Address:  market.Address,
		Borrower: market.Borrower,
		BorrowAsset: response.TokenInfo{
			Address:  market.BorrowAsset,
			Symbol:   market.BorrowAssetSymbol.String,
			Decimals: market.BorrowAssetDecimals,
		},
		CollateralAsset: response.TokenInfo{
			Address:  market.CollateralAsset,
			Symbol:   market.CollateralAssetSymbol.String,
			Decimals: market.CollateralAssetDecimals,
		},
		TotalPrincipal:  stringOrZero(market.TotalPrincipal),
		TotalLiquidity:  stringOrZero(market.TotalLiquidity),
		TotalDebt:       totalDebt,
		UtilizationRate: market.UtilizationRate,
		WeightedAPR:     market.WeightedAvgAPR,
		IsActive:        market.IsActive,
		IsLiquidating:   market.IsLiquidating,
		IsClosed:        market.IsClosed,
		CreatedAt:       market.CreatedAt.Unix(),
	}
}

/*
@function offerResponse

@desc
Maps an offer domain model to API response DTO.

@responsibilities
- Convert amounts to strings
- Calculate filled amount and fill percentage
- Preserve offer lifecycle metadata

@params
- offer: domain offer model

@returns
- response.OfferResponse
*/
func offerResponse(offer *models.Offer) response.OfferResponse {
	filled := big.NewInt(0)
	fillPercentage := 0.0
	if offer.Amount != nil && offer.RemainingAmount != nil {
		filled = new(big.Int).Sub(offer.Amount, offer.RemainingAmount)
		if offer.Amount.Sign() > 0 {
			ratio := new(big.Rat).SetFrac(filled, offer.Amount)
			value, _ := ratio.Float64()
			fillPercentage = value * 100
		}
	}

	return response.OfferResponse{
		OfferID:         offer.OfferID,
		Lender:          offer.Lender,
		MarketAddress:   offer.MarketAddress,
		Amount:          stringOrZero(offer.Amount),
		RemainingAmount: stringOrZero(offer.RemainingAmount),
		FilledAmount:    filled.String(),
		APR:             offer.APR,
		Seniority:       offer.Seniority,
		Status:          offer.Status,
		FillPercentage:  fillPercentage,
		Expiry:          offer.Expiry.Unix(),
		CreatedAt:       offer.CreatedAt.Unix(),
	}
}

/*
@function positionResponse

@desc
Maps a position domain model to API response DTO.

@responsibilities
- Calculate current value from principal and interest
- Convert nullable settlement time
- Preserve position status fields

@params
- position: domain position model

@returns
- response.PositionResponse
*/
func positionResponse(position *models.Position) response.PositionResponse {
	currentValue := big.NewInt(0)
	if position.CurrentPrincipal != nil {
		currentValue.Add(currentValue, position.CurrentPrincipal)
	}
	if position.AccruedInterest != nil {
		currentValue.Add(currentValue, position.AccruedInterest)
	}

	var settledAt *int64
	if position.SettledAt.Valid {
		value := position.SettledAt.Time.Unix()
		settledAt = &value
	}

	return response.PositionResponse{
		TokenID:         position.TokenID,
		Lender:          position.Lender,
		MarketAddress:   position.MarketAddress,
		Principal:       stringOrZero(position.Principal),
		CurrentValue:    currentValue.String(),
		AccruedInterest: stringOrZero(position.AccruedInterest),
		ClaimableAmount: stringOrZero(position.ClaimableAmount),
		APR:             position.APR,
		Seniority:       position.Seniority,
		Status:          position.Status,
		MintedAt:        position.MintedAt.Unix(),
		SettledAt:       settledAt,
	}
}

/*
@function borrowerResponse

@desc
Maps a borrower domain model to reputation response DTO.

@responsibilities
- Convert monetary fields to strings
- Convert nullable last activity timestamp
- Preserve reputation and risk metrics

@params
- borrower: domain borrower model

@returns
- response.BorrowerReputationResponse
*/
func borrowerResponse(borrower *models.Borrower) response.BorrowerReputationResponse {
	var lastActivity *int64
	if borrower.LastActivity.Valid {
		value := borrower.LastActivity.Time.Unix()
		lastActivity = &value
	}

	defaultRate := 0.0
	if borrower.TotalLoans > 0 {
		defaultRate = float64(borrower.DefaultedLoans) / float64(borrower.TotalLoans) * 100
	}

	return response.BorrowerReputationResponse{
		Address:         borrower.Address,
		ReputationScore: borrower.ReputationScore,
		RiskLabel:       borrower.RiskLabel,
		SuccessRate:     borrower.SuccessRate,
		DefaultRate:     defaultRate,
		TotalBorrowed:   stringOrZero(borrower.TotalBorrowed),
		TotalRepaid:     stringOrZero(borrower.TotalRepaid),
		OutstandingDebt: stringOrZero(borrower.OutstandingDebt),
		ActiveLoans:     borrower.ActiveLoans,
		FailedLoans:     borrower.DefaultedLoans,
		RegisteredAt:    borrower.RegisteredAt.Unix(),
		LastActivity:    lastActivity,
	}
}

/*
@function auctionResponse

@desc
Maps an auction domain model to API response DTO.

@responsibilities
- Convert bid and debt fields to strings
- Calculate remaining time
- Preserve auction lifecycle fields

@params
- auction: domain auction model

@returns
- response.AuctionResponse
*/
func auctionResponse(auction *models.Auction) response.AuctionResponse {
	remaining := max(time.Until(auction.EndTime), 0)

	return response.AuctionResponse{
		AuctionID:        auction.AuctionID,
		MarketAddress:    auction.MarketAddress,
		Borrower:         auction.BorrowerAddress,
		CollateralAmount: stringOrZero(auction.CollateralAmount),
		DebtAmount:       stringOrZero(auction.DebtAmount),
		CurrentPrice:     stringOrZero(auction.CurrentPrice),
		HighestBid:       stringOrZero(auction.HighestBid),
		HighestBidder:    auction.HighestBidder.String,
		Status:           auction.Status,
		TimeRemaining:    int64(remaining.Seconds()),
		StartTime:        auction.StartTime.Unix(),
		EndTime:          auction.EndTime.Unix(),
	}
}

/*
@function withdrawalResponse

@desc
Maps a withdrawal request domain model to API response DTO.

@responsibilities
- Convert amount fields to strings
- Calculate remaining amount when not persisted
- Preserve epoch and lifecycle fields

@params
- request: domain withdrawal request model

@returns
- response.WithdrawalRequestResponse
*/
func withdrawalResponse(request *models.WithdrawalRequest) response.WithdrawalRequestResponse {
	remaining := request.RemainingAmount
	if remaining == nil {
		remaining = big.NewInt(0)
		if request.RequestedAmount != nil {
			remaining.Set(request.RequestedAmount)
		}
		if request.FulfilledAmount != nil {
			remaining.Sub(remaining, request.FulfilledAmount)
		}
	}

	return response.WithdrawalRequestResponse{
		RequestID:       request.RequestID,
		Lender:          request.Lender,
		PositionID:      request.PositionID,
		MarketAddress:   request.MarketAddress,
		RequestedAmount: stringOrZero(request.RequestedAmount),
		FulfilledAmount: stringOrZero(request.FulfilledAmount),
		RemainingAmount: remaining.String(),
		EpochNumber:     request.EpochNumber,
		Status:          request.Status,
		CreatedAt:       request.CreatedAt.Unix(),
	}
}

/*
@function stringOrZero

@desc
Converts a nullable big.Int pointer into a decimal string.

@responsibilities
- Avoid nil pointer panics
- Keep API amount serialization consistent

@params
- value: integer pointer

@returns
- string
*/
func stringOrZero(value *big.Int) string {
	if value == nil {
		return "0"
	}
	return value.String()
}
