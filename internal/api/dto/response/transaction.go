package response

/*
@file transaction.go

@desc
Response DTOs for transaction building endpoints.
Returns transaction payloads for wallet signing.
*/

/*
@struct TransactionQuoteResponse

@desc
Response from transaction gas estimation.

@fields
- GasEstimate: estimated gas units
- GasPrice: current gas price (wei)
- MaxFeePerGas: EIP-1559 max fee
- MaxPriorityFeePerGas: EIP-1559 priority fee
- EstimatedCost: total estimated cost
- EstimatedTime: estimated confirmation time
*/
type TransactionQuoteResponse struct {
	GasEstimate         string  `json:"gas_estimate"`
	GasPrice            string  `json:"gas_price"`
	MaxFeePerGas        string  `json:"max_fee_per_gas,omitempty"`
	MaxPriorityFeePerGas string  `json:"max_priority_fee_per_gas,omitempty"`
	EstimatedCost       string  `json:"estimated_cost"`
	EstimatedTime       int64   `json:"estimated_time"`
}

/*
@struct TransactionBuildResponse

@desc
Response containing unsigned transaction payload for wallet signing.

@fields
- To: recipient address
- From: sender address
- Value: ETH value (0 for most operations)
- Data: transaction calldata
- GasLimit: gas limit
- GasPrice: gas price (or maxFeePerGas for EIP-1559)
- Nonce: transaction nonce
- ChainID: chain identifier
*/
type TransactionBuildResponse struct {
	To        string `json:"to"`
	From      string `json:"from"`
	Value     string `json:"value"`
	Data      string `json:"data"`
	GasLimit  string `json:"gas_limit"`
	GasPrice  string `json:"gas_price,omitempty"`
	MaxFeePerGas string `json:"max_fee_per_gas,omitempty"`
	MaxPriorityFeePerGas string `json:"max_priority_fee_per_gas,omitempty"`
	Nonce     int64  `json:"nonce"`
	ChainID   int64  `json:"chain_id"`
}

/*
@struct MultiCallResponse

@desc
Response containing batched multicall transaction payload.

@fields
- Targets: array of contract addresses
- Calldata: array of call data
- Value: array of ETH values
- GasLimit: total gas limit
- To: Multicall3 contract address
- Data: encoded multicall data
*/
type MultiCallResponse struct {
	Targets    []string `json:"targets"`
	Calldata   []string `json:"calldata"`
	Value      []string `json:"value"`
	GasLimit   string   `json:"gas_limit"`
	To         string   `json:"to"`
	Data       string   `json:"data"`
}

/*
@struct GasPriceResponse

@desc
Response containing current network gas pricing.

@fields
- GasPrice: standard gas price
- SafeGasPrice: safe gas price
- StandardGasPrice: standard gas price
- FastGasPrice: fast gas price
- UrgentGasPrice: urgent gas price
- MaxFeePerGas: EIP-1559 max fee (if available)
- MaxPriorityFeePerGas: EIP-1559 priority fee (if available)
- BaseFee: EIP-1559 base fee (if available)
*/
type GasPriceResponse struct {
	GasPrice            string `json:"gas_price,omitempty"`
	SafeGasPrice        string `json:"safe_gas_price,omitempty"`
	StandardGasPrice    string `json:"standard_gas_price,omitempty"`
	FastGasPrice        string `json:"fast_gas_price,omitempty"`
	UrgentGasPrice      string `json:"urgent_gas_price,omitempty"`
	MaxFeePerGas        string `json:"max_fee_per_gas,omitempty"`
	MaxPriorityFeePerGas string `json:"max_priority_fee_per_gas,omitempty"`
	BaseFee             string `json:"base_fee,omitempty"`
}
