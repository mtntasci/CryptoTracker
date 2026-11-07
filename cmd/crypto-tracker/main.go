package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"crypto-tracker/internal/attribution"
	"crypto-tracker/internal/chain"
	"crypto-tracker/internal/chain/account"
	"crypto-tracker/internal/chain/utxo"
	"crypto-tracker/internal/config"
	"crypto-tracker/internal/model"
	"crypto-tracker/internal/tracer"
	"crypto-tracker/internal/ui"
	"crypto-tracker/internal/ui/views"
)

func main() {
	cfg := config.DefaultConfig()

	var (
		addrFlag     = flag.String("address", "", "Starting wallet address to trace (BTC, ETH, or TRON)")
		chainFlag    = flag.String("chain", "", "Chain override: BTC, ETH, or TRON (auto-detected if omitted)")
		depthFlag    = flag.Int("depth", cfg.MaxDepth, "Maximum BFS hop depth (default: 5)")
		branchesFlag = flag.Int("branches", cfg.MaxBranches, "Maximum outbound branches followed per node")
		headlessFlag = flag.Bool("headless", false, "Run in non-interactive streaming CLI mode instead of Bubbletea TUI")
	)
	flag.Parse()

	if *depthFlag > 0 && *depthFlag <= 15 {
		cfg.MaxDepth = *depthFlag
	}
	if *branchesFlag > 0 && *branchesFlag <= 20 {
		cfg.MaxBranches = *branchesFlag
	}

	address := strings.TrimSpace(*addrFlag)
	if address == "" && flag.NArg() > 0 {
		address = strings.TrimSpace(flag.Arg(0))
	}

	var selectedChain model.ChainType
	if *chainFlag != "" {
		switch strings.ToUpper(strings.TrimSpace(*chainFlag)) {
		case "BTC", "BITCOIN":
			selectedChain = model.ChainBitcoin
		case "ETH", "ETHEREUM", "ERC20":
			selectedChain = model.ChainEthereum
		case "TRON", "TRX", "TRC20":
			selectedChain = model.ChainTron
		default:
			fmt.Fprintf(os.Stderr, "Unsupported chain %q (use BTC, ETH, or TRON)\n", *chainFlag)
			os.Exit(1)
		}
	} else if address != "" {
		detected, err := model.DetectChainFromAddress(address)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		selectedChain = detected
	}

	if *headlessFlag {
		if address == "" {
			fmt.Fprintln(os.Stderr, "Error: -address is required when running with -headless")
			os.Exit(1)
		}
		runHeadless(cfg, address, selectedChain)
		return
	}

	app := ui.NewAppModel(cfg, address, selectedChain)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running CryptoTracker UI: %v\n", err)
		os.Exit(1)
	}
}

func runHeadless(cfg *config.Config, address string, chainType model.ChainType) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var provider chain.Provider
	switch chainType {
	case model.ChainBitcoin:
		provider = utxo.NewBitcoinProvider(cfg)
	case model.ChainEthereum:
		provider = account.NewEthereumProvider(cfg)
	case model.ChainTron:
		provider = account.NewTronProvider(cfg)
	default:
		provider = utxo.NewBitcoinProvider(cfg)
	}

	attrEngine := attribution.NewEngine(cfg)
	engine := tracer.NewEngine(cfg, provider, attrEngine)

	eventCh := make(chan tracer.TraceEvent, 64)
	go func() {
		_, _ = engine.Trace(ctx, address, eventCh)
	}()

	var finalRes *model.TraceResult
	for ev := range eventCh {
		switch ev.Type {
		case tracer.EventInit, tracer.EventScanningNode, tracer.EventHopDiscovered, tracer.EventExchangeHit:
			fmt.Printf("[%s] %s\n", ev.Type, ev.Message)
		case tracer.EventOutputSkipped:
			fmt.Printf("  [FILTER] %s\n", ev.Message)
		case tracer.EventError:
			fmt.Printf("  [WARN] %s\n", ev.Message)
		case tracer.EventComplete:
			finalRes = ev.Result
		}
	}

	if finalRes != nil {
		fmt.Println()
		fmt.Println(views.RenderResultView(views.ResultViewProps{
			Result:       finalRes,
			ScrollOffset: 0,
		}))
	}
}

