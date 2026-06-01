package response

/*
@file admin_protocol.go

@desc
Response DTOs for admin protocol configuration endpoints.
Provides protocol state, configuration parameters, and transaction preparation results.

@responsibilities
- Define protocol configuration response payloads
- Structure protocol state and metrics
- Support transaction preparation responses
*/

/*
@struct ProtocolConfig

@desc
Complete protocol configuration state.
Includes all protocol parameters and contract addresses.

@fields
- deployment_fee_eth: deployment fee in eth (human-readable)
- deployment_fee_wei: deployment fee in wei (for precision)
- fee_recipient: address receiving protocol fees
- arch_controller: arch controller contract address
- position_nft: position NFT contract address
- liquidator: liquidator contract address
- reputation_registry: reputation registry contract address
- timelock_delay_days: timelock delay in days
- is_core_contracts_set: whether core contracts are initialized
- last_updated_at: last configuration update timestamp
*/
type ProtocolConfig struct {
	DeploymentFeeEth   string `json:"deployment_fee_eth"`
	DeploymentFeeWei   string `json:"deployment_fee_wei"`
	FeeRecipient       string `json:"fee_recipient"`
	ArchController     string `json:"arch_controller"`
	PositionNFT        string `json:"position_nft"`
	Liquidator         string `json:"liquidator"`
	ReputationRegistry string `json:"reputation_registry"`
	TimelockDelayDays  int    `json:"timelock_delay_days"`
	IsCoreContractsSet bool   `json:"is_core_contracts_set"`
	LastUpdatedAt      int64  `json:"last_updated_at"`
}

/*
@struct ProtocolConfigResponse

@desc
Response wrapper for protocol configuration query.
Includes configuration data and pagination metadata.

@fields
- config: protocol configuration object
- pagination: pagination metadata
*/
type ProtocolConfigResponse struct {
	Config     *ProtocolConfig `json:"config"`
	Pagination *PaginationInfo `json:"pagination,omitempty"`
}

/*
@struct FeeCollectedResponse

@desc
Response containing accumulated fees collected by protocol.
Includes both wei precision and human-readable formats.

@fields
- total_fees_eth: total fees collected in eth
- total_fees_wei: total fees collected in wei
- recipient_address: current fee recipient
- collection_period_start: collection period start timestamp
- collection_period_end: collection period end timestamp
*/
type FeeCollectedResponse struct {
	TotalFeesEth          string `json:"total_fees_eth"`
	TotalFeesWei          string `json:"total_fees_wei"`
	RecipientAddress      string `json:"recipient_address"`
	CollectionPeriodStart int64  `json:"collection_period_start"`
	CollectionPeriodEnd   int64  `json:"collection_period_end"`
}

/*
@struct ContractStatus

@desc
Status of core contract deployment.
Indicates which contracts are initialized and active.

@fields
- is_core_contracts_set: whether core contracts initialized
- position_nft_address: position NFT contract address
- liquidator_address: liquidator contract address
- reputation_registry_address: reputation registry address
- deployment_block: block number of deployment
- deployment_timestamp: deployment timestamp
*/
type ContractStatus struct {
	IsCoreContractsSet        bool   `json:"is_core_contracts_set"`
	PositionNFTAddress        string `json:"position_nft_address"`
	LiquidatorAddress         string `json:"liquidator_address"`
	ReputationRegistryAddress string `json:"reputation_registry_address"`
	DeploymentBlock           uint64 `json:"deployment_block"`
	DeploymentTimestamp       int64  `json:"deployment_timestamp"`
}

/*
@struct PendingUpgrade

@desc
Pending timelock upgrade waiting for execution.
Includes upgrade details and timelock status.

@fields
- upgrade_id: unique upgrade identifier
- target_address: target contract address
- new_implementation: new implementation address
- upgrade_type: type of upgrade (market_impl, etc.)
- timelock_expiry: when timelock expires and upgrade can execute
- created_at: when upgrade was scheduled
- eta: estimated time of execution
*/
type PendingUpgrade struct {
	UpgradeID         string `json:"upgrade_id"`
	TargetAddress     string `json:"target_address"`
	NewImplementation string `json:"new_implementation"`
	UpgradeType       string `json:"upgrade_type"`
	TimelockExpiry    int64  `json:"timelock_expiry"`
	CreatedAt         int64  `json:"created_at"`
	ETA               int64  `json:"eta"`
}

/*
@struct PendingUpgradesResponse

@desc
Response containing pending protocol upgrades.
Includes list of upgrades and pagination.

@fields
- upgrades: list of pending upgrades
- pagination: pagination metadata
- total_pending: total count of pending upgrades
*/
type PendingUpgradesResponse struct {
	Upgrades      []PendingUpgrade `json:"upgrades"`
	Pagination    *PaginationInfo  `json:"pagination"`
	TotalPending  int              `json:"total_pending"`
}

/*
@struct TimelockStatus

@desc
Status of timelock mechanism.
Provides timelock parameters and delay information.

@fields
- delay_seconds: current timelock delay in seconds
- delay_days: delay in days (human-readable)
- min_delay: minimum allowed delay
- max_delay: maximum allowed delay
*/
type TimelockStatus struct {
	DelaySeconds int `json:"delay_seconds"`
	DelayDays    int `json:"delay_days"`
	MinDelay     int `json:"min_delay"`
	MaxDelay     int `json:"max_delay"`
}

/*
@struct TransactionPreparation

@desc
Prepared transaction data for admin actions.
Contains encoded transaction data ready for execution.

@fields
- tx_data: encoded transaction calldata
- target: target contract address
- value: eth value to send (in wei)
- description: human-readable description of action
- encoded_proposal: fully encoded proposal for governance
*/
type TransactionPreparation struct {
	TxData           string `json:"tx_data"`
	Target           string `json:"target"`
	Value            string `json:"value"`
	Description      string `json:"description"`
	EncodedProposal  string `json:"encoded_proposal"`
}
/*
@struct UpgradeQueueItem

@desc
Item in the upgrade queue waiting for timelock execution.
Used for monitoring pending upgrades.

@fields
- id: unique queue item identifier
- upgrade_type: type of upgrade (market, factory, arch_controller)
- target_address: contract being upgraded
- new_address: new implementation address
- queue_position: position in queue (0 = next)
- ready_at: timestamp when ready for execution
- status: pending, ready, executing, executed, cancelled
- proposed_by: address that proposed the upgrade
- proposed_at: when upgrade was proposed
*/
type UpgradeQueueItem struct {
    ID             string `json:"id"`
    UpgradeType    string `json:"upgrade_type"`
    TargetAddress  string `json:"target_address"`
    NewAddress     string `json:"new_address"`
    QueuePosition  int    `json:"queue_position"`
    ReadyAt        int64  `json:"ready_at"`
    Status         string `json:"status"`
    ProposedBy     string `json:"proposed_by"`
    ProposedAt     int64  `json:"proposed_at"`
}
