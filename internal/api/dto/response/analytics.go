package response

/*
@file analytics.go

@desc
Response DTOs for analytics and statistics endpoints.
Returns protocol metrics and leaderboards.
*/

/*
@struct GlobalStatsResponse

@desc
Response containing global protocol statistics.

@fields
- TotalTVL: total value locked
- TotalVolume: total volume traded
- ActiveMarkets: number of active markets
- ActiveUsers: number of active users
- AvgAPR: protocol average APR
- TotalBorrowed: total borrowed amount
- TotalLent: total lent amount
*/
type GlobalStatsResponse struct {
	TotalTVL       string  `json:"total_tvl"`
	TotalVolume    string  `json:"total_volume"`
	ActiveMarkets  int32   `json:"active_markets"`
	ActiveUsers    int32   `json:"active_users"`
	AvgAPR         float64 `json:"avg_apr"`
	TotalBorrowed  string  `json:"total_borrowed"`
	TotalLent      string  `json:"total_lent"`
	ProtocolUptime float64 `json:"protocol_uptime"`
}

/*
@struct TrendingResponse

@desc
Response containing trending markets and activity.

@fields
- Markets: trending market data
- Growth24h: 24-hour growth percentage
- TopMarkets: top performing markets
- MostActive: most active markets
*/
type TrendingResponse struct {
	Markets      []TrendingMarket `json:"markets"`
	Growth24h    float64          `json:"growth_24h"`
	TopMarkets   []MarketResponse `json:"top_markets"`
	MostActive   []MarketResponse `json:"most_active"`
}

/*
@struct TrendingMarket

@desc
Trending market data point.

@fields
- Address: market address
- TVL24hChange: 24h TVL change
- Volume24h: 24h volume
- ActivityScore: activity ranking score
*/
type TrendingMarket struct {
	Address       string  `json:"address"`
	TVL24hChange  float64 `json:"tvl_24h_change"`
	Volume24h     string  `json:"volume_24h"`
	ActivityScore int32   `json:"activity_score"`
}

/*
@struct LenderLeaderboardResponse

@desc
Response containing top lender rankings.

@fields
- Leaderboard: ranked lender data
- YourRank: authenticated lender rank (if applicable)
- YourStats: authenticated lender stats
*/
type LenderLeaderboardResponse struct {
	Leaderboard []LenderRank         `json:"leaderboard"`
	YourRank    *LenderRank          `json:"your_rank,omitempty"`
	YourStats   *LenderStatsResponse `json:"your_stats,omitempty"`
}

/*
@struct LenderRank

@desc
Ranked lender information.

@fields
- Rank: position in ranking
- Address: lender address
- TotalSupplied: total capital supplied
- TotalEarned: total interest earned
- AvgAPR: average APR
- PositionCount: number of positions
- ReputationScore: lender reputation
*/
type LenderRank struct {
	Rank             int32  `json:"rank"`
	Address          string `json:"address"`
	TotalSupplied    string `json:"total_supplied"`
	TotalEarned      string `json:"total_earned"`
	AvgAPR           float64 `json:"avg_apr"`
	PositionCount    int32  `json:"position_count"`
	ReputationScore  int32  `json:"reputation_score"`
}

/*
@struct LenderStatsResponse

@desc
Lender statistics and performance.

@fields
- TotalSupplied: total supplied amount
- TotalEarned: total earned amount
- AvgAPR: average APR
- PositionCount: number of positions
- SuccessRate: successful transaction rate
*/
type LenderStatsResponse struct {
	TotalSupplied  string  `json:"total_supplied"`
	TotalEarned    string  `json:"total_earned"`
	AvgAPR         float64 `json:"avg_apr"`
	PositionCount  int32   `json:"position_count"`
	SuccessRate    float64 `json:"success_rate"`
}

/*
@struct BorrowerLeaderboardResponse

@desc
Response containing top borrower rankings.

@fields
- Leaderboard: ranked borrower data
- YourRank: authenticated borrower rank
- YourStats: authenticated borrower stats
*/
type BorrowerLeaderboardResponse struct {
	Leaderboard []BorrowerRank         `json:"leaderboard"`
	YourRank    *BorrowerRank          `json:"your_rank,omitempty"`
	YourStats   *BorrowerStatsResponse `json:"your_stats,omitempty"`
}

/*
@struct BorrowerRank

@desc
Ranked borrower information.

@fields
- Rank: position in ranking
- Address: borrower address
- ReputationScore: reputation score
- RiskLabel: AAA|AA|A|B|C|D
- TotalBorrowed: cumulative borrowed
- SuccessRate: repayment success rate
*/
type BorrowerRank struct {
	Rank            int32   `json:"rank"`
	Address         string  `json:"address"`
	ReputationScore int32   `json:"reputation_score"`
	RiskLabel       string  `json:"risk_label"`
	TotalBorrowed   string  `json:"total_borrowed"`
	SuccessRate     float64 `json:"success_rate"`
}

/*
@struct BorrowerStatsResponse

@desc
Borrower statistics and performance.

@fields
- ReputationScore: reputation score
- RiskLabel: risk label
- TotalBorrowed: total borrowed
- OutstandingDebt: current debt
- SuccessRate: success rate
- ActiveLoans: active loans count
*/
type BorrowerStatsResponse struct {
	ReputationScore int32   `json:"reputation_score"`
	RiskLabel       string  `json:"risk_label"`
	TotalBorrowed   string  `json:"total_borrowed"`
	OutstandingDebt string  `json:"outstanding_debt"`
	SuccessRate     float64 `json:"success_rate"`
	ActiveLoans     int32   `json:"active_loans"`
}
