package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

/*
@struct Config

@desc
Holds runtime configuration for the API, database, authentication, rate limiting, and blockchain integration.

@responsibilities
- Centralize environment-backed configuration
- Provide typed values to application wiring
- Keep package consumers independent from raw environment variables
*/
type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	JWT         JWTConfig
	CORS        CORSConfig
	RateLimit   RateLimitConfig
	Blockchain  BlockchainConfig
	Auth        AuthConfig
}

/*
@struct ServerConfig

@desc
Defines HTTP server runtime settings.

@responsibilities
- Configure host, port, route prefix, and shutdown timeout
*/
type ServerConfig struct {
	Host            string
	Port            string
	BasePath        string
	ShutdownTimeout time.Duration
}

/*
@struct DatabaseConfig

@desc
Defines PostgreSQL connection and pool settings.

@responsibilities
- Configure database/sql PostgreSQL connections
- Keep pool sizing and SSL settings explicit
*/
type DatabaseConfig struct {
	URL             string
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

/*
@struct JWTConfig

@desc
Defines JWT signing and lifecycle settings.

@responsibilities
- Configure auth token signing secret
- Configure token TTL and issuer
*/
type JWTConfig struct {
	Secret string
	TTL    time.Duration
	Issuer string
}

/*
@struct CORSConfig

@desc
Defines browser cross-origin request policy.

@responsibilities
- Configure allowed origins for API clients
- Configure allowed headers and methods
*/
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

/*
@struct RateLimitConfig

@desc
Defines in-memory per-client rate limiting policy.

@responsibilities
- Configure request allowance per window
- Configure window duration
*/
type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

/*
@struct BlockchainConfig

@desc
Defines Ethereum and RevvFi contract integration settings.

@responsibilities
- Configure RPC and chain identity
- Configure ABI artifact path
- Configure deployed protocol contract addresses
*/
type BlockchainConfig struct {
	RPCURL                    string
	ChainID                   int64
	ArtifactPath              string
	MulticallAddress          string
	FactoryAddress            string
	ArchControllerAddress     string
	PositionNFTAddress        string
	LiquidatorAddress         string
	ReputationRegistryAddress string
	MarketAddress             string
	OfferBookAddress          string
	CollateralEscrowAddress   string
	StartBlock                uint64
	AdminWallets              []string
}

/*
@struct AuthConfig

@desc
Defines SIWE authentication configuration.

@responsibilities
- Configure domain for SIWE messages
- Configure URI for authentication
- Configure chain ID for signature verification
*/
type AuthConfig struct {
	Domain    string        `json:"domain"`
	URI       string        `json:"uri"`
	Statement string        `json:"statement"`
	Version   string        `json:"version"`
	ChainID   int64         `json:"chain_id"`
	NonceTTL  time.Duration `json:"nonce_ttl"`
	SessionTTL time.Duration `json:"session_ttl"`
}

/*
@function Load

@desc
Loads application configuration from environment variables with development-safe defaults.

@responsibilities
- Read and parse environment variables
- Apply defaults for local development
- Return typed configuration to application entrypoints

@returns
- *Config
- error
*/
func Load() (*Config, error) {
	// Load base .env first
	if err := loadDotEnvUpward(".env"); err != nil {
		return nil, err
	}

	// Check ENVIRONMENT to determine which env file to load
	env := getEnv("ENVIRONMENT", "development")

	// Load environment-specific config
	if env == "production" {
		if err := loadDotEnvUpward(".env.production"); err != nil {
			return nil, err
		}
	} else {
		// Load .env.local for local development
		if err := loadDotEnvUpward(".env.local"); err != nil {
			return nil, err
		}
	}

	cfg := &Config{
		Environment: env,
		Server: ServerConfig{
			Host:            getEnv("API_HOST", "0.0.0.0"),
			Port:            getEnv("API_PORT", "3000"),
			BasePath:        getEnv("API_BASE_PATH", "/api/v1"),
			ShutdownTimeout: getDurationEnv("API_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Database: DatabaseConfig{
			URL:             firstEnv("DB_URL", "DATABASE_URL"),
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			Name:            getEnv("DB_NAME", "revvfi_db"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getIntEnv("DB_MAX_CONN", 25),
			MaxIdleConns:    getIntEnv("DB_MIN_CONN", 5),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 30*time.Minute),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "development-jwt-secret-change-me-32-bytes"),
			TTL:    time.Duration(getIntEnv("JWT_EXPIRY", 86400)) * time.Second,
			Issuer: getEnv("JWT_ISSUER", "revvfi"),
		},
		CORS: CORSConfig{
			AllowedOrigins: splitEnv("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:3001"}),
			AllowedMethods: splitEnv("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
			AllowedHeaders: splitEnv("CORS_ALLOWED_HEADERS", []string{"Authorization", "Content-Type", "X-Requested-With", "X-Correlation-ID"}),
		},
		RateLimit: RateLimitConfig{
			Requests: getIntEnv("RATE_LIMIT_REQUESTS", 120),
			Window:   getDurationEnv("RATE_LIMIT_WINDOW", time.Minute),
		},
		Blockchain: BlockchainConfig{
			RPCURL:                    getEnv("RPC_URL", ""),
			ChainID:                   int64(getIntEnv("CHAIN_ID", 1)),
			ArtifactPath:              getEnv("ARTIFACT_PATH", "../out"),
			MulticallAddress:          getEnv("MULTICALL_ADDRESS", ""),
			FactoryAddress:            getEnv("FACTORY_ADDRESS", ""),
			ArchControllerAddress:     getEnv("ARCH_CONTROLLER_ADDRESS", ""),
			PositionNFTAddress:        getEnv("POSITION_NFT_ADDRESS", ""),
			LiquidatorAddress:         getEnv("LIQUIDATOR_ADDRESS", ""),
			ReputationRegistryAddress: getEnv("REPUTATION_REGISTRY_ADDRESS", ""),
			MarketAddress:             getEnv("MARKET_ADDRESS", ""),
			OfferBookAddress:          getEnv("OFFERBOOK_ADDRESS", ""),
			CollateralEscrowAddress:   getEnv("COLLATERAL_ESCROW_ADDRESS", ""),
			AdminWallets:              splitEnv("ADMIN_WALLETS", []string{}),
		},
		Auth: AuthConfig{ 
			Domain:     getEnv("AUTH_DOMAIN", "revvfi.com"),
			URI:        getEnv("AUTH_URI", "https://revvfi.com/api/v1/auth/login"),
			Statement:  getEnv("AUTH_STATEMENT", "Sign in to RevvFi Protocol\n\nThis signature will not trigger a blockchain transaction or cost any gas.\n\nBy signing, you agree to the RevvFi Terms of Service (https://revvfi.com/terms)"),
			Version:    getEnv("AUTH_VERSION", "1"),
			ChainID:    int64(getIntEnv("AUTH_CHAIN_ID", 1)),
			NonceTTL:   getDurationEnv("AUTH_NONCE_TTL", 5*time.Minute),
			SessionTTL: getDurationEnv("AUTH_SESSION_TTL", 24*time.Hour),
		},
	}

	return cfg, nil
}

/*
@function loadDotEnvUpward

@desc
Finds and loads a dotenv file from the current directory or one of its parents.

@responsibilities
- Support running binaries from cmd/api or repository root
- Stop at filesystem root when no dotenv file exists
- Delegate dotenv parsing to loadDotEnv

@params
- filename: dotenv filename to locate

@returns
- error
*/
func loadDotEnvUpward(filename string) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	for {
		candidate := filepath.Join(dir, filename)
		if _, err := os.Stat(candidate); err == nil {
			return loadDotEnv(candidate)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("stat %s: %w", candidate, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return nil
		}
		dir = parent
	}
}

/*
@method DSN

@desc
Builds a lib/pq connection string from database configuration.

@responsibilities
- Format PostgreSQL connection settings
- Avoid leaking connection formatting into callers

@returns
- string
*/
func (c DatabaseConfig) DSN() string {
	if strings.TrimSpace(c.URL) != "" {
		return postgresURLWithSSL(c.URL)
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Name,
		c.SSLMode,
	)
}

/*
@function loadDotEnv

@desc
Loads simple KEY=VALUE pairs from a dotenv file when it exists.

@responsibilities
- Support local .env and .env.local development files
- Preserve already-exported environment variables
- Avoid adding another runtime dependency for dotenv parsing

@params
- path: dotenv file path

@returns
- error
*/
func loadDotEnv(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s: %w", path, err)
	}

	lines := strings.Split(string(raw), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		key, value, ok := strings.Cut(trimmed, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set %s from %s: %w", key, path, err)
		}
	}

	return nil
}

