package request

type SetDeploymentFeeRequest struct {
	FeeWei string `json:"fee_wei" binding:"required"`
	FeeEth string `json:"fee_eth"` // Optional human-readable field
}

type SetFeeRecipientRequest struct {
	RecipientAddress string `json:"recipient_address" binding:"required,eth_addr"`
	Reason           string `json:"reason" binding:"max=500"`
}

type SetCoreContractsRequest struct {
	PositionNFT        string `json:"position_nft" binding:"required,eth_addr"`
	Liquidator         string `json:"liquidator" binding:"required,eth_addr"`
	ReputationRegistry string `json:"reputation_registry" binding:"required,eth_addr"`
}

type SetMarketImplRequest struct {
	MarketImplAddress string `json:"market_impl_address" binding:"required,eth_addr"`
	Reason            string `json:"reason" binding:"required,max=500"`
}

type WithdrawFeesRequest struct {
	Amount    string `json:"amount" binding:"required"`
	Recipient string `json:"recipient" binding:"required,eth_addr"`
}

type CancelUpgradeRequest struct {
	Reason string `json:"reason" binding:"max=500"`
}