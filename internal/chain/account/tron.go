package account

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

const sunPerTRX = 1e6

// TronProvider implements chain.Provider for Tron and TRC-20 Smart Contract Event Logs (USDT TRC-20)
// using keyless TronGrid v1 and Tronscan Public APIs.
type TronProvider struct {
	cfg    *config.Config
	client *http.Client
}

// NewTronProvider creates a new Tron & TRC-20 provider.
func NewTronProvider(cfg *config.Config) *TronProvider {
	return &TronProvider{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
	}
}

func (p *TronProvider) ChainType() model.ChainType {
	return model.ChainTron
}

func (p *TronProvider) Architecture() model.ChainArchitecture {
	return model.ArchAccount
}

type tronscanAccountResp struct {
	Name          string `json:"name"`
	PublicTag     string `json:"publicTag"`
	BlueTag       string `json:"blueTag"`
	GreyTag       string `json:"greyTag"`
	Balance       int64  `json:"balance"`
	TotalTxCount  int    `json:"totalTransactionCount"`
	TransactionsIn  int  `json:"transactions_in"`
	TransactionsOut int  `json:"transactions_out"`
}

type trongridTRC20Resp struct {
	Data []struct {
		TransactionID  string `json:"transaction_id"`
		BlockTimestamp int64  `json:"block_timestamp"`
		From           string `json:"from"`
		To             string `json:"to"`
		Type           string `json:"type"`
		Value          string `json:"value"`
		TokenInfo      struct {
			Symbol   string `json:"symbol"`
			Address  string `json:"address"`
			Decimals int    `json:"decimals"`
			Name     string `json:"name"`
		} `json:"token_info"`
	} `json:"data"`
}

type tronscanTRC20TransfersResp struct {
	TokenTransfers []struct {
		TransactionID string `json:"transaction_id"`
		BlockTs       int64  `json:"block_ts"`
		Block         int64  `json:"block"`
		FromAddress   string `json:"from_address"`
		ToAddress     string `json:"to_address"`
		Quant         string `json:"quant"`
		TokenInfo     struct {
			TokenAbbr    string `json:"tokenAbbr"`
			TokenName    string `json:"tokenName"`
			TokenDecimal int    `json:"tokenDecimal"`
			TokenID      string `json:"tokenId"`
		} `json:"tokenInfo"`
	} `json:"token_transfers"`
}

// FetchAddressSummary fetches Tron account statistics and public tags via keyless Tronscan API.
func (p *TronProvider) FetchAddressSummary(ctx context.Context, address string) (*model.AddressSummary, error) {
	trimmed := strings.TrimSpace(address)
	url := fmt.Sprintf("%s/accountv2?address=%s", strings.TrimRight(p.cfg.TronscanBaseURL, "/"), trimmed)

	var resp tronscanAccountResp
	if err := p.getJSON(ctx, url, &resp); err != nil {
		return &model.AddressSummary{
			Address: trimmed,
			Chain:   model.ChainTron,
		}, nil
	}

	var tags []string
	for _, t := range []string{resp.PublicTag, resp.BlueTag, resp.GreyTag} {
		if t != "" {
			tags = append(tags, t)
		}
	}

	label := resp.PublicTag
	if label == "" {
		label = resp.Name
	}

	return &model.AddressSummary{
		Address:       trimmed,
		Chain:         model.ChainTron,
		TxCount:       resp.TotalTxCount,
		FundedCount:   resp.TransactionsIn,
		SpentCount:    resp.TransactionsOut,
		BalanceFloat:  float64(resp.Balance) / sunPerTRX,
		ExplorerLabel: label,
		ExplorerTags:  tags,
	}, nil
}

