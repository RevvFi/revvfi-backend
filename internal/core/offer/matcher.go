package offer

import (
"math/big"
"sort"

"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@file matcher.go

@desc
Offer matching algorithm for optimal liquidity allocation.
Matches borrow requests with available offers in APR order.

@responsibilities
- Prioritize offers by APR (lowest first)
- Match requested amount across multiple offers
- Respect seniority levels
- Calculate effective rates
*/

/*
@struct Matcher

@desc
Matches borrow requests with available offers.
*/
type Matcher struct{}

/*
@function NewMatcher

@desc
Creates new matcher instance.

@returns
- *Matcher
*/
func NewMatcher() *Matcher {
	return &Matcher{}
}

/*
@function MatchOffers

@desc
Matches borrow request with best available offers.
Offers sorted by APR (lowest first), then by seniority (senior first).

@params
- offers: available offers
- borrowAmount: amount to borrow in wei
- maxAPR: maximum acceptable APR in bps

@returns
- []models.Offer: sorted offers that match criteria
*/
func (m *Matcher) MatchOffers(
offers []models.Offer,
borrowAmount *big.Int,
maxAPR int32,
) []models.Offer {
	// Filter offers by criteria
	var filtered []models.Offer
	for _, offer := range offers {
		if offer.Status != "active" && offer.Status != "partially_filled" {
			continue
		}
		if offer.APR > maxAPR {
			continue
		}
		if offer.RemainingAmount.Sign() <= 0 {
			continue
		}
		filtered = append(filtered, offer)
	}

	// Sort by APR (ascending), then by seniority (senior first = 0 first)
	sort.Slice(filtered, func(i, j int) bool {
if filtered[i].APR != filtered[j].APR {
return filtered[i].APR < filtered[j].APR
		}
		return filtered[i].Seniority < filtered[j].Seniority
	})

	// Select offers until we have enough liquidity
	var matched []models.Offer
	accumulated := big.NewInt(0)

	for _, offer := range filtered {
		matched = append(matched, offer)
		accumulated.Add(accumulated, offer.RemainingAmount)

		if accumulated.Cmp(borrowAmount) >= 0 {
			break
		}
	}

	return matched
}

/*
@function RankOffersByAPR

@desc
Ranks offers from best (lowest APR) to worst (highest APR).

@params
- offers: offers to rank

@returns
- []models.Offer: ranked offers
*/
func (m *Matcher) RankOffersByAPR(offers []models.Offer) []models.Offer {
	ranked := make([]models.Offer, len(offers))
	copy(ranked, offers)

	sort.Slice(ranked, func(i, j int) bool {
return ranked[i].APR < ranked[j].APR
	})

	return ranked
}

/*
@function FilterByAPR

@desc
Filters offers within APR range.

@params
- offers: offers to filter
- minAPR: minimum APR in bps
- maxAPR: maximum APR in bps

@returns
- []models.Offer: filtered offers
*/
func (m *Matcher) FilterByAPR(
offers []models.Offer,
minAPR, maxAPR int32,
) []models.Offer {
	var filtered []models.Offer
	for _, offer := range offers {
		if offer.APR >= minAPR && offer.APR <= maxAPR {
			filtered = append(filtered, offer)
		}
	}
	return filtered
}

/*
@function FilterBySeniority

@desc
Filters offers by seniority level.

@params
- offers: offers to filter
- seniority: seniority level (0=Senior, 1=Junior)

@returns
- []models.Offer: filtered offers
*/
func (m *Matcher) FilterBySeniority(
offers []models.Offer,
seniority int16,
) []models.Offer {
	var filtered []models.Offer
	for _, offer := range offers {
		if offer.Seniority == seniority {
			filtered = append(filtered, offer)
		}
	}
	return filtered
}
