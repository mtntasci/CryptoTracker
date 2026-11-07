package model

import (
	"fmt"
	"strings"
)

// ChainType represents supported blockchain networks.
type ChainType string

const (
	ChainBitcoin  ChainType = "BTC"
	ChainEthereum ChainType = "ETH"
	ChainTron     ChainType = "TRON"
)

// ChainArchitecture distinguishes between UTXO and Account/Token chains.
type ChainArchitecture string

const (
	ArchUTXO    ChainArchitecture = "UTXO"
	ArchAccount ChainArchitecture = "ACCOUNT"
)

// ChainMetadata holds descriptive details about a supported blockchain.
type ChainMetadata struct {
	Type         ChainType
	Name         string
	Symbol       string
	Architecture ChainArchitecture
	ExplorerURL  string
}

// AllChains returns metadata for all supported chains.
func AllChains() []ChainMetadata {
	return []ChainMetadata{
		{
			Type:         ChainBitcoin,
			Name:         "Bitcoin (UTXO)",
			Symbol:       "BTC",
			Architecture: ArchUTXO,
			ExplorerURL:  "https://mempool.space",
		},
		{
			Type:         ChainEthereum,
			Name:         "Ethereum / USDT ERC-20",
			Symbol:       "ETH/USDT",
			Architecture: ArchAccount,
			ExplorerURL:  "https://etherscan.io",
		},
		{
			Type:         ChainTron,
			Name:         "Tron / USDT TRC-20",
			Symbol:       "TRX/USDT",
			Architecture: ArchAccount,
			ExplorerURL:  "https://tronscan.org",
		},
	}
}

// GetChainMetadata returns metadata for a given ChainType.
func GetChainMetadata(c ChainType) ChainMetadata {
	for _, meta := range AllChains() {
		if meta.Type == c {
			return meta
		}
	}
	return ChainMetadata{
		Type:         c,
		Name:         string(c),
		Symbol:       string(c),
		Architecture: ArchAccount,
	}
}

// DetectChainFromAddress attempts to infer the blockchain from an address format.
func DetectChainFromAddress(addr string) (ChainType, error) {
	trimmed := strings.TrimSpace(addr)
	if trimmed == "" {
		return "", fmt.Errorf("address cannot be empty")
	}

	if strings.HasPrefix(trimmed, "0x") && len(trimmed) == 42 {
		return ChainEthereum, nil
	}
	if strings.HasPrefix(trimmed, "T") && len(trimmed) == 34 {
		return ChainTron, nil
	}
	if strings.HasPrefix(trimmed, "bc1") || strings.HasPrefix(trimmed, "1") || strings.HasPrefix(trimmed, "3") {
		if len(trimmed) >= 26 && len(trimmed) <= 90 {
			return ChainBitcoin, nil
		}
	}

	return "", fmt.Errorf("unable to auto-detect chain for address %q", trimmed)
}