// FetchOutboundTransactions retrieves outbound TRC-20 smart contract Transfer events (USDT TRC-20, etc.)
// from TronGrid v1 (keyless) with automatic fallback to Tronscan public API.
func (p *TronProvider) FetchOutboundTransactions(ctx context.Context, address string, limit int) ([]model.Transaction, error) {
	trimmed := strings.TrimSpace(address)
	fetchLimit := 25
	if limit > 0 && limit < fetchLimit {
		fetchLimit = limit * 2
	}

	// 1. Primary: Keyless TronGrid v1 TRC-20 Event Log API
	trongridURL := fmt.Sprintf("%s/v1/accounts/%s/transactions/trc20?only_from=true&limit=%d",
		strings.TrimRight(p.cfg.TronGridBaseURL, "/"), trimmed, fetchLimit)

	var tgResp trongridTRC20Resp
	var out []model.Transaction

	if err := p.getJSON(ctx, trongridURL, &tgResp); err == nil && len(tgResp.Data) > 0 {
		for _, item := range tgResp.Data {
			if !strings.EqualFold(item.From, trimmed) || item.To == "" {
				continue
			}

			dec := item.TokenInfo.Decimals
			if dec <= 0 {
				dec = 6
			}
			rawVal := ParseDecimalString(item.Value)
			valFloat := FormatTokenUnits(rawVal, dec)

			symbol := item.TokenInfo.Symbol
			if symbol == "" {
				symbol = "TRC20"
			}

			ts := time.Now()
			if item.BlockTimestamp > 0 {
				ts = time.UnixMilli(item.BlockTimestamp)
			}

			out = append(out, model.Transaction{
				TxID:          item.TransactionID,
				Chain:         model.ChainTron,
				Timestamp:     ts,
				From:          item.From,
				To:            item.To,
				ValueFloat:    valFloat,
				AssetSymbol:   symbol,
				IsTokenTx:     true,
				ContractEvent: "Transfer(address,address,uint256)",
				TokenTransfers: []model.TokenTransfer{
					{
						ContractAddress: item.TokenInfo.Address,
						TokenSymbol:     symbol,
						TokenName:       item.TokenInfo.Name,
						Decimals:        dec,
						From:            item.From,
						To:              item.To,
						ValueRaw:        rawVal,
						ValueFloat:      valFloat,
					},
				},
			})

			if limit > 0 && len(out) >= limit {
				return out, nil
			}
		}
	}

	// 2. Fallback: Keyless Tronscan TRC-20 Transfers API
	if len(out) == 0 {
		tsURL := fmt.Sprintf("%s/filter/trc20/transfers?limit=%d&start=0&fromAddress=%s",
			strings.TrimRight(p.cfg.TronscanBaseURL, "/"), fetchLimit, trimmed)
		var tsResp tronscanTRC20TransfersResp
		if err := p.getJSON(ctx, tsURL, &tsResp); err == nil {
			for _, item := range tsResp.TokenTransfers {
				if !strings.EqualFold(item.FromAddress, trimmed) || item.ToAddress == "" {
					continue
				}
				dec := item.TokenInfo.TokenDecimal
				if dec <= 0 {
					dec = 6
				}
				rawVal := ParseDecimalString(item.Quant)
				valFloat := FormatTokenUnits(rawVal, dec)

				symbol := item.TokenInfo.TokenAbbr
				if symbol == "" {
					symbol = "USDT"
				}

				ts := time.Now()
				if item.BlockTs > 0 {
					ts = time.UnixMilli(item.BlockTs)
				}

				out = append(out, model.Transaction{
					TxID:          item.TransactionID,
					Chain:         model.ChainTron,
					BlockHeight:   item.Block,
					Timestamp:     ts,
					From:          item.FromAddress,
					To:            item.ToAddress,
					ValueFloat:    valFloat,
					AssetSymbol:   symbol,
					IsTokenTx:     true,
					ContractEvent: "Transfer(address,address,uint256)",
					TokenTransfers: []model.TokenTransfer{
						{
							ContractAddress: item.TokenInfo.TokenID,
							TokenSymbol:     symbol,
							TokenName:       item.TokenInfo.TokenName,
							Decimals:        dec,
							From:            item.FromAddress,
							To:              item.ToAddress,
							ValueRaw:        rawVal,
							ValueFloat:      valFloat,
						},
					},
				})

				if limit > 0 && len(out) >= limit {
					break
				}
			}
		}
	}

	return out, nil
}

// ResolveHopTargets evaluates TRC-20 event logs and TRX outbound transfers.
func (p *TronProvider) ResolveHopTargets(_ context.Context, tx model.Transaction, sourceAddress string) ([]model.HopTarget, error) {
	return ResolveAccountHopTargets(tx, sourceAddress, p.cfg), nil
}

func (p *TronProvider) getJSON(ctx context.Context, url string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "crypto-tracker/1.0")
	if p.cfg.TronGridAPIKey != "" && strings.Contains(url, "trongrid.io") {
		req.Header.Set("TRON-PRO-API-KEY", p.cfg.TronGridAPIKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

