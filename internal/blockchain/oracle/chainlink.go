package oracle

import (
	"context"
	"fmt"
	"math/big"

	"github.com/Revvfi/revvfi-backend/internal/blockchain/types"
)

/*
@interface RPCReader

@desc
Minimal RPC dependency for oracle reads.

@responsibilities
- Provide a context-aware call boundary for future Ethereum RPC integration
*/
type RPCReader interface {
	Ping(ctx context.Context) error
}

/*
@struct ChainlinkClient

@desc
Chainlink oracle integration scaffold.

@responsibilities
- Hold RPC dependency
- Expose latest price read contract
- Enforce oracle freshness policy
*/
type ChainlinkClient struct {
	rpc RPCReader
}

/*
@function NewChainlinkClient

@desc
Creates a Chainlink oracle client.

@params
- rpc: RPC dependency

@returns
- *ChainlinkClient
*/
func NewChainlinkClient(rpc RPCReader) *ChainlinkClient {
	return &ChainlinkClient{rpc: rpc}
}

/*
@method LatestPrice

@desc
Returns latest oracle price data.

@params
- ctx: request lifecycle context
- oracleAddress: oracle contract address

@returns
- *types.OraclePrice
- error
*/
func (c *ChainlinkClient) LatestPrice(ctx context.Context, oracleAddress string) (*types.OraclePrice, error) {
	if c.rpc == nil {
		return nil, fmt.Errorf("rpc client is not configured")
	}
	if err := c.rpc.Ping(ctx); err != nil {
		return nil, err
	}
	return &types.OraclePrice{OracleAddress: oracleAddress, Price: big.NewInt(0), Decimals: 8}, nil
}
