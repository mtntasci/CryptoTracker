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

	b.WriteString(RenderBigBanner())
	b.WriteString("\n")

	// 1. Attribution Outcome Banner
	if len(res.TerminalHits) > 0 {
		var hitBanner strings.Builder
		hitBanner.WriteString(BadgeExchangeHitStyle.Render(" ✔ DESTINATION EXCHANGE HOT WALLET IDENTIFIED ") + "\n\n")
		for idx, hit := range res.TerminalHits {
			attr := hit.Attribution
			stageBadge := StageBadgeStyle(hit.Depth).Render(fmt.Sprintf("Stage %d", hit.Depth))
			hitBanner.WriteString(fmt.Sprintf("  %s : %s %s\n",
				BoldWhiteStyle.Render(fmt.Sprintf("%d. Exchange Entity  ", idx+1)),
				SubHeaderStyle.Render(attr.EntityName),
				BoldWhiteStyle.Render(fmt.Sprintf("(%s)", attr.WalletLabel)),
			))
			hitBanner.WriteString(fmt.Sprintf("     %s : %s\n",
				BoldWhiteStyle.Render("Wallet Category  "),
				TextStyle.Render(fmt.Sprintf("%s (Confidence: %.0f%% | Source: %s)", attr.Category, attr.Confidence*100, attr.Source)),
			))
			hitBanner.WriteString(fmt.Sprintf("     %s : %s\n",
				BoldWhiteStyle.Render("Hot Wallet Addr  "),
				AddressStyle.Render(hit.ToAddress),
			))
			hitBanner.WriteString(fmt.Sprintf("     %s : %s at %s %s\n",
				BoldWhiteStyle.Render("Sweep / Transfer "),
				AmountStyle.Render(fmt.Sprintf("%.4f %s", hit.AmountFloat, hit.AssetSymbol)),
				stageBadge,
				TextStyle.Render(fmt.Sprintf("(Tx: %s)", FormatShortAddr(hit.TxID))),
			))
			if idx < len(res.TerminalHits)-1 {
				hitBanner.WriteString("\n")
			}
		}
		b.WriteString(SuccessBannerStyle.Render(hitBanner.String()))
	} else {
		var ongoingBanner strings.Builder
		ongoingBanner.WriteString(BadgeSkippedStyle.Render(" ◉ ONGOING TRACKING // NO KNOWN EXCHANGE HOT WALLET REACHED WITHIN MAX STAGE ") + "\n\n")
		ongoingBanner.WriteString(TextStyle.Render(fmt.Sprintf("  Funds have not entered a known exchange hot wallet within %d stages.\n", res.MaxDepthReached)))
		if len(res.ActiveBranches) > 0 {
			ongoingBanner.WriteString(BoldWhiteStyle.Render("  Current Terminal / Unspent Frontier Wallets for Ongoing Monitoring:\n"))
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
	summary.WriteString(fmt.Sprintf("  %s : %s\n", BoldWhiteStyle.Render("Root Wallet         "), AddressStyle.Render(res.RootAddress)))
	summary.WriteString(fmt.Sprintf("  %s : %s\n", BoldWhiteStyle.Render("Chain & Architecture"), TextStyle.Render(fmt.Sprintf("%s (%s)  |  Elapsed Time: %s", meta.Name, meta.Architecture, duration))))
	summary.WriteString(fmt.Sprintf("  %s : %s\n", BoldWhiteStyle.Render("Traversal Stats     "), TextStyle.Render(fmt.Sprintf("Max Stage Reached: %d  |  Wallets Visited: %d  |  Transactions Analyzed: %d",
		res.MaxDepthReached, res.TotalVisited, res.TotalTxAnalyzed))))
	summary.WriteString(fmt.Sprintf("  %s : %s\n", BoldWhiteStyle.Render("Heuristics Applied  "), TextStyle.Render(fmt.Sprintf("%d Valid Outbound Transfers  |  %d Filtered Outputs (Dust / Change / Self)",
		len(res.Steps), len(res.SkippedOutputs)))))

	b.WriteString(PanelStyle.Render(summary.String()))
	b.WriteString("\n")

	// 3. Complete Step-by-Step Stage Breakdown
	var stepsBox strings.Builder
	stepsBox.WriteString(SubHeaderStyle.Render(fmt.Sprintf("Step-by-Step Stage Path (%d Total Steps)", len(res.Steps))) + "\n\n")

	if len(res.Steps) == 0 {
		stepsBox.WriteString(TextStyle.Render("  No outbound transfers met heuristic criteria from the starting address."))
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
			stageStyle := StageBadgeStyle(step.Depth)
			stageLabel := stageStyle.Render(fmt.Sprintf("[Step #%d | Stage %d]", step.StepIndex, step.Depth))
			fromStr := AddressStyle.Render(FormatShortAddr(step.FromAddress))
			toStr := AddressStyle.Render(FormatShortAddr(step.ToAddress))
			amtStr := AmountStyle.Render(fmt.Sprintf("%.4f %s", step.AmountFloat, step.AssetSymbol))
			arrow := stageStyle.Render(" ──(") + amtStr + stageStyle.Render(")──▶ ")

			tag := ""
			if step.Attribution != nil && step.Attribution.IsTerminal {
				tag = " " + BadgeExchangeHitStyle.Render(fmt.Sprintf("★ %s (%s)", step.Attribution.EntityName, step.Attribution.WalletLabel))
			}

			stepsBox.WriteString(fmt.Sprintf("  %s %s%s%s%s\n", stageLabel, fromStr, arrow, toStr, tag))
			stepsBox.WriteString(fmt.Sprintf("     %s\n", TextStyle.Render(fmt.Sprintf("TxID: %s | %s", step.TxID, step.HeuristicDetail))))
		}

		if len(res.Steps) > pageSize {
			stepsBox.WriteString("\n" + BoldWhiteStyle.Render(fmt.Sprintf("  Showing steps %d-%d of %d (Use ↑/↓ to scroll)", start+1, end, len(res.Steps))))
		}
	}

	b.WriteString(PanelStyle.Render(stepsBox.String()))
	b.WriteString("\n")
	b.WriteString(BoldWhiteStyle.Render("  [↑/↓] Scroll Steps   [n] Trace Another Address   [q / Ctrl+C] Quit"))
	b.WriteString("\n")

	return b.String()
}
