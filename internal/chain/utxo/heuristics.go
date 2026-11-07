package utxo

import (
	"fmt"
	"math/big"
	"strings"

	"crypto-tracker/internal/model"
)

// AnalyzeUTXOOutputs evaluates all outputs of a Bitcoin transaction spent from sourceAddress,
// filtering dust, skipping self/co-input transfers, and applying multi-factor change heuristics.
func AnalyzeUTXOOutputs(tx model.Transaction, sourceAddress string, dustThresholdSats int64) []model.HopTarget {
	// 1. Collect all input addresses (Common-Input Ownership / Co-spend cluster)
	inputAddrs := make(map[string]bool, len(tx.Inputs))
	var sourceScriptType model.ScriptType = model.ScriptUnknown
	var smallestInputSats int64 = -1

	for _, in := range tx.Inputs {
		if in.Address != "" {
			inputAddrs[strings.TrimSpace(in.Address)] = true
		}
		if strings.EqualFold(in.Address, sourceAddress) && in.ScriptType != model.ScriptUnknown {
			sourceScriptType = in.ScriptType
		}
		if in.ValueRaw != nil && in.ValueRaw.IsInt64() {
			sats := in.ValueRaw.Int64()
			if sats > 0 && (smallestInputSats == -1 || sats < smallestInputSats) {
				smallestInputSats = sats
			}
		}
	}

	if sourceScriptType == model.ScriptUnknown {
		sourceScriptType = inferScriptTypeFromAddress(sourceAddress)
	}

	var targets []model.HopTarget
	var candidateIndices []int

	for _, out := range tx.Outputs {
		addr := strings.TrimSpace(out.Address)
		if addr == "" {
			// OP_RETURN or unparseable output script
			continue
		}

		sats := int64(0)
		if out.ValueRaw != nil && out.ValueRaw.IsInt64() {
			sats = out.ValueRaw.Int64()
		}

		// Check 1: Self-transfer or return to a co-spent input address
		if inputAddrs[addr] {
			targets = append(targets, model.HopTarget{
				TxID:            tx.TxID,
				Timestamp:       tx.Timestamp,
				FromAddress:     sourceAddress,
				ToAddress:       addr,
				AmountFloat:     out.ValueFloat,
				AssetSymbol:     "BTC",
				IsSkipped:       true,
				Reason:          model.ReasonSelfTransfer,
				HeuristicDetail: "Output returns directly to input cluster address (address reuse)",
			})
			continue
		}

		// Check 2: Dust Filter
		if sats <= dustThresholdSats {
			targets = append(targets, model.HopTarget{
				TxID:            tx.TxID,
				Timestamp:       tx.Timestamp,
				FromAddress:     sourceAddress,
				ToAddress:       addr,
				AmountFloat:     out.ValueFloat,
				AssetSymbol:     "BTC",
				IsSkipped:       true,
				Reason:          model.ReasonDustFiltered,
				HeuristicDetail: fmt.Sprintf("Output %d sats <= dust threshold (%d sats)", sats, dustThresholdSats),
			})
			continue
		}

		idx := len(targets)
		targets = append(targets, model.HopTarget{
			TxID:            tx.TxID,
			Timestamp:       tx.Timestamp,
			FromAddress:     sourceAddress,
			ToAddress:       addr,
			AmountFloat:     out.ValueFloat,
			AssetSymbol:     "BTC",
			IsSkipped:       false,
			Reason:          model.ReasonOutboundTransfer,
			HeuristicDetail: "Primary outbound UTXO transfer",
		})
		candidateIndices = append(candidateIndices, idx)
	}

	// If there is only 1 candidate output (e.g. full sweep / consolidation), it is a direct sweep
	if len(candidateIndices) == 1 {
		idx := candidateIndices[0]
		targets[idx].Reason = model.ReasonDepositSweep
		targets[idx].HeuristicDetail = "Single non-dust output (100% UTXO sweep / consolidation)"
		return targets
	}

	// Apply Change Address Heuristics when there are 2 candidate outputs (standard payment + change)
	if len(candidateIndices) == 2 {
		i0, i1 := candidateIndices[0], candidateIndices[1]
		out0 := findOutputByAddress(tx.Outputs, targets[i0].ToAddress)
		out1 := findOutputByAddress(tx.Outputs, targets[i1].ToAddress)

		changeIdx, reason := detectChangeBetweenTwoOutputs(out0, out1, sourceScriptType, smallestInputSats)
		if changeIdx == 0 {
			targets[i0].IsSkipped = true
			targets[i0].Reason = model.ReasonChangeSkipped
			targets[i0].HeuristicDetail = reason
		} else if changeIdx == 1 {
			targets[i1].IsSkipped = true
			targets[i1].Reason = model.ReasonChangeSkipped
			targets[i1].HeuristicDetail = reason
		}
	}

	return targets
}

