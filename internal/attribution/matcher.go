package attribution

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"crypto-tracker/internal/config"
	"crypto-tracker/internal/model"
)

// Engine resolves wallet addresses to known exchange entities or behavioral classifications.
type Engine struct {
	cfg      *config.Config
	client   *http.Client
	registry map[string]KnownWalletEntry
	cacheMu  sync.RWMutex
	cache    map[string]*model.AttributionInfo
}

// NewEngine creates a new Attribution Engine with verified registry and live keyless API lookup.
func NewEngine(cfg *config.Config) *Engine {
	return &Engine{
		cfg:      cfg,
		client:   NewHTTPClient(cfg),
		registry: buildVerifiedRegistry(),
		cache:    make(map[string]*model.AttributionInfo),
	}
}

// IdentifyFast checks the in-memory cache and verified internal hot wallet registry without network I/O.
func (e *Engine) IdentifyFast(chain model.ChainType, address string) *model.AttributionInfo {
	norm := normalizeAddress(chain, address)

	e.cacheMu.RLock()
	if cached, exists := e.cache[norm]; exists {
		e.cacheMu.RUnlock()
		return cached
	}
	e.cacheMu.RUnlock()

	if entry, found := e.registry[norm]; found {
		info := &model.AttributionInfo{
			Address:     address,
			Chain:       chain,
			EntityName:  entry.EntityName,
			WalletLabel: entry.Label,
			Category:    entry.Category,
			Confidence:  1.0,
			Source:      "Verified Exchange Hot Wallet Registry",
			IsTerminal:  true,
		}
		e.storeCache(norm, info)
		return info
	}

	return nil
}

// Identify checks whether an address belongs to a known exchange hot/cold wallet.
func (e *Engine) Identify(ctx context.Context, chain model.ChainType, address string, summary *model.AddressSummary) *model.AttributionInfo {
	if fast := e.IdentifyFast(chain, address); fast != nil {
		return fast
	}
	norm := normalizeAddress(chain, address)

	// 3. Check if AddressSummary already carries an explorer label
	if summary != nil {
		if summary.ExplorerLabel != "" {
			if entity, ok := MatchLabelToExchange(summary.ExplorerLabel); ok {
				info := &model.AttributionInfo{
					Address:     address,
					Chain:       chain,
					EntityName:  entity,
					WalletLabel: summary.ExplorerLabel,
					Category:    model.CategoryExchangeHotWallet,
					Confidence:  0.95,
					Source:      "Public Explorer Tag",
					IsTerminal:  true,
				}
				e.storeCache(norm, info)
				return info
			}
		}
		for _, tag := range summary.ExplorerTags {
			if entity, ok := MatchLabelToExchange(tag); ok {
				info := &model.AttributionInfo{
					Address:     address,
					Chain:       chain,
					EntityName:  entity,
					WalletLabel: tag,
					Category:    model.CategoryExchangeHotWallet,
					Confidence:  0.92,
					Source:      "Public Explorer Metadata",
					IsTerminal:  true,
				}
				e.storeCache(norm, info)
				return info
			}
		}
	}

	// 4. Query live keyless public label APIs
	switch chain {
	case model.ChainEthereum:
		if info, err := queryBlockscoutLabel(ctx, e.client, e.cfg.BlockscoutETHURL, address); err == nil && info != nil {
			e.storeCache(norm, info)
			return info
		}
	case model.ChainTron:
		if info, err := queryTronscanLabel(ctx, e.client, e.cfg.TronscanBaseURL, address); err == nil && info != nil {
			e.storeCache(norm, info)
			return info
		}
	}

	// 5. Behavioral Hot-Wallet / Sweep Cluster Detection based on on-chain metrics
	if summary != nil && (summary.TxCount >= 25000 || summary.FundedCount >= 15000) {
		count := summary.TxCount
		if summary.FundedCount > count {
			count = summary.FundedCount
		}
		info := &model.AttributionInfo{
			Address:     address,
			Chain:       chain,
			EntityName:  "Institutional / Exchange Hot Cluster",
			WalletLabel: fmt.Sprintf("High-Frequency Consolidation Wallet (%d+ txs)", count),
			Category:    model.CategoryExchangeHotWallet,
			Confidence:  0.78,
			Source:      "Behavioral Sweep Consolidation Heuristic",
			IsTerminal:  true,
		}
		e.storeCache(norm, info)
		return info
	}

	return nil
}

func (e *Engine) storeCache(key string, info *model.AttributionInfo) {
	e.cacheMu.Lock()
	defer e.cacheMu.Unlock()
	e.cache[key] = info
}
