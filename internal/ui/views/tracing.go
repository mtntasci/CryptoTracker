package views

import (
	"fmt"
	"strings"

	"crypto-tracker/internal/model"
)

// TracingViewProps contains the live state of an active BFS tracing session.
type TracingViewProps struct {
	Chain          model.ChainType
	RootAddress    string
	SpinnerView    string
	CurrentDepth   int
	MaxDepth       int
	VisitedCount   int
	QueueSize      int
	CurrentAddress string
	StatusMessage  string
	Steps          []model.TraceStep
	Skipped        []model.SkippedOutput
	ExchangeHits   []model.TraceStep
}

// RenderTracingView renders the animated step-by-step hopping UI.
func RenderTracingView(p TracingViewProps) string {
	var b strings.Builder

	meta := model.GetChainMetadata(p.Chain)

	b.WriteString(HeaderStyle.Render(fmt.Sprintf(" CRYPTO-TRACKER // LIVE %s (%s) BFS TRACING ", meta.Type, meta.Architecture)))
	b.WriteString("\n\n")

	// 1. Live Status & Progress Indicator Panel
	var statusPanel strings.Builder
	depthBar := renderDepthProgressBar(p.CurrentDepth, p.MaxDepth)
	statusPanel.WriteString(fmt.Sprintf("  %s %s\n\n", p.SpinnerView, SubHeaderStyle.Render(p.StatusMessage)))
	statusPanel.WriteString(fmt.Sprintf("  Origin Address  : %s\n", AddressStyle.Render(p.RootAddress)))
	statusPanel.WriteString(fmt.Sprintf("  Active Node     : %s\n", AddressStyle.Render(p.CurrentAddress)))
	statusPanel.WriteString(fmt.Sprintf("  BFS Depth       : %s  (%d / %d Hops)\n", depthBar, p.CurrentDepth, p.MaxDepth))
	statusPanel.WriteString(fmt.Sprintf("  Metrics         : %d Wallets Scanned  |  %d In Queue  |  %d Filtered (Dust/Change)  |  %d Exchange Hits",
		p.VisitedCount, p.QueueSize, len(p.Skipped), len(p.ExchangeHits)))

	b.WriteString(ActivePanelStyle.Render(statusPanel.String()))
	b.WriteString("\n")

	// 2. Step-by-Step Animated Hop Graph
	var hopPanel strings.Builder
	hopPanel.WriteString(SubHeaderStyle.Render("Live Transaction Hop Graph (BFS Traversal)") + "\n\n")

	if len(p.Steps) == 0 {
		hopPanel.WriteString(MutedStyle.Render("  Scanning initial outbound transactions from starting wallet..."))
	} else {
		startIdx := 0
		if len(p.Steps) > 10 {
			startIdx = len(p.Steps) - 10
			hopPanel.WriteString(MutedStyle.Render(fmt.Sprintf("  ... (%d earlier hops omitted for view) ...\n", startIdx)))
		}
		for _, step := range p.Steps[startIdx:] {
			indent := strings.Repeat("  ", step.Depth-1)
			hopTag := BadgeHopStyle.Render(fmt.Sprintf("[Hop %d]", step.Depth))
			fromStr := AddressStyle.Render(FormatShortAddr(step.FromAddress))
			toStr := AddressStyle.Render(FormatShortAddr(step.ToAddress))
			amtStr := AmountStyle.Render(fmt.Sprintf("%.4f %s", step.AmountFloat, step.AssetSymbol))
			arrow := ArrowStyle.Render(fmt.Sprintf(" ──(%s)──▶ ", amtStr))

			destBadge := ""
			if step.Attribution != nil && step.Attribution.IsTerminal {
				destBadge = " " + BadgeExchangeHitStyle.Render(fmt.Sprintf("★ %s [%s]", step.Attribution.EntityName, step.Attribution.WalletLabel))
			}

			hopPanel.WriteString(fmt.Sprintf("  %s%s %s%s%s%s\n", indent, hopTag, fromStr, arrow, toStr, destBadge))
			hopPanel.WriteString(fmt.Sprintf("  %s       %s\n", indent, MutedStyle.Render(fmt.Sprintf("└─ Tx: %s | %s", FormatShortAddr(step.TxID), step.HeuristicDetail))))
		}
	}

	b.WriteString(PanelStyle.Render(hopPanel.String()))

	// 3. Heuristic Filter Activity (Change Addresses Skipped & Dust Filtered)
	if len(p.Skipped) > 0 {
		var filterPanel strings.Builder
		filterPanel.WriteString(SubHeaderStyle.Render("Heuristic Filter Log (Dust & Change Address Exclusion)") + "\n\n")
		startSkip := 0
		if len(p.Skipped) > 4 {
			startSkip = len(p.Skipped) - 4
		}
		for _, sk := range p.Skipped[startSkip:] {
			filterPanel.WriteString(fmt.Sprintf("  %s [Hop %d] %s (%.6f %s) — %s\n",
				BadgeSkippedStyle.Render("⊘ "+string(sk.Reason)),
				sk.Depth,
				AddressStyle.Render(FormatShortAddr(sk.ToAddress)),
				sk.AmountFloat,
				sk.AssetSymbol,
				MutedStyle.Render(sk.Detail),
			))
		}
		b.WriteString("\n")
		b.WriteString(PanelStyle.Render(filterPanel.String()))
	}

	b.WriteString("\n")
	b.WriteString(MutedStyle.Render("  [Ctrl+C / q] Cancel Trace & Exit"))
	b.WriteString("\n")

	return b.String()
}

func renderDepthProgressBar(current, max int) string {
	if max <= 0 {
		max = 5
	}
	var segments []string
	for i := 1; i <= max; i++ {
		if i <= current {
			segments = append(segments, BadgeSelectedChainStyle.Render(fmt.Sprintf("H%d", i)))
		} else {
			segments = append(segments, MutedStyle.Render(fmt.Sprintf("[H%d]", i)))
		}
	}
	return strings.Join(segments, "━")
}

