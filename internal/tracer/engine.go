package tracer

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"crypto-tracker/internal/attribution"
	"crypto-tracker/internal/chain"
	"crypto-tracker/internal/config"
	"crypto-tracker/internal/model"
)

type bfsItem struct {
	Address string
	Depth   int
}

// Engine coordinates Breadth-First Search (BFS) transaction tracing across chains
// and integrates with the Attribution Engine.
type Engine struct {
	cfg         *config.Config
	provider    chain.Provider
	attribution *attribution.Engine
}

// NewEngine creates a new BFS Tracing Engine.
func NewEngine(cfg *config.Config, provider chain.Provider, attr *attribution.Engine) *Engine {
	return &Engine{
		cfg:         cfg,
		provider:    provider,
		attribution: attr,
	}
}

// Trace executes a Breadth-First Search (BFS) traversal starting from rootAddress up to cfg.MaxDepth,
// streaming live events to eventCh.
func (e *Engine) Trace(ctx context.Context, rootAddress string, eventCh chan<- TraceEvent) (*model.TraceResult, error) {
	defer close(eventCh)

	start := time.Now()
	trimmedRoot := strings.TrimSpace(rootAddress)
	chainType := e.provider.ChainType()

	res := &model.TraceResult{
		RootAddress: trimmedRoot,
		Chain:       chainType,
		StartTime:   start,
	}

	e.emit(ctx, eventCh, TraceEvent{
		Type:        EventInit,
		Timestamp:   time.Now(),
		Depth:       0,
		MaxDepth:    e.cfg.MaxDepth,
		CurrentAddr: trimmedRoot,
		Message:     fmt.Sprintf("Initializing %s (%s) BFS Tracing Engine (Max Depth: %d hops)", chainType, e.provider.Architecture(), e.cfg.MaxDepth),
	})

	// Check if starting address itself is already an attributed Exchange Hot/Cold Wallet
	rootAttr := e.attribution.IdentifyFast(chainType, trimmedRoot)
	var rootSummary *model.AddressSummary
	if rootAttr == nil {
		rootSummary, _ = e.provider.FetchAddressSummary(ctx, trimmedRoot)
		rootAttr = e.attribution.Identify(ctx, chainType, trimmedRoot, rootSummary)
	}
	if rootAttr != nil && rootAttr.IsTerminal {
		hitStep := model.TraceStep{
			StepIndex:       1,
			Depth:           0,
			FromAddress:     trimmedRoot,
			ToAddress:       trimmedRoot,
			TxID:            "DIRECT_ADDRESS_MATCH",
			Timestamp:       time.Now(),
			AmountFloat:     rootSummaryBalance(rootSummary),
			AssetSymbol:     string(chainType),
			Reason:          model.ReasonOutboundTransfer,
			HeuristicDetail: "Starting address is directly identified as an Exchange Wallet",
			Attribution:     rootAttr,
		}
		res.Steps = append(res.Steps, hitStep)
		res.TerminalHits = append(res.TerminalHits, hitStep)
		res.TotalVisited = 1
		res.EndTime = time.Now()

		e.emit(ctx, eventCh, TraceEvent{
			Type:        EventExchangeHit,
			Timestamp:   time.Now(),
			Depth:       0,
			MaxDepth:    e.cfg.MaxDepth,
			VisitedCnt:  1,
			CurrentAddr: trimmedRoot,
			Message:     fmt.Sprintf("Starting address belongs to %s (%s)", rootAttr.EntityName, rootAttr.WalletLabel),
			Step:        &hitStep,
		})

		e.emit(ctx, eventCh, TraceEvent{
			Type:       EventComplete,
			Timestamp:  time.Now(),
			Depth:      0,
			MaxDepth:   e.cfg.MaxDepth,
			VisitedCnt: 1,
			Message:    "Tracing complete: starting address is a verified exchange wallet.",
			Result:     res,
		})
		return res, nil
	}

	queue := []bfsItem{{Address: trimmedRoot, Depth: 1}}
	visited := make(map[string]bool)
	visited[normalizeKey(chainType, trimmedRoot)] = true

	stepIndex := 0

	for len(queue) > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		current := queue[0]
		queue = queue[1:]

		if current.Depth > e.cfg.MaxDepth {
			res.ActiveBranches = appendUnique(res.ActiveBranches, current.Address)
			continue
		}

		if current.Depth > res.MaxDepthReached {
			res.MaxDepthReached = current.Depth
		}
		res.TotalVisited++

		e.emit(ctx, eventCh, TraceEvent{
			Type:        EventScanningNode,
			Timestamp:   time.Now(),
			Depth:       current.Depth,
			MaxDepth:    e.cfg.MaxDepth,
			QueueSize:   len(queue),
			VisitedCnt:  res.TotalVisited,
			CurrentAddr: current.Address,
			Message:     fmt.Sprintf("[Hop %d/%d] Querying outbound transactions for %s...", current.Depth, e.cfg.MaxDepth, shortAddr(current.Address)),
		})

		txs, err := e.provider.FetchOutboundTransactions(ctx, current.Address, e.cfg.MaxTxPerAddress)
		if err != nil {
			e.emit(ctx, eventCh, TraceEvent{
				Type:        EventError,
				Timestamp:   time.Now(),
				Depth:       current.Depth,
				MaxDepth:    e.cfg.MaxDepth,
				CurrentAddr: current.Address,
				Message:     fmt.Sprintf("Warning fetching txs for %s: %v", shortAddr(current.Address), err),
				Err:         err,
			})
			res.ActiveBranches = appendUnique(res.ActiveBranches, current.Address)
			continue
		}

		res.TotalTxAnalyzed += len(txs)
		if len(txs) == 0 {
			// Unspent terminal wallet (funds still sitting in this address or no outbound activity)
			res.ActiveBranches = appendUnique(res.ActiveBranches, current.Address)
			continue
		}

		var validTargets []model.HopTarget
		for _, tx := range txs {
			targets, err := e.provider.ResolveHopTargets(ctx, tx, current.Address)
			if err != nil {
				continue
			}

			for _, tgt := range targets {
				if tgt.IsSkipped {
					skipped := model.SkippedOutput{
						Depth:       current.Depth,
						FromAddress: tgt.FromAddress,
						ToAddress:   tgt.ToAddress,
						TxID:        tgt.TxID,
						AmountFloat: tgt.AmountFloat,
						AssetSymbol: tgt.AssetSymbol,
						Reason:      tgt.Reason,
						Detail:      tgt.HeuristicDetail,
					}
					res.SkippedOutputs = append(res.SkippedOutputs, skipped)
					e.emit(ctx, eventCh, TraceEvent{
						Type:        EventOutputSkipped,
						Timestamp:   time.Now(),
						Depth:       current.Depth,
						MaxDepth:    e.cfg.MaxDepth,
						QueueSize:   len(queue),
						VisitedCnt:  res.TotalVisited,
						CurrentAddr: current.Address,
						Message:     fmt.Sprintf("%s -> %s (%.6f %s)", tgt.Reason, shortAddr(tgt.ToAddress), tgt.AmountFloat, tgt.AssetSymbol),
						Skipped:     &skipped,
					})
					continue
				}
				validTargets = append(validTargets, tgt)
			}
		}

		if len(validTargets) == 0 {
			res.ActiveBranches = appendUnique(res.ActiveBranches, current.Address)
			continue
		}

		// Prioritize highest value outbound transfers first
		sort.SliceStable(validTargets, func(i, j int) bool {
			return validTargets[i].AmountFloat > validTargets[j].AmountFloat
		})

		// Deduplicate destination addresses within the same hop and cap at MaxBranches
		seenInHop := make(map[string]bool)
		branchesAdded := 0

		for _, tgt := range validTargets {
			if branchesAdded >= e.cfg.MaxBranches {
				break
			}
			destKey := normalizeKey(chainType, tgt.ToAddress)
			if seenInHop[destKey] {
				continue
			}
			seenInHop[destKey] = true
			branchesAdded++
			stepIndex++

			// Check fast attribution first; only query remote summary if not in local verified registry
			attrInfo := e.attribution.IdentifyFast(chainType, tgt.ToAddress)
			if attrInfo == nil {
				destSummary, _ := e.provider.FetchAddressSummary(ctx, tgt.ToAddress)
				attrInfo = e.attribution.Identify(ctx, chainType, tgt.ToAddress, destSummary)
			}

			step := model.TraceStep{
				StepIndex:       stepIndex,
				Depth:           current.Depth,
				FromAddress:     tgt.FromAddress,
				ToAddress:       tgt.ToAddress,
				TxID:            tgt.TxID,
				Timestamp:       tgt.Timestamp,
				AmountFloat:     tgt.AmountFloat,
				AssetSymbol:     tgt.AssetSymbol,
				Reason:          tgt.Reason,
				HeuristicDetail: tgt.HeuristicDetail,
				Attribution:     attrInfo,
			}
			res.Steps = append(res.Steps, step)

			if e.cfg.StepDelay > 0 {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(e.cfg.StepDelay):
				}
			}

			if attrInfo != nil && attrInfo.IsTerminal {
				res.TerminalHits = append(res.TerminalHits, step)
				e.emit(ctx, eventCh, TraceEvent{
					Type:        EventExchangeHit,
					Timestamp:   time.Now(),
					Depth:       current.Depth,
					MaxDepth:    e.cfg.MaxDepth,
					QueueSize:   len(queue),
					VisitedCnt:  res.TotalVisited,
					CurrentAddr: tgt.ToAddress,
					Message:     fmt.Sprintf("EXCHANGE IDENTIFIED: %s [%s] via %.4f %s", attrInfo.EntityName, attrInfo.WalletLabel, tgt.AmountFloat, tgt.AssetSymbol),
					Step:        &step,
				})
				// Do not enqueue exchange hot wallets into BFS (mixing pool reached)
				continue
			}

			e.emit(ctx, eventCh, TraceEvent{
				Type:        EventHopDiscovered,
				Timestamp:   time.Now(),
				Depth:       current.Depth,
				MaxDepth:    e.cfg.MaxDepth,
				QueueSize:   len(queue),
				VisitedCnt:  res.TotalVisited,
				CurrentAddr: tgt.ToAddress,
				Message:     fmt.Sprintf("Hop %d: %s ──[%.4f %s]──> %s", current.Depth, shortAddr(tgt.FromAddress), tgt.AmountFloat, tgt.AssetSymbol, shortAddr(tgt.ToAddress)),
				Step:        &step,
			})

			if !visited[destKey] {
				visited[destKey] = true
				if current.Depth < e.cfg.MaxDepth {
					queue = append(queue, bfsItem{
						Address: tgt.ToAddress,
						Depth:   current.Depth + 1,
					})
				} else {
					res.ActiveBranches = appendUnique(res.ActiveBranches, tgt.ToAddress)
				}
			}
		}
	}

	res.EndTime = time.Now()
	e.emit(ctx, eventCh, TraceEvent{
		Type:       EventComplete,
		Timestamp:  time.Now(),
		Depth:      res.MaxDepthReached,
		MaxDepth:   e.cfg.MaxDepth,
		VisitedCnt: res.TotalVisited,
		Message:    fmt.Sprintf("BFS Tracing finished in %s (%d hops, %d exchange hits)", res.EndTime.Sub(res.StartTime).Round(time.Millisecond), len(res.Steps), len(res.TerminalHits)),
		Result:     res,
	})

	return res, nil
}

func (e *Engine) emit(ctx context.Context, ch chan<- TraceEvent, ev TraceEvent) {
	select {
	case <-ctx.Done():
	case ch <- ev:
	}
}

func normalizeKey(c model.ChainType, addr string) string {
	trimmed := strings.TrimSpace(addr)
	if c == model.ChainEthereum {
		return strings.ToLower(trimmed)
	}
	return trimmed
}

func shortAddr(addr string) string {
	if len(addr) <= 14 {
		return addr
	}
	return addr[:7] + "..." + addr[len(addr)-5:]
}

func rootSummaryBalance(s *model.AddressSummary) float64 {
	if s == nil {
		return 0
	}
	return s.BalanceFloat
}

func appendUnique(slice []string, item string) []string {
	for _, existing := range slice {
		if strings.EqualFold(existing, item) {
			return slice
		}
	}
	return append(slice, item)
}