// detectChangeBetweenTwoOutputs applies multi-factor Bitcoin change heuristics between two candidate outputs.
// Returns 0 if out0 is change, 1 if out1 is change, or -1 if ambiguous.
func detectChangeBetweenTwoOutputs(
	out0, out1 model.TxOutput,
	senderScript model.ScriptType,
	smallestInputSats int64,
) (int, string) {
	sats0 := getSats(out0.ValueRaw)
	sats1 := getSats(out1.ValueRaw)

	script0 := out0.ScriptType
	if script0 == model.ScriptUnknown {
		script0 = inferScriptTypeFromAddress(out0.Address)
	}
	script1 := out1.ScriptType
	if script1 == model.ScriptUnknown {
		script1 = inferScriptTypeFromAddress(out1.Address)
	}

	// Heuristic A: Script Type Match
	// Wallet software generates change addresses matching its own native script type.
	if senderScript != model.ScriptUnknown && script0 != script1 {
		if script0 == senderScript {
			return 0, fmt.Sprintf("Script-Type Heuristic: matches sender script (%s) while payment uses %s", script0, script1)
		}
		if script1 == senderScript {
			return 1, fmt.Sprintf("Script-Type Heuristic: matches sender script (%s) while payment uses %s", script1, script0)
		}
	}

	// Heuristic B: Round Number Payment Heuristic
	// Payments are frequently round BTC figures (divisible by 100,000 sats = 0.001 BTC), whereas change has fee residue.
	round0 := isRoundSats(sats0)
	round1 := isRoundSats(sats1)
	if round0 && !round1 {
		return 1, "Round-Amount Heuristic: non-round residual output after fee deduction"
	}
	if round1 && !round0 {
		return 0, "Round-Amount Heuristic: non-round residual output after fee deduction"
	}

	// Heuristic C: Optimal Change / Unnecessary Input Heuristic (UIH1)
	// In multi-input transactions, if exactly one output is smaller than the smallest input, that output is change.
	if smallestInputSats > 0 {
		smaller0 := sats0 < smallestInputSats
		smaller1 := sats1 < smallestInputSats
		if smaller0 && !smaller1 {
			return 0, fmt.Sprintf("Optimal-Change (UIH1) Heuristic: output (%d sats) < smallest input (%d sats)", sats0, smallestInputSats)
		}
		if smaller1 && !smaller0 {
			return 1, fmt.Sprintf("Optimal-Change (UIH1) Heuristic: output (%d sats) < smallest input (%d sats)", sats1, smallestInputSats)
		}
	}

	return -1, ""
}

func isRoundSats(sats int64) bool {
	if sats <= 0 {
		return false
	}
	// Divisible by 0.001 BTC (100,000 sats) or 0.0005 BTC (50,000 sats)
	return sats%100_000 == 0 || sats%50_000 == 0
}

func getSats(v *big.Int) int64 {
	if v == nil || !v.IsInt64() {
		return 0
	}
	return v.Int64()
}

func findOutputByAddress(outputs []model.TxOutput, addr string) model.TxOutput {
	for _, out := range outputs {
		if out.Address == addr {
			return out
		}
	}
	return model.TxOutput{}
}

func inferScriptTypeFromAddress(addr string) model.ScriptType {
	a := strings.TrimSpace(addr)
	switch {
	case strings.HasPrefix(a, "bc1p"):
		return model.ScriptP2TR
	case strings.HasPrefix(a, "bc1q") && len(a) == 42:
		return model.ScriptP2WPKH
	case strings.HasPrefix(a, "bc1q") && len(a) > 42:
		return model.ScriptP2WSH
	case strings.HasPrefix(a, "3"):
		return model.ScriptP2SH
	case strings.HasPrefix(a, "1"):
		return model.ScriptP2PKH
	default:
		return model.ScriptUnknown
	}
}

