package views

import (
	"fmt"
	"strings"
	"time"

	"crypto-tracker/internal/model"
)

// ResultViewProps holds the completed trace result and scroll offset for the UI.
type ResultViewProps struct {
	Result       *model.TraceResult
	ScrollOffset int
}

// RenderResultView renders the final attribution report and transaction trace tree.
func RenderResultView(p ResultViewProps) string {
	var b strings.Builder
	res := p.Result
	if res == nil {
		return "No trace result available."
	}

	meta := model.GetChainMetadata(res.Chain)
	duration := res.EndTime.Sub(res.StartTime).Round(time.Millisecond)

	b.WriteString(HeaderStyle.Render(fmt.Sprintf(" CRYPTO-TRACKER // FINAL ATTRIBUTION REPORT (%s) ", meta.Name)))
	b.WriteString("\n\n")

	// 1. Attribution Outcome Banner
	if len(res.TerminalHits) > 0 {
		var hitBanner strings.Builder
		hitBanner.WriteString(BadgeExchangeHitStyle.Render(" ✔ DESTINATION EXCHANGE HOT WALLET IDENTIFIED ") + "\n\n")
		for idx, hit := range res.TerminalHits {
			attr := hit.Attribution
			hitBanner.WriteString(fmt.Sprintf("  %d. Exchange Entity   : %s (%s)\n", idx+1, SubHeaderStyle.Render(attr.EntityName), attr.WalletLabel))
			hitBanner.WriteString(fmt.Sprintf("     Wallet Category   : %s (Confidence: %.0f%% | Source: %s)\n", attr.Category, attr.Confidence*100, attr.Source))
			hitBanner.WriteString(fmt.Sprintf("     Hot Wallet Addr   : %s\n", AddressStyle.Render(hit.ToAddress)))
			hitBanner.WriteString(fmt.Sprintf("     Sweep / Transfer  : %s at Hop Depth %d (Tx: %s)\n",
				AmountStyle.Render(fmt.Sprintf("%.4f %s", hit.AmountFloat, hit.AssetSymbol)),
				hit.Depth,
				FormatShortAddr(hit.TxID),
			))
			if idx < len(res.TerminalHits)-1 {
				hitBanner.WriteString("\n")
			}
		}
		b.WriteString(SuccessBannerStyle.Render(hitBanner.String()))
	} else {
		var ongoingBanner strings.Builder
		ongoingBanner.WriteString(BadgeSkippedStyle.Render(" ◉ ONGOING TRACKING // NO KNOWN EXCHANGE HOT WALLET REACHED WITHIN MAX DEPTH ") + "\n\n")
		ongoingBanner.WriteString(fmt.Sprintf("  Funds have not entered a known exchange hot wallet within %d hops.\n", res.MaxDepthReached))
		if len(res.ActiveBranches) > 0 {
			ongoingBanner.WriteString("  Current Terminal / Unspent Frontier Wallets for Ongoing Monitoring:\n")
			maxShow := len(res.ActiveBranches)
			if maxShow > 5 {
				maxShow = 5
			}
			for i := 0; i < maxShow; i++ {
				ongoingBanner.WriteString(fmt.Sprintf("    • %s\n", AddressStyle.Render(res.ActiveBranches[i])))
			}
		}
		b.WriteString(WarningBannerStyle.Render(ongoingBanner.String()))
	}
	b.WriteString("\n")

	// 2. Session Telemetry Summary
	var summary strings.Builder
	summary.WriteString(SubHeaderStyle.Render("Trace Session Telemetry") + "\n\n")
	summary.WriteString(fmt.Sprintf("  Root Wallet          : %s\n", AddressStyle.Render(res.RootAddress)))
	summary.WriteString(fmt.Sprintf("  Chain & Architecture : %s (%s)  |  Elapsed Time: %s\n", meta.Name, meta.Architecture, duration))
	summary.WriteString(fmt.Sprintf("  Traversal Stats      : Max Depth Reached: %d  |  Wallets Visited: %d  |  Transactions Analyzed: %d\n",
		res.MaxDepthReached, res.TotalVisited, res.TotalTxAnalyzed))
	summary.WriteString(fmt.Sprintf("  Heuristics Applied   : %d Valid Outbound Hops  |  %d Filtered Outputs (Dust / Change / Self)\n",
		len(res.Steps), len(res.SkippedOutputs)))

	b.WriteString(PanelStyle.Render(summary.String()))
	b.WriteString("\n")

	// 3. Complete Step-by-Step Hop Breakdown
	var stepsBox strings.Builder
	stepsBox.WriteString(SubHeaderStyle.Render(fmt.Sprintf("Step-by-Step Hop Path (%d Total Hops)", len(res.Steps))) + "\n\n")

	if len(res.Steps) == 0 {
		stepsBox.WriteString(MutedStyle.Render("  No outbound transfers met heuristic criteria from the starting address."))
	} else {
		pageSize := 8
		start := p.ScrollOffset
		if start < 0 {
			start = 0
		}
		if start >= len(res.Steps) {
			start = len(res.Steps) - 1
		}
		end := start + pageSize
		if end > len(res.Steps) {
			end = len(res.Steps)
		}

		for _, step := range res.Steps[start:end] {
			hopLabel := BadgeHopStyle.Render(fmt.Sprintf("[Step #%d | Hop %d]", step.StepIndex, step.Depth))
			fromStr := AddressStyle.Render(FormatShortAddr(step.FromAddress))
			toStr := AddressStyle.Render(FormatShortAddr(step.ToAddress))
			amtStr := AmountStyle.Render(fmt.Sprintf("%.4f %s", step.AmountFloat, step.AssetSymbol))

			tag := ""
			if step.Attribution != nil && step.Attribution.IsTerminal {
				tag = " " + BadgeExchangeHitStyle.Render(fmt.Sprintf("★ %s (%s)", step.Attribution.EntityName, step.Attribution.WalletLabel))
			}

			stepsBox.WriteString(fmt.Sprintf("  %s %s ──(%s)──▶ %s%s\n", hopLabel, fromStr, amtStr, toStr, tag))
			stepsBox.WriteString(fmt.Sprintf("     %s\n", MutedStyle.Render(fmt.Sprintf("TxID: %s | %s", step.TxID, step.HeuristicDetail))))
		}

		if len(res.Steps) > pageSize {
			stepsBox.WriteString("\n" + MutedStyle.Render(fmt.Sprintf("  Showing steps %d-%d of %d (Use ↑/↓ to scroll)", start+1, end, len(res.Steps))))
		}
	}

	b.WriteString(PanelStyle.Render(stepsBox.String()))
	b.WriteString("\n")
	b.WriteString(MutedStyle.Render("  [↑/↓] Scroll Steps   [n] Trace Another Address   [q / Ctrl+C] Quit"))
	b.WriteString("\n")

	return b.String()
}

