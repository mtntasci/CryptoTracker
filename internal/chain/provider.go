package chain

import (
	"context"

	"crypto-tracker/internal/model"
)

// Provider defines the modular interface that every blockchain adapter (UTXO or Account/Token) must implement.
type Provider interface {
	// ChainType returns the blockchain identifier (BTC, ETH, TRON).
	ChainType() model.ChainType

	// Architecture returns whether the chain uses UTXO or Account/Token mechanics.
	Architecture() model.ChainArchitecture

	// FetchAddressSummary retrieves on-chain metrics and public explorer tags for an address.
	FetchAddressSummary(ctx context.Context, address string) (*model.AddressSummary, error)

	// FetchOutboundTransactions retrieves transactions where funds or tokens leave the given address.
	FetchOutboundTransactions(ctx context.Context, address string, limit int) ([]model.Transaction, error)

	// ResolveHopTargets analyzes a transaction originating from sourceAddress, applying chain-specific
	// heuristics (dust filtering, change detection for UTXO; event log decoding & sweep detection for Account).
	ResolveHopTargets(ctx context.Context, tx model.Transaction, sourceAddress string) ([]model.HopTarget, error)
}

