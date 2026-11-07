package views

import (
	"fmt"
	"strings"

	"crypto-tracker/internal/config"
	"crypto-tracker/internal/model"
)

// InputViewProps holds state needed to render the initialization and address input screen.
type InputViewProps struct {
	SelectedChain model.ChainType
	TextInputView string
	ErrorMessage  string
	Config        *config.Config
}

// RenderInputView renders the initialization and wallet input screen with the giant CRYPTO TRACKER banner.
func RenderInputView(p InputViewProps) string {
	var b strings.Builder

	// Big "CRYPTO TRACKER" Banner Only
	b.WriteString(RenderBigBanner())
	b.WriteString("\n")

	// Chain selector bar
	var chainTabs []string
	for _, meta := range model.AllChains() {
		label := fmt.Sprintf(" %s (%s) ", meta.Type, meta.Architecture)
		if meta.Type == p.SelectedChain {
			chainTabs = append(chainTabs, BadgeSelectedChainStyle.Render("●"+label))
		} else {
			chainTabs = append(chainTabs, BadgeChainStyle.Render("○"+label))
		}
	}

	selectedMeta := model.GetChainMetadata(p.SelectedChain)

	var configLines strings.Builder
	configLines.WriteString(SubHeaderStyle.Render("Engine Initialization & Active Parameters") + "\n\n")
	configLines.WriteString(fmt.Sprintf("  %s : %s\n", BoldWhiteStyle.Render("Selected Network"), strings.Join(chainTabs, "  ")))
	configLines.WriteString(fmt.Sprintf("  %s : %s\n", BoldWhiteStyle.Render("Chain Model     "), TextStyle.Render(fmt.Sprintf("%s (%s)", selectedMeta.Name, selectedMeta.Architecture))))
	configLines.WriteString(fmt.Sprintf("  %s : %s\n", BoldWhiteStyle.Render("BFS Max Depth   "), TextStyle.Render(fmt.Sprintf("%d Stages (Max %d outbound branches/node)", p.Config.MaxDepth, p.Config.MaxBranches))))

	switch p.SelectedChain {
	case model.ChainBitcoin:
		configLines.WriteString(fmt.Sprintf("  %s : %s\n", BoldWhiteStyle.Render("UTXO Heuristics "), TextStyle.Render(fmt.Sprintf("Dust Filter (<= %d sats), Script-Type, Round-Amount & Optimal Change (UIH1)", p.Config.BTCDustThresholdSats))))
		configLines.WriteString(fmt.Sprintf("  %s : %s\n", BoldWhiteStyle.Render("Keyless API     "), AddressStyle.Render(fmt.Sprintf("%s (Fallback: %s)", p.Config.MempoolBaseURL, p.Config.BlockstreamBaseURL))))
	case model.ChainEthereum:
		configLines.WriteString(fmt.Sprintf("  %s : %s\n", BoldWhiteStyle.Render("Account Engine  "), TextStyle.Render("ERC-20 Smart Contract Transfer Event Log Monitor & Native ETH Sweep Detector")))
		configLines.WriteString(fmt.Sprintf("  %s : %s\n", BoldWhiteStyle.Render("Keyless API     "), AddressStyle.Render(fmt.Sprintf("%s & Public JSON-RPC (%s)", p.Config.BlockscoutETHURL, p.Config.PublicETHRPCURL))))
	case model.ChainTron:
		configLines.WriteString(fmt.Sprintf("  %s : %s\n", BoldWhiteStyle.Render("Account Engine  "), TextStyle.Render("TRC-20 Smart Contract Transfer Event Log Monitor (USDT) & Native TRX")))
		configLines.WriteString(fmt.Sprintf("  %s : %s\n", BoldWhiteStyle.Render("Keyless API     "), AddressStyle.Render(fmt.Sprintf("%s & %s", p.Config.TronGridBaseURL, p.Config.TronscanBaseURL))))
	}

	b.WriteString(PanelStyle.Render(configLines.String()))
	b.WriteString("\n")

	// Address Input Box
	var inputBox strings.Builder
	inputBox.WriteString(SubHeaderStyle.Render("Target Starting Wallet Address") + "\n")
	inputBox.WriteString(TextStyle.Render("Paste a Bitcoin (1..., 3..., bc1...), Ethereum (0x...), or Tron (T...) address:") + "\n\n")
	inputBox.WriteString("  " + p.TextInputView + "\n")

	if p.ErrorMessage != "" {
		inputBox.WriteString("\n  " + ErrorStyle.Render("✖ "+p.ErrorMessage) + "\n")
	}

	b.WriteString(ActivePanelStyle.Render(inputBox.String()))
	b.WriteString("\n")
	b.WriteString(BoldWhiteStyle.Render("  [Tab] Switch Chain   [Enter] Start Live BFS Trace   [Ctrl+C / Esc] Quit"))
	b.WriteString("\n")

	return b.String()
}
