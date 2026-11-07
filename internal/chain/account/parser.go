package account

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	"crypto-tracker/internal/config"
	"crypto-tracker/internal/model"
)

// TransferEventTopic0 is the Keccak-256 signature of Transfer(address,address,uint256)
// used across both Ethereum ERC-20 and Tron TRC-20 smart contracts.
const TransferEventTopic0 = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

// Well-known stablecoin contract addresses
const (
	ETHUSDTContract  = "0xdac17f958d2ee523a2206206994597c13d831ec7"
	ETHUSDCContract  = "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"
	TRONUSDTContract = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
)

// FormatTokenUnits converts a raw integer token balance/transfer value into float64 using decimals.
func FormatTokenUnits(raw *big.Int, decimals int) float64 {
	if raw == nil || raw.Sign() == 0 {
		return 0
	}
	if decimals <= 0 {
		decimals = 18
	}
	f := new(big.Float).SetInt(raw)
	divisor := new(big.Float).SetFloat64(math.Pow10(decimals))
	res := new(big.Float).Quo(f, divisor)
	val, _ := res.Float64()
	return val
}

// ParseDecimalString parses a decimal string (e.g. "1000000000") into *big.Int.
func ParseDecimalString(s string) *big.Int {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return big.NewInt(0)
	}
	v, ok := new(big.Int).SetString(trimmed, 10)
	if !ok {
		return big.NewInt(0)
	}
	return v
}

// ParseHexBigInt parses a 0x-prefixed hex string into *big.Int.
func ParseHexBigInt(hexStr string) *big.Int {
	clean := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(hexStr)), "0x")
	if clean == "" {
		return big.NewInt(0)
	}
	v, ok := new(big.Int).SetString(clean, 16)
	if !ok {
		return big.NewInt(0)
	}
	return v
}

// DecodeIndexedAddress extracts a 20-byte EVM address from a 32-byte event log topic.
func DecodeIndexedAddress(topic string) string {
	clean := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(topic)), "0x")
	if len(clean) >= 40 {
		return "0x" + clean[len(clean)-40:]
	}
	return "0x" + clean
}

// ResolveAccountHopTargets converts an Account/Token model transaction into HopTargets,
// filtering out zero-value/dust transfers and flagging smart contract Transfer event sweeps.
func ResolveAccountHopTargets(tx model.Transaction, sourceAddress string, cfg *config.Config) []model.HopTarget {
	var targets []model.HopTarget

	// 1. Process Smart Contract Token Transfer Event Logs (ERC-20 / TRC-20)
	if len(tx.TokenTransfers) > 0 {
		for _, tt := range tx.TokenTransfers {
			if !strings.EqualFold(tt.From, sourceAddress) {
				continue
			}
			if strings.EqualFold(tt.To, sourceAddress) {
				targets = append(targets, model.HopTarget{
					TxID:            tx.TxID,
					Timestamp:       tx.Timestamp,
					FromAddress:     tt.From,
					ToAddress:       tt.To,
					AmountFloat:     tt.ValueFloat,
					AssetSymbol:     tt.TokenSymbol,
					IsSkipped:       true,
					Reason:          model.ReasonSelfTransfer,
					HeuristicDetail: "Token transfer back to same address",
					IsTokenTransfer: true,
					ContractAddress: tt.ContractAddress,
				})
				continue
			}

			if tt.ValueFloat < cfg.TokenDustThreshold {
				targets = append(targets, model.HopTarget{
					TxID:            tx.TxID,
					Timestamp:       tx.Timestamp,
					FromAddress:     tt.From,
					ToAddress:       tt.To,
					AmountFloat:     tt.ValueFloat,
					AssetSymbol:     tt.TokenSymbol,
					IsSkipped:       true,
					Reason:          model.ReasonDustFiltered,
					HeuristicDetail: fmt.Sprintf("Token transfer %.4f %s < dust limit (%.2f) (zero-value/address-poisoning filter)", tt.ValueFloat, tt.TokenSymbol, cfg.TokenDustThreshold),
					IsTokenTransfer: true,
					ContractAddress: tt.ContractAddress,
				})
				continue
			}

			targets = append(targets, model.HopTarget{
				TxID:            tx.TxID,
				Timestamp:       tx.Timestamp,
				FromAddress:     tt.From,
				ToAddress:       tt.To,
				AmountFloat:     tt.ValueFloat,
				AssetSymbol:     tt.TokenSymbol,
				IsSkipped:       false,
				Reason:          model.ReasonContractEvent,
				HeuristicDetail: fmt.Sprintf("Contract Event Log: Transfer(from, to, %.2f %s)", tt.ValueFloat, tt.TokenSymbol),
				IsTokenTransfer: true,
				ContractAddress: tt.ContractAddress,
			})
		}
		return targets
	}

	// 2. Process Direct Native Balance Transfer (ETH / TRX)
	if !strings.EqualFold(tx.From, sourceAddress) || tx.To == "" {
		return nil
	}

	if strings.EqualFold(tx.To, sourceAddress) {
		return []model.HopTarget{
			{
				TxID:            tx.TxID,
				Timestamp:       tx.Timestamp,
				FromAddress:     tx.From,
				ToAddress:       tx.To,
				AmountFloat:     tx.ValueFloat,
				AssetSymbol:     tx.AssetSymbol,
				IsSkipped:       true,
				Reason:          model.ReasonSelfTransfer,
				HeuristicDetail: "Native self-transfer",
			},
		}
	}

	dustLimit := cfg.ETHDustThreshold
	if tx.Chain == model.ChainTron {
		dustLimit = cfg.TRXDustThreshold
	}

	if tx.ValueFloat < dustLimit {
		return []model.HopTarget{
			{
				TxID:            tx.TxID,
				Timestamp:       tx.Timestamp,
				FromAddress:     tx.From,
				ToAddress:       tx.To,
				AmountFloat:     tx.ValueFloat,
				AssetSymbol:     tx.AssetSymbol,
				IsSkipped:       true,
				Reason:          model.ReasonDustFiltered,
				HeuristicDetail: fmt.Sprintf("Native transfer %.6f %s < dust threshold (%.6f)", tx.ValueFloat, tx.AssetSymbol, dustLimit),
			},
		}
	}

	return []model.HopTarget{
		{
			TxID:            tx.TxID,
			Timestamp:       tx.Timestamp,
			FromAddress:     tx.From,
			ToAddress:       tx.To,
			AmountFloat:     tx.ValueFloat,
			AssetSymbol:     tx.AssetSymbol,
			IsSkipped:       false,
			Reason:          model.ReasonOutboundTransfer,
			HeuristicDetail: fmt.Sprintf("Direct account balance transfer (%.4f %s)", tx.ValueFloat, tx.AssetSymbol),
		},
	}
}

