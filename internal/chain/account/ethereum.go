package account

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"crypto-tracker/internal/config"
	"crypto-tracker/internal/model"
)

// EthereumProvider implements chain.Provider for Ethereum and ERC-20 Token Event Logs
// using keyless Blockscout v2 API and Public Ethereum JSON-RPC.
type EthereumProvider struct {
	cfg    *config.Config
	client *http.Client
}

// NewEthereumProvider creates a new Ethereum & ERC-20 provider.
func NewEthereumProvider(cfg *config.Config) *EthereumProvider {
	return &EthereumProvider{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
	}
}

func (p *EthereumProvider) ChainType() model.ChainType {
	return model.ChainEthereum
}

func (p *EthereumProvider) Architecture() model.ChainArchitecture {
	return model.ArchAccount
}

type blockscoutAddressResp struct {
	Hash        string `json:"hash"`
	Name        string `json:"name"`
	IsContract  bool   `json:"is_contract"`
	CoinBalance string `json:"coin_balance"`
	PublicTags  []struct {
		Label       string `json:"label"`
		DisplayName string `json:"display_name"`
	} `json:"public_tags"`
	Metadata *struct {
		Tags []struct {
			Name string `json:"name"`
		} `json:"tags"`
	} `json:"metadata"`
}

type blockscoutTokenTransfersResp struct {
	Items []struct {
		TxHash      string `json:"tx_hash"`
		Transaction string `json:"transaction_hash"`
		Timestamp   string `json:"timestamp"`
		BlockNumber int64  `json:"block_number"`
		From        struct {
			Hash string `json:"hash"`
			Name string `json:"name"`
		} `json:"from"`
		To struct {
			Hash string `json:"hash"`
			Name string `json:"name"`
		} `json:"to"`
		Token struct {
			Address  string `json:"address"`
			Address2 string `json:"address_hash"`
			Symbol   string `json:"symbol"`
			Name     string `json:"name"`
			Decimals string `json:"decimals"`
		} `json:"token"`
		Total struct {
			Value    string `json:"value"`
			Decimals string `json:"decimals"`
		} `json:"total"`
	} `json:"items"`
}

type blockscoutTxsResp struct {
	Items []struct {
		Hash        string `json:"hash"`
		Timestamp   string `json:"timestamp"`
		BlockNumber int64  `json:"block_number"`
		Value       string `json:"value"`
		Fee         struct {
			Value string `json:"value"`
		} `json:"fee"`
		From struct {
			Hash string `json:"hash"`
		} `json:"from"`
		To *struct {
			Hash       string `json:"hash"`
			IsContract bool   `json:"is_contract"`
		} `json:"to"`
	} `json:"items"`
}

// FetchAddressSummary fetches Ethereum address metadata and public labels from keyless Blockscout v2 API.
func (p *EthereumProvider) FetchAddressSummary(ctx context.Context, address string) (*model.AddressSummary, error) {
	url := fmt.Sprintf("%s/addresses/%s", strings.TrimRight(p.cfg.BlockscoutETHURL, "/"), strings.TrimSpace(address))

	var resp blockscoutAddressResp
	if err := p.getJSON(ctx, url, &resp); err != nil {
		// Return minimal summary if endpoint is temporarily unavailable
		return &model.AddressSummary{
			Address: address,
			Chain:   model.ChainEthereum,
		}, nil
	}

	rawBal := ParseDecimalString(resp.CoinBalance)
	balETH := FormatTokenUnits(rawBal, 18)

	var tags []string
	for _, pt := range resp.PublicTags {
		if pt.DisplayName != "" {
			tags = append(tags, pt.DisplayName)
		} else if pt.Label != "" {
			tags = append(tags, pt.Label)
		}
	}
	if resp.Metadata != nil {
		for _, mt := range resp.Metadata.Tags {
			if mt.Name != "" {
				tags = append(tags, mt.Name)
			}
		}
	}

	return &model.AddressSummary{
		Address:       address,
		Chain:         model.ChainEthereum,
		BalanceFloat:  balETH,
		IsContract:    resp.IsContract,
		ExplorerLabel: resp.Name,
		ExplorerTags:  tags,
	}, nil
}

