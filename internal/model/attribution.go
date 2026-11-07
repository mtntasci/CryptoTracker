package model

// EntityCategory classifies the role of an attributed blockchain address.
type EntityCategory string

const (
	CategoryExchangeHotWallet  EntityCategory = "Exchange Hot Wallet"
	CategoryExchangeColdWallet EntityCategory = "Exchange Cold Wallet"
	CategoryExchangeDeposit    EntityCategory = "Deposit / Sweep Intermediary"
	CategoryBridgeOrDeFi       EntityCategory = "Bridge / DeFi Protocol"
	CategoryUnknown            EntityCategory = "Unknown Wallet"
)

// AttributionInfo holds identification details for a wallet address.
type AttributionInfo struct {
	Address     string
	Chain       ChainType
	EntityName  string         // e.g., "Binance", "OKX", "Coinbase", "Kraken", "Bybit", "HTX", "Bitfinex"
	WalletLabel string         // e.g., "Binance Hot Wallet 14", "OKX Hot Wallet"
	Category    EntityCategory
	Confidence  float64        // 0.0 to 1.0
	Source      string         // e.g., "Internal Hot Wallet Registry", "Blockscout Public Tag", "Tronscan Public Label", "Behavioral Sweep Heuristic"
	IsTerminal  bool           // True if tracing along this branch should stop (reached exchange hot/cold wallet)
}

