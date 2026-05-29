package client

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"strings"
)

/*
@struct EthereumClient

@desc
Minimal Ethereum JSON-RPC client wrapper.

@responsibilities
- Store RPC endpoint and chain ID
- Provide connectivity health checks
- Leave transaction signing to user wallets
*/
type EthereumClient struct {
	rpcURL     string
	chainID    *big.Int
	httpClient *http.Client
}

/*
@function NewEthereumClient

@desc
Creates an Ethereum RPC client wrapper.

@params
- rpcURL: Ethereum JSON-RPC endpoint
- chainID: chain identifier

@returns
- *EthereumClient
- error
*/
func NewEthereumClient(rpcURL string, chainID int64) (*EthereumClient, error) {
	if strings.TrimSpace(rpcURL) == "" {
		return nil, fmt.Errorf("rpc url is required")
	}
	return &EthereumClient{rpcURL: rpcURL, chainID: big.NewInt(chainID), httpClient: http.DefaultClient}, nil
}

/*
@method ChainID

@desc
Returns configured chain ID.

@returns
- *big.Int
*/
func (c *EthereumClient) ChainID() *big.Int {
	return new(big.Int).Set(c.chainID)
}

/*
@method Ping

@desc
Checks whether the configured RPC endpoint is reachable.

@params
- ctx: request lifecycle context

@returns
- error
*/
func (c *EthereumClient) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.rpcURL, strings.NewReader(`{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}`))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("rpc status %d", resp.StatusCode)
	}
	return nil
}
