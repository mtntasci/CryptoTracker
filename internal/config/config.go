package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds application-wide settings for tracing, API endpoints, and heuristics.
type Config struct {
	// Tracing Engine Parameters
	MaxDepth        int           // Default: 5 hops
	MaxBranches     int           // Max outbound branches followed per address
	MaxTxPerAddress int           // Max recent outbound transactions inspected per hop
	RequestTimeout  time.Duration // HTTP timeout per request
	StepDelay       time.Duration // Visual pacing delay for TUI hop animation

	// Heuristic Thresholds
	BTCDustThresholdSats int64   // Default: 1000 sats (filters 546-sat dust & spam)
	ETHDustThreshold     float64 // Default: 0.0001 ETH
	TokenDustThreshold   float64 // Default: 1.0 USDT/USDC
	TRXDustThreshold     float64 // Default: 1.0 TRX

	// Keyless Public Endpoints
	MempoolBaseURL     string
	BlockstreamBaseURL string
	BlockscoutETHURL   string
	PublicETHRPCURL    string
	TronGridBaseURL    string
	TronscanBaseURL    string

	// Optional API Keys (not required; used only if present in environment)
	EtherscanAPIKey string
	TronGridAPIKey  string
}

// DefaultConfig returns production-ready defaults using free, keyless public APIs and MaxDepth = 5.
func DefaultConfig() *Config {
	cfg := &Config{
		MaxDepth:             5,
		MaxBranches:          4,
		MaxTxPerAddress:      15,
		RequestTimeout:       15 * time.Second,
		StepDelay:            350 * time.Millisecond,
		BTCDustThresholdSats: 1000, // 0.00001000 BTC
		ETHDustThreshold:     0.0001,
		TokenDustThreshold:   1.0,
		TRXDustThreshold:     1.0,

		MempoolBaseURL:     getEnv("MEMPOOL_API_URL", "https://mempool.space/api"),
		BlockstreamBaseURL: getEnv("BLOCKSTREAM_API_URL", "https://blockstream.info/api"),
		BlockscoutETHURL:   getEnv("BLOCKSCOUT_ETH_URL", "https://eth.blockscout.com/api/v2"),
		PublicETHRPCURL:    getEnv("ETH_RPC_URL", "https://ethereum-rpc.publicnode.com"),
		TronGridBaseURL:    getEnv("TRONGRID_API_URL", "https://api.trongrid.io"),
		TronscanBaseURL:    getEnv("TRONSCAN_API_URL", "https://apilist.tronscanapi.com/api"),

		EtherscanAPIKey: os.Getenv("ETHERSCAN_API_KEY"),
		TronGridAPIKey:  os.Getenv("TRONGRID_API_KEY"),
	}

	if v := os.Getenv("MAX_DEPTH"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 15 {
			cfg.MaxDepth = parsed
		}
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