// FetchOutboundTransactions retrieves both ERC-20 Smart Contract Transfer Event Logs (prioritizing USDT/USDC)
// and native ETH outbound transfers from the target address.
func (p *EthereumProvider) FetchOutboundTransactions(ctx context.Context, address string, limit int) ([]model.Transaction, error) {
	trimmed := strings.TrimSpace(address)
	var results []model.Transaction

	// 1. Fetch ERC-20 Smart Contract Transfer Event Logs via Blockscout v2
	tokenURL := fmt.Sprintf("%s/addresses/%s/token-transfers?type=ERC-20", strings.TrimRight(p.cfg.BlockscoutETHURL, "/"), trimmed)
	var tokenResp blockscoutTokenTransfersResp
	if err := p.getJSON(ctx, tokenURL, &tokenResp); err == nil {
		for _, item := range tokenResp.Items {
			if !strings.EqualFold(item.From.Hash, trimmed) {
				continue
			}

			txHash := item.Transaction
			if txHash == "" {
				txHash = item.TxHash
			}

			dec := 18
			if d, err := strconv.Atoi(item.Total.Decimals); err == nil && d > 0 {
				dec = d
			} else if d, err := strconv.Atoi(item.Token.Decimals); err == nil && d > 0 {
				dec = d
			}

			rawVal := ParseDecimalString(item.Total.Value)
			valFloat := FormatTokenUnits(rawVal, dec)

			symbol := item.Token.Symbol
			if symbol == "" {
				symbol = "ERC20"
			}
			contractAddr := item.Token.Address
			if contractAddr == "" {
				contractAddr = item.Token.Address2
			}

			ts := time.Now()
			if parsedTS, err := time.Parse(time.RFC3339, item.Timestamp); err == nil {
				ts = parsedTS
			}

			results = append(results, model.Transaction{
				TxID:          txHash,
				Chain:         model.ChainEthereum,
				BlockHeight:   item.BlockNumber,
				Timestamp:     ts,
				From:          item.From.Hash,
				To:            item.To.Hash,
				ValueFloat:    valFloat,
				AssetSymbol:   symbol,
				IsTokenTx:     true,
				ContractEvent: "Transfer(address,address,uint256)",
				TokenTransfers: []model.TokenTransfer{
					{
						ContractAddress: contractAddr,
						TokenSymbol:     symbol,
						TokenName:       item.Token.Name,
						Decimals:        dec,
						From:            item.From.Hash,
						To:              item.To.Hash,
						ValueRaw:        rawVal,
						ValueFloat:      valFloat,
					},
				},
			})

			if limit > 0 && len(results) >= limit {
				return results, nil
			}
		}
	} else {
		// Fallback: Direct eth_getLogs on Public Ethereum JSON-RPC for USDT ERC-20 Transfer events
		if rpcTxs, rpcErr := p.fetchUSDTLogsViaRPC(ctx, trimmed, limit); rpcErr == nil {
			results = append(results, rpcTxs...)
		}
	}

	// 2. Fetch Native ETH Outbound Transfers
	txURL := fmt.Sprintf("%s/addresses/%s/transactions?filter=from", strings.TrimRight(p.cfg.BlockscoutETHURL, "/"), trimmed)
	var txResp blockscoutTxsResp
	if err := p.getJSON(ctx, txURL, &txResp); err == nil {
		for _, item := range txResp.Items {
			if !strings.EqualFold(item.From.Hash, trimmed) || item.To == nil || item.To.Hash == "" {
				continue
			}

			rawVal := ParseDecimalString(item.Value)
			valETH := FormatTokenUnits(rawVal, 18)
			if valETH <= 0 {
				// Pure contract call with 0 native ETH (already covered by ERC-20 event logs above)
				continue
			}

			ts := time.Now()
			if parsedTS, err := time.Parse(time.RFC3339, item.Timestamp); err == nil {
				ts = parsedTS
			}

			results = append(results, model.Transaction{
				TxID:        item.Hash,
				Chain:       model.ChainEthereum,
				BlockHeight: item.BlockNumber,
				Timestamp:   ts,
				From:        item.From.Hash,
				To:          item.To.Hash,
				ValueFloat:  valETH,
				AssetSymbol: "ETH",
				IsTokenTx:   false,
			})

			if limit > 0 && len(results) >= limit {
				break
			}
		}
	}

	return results, nil
}

