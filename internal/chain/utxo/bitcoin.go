package utxo

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"crypto-tracker/internal/config"
	"crypto-tracker/internal/model"
)

const satsPerBTC = 1e8

// BitcoinProvider implements chain.Provider for Bitcoin (UTXO model) via keyless Esplora APIs
// (Mempool.space primary, Blockstream.info automatic fallback).
type BitcoinProvider struct {
	cfg    *config.Config
	client *http.Client
}

// NewBitcoinProvider creates a new Bitcoin UTXO chain provider.
func NewBitcoinProvider(cfg *config.Config) *BitcoinProvider {
	return &BitcoinProvider{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
	}
}

func (p *BitcoinProvider) ChainType() model.ChainType {
	return model.ChainBitcoin
}

func (p *BitcoinProvider) Architecture() model.ChainArchitecture {
	return model.ArchUTXO
}

type esploraAddressResponse struct {
	Address    string `json:"address"`
	ChainStats struct {
		FundedTxoCount int   `json:"funded_txo_count"`
		FundedTxoSum   int64 `json:"funded_txo_sum"`
		SpentTxoCount  int   `json:"spent_txo_count"`
		SpentTxoSum    int64 `json:"spent_txo_sum"`
		TxCount        int   `json:"tx_count"`
	} `json:"chain_stats"`
}

type esploraTx struct {
	TxID   string `json:"txid"`
	Fee    int64  `json:"fee"`
	Status struct {
		Confirmed   bool  `json:"confirmed"`
		BlockHeight int64 `json:"block_height"`
		BlockTime   int64 `json:"block_time"`
	} `json:"status"`
	Vin []struct {
		TxID    string `json:"txid"`
		Vout    uint32 `json:"vout"`
		Prevout *struct {
			ScriptPubKeyAddress string `json:"scriptpubkey_address"`
			ScriptPubKeyType    string `json:"scriptpubkey_type"`
			Value               int64  `json:"value"`
		} `json:"prevout"`
	} `json:"vin"`
	Vout []struct {
		ScriptPubKeyAddress string `json:"scriptpubkey_address"`
		ScriptPubKeyType    string `json:"scriptpubkey_type"`
		Value               int64  `json:"value"`
	} `json:"vout"`
}

// FetchAddressSummary fetches on-chain statistics for a Bitcoin address from Mempool.space (or Blockstream fallback).
func (p *BitcoinProvider) FetchAddressSummary(ctx context.Context, address string) (*model.AddressSummary, error) {
	path := fmt.Sprintf("/address/%s", strings.TrimSpace(address))
	var resp esploraAddressResponse
	if err := p.getJSONWithFallback(ctx, path, &resp); err != nil {
		return nil, err
	}

	balSats := resp.ChainStats.FundedTxoSum - resp.ChainStats.SpentTxoSum
	if balSats < 0 {
		balSats = 0
	}

	return &model.AddressSummary{
		Address:      address,
		Chain:        model.ChainBitcoin,
		TxCount:      resp.ChainStats.TxCount,
		FundedCount:  resp.ChainStats.FundedTxoCount,
		SpentCount:   resp.ChainStats.SpentTxoCount,
		BalanceFloat: float64(balSats) / satsPerBTC,
	}, nil
}

// FetchOutboundTransactions retrieves recent transactions where source address appears in Vin (spending UTXOs).
func (p *BitcoinProvider) FetchOutboundTransactions(ctx context.Context, address string, limit int) ([]model.Transaction, error) {
	trimmedAddr := strings.TrimSpace(address)
	path := fmt.Sprintf("/address/%s/txs", trimmedAddr)

	var rawTxs []esploraTx
	if err := p.getJSONWithFallback(ctx, path, &rawTxs); err != nil {
		return nil, err
	}

	var outbound []model.Transaction
	for _, rtx := range rawTxs {
		isSender := false
		var inputs []model.TxInput
		for _, vin := range rtx.Vin {
			if vin.Prevout == nil {
				continue
			}
			inAddr := vin.Prevout.ScriptPubKeyAddress
			if strings.EqualFold(inAddr, trimmedAddr) {
				isSender = true
			}
			inputs = append(inputs, model.TxInput{
				PrevTxID:   vin.TxID,
				Vout:       vin.Vout,
				Address:    inAddr,
				ValueRaw:   big.NewInt(vin.Prevout.Value),
				ValueFloat: float64(vin.Prevout.Value) / satsPerBTC,
				ScriptType: model.ScriptType(vin.Prevout.ScriptPubKeyType),
			})
		}

		// Only trace transactions where the target address is spending funds outbound
		if !isSender {
			continue
		}

		var outputs []model.TxOutput
		for idx, vout := range rtx.Vout {
			outputs = append(outputs, model.TxOutput{
				Index:      uint32(idx),
				Address:    vout.ScriptPubKeyAddress,
				ValueRaw:   big.NewInt(vout.Value),
				ValueFloat: float64(vout.Value) / satsPerBTC,
				ScriptType: model.ScriptType(vout.ScriptPubKeyType),
			})
		}

		ts := time.Now()
		if rtx.Status.BlockTime > 0 {
			ts = time.Unix(rtx.Status.BlockTime, 0)
		}

		outbound = append(outbound, model.Transaction{
			TxID:        rtx.TxID,
			Chain:       model.ChainBitcoin,
			BlockHeight: rtx.Status.BlockHeight,
			Timestamp:   ts,
			FeeFloat:    float64(rtx.Fee) / satsPerBTC,
			Inputs:      inputs,
			Outputs:     outputs,
			AssetSymbol: "BTC",
		})

		if limit > 0 && len(outbound) >= limit {
			break
		}
	}

	return outbound, nil
}

// ResolveHopTargets applies UTXO dust filtering and change detection heuristics.
func (p *BitcoinProvider) ResolveHopTargets(_ context.Context, tx model.Transaction, sourceAddress string) ([]model.HopTarget, error) {
	targets := AnalyzeUTXOOutputs(tx, sourceAddress, p.cfg.BTCDustThresholdSats)
	return targets, nil
}

func (p *BitcoinProvider) getJSONWithFallback(ctx context.Context, path string, target any) error {
	endpoints := []string{
		strings.TrimRight(p.cfg.MempoolBaseURL, "/") + path,
		strings.TrimRight(p.cfg.BlockstreamBaseURL, "/") + path,
	}

	var lastErr error
	for _, url := range endpoints {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "crypto-tracker/1.0")

		resp, err := p.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("endpoint %s returned HTTP %d", url, resp.StatusCode)
			continue
		}

		err = json.NewDecoder(resp.Body).Decode(target)
		resp.Body.Close()
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("bitcoin API request failed: %w", lastErr)
}

