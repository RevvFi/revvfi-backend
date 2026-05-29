package types

/*
@struct ContractRegistry

@desc
Stores configured RevvFi protocol contract addresses.

@responsibilities
- Keep deployed contract addresses in one typed value
- Provide blockchain clients and indexers with address metadata
*/
type ContractRegistry struct {
	Factory            string
	ArchController     string
	PositionNFT        string
	Liquidator         string
	ReputationRegistry string
	Multicall          string
}