/*
@function postgresURLWithSSL

@desc
Ensures URL-style PostgreSQL DSNs include an SSL mode for hosted providers.

@responsibilities
- Preserve explicitly configured SSL modes
- Default URL DSNs to sslmode=require for Supabase compatibility

@params
- value: PostgreSQL URL

@returns
- string
*/
func postgresURLWithSSL(value string) string {
	if strings.Contains(value, "sslmode=") {
		return value
	}
	separator := "?"
	if strings.Contains(value, "?") {
		separator = "&"
	}
	return value + separator + "sslmode=require"
}

/*
@function firstEnv

@desc
Returns the first non-empty value from a list of environment variables.

@responsibilities
- Support alias environment variable names
- Keep precedence order explicit

@params
- keys: environment variable names

@returns
- string
*/
func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

/*
@function getEnv

@desc
Reads an environment variable with a fallback default.

@responsibilities
- Return configured value when present
- Return fallback when empty

@returns
- string
*/
func getEnv(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

/*
@function getIntEnv

@desc
Reads an integer environment variable with a fallback default.

@responsibilities
- Parse integer values
- Return fallback on missing or invalid input

@returns
- int
*/
func getIntEnv(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

/*
@function getDurationEnv

@desc
Reads a duration environment variable with a fallback default.

@responsibilities
- Parse Go duration strings
- Return fallback on missing or invalid input

@returns
- time.Duration
*/
func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

/*
@function splitEnv

@desc
Reads a comma-separated environment variable into a string slice.

@responsibilities
- Split comma-separated config values
- Trim empty values
- Return fallback when no configured values are present

@returns
- []string
*/
func splitEnv(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}