// ResolveHopTargets evaluates Ethereum native and ERC-20 event log transfers.
func (p *EthereumProvider) ResolveHopTargets(_ context.Context, tx model.Transaction, sourceAddress string) ([]model.HopTarget, error) {
	return ResolveAccountHopTargets(tx, sourceAddress, p.cfg), nil
}

// fetchUSDTLogsViaRPC queries Public Ethereum JSON-RPC eth_getLogs directly for Transfer events from address.
func (p *EthereumProvider) fetchUSDTLogsViaRPC(ctx context.Context, address string, limit int) ([]model.Transaction, error) {
	cleanAddr := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(address)), "0x")
	if len(cleanAddr) != 40 {
		return nil, fmt.Errorf("invalid EVM address length")
	}
	paddedSenderTopic := "0x000000000000000000000000" + cleanAddr

	// Get latest block number first
	latestBlockHex, err := p.rpcBlockNumber(ctx)
	if err != nil {
		return nil, err
	}
	latestBlock := ParseHexBigInt(latestBlockHex).Int64()
	fromBlock := latestBlock - 20000
	if fromBlock < 0 {
		fromBlock = 0
	}

	reqBody := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "eth_getLogs",
		"params": []any{
			map[string]any{
				"fromBlock": fmt.Sprintf("0x%x", fromBlock),
				"toBlock":   "latest",
				"address":   ETHUSDTContract,
				"topics": []any{
					TransferEventTopic0,
					paddedSenderTopic,
				},
			},
		},
	}

	var rpcResp struct {
		Result []struct {
			Address          string   `json:"address"`
			Topics           []string `json:"topics"`
			Data             string   `json:"data"`
			BlockNumber      string   `json:"blockNumber"`
			TransactionHash  string   `json:"transactionHash"`
		} `json:"result"`
	}

	if err := p.postRPC(ctx, reqBody, &rpcResp); err != nil {
		return nil, err
	}

	var out []model.Transaction
	for _, logItem := range rpcResp.Result {
		if len(logItem.Topics) < 3 {
			continue
		}
		fromAddr := DecodeIndexedAddress(logItem.Topics[1])
		toAddr := DecodeIndexedAddress(logItem.Topics[2])
		rawAmount := ParseHexBigInt(logItem.Data)
		valUSDT := FormatTokenUnits(rawAmount, 6)
		blockNum := ParseHexBigInt(logItem.BlockNumber).Int64()

		out = append(out, model.Transaction{
			TxID:          logItem.TransactionHash,
			Chain:         model.ChainEthereum,
			BlockHeight:   blockNum,
			Timestamp:     time.Now(),
			From:          fromAddr,
			To:            toAddr,
			ValueFloat:    valUSDT,
			AssetSymbol:   "USDT",
			IsTokenTx:     true,
			ContractEvent: "Transfer(address,address,uint256)",
			TokenTransfers: []model.TokenTransfer{
				{
					ContractAddress: ETHUSDTContract,
					TokenSymbol:     "USDT",
					TokenName:       "Tether USD",
					Decimals:        6,
					From:            fromAddr,
					To:              toAddr,
					ValueRaw:        rawAmount,
					ValueFloat:      valUSDT,
				},
			},
		})
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (p *EthereumProvider) rpcBlockNumber(ctx context.Context) (string, error) {
	reqBody := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "eth_blockNumber",
		"params":  []any{},
	}
	var resp struct {
		Result string `json:"result"`
	}
	if err := p.postRPC(ctx, reqBody, &resp); err != nil {
		return "", err
	}
	return resp.Result, nil
}

func (p *EthereumProvider) postRPC(ctx context.Context, payload any, target any) error {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.PublicETHRPCURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("RPC HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func (p *EthereumProvider) getJSON(ctx context.Context, url string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "crypto-tracker/1.0")

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

