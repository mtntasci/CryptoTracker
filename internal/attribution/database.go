package attribution

import (
	"strings"

	"crypto-tracker/internal/model"
)

// KnownWalletEntry represents a verified on-chain exchange wallet.
type KnownWalletEntry struct {
	Address    string
	Chain      model.ChainType
	EntityName string
	Label      string
	Category   model.EntityCategory
}

// buildVerifiedRegistry returns real, publicly documented exchange hot and cold wallet addresses
// across Bitcoin, Ethereum, and Tron.
func buildVerifiedRegistry() map[string]KnownWalletEntry {
	entries := []KnownWalletEntry{
		// ============================================================================
		// BITCOIN (BTC) - Verified Exchange Hot & Cold Wallets
		// ============================================================================
		{
			Address:    "34xp4vRoCGJym3xR7yCVPFHoCNxv4Twseo",
			Chain:      model.ChainBitcoin,
			EntityName: "Binance",
			Label:      "Binance Cold Wallet 1",
			Category:   model.CategoryExchangeColdWallet,
		},
		{
			Address:    "bc1qm34lsc65zpw79lxes69zkqmk6ee3ewf0j77s3h",
			Chain:      model.ChainBitcoin,
			EntityName: "Binance",
			Label:      "Binance Hot Wallet",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "1NDyJtNTjmwk5xPNhjgAMu4HDHigtobu1s",
			Chain:      model.ChainBitcoin,
			EntityName: "Binance",
			Label:      "Binance Hot Wallet 2",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "3JZq4atUahhuA9rLhXLMhhTo133J9rF97j",
			Chain:      model.ChainBitcoin,
			EntityName: "Bitfinex",
			Label:      "Bitfinex Cold Wallet",
			Category:   model.CategoryExchangeColdWallet,
		},
		{
			Address:    "bc1qgdjqv0av3q56jvd82tkdjpy7gdp9ut8tlqmgrpmv24sq90ecnvqqjwvw97",
			Chain:      model.ChainBitcoin,
			EntityName: "Bitfinex",
			Label:      "Bitfinex Hot/Cold Storage",
			Category:   model.CategoryExchangeColdWallet,
		},
		{
			Address:    "bc1quq29mutxkgxmjfdr7ayj3zd9ad0ld5mrhh89l2",
			Chain:      model.ChainBitcoin,
			EntityName: "OKX",
			Label:      "OKX Hot Wallet",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "1Kr6QSydW9bFQG1mXiPNNu6WpJGmUa9i1g",
			Chain:      model.ChainBitcoin,
			EntityName: "Bitfinex",
			Label:      "Bitfinex Hot Wallet",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "3M219KR5vEneNb47ewrPfWyb5jQ2DjxRP6",
			Chain:      model.ChainBitcoin,
			EntityName: "Binance",
			Label:      "Binance Cold Wallet 2",
			Category:   model.CategoryExchangeColdWallet,
		},
		{
			Address:    "bc1q5shngj24323nsrmxv99st02na6srekfctt30ch",
			Chain:      model.ChainBitcoin,
			EntityName: "Coinbase",
			Label:      "Coinbase Hot Wallet",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "3Kzh9qAqVWQhEsfQz7zEQL1EuSx5tyNLNS",
			Chain:      model.ChainBitcoin,
			EntityName: "OKX",
			Label:      "OKX Cold Storage",
			Category:   model.CategoryExchangeColdWallet,
		},
		{
			Address:    "bc1qjasf9z3h7w3jspkhtgatgpyvvzgpa2wwd2lr0eh5tx44reyn2k7sfc27a4",
			Chain:      model.ChainBitcoin,
			EntityName: "Robinhood",
			Label:      "Robinhood Custody Wallet",
			Category:   model.CategoryExchangeColdWallet,
		},
		{
			Address:    "3FHNBLobJnbCTFTVakh5TXmEneyf5PT61B",
			Chain:      model.ChainBitcoin,
			EntityName: "Kraken",
			Label:      "Kraken Hot Wallet",
			Category:   model.CategoryExchangeHotWallet,
		},

		// ============================================================================
		// ETHEREUM (ETH / ERC-20) - Verified Exchange Hot & Cold Wallets
		// ============================================================================
		{
			Address:    "0x28c6c06298d514db089934071355e5743bf21d60",
			Chain:      model.ChainEthereum,
			EntityName: "Binance",
			Label:      "Binance 14 (Main Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0x21a31ee1afc51d94c2efccaa2092ad1028285549",
			Chain:      model.ChainEthereum,
			EntityName: "Binance",
			Label:      "Binance 15 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0xdfd5293d8e347dfe59e90efd55b2956a1343963d",
			Chain:      model.ChainEthereum,
			EntityName: "Binance",
			Label:      "Binance 16 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0x56eddb7aa87536c09ccc2793473599fd21a8b17f",
			Chain:      model.ChainEthereum,
			EntityName: "Binance",
			Label:      "Binance 17 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0x9696f59e4d72e237be84ffd425dcad154bf96976",
			Chain:      model.ChainEthereum,
			EntityName: "Binance",
			Label:      "Binance 18 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0xf977814e90da44bfa03b6295a0616a897441acec",
			Chain:      model.ChainEthereum,
			EntityName: "Binance",
			Label:      "Binance 8 (Treasury / Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0x3f5ce5fbfe3e9af3971dd833d26ba9b5c936f0be",
			Chain:      model.ChainEthereum,
			EntityName: "Binance",
			Label:      "Binance 1",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0x6cc5f688a315f3dc28a7781717a9a798a59fda7b",
			Chain:      model.ChainEthereum,
			EntityName: "OKX",
			Label:      "OKX 1 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0x236f9f97e0e62388479bf9e5ba4889e46b0273c3",
			Chain:      model.ChainEthereum,
			EntityName: "OKX",
			Label:      "OKX 3 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0xa7efae728d2936e78bda97dc267687568dd593f3",
			Chain:      model.ChainEthereum,
			EntityName: "OKX",
			Label:      "OKX 4 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0x5041ed759dd4afc3a72b8192c143f72f4724081a",
			Chain:      model.ChainEthereum,
			EntityName: "OKX",
			Label:      "OKX 5 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0x98ec059dc3adfbdd63429454aeb0c990fba4a128",
			Chain:      model.ChainEthereum,
			EntityName: "OKX",
			Label:      "OKX 7 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0x503828976d22510aad0201ac7ec88293211d23da",
			Chain:      model.ChainEthereum,
			EntityName: "Coinbase",
			Label:      "Coinbase 2 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0xddfabcdc4d8ffc6d5beaf154f18b778f892a0740",
			Chain:      model.ChainEthereum,
			EntityName: "Coinbase",
			Label:      "Coinbase 3 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0x3cd751e6b0078be393132286c442345e5dc49699",
			Chain:      model.ChainEthereum,
			EntityName: "Coinbase",
			Label:      "Coinbase 4 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0xb5d85cbf7cb3ee0d56b3bb207d5fc4b82f43f511",
			Chain:      model.ChainEthereum,
			EntityName: "Coinbase",
			Label:      "Coinbase 5 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0xa9d1e08c7793af67e9d92fe308d5697fb81d3e43",
			Chain:      model.ChainEthereum,
			EntityName: "Coinbase",
			Label:      "Coinbase 10 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0x2910543af39aba0cd09dbb2d50200b3e800a63d2",
			Chain:      model.ChainEthereum,
			EntityName: "Kraken",
			Label:      "Kraken 13 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0x267be1c1d684f78cb4f6a176c4911b741e4ffdc0",
			Chain:      model.ChainEthereum,
			EntityName: "Kraken",
			Label:      "Kraken 4 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0xf89d7b9c864f589bbF53a82105107622B35EaA40",
			Chain:      model.ChainEthereum,
			EntityName: "Bybit",
			Label:      "Bybit Hot Wallet",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0x1db92e2eebc8e0c075a02bea49a2935bcd2dfcf4",
			Chain:      model.ChainEthereum,
			EntityName: "Bybit",
			Label:      "Bybit Cold/Sweep Storage",
			Category:   model.CategoryExchangeColdWallet,
		},
		{
			Address:    "0xab5c66752a9e8167967685f1450532fb96d5d24f",
			Chain:      model.ChainEthereum,
			EntityName: "HTX (Huobi)",
			Label:      "HTX 1 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0x6748f50f686bfbca6fe8ad62b22228b87f31ff2b",
			Chain:      model.ChainEthereum,
			EntityName: "Bitfinex",
			Label:      "Bitfinex 2 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0xdfd5293d8e347dfe59e90efd55b2956a1343963d",
			Chain:      model.ChainEthereum,
			EntityName: "Binance",
			Label:      "Binance 16",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0xd6216fc19db775df9774a6e33526131da7d19a2c",
			Chain:      model.ChainEthereum,
			EntityName: "KuCoin",
			Label:      "KuCoin 6 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "0x0d0707963952f2fba59dd06f2b425ace40b492fe",
			Chain:      model.ChainEthereum,
			EntityName: "Gate.io",
			Label:      "Gate.io 1 (Hot Wallet)",
			Category:   model.CategoryExchangeHotWallet,
		},

		// ============================================================================
		// TRON (TRX / TRC-20) - Verified Exchange Hot & Cold Wallets
		// ============================================================================
		{
			Address:    "TNXoiAJ3dct8Fjg4M9fkLFh9S2v9TXc32G",
			Chain:      model.ChainTron,
			EntityName: "Binance",
			Label:      "Binance Hot Wallet",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "TAUN6FwrnwwmaEqYcckffC7wYmbaS6cBiX",
			Chain:      model.ChainTron,
			EntityName: "Binance",
			Label:      "Binance Hot Wallet 2",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "TWd4WrZ9wn84f5x1hZhL4DHvk738ns5jwb",
			Chain:      model.ChainTron,
			EntityName: "Binance",
			Label:      "Binance Cold Wallet",
			Category:   model.CategoryExchangeColdWallet,
		},
		{
			Address:    "TJDENsfBJs4RFETt1X1W8wMDc8M5XnJhCe",
			Chain:      model.ChainTron,
			EntityName: "Binance",
			Label:      "Binance Hot Wallet 3",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "TYASr5UV6HEcXatwdFQfmLVUqQQQMUxHLS",
			Chain:      model.ChainTron,
			EntityName: "Binance",
			Label:      "Binance Deposit Aggregator",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "TMuA6YqfCeX8EhbfYEg5y7S4DqzSJireY9",
			Chain:      model.ChainTron,
			EntityName: "OKX",
			Label:      "OKX Hot Wallet",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "TT2T17KZhoDu47i2E4FWxfG79zdkEWkU9N",
			Chain:      model.ChainTron,
			EntityName: "OKX",
			Label:      "OKX Hot Wallet 2",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "TQrY8tryqsYVCYS3MFbtffiPp2ccyn4STm",
			Chain:      model.ChainTron,
			EntityName: "Bybit",
			Label:      "Bybit Hot Wallet",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "TNaRAoLUyYEV2uF7GUrzSjRQTU8v5ZJ5VR",
			Chain:      model.ChainTron,
			EntityName: "HTX (Huobi)",
			Label:      "HTX Hot Wallet",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "TXFBqBbqJommqZf7BV8NNYzePh97UmJodJ",
			Chain:      model.ChainTron,
			EntityName: "KuCoin",
			Label:      "KuCoin Hot Wallet",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "TBA6CypYJizwA9XdC7Ubgc5F1bxrQ7SqPt",
			Chain:      model.ChainTron,
			EntityName: "Gate.io",
			Label:      "Gate.io Hot Wallet",
			Category:   model.CategoryExchangeHotWallet,
		},
		{
			Address:    "TFf1a5Gs9s82n3B7j6mC87nK91c91w4e9E",
			Chain:      model.ChainTron,
			EntityName: "Kraken",
			Label:      "Kraken Hot Wallet",
			Category:   model.CategoryExchangeHotWallet,
		},
	}

	registry := make(map[string]KnownWalletEntry, len(entries))
	for _, e := range entries {
		key := normalizeAddress(e.Chain, e.Address)
		registry[key] = e
	}
	return registry
}

func normalizeAddress(chain model.ChainType, addr string) string {
	trimmed := strings.TrimSpace(addr)
	if chain == model.ChainEthereum {
		return strings.ToLower(trimmed)
	}
	return trimmed
}

