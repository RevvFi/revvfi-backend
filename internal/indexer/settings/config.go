package settings

import "time"

/*@
 * Config
 * @desc Leaf-package indexer settings shared by worker and subpackages.
 *
 * Keeping this type outside package indexer prevents subpackages such as
 * poller from importing their parent package, which would create an import
 * cycle once the parent imports those subpackages.
 */
type Config struct {
	// RPC Configuration
	RPCURL  string
	ChainID int64

	// ABI artifacts
	ArtifactPath string

	// Block Processing
	StartBlock         int64
	BlockConfirmations int
	PollInterval       time.Duration

	// Contract Addresses
	FactoryAddress            string
	ArchControllerAddress     string
	PositionNFTAddress        string
	LiquidatorAddress         string
	ReputationRegistryAddress string
	MarketAddresses           []string
	MarketAddress             string
	OfferBookAddress          string
	CollateralEscrowAddress   string
	// Performance
	BatchSize    int
	MaxRetries   int
	RetryBackoff time.Duration

	// Database
	DBConnectionString string
}

/*@
 * DefaultConfig
 * @desc Returns conservative defaults for local indexer execution.
 */
func DefaultConfig() *Config {
	return &Config{
		StartBlock:         0,
		BlockConfirmations: 0, // for development purpoose i was getting lot of errror because of this
		PollInterval:       3 * time.Second,
		BatchSize:          100,
		MaxRetries:         5,
		RetryBackoff:       2 * time.Second,
	}
}
