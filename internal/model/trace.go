package model

import "time"

// FilterReason explains why a particular transaction output was skipped or selected.
type FilterReason string

const (
	ReasonOutboundTransfer FilterReason = "Outbound Transfer"
	ReasonDustFiltered     FilterReason = "Filtered: Dust Output"
	ReasonChangeSkipped    FilterReason = "Skipped: Change Address Heuristic"
	ReasonSelfTransfer     FilterReason = "Skipped: Self-Transfer"
	ReasonDepositSweep     FilterReason = "Deposit Sweep Detected"
	ReasonContractEvent    FilterReason = "Smart Contract Transfer Event"
)

// HopTarget represents a resolved destination from a single transaction.
type HopTarget struct {
	TxID            string
	Timestamp       time.Time
	FromAddress     string
	ToAddress       string
	AmountFloat     float64
	AssetSymbol     string
	IsSkipped       bool
	Reason          FilterReason
	HeuristicDetail string
	IsTokenTransfer bool
	ContractAddress string
}

// TraceStep represents one hop in the BFS graph traversal.
type TraceStep struct {
	StepIndex       int
	Depth           int
	FromAddress     string
	ToAddress       string
	TxID            string
	Timestamp       time.Time
	AmountFloat     float64
	AssetSymbol     string
	Reason          FilterReason
	HeuristicDetail string
	Attribution     *AttributionInfo
}

// SkippedOutput records an output that was filtered out (dust or change) for transparency in UI.
type SkippedOutput struct {
	Depth       int
	FromAddress string
	ToAddress   string
	TxID        string
	AmountFloat float64
	AssetSymbol string
	Reason      FilterReason
	Detail      string
}

// TraceResult summarizes the entire tracing session.
type TraceResult struct {
	RootAddress     string
	Chain           ChainType
	StartTime       time.Time
	EndTime         time.Time
	MaxDepthReached int
	TotalVisited    int
	TotalTxAnalyzed int
	Steps           []TraceStep
	SkippedOutputs  []SkippedOutput
	TerminalHits    []TraceStep // Steps that landed on a confirmed Exchange Hot/Cold Wallet
	ActiveBranches  []string    // Addresses where funds remain or tracing reached max depth
}

