package response

/*
@file health.go

@desc
Response DTOs for health and status endpoints.
Returns service health, readiness, and sync status.
*/

/*
@struct HealthResponse

@desc
Response containing overall service health status.

@fields
- Status: healthy|degraded|unhealthy
- Database: database connectivity status
- RPC: blockchain RPC connectivity
- Cache: cache service status
- Uptime: service uptime in seconds
- Timestamp: health check timestamp
*/
type HealthResponse struct {
	Status    string `json:"status"`
	Database  string `json:"database"`
	RPC       string `json:"rpc"`
	Cache     string `json:"cache,omitempty"`
	Uptime    int64  `json:"uptime"`
	Timestamp int64  `json:"timestamp"`
}

/*
@struct OracleHealthResponse

@desc
Response containing oracle freshness and health.

@fields
- Status: healthy|stale|unavailable
- Oracles: individual oracle statuses
- LastUpdate: last oracle update timestamp
- UpdateLag: seconds since last update
*/
type OracleHealthResponse struct {
	Status     string            `json:"status"`
	Oracles    []OracleStatus    `json:"oracles"`
	LastUpdate int64             `json:"last_update"`
	UpdateLag  int64             `json:"update_lag"`
}

/*
@struct OracleStatus

@desc
Individual oracle status.

@fields
- Address: oracle address
- Pair: trading pair (e.g., ETH/USD)
- LastUpdate: last price update
- IsStale: if data is stale
- Age: seconds since update
- Price: current price
*/
type OracleStatus struct {
	Address    string `json:"address"`
	Pair       string `json:"pair"`
	LastUpdate int64  `json:"last_update"`
	IsStale    bool   `json:"is_stale"`
	Age        int64  `json:"age"`
	Price      string `json:"price,omitempty"`
}

/*
@struct SyncStatusResponse

@desc
Response containing blockchain indexer sync status.

@fields
- Status: syncing|synced|lagging|failed
- LatestBlock: latest indexed block
- LatestBlockHash: latest block hash
- LatestBlockTime: latest block timestamp
- ChainHeight: current chain height
- LagBlocks: blocks behind
- LagSeconds: time lag in seconds
- SyncProgress: percentage synced (0-100)
*/
type SyncStatusResponse struct {
	Status           string  `json:"status"`
	LatestBlock      int64   `json:"latest_block"`
	LatestBlockHash  string  `json:"latest_block_hash"`
	LatestBlockTime  int64   `json:"latest_block_time"`
	ChainHeight      int64   `json:"chain_height"`
	LagBlocks        int64   `json:"lag_blocks"`
	LagSeconds       int64   `json:"lag_seconds"`
	SyncProgress     float64 `json:"sync_progress"`
}

/*
@struct ReadinessResponse

@desc
Response for readiness probe (Kubernetes liveness).

@fields
- Ready: if service is ready
- Reason: reason if not ready
*/
type ReadinessResponse struct {
	Ready  bool   `json:"ready"`
	Reason string `json:"reason,omitempty"`
}
