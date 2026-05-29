package types

import "math/big"

/*
@struct Event

@desc
Normalized blockchain event payload for processor input.

@responsibilities
- Preserve log identity for idempotency
- Carry decoded event fields as attributes
*/
type Event struct {
	Name            string
	ContractAddress string
	TxHash          string
	LogIndex        int32
	BlockNumber     int64
	Attributes      map[string]interface{}
}

/*
@struct OraclePrice

@desc
Normalized oracle price result.

@responsibilities
- Carry price answer and oracle metadata
- Preserve update timestamp for freshness checks
*/
type OraclePrice struct {
	OracleAddress string
	Price         *big.Int
	Decimals      uint8
	UpdatedAt     int64
	RoundID       uint64
}
