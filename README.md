# CryptoTracker (`crypto-tracker`)

A modular, production-ready CLI application in Go built with the [Charm Bubble Tea](https://github.com/charmbracelet/bubbletea) framework for interactive multi-chain cryptocurrency transaction tracing and exchange hot wallet attribution.

## Key Features

- **Multi-Chain Architecture**:
  - **UTXO Model (`BTC` - Bitcoin)**: Breadth-First Search (BFS) transaction graph traversal powered by keyless public Esplora APIs (`Mempool.space` primary, `Blockstream.info` automatic fallback).
    - **Dust Filtering**: Automatically filters out dust outputs (`<= 1000 sats`) and spam/dusting attacks.
    - **Change Address Heuristics**: Identifies and skips change outputs using Script-Type Matching, Round-Number Payment detection, Common-Input Ownership (self/co-spend filtering), and Optimal Change (`UIH1`).
  - **Account & Token Model (`ETH`, `TRON`, `USDT ERC-20 / TRC-20`)**:
    - Monitors `Transfer(address indexed from, address indexed to, uint256 value)` smart contract event logs (`0xddf252ad...`) and direct native transfers without requiring paid API keys (`Blockscout v2`, `Public Ethereum JSON-RPC`, `TronGrid v1`, and `Tronscan Public API`).
- **Attribution Engine**:
  - Combines a verified internal registry of major Exchange Hot & Cold Wallets (**Binance, OKX, Coinbase, Kraken, Bybit, Bitfinex, HTX, KuCoin, Gate.io, Robinhood**) with live keyless public tag resolution (`Blockscout Public Tags` & `Tronscan Account Tags`) and behavioral deposit-sweeping consolidation heuristics.
- **Interactive Terminal UI (Bubble Tea)**:
  - **Initialization View**: Chain auto-detection, network switcher (`Tab`), and live parameter display.
  - **Animated BFS Tracing View**: Real-time hop-by-hop visualization, depth progress indicator (up to 5 hops), and live heuristic filter feed.
  - **Final Attribution Report**: Highlights confirmed Exchange Hot Wallet hits or lists active frontier addresses for ongoing monitoring.

## Project Structure

```text
CryptoTracker/
├── cmd/
│   └── crypto-tracker/
│       └── main.go                  # CLI & Bubbletea TUI entrypoint
├── internal/
│   ├── attribution/
│   │   ├── database.go              # Verified on-chain Exchange Hot/Cold Wallet registry
│   │   ├── matcher.go               # Multi-layer attribution & behavioral sweep engine
│   │   └── online.go                # Keyless Blockscout & Tronscan public label resolvers
│   ├── chain/
│   │   ├── provider.go              # Modular ChainProvider interface
│   │   ├── utxo/
│   │   │   ├── bitcoin.go           # Bitcoin Mempool.space & Blockstream Esplora provider
│   │   │   ├── heuristics.go        # UTXO dust filter & change address heuristics
│   │   │   └── btc_test.go          # Heuristic unit tests
│   │   └── account/
│   │       ├── ethereum.go          # Ethereum & ERC-20 event log provider (Blockscout + RPC)
│   │       ├── tron.go              # Tron & TRC-20 event log provider (TronGrid + Tronscan)
│   │       └── parser.go            # ERC-20/TRC-20 Transfer topic decoder & dust filters
│   ├── config/
│   │   └── config.go                # Configuration & keyless endpoint defaults (MaxDepth = 5)
│   ├── model/
│   │   ├── attribution.go           # Attribution & entity category models
│   │   ├── chain.go                 # Chain metadata & address auto-detection
│   │   ├── trace.go                 # BFS hop steps, skipped outputs, and trace report models
│   │   └── transaction.go           # Normalized UTXO & Account/Token transaction models
│   ├── tracer/
│   │   ├── engine.go                # Core BFS traversal engine (depth=5, cycle prevention)
│   │   └── events.go                # Real-time event stream definitions for Bubbletea TUI
│   └── ui/
│       ├── app.go                   # Root Bubbletea state machine (Input -> Tracing -> Result)
│       └── views/
│           ├── input.go             # Initialization & address prompt view
│           ├── result.go            # Final attribution report view
│           ├── styles.go            # Lip Gloss styling palette
│           └── tracing.go           # Animated step-by-step hop visualization view
├── go.mod
└── go.sum
```

## Build & Run

```bash
# Build the binary
go build -o crypto-tracker ./cmd/crypto-tracker

# Launch Interactive Bubbletea TUI
./crypto-tracker

# Launch TUI directly with a starting address (auto-detects BTC, ETH, or TRON)
./crypto-tracker -address <WALLET_ADDRESS> -depth 5

# Run in headless streaming CLI mode
./crypto-tracker -address <WALLET_ADDRESS> -depth 5 -headless
```
