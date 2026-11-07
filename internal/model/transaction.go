package model

import (
	"math/big"
	"time"
)

// ScriptType identifies Bitcoin/UTXO locking script types for change heuristics.
type ScriptType string

const (
	ScriptP2PKH   ScriptType = "p2pkh"      // 1...
	ScriptP2SH    ScriptType = "p2sh"       // 3...
	ScriptP2WPKH  ScriptType = "v0_p2wpkh"  // bc1q... (42 chars)
	ScriptP2WSH   ScriptType = "v0_p2wsh"   // bc1q... (62 chars)
	ScriptP2TR    ScriptType = "v1_p2tr"    // bc1p... (Taproot)
	ScriptUnknown ScriptType = "unknown"
)

// TxInput represents a single input in a UTXO or Account transaction.
type TxInput struct {
	PrevTxID   string
	Vout       uint32
	Address    string
	ValueRaw   *big.Int // In satoshis / wei / sun
	ValueFloat float64  // Human-readable unit (BTC, ETH, TRX, USDT)
	ScriptType ScriptType
}

// TxOutput represents a single output in a UTXO or Account transaction.
type TxOutput struct {
	Index      uint32
	Address    string
	ValueRaw   *big.Int
	ValueFloat float64
	ScriptType ScriptType
	IsSpent    bool
}

// TokenTransfer represents a smart contract event log transfer (e.g. ERC-20 / TRC-20 Transfer event).
type TokenTransfer struct {
	ContractAddress string
	TokenSymbol     string
	TokenName       string
	Decimals        int
	From            string
	To              string
	ValueRaw        *big.Int
	ValueFloat      float64
	LogIndex        uint
}

// Transaction is a normalized blockchain transaction supporting both UTXO and Account/Token models.
type Transaction struct {
	TxID           string
	Chain          ChainType
	BlockHeight    int64
	Timestamp      time.Time
	FeeFloat       float64
	Inputs         []TxInput
	Outputs        []TxOutput
	TokenTransfers []TokenTransfer

	// Account-model direct fields
	From          string
	To            string
	ValueFloat    float64
	AssetSymbol   string
	IsTokenTx     bool
	ContractEvent string
}

// AddressSummary contains on-chain metadata about an address used in heuristics and attribution.
type AddressSummary struct {
	Address       string
	Chain         ChainType
	TxCount       int
	FundedCount   int
	SpentCount    int
	BalanceFloat  float64
	IsContract    bool
	ExplorerLabel string // Label returned directly by public indexer (e.g. Blockscout / Tronscan)
	ExplorerTags  []string
}

