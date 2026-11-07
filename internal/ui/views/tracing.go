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

// RenderTracingView renders the animated step-by-step stage traversal UI.
func RenderTracingView(p TracingViewProps) string {
	var b strings.Builder

	b.WriteString(RenderBigBanner())
	b.WriteString("\n")

	// 1. Live Status & Progress Indicator Panel
	var statusPanel strings.Builder
	depthBar := renderStageProgressBar(p.CurrentDepth, p.MaxDepth)
	statusPanel.WriteString(fmt.Sprintf("  %s %s\n\n", p.SpinnerView, BoldWhiteStyle.Render(p.StatusMessage)))
	statusPanel.WriteString(fmt.Sprintf("  %s : %s\n", BoldWhiteStyle.Render("Origin Address "), AddressStyle.Render(p.RootAddress)))
	statusPanel.WriteString(fmt.Sprintf("  %s : %s\n", BoldWhiteStyle.Render("Active Node    "), AddressStyle.Render(p.CurrentAddress)))
	statusPanel.WriteString(fmt.Sprintf("  %s : %s  %s\n", BoldWhiteStyle.Render("BFS Stage      "), depthBar, BoldWhiteStyle.Render(fmt.Sprintf("(%d / %d Stages)", p.CurrentDepth, p.MaxDepth))))
	statusPanel.WriteString(fmt.Sprintf("  %s : %s",
		BoldWhiteStyle.Render("Metrics        "),
		TextStyle.Render(fmt.Sprintf("%d Wallets Scanned  |  %d In Queue  |  %d Filtered (Dust/Change)  |  %d Exchange Hits",
			p.VisitedCount, p.QueueSize, len(p.Skipped), len(p.ExchangeHits))),
	))

	b.WriteString(ActivePanelStyle.Render(statusPanel.String()))
	b.WriteString("\n")

	// 2. Step-by-Step Animated Stage Graph
	var hopPanel strings.Builder
	hopPanel.WriteString(SubHeaderStyle.Render("Live Transaction Stage Graph (BFS Traversal)") + "\n\n")

	if len(p.Steps) == 0 {
		hopPanel.WriteString(TextStyle.Render("  Scanning initial outbound transactions from starting wallet..."))
	} else {
		startIdx := 0
		if len(p.Steps) > 10 {
			startIdx = len(p.Steps) - 10
			hopPanel.WriteString(TextStyle.Render(fmt.Sprintf("  ... (%d earlier steps omitted for view) ...\n", startIdx)))
		}
		for _, step := range p.Steps[startIdx:] {
			indentDepth := step.Depth - 1
			if indentDepth < 0 {
				indentDepth = 0
			}
			indent := strings.Repeat("  ", indentDepth)

			stageStyle := StageBadgeStyle(step.Depth)
			stageTag := stageStyle.Render(fmt.Sprintf("[Stage %d]", step.Depth))
			fromStr := AddressStyle.Render(FormatShortAddr(step.FromAddress))
			toStr := AddressStyle.Render(FormatShortAddr(step.ToAddress))
			amtStr := AmountStyle.Render(fmt.Sprintf("%.4f %s", step.AmountFloat, step.AssetSymbol))
			arrow := stageStyle.Render(" ──(") + amtStr + stageStyle.Render(")──▶ ")

			destBadge := ""
			if step.Attribution != nil && step.Attribution.IsTerminal {
				destBadge = " " + BadgeExchangeHitStyle.Render(fmt.Sprintf("★ %s [%s]", step.Attribution.EntityName, step.Attribution.WalletLabel))
			}

			hopPanel.WriteString(fmt.Sprintf("  %s%s %s%s%s%s\n", indent, stageTag, fromStr, arrow, toStr, destBadge))
			hopPanel.WriteString(fmt.Sprintf("  %s         %s\n", indent, TextStyle.Render(fmt.Sprintf("└─ Tx: %s | %s", FormatShortAddr(step.TxID), step.HeuristicDetail))))
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
			stageLabel := StageBadgeStyle(sk.Depth).Render(fmt.Sprintf("[Stage %d]", sk.Depth))
			filterPanel.WriteString(fmt.Sprintf("  %s %s %s %s — %s\n",
				BadgeSkippedStyle.Render("⊘ "+string(sk.Reason)),
				stageLabel,
				AddressStyle.Render(FormatShortAddr(sk.ToAddress)),
				BoldWhiteStyle.Render(fmt.Sprintf("(%.6f %s)", sk.AmountFloat, sk.AssetSymbol)),
				TextStyle.Render(sk.Detail),
			))
		}
		b.WriteString("\n")
		b.WriteString(PanelStyle.Render(filterPanel.String()))
	}

	b.WriteString("\n")
	b.WriteString(BoldWhiteStyle.Render("  [Ctrl+C / q] Cancel Trace & Exit"))
	b.WriteString("\n")

	return b.String()
}

func renderStageProgressBar(current, max int) string {
	if max <= 0 {
		max = 5
	}
	var segments []string
	for i := 1; i <= max; i++ {
		if i <= current {
			segments = append(segments, StagePillStyle(i).Render(fmt.Sprintf("Stage %d", i)))
		} else {
			segments = append(segments, TextStyle.Render(fmt.Sprintf("[Stage %d]", i)))
		}
	}
	return strings.Join(segments, "━")
}
