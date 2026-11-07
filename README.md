# CryptoTracker

**Trace crypto. Follow the flow. Find the destination.**

CryptoTracker is a fast, stateless, cross-platform terminal application written in Go. It traces cryptocurrency transfers hop by hop across Bitcoin, Ethereum, and TRON, helping identify destination centralized exchanges (CEX) by following funds through intermediate wallets and exchange deposit sweeps.

No local database. No full-node synchronization. Just a starting wallet address and an interactive view of where the funds go.

[![Go](https://img.shields.io/badge/Go-1.0%2B-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE.md)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-blue)](#installation)

---

## Overview

Following cryptocurrency across a public blockchain is relatively straightforward when looking at individual transactions. Understanding where the funds ultimately end up is much harder.

Bitcoin's UTXO model introduces change outputs and complex transaction graphs. Ethereum and TRON add token transfers, smart contract events, and intermediate addresses. Centralized exchanges make tracking even more challenging by assigning unique deposit addresses to users and periodically consolidating those funds into their main wallets.

CryptoTracker addresses this problem by traversing transaction flows, applying chain-specific heuristics, and identifying known exchange hot wallets.

The result is an interactive terminal experience that makes complex on-chain transfer paths easier to understand.

## Features

### Multi-chain tracking

Track native cryptocurrency and token transfers across multiple blockchain networks.

| Network | Asset / Standard | Tracking |
|---|---|---|
| Bitcoin | BTC | UTXO transaction graph |
| Ethereum | ETH | Native transfers |
| Ethereum | ERC-20 / USDT | `Transfer` event logs |
| TRON | TRX | Native transfers |
| TRON | TRC-20 / USDT | `Transfer` event logs |

### Bitcoin UTXO graph traversal

Bitcoin transactions do not behave like conventional account-based transfers. A single transaction can contain multiple inputs and outputs, making it difficult to distinguish actual fund movement from change.

CryptoTracker uses Breadth-First Search (BFS) to traverse the transaction graph and applies heuristics to reduce irrelevant branches.

Current heuristic approaches include:

- Change-output detection.
- Script-type matching.
- Round-number analysis.
- Address reuse checks.
- Dust transaction filtering.

The goal is to prioritize meaningful fund movements instead of following every output indiscriminately.

### Exchange deposit sweep detection

A deposit address does not necessarily reveal which exchange controls it.

Exchanges commonly generate individual deposit addresses and later consolidate incoming funds into shared hot wallets.

CryptoTracker follows intermediate transfers and looks for these consolidation patterns, typically within 1–3 hops, to identify potential destination exchanges.

Known exchange labels include:

- Binance
- OKX
- Coinbase
- Kraken
- Paribu
- Bybit

The exchange label dataset can be expanded as new wallet information becomes available.

### Interactive terminal interface

CryptoTracker uses [Charm Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss) to provide an interactive terminal experience.

Instead of dumping raw blockchain data into a terminal, it presents transfer paths as a visual, step-by-step flow.

- Animated transaction traversal.
- Hop-by-hop transfer visualization.
- Branch analysis.
- Intermediate wallet inspection.
- Exchange detection and confidence indicators.

### Stateless architecture

CryptoTracker does not require a local database or a synchronized blockchain node.

Its tracking process runs in memory, avoiding the operational overhead of maintaining a persistent local transaction database.

This makes it lightweight and convenient to run across different environments.

### Cross-platform

CryptoTracker is written in Go and designed to build as a standalone executable for:

- macOS
- Linux
- Windows

No language runtime or local database installation is required to run a compiled binary.

---

## How It Works

At a high level, CryptoTracker follows this process:

1. **Starting address** — The user provides a wallet address.
2. **Transaction discovery** — The application retrieves relevant on-chain transaction information.
3. **Graph traversal** — Transfers are followed hop by hop using chain-specific tracking logic.
4. **Heuristic analysis** — Change outputs, dust transactions, and irrelevant branches are filtered where applicable.
5. **Intermediate wallet tracking** — Unlabeled deposit addresses and subsequent consolidation movements are examined.
6. **Exchange identification** — Known exchange hot-wallet labels are matched against the observed transfer path.
7. **Terminal visualization** — The resulting transfer path and detection information are presented interactively.

---

## Technology Stack

| Component | Technology |
|---|---|
| Programming language | Go |
| Terminal UI | Bubble Tea |
| Terminal styling | Lip Gloss |
| Bitcoin tracking | UTXO graph traversal / BFS |
| Ethereum tracking | Native transfers / ERC-20 events |
| TRON tracking | Native transfers / TRC-20 events |
| Architecture | Stateless, in-memory |

---

## Installation

### Download a release

Download the appropriate precompiled binary for your operating system from the project's GitHub Releases page.

### Build from source

Make sure [Go](https://go.dev/dl/) is installed.

Clone the repository:

```bash
git clone https://github.com/mtntasci/crypto-tracker.git
cd crypto-tracker
```

Build:

```bash
go build -o cryptotracker .
```

Run the resulting executable:

```bash
./cryptotracker
```

On Windows, run the generated executable:

```powershell
.\cryptotracker.exe
```

*Note: Runtime configuration and command-line options should be documented here as they are finalized.*

---

## Project Goals

CryptoTracker aims to make multi-chain fund tracing accessible directly from the terminal.

The project focuses on:

- Accurate transaction-flow analysis.
- Chain-specific tracking and heuristics.
- Exchange deposit consolidation detection.
- Minimal operational requirements.
- A portable, efficient Go implementation.
- A clear and interactive terminal experience.

---

## Limitations and Accuracy

CryptoTracker uses publicly observable blockchain activity and heuristic analysis to identify potential fund destinations.

A detected exchange hot wallet does not, by itself, prove the identity of a wallet owner or the identity of an individual exchange customer.

Heuristic results may be affected by:
- Unlabeled or outdated wallet addresses.
- Complex transaction patterns.
- Exchange wallet infrastructure changes.
- Multiple possible interpretations of UTXO transactions.
- Incomplete or delayed blockchain data.

Detection results should therefore be treated as analytical indicators, not definitive attribution or forensic proof.

---

## Contributing

Contributions, bug reports, heuristic improvements, and additional exchange wallet labels are welcome.

Areas where contributions can help:

- New exchange wallet labels.
- Improved Bitcoin change-output detection.
- Better dust and spam filtering.
- Additional blockchain and token support.
- Transaction graph analysis improvements.
- Performance and terminal UI improvements.

Feel free to open an issue or submit a pull request.

---

## License

CryptoTracker is released under the [MIT License](LICENSE.md).

---

**Built with Go. Designed for the terminal. Following funds across chains.**