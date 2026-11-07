package attribution

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"crypto-tracker/internal/config"
	"crypto-tracker/internal/model"
)

// knownExchangeKeywords maps lowercase exchange keywords to canonical Exchange Entity Names.
var knownExchangeKeywords = []struct {
	Keyword string
	Entity  string
}{
	{"binance", "Binance"},
	{"okx", "OKX"},
	{"okex", "OKX"},
	{"coinbase", "Coinbase"},
	{"kraken", "Kraken"},
	{"bybit", "Bybit"},
	{"bitfinex", "Bitfinex"},
	{"huobi", "HTX (Huobi)"},
	{"htx", "HTX (Huobi)"},
	{"kucoin", "KuCoin"},
	{"gate.io", "Gate.io"},
	{"gateio", "Gate.io"},
	{"bitget", "Bitget"},
	{"mexc", "MEXC"},
	{"crypto.com", "Crypto.com"},
	{"gemini", "Gemini"},
	{"bitstamp", "Bitstamp"},
	{"upbit", "Upbit"},
	{"bithumb", "Bithumb"},
	{"deribit", "Deribit"},
	{"robinhood", "Robinhood"},
	{"paribu", "Paribu"},
	{"btcturk", "BtcTurk"},
}

// MatchLabelToExchange checks if a public explorer tag/label matches a known exchange.
func MatchLabelToExchange(label string) (entityName string, isExchange bool) {
	lower := strings.ToLower(strings.TrimSpace(label))
	if lower == "" {
		return "", false
	}
	for _, kw := range knownExchangeKeywords {
		if strings.Contains(lower, kw.Keyword) {
			return kw.Entity, true
		}
	}
	return "", false
}

// queryBlockscoutLabel queries Blockscout's keyless v2 API for public address tags on Ethereum.
func queryBlockscoutLabel(ctx context.Context, client *http.Client, baseURL, address string) (*model.AttributionInfo, error) {
	url := fmt.Sprintf("%s/addresses/%s", strings.TrimRight(baseURL, "/"), address)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "crypto-tracker/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("blockscout status %d", resp.StatusCode)
	}

	var payload struct {
		Name       string `json:"name"`
		IsContract bool   `json:"is_contract"`
		PublicTags []struct {
			Label       string `json:"label"`
			DisplayName string `json:"display_name"`
		} `json:"public_tags"`
		Metadata *struct {
			Tags []struct {
				Name    string `json:"name"`
				TagType string `json:"tagType"`
			} `json:"tags"`
		} `json:"metadata"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	var candidateLabels []string
	if payload.Name != "" {
		candidateLabels = append(candidateLabels, payload.Name)
	}
	for _, t := range payload.PublicTags {
		if t.DisplayName != "" {
			candidateLabels = append(candidateLabels, t.DisplayName)
		}
		if t.Label != "" {
			candidateLabels = append(candidateLabels, t.Label)
		}
	}
	if payload.Metadata != nil {
		for _, t := range payload.Metadata.Tags {
			if t.Name != "" {
				candidateLabels = append(candidateLabels, t.Name)
			}
		}
	}

	for _, lbl := range candidateLabels {
		if entity, ok := MatchLabelToExchange(lbl); ok {
			return &model.AttributionInfo{
				Address:     address,
				Chain:       model.ChainEthereum,
				EntityName:  entity,
				WalletLabel: lbl,
				Category:    model.CategoryExchangeHotWallet,
				Confidence:  0.95,
				Source:      "Blockscout Public Tag API",
				IsTerminal:  true,
			}, nil
		}
	}

	return nil, nil
}

// queryTronscanLabel queries Tronscan's keyless public API for account tags on Tron.
func queryTronscanLabel(ctx context.Context, client *http.Client, baseURL, address string) (*model.AttributionInfo, error) {
	url := fmt.Sprintf("%s/accountv2?address=%s", strings.TrimRight(baseURL, "/"), address)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "crypto-tracker/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tronscan status %d", resp.StatusCode)
	}

	var payload struct {
		Name            string `json:"name"`
		PublicTag       string `json:"publicTag"`
		BlueTag         string `json:"blueTag"`
		GreyTag         string `json:"greyTag"`
		TransactionsIn  int    `json:"transactions_in"`
		TransactionsOut int    `json:"transactions_out"`
		TotalTxCount    int    `json:"totalTransactionCount"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	candidates := []string{payload.PublicTag, payload.BlueTag, payload.GreyTag, payload.Name}
	for _, lbl := range candidates {
		if entity, ok := MatchLabelToExchange(lbl); ok {
			return &model.AttributionInfo{
				Address:     address,
				Chain:       model.ChainTron,
				EntityName:  entity,
				WalletLabel: lbl,
				Category:    model.CategoryExchangeHotWallet,
				Confidence:  0.95,
				Source:      "Tronscan Public Tag API",
				IsTerminal:  true,
			}, nil
		}
	}

	// Behavioral detection: Massive transaction volume on Tron account
	if payload.TotalTxCount > 50000 {
		tagName := "High-Volume Consolidation Hot Wallet"
		if payload.PublicTag != "" {
			tagName = payload.PublicTag
		}
		return &model.AttributionInfo{
			Address:     address,
			Chain:       model.ChainTron,
			EntityName:  "Unlabeled Exchange / Service Cluster",
			WalletLabel: fmt.Sprintf("%s (%d+ txs)", tagName, payload.TotalTxCount),
			Category:    model.CategoryExchangeHotWallet,
			Confidence:  0.75,
			Source:      "Behavioral Sweep / Volume Heuristic",
			IsTerminal:  true,
		}, nil
	}

	return nil, nil
}

// NewHTTPClient creates a shared HTTP client with timeout.
func NewHTTPClient(cfg *config.Config) *http.Client {
	timeout := 15 * time.Second
	if cfg != nil && cfg.RequestTimeout > 0 {
		timeout = cfg.RequestTimeout
	}
	return &http.Client{
		Timeout: timeout,
	}
}

